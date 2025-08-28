package traceloop

import (
	"context"
	"fmt"
	"os"
	"strings"

	otlp "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	otlpgrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otlphttp "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func newTraceloopExporter(ctx context.Context, config Config) (*otlp.Exporter, error) {
	headers := make(map[string]string)
	for k, v := range config.Headers {
		headers[k] = v
	}

	if config.APIKey != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", config.APIKey)
	}

	return otlp.New(
		ctx,
		otlphttp.NewClient(
			otlphttp.WithEndpointURL(config.BaseURL),
			otlphttp.WithHeaders(headers),
		),
	)
}

func parseHeaders(headers string) map[string]string {
	headersArr := strings.Split(headers, ",")
	headersMap := make(map[string]string)
	for _, header := range headersArr {
		h := strings.Split(header, "=")

		if len(h) == 2 {
			headersMap[h[0]] = h[1]
		}
	}

	return headersMap
}

func newGenericExporter(ctx context.Context) (*otlp.Exporter, error) {
	protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	headers := parseHeaders(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS"))

	// Default to http/protobuf
	if protocol == "" || protocol == "http/protobuf" {
		return otlp.New(
			ctx,
			otlphttp.NewClient(
				otlphttp.WithEndpoint(endpoint),
				otlphttp.WithHeaders(headers),
			),
		)
	} else if protocol == "grpc" {
		return otlp.New(
			ctx,
			otlpgrpc.NewClient(
				otlpgrpc.WithEndpoint(endpoint),
				otlpgrpc.WithHeaders(headers),
			),
		)
	} else {
		// Not supporting http/json for now
		return nil, fmt.Errorf("invalid OTLP exporter type")
	}
}

func newOtlpExporter(ctx context.Context, config Config) (*otlp.Exporter, error) {
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		return newTraceloopExporter(ctx, config)
	} else {
		return newGenericExporter(ctx)
	}
}

func resourceName(serviceName string) string {
	if serviceName != "" {
		return serviceName
	}

	envVar := os.Getenv("OTEL_SERVICE_NAME")
	if envVar != "" {
		return envVar
	}

	if len(os.Args) > 0 {
		return os.Args[0]
	}

	return "unknown_service"
}

func newTracerProvider(ctx context.Context, serviceName string, exp trace.SpanExporter) (*trace.TracerProvider, error) {
	r, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(resourceName(serviceName)),
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
	exp, err := newTraceloopExporter(ctx, instance.config)
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
