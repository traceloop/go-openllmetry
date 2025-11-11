package model

type SpanKind string

const (
	SpanKindTool     SpanKind = "tool"
	SpanKindAgent    SpanKind = "agent"
	SpanKindTask     SpanKind = "task"
	SpanKindWorkflow SpanKind = "workflow"
)

// The varient that is active will be added to the trace.
type ABTest struct {
	VarientKeys map[string]bool `json:"varient_keys"`
}
