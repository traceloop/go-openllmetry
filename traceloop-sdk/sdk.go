package traceloop

import (
	"context"
	"encoding/json"
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

	instance.pollPrompts()
	instance.initTracer(ctx, "GoExampleService")
}

func setMessagesAttribute(span *apitrace.Span, prefix string, messages []dto.Message) {
	for _, message := range messages {
		attrsPrefix := fmt.Sprintf("%s.%d", prefix, message.Index)
		(*span).SetAttributes(
			attribute.KeyValue{
				Key:   attribute.Key(attrsPrefix + ".content"),
				Value: attribute.StringValue(message.Content),
			},
			attribute.KeyValue{
				Key:   attribute.Key(attrsPrefix + ".role"),
				Value: attribute.StringValue(message.Role),
			},
		)

		if len(message.ToolCalls) > 0 {
			setMessageToolCallsAttribute(span, attrsPrefix, message.ToolCalls)
		}
	}
}

func setMessageToolCallsAttribute(span *apitrace.Span, messagePrefix string, toolCalls []dto.ToolCall) {
	for i, toolCall := range toolCalls {
		toolCallPrefix := fmt.Sprintf("%s.tool_calls.%d", messagePrefix, i)
		(*span).SetAttributes(
			attribute.KeyValue{
				Key:   attribute.Key(toolCallPrefix + ".id"),
				Value: attribute.StringValue(toolCall.ID),
			},
			attribute.KeyValue{
				Key:   attribute.Key(toolCallPrefix + ".type"),
				Value: attribute.StringValue(toolCall.Type),
			},
			attribute.KeyValue{
				Key:   attribute.Key(toolCallPrefix + ".name"),
				Value: attribute.StringValue(toolCall.Function.Name),
			},
			attribute.KeyValue{
				Key:   attribute.Key(toolCallPrefix + ".arguments"),
				Value: attribute.StringValue(toolCall.Function.Arguments),
			},
		)
	}
}

func setCompletionsAttribute(span *apitrace.Span, messages []dto.Message) {
	for _, message := range messages {
		prefix := fmt.Sprintf("llm.completions.%d", message.Index)
		attrs := []attribute.KeyValue{
			{Key: attribute.Key(prefix + ".role"), Value: attribute.StringValue(message.Role)},
			{Key: attribute.Key(prefix + ".content"), Value: attribute.StringValue(message.Content)},
		}
		
		// Set tool calls attributes exactly like Python version
		for i, toolCall := range message.ToolCalls {
			toolCallPrefix := fmt.Sprintf("%s.tool_calls.%d", prefix, i)
			attrs = append(attrs, 
				attribute.KeyValue{Key: attribute.Key(toolCallPrefix + ".id"), Value: attribute.StringValue(toolCall.ID)},
				attribute.KeyValue{Key: attribute.Key(toolCallPrefix + ".type"), Value: attribute.StringValue(toolCall.Type)},
				attribute.KeyValue{Key: attribute.Key(toolCallPrefix + ".name"), Value: attribute.StringValue(toolCall.Function.Name)},
				attribute.KeyValue{Key: attribute.Key(toolCallPrefix + ".arguments"), Value: attribute.StringValue(toolCall.Function.Arguments)},
			)
		}
		
		(*span).SetAttributes(attrs...)
	}
}

func setToolsAttribute(span *apitrace.Span, tools []dto.Tool) {
	if len(tools) == 0 {
		return
	}

	for i, tool := range tools {
		prefix := fmt.Sprintf("%s.%d", string(semconvai.LLMRequestFunctions), i)
		(*span).SetAttributes(
			attribute.KeyValue{
				Key:   attribute.Key(prefix + ".name"),
				Value: attribute.StringValue(tool.Function.Name),
			},
			attribute.KeyValue{
				Key:   attribute.Key(prefix + ".description"),
				Value: attribute.StringValue(tool.Function.Description),
			},
		)

		if tool.Function.Parameters != nil {
			parametersJSON, err := json.Marshal(tool.Function.Parameters)
			if err == nil {
				(*span).SetAttributes(
					attribute.KeyValue{
						Key:   attribute.Key(prefix + ".parameters"),
						Value: attribute.StringValue(string(parametersJSON)),
					},
				)
			} else {
				fmt.Printf("Failed to marshal tool parameters for %s: %v\n", tool.Function.Name, err)
			}
		}
	}
}

func (instance *Traceloop) LogPrompt(ctx context.Context, attrs dto.PromptLogAttributes) error {
	spanName := fmt.Sprintf("%s.%s", attrs.Prompt.Vendor, attrs.Prompt.Mode)
	
	// Calculate start time based on duration
	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(attrs.Duration) * time.Millisecond)
	
	// Create span with historical start time
	spanCtx, span := (*instance.tracerProvider).Tracer(os.Args[0]).Start(
		ctx, 
		spanName,
		apitrace.WithTimestamp(startTime),
	)
	
	// Serialize messages to JSON for main attributes (both needed)
	promptsJSON, _ := json.Marshal(attrs.Prompt.Messages)
	completionsJSON, _ := json.Marshal(attrs.Completion.Messages)
	
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
		semconvai.LLMPrompts.String(string(promptsJSON)),
		semconvai.LLMCompletions.String(string(completionsJSON)),
	)

	setMessagesAttribute(&span, "llm.prompts", attrs.Prompt.Messages)
	setCompletionsAttribute(&span, attrs.Completion.Messages)
	setToolsAttribute(&span, attrs.Prompt.Tools)

	// End span with correct end time
	span.End(apitrace.WithTimestamp(endTime))

	_ = spanCtx // avoid unused variable

	return nil
}

func (instance *Traceloop) Shutdown(ctx context.Context) {
	instance.tracerProvider.Shutdown(ctx)
}
