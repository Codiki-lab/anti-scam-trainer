// Package aiprovider defines the infrastructure boundary for local generative models.
package aiprovider

import "context"

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message is one ordered conversation message supplied by a caller.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type ContextRisk string

const (
	ContextRiskLow    ContextRisk = "low"
	ContextRiskMedium ContextRisk = "medium"
	ContextRiskHigh   ContextRisk = "high"
)

type Usage struct {
	PromptTokens          int         `json:"prompt_tokens"`
	CompletionTokens      int         `json:"completion_tokens"`
	EstimatedPromptTokens int         `json:"estimated_prompt_tokens"`
	ReservedOutputTokens  int         `json:"reserved_output_tokens"`
	ContextWindowTokens   int         `json:"context_window_tokens"`
	ContextUsage          float64     `json:"context_usage"`
	ContextRisk           ContextRisk `json:"context_risk"`
}

type Result struct {
	Content string `json:"content"`
	Usage   Usage  `json:"usage"`
}

// Provider is independent of the wire protocol used by an AI runtime.
type Provider interface {
	Generate(context.Context, []Message) (Result, error)
	GenerateStructured(context.Context, StructuredRequest) (Result, error)
}

type StructuredRequest struct {
	Messages      []Message
	Schema        map[string]any
	OutputTokens  int
	Temperature   float64
	TopP          float64
	TopK          int
	RepeatPenalty float64
}
