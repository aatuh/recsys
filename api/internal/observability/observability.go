package observability

import (
	"context"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"recsys/internal/config"
	httpmetrics "recsys/internal/http/middleware"
)

const serviceName = "recsys-api"

// Provider wires metrics and tracing instrumentation.
type Provider struct {
	MetricsMiddleware func(http.Handler) http.Handler
	MetricsHandler    http.Handler
	TraceMiddleware   func(http.Handler) http.Handler
	Shutdown          func(context.Context) error
}

// Setup initializes observability components based on configuration.
func Setup(ctx context.Context, cfg config.ObservabilityConfig, logger *zap.Logger) (*Provider, error) {
	prov := &Provider{}
	var shutdowns []func(context.Context) error

	if cfg.MetricsEnabled {
		metrics := httpmetrics.NewHTTPMetrics()
		metrics.SetMetricsPath(cfg.MetricsPath)
		prov.MetricsMiddleware = metrics.Middleware
		prov.MetricsHandler = metrics.Handler()
	}

	if cfg.TracingEnabled {
		if logger != nil {
			logger.Info("tracing enabled", zap.String("exporter", cfg.TraceExporter))
		}
		exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, err
		}
		tp := tracesdk.NewTracerProvider(
			tracesdk.WithBatcher(exp),
			tracesdk.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
			)),
		)
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.TraceContext{})
		prov.TraceMiddleware = otelhttp.NewMiddleware(serviceName, otelhttp.WithTracerProvider(tp))
		shutdowns = append(shutdowns, tp.Shutdown)
	} else {
		if logger != nil {
			logger.Debug("tracing disabled")
		}
		tp := trace.NewNoopTracerProvider()
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.TraceContext{})
		prov.TraceMiddleware = otelhttp.NewMiddleware(serviceName, otelhttp.WithTracerProvider(tp))
	}

	if len(shutdowns) > 0 {
		prov.Shutdown = func(ctx context.Context) error {
			var first error
			for _, fn := range shutdowns {
				if err := fn(ctx); err != nil && first == nil {
					first = err
				}
			}
			return first
		}
	}

	return prov, nil
}
