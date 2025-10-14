package logger

import (
	"context"
	"io"
	"log/slog"
	"regexp"

	"go.opentelemetry.io/otel/trace"
)

// PII patterns to scrub
var piiPatterns = []*regexp.Regexp{
	regexp.MustCompile(`"email":"[^"]+"`),
	regexp.MustCompile(`"password":"[^"]+"`),
	// Add more patterns as needed
}

// NewLogger creates a new slog logger with PII scrubbing and trace context.
func NewLogger(out io.Writer, level slog.Level) *slog.Logger {
	handler := &piiScrubbingHandler{slog.NewJSONHandler(out, &slog.HandlerOptions{Level: level})}
	return slog.New(handler)
}

type piiScrubbingHandler struct {
	slog.Handler
}

func (h *piiScrubbingHandler) Handle(ctx context.Context, r slog.Record) error {
	// Scrub PII from the message and attributes
	for _, p := range piiPatterns {
		r.Message = p.ReplaceAllString(r.Message, `"<REDACTED>"`)
	}

	// Add trace and span ID if available
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
		)
	}

	return h.Handler.Handle(ctx, r)
}

func (h *piiScrubbingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &piiScrubbingHandler{h.Handler.WithAttrs(attrs)}
}

func (h *piiScrubbingHandler) WithGroup(name string) slog.Handler {
	return &piiScrubbingHandler{h.Handler.WithGroup(name)}
}
