package traceloop

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kluctl/go-jinja2"
	"github.com/sashabaranov/go-openai"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/model"
)

type PromptsResponse struct {
	Prompts     []model.Prompt `json:"prompts"`
	Environment string         `json:"environment"`
}

func (instance *Traceloop) populatePromptRegistry() {
	resp, err := instance.fetchPathWithRetry(PromptsPath, instance.config.BackoffConfig.MaxRetries)
	if err != nil {
		fmt.Println("Failed to fetch prompts", err)
		return
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	var response PromptsResponse
	err = decoder.Decode(&response)
	if err != nil {
		fmt.Println("Failed to decode response", err)
		return
	}

	for _, prompt := range response.Prompts {
		instance.promptRegistry[prompt.Key] = &prompt
	}
}

func (instance *Traceloop) pollPrompts() {
	prompts := make(chan []model.Prompt)
	errs := make(chan error)

	instance.populatePromptRegistry()

	go func() {
		defer close(prompts)
		defer close(errs)

		ticker := time.NewTicker(instance.config.PollingInterval)

		for range ticker.C {
			instance.populatePromptRegistry()
		}
	}()
}

func (instance *Traceloop) getPromptVersion(key string) (*model.PromptVersion, error) {
	if instance.promptRegistry[key] == nil {
		return nil, fmt.Errorf("prompt with key %s not found", key)
	}

	if instance.promptRegistry[key].Target.Version == "" {
		return nil, fmt.Errorf("prompt with key %s has no version", key)
	}

	var promptVersion model.PromptVersion
	for _, version := range instance.promptRegistry[key].Versions {
		if version.Id == instance.promptRegistry[key].Target.Version {
			promptVersion = version
		}
	}

	if promptVersion.Id == "" {
		return nil, fmt.Errorf("prompt version was not found")
	}

	return &promptVersion, nil
}

func (instance *Traceloop) GetOpenAIChatCompletionRequest(key string, variables map[string]any) (*openai.ChatCompletionRequest, error) {
	promptVersion, err := instance.getPromptVersion(key)
	if err != nil {
		return nil, err
	}

	jinjaRenderer, err := jinja2.NewJinja2("renderer", 1, jinja2.WithGlobals(variables))
	if err != nil {
		return nil, err
	}

	var messages []openai.ChatCompletionMessage

	for _, message := range promptVersion.Messages {
		renderedMessage, err := jinjaRenderer.RenderString(message.Template)
		if err != nil {
			return nil, err
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    message.Role,
			Content: renderedMessage,
		})
	}

	return &openai.ChatCompletionRequest{
		Model:            promptVersion.LlmConfig.Model,
		Temperature:      promptVersion.LlmConfig.Temperature,
		TopP:             promptVersion.LlmConfig.TopP,
		Stop:             promptVersion.LlmConfig.Stop,
		FrequencyPenalty: promptVersion.LlmConfig.FrequencyPenalty,
		PresencePenalty:  promptVersion.LlmConfig.PresencePenalty,
		Messages:         messages,
	}, nil
}
