package traceloop

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type Tool struct {
	agent    Agent
	ctx      context.Context
	Name     string       `json:"name"`
	Type     string       `json:"type"`
	Function ToolFunction `json:"function,omitempty"`
}

func (tool *Tool) End() {
	trace.SpanFromContext(tool.ctx).End()
}

func (tool *Tool) LogPrompt(prompt Prompt) LLMSpan {
	// Merge workflow and agent association properties
	contextAttrs := ContextAttributes{
		AssociationProperties: make(map[string]string),
	}

	// Start with workflow properties if available
	if tool.agent.workflow != nil {
		contextAttrs.WorkflowName = &tool.agent.workflow.Attributes.Name
		for key, value := range tool.agent.workflow.Attributes.AssociationProperties {
			contextAttrs.AssociationProperties[key] = value
		}
	}

	// Agent properties override workflow properties
	for key, value := range tool.agent.Attributes.AssociationProperties {
		contextAttrs.AssociationProperties[key] = value
	}

	return tool.agent.sdk.LogPrompt(tool.ctx, prompt, contextAttrs)
}
