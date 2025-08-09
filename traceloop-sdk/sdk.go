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
	"github.com/traceloop/go-openllmetry/traceloop-sdk/dto"
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

func setMessagesAttribute(span apitrace.Span, prefix string, messages []Message) {
	for _, message := range messages {
		attrsPrefix := fmt.Sprintf("%s.%d", prefix, message.Index)
		span.SetAttributes(
			attribute.String(attrsPrefix+".content", message.Content),
			attribute.String(attrsPrefix+".role", message.Role),
		)

		if len(message.ToolCalls) > 0 {
			setToolCallsAttribute(span, attrsPrefix, message.ToolCalls)
		}
	}
}

// Overload for DTO messages to support backward compatibility
func setDTOMessagesAttribute(span apitrace.Span, prefix string, messages []dto.Message) {
	for _, message := range messages {
		attrsPrefix := fmt.Sprintf("%s.%d", prefix, message.Index)
		span.SetAttributes(
			attribute.String(attrsPrefix+".content", message.Content),
			attribute.String(attrsPrefix+".role", message.Role),
		)

		if len(message.ToolCalls) > 0 {
			setDTOMessageToolCallsAttribute(span, attrsPrefix, message.ToolCalls)
		}
	}
}



// Tool calling attribute helpers for new types
func setToolCallsAttribute(span apitrace.Span, messagePrefix string, toolCalls []ToolCall) {
	for i, toolCall := range toolCalls {
		toolCallPrefix := fmt.Sprintf("%s.tool_calls.%d", messagePrefix, i)
		span.SetAttributes(
			attribute.String(toolCallPrefix+".id", toolCall.ID),
			attribute.String(toolCallPrefix+".type", toolCall.Type),
			attribute.String(toolCallPrefix+".name", toolCall.Function.Name),
			attribute.String(toolCallPrefix+".arguments", toolCall.Function.Arguments),
		)
	}
}

func setDTOMessageToolCallsAttribute(span apitrace.Span, messagePrefix string, toolCalls []dto.ToolCall) {
	for i, toolCall := range toolCalls {
		toolCallPrefix := fmt.Sprintf("%s.tool_calls.%d", messagePrefix, i)
		span.SetAttributes(
			attribute.String(toolCallPrefix+".id", toolCall.ID),
			attribute.String(toolCallPrefix+".type", toolCall.Type),
			attribute.String(toolCallPrefix+".name", toolCall.Function.Name),
			attribute.String(toolCallPrefix+".arguments", toolCall.Function.Arguments),
		)
	}
}

func setDTOCompletionsAttribute(span apitrace.Span, messages []dto.Message) {
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
		
		span.SetAttributes(attrs...)
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
			} else {
				fmt.Printf("Failed to marshal tool parameters for %s: %v\n", tool.Function.Name, err)
			}
		}
	}
}

func setDTOToolsAttribute(span apitrace.Span, tools []dto.Tool) {
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
			} else {
				fmt.Printf("Failed to marshal tool parameters for %s: %v\n", tool.Function.Name, err)
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

// New workflow-based API
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

// Legacy DTO-based API for backward compatibility
func (instance *Traceloop) LogPromptLegacy(ctx context.Context, attrs dto.PromptLogAttributes) error {
	spanName := fmt.Sprintf("%s.%s", attrs.Prompt.Vendor, attrs.Prompt.Mode)
	
	// Calculate start time based on duration
	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(attrs.Duration) * time.Millisecond)
	
	// Create span with historical start time
	spanCtx, span := instance.getTracer().Start(
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

	setDTOMessagesAttribute(span, "llm.prompts", attrs.Prompt.Messages)
	setDTOCompletionsAttribute(span, attrs.Completion.Messages)
	setDTOToolsAttribute(span, attrs.Prompt.Tools)

	// End span with correct end time
	span.End(apitrace.WithTimestamp(endTime))

	_ = spanCtx // avoid unused variable
	return nil
}

func (instance *Traceloop) Shutdown(ctx context.Context) {
	if instance.tracerProvider != nil {
		instance.tracerProvider.Shutdown(ctx)
	}
}
