package traceloop

import (
	"context"
	"fmt"

	otlp "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	otlpclient "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func newOtlpExporter(ctx context.Context, endpoint string, apiKey string) (*otlp.Exporter, error) {
	return otlp.New(
		ctx,
		otlpclient.NewClient(
			otlpclient.WithEndpoint(endpoint),
			otlpclient.WithHeaders(map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", apiKey),
			}),
		),
	)
}

func newTracerProvider(ctx context.Context, serviceName string, exp trace.SpanExporter) (*trace.TracerProvider, error) {
	r, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)

	if err != nil {
		return nil, err
	}

	return trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(r),
	), nil
}

func (instance *Traceloop) initTracer(ctx context.Context, serviceName string) error {
	exp, err := newOtlpExporter(ctx, instance.config.BaseURL, instance.config.APIKey)
	if err != nil {
		return fmt.Errorf("create otlp exporter: %w", err)
	}
	
	tp, err := newTracerProvider(ctx, serviceName, exp)
	if err != nil {
		return fmt.Errorf("create tracer provider: %w", err)
	}

	instance.tracerProvider = tp

	return nil
}
