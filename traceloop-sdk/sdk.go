package traceloop

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	semconvai "github.com/traceloop/go-openllmetry/semconv-ai"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/config"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/dto"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/model"
)

const PromptsPath = "/v1/traceloop/prompts"

type Traceloop struct {
    config            config.Config
    promptRegistry    model.PromptRegistry
	tracer    		  trace.Tracer
    http.Client
}

func NewClient(config config.Config) *Traceloop {
	return &Traceloop{
		config:         config,
		promptRegistry: make(model.PromptRegistry),
		Client:         http.Client{},
	}
}

func (instance *Traceloop) Initialize(ctx context.Context) {
	if instance.config.BaseURL == "" {
		baseUrl := os.Getenv("TRACELOOP_BASE_URL")
		if baseUrl == "" {		
			instance.config.BaseURL = "https://api.traceloop.com"
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

	instance.initTracer(ctx, "GoExampleService")
	instance.pollPrompts()
}

func (instance *Traceloop) LogPrompt(ctx context.Context, prompt dto.Prompt, completion dto.Completion, traceloopAttrs dto.TraceloopAttributes) error {
	spanName := fmt.Sprintf("%s.%s", prompt.Vendor, prompt.Mode)
	_, span := instance.tracer.Start(ctx, spanName)
	
	span.SetAttributes(
		semconvai.LLMVendor.String(prompt.Vendor),
		semconvai.LLMRequestModel.String(prompt.Model),
		semconvai.LLMRequestType.String(prompt.Mode),
		semconvai.LLMResponseModel.String(completion.Model),
		semconvai.TraceloopWorkflowName.String(traceloopAttrs.WorkflowName),
		semconvai.TraceloopEntityName.String(traceloopAttrs.EntityName),
	)

	for _, message := range prompt.Messages {
		attrsPrefix := fmt.Sprintf("llm.prompts.%d", message.Index)
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

	for _, message := range completion.Messages {
		attrsPrefix := fmt.Sprintf("llm.completions.%d", message.Index)
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

	defer span.End()

	return nil
}
