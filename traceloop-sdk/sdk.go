package sdk

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

func NewTraceloop(config config.Config) *Traceloop {
	return &Traceloop{
		Config:         config,
		PromptRegistry: make(model.PromptRegistry),
		Client:         http.Client{},
	}
}

func (sdk *Traceloop) Initialize() {
	sdk.PromptRegistry = make(model.PromptRegistry)

	if sdk.Config.BaseURL == "" {
		baseUrl := os.Getenv("TRACELOOP_BASE_URL")
		if baseUrl == "" {		
			sdk.Config.BaseURL = "https://api.traceloop.com"
		} else {
			sdk.Config.BaseURL = baseUrl
		}
	}

	if sdk.Config.PollingInterval == 0 {
		pollingInterval := os.Getenv("TRACELOOP_SECONDS_POLLING_INTERVAL")
		if pollingInterval == "" {
			sdk.Config.PollingInterval = 5 * time.Second
		} else {
			sdk.Config.PollingInterval, _ = time.ParseDuration(pollingInterval)
		}
	}

	fmt.Printf("Traceloop %s SDK initialized. Connecting to %s\n", sdk.GetVersion(), sdk.Config.BaseURL)

	sdk.pollPrompts()
}

func (sdk *Traceloop) populatePromptRegistry() {
	resp, err := sdk.fetchPathWithRetry(PromptsPath, sdk.Config.BackoffConfig.MaxRetries)
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
		sdk.PromptRegistry[prompt.Key] = &prompt
	}
}

func (sdk *Traceloop) pollPrompts() {
	prompts := make(chan []model.Prompt)
    errs := make(chan error)

	sdk.populatePromptRegistry()

go func() {
	defer close(prompts)
	defer close(errs)

	ticker := time.NewTicker(sdk.Config.PollingInterval)

	for range ticker.C {
		sdk.populatePromptRegistry()
	}
}()
}

func (sdk *Traceloop) getPromptVersion(key string) (*model.PromptVersion, error) {
	if sdk.PromptRegistry[key] == nil {
		return nil, fmt.Errorf("prompt with key %s not found", key)
	}

	if sdk.PromptRegistry[key].Target.Version == "" {
		return nil, fmt.Errorf("prompt with key %s has no version", key)
	}

	var promptVersion model.PromptVersion
	for _, version := range sdk.PromptRegistry[key].Versions {
		if version.Id == sdk.PromptRegistry[key].Target.Version {
			promptVersion = version
		}
	}

	if promptVersion.Id == "" {
		return nil, fmt.Errorf("prompt version was not found")
	}

	return &promptVersion, nil
}

func (sdk *Traceloop) GetOpenAIChatCompletionRequest(key string, variables map[string]any) (*openai.ChatCompletionRequest, error) {
	promptVersion, err := sdk.getPromptVersion(key)
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
