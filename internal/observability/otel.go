package observability

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go-api-kbt/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Provider wraps a TracerProvider to allow clean shutdown.
type Provider struct {
	tracerProvider *sdktrace.TracerProvider
}

// Setup initialises OpenTelemetry tracing with a stdout or OTLP exporter.
func Setup(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*Provider, error) {
	if !cfg.Telemetry.EnableTracing {
		logger.Info("tracing disabled")
		return &Provider{}, nil
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

	var exporter sdktrace.SpanExporter
	switch cfg.Telemetry.Exporter {
	case "otlp":
		logger.Info("using OTLP trace exporter", slog.String("endpoint", cfg.Telemetry.CollectorEndpoint))
		conn, err := grpc.DialContext(ctx, cfg.Telemetry.CollectorEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err != nil {
			return nil, fmt.Errorf("create OTLP gRPC connection: %w", err)
		}
		otlpExporter, err := otlptrace.New(ctx, otlptracegrpc.WithGRPCConn(conn))
		if err != nil {
			return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
		}
		exporter = otlpExporter
	case "stdout":
		fallthrough
	default:
		logger.Info("using stdout trace exporter")
		stdoutExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("create stdout exporter: %w", err)
		}
		exporter = stdoutExporter
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(5*time.Second)),
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
