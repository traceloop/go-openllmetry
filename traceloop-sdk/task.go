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

func (task *Task) LogPrompt(prompt Prompt) LLMSpan {
	contextAttrs := ContextAttributes{
		WorkflowName: &task.workflow.Attributes.Name,
		AssociationProperties: task.workflow.Attributes.AssociationProperties,
	}
	return task.workflow.sdk.LogPrompt(task.ctx, prompt, contextAttrs)
}
