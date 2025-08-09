package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	sdk "github.com/traceloop/go-openllmetry/traceloop-sdk"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	if len(os.Args) > 1 && os.Args[1] == "tool-calling" {
		runToolCallingExample()
		return
	}
	
	// Default to workflow example using prompt registry
	workflowExample()
}

func workflowExample() {
	ctx := context.Background()

	traceloop, err := sdk.NewClient(ctx, sdk.Config{
		APIKey: os.Getenv("TRACELOOP_API_KEY"),
	})
	if err != nil {
		fmt.Printf("NewClient error: %v\n", err)
		return
	}
	defer func() { traceloop.Shutdown(ctx) }()

	// Wait a bit for prompt registry to populate
	time.Sleep(2 * time.Second)
	
	// Get prompt from registry
	request, err := traceloop.GetOpenAIChatCompletionRequest("question_answering", map[string]interface{}{
		"date":        time.Now().Format("01/02"),
		"question":    "What's the weather like today?",
		"information": "The current weather is sunny and 75 degrees.",
	})
	if err != nil {
		fmt.Printf("GetOpenAIChatCompletionRequest error: %v\n", err)
		return
	}

	// Convert to our format for logging
	var promptMsgs []sdk.Message
	for i, message := range request.Messages {
		promptMsgs = append(promptMsgs, sdk.Message{
			Index:   i,
			Content: message.Content,
			Role:    message.Role,
		})
	}

	// Log the prompt
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
			AssociationProperties: map[string]string{
				"user_id": "demo-user",
			},
		},
	)
	if err != nil {
		fmt.Printf("LogPrompt error: %v\n", err)
		return
	}

	// Make actual OpenAI API call
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		*request,
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	// Convert response to our format for logging
	var completionMsgs []sdk.Message
	for _, choice := range resp.Choices {
		completionMsgs = append(completionMsgs, sdk.Message{
			Index:   choice.Index,
			Content: choice.Message.Content,
			Role:    choice.Message.Role,
		})
	}

	// Log the completion
	err = llmSpan.LogCompletion(ctx, sdk.Completion{
		Model:    resp.Model,
		Messages: completionMsgs,
	}, sdk.Usage{
		TotalTokens:      resp.Usage.TotalTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		PromptTokens:     resp.Usage.PromptTokens,
	})
	if err != nil {
		fmt.Printf("LogCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
}