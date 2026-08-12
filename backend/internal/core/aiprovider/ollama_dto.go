package aiprovider

import (
	"encoding/json"
	"errors"
)

type ollamaRequest struct {
	Model    string         `json:"model"`
	Messages []Message      `json:"messages"`
	Stream   bool           `json:"stream"`
	Format   map[string]any `json:"format,omitempty"`
	Think    bool           `json:"think"`
	Options  ollamaOptions  `json:"options"`
}

type ollamaOptions struct {
	ContextWindowTokens int     `json:"num_ctx"`
	OutputTokens        int     `json:"num_predict"`
	Temperature         float64 `json:"temperature"`
	TopP                float64 `json:"top_p,omitempty"`
	TopK                int     `json:"top_k,omitempty"`
	RepeatPenalty       float64 `json:"repeat_penalty,omitempty"`
}

type optionalInt struct {
	Value int
	Set   bool
}

func (i *optionalInt) UnmarshalJSON(data []byte) error {
	i.Set = true
	return json.Unmarshal(data, &i.Value)
}

type ollamaResponse struct {
	Message struct {
		Role    string  `json:"role"`
		Content *string `json:"content"`
	} `json:"message"`
	Done            bool        `json:"done"`
	PromptEvalCount optionalInt `json:"prompt_eval_count"`
	EvalCount       optionalInt `json:"eval_count"`
}

func (r ollamaResponse) validate() error {
	if !r.Done {
		return errors.New("response is not complete")
	}
	if r.Message.Role != string(RoleAssistant) || r.Message.Content == nil {
		return errors.New("response has no assistant message")
	}
	if !r.PromptEvalCount.Set || !r.EvalCount.Set || r.PromptEvalCount.Value < 0 || r.EvalCount.Value < 0 {
		return errors.New("response has invalid token usage")
	}
	return nil
}
