package traceloop

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type Task struct {
	workflow *Workflow
	ctx      context.Context
	Name     string `json:"name"`
}

func (task *Task) End() {
	trace.SpanFromContext(task.ctx).End()
}

func (task *Task) LogPrompt(prompt Prompt) (LLMSpan, error) {
	return task.workflow.sdk.LogPrompt(task.ctx, prompt, task.workflow.Attributes)
}

func (task *Task) LogAgent(attrs AgentAttributes) LLMSpan {
	return task.workflow.sdk.LogAgent(task.ctx, attrs, task.workflow.Attributes)
}

func (task *Task) LogToolCall(attrs ToolCallAttributes) LLMSpan {
	return task.workflow.sdk.LogToolCall(task.ctx, attrs, task.workflow.Attributes)
}