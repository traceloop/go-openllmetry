package traceloop

import (
	"context"

	semconvai "github.com/traceloop/go-openllmetry/semconv-ai"
	apitrace "go.opentelemetry.io/otel/trace"
)

type LLMSpan struct {
	span apitrace.Span
}

func (llmSpan *LLMSpan) LogPrompt(ctx context.Context, prompt Prompt) {
	llmSpan.span.SetAttributes(
		semconvai.LLMRequestModel.String(prompt.Model),
		semconvai.LLMRequestType.String(prompt.Mode),
	)

	setMessagesAttribute(llmSpan.span, "llm.prompts", prompt.Messages)
}

func (llmSpan *LLMSpan) LogCompletion(ctx context.Context, completion Completion, usage Usage) {
	llmSpan.span.SetAttributes(
		semconvai.LLMResponseModel.String(completion.Model),
		semconvai.LLMUsageTotalTokens.Int(usage.TotalTokens),
		semconvai.LLMUsageCompletionTokens.Int(usage.CompletionTokens),
		semconvai.LLMUsagePromptTokens.Int(usage.PromptTokens),
	)

	setMessagesAttribute(llmSpan.span, "llm.completions", completion.Messages)

	defer llmSpan.span.End()
}
