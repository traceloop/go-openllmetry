package traceloop

import (
	"context"
	"fmt"

	semconvai "github.com/traceloop/go-openllmetry/semconv-ai"
	"go.opentelemetry.io/otel/trace"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/model"
)

type Agent struct {
	sdk        *Traceloop
	workflow   *Workflow
	ctx        context.Context
	Attributes AgentAttributes `json:"agent_attributes"`
}

func (agent *Agent) End() {
	trace.SpanFromContext(agent.ctx).End()
}

func (agent *Agent) LogPrompt(prompt Prompt) LLMSpan {
	// Merge workflow and agent association properties
	contextAttrs := ContextAttributes{
		AssociationProperties: make(map[string]string),
	}

	// Start with workflow properties if available
	if agent.workflow != nil {
		contextAttrs.WorkflowName = &agent.workflow.Attributes.Name
		for key, value := range agent.workflow.Attributes.AssociationProperties {
			contextAttrs.AssociationProperties[key] = value
		}
	}

	// Agent properties override workflow properties
	for key, value := range agent.Attributes.AssociationProperties {
		contextAttrs.AssociationProperties[key] = value
	}

	return agent.sdk.LogPrompt(agent.ctx, prompt, contextAttrs)
}

func (agent *Agent) NewTool(name string, toolType string, toolFunction ToolFunction) *Tool {
	toolCtx, span := agent.sdk.getTracer().Start(agent.ctx, fmt.Sprintf("%s.tool", name))
	span.SetAttributes(
		semconvai.LLMAgentName.String(agent.Attributes.Name),
		semconvai.TraceloopSpanKind.String(string(model.SpanKindTool)),
		semconvai.TraceloopEntityName.String(name),
	)

	return &Tool{
		agent:    *agent,
		ctx:      toolCtx,
		Name:     name,
		Type:     toolType,
		Function: toolFunction,
	}
}


