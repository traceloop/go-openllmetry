package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	sdk "github.com/traceloop/go-openllmetry/traceloop-sdk"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/config"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/dto"
)

// WeatherParams represents the parameters for the get_weather function
type WeatherParams struct {
	Location string `json:"location"`
	Unit     string `json:"unit,omitempty"`
}

// getWeather simulates a weather API call
func getWeather(location, unit string) string {
	return fmt.Sprintf("The weather in %s is sunny and 72°%s", location, unit)
}

// convertOpenAIToolCallsToDTO converts OpenAI tool calls to traceloop DTO format
func convertOpenAIToolCallsToDTO(toolCalls []openai.ChatCompletionMessageToolCall) []dto.ToolCall {
	var dtoToolCalls []dto.ToolCall
	for _, tc := range toolCalls {
		dtoToolCalls = append(dtoToolCalls, dto.ToolCall{
			ID:   tc.ID,
			Type: string(tc.Type),
			Function: dto.ToolCallFunction{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		})
	}
	return dtoToolCalls
}

// createWeatherTool creates the weather tool definition
func createWeatherTool() openai.ChatCompletionToolParam {
	return openai.ChatCompletionToolParam{
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
	}
}

// convertOpenAIToolsToDTO converts OpenAI tools to traceloop DTO format (simplified)
func convertToolsToDTO() []dto.Tool {
	return []dto.Tool{
		{
			Type: "function",
			Function: dto.ToolFunction{
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
							"enum":        []string{"C", "F"},
							"description": "The unit for temperature",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}
}

func main() {
	ctx := context.Background()

	// Initialize Traceloop SDK
	traceloop := sdk.NewClient(config.Config{
		BaseURL: os.Getenv("TRACELOOP_BASE_URL"),
		APIKey:  os.Getenv("TRACELOOP_API_KEY"),
	})
	defer func() { traceloop.Shutdown(ctx) }()

	traceloop.Initialize(ctx)

	// Initialize OpenAI client
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	)

	// Define available tools
	tools := []openai.ChatCompletionToolParam{
		createWeatherTool(),
	}

	// Create initial message
	userPrompt := "What's the weather like in San Francisco?"
	fmt.Printf("User: %s\n", userPrompt)

	// First API call with tool calling enabled
	startTime := time.Now()
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.F(openai.ChatModelGPT4oMini),
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(userPrompt),
		}),
		Tools: openai.F(tools),
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	duration := time.Since(startTime)

	fmt.Printf("\nAssistant: %s\n", resp.Choices[0].Message.Content)

	// Log the initial request and response with tools
	log := dto.PromptLogAttributes{
		Prompt: dto.Prompt{
			Vendor:      "openai",
			Mode:        "chat",
			Model:       string(openai.ChatModelGPT4oMini),
			Temperature: 0.7,
			Tools:       convertToolsToDTO(),
			Messages: []dto.Message{
				{
					Index:   0,
					Content: userPrompt,
					Role:    "user",
				},
			},
		},
		Completion: dto.Completion{
			Model: resp.Model,
		},
		Usage: dto.Usage{
			TotalTokens:      int(resp.Usage.TotalTokens),
			CompletionTokens: int(resp.Usage.CompletionTokens),
			PromptTokens:     int(resp.Usage.PromptTokens),
		},
		Duration: int(duration.Milliseconds()),
	}

	// Add response message with tool calls
	completionMsg := dto.Message{
		Index:   0,
		Content: resp.Choices[0].Message.Content,
		Role:    "assistant",
	}

	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		completionMsg.ToolCalls = convertOpenAIToolCallsToDTO(resp.Choices[0].Message.ToolCalls)
	}

	log.Completion.Messages = append(log.Completion.Messages, completionMsg)

	// Log the first interaction
	if err := traceloop.LogPrompt(ctx, log); err != nil {
		fmt.Printf("Error logging prompt: %v\n", err)
	}

	// Handle tool calls if any
	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		fmt.Println("\nTool calls requested:")

		var toolMessages []openai.ChatCompletionMessageParamUnion
		toolMessages = append(toolMessages, openai.UserMessage(userPrompt))
		toolMessages = append(toolMessages, openai.AssistantMessage(resp.Choices[0].Message.Content))

		// Execute each tool call
		for _, toolCall := range resp.Choices[0].Message.ToolCalls {
			fmt.Printf("- Calling %s with arguments: %s\n", toolCall.Function.Name, toolCall.Function.Arguments)

			// Handle the tool call
			var result string
			if toolCall.Function.Name == "get_weather" {
				var params WeatherParams
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
					result = fmt.Sprintf("Error parsing parameters: %v", err)
				} else {
					if params.Unit == "" {
						params.Unit = "F"
					}
					result = getWeather(params.Location, params.Unit)
				}
			} else {
				result = fmt.Sprintf("Unknown function: %s", toolCall.Function.Name)
			}

			fmt.Printf("  Result: %s\n", result)

			// Add tool result to conversation
			toolMessages = append(toolMessages, openai.ToolMessage(toolCall.ID, result))
		}

		// Make follow-up call to get final response
		fmt.Println("\nGetting final response...")
		startTime = time.Now()
		finalResp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model:    openai.F(openai.ChatModelGPT4oMini),
			Messages: openai.F(toolMessages),
		})
		if err != nil {
			fmt.Printf("Error in follow-up call: %v\n", err)
			return
		}
		duration = time.Since(startTime)

		fmt.Printf("\nFinal Assistant Response: %s\n", finalResp.Choices[0].Message.Content)

		// Log the follow-up interaction (simplified)
		followUpLog := dto.PromptLogAttributes{
			Prompt: dto.Prompt{
				Vendor: "openai",
				Mode:   "chat",
				Model:  string(openai.ChatModelGPT4oMini),
				Messages: []dto.Message{
					{Index: 0, Content: userPrompt, Role: "user"},
					{Index: 1, Content: resp.Choices[0].Message.Content, Role: "assistant", ToolCalls: convertOpenAIToolCallsToDTO(resp.Choices[0].Message.ToolCalls)},
					{Index: 2, Content: fmt.Sprintf("Tool result for get_weather"), Role: "tool"},
				},
			},
			Completion: dto.Completion{
				Model: finalResp.Model,
				Messages: []dto.Message{
					{Index: 0, Content: finalResp.Choices[0].Message.Content, Role: "assistant"},
				},
			},
			Usage: dto.Usage{
				TotalTokens:      int(finalResp.Usage.TotalTokens),
				CompletionTokens: int(finalResp.Usage.CompletionTokens),
				PromptTokens:     int(finalResp.Usage.PromptTokens),
			},
			Duration: int(duration.Milliseconds()),
		}

		// Log the follow-up interaction
		if err := traceloop.LogPrompt(ctx, followUpLog); err != nil {
			fmt.Printf("Error logging follow-up prompt: %v\n", err)
		}
	}

	fmt.Println("\nDone! Check your Traceloop dashboard to see the traced interactions with tool calling.")
}