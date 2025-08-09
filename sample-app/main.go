package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
	sdk "github.com/traceloop/go-openllmetry/traceloop-sdk"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/dto"
)

func main() {
	ctx := context.Background()
	
	if len(os.Args) > 1 && os.Args[1] == "tool-calling" {
		runToolCallingExample()
		return
	}
	
	if len(os.Args) > 1 && os.Args[1] == "workflow" {
		workflowExample()
		return
	}

	// Default to legacy example
	legacyExample()
}

func workflowExample() {
	ctx := context.Background()

	traceloop, err := sdk.NewClient(ctx, sdk.Config{
		BaseURL: "https://api.traceloop.com",
		APIKey:  os.Getenv("TRACELOOP_API_KEY"),
	})
	if err != nil {
		fmt.Printf("NewClient error: %v\n", err)
		return
	}
	defer func() { traceloop.Shutdown(ctx) }()

	request, err := traceloop.GetOpenAIChatCompletionRequest("example-prompt", map[string]interface{}{"date": time.Now().Format("01/02")})
	if err != nil {
		fmt.Printf("GetOpenAIChatCompletionRequest error: %v\n", err)
		return
	}

	var promptMsgs []sdk.Message
	for i, message := range request.Messages {
		promptMsgs = append(promptMsgs, sdk.Message{
			Index:   i,
			Content: message.Content,
			Role:    message.Role,
		})
	}

	llmSpan, err := traceloop.LogPrompt(
		ctx,
		sdk.Prompt{
			Vendor:   "openai",
			Mode:     "chat",
			Model:    request.Model,
			Messages: promptMsgs,
		},
		sdk.WorkflowAttributes{
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

	fmt.Println(resp.Choices[0].Message.Content)
}

func legacyExample() {
	ctx := context.Background()

	// For backward compatibility, provide a constructor that mimics old API
	traceloop, err := sdk.NewClient(ctx, sdk.Config{
		BaseURL: "https://api.traceloop.com",
		APIKey:  os.Getenv("TRACELOOP_API_KEY"),
	})
	if err != nil {
		fmt.Printf("NewClient error: %v\n", err)
		return
	}
	defer func() { traceloop.Shutdown(ctx) }()

	request, err := traceloop.GetOpenAIChatCompletionRequest("example-prompt", map[string]interface{}{"date": time.Now().Format("01/02")})
	if err != nil {
		fmt.Printf("GetOpenAIChatCompletionRequest error: %v\n", err)
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

	fmt.Println(resp.Choices[0].Message.Content)

	log := dto.PromptLogAttributes{
		Prompt: dto.Prompt{
			Vendor: "openai",
			Mode:   "chat",
			Model: request.Model,
		},
		Completion: dto.Completion{
			Model: resp.Model,
		},
		Usage: dto.Usage{
			TotalTokens: resp.Usage.TotalTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			PromptTokens: resp.Usage.PromptTokens,
		},
		Traceloop: dto.TraceloopAttributes{
			WorkflowName: "legacy-example",
		},
	}

	for i, message := range request.Messages {
		log.Prompt.Messages = append(log.Prompt.Messages, dto.Message{
			Index:   i,
			Content: message.Content,
			Role:    message.Role,
		})
	}

	for _, choice := range resp.Choices {
		log.Completion.Messages = append(log.Completion.Messages, dto.Message{
			Index:   choice.Index,
			Content: choice.Message.Content,
			Role:    choice.Message.Role,
		})
	}

	traceloop.LogPromptLegacy(ctx, log)
}
