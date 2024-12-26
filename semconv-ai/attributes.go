package semconvai

import "go.opentelemetry.io/otel/attribute"

const (
	// LLM
	LLMVendor                = attribute.Key("llm.vendor")
	LLMRequestType           = attribute.Key("llm.request.type")
	LLMRequestModel          = attribute.Key("llm.request.model")
	LLMResponseModel         = attribute.Key("llm.response.model")
	LLMRequestMaxTokens      = attribute.Key("llm.request.max_tokens")
	LLMUsageTotalTokens      = attribute.Key("llm.usage.total_tokens")
	LLMUsageCompletionTokens = attribute.Key("llm.usage.completion_tokens")
	LLMUsagePromptTokens     = attribute.Key("llm.usage.prompt_tokens")
	LLMTemperature           = attribute.Key("llm.temperature")
	LLMUser                  = attribute.Key("llm.user")
	LLMHeaders               = attribute.Key("llm.headers")
	LLMTopP                  = attribute.Key("llm.top_p")
	LLMTopK                  = attribute.Key("llm.top_k")
	LLMFrequencyPenalty      = attribute.Key("llm.frequency_penalty")
	LLMPresencePenalty       = attribute.Key("llm.presence_penalty")
	LLMPrompts               = attribute.Key("llm.prompts")
	LLMCompletions           = attribute.Key("llm.completions")
	LLMChatStopSequence      = attribute.Key("llm.chat.stop_sequences")
	LLMRequestFunctions      = attribute.Key("llm.request.functions")

	// Vector DB
	VectorDBVendor    = attribute.Key("vector_db.vendor")
	VectorDBQueryTopK = attribute.Key("vector_db.query.top_k")

	// LLM Workflows
	TraceloopSpanKind              = attribute.Key("traceloop.span.kind")
	TraceloopWorkflowName          = attribute.Key("traceloop.workflow.name")
	TraceloopEntityName            = attribute.Key("traceloop.entity.name")
	TraceloopAssociationProperties = attribute.Key("traceloop.association.properties")
)
