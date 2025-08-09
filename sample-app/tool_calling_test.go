package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v2/recorder"
	"gopkg.in/dnaeon/go-vcr.v2/cassette"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	sdk "github.com/traceloop/go-openllmetry/traceloop-sdk"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/config"
)

func TestToolCallingWithMock(t *testing.T) {
	// Create VCR recorder
	cassettePath := filepath.Join("testdata", "tool_calling_cassette")
	r, err := recorder.New(cassettePath)
	if err != nil {
		t.Fatalf("Failed to create recorder: %v", err)
	}
	defer r.Stop()

	// Configure recorder to sanitize sensitive data
	r.AddFilter(func(i *cassette.Interaction) error {
		// Remove Authorization header from requests
		delete(i.Request.Headers, "Authorization")
		
		// Remove any OpenAI API key patterns from the request body
		if i.Request.Body != "" {
			// This is just extra safety - OpenAI keys shouldn't be in request bodies anyway
			i.Request.Body = ""
		}
		
		return nil
	})

	// Create custom HTTP client with recorder
	httpClient := &http.Client{
		Transport: r,
	}

	ctx := context.Background()

	// Initialize traceloop (will work without real API key in replay mode)
	traceloop := sdk.NewClient(config.Config{
		BaseURL: "https://api.traceloop.com",
		APIKey:  "test-key-for-mocking",
	})
	defer func() { traceloop.Shutdown(ctx) }()

	traceloop.Initialize(ctx)

	// Create OpenAI client with custom HTTP transport
	// In recording mode, use real API key. In replay mode, any key works.
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = "mock-api-key-for-testing"
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithHTTPClient(httpClient),
	)

	// Create weather tool (same as main example)
	tools := []openai.ChatCompletionToolParam{
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
						"unit": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"C", "F"},
							"description": "The unit for temperature",
						},
					},
					"required": []string{"location"},
				}),
			}),
		},
	}

	// Test the tool calling flow
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.F(openai.ChatModelGPT4oMini),
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("What's the weather like in San Francisco?"),
		}),
		Tools: openai.F(tools),
	})
	if err != nil {
		t.Fatalf("OpenAI API call failed: %v", err)
	}

	// Verify we got tool calls
	if len(resp.Choices) == 0 {
		t.Fatal("Expected at least one choice in response")
	}

	choice := resp.Choices[0]
	if len(choice.Message.ToolCalls) == 0 {
		t.Fatal("Expected tool calls in response")
	}

	// Verify the tool call
	toolCall := choice.Message.ToolCalls[0]
	if toolCall.Function.Name != "get_weather" {
		t.Errorf("Expected tool call name 'get_weather', got '%s'", toolCall.Function.Name)
	}

	t.Logf("Successfully got tool call: %s with args: %s", toolCall.Function.Name, toolCall.Function.Arguments)
	t.Logf("Tool call ID: %s, Type: %s", toolCall.ID, toolCall.Type)
}

func TestToolCallingIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	// This test runs the actual tool calling example
	// It will use real API keys when INTEGRATION_TEST is set
	runToolCallingExample()
}