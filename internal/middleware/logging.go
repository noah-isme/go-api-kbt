package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	chim "github.com/go-chi/chi/v5/middleware"
)

// loggerKey is used to store the logger in the request context.
type loggerKey struct{}

// LoggerFromContext retrieves the request scoped logger.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// NewStructuredLogger returns a middleware that enriches the request context
// with a structured logger and emits an access log once the request completes.
func NewStructuredLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := chim.GetReqID(r.Context())
			logger := slog.Default().With(
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)

			ctx := context.WithValue(r.Context(), loggerKey{}, logger)
			ww := chim.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			next.ServeHTTP(ww, r.WithContext(ctx))

			logger.Info("request completed",
				slog.Int("status", ww.Status()),
				slog.Int("bytes", ww.BytesWritten()),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}
