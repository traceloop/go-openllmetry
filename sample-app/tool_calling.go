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

type WeatherParams struct {
	Location string `json:"location"`
	Unit     string `json:"unit,omitempty"`
}

func getWeather(location, unit string) string {
	return fmt.Sprintf("The weather in %s is sunny and 72°%s", location, unit)
}

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

func createWeatherTool() openai.ChatCompletionToolParam {
	return openai.ChatCompletionToolParam{
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
	}
}

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

func runToolCallingExample() {
	ctx := context.Background()

	traceloop := sdk.NewClient(config.Config{
		// BaseURL: os.Getenv("TRACELOOP_BASE_URL"),
		APIKey: "tl_4be59d06bb644ced90f8b21e2924a31e",
	})
	defer func() { traceloop.Shutdown(ctx) }()

	traceloop.Initialize(ctx)

	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	)

	tools := []openai.ChatCompletionToolParam{
		createWeatherTool(),
	}

	userPrompt := "What's the weather like in San Francisco?"
	fmt.Printf("User: %s\n", userPrompt)

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

	// Log the first API call (with tool calling)
	firstLog := dto.PromptLogAttributes{
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
			Messages: []dto.Message{
				{
					Index:     0,
					Content:   resp.Choices[0].Message.Content,
					Role:      "assistant",
					ToolCalls: convertOpenAIToolCallsToDTO(resp.Choices[0].Message.ToolCalls),
				},
			},
		},
		Usage: dto.Usage{
			TotalTokens:      int(resp.Usage.TotalTokens),
			CompletionTokens: int(resp.Usage.CompletionTokens),
			PromptTokens:     int(resp.Usage.PromptTokens),
		},
		Duration: int(duration.Milliseconds()),
	}

	if err := traceloop.LogPrompt(ctx, firstLog); err != nil {
		fmt.Printf("Error logging first API call: %v\n", err)
	}

	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		fmt.Println("\nTool calls requested:")

		var toolMessages []openai.ChatCompletionMessageParamUnion
		var toolCallResults []struct{
			ID string
			Result string
		}
		
		toolMessages = append(toolMessages, openai.UserMessage(userPrompt))
		
		toolMessages = append(toolMessages, openai.ChatCompletionMessage{
			Role:      openai.ChatCompletionMessageRoleAssistant,
			Content:   resp.Choices[0].Message.Content,
			ToolCalls: resp.Choices[0].Message.ToolCalls,
		})

		for _, toolCall := range resp.Choices[0].Message.ToolCalls {
			fmt.Printf("- Calling %s with arguments: %s\n", toolCall.Function.Name, toolCall.Function.Arguments)

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
			toolCallResults = append(toolCallResults, struct{ID string; Result string}{
				ID: toolCall.ID,
				Result: result,
			})
			toolMessages = append(toolMessages, openai.ToolMessage(toolCall.ID, result))
		}

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

		// Build the follow-up conversation context
		var followUpMessages []dto.Message
		// User message
		followUpMessages = append(followUpMessages, dto.Message{
			Index:   0,
			Content: userPrompt,
			Role:    "user",
		})
		// Assistant message with tool calls
		followUpMessages = append(followUpMessages, dto.Message{
			Index:     1,
			Content:   resp.Choices[0].Message.Content,
			Role:      "assistant",
			ToolCalls: convertOpenAIToolCallsToDTO(resp.Choices[0].Message.ToolCalls),
		})
		// Tool result messages
		for i, toolResult := range toolCallResults {
			followUpMessages = append(followUpMessages, dto.Message{
				Index:   i + 2,
				Content: toolResult.Result,
				Role:    "tool",
			})
		}
		
		// Log the second API call (follow-up with tool results)
		secondLog := dto.PromptLogAttributes{
			Prompt: dto.Prompt{
				Vendor:   "openai",
				Mode:     "chat",
				Model:    string(openai.ChatModelGPT4oMini),
				Messages: followUpMessages,
			},
			Completion: dto.Completion{
				Model: finalResp.Model,
				Messages: []dto.Message{
					{
						Index:   0,
						Content: finalResp.Choices[0].Message.Content,
						Role:    "assistant",
					},
				},
			},
			Usage: dto.Usage{
				TotalTokens:      int(finalResp.Usage.TotalTokens),
				CompletionTokens: int(finalResp.Usage.CompletionTokens),
				PromptTokens:     int(finalResp.Usage.PromptTokens),
			},
			Duration: int(duration.Milliseconds()),
		}

		if err := traceloop.LogPrompt(ctx, secondLog); err != nil {
			fmt.Printf("Error logging second API call: %v\n", err)
		}
	} else {
		// No tool calls - log simple interaction
		simpleLog := dto.PromptLogAttributes{
			Prompt: dto.Prompt{
				Vendor:      "openai",
				Mode:        "chat",
				Model:       string(openai.ChatModelGPT4oMini),
				Temperature: 0.7,
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
				Messages: []dto.Message{
					{
						Index:   0,
						Content: resp.Choices[0].Message.Content,
						Role:    "assistant",
					},
				},
			},
			Usage: dto.Usage{
				TotalTokens:      int(resp.Usage.TotalTokens),
				CompletionTokens: int(resp.Usage.CompletionTokens),
				PromptTokens:     int(resp.Usage.PromptTokens),
			},
			Duration: int(duration.Milliseconds()),
		}

		if err := traceloop.LogPrompt(ctx, simpleLog); err != nil {
			fmt.Printf("Error logging simple interaction: %v\n", err)
		}
	}

	fmt.Println("\nDone! Check your Traceloop dashboard to see the traced interactions with tool calling.")
}
