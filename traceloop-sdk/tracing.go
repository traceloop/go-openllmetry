package traceloop

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	otlp "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	otlpclient "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	stdoutexp "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
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

func newStdoutExporter(ctx context.Context) (*stdoutexp.Exporter, error) {
	return stdoutexp.New(stdoutexp.WithPrettyPrint())
}

func newTracerProvider(ctx context.Context, serviceName string, exp trace.SpanExporter) *trace.TracerProvider {
	r, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)

	if err != nil {
		panic(err)
	}

	return trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(r),
	)
}

func (instance *Traceloop) initTracer(ctx context.Context, serviceName string) error{
	fmt.Println("init tracer")

	// exp, err := newOtlpExporter(ctx, instance.config.BaseURL, instance.config.APIKey)
	exp, err := newStdoutExporter(ctx)
	if err != nil {
		panic(err)
	}
	
	tp := newTracerProvider(ctx, serviceName, exp)
	defer func() { _ = tp.Shutdown(ctx) }()

	otel.SetTracerProvider(tp)

	instance.tracer = tp.Tracer(serviceName)
	
	return nil
}
