package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	sdk "github.com/traceloop/go-openllmetry/traceloop-sdk"
)

type WeatherParams struct {
	Location string `json:"location"`
	Unit     string `json:"unit,omitempty"`
}

func getWeather(location, unit string) string {
	return fmt.Sprintf("The weather in %s is sunny and 72°%s", location, unit)
}

func runToolCallingExample() {
	ctx := context.Background()

	baseURL := os.Getenv("TRACELOOP_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.traceloop.com"
	}
	
	traceloop, err := sdk.NewClient(ctx, sdk.Config{
		BaseURL: baseURL,
		APIKey:  os.Getenv("TRACELOOP_API_KEY"),
	})
	if err != nil {
		log.Printf("NewClient error: %v", err)
		return
	}
	defer func() { traceloop.Shutdown(ctx) }()

	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	)

	userPrompt := "What's the weather like in San Francisco?"
	
	// Define tools
	tools := []sdk.Tool{
		{
			Type: "function",
			Function: sdk.ToolFunction{
				Name:        "get_weather",
				Description: "Get the current weather for a given location",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
						"unit": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"celsius", "fahrenheit"},
							"description": "The unit of temperature",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	// Create prompt
	prompt := sdk.Prompt{
		Vendor: "openai",
		Mode:   "chat",
		Model:  "gpt-4o-mini",
		Messages: []sdk.Message{
			{
				Index:   0,
				Role:    "user",
				Content: userPrompt,
			},
		},
		Tools: tools,
	}

	workflowAttrs := sdk.WorkflowAttributes{
		Name: "tool-calling-example",
		AssociationProperties: map[string]string{
			"user_id": "demo-user",
		},
	}

	fmt.Printf("User: %s\n", userPrompt)
	
	// Log the prompt
	llmSpan, err := traceloop.LogPrompt(ctx, prompt, workflowAttrs)
	if err != nil {
		fmt.Printf("Error logging prompt: %v\n", err)
		return
	}

	// Make API call to OpenAI
	startTime := time.Now()
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(userPrompt),
		}),
		Model: openai.F(openai.ChatModelGPT4oMini),
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
							"unit": map[string]interface{}{
								"type":        "string",
								"enum":        []string{"celsius", "fahrenheit"},
								"description": "The unit of temperature",
							},
						},
						"required": []string{"location"},
					}),
				}),
			},
		}),
		Temperature: openai.F(0.7),
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	duration := time.Since(startTime)

	fmt.Printf("\nAssistant: %s\n", resp.Choices[0].Message.Content)

	// Convert response to our format
	var completionMessages []sdk.Message
	for _, choice := range resp.Choices {
		message := sdk.Message{
			Index:   int(choice.Index),
			Role:    string(choice.Message.Role),
			Content: choice.Message.Content,
		}
		
		// Convert tool calls if present
		if len(choice.Message.ToolCalls) > 0 {
			for _, toolCall := range choice.Message.ToolCalls {
				message.ToolCalls = append(message.ToolCalls, sdk.ToolCall{
					ID:   toolCall.ID,
					Type: string(toolCall.Type),
					Function: sdk.ToolCallFunction{
						Name:      toolCall.Function.Name,
						Arguments: toolCall.Function.Arguments,
					},
				})
			}
		}
		completionMessages = append(completionMessages, message)
	}

	// Log the completion
	completion := sdk.Completion{
		Model:    resp.Model,
		Messages: completionMessages,
	}

	usage := sdk.Usage{
		TotalTokens:      int(resp.Usage.TotalTokens),
		CompletionTokens: int(resp.Usage.CompletionTokens),
		PromptTokens:     int(resp.Usage.PromptTokens),
	}

	err = llmSpan.LogCompletion(ctx, completion, usage)
	if err != nil {
		fmt.Printf("Error logging completion: %v\n", err)
		return
	}

	// If tool calls were made, execute them
	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		fmt.Println("\nTool calls requested:")

		for _, toolCall := range resp.Choices[0].Message.ToolCalls {
			if toolCall.Function.Name == "get_weather" {
				fmt.Printf("Tool call: %s with args: %s\n", toolCall.Function.Name, toolCall.Function.Arguments)
				
				var params WeatherParams
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
					fmt.Printf("Error parsing arguments: %v\n", err)
					continue
				}

				result := getWeather(params.Location, params.Unit)
				fmt.Printf("Function result: %s\n", result)
			}
		}
	}

	fmt.Printf("\nRequest completed in %v\n", duration)
}