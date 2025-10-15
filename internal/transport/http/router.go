package http

import (
	"net/http"
	"time"

	"go-api-kbt/internal/config"
	"go-api-kbt/internal/database"
	"go-api-kbt/internal/middleware"
	handler "go-api-kbt/internal/transport/http/handler"
	httpmiddleware "go-api-kbt/internal/transport/http/middleware"
	httputil "go-api-kbt/internal/transport/httputil"

	"github.com/go-chi/chi/v5"
	chim "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// RouterBuilder wires handlers and middleware together.
type RouterBuilder struct {
	userHandler     *handler.UserHandler
	eventHandler    *handler.EventHandler
	locationHandler *handler.LocationHandler
	medalerHandler  *handler.MedalerHandler
	activityHandler *handler.ActivityHandler
	authHandler     *handler.AuthHandler
	middlewares     []func(http.Handler) http.Handler
	cfg             *config.Config
	db              database.Database
	redis           *redis.Client
}

// NewRouterBuilder constructs a new router builder instance.
func NewRouterBuilder(userHandler *handler.UserHandler) *RouterBuilder {
	return &RouterBuilder{userHandler: userHandler}
}

// WithSystemDeps sets shared configuration and infrastructure dependencies used by the router.
func (b *RouterBuilder) WithSystemDeps(cfg *config.Config, db database.Database, redis *redis.Client) *RouterBuilder {
	b.cfg = cfg
	b.db = db
	b.redis = redis
	return b
}

// WithEventHandler attaches an event handler.
func (b *RouterBuilder) WithEventHandler(h *handler.EventHandler) *RouterBuilder {
	b.eventHandler = h
	return b
}

// WithLocationHandler attaches a location handler.
func (b *RouterBuilder) WithLocationHandler(h *handler.LocationHandler) *RouterBuilder {
	b.locationHandler = h
	return b
}

// WithMedalerHandler attaches a medaler handler.
func (b *RouterBuilder) WithMedalerHandler(h *handler.MedalerHandler) *RouterBuilder {
	b.medalerHandler = h
	return b
}

// WithActivityHandler attaches an activity handler.
func (b *RouterBuilder) WithActivityHandler(h *handler.ActivityHandler) *RouterBuilder {
	b.activityHandler = h
	return b
}

// WithAuthHandler attaches an auth handler.
func (b *RouterBuilder) WithAuthHandler(h *handler.AuthHandler) *RouterBuilder {
	b.authHandler = h
	return b
}

// WithMiddlewares sets additional middlewares to apply to the router.
func (b *RouterBuilder) WithMiddlewares(mw ...func(http.Handler) http.Handler) *RouterBuilder {
	b.middlewares = append(b.middlewares, mw...)
	return b
}

// Build returns a configured chi.Router instance.
func (b *RouterBuilder) Build() http.Handler {
	r := chi.NewRouter()

	// Built-in middlewares
	r.Use(chim.RequestID)
	r.Use(chim.RealIP)
	r.Use(middleware.NewStructuredLogger())
	r.Use(chim.Recoverer)
	r.Use(httpmiddleware.SecurityHeadersMiddleware)

	// CORS middleware
	corsOptions := cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Idempotency-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}
	if b.cfg != nil {
		corsOptions.AllowedOrigins = b.cfg.CORS.AllowedOrigins
	}
	corsMiddleware := cors.New(corsOptions)
	r.Use(corsMiddleware.Handler)

	rateLimitPerMin := 100
	if b.cfg != nil && b.cfg.RateLimit.PerMinute > 0 {
		rateLimitPerMin = b.cfg.RateLimit.PerMinute
	}
	r.Use(httprate.LimitByIP(rateLimitPerMin, time.Minute))

	for _, mw := range b.middlewares {
		r.Use(mw)
	}

	r.Route("/api/v1", func(r chi.Router) {
		if b.userHandler != nil {
			b.userHandler.RegisterRoutes(r)
		}
		if b.eventHandler != nil {
			b.eventHandler.RegisterRoutes(r)
		}
		if b.locationHandler != nil {
			b.locationHandler.RegisterRoutes(r)
		}
		if b.medalerHandler != nil {
			b.medalerHandler.RegisterRoutes(r)
		}
		if b.activityHandler != nil {
			b.activityHandler.RegisterRoutes(r)
		}
		if b.authHandler != nil {
			b.authHandler.RegisterRoutes(r)
		}
	})

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// Check database connectivity
		if b.db != nil {
			sqlDB, err := b.db.GormDB().DB()
			if err != nil {
				httputil.RespondError(w, http.StatusInternalServerError, "database connection error")
				return
			}
			if err := sqlDB.PingContext(r.Context()); err != nil {
				httputil.RespondError(w, http.StatusInternalServerError, "database ping failed")
				return
			}
		}

		// Check Redis connectivity
		if b.redis != nil {
			if err := b.redis.Ping(r.Context()).Err(); err != nil {
				httputil.RespondError(w, http.StatusInternalServerError, "redis ping failed")
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	RegisterDocsRoutes(r)

	// Serve Swagger UI
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	return r
}
