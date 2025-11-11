package traceloop

import (
	"context"
	"fmt"

	semconvai "github.com/traceloop/go-openllmetry/semconv-ai"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/model"
	"go.opentelemetry.io/otel/attribute"
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

	if attrs.ABTest != nil {
		for key, activeVariant := range attrs.ABTest.VariantKeys {
			if activeVariant {
				span.SetAttributes(attribute.String("traceloop.association.properties.ab_testing_variant", key))
				break
			}
		}
	}

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
	contextAttrs := ContextAttributes{
		WorkflowName:          &workflow.Attributes.Name,
		AssociationProperties: workflow.Attributes.AssociationProperties,
	}
	return workflow.sdk.LogPrompt(workflow.ctx, prompt, contextAttrs)
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

func (workflow *Workflow) NewAgent(name string, associationProperties map[string]string) *Agent {
	aCtx, span := workflow.sdk.getTracer().Start(workflow.ctx, fmt.Sprintf("%s.agent", name))

	attrs := []attribute.KeyValue{
		semconvai.TraceloopWorkflowName.String(workflow.Attributes.Name),
		semconvai.TraceloopSpanKind.String(string(model.SpanKindAgent)),
		semconvai.TraceloopEntityName.String(name),
	}

	if workflow.Attributes.ABTest != nil {
		for key, activeVariant := range workflow.Attributes.ABTest.VariantKeys {
			if activeVariant {
				associationProperties["ab_testing_variant"] = key
			}
		}
	}
	// Add agent-specific association properties to the span
	for key, value := range associationProperties {
		attrs = append(attrs, attribute.String("traceloop.association.properties."+key, value))
	}

	span.SetAttributes(attrs...)

	return &Agent{
		sdk:      workflow.sdk,
		workflow: workflow,
		ctx:      aCtx,
		Attributes: AgentAttributes{
			Name:                  name,
			AssociationProperties: associationProperties,
		},
	}
}
