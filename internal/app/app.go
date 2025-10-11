package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"go-api-kbt/internal/config"
	"go-api-kbt/internal/database"
	domainUser "go-api-kbt/internal/domain/user"
	domainEvent "go-api-kbt/internal/domain/event"
	domainLocation "go-api-kbt/internal/domain/location"

	repoUser "go-api-kbt/internal/repository/user"
	repoEvent "go-api-kbt/internal/repository/event"
	repoLocation "go-api-kbt/internal/repository/location"

	serviceUser "go-api-kbt/internal/service/user"
	serviceEvent "go-api-kbt/internal/service/event"
	serviceLocation "go-api-kbt/internal/service/location"

	transport "go-api-kbt/internal/transport/http"
	handler "go-api-kbt/internal/transport/http/handler"
)

// Application wires dependencies and exposes a runnable HTTP server.
type Application struct {
	server *http.Server
	logger *slog.Logger
	db     database.Database
}

// New constructs an application instance using the supplied configuration.
func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*Application, error) {
	db, err := database.NewPostgres(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	if err := db.AutoMigrate(&domainUser.Entity{}, &domainEvent.Entity{}, &domainLocation.Entity{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	if err := database.SeedAdmin(ctx, db, cfg); err != nil {
		return nil, fmt.Errorf("seed admin: %w", err)
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

	router := transport.NewRouterBuilder(userHandler).
		WithEventHandler(eventHandler).
		WithLocationHandler(locationHandler).
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

	return a.db.Close()
}
