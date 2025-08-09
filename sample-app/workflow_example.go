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
		BaseURL: "api-staging.traceloop.com",
		APIKey:  os.Getenv("TRACELOOP_API_KEY"),
	})
	defer func() { traceloop.Shutdown(ctx) }()

	if err != nil {
		fmt.Printf("NewClient error: %v\n", err)
		return
	}

	wf := traceloop.NewWorkflow(ctx, tlp.WorkflowAttributes{
		Name: "history_generation",
	})
	defer wf.End()

	factGenTask := wf.NewTask("current_date_fact_generation")
	defer factGenTask.End()

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

	llmSpan, err := factGenTask.LogPrompt(
		tlp.Prompt{
			Vendor:   "openai",
			Mode:     "chat",
			Model:    request.Model,
			Messages: promptMsgs,
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

	someOtherTask := wf.NewTask("some_other_task")
	defer someOtherTask.End()

	otherPrompt, _ := someOtherTask.LogPrompt(tlp.Prompt{
		Vendor: "openai",
		Mode:   "chat",
		Model:  request.Model,
		Messages: []tlp.Message{
			{
				Index:   0,
				Content: "some other prompt",
				Role:    "user",
			},
		},
	})

	otherPrompt.LogCompletion(ctx, tlp.Completion{
		Model:    resp.Model,
		Messages: completionMsgs,
	}, tlp.Usage{
		TotalTokens:      resp.Usage.TotalTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		PromptTokens:     resp.Usage.PromptTokens,
	})

	fmt.Println(resp.Choices[0].Message.Content)
}
