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
	Name       string `json:"name"`
}

func (agent *Agent) End() {
	trace.SpanFromContext(agent.ctx).End()
}

func (agent *Agent) LogPrompt(prompt Prompt) LLMSpan {
	if agent.workflow != nil {
		return agent.workflow.LogPrompt(prompt)
	}
	return agent.sdk.LogPrompt(agent.ctx, prompt, WorkflowAttributes{})
}

func (agent *Agent) NewTool(name string, toolType string, toolFunction ToolFunction) *Tool {
	toolCtx, span := agent.sdk.getTracer().Start(agent.ctx, fmt.Sprintf("%s.tool", name))
	span.SetAttributes(
		semconvai.LLMAgentName.String(agent.Name),
		semconvai.TraceloopSpanKind.String(string(model.SpanKindTool)),
		semconvai.TraceloopEntityName.String(name),
	)

	return &Tool{
		agent:    agent,
		ctx:      toolCtx,
		Name:     name,
		Type:     toolType,
		Function: toolFunction,
	}
}


