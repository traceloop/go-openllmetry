package traceloop

import (
	"context"
	"fmt"

	semconvai "github.com/traceloop/go-openllmetry/semconv-ai"
	"go.opentelemetry.io/otel/trace"
)

type Workflow struct {
	sdk        *Traceloop
	ctx        context.Context
	Attributes WorkflowAttributes `json:"workflow_attributes"`
}

type Task struct {
	workflow *Workflow
	ctx      context.Context
	Name     string `json:"name"`
}

func (instance *Traceloop) NewWorkflow(ctx context.Context, attrs WorkflowAttributes) *Workflow {
	wCtx, span := instance.getTracer().Start(ctx, fmt.Sprintf("%s.workflow", attrs.Name), trace.WithNewRoot())

	span.SetAttributes(
		semconvai.TraceloopWorkflowName.String(attrs.Name),
		semconvai.TraceloopSpanKind.String("workflow"),
		semconvai.TraceloopEntityName.String(attrs.Name),
	)

	return &Workflow{
		sdk:        instance,
		ctx:        wCtx,
		Attributes: attrs,
	}
}

func (workflow *Workflow) End() {
	trace.SpanFromContext(workflow.ctx).End()
}

func (workflow *Workflow) LogPrompt(prompt Prompt) (LLMSpan, error) {
	return workflow.sdk.LogPrompt(workflow.ctx, prompt, workflow.Attributes)
}

func (workflow *Workflow) NewTask(name string) *Task {
	tCtx, span := workflow.sdk.getTracer().Start(workflow.ctx, fmt.Sprintf("%s.task", name))

	span.SetAttributes(
		semconvai.TraceloopWorkflowName.String(workflow.Attributes.Name),
		semconvai.TraceloopSpanKind.String("task"),
		semconvai.TraceloopEntityName.String(name),
	)

	return &Task{
		workflow: workflow,
		ctx:      tCtx,
		Name:     name,
	}
}

func (task *Task) End() {
	trace.SpanFromContext(task.ctx).End()
}

func (task *Task) LogPrompt(prompt Prompt) (LLMSpan, error) {
	return task.workflow.sdk.LogPrompt(task.ctx, prompt, task.workflow.Attributes)
}
