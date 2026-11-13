package tracer

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

type Config struct {
	// ServiceName sets the service.name resource attribute.
	ServiceName string

	// Endpoint is the OTLP HTTP collector endpoint (e.g. "localhost:4318").
	Endpoint string

	// Insecure disables TLS — use HTTP instead of HTTPS.
	Insecure bool

	// SamplerRatio controls sampling probability (0.0-1.0).
	// 1.0 = sample everything, 0.01 = sample 1%.
	SamplerRatio float64

	// ExportTimeout is the maximum time to wait for an export to complete.
	ExportTimeout time.Duration
}

func New(ctx context.Context, cfg Config) (*sdktrace.TracerProvider, error) {
	opts := []otlptracehttp.Option{}
	if cfg.Endpoint != "" {
		opts = append(opts, otlptracehttp.WithEndpoint(cfg.Endpoint))
	}
	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if cfg.ExportTimeout > 0 {
		opts = append(opts, otlptracehttp.WithTimeout(cfg.ExportTimeout))
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create otlp exporter: %w", err)
	}

	var sampler sdktrace.Sampler
	switch {
	case cfg.SamplerRatio >= 1.0:
		sampler = sdktrace.AlwaysSample()
	case cfg.SamplerRatio <= 0.0:
		sampler = sdktrace.NeverSample()
	default:
		sampler = sdktrace.TraceIDRatioBased(cfg.SamplerRatio)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}
