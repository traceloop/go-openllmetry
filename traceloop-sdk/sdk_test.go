package traceloop

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestLogPromptSpanAttributes(t *testing.T) {
	// Create in-memory exporter for testing
	exporter := tracetest.NewInMemoryExporter()
	
	// Create tracer provider with in-memory exporter
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	// Create traceloop instance
	tl := &Traceloop{
		config: Config{
			BaseURL: "https://api.traceloop.com",
			APIKey:  "test-key",
		},
		tracerProvider: tp,
	}

	// Create prompt with tool calling using new API
	prompt := Prompt{
		Vendor:  "openai",
		Mode:    "chat",
		Model:   "gpt-4o-mini",
		Messages: []Message{
			{
				Index:   0,
				Role:    "user",
				Content: "What's the weather like in San Francisco?",
			},
		},
		Tools: []Tool{
			{
				Type: "function",
				Function: ToolFunction{
					Name:        "get_weather",
					Description: "Get the current weather for a given location",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "The city and state, e.g. San Francisco, CA",
							},
						},
						"required": []string{"location"},
					},
				},
			},
		},
	}

	workflowAttrs := WorkflowAttributes{
		Name: "test-workflow",
		AssociationProperties: map[string]string{
			"entity_name": "test-entity",
		},
	}

	// Log the prompt using new workflow API
	llmSpan, err := tl.LogPrompt(context.Background(), prompt, workflowAttrs)
	if err != nil {
		t.Fatalf("LogPrompt failed: %v", err)
	}

	// Log completion with tool calls
	completion := Completion{
		Model: "gpt-4o-mini-2024-07-18",
		Messages: []Message{
			{
				Index:   0,
				Role:    "assistant",
				Content: "",
				ToolCalls: []ToolCall{
					{
						ID:   "call_YkIfypBQrmpUpxsKuS9aNdKg",
						Type: "function",
						Function: ToolCallFunction{
							Name:      "get_weather",
							Arguments: "{\"location\":\"San Francisco, CA\"}",
						},
					},
				},
			},
		},
	}

	usage := Usage{
		TotalTokens:      99,
		CompletionTokens: 17,
		PromptTokens:     82,
	}

	err = llmSpan.LogCompletion(context.Background(), completion, usage)
	if err != nil {
		t.Fatalf("LogCompletion failed: %v", err)
	}

	// Get the recorded spans
	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	t.Logf("Span name: %s", span.Name)
	t.Logf("Total attributes: %d", len(span.Attributes))

	// Print all attributes for debugging
	attributeMap := make(map[string]interface{})
	for _, attr := range span.Attributes {
		key := string(attr.Key)
		value := attr.Value.AsInterface()
		attributeMap[key] = value
		t.Logf("Attribute: %s = %v", key, value)
	}

	// Assert on specific attributes
	expectedAttrs := map[string]interface{}{
		"llm.vendor":                    "openai",
		"llm.request.model":            "gpt-4o-mini",
		"llm.request.type":             "chat",
		"llm.response.model":           "gpt-4o-mini-2024-07-18",
		"llm.usage.total_tokens":       int64(99),
		"llm.usage.completion_tokens":  int64(17),
		"llm.usage.prompt_tokens":      int64(82),
		"traceloop.workflow.name":      "test-workflow",
		"traceloop.association.properties.entity_name": "test-entity",
		"llm.prompts.0.content":        "What's the weather like in San Francisco?",
		"llm.prompts.0.role":           "user",
		"llm.completions.0.content":    "",
		"llm.completions.0.role":       "assistant",
		"llm.completions.0.tool_calls.0.id": "call_YkIfypBQrmpUpxsKuS9aNdKg",
		"llm.completions.0.tool_calls.0.type": "function",
		"llm.completions.0.tool_calls.0.name": "get_weather",
		"llm.completions.0.tool_calls.0.arguments": "{\"location\":\"San Francisco, CA\"}",
		"llm.request.functions.0.name": "get_weather",
		"llm.request.functions.0.description": "Get the current weather for a given location",
	}

	for expectedKey, expectedValue := range expectedAttrs {
		actualValue, exists := attributeMap[expectedKey]
		if !exists {
			t.Errorf("Expected attribute %s not found", expectedKey)
		} else if actualValue != expectedValue {
			t.Errorf("Attribute %s: expected %v, got %v", expectedKey, expectedValue, actualValue)
		}
	}

	// Check for JSON attributes as well
	if _, exists := attributeMap["llm.prompts"]; !exists {
		t.Error("Expected llm.prompts JSON attribute not found")
	}
	if _, exists := attributeMap["llm.completions"]; !exists {
		t.Error("Expected llm.completions JSON attribute not found")
	}
}