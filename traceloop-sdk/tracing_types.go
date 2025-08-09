package traceloop

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

type ToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type Message struct {
	Index     int        `json:"index"`
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type Prompt struct {
	Vendor           string    `json:"vendor"`
	Model            string    `json:"model"`
	Mode             string    `json:"mode"`
	Temperature      float32   `json:"temperature"`
	TopP             float32   `json:"top_p"`
	Stop             []string  `json:"stop"`
	FrequencyPenalty float32   `json:"frequency_penalty"`
	PresencePenalty  float32   `json:"presence_penalty"`
	Messages         []Message `json:"messages"`
	Tools            []Tool    `json:"tools,omitempty"`
}

type Completion struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type WorkflowAttributes struct {
	Name                  string            `json:"workflow_name"`
	AssociationProperties map[string]string `json:"association_properties"`
}

type Usage struct {
	TotalTokens      int `json:"total_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
}
