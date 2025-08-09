module github.com/traceloop/go-openllmetry

go 1.21

require (
	github.com/traceloop/go-openllmetry/traceloop-sdk v0.0.0-00010101000000-000000000000
	github.com/traceloop/go-openllmetry/semconv-ai v0.0.0-00010101000000-000000000000
)

replace github.com/traceloop/go-openllmetry/traceloop-sdk => ./traceloop-sdk

replace github.com/traceloop/go-openllmetry/semconv-ai => ./semconv-ai
