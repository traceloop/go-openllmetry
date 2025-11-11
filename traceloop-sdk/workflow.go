package traceloop

import (
	"context"
	"fmt"

	"github.com/traceloop/go-openllmetry/traceloop-sdk/model"
	semconvai "github.com/traceloop/go-openllmetry/semconv-ai"
	"go.opentelemetry.io/otel/trace"
)

type Workflow struct {
	sdk        *Traceloop
	ctx        context.Context
	Attributes WorkflowAttributes `json:"workflow_attributes"`
}

func (instance *Traceloop) NewWorkflow(ctx context.Context, attrs WorkflowAttributes) *Workflow {
	wCtx, span := instance.getTracer().Start(ctx, fmt.Sprintf("%s.workflow", attrs.Name), trace.WithNewRoot())

	span.SetAttributes(
		semconvai.TraceloopWorkflowName.String(attrs.Name),
		semconvai.TraceloopSpanKind.String(string(model.SpanKindWorkflow)),
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

func (workflow *Workflow) LogPrompt(prompt Prompt) LLMSpan {
	return workflow.sdk.LogPrompt(workflow.ctx, prompt, workflow.Attributes)
}

func (workflow *Workflow) NewTask(name string) *Task {
	tCtx, span := workflow.sdk.getTracer().Start(workflow.ctx, fmt.Sprintf("%s.task", name))

	span.SetAttributes(
		semconvai.TraceloopWorkflowName.String(workflow.Attributes.Name),
		semconvai.TraceloopSpanKind.String(string(model.SpanKindTask)),
		semconvai.TraceloopEntityName.String(name),
	)

	return &Task{
		workflow: workflow,
		ctx:      tCtx,
		Name:     name,
	}
}

func (workflow *Workflow) NewAgent(name string) *Agent {
	aCtx, span := workflow.sdk.getTracer().Start(workflow.ctx, fmt.Sprintf("%s.agent", name))

	span.SetAttributes(
		semconvai.TraceloopWorkflowName.String(workflow.Attributes.Name),
		semconvai.TraceloopSpanKind.String(string(model.SpanKindAgent)),
		semconvai.TraceloopEntityName.String(name),
	)

	return &Agent{
		workflow:   workflow,
		ctx:        aCtx,
		Name:       name,
	}
}

