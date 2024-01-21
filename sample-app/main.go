package main

import (
	"context"
	"os"

	sdk "github.com/traceloop/go-openllmetry/traceloop-sdk"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/config"
)

func main() {
	ctx := context.Background()

	traceloop := sdk.NewClient(config.Config{
		BaseURL: "api-staging.traceloop.com",
		APIKey: os.Getenv("TRACELOOP_API_KEY"),
	})

	traceloop.Initialize(ctx)


	// request, err := traceloop.GetOpenAIChatCompletionRequest("eval-test", map[string]interface{}{ "a": "workout" })
	// if err != nil {
	// 	fmt.Printf("GetOpenAIChatCompletionRequest error: %v\n", err)
	// 	return
	// }
	
	// client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	// resp, err := client.CreateChatCompletion(
	// 	context.Background(),
	// 	*request,
	// )

	// if err != nil {
	// 	fmt.Printf("ChatCompletion error: %v\n", err)
	// 	return
	// }

	// fmt.Println(resp.Choices[0].Message.Content)

	traceloop.LogPrompt(ctx)
}
