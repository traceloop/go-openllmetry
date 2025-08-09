package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
	tlp "github.com/traceloop/go-openllmetry/traceloop-sdk"
)

func workflow_example() {
	ctx := context.Background()

	traceloop, err := tlp.NewClient(ctx, tlp.Config{
		BaseURL: "api-staging.traceloop.com",
		APIKey:  os.Getenv("TRACELOOP_API_KEY"),
	})
	defer func() { traceloop.Shutdown(ctx) }()

	if err != nil {
		fmt.Printf("NewClient error: %v\n", err)
		return
	}

	request, err := traceloop.GetOpenAIChatCompletionRequest("example-prompt", map[string]interface{}{"date": time.Now().Format("01/02")})
	if err != nil {
		fmt.Printf("GetOpenAIChatCompletionRequest error: %v\n", err)
		return
	}

	var promptMsgs []tlp.Message
	for i, message := range request.Messages {
		promptMsgs = append(promptMsgs, tlp.Message{
			Index:   i,
			Content: message.Content,
			Role:    message.Role,
		})
	}

	llmSpan, err := traceloop.LogPrompt(
		ctx,
		tlp.Prompt{
			Vendor:   "openai",
			Mode:     "chat",
			Model:    request.Model,
			Messages: promptMsgs,
		},
		tlp.WorkflowAttributes{
			Name: "example-workflow",
		},
	)
	if err != nil {
		fmt.Printf("LogPrompt error: %v\n", err)
		return
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		*request,
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	var completionMsgs []tlp.Message
	for _, choice := range resp.Choices {
		completionMsgs = append(completionMsgs, tlp.Message{
			Index:   choice.Index,
			Content: choice.Message.Content,
			Role:    choice.Message.Role,
		})
	}

	llmSpan.LogCompletion(ctx, tlp.Completion{
		Model:    resp.Model,
		Messages: completionMsgs,
	}, tlp.Usage{
		TotalTokens:      resp.Usage.TotalTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		PromptTokens:     resp.Usage.PromptTokens,
	})

	fmt.Println(resp.Choices[0].Message.Content)
}

type WeatherParams struct {
	Location string `json:"location"`
	Unit     string `json:"unit,omitempty"`
}

func getWeather(location, unit string) string {
	return fmt.Sprintf("The weather in %s is sunny and 72°%s", location, unit)
}

func convertOpenAIToolCallsToTLP(toolCalls []openai.ToolCall) []tlp.ToolCall {
	var tlpToolCalls []tlp.ToolCall
	for _, tc := range toolCalls {
		tlpToolCalls = append(tlpToolCalls, tlp.ToolCall{
			ID:   tc.ID,
			Type: string(tc.Type),
			Function: tlp.ToolCallFunction{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		})
	}
	return tlpToolCalls
}

func createWeatherTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: openai.FunctionDefinition{
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
	}
}

func convertToolsToTLP() []tlp.Tool {
	return []tlp.Tool{
		{
			Type: "function",
			Function: tlp.ToolFunction{
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

	traceloop, err := tlp.NewClient(ctx, tlp.Config{
		BaseURL: "api-staging.traceloop.com",
		APIKey:  os.Getenv("TRACELOOP_API_KEY"),
	})
	defer func() { traceloop.Shutdown(ctx) }()

	if err != nil {
		fmt.Printf("NewClient error: %v\n", err)
		return
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	tools := []openai.Tool{createWeatherTool()}

	userPrompt := "What's the weather like in San Francisco?"
	fmt.Printf("User: %s\n", userPrompt)

	// Log the first API call (with tool calling request)
	promptMessages := []tlp.Message{
		{
			Index:   0,
			Content: userPrompt,
			Role:    "user",
		},
	}

	llmSpan, err := traceloop.LogPrompt(
		ctx,
		tlp.Prompt{
			Vendor:   "openai",
			Mode:     "chat",
			Model:    "gpt-4o-mini",
			Messages: promptMessages,
			Tools:    convertToolsToTLP(),
		},
		tlp.WorkflowAttributes{
			Name: "tool-calling-workflow",
		},
	)
	if err != nil {
		fmt.Printf("LogPrompt error: %v\n", err)
		return
	}

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		Tools: tools,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nAssistant: %s\n", resp.Choices[0].Message.Content)

	// Log the completion
	completionMessages := []tlp.Message{
		{
			Index:     0,
			Content:   resp.Choices[0].Message.Content,
			Role:      "assistant",
			ToolCalls: convertOpenAIToolCallsToTLP(resp.Choices[0].Message.ToolCalls),
		},
	}

	llmSpan.LogCompletion(ctx, tlp.Completion{
		Model:    resp.Model,
		Messages: completionMessages,
	}, tlp.Usage{
		TotalTokens:      int(resp.Usage.TotalTokens),
		CompletionTokens: int(resp.Usage.CompletionTokens),
		PromptTokens:     int(resp.Usage.PromptTokens),
	})

	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		fmt.Println("\nTool calls requested:")

		var toolMessages []openai.ChatCompletionMessage
		var toolCallResults []struct {
			ID     string
			Result string
		}

		toolMessages = append(toolMessages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: userPrompt,
		})
		toolMessages = append(toolMessages, openai.ChatCompletionMessage{
			Role:      openai.ChatMessageRoleAssistant,
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
			toolCallResults = append(toolCallResults, struct {
				ID     string
				Result string
			}{
				ID:     toolCall.ID,
				Result: result,
			})
			toolMessages = append(toolMessages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result,
				ToolCallID: toolCall.ID,
			})
		}

		fmt.Println("\nGetting final response...")

		// Log the follow-up prompt with tool results
		followUpMessages := []tlp.Message{
			{
				Index:   0,
				Content: userPrompt,
				Role:    "user",
			},
			{
				Index:     1,
				Content:   resp.Choices[0].Message.Content,
				Role:      "assistant",
				ToolCalls: convertOpenAIToolCallsToTLP(resp.Choices[0].Message.ToolCalls),
			},
		}

		for i, toolResult := range toolCallResults {
			followUpMessages = append(followUpMessages, tlp.Message{
				Index:   i + 2,
				Content: toolResult.Result,
				Role:    "tool",
			})
		}

		followUpSpan, err := traceloop.LogPrompt(
			ctx,
			tlp.Prompt{
				Vendor:   "openai",
				Mode:     "chat",
				Model:    "gpt-4o-mini",
				Messages: followUpMessages,
			},
			tlp.WorkflowAttributes{
				Name: "tool-calling-workflow",
			},
		)
		if err != nil {
			fmt.Printf("LogPrompt error for follow-up: %v\n", err)
			return
		}

		finalResp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    "gpt-4o-mini",
			Messages: toolMessages,
		})
		if err != nil {
			fmt.Printf("Error in follow-up call: %v\n", err)
			return
		}

		fmt.Printf("\nFinal Assistant Response: %s\n", finalResp.Choices[0].Message.Content)

		followUpSpan.LogCompletion(ctx, tlp.Completion{
			Model: finalResp.Model,
			Messages: []tlp.Message{
				{
					Index:   0,
					Content: finalResp.Choices[0].Message.Content,
					Role:    "assistant",
				},
			},
		}, tlp.Usage{
			TotalTokens:      int(finalResp.Usage.TotalTokens),
			CompletionTokens: int(finalResp.Usage.CompletionTokens),
			PromptTokens:     int(finalResp.Usage.PromptTokens),
		})
	}

	fmt.Println("\nDone! Check your Traceloop dashboard to see the traced interactions with tool calling.")
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "tool-calling" {
		runToolCallingExample()
		return
	}

	workflow_example()
}
