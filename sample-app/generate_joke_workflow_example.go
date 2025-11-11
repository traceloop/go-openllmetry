package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
	sdk "github.com/traceloop/go-openllmetry/traceloop-sdk"
)

func createJoke(ctx context.Context, workflow *sdk.Workflow, client *openai.Client) (string, error) {
	task := workflow.NewTask("joke_creation")
	defer task.End()

	// Log prompt
	prompt := sdk.Prompt{
		Vendor: "openai",
		Mode:   "chat",
		Model:  "gpt-3.5-turbo",
		Messages: []sdk.Message{
			{
				Index:   0,
				Role:    "user",
				Content: "Tell me a joke about opentelemetry",
			},
		},
	}

	llmSpan := task.LogPrompt(prompt)

	// Make API call
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: "Tell me a joke about opentelemetry",
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("CreateChatCompletion error: %w", err)
	}

	// Log completion
	var completionMsgs []sdk.Message
	for _, choice := range resp.Choices {
		completionMsgs = append(completionMsgs, sdk.Message{
			Index:   choice.Index,
			Content: choice.Message.Content,
			Role:    choice.Message.Role,
		})
	}

	llmSpan.LogCompletion(ctx, sdk.Completion{
		Model:    resp.Model,
		Messages: completionMsgs,
	}, sdk.Usage{
		TotalTokens:      resp.Usage.TotalTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		PromptTokens:     resp.Usage.PromptTokens,
	})

	return resp.Choices[0].Message.Content, nil
}

func translateJokeToPirate(ctx context.Context, workflow *sdk.Workflow, client *openai.Client, joke string) (string, error) {
	// Log prompt
	piratePrompt := fmt.Sprintf("Translate the below joke to pirate-like english:\n\n%s", joke)
	prompt := sdk.Prompt{
		Vendor: "openai",
		Mode:   "chat",
		Model:  "gpt-3.5-turbo",
		Messages: []sdk.Message{
			{
				Index:   0,
				Role:    "user",
				Content: piratePrompt,
			},
		},
	}

	agent := workflow.NewAgent("joke_translation", map[string]string{
		"translation_type": "pirate",
	})
	defer agent.End()

	llmSpan := agent.LogPrompt(prompt)
	
	// Make API call
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: piratePrompt,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("CreateChatCompletion error: %w", err)
	}

	// Log completion
	var completionMsgs []sdk.Message
	for _, choice := range resp.Choices {
		completionMsgs = append(completionMsgs, sdk.Message{
			Index:   choice.Index,
			Content: choice.Message.Content,
			Role:    choice.Message.Role,
		})
	}

	llmSpan.LogCompletion(ctx, sdk.Completion{
		Model:    resp.Model,
		Messages: completionMsgs,
	}, sdk.Usage{
		TotalTokens:      resp.Usage.TotalTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		PromptTokens:     resp.Usage.PromptTokens,
	})

	// Call history jokes tool
	_, err = historyJokesTool(ctx, agent, client)
	if err != nil {
		fmt.Printf("Warning: history_jokes_tool error: %v\n", err)
	}

	return resp.Choices[0].Message.Content, nil
}

func historyJokesTool(ctx context.Context, agent *sdk.Agent, client *openai.Client) (string, error) {
	// Log prompt
	prompt := sdk.Prompt{
		Vendor: "openai",
		Mode:   "chat",
		Model:  "gpt-3.5-turbo",
		Messages: []sdk.Message{
			{
				Index:   0,
				Role:    "user",
				Content: "get some history jokes",
			},
		},
	}
	
	tool := agent.NewTool("history_jokes", "function", sdk.ToolFunction{
		Name:        "history_jokes",
		Description: "Get some history jokes",
		Parameters:  map[string]interface{}{},
	})
	defer tool.End()

	llmSpan := tool.LogPrompt(prompt)

	// Make API call
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: "get some history jokes",
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("CreateChatCompletion error: %w", err)
	}

	// Log completion
	var completionMsgs []sdk.Message
	for _, choice := range resp.Choices {
		completionMsgs = append(completionMsgs, sdk.Message{
			Index:   choice.Index,
			Content: choice.Message.Content,
			Role:    choice.Message.Role,
		})
	}

	llmSpan.LogCompletion(ctx, sdk.Completion{
		Model:    resp.Model,
		Messages: completionMsgs,
	}, sdk.Usage{
		TotalTokens:      resp.Usage.TotalTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		PromptTokens:     resp.Usage.PromptTokens,
	})

	return resp.Choices[0].Message.Content, nil
}

func generateSignature(ctx context.Context, workflow *sdk.Workflow, client *openai.Client, joke string) (string, error) {
	task := workflow.NewTask("signature_generation")
	defer task.End()

	signaturePrompt := "add a signature to the joke:\n\n" + joke

	// Log prompt
	prompt := sdk.Prompt{
		Vendor: "openai",
		Mode:   "completion",
		Model:  "davinci-002",
		Messages: []sdk.Message{
			{
				Index:   0,
				Role:    "user",
				Content: signaturePrompt,
			},
		},
	}

	llmSpan := task.LogPrompt(prompt)

	// Make API call
	resp, err := client.CreateCompletion(ctx, openai.CompletionRequest{
		Model:  "davinci-002",
		Prompt: signaturePrompt,
	})
	if err != nil {
		return "", fmt.Errorf("CreateCompletion error: %w", err)
	}

	// Log completion
	llmSpan.LogCompletion(ctx, sdk.Completion{
		Model: resp.Model,
		Messages: []sdk.Message{
			{
				Index:   0,
				Role:    "assistant",
				Content: resp.Choices[0].Text,
			},
		},
	}, sdk.Usage{
		TotalTokens:      resp.Usage.TotalTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		PromptTokens:     resp.Usage.PromptTokens,
	})

	return resp.Choices[0].Text, nil
}

func runJokeWorkflow() {
	ctx := context.Background()

	// Initialize Traceloop SDK
	traceloop, err := sdk.NewClient(ctx, sdk.Config{
		APIKey: os.Getenv("TRACELOOP_API_KEY"),
	})
	if err != nil {
		fmt.Printf("NewClient error: %v\n", err)
		return
	}
	defer func() { traceloop.Shutdown(ctx) }()

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	// Create workflow
	wf := traceloop.NewWorkflow(ctx, sdk.WorkflowAttributes{
		Name: "go-joke_generator",
		AssociationProperties: map[string]string{
			"user_id": "user_12345",
			"chat_id": "chat_1234",
		},
	})
	defer wf.End()

	// Execute workflow steps
	fmt.Println("Creating joke...")
	engJoke, err := createJoke(ctx, wf, client)
	if err != nil {
		fmt.Printf("Error creating joke: %v\n", err)
		return
	}
	fmt.Printf("\nEnglish joke:\n%s\n\n", engJoke)

	fmt.Println("Translating to pirate...")
	pirateJoke, err := translateJokeToPirate(ctx, wf, client, engJoke)
	if err != nil {
		fmt.Printf("Error translating joke: %v\n", err)
		return
	}
	fmt.Printf("\nPirate joke:\n%s\n\n", pirateJoke)

	fmt.Println("Generating signature...")
	signature, err := generateSignature(ctx, wf, client, pirateJoke)
	if err != nil {
		fmt.Printf("Error generating signature: %v\n", err)
		return
	}

	// Combine result
	result := pirateJoke + "\n\n" + signature
	fmt.Printf("\n=== Final Result ===\n%s\n", result)
}
