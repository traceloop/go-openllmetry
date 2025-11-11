package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
	sdk "github.com/traceloop/go-openllmetry/traceloop-sdk"
)

var associationProperties = map[string]string{
	"user_id": "user_67890",
	"ab_testing_variant": "variant_a",
}

// ingredientValidatorTool validates that requested ingredients are available/safe
func ingredientValidatorTool(ctx context.Context, agent *sdk.Agent, client *openai.Client, ingredients string) (string, error) {
	tool := agent.NewTool("ingredient_validator", "function", sdk.ToolFunction{
		Name:        "ingredient_validator",
		Description: "Validates that requested ingredients are available and safe to use",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"ingredients": map[string]string{
					"type":        "string",
					"description": "Comma-separated list of ingredients to validate",
				},
			},
		},
	}, associationProperties)
	defer tool.End()

	prompt := sdk.Prompt{
		Vendor: "openai",
		Mode:   "chat",
		Model:  "gpt-3.5-turbo",
		Messages: []sdk.Message{
			{
				Index:   0,
				Role:    "user",
				Content: fmt.Sprintf("Validate these ingredients are commonly available and safe: %s. Respond with 'Valid' or list any concerns.", ingredients),
			},
		},
	}

	llmSpan := tool.LogPrompt(prompt)

	// Make API call
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: prompt.Messages[0].Content,
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

// nutritionCalculatorTool calculates nutritional information for the recipe
func nutritionCalculatorTool(ctx context.Context, agent *sdk.Agent, client *openai.Client, recipe string) (string, error) {
	tool := agent.NewTool("nutrition_calculator", "function", sdk.ToolFunction{
		Name:        "nutrition_calculator",
		Description: "Calculates estimated nutritional information for a recipe",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"recipe": map[string]string{
					"type":        "string",
					"description": "The recipe description",
				},
			},
		},
	}, associationProperties)
	defer tool.End()

	prompt := sdk.Prompt{
		Vendor: "openai",
		Mode:   "chat",
		Model:  "gpt-3.5-turbo",
		Messages: []sdk.Message{
			{
				Index:   0,
				Role:    "user",
				Content: fmt.Sprintf("Estimate the nutritional information (calories, protein, carbs, fat) per serving for this recipe: %s", recipe),
			},
		},
	}

	llmSpan := tool.LogPrompt(prompt)

	// Make API call
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: prompt.Messages[0].Content,
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

// cookingTimeEstimatorTool estimates preparation and cooking time
func cookingTimeEstimatorTool(ctx context.Context, agent *sdk.Agent, client *openai.Client, recipe string) (string, error) {
	tool := agent.NewTool("cooking_time_estimator", "function", sdk.ToolFunction{
		Name:        "cooking_time_estimator",
		Description: "Estimates preparation and cooking time based on recipe complexity",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"recipe": map[string]string{
					"type":        "string",
					"description": "The recipe description",
				},
			},
		},
	}, map[string]string{
		"user_id": "user_67890",
	})
	defer tool.End()

	prompt := sdk.Prompt{
		Vendor: "openai",
		Mode:   "chat",
		Model:  "gpt-3.5-turbo",
		Messages: []sdk.Message{
			{
				Index:   0,
				Role:    "user",
				Content: fmt.Sprintf("Estimate the preparation time and cooking time for this recipe: %s", recipe),
			},
		},
	}

	llmSpan := tool.LogPrompt(prompt)

	// Make API call
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: prompt.Messages[0].Content,
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

func runRecipeAgent() {
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

	// Create standalone agent with association properties
	agent := traceloop.NewAgent(ctx, "recipe_generator", associationProperties)
	defer agent.End()

	// User request
	userRequest := "Create a healthy pasta dish with vegetables"
	fmt.Printf("User request: %s\n\n", userRequest)

	// Agent generates initial recipe
	fmt.Println("Generating recipe...")
	recipePrompt := sdk.Prompt{
		Vendor: "openai",
		Mode:   "chat",
		Model:  "gpt-3.5-turbo",
		Messages: []sdk.Message{
			{
				Index:   0,
				Role:    "user",
				Content: fmt.Sprintf("Create a detailed recipe for: %s. Include ingredients and instructions.", userRequest),
			},
		},
	}

	llmSpan := agent.LogPrompt(recipePrompt)

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: recipePrompt.Messages[0].Content,
			},
		},
	})
	if err != nil {
		fmt.Printf("Error generating recipe: %v\n", err)
		return
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

	recipe := resp.Choices[0].Message.Content
	fmt.Printf("\nGenerated Recipe:\n%s\n\n", recipe)

	// Tool 1: Validate ingredients
	fmt.Println("Validating ingredients...")
	validation, err := ingredientValidatorTool(ctx, agent, client, "pasta, tomatoes, spinach, garlic, olive oil")
	if err != nil {
		fmt.Printf("Warning: ingredient validation error: %v\n", err)
	} else {
		fmt.Printf("Validation Result: %s\n\n", validation)
	}

	// Tool 2: Calculate nutrition
	fmt.Println("Calculating nutrition...")
	nutrition, err := nutritionCalculatorTool(ctx, agent, client, recipe)
	if err != nil {
		fmt.Printf("Warning: nutrition calculation error: %v\n", err)
	} else {
		fmt.Printf("Nutrition Info:\n%s\n\n", nutrition)
	}

	// Tool 3: Estimate cooking time
	fmt.Println("Estimating cooking time...")
	cookingTime, err := cookingTimeEstimatorTool(ctx, agent, client, recipe)
	if err != nil {
		fmt.Printf("Warning: cooking time estimation error: %v\n", err)
	} else {
		fmt.Printf("Time Estimate:\n%s\n\n", cookingTime)
	}

	fmt.Println("=== Recipe Agent Complete ===")
}
