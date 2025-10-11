package http

import (
	"net/http"
	"time"

	"go-api-kbt/internal/middleware"
	handler "go-api-kbt/internal/transport/http/handler"

	"github.com/go-chi/chi/v5"
	chim "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

// RouterBuilder wires handlers and middleware together.
type RouterBuilder struct {
	userHandler     *handler.UserHandler
	eventHandler    *handler.EventHandler
	locationHandler *handler.LocationHandler
	middlewares     []func(http.Handler) http.Handler
}

// NewRouterBuilder constructs a new router builder instance.
func NewRouterBuilder(userHandler *handler.UserHandler) *RouterBuilder {
	return &RouterBuilder{userHandler: userHandler}
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
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(httprate.LimitByIP(100, time.Minute))

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
	})

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Serve API docs (openapi + Swagger UI)
	RegisterDocsRoutes(r)

	return r
}
