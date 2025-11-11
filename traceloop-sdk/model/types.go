package model

type SpanKind string

const (
	SpanKindTool SpanKind = "tool"
	SpanKindAgent SpanKind = "agent"
	SpanKindTask SpanKind = "task"
	SpanKindWorkflow SpanKind = "workflow"
)

// The varient that is active will be added to the trace.
type ABTest struct {
	VarientsKeys map[string]bool `json:"varients_keys"`
}
