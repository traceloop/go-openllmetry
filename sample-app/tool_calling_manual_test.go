package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	sdk "github.com/traceloop/go-openllmetry/traceloop-sdk"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/config"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/dto"
)

// Mock OpenAI response for tool calling
const mockToolCallingResponse = `{
  "id": "chatcmpl-test123",
  "object": "chat.completion",
  "created": 1699014393,
  "model": "gpt-4o-mini-2024-07-18",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "",
      "tool_calls": [{
        "id": "call_YkIfypBQrmpUpxsKuS9aNdKg",
        "type": "function",
        "function": {
          "name": "get_weather",
          "arguments": "{\"location\":\"San Francisco, CA\"}"
        }
      }]
    },
    "finish_reason": "tool_calls"
  }],
  "usage": {
    "prompt_tokens": 82,
    "completion_tokens": 17,
    "total_tokens": 99
  }
}`

func TestToolCallingWithHTTPMock(t *testing.T) {
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if it's an OpenAI request
		if strings.Contains(r.URL.Path, "chat/completions") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockToolCallingResponse))
			return
		}
		// For other requests (like traceloop), return OK
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer mockServer.Close()

	ctx := context.Background()

	// Initialize traceloop with mock
	traceloop := sdk.NewClient(config.Config{
		BaseURL: mockServer.URL, // Point to our mock server
		APIKey:  "test-key-for-mocking",
	})
	defer func() { traceloop.Shutdown(ctx) }()

	traceloop.Initialize(ctx)

	// Create OpenAI client pointing to our mock server
	client := openai.NewClient(
		option.WithAPIKey("mock-api-key"),
		option.WithBaseURL(mockServer.URL), // Point to our mock server
	)

	// Test the tool calling request
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.F(openai.ChatModelGPT4oMini),
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("What's the weather like in San Francisco?"),
		}),
		Tools: openai.F([]openai.ChatCompletionToolParam{
			{
				Type: openai.F(openai.ChatCompletionToolTypeFunction),
				Function: openai.F(openai.FunctionDefinitionParam{
					Name:        openai.F("get_weather"),
					Description: openai.F("Get the current weather for a given location"),
					Parameters: openai.F(openai.FunctionParameters{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "The city and state, e.g. San Francisco, CA",
							},
						},
						"required": []string{"location"},
					}),
				}),
			},
		}),
	})

	if err != nil {
		t.Fatalf("Mock OpenAI API call failed: %v", err)
	}

	// Verify the response structure
	if len(resp.Choices) == 0 {
		t.Fatal("Expected at least one choice in response")
	}

	choice := resp.Choices[0]
	if len(choice.Message.ToolCalls) == 0 {
		t.Fatal("Expected tool calls in response")
	}

	// Verify the tool call details
	toolCall := choice.Message.ToolCalls[0]
	if toolCall.Function.Name != "get_weather" {
		t.Errorf("Expected tool call name 'get_weather', got '%s'", toolCall.Function.Name)
	}

	if toolCall.ID != "call_YkIfypBQrmpUpxsKuS9aNdKg" {
		t.Errorf("Expected tool call ID 'call_YkIfypBQrmpUpxsKuS9aNdKg', got '%s'", toolCall.ID)
	}

	// Test the traceloop logging with mock data
	log := dto.PromptLogAttributes{
		Prompt: dto.Prompt{
			Vendor:      "openai",
			Mode:        "chat",
			Model:       "gpt-4o-mini",
			Temperature: 0.7,
			Tools: []dto.Tool{
				{
					Type: "function",
					Function: dto.ToolFunction{
						Name:        "get_weather",
						Description: "Get the current weather for a given location",
					},
				},
			},
			Messages: []dto.Message{
				{
					Index:   0,
					Content: "What's the weather like in San Francisco?",
					Role:    "user",
				},
			},
		},
		Completion: dto.Completion{
			Model: resp.Model,
			Messages: []dto.Message{
				{
					Index:   0,
					Content: choice.Message.Content,
					Role:    "assistant",
					ToolCalls: []dto.ToolCall{
						{
							ID:   toolCall.ID,
							Type: string(toolCall.Type),
							Function: dto.ToolCallFunction{
								Name:      toolCall.Function.Name,
								Arguments: toolCall.Function.Arguments,
							},
						},
					},
				},
			},
		},
		Usage: dto.Usage{
			TotalTokens:      int(resp.Usage.TotalTokens),
			CompletionTokens: int(resp.Usage.CompletionTokens),
			PromptTokens:     int(resp.Usage.PromptTokens),
		},
		Duration: 1500,
	}

	// Test logging (this will hit our mock server)
	err = traceloop.LogPrompt(ctx, log)
	if err != nil {
		t.Fatalf("LogPrompt failed: %v", err)
	}

	t.Log("Successfully tested tool calling with HTTP mocks")
	t.Logf("Tool call: %s(%s)", toolCall.Function.Name, toolCall.Function.Arguments)
}