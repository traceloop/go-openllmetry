package traceloop

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	apitrace "go.opentelemetry.io/otel/trace"

	semconvai "github.com/traceloop/go-openllmetry/semconv-ai"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/config"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/dto"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/model"
)

const PromptsPath = "/v1/traceloop/prompts"

type Traceloop struct {
    config            config.Config
    promptRegistry    model.PromptRegistry
	tracerProvider    *trace.TracerProvider
    http.Client
}

func NewClient(ctx context.Context, config config.Config) *Traceloop {
	instance := Traceloop{
		config:         config,
		promptRegistry: make(model.PromptRegistry),
		Client:         http.Client{},
	}

	instance.initialize(ctx)

	return &instance
}

func (instance *Traceloop) initialize(ctx context.Context) {
	if instance.config.BaseURL == "" {
		baseUrl := os.Getenv("TRACELOOP_BASE_URL")
		if baseUrl == "" {		
			instance.config.BaseURL = "api.traceloop.com"
		} else {
			instance.config.BaseURL = baseUrl
		}
	}

	if instance.config.PollingInterval == 0 {
		pollingInterval := os.Getenv("TRACELOOP_SECONDS_POLLING_INTERVAL")
		if pollingInterval == "" {
			instance.config.PollingInterval = 5 * time.Second
		} else {
			instance.config.PollingInterval, _ = time.ParseDuration(pollingInterval)
		}
	}

	fmt.Printf("Traceloop %s SDK initialized. Connecting to %s\n", instance.GetVersion(), instance.config.BaseURL)

	instance.pollPrompts()
	instance.initTracer(ctx, "GoExampleService")
}

func setMessagesAttribute(span apitrace.Span, prefix string, messages []dto.Message) {
	for _, message := range messages {
		attrsPrefix := fmt.Sprintf("%s.%d", prefix, message.Index)
		span.SetAttributes(
			attribute.KeyValue{
				Key:   attribute.Key(attrsPrefix + ".content"),
				Value: attribute.StringValue(message.Content),
			},
			attribute.KeyValue{
				Key:   attribute.Key(attrsPrefix + ".role"),
				Value: attribute.StringValue(message.Role),
			},
		)
	}
}

func (instance *Traceloop) LogPrompt(ctx context.Context, attrs dto.PromptLogAttributes) error {
	spanName := fmt.Sprintf("%s.%s", attrs.Prompt.Vendor, attrs.Prompt.Mode)
	_, span := (*instance.tracerProvider).Tracer(os.Args[0]).Start(ctx, spanName)
	
	span.SetAttributes(
		semconvai.LLMVendor.String(attrs.Prompt.Vendor),
		semconvai.LLMRequestModel.String(attrs.Prompt.Model),
		semconvai.LLMRequestType.String(attrs.Prompt.Mode),
		semconvai.LLMResponseModel.String(attrs.Completion.Model),
		semconvai.LLMUsageTotalTokens.Int(attrs.Usage.TotalTokens),
		semconvai.LLMUsageCompletionTokens.Int(attrs.Usage.CompletionTokens),
		semconvai.LLMUsagePromptTokens.Int(attrs.Usage.PromptTokens),
		semconvai.TraceloopWorkflowName.String(attrs.Traceloop.WorkflowName),
		semconvai.TraceloopEntityName.String(attrs.Traceloop.EntityName),
	)

	setMessagesAttribute(span, "llm.prompts", attrs.Prompt.Messages)
	setMessagesAttribute(span, "llm.completions", attrs.Completion.Messages)

	defer span.End()

	return nil
}

func (instance *Traceloop) Shutdown(ctx context.Context) {
	instance.tracerProvider.Shutdown(ctx)
}
