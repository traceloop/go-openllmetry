package traceloop

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	apitrace "go.opentelemetry.io/otel/trace"

	semconvai "github.com/traceloop/go-openllmetry/semconv-ai"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/model"
)

const PromptsPath = "/v1/traceloop/prompts"

type Traceloop struct {
	config         Config
	promptRegistry model.PromptRegistry
	registryMutex  sync.RWMutex
	tracerProvider *trace.TracerProvider
	http.Client
}

type LLMSpan struct {
	span apitrace.Span
}

func NewClient(ctx context.Context, config Config) (*Traceloop, error) {
	instance := Traceloop{
		config:         config,
		promptRegistry: make(model.PromptRegistry),
		Client:         http.Client{},
	}

	err := instance.initialize(ctx)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

func (instance *Traceloop) initialize(ctx context.Context) error {
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

	log.Printf("Traceloop %s SDK initialized. Connecting to %s\n", Version(), instance.config.BaseURL)

	instance.pollPrompts()
	err := instance.initTracer(ctx, instance.config.ServiceName)
	if err != nil {
		return err
	}

	return nil
}

func setMessageToolCallsAttribute(span apitrace.Span, messagePrefix string, toolCalls []ToolCall) {
	for i, toolCall := range toolCalls {
		toolCallPrefix := fmt.Sprintf("%s.tool_calls.%d", messagePrefix, i)
		span.SetAttributes(
			attribute.String(toolCallPrefix+".id", toolCall.ID),
			attribute.String(toolCallPrefix+".name", toolCall.Function.Name),
			attribute.String(toolCallPrefix+".arguments", toolCall.Function.Arguments),
		)
	}
}

func setMessagesAttribute(span apitrace.Span, prefix string, messages []Message) {
	for _, message := range messages {
		attrsPrefix := fmt.Sprintf("%s.%d", prefix, message.Index)
		span.SetAttributes(
			attribute.String(attrsPrefix+".content", message.Content),
			attribute.String(attrsPrefix+".role", message.Role),
		)

		if len(message.ToolCalls) > 0 {
			setMessageToolCallsAttribute(span, attrsPrefix, message.ToolCalls)
		}
	}
}

func setToolsAttribute(span apitrace.Span, tools []Tool) {
	if len(tools) == 0 {
		return
	}

	for i, tool := range tools {
		prefix := fmt.Sprintf("%s.%d", string(semconvai.LLMRequestFunctions), i)
		span.SetAttributes(
			attribute.String(prefix+".name", tool.Function.Name),
			attribute.String(prefix+".description", tool.Function.Description),
		)

		if tool.Function.Parameters != nil {
			parametersJSON, err := json.Marshal(tool.Function.Parameters)
			if err == nil {
				span.SetAttributes(
					attribute.String(prefix+".parameters", string(parametersJSON)),
				)
			}
		}
	}
}

func (instance *Traceloop) tracerName() string {
	if instance.config.TracerName != "" {
		return instance.config.TracerName
	} else {
		return "traceloop.tracer"
	}
}

func (instance *Traceloop) getTracer() apitrace.Tracer {
	return (*instance.tracerProvider).Tracer(instance.tracerName())
}

func (instance *Traceloop) LogPrompt(ctx context.Context, prompt Prompt, workflowAttrs WorkflowAttributes) (LLMSpan, error) {
	spanName := fmt.Sprintf("%s.%s", prompt.Vendor, prompt.Mode)
	_, span := instance.getTracer().Start(ctx, spanName)

	span.SetAttributes(
		semconvai.LLMVendor.String(prompt.Vendor),
		semconvai.LLMRequestModel.String(prompt.Model),
		semconvai.LLMRequestType.String(prompt.Mode),
		semconvai.TraceloopWorkflowName.String(workflowAttrs.Name),
	)

	setMessagesAttribute(span, "llm.prompts", prompt.Messages)
	setToolsAttribute(span, prompt.Tools)

	return LLMSpan{
		span: span,
	}, nil
}

func (llmSpan *LLMSpan) LogCompletion(ctx context.Context, completion Completion, usage Usage) error {
	llmSpan.span.SetAttributes(
		semconvai.LLMResponseModel.String(completion.Model),
		semconvai.LLMUsageTotalTokens.Int(usage.TotalTokens),
		semconvai.LLMUsageCompletionTokens.Int(usage.CompletionTokens),
		semconvai.LLMUsagePromptTokens.Int(usage.PromptTokens),
	)

	setMessagesAttribute(llmSpan.span, "llm.completions", completion.Messages)

	defer llmSpan.span.End()

	return nil
}

func (instance *Traceloop) Shutdown(ctx context.Context) {
	if instance.tracerProvider != nil {
		instance.tracerProvider.Shutdown(ctx)
	}
}
