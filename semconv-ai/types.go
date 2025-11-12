package semconvai

type SpanKind string

const (
	SpanKindTool     SpanKind = "tool"
	SpanKindAgent    SpanKind = "agent"
	SpanKindTask     SpanKind = "task"
	SpanKindWorkflow SpanKind = "workflow"
)

// The variant that is active will be added to the trace.
type ABTest struct {
	VariantKeys map[string]bool `json:"variant_keys"`
}
