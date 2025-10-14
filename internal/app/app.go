package app

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"go-api-kbt/internal/config"
	"go-api-kbt/internal/database"
	domainActivity "go-api-kbt/internal/domain/activity"
	domainEvent "go-api-kbt/internal/domain/event"
	domainLocation "go-api-kbt/internal/domain/location"
	domainMedaler "go-api-kbt/internal/domain/medaler"
	domainUser "go-api-kbt/internal/domain/user"

	repoActivity "go-api-kbt/internal/repository/activity"
	repoAuth "go-api-kbt/internal/repository/auth"
	repoEvent "go-api-kbt/internal/repository/event"
	repoLocation "go-api-kbt/internal/repository/location"
	repoMedaler "go-api-kbt/internal/repository/medaler"
	repoUser "go-api-kbt/internal/repository/user"

	serviceActivity "go-api-kbt/internal/service/activity"
	serviceAuth "go-api-kbt/internal/service/auth"
	serviceEvent "go-api-kbt/internal/service/event"
	serviceLocation "go-api-kbt/internal/service/location"
	serviceMedaler "go-api-kbt/internal/service/medaler"
	serviceUser "go-api-kbt/internal/service/user"

	transport "go-api-kbt/internal/transport/http"
	handler "go-api-kbt/internal/transport/http/handler"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/redis/go-redis/v9"
)

//go:embed ../../migrations/*.sql
var fs embed.FS

// Application wires dependencies and exposes a runnable HTTP server.
type Application struct {
	server *http.Server
	logger *slog.Logger
	db     database.Database
	redis  *redis.Client
}

// New constructs an application instance using the supplied configuration.
func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*Application, error) {
	db, err := database.NewPostgres(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	// Run migrations if enabled
	if cfg.Database.RunMigrations {
		logger.Info("running database migrations")
		source, err := iofs.New(fs, "../../migrations")
		if err != nil {
			return nil, fmt.Errorf("failed to create iofs source: %w", err)
		}
		m, err := migrate.NewWithSourceInstance("iofs", source, db.DSN())
		if err != nil {
			return nil, fmt.Errorf("failed to create migrate instance: %w", err)
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return nil, fmt.Errorf("failed to apply migrations: %w", err)
		}
		logger.Info("database migrations applied successfully")
	}

	// GORM AutoMigrate is removed, as golang-migrate handles schema management.
	// However, for initial development or if not using golang-migrate, you might keep it.
	// if err := db.AutoMigrate(&domainUser.Entity{}, &domainEvent.Entity{}, &domainLocation.Entity{}, &domainMedaler.Medaler{}, &domainActivity.Activity{}); err != nil {
	// 	return nil, fmt.Errorf("auto migrate: %w", err)
	// }

	if err := database.SeedAdmin(ctx, db, cfg); err != nil {
		return nil, fmt.Errorf("seed admin: %w", err)
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: 100,
		PoolTimeout: cfg.Redis.Timeout,
		ReadTimeout: cfg.Redis.Timeout,
		WriteTimeout: cfg.Redis.Timeout,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	userRepository := repoUser.NewGormRepository(db.GormDB(), cfg.Database.StatementTimeout)
	userService := serviceUser.NewService(userRepository, cfg.Auth.HashCost)
	userHandler := handler.NewUserHandler(userService, logger)

	eventRepository := repoEvent.NewGormRepository(db.GormDB())
	eventService := serviceEvent.NewService(eventRepository)
	eventHandler := handler.NewEventHandler(eventService, logger)

	locationRepository := repoLocation.NewGormRepository(db.GormDB())
	locationService := serviceLocation.NewService(locationRepository)
	locationHandler := handler.NewLocationHandler(locationService, logger)

	medalerRepository := repoMedaler.NewGormRepository(db.GormDB(), cfg.Database.StatementTimeout)
	medalerService := serviceMedaler.NewService(medalerRepository)
	medalerHandler := handler.NewMedalerHandler(medalerService, logger)

	activityRepository := repoActivity.NewGormRepository(db.GormDB(), cfg.Database.StatementTimeout)
	activityService := serviceActivity.NewService(activityRepository)
	activityHandler := handler.NewActivityHandler(activityService, logger)

	authRepository := repoAuth.NewRedisRepository(redisClient)
	authService := serviceAuth.NewService(userRepository, authRepository, &cfg.Auth)
	authHandler := handler.NewAuthHandler(authService, logger)

	router := transport.NewRouterBuilder(userHandler, cfg, db, redisClient).
		WithEventHandler(eventHandler).
		WithLocationHandler(locationHandler).
		WithMedalerHandler(medalerHandler).
		WithActivityHandler(activityHandler).
		WithAuthHandler(authHandler).
		WithMiddlewares().
		Build()

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Application{
		server: srv,
		logger: logger,
		db:     db,
		redis:  redisClient,
	}, nil
}

// Start launches the HTTP server.
func (a *Application) Start() error {
	a.logger.Info("starting HTTP server", slog.String("addr", a.server.Addr))
	return a.server.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (a *Application) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			return fmt.Errorf("close redis client: %w", err)
		}
	}

	return a.db.Close()
}
