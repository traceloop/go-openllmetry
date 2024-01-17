package dto

import "github.com/traceloop/go-openllmetry/traceloop-sdk/model"

type PromptsResponse struct {
	Prompts 			[]model.Prompt 			`json:"prompts"`
	Environment 		string 					`json:"environment"`
}