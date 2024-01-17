package traceloop

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kluctl/go-jinja2"
	"github.com/sashabaranov/go-openai"

	"github.com/traceloop/go-openllmetry/traceloop-sdk/config"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/dto"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/model"
)

const PromptsPath = "/v1/traceloop/prompts"

type Traceloop struct {
    Config            config.Config
    PromptRegistry    model.PromptRegistry
    http.Client
}

func NewClient(config config.Config) *Traceloop {
	return &Traceloop{
		Config:         config,
		PromptRegistry: make(model.PromptRegistry),
		Client:         http.Client{},
	}
}

func (instance *Traceloop) Initialize() {
	instance.PromptRegistry = make(model.PromptRegistry)

	if instance.Config.BaseURL == "" {
		baseUrl := os.Getenv("TRACELOOP_BASE_URL")
		if baseUrl == "" {		
			instance.Config.BaseURL = "https://api.traceloop.com"
		} else {
			instance.Config.BaseURL = baseUrl
		}
	}

	if instance.Config.PollingInterval == 0 {
		pollingInterval := os.Getenv("TRACELOOP_SECONDS_POLLING_INTERVAL")
		if pollingInterval == "" {
			instance.Config.PollingInterval = 5 * time.Second
		} else {
			instance.Config.PollingInterval, _ = time.ParseDuration(pollingInterval)
		}
	}

	fmt.Printf("Traceloop %s SDK initialized. Connecting to %s\n", instance.GetVersion(), instance.Config.BaseURL)

	instance.pollPrompts()
}

func (instance *Traceloop) populatePromptRegistry() {
	resp, err := instance.fetchPathWithRetry(PromptsPath, instance.Config.BackoffConfig.MaxRetries)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	var response dto.PromptsResponse
	err = decoder.Decode(&response)
	if err != nil {
		fmt.Println(err)
	}

	for _, prompt := range response.Prompts {
		instance.PromptRegistry[prompt.Key] = &prompt
	}
}

func (instance *Traceloop) pollPrompts() {
	prompts := make(chan []model.Prompt)
    errs := make(chan error)

	instance.populatePromptRegistry()

go func() {
	defer close(prompts)
	defer close(errs)

	ticker := time.NewTicker(instance.Config.PollingInterval)

	for range ticker.C {
		instance.populatePromptRegistry()
	}
}()
}

func (instance *Traceloop) getPromptVersion(key string) (*model.PromptVersion, error) {
	if instance.PromptRegistry[key] == nil {
		return nil, fmt.Errorf("prompt with key %s not found", key)
	}

	if instance.PromptRegistry[key].Target.Version == "" {
		return nil, fmt.Errorf("prompt with key %s has no version", key)
	}

	var promptVersion model.PromptVersion
	for _, version := range instance.PromptRegistry[key].Versions {
		if version.Id == instance.PromptRegistry[key].Target.Version {
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
		Model: promptVersion.LlmConfig.Model,
		Temperature: promptVersion.LlmConfig.Temperature,
		TopP: promptVersion.LlmConfig.TopP,
		Stop: promptVersion.LlmConfig.Stop,
		FrequencyPenalty: promptVersion.LlmConfig.FrequencyPenalty,
		PresencePenalty: promptVersion.LlmConfig.PresencePenalty,
		Messages: messages,
	}, nil
}
