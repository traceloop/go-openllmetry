package traceloop

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type Tool struct {
	agent    *Agent
	ctx      context.Context
	Name     string       `json:"name"`
	Type     string       `json:"type"`
	Function ToolFunction `json:"function,omitempty"`
}

func (tool *Tool) End() {
	trace.SpanFromContext(tool.ctx).End()
}

func (tool *Tool) LogPrompt(prompt Prompt) LLMSpan {
	if tool.agent.workflow != nil {
		return tool.agent.sdk.LogPrompt(tool.ctx, prompt, &tool.agent.workflow.Attributes)
	}
	return tool.agent.sdk.LogPrompt(tool.ctx, prompt, nil)
}
