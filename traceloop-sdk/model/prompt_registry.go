package model

import "time"

type ModelConfig struct {
	Mode             string   `json:"mode"`
	Model            string   `json:"model"`
	Temperature      float32  `json:"temperature"`
	TopP             float32  `json:"top_p"`
	Stop             []string `json:"stop"`
	FrequencyPenalty float32  `json:"frequency_penalty"`
	PresencePenalty  float32  `json:"presence_penalty"`
}

type Message struct {
	Index     int      `json:"index"`
	Role      string   `json:"role"`
	Template  string   `json:"template"`
	Variables []string `json:"variables"`
}

type PromptVersion struct {
	Id               string      `json:"id"`
	Hash             string      `json:"hash"`
	Version          uint        `json:"version"`
	Name             string      `json:"name"`
	CreatedAt        time.Time   `json:"created_at"`
	Provider         string      `json:"provider"`
	TemplatingEngine string      `json:"templating_engine"`
	Messages         []Message   `json:"messages"`
	LlmConfig        ModelConfig `json:"llm_config"`
}

type Target struct {
	Id        string    `json:"id"`
	PromptId  string    `json:"prompt_id"`
	Version   string    `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Prompt struct {
	Id        string          `json:"id"`
	Versions  []PromptVersion `json:"versions"`
	Target    Target          `json:"target"`
	Key       string          `json:"key"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type PromptRegistry map[string]*Prompt
