package observability

import (
	"context"
	"fmt"
	"log/slog"

	"go-api-kbt/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Provider wraps a TracerProvider to allow clean shutdown.
type Provider struct {
	tracerProvider *sdktrace.TracerProvider
}

// Setup initialises OpenTelemetry tracing with a stdout exporter for local use.
func Setup(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*Provider, error) {
	if !cfg.Telemetry.EnableTracing {
		logger.Info("tracing disabled")
		return &Provider{}, nil
	}

	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("create stdout exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.Telemetry.ServiceName),
			semconv.ServiceVersionKey.String(cfg.App.Version),
			semconv.DeploymentEnvironmentKey.String(cfg.App.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return &Provider{tracerProvider: tp}, nil
}

// Shutdown gracefully stops the tracer provider.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p == nil || p.tracerProvider == nil {
		return nil
	}
	return p.tracerProvider.Shutdown(ctx)
}
