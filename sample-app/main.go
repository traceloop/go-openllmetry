package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
	tlp "github.com/traceloop/go-openllmetry/traceloop-sdk"
)

func main() {
	ctx := context.Background()

	traceloop, err := tlp.NewClient(ctx, tlp.Config{
		APIKey: os.Getenv("TRACELOOP_API_KEY"),
	})
	defer func() { traceloop.Shutdown(ctx) }()

	if err != nil {
		fmt.Printf("NewClient error: %v\n", err)
		return
	}

	request, err := traceloop.GetOpenAIChatCompletionRequest("example-prompt", map[string]interface{}{ "date": time.Now().Format("01/02") })
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


	log := tlp.PromptLogAttributes{
		Prompt: tlp.Prompt{
			Vendor: "openai",
			Mode:   "chat",
			Model: request.Model,
		},
		Completion: tlp.Completion{
			Model: resp.Model,
		},
		Usage: tlp.Usage{
			TotalTokens: resp.Usage.TotalTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			PromptTokens: resp.Usage.PromptTokens,
		},
	}	

	for i, message := range request.Messages {
		log.Prompt.Messages = append(log.Prompt.Messages, tlp.Message{
			Index:   i,
			Content: message.Content,
			Role:    message.Role,
		})
	}

	for _, choice := range resp.Choices {
		log.Completion.Messages = append(log.Completion.Messages, tlp.Message{
			Index:   choice.Index,
			Content: choice.Message.Content,
			Role:    choice.Message.Role,
		})
	}

	traceloop.LogPrompt(ctx, log)
}
