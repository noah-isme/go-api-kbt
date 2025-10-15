package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"go-api-kbt/internal/config"
	serviceAuth "go-api-kbt/internal/service/auth"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware creates a new span for each request and propagates trace context.
func TracingMiddleware(cfg *config.TelemetryConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.EnableTracing {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			// Extract trace context from request headers
			propagator := propagation.TraceContext{}
			ctx = propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))

			// Start a new span
			tracer := otel.Tracer(cfg.ServiceName)
			ctx, span := tracer.Start(ctx, r.URL.Path, trace.WithSpanKind(trace.SpanKindServer))
			defer span.End()

			// Add common span attributes
			routePattern := r.URL.Path
			if routeContext := chi.RouteContext(ctx); routeContext != nil {
				if pattern := routeContext.RoutePattern(); pattern != "" {
					routePattern = pattern
				}
			}

			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", routePattern),
				attribute.String("http.target", r.URL.Path),
				attribute.String("http.flavor", fmt.Sprintf("1.%d", r.ProtoMajor)),
				attribute.String("net.host.name", r.Host),
				attribute.String("user_agent.original", r.UserAgent()),
			)

			// Add user ID to span attributes if authenticated
			claims, ok := ctx.Value(ContextKeyUser).(*serviceAuth.Claims)
			if ok {
				span.SetAttributes(attribute.Int64("user.id", int64(claims.UserID)))
			}

			// Inject trace context into response headers
			propagator.Inject(ctx, propagation.HeaderCarrier(w.Header()))

			// Add trace and span ID to logger
			slog.Default().With(
				slog.String("trace_id", span.SpanContext().TraceID().String()),
				slog.String("span_id", span.SpanContext().SpanID().String()),
			)

			// Call the next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
