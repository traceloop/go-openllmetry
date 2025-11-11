package model

type SpanKind string

const (
	SpanKindTool SpanKind = "tool"
	SpanKindAgent SpanKind = "agent"
	SpanKindTask SpanKind = "task"
	SpanKindWorkflow SpanKind = "workflow"
)