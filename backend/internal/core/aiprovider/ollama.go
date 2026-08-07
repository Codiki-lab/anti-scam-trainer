// Package aiprovider defines the infrastructure boundary for local generative models.
package aiprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

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
}

type Config struct {
	URL                 string
	Model               string
	RequestTimeout      time.Duration
	ContextWindowTokens int
	OutputReserveTokens int
	MediumRiskThreshold float64
	HighRiskThreshold   float64
}

// TransportError is an unavailable runtime, network failure, or request timeout.
type TransportError struct {
	Err error
}

func (e *TransportError) Error() string { return "ollama transport failure: " + e.Err.Error() }
func (e *TransportError) Unwrap() error { return e.Err }

// OllamaError is a non-success response returned by the Ollama runtime.
type OllamaError struct {
	StatusCode int
	Body       string
}

func (e *OllamaError) Error() string {
	return fmt.Sprintf("ollama returned HTTP %d: %s", e.StatusCode, e.Body)
}

// MalformedResponseError means a success response cannot satisfy Provider's contract.
type MalformedResponseError struct {
	Err error
}

func (e *MalformedResponseError) Error() string { return "invalid ollama response: " + e.Err.Error() }
func (e *MalformedResponseError) Unwrap() error { return e.Err }

// ContextCapacityError is returned before a network call when the requested output
// reserve cannot fit alongside the estimated prompt in the configured context window.
type ContextCapacityError struct {
	EstimatedPromptTokens int
	ReservedOutputTokens  int
	ContextWindowTokens   int
}

func (e *ContextCapacityError) Error() string {
	return fmt.Sprintf("estimated prompt (%d) plus output reserve (%d) exceeds context window (%d)", e.EstimatedPromptTokens, e.ReservedOutputTokens, e.ContextWindowTokens)
}

type Ollama struct {
	url                 string
	model               string
	client              *http.Client
	contextWindowTokens int
	outputReserveTokens int
	mediumRiskThreshold float64
	highRiskThreshold   float64
}

func NewOllama(config Config) (Provider, error) {
	if config.URL == "" {
		return nil, errors.New("ollama URL is required")
	}
	parsedURL, err := url.Parse(config.URL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("invalid ollama URL: %q", config.URL)
	}
	if config.Model == "" {
		return nil, errors.New("ollama model is required")
	}
	if config.ContextWindowTokens <= 0 {
		return nil, errors.New("ollama context window must be positive")
	}
	if config.RequestTimeout <= 0 {
		config.RequestTimeout = 30 * time.Second
	}
	if config.MediumRiskThreshold == 0 {
		config.MediumRiskThreshold = 0.60
	}
	if config.HighRiskThreshold == 0 {
		config.HighRiskThreshold = 0.75
	}
	if config.MediumRiskThreshold <= 0 || config.MediumRiskThreshold >= config.HighRiskThreshold || config.HighRiskThreshold >= 1 {
		return nil, errors.New("ollama context-risk thresholds must be ordered between 0 and 1")
	}
	if config.OutputReserveTokens == 0 {
		config.OutputReserveTokens = max(256, int(math.Ceil(float64(config.ContextWindowTokens)*0.20)))
	}
	if config.OutputReserveTokens <= 0 || config.OutputReserveTokens >= config.ContextWindowTokens {
		return nil, errors.New("ollama output reserve must be positive and smaller than the context window")
	}
	return &Ollama{
		url:                 strings.TrimRight(config.URL, "/"),
		model:               config.Model,
		client:              &http.Client{Timeout: config.RequestTimeout},
		contextWindowTokens: config.ContextWindowTokens,
		outputReserveTokens: config.OutputReserveTokens,
		mediumRiskThreshold: config.MediumRiskThreshold,
		highRiskThreshold:   config.HighRiskThreshold,
	}, nil
}

func (o *Ollama) Generate(ctx context.Context, messages []Message) (Result, error) {
	if err := validateMessages(messages); err != nil {
		return Result{}, err
	}
	estimatedPromptTokens := EstimateLlama32PromptTokens(messages)
	if estimatedPromptTokens+o.outputReserveTokens > o.contextWindowTokens {
		return Result{}, &ContextCapacityError{EstimatedPromptTokens: estimatedPromptTokens, ReservedOutputTokens: o.outputReserveTokens, ContextWindowTokens: o.contextWindowTokens}
	}
	requestBody, err := json.Marshal(ollamaRequest{Model: o.model, Messages: messages, Stream: false, Options: ollamaOptions{ContextWindowTokens: o.contextWindowTokens}})
	if err != nil {
		return Result{}, fmt.Errorf("encode ollama request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, o.url+"/api/chat", bytes.NewReader(requestBody))
	if err != nil {
		return Result{}, fmt.Errorf("create ollama request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := o.client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return Result{}, ctx.Err()
		}
		return Result{}, &TransportError{Err: err}
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return Result{}, &TransportError{Err: fmt.Errorf("read ollama response: %w", err)}
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return Result{}, &OllamaError{StatusCode: response.StatusCode, Body: string(responseBody)}
	}
	var decoded ollamaResponse
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return Result{}, &MalformedResponseError{Err: err}
	}
	if err := decoded.validate(); err != nil {
		return Result{}, &MalformedResponseError{Err: err}
	}
	usageFraction := float64(estimatedPromptTokens+o.outputReserveTokens) / float64(o.contextWindowTokens)
	return Result{Content: *decoded.Message.Content, Usage: Usage{
		PromptTokens:          decoded.PromptEvalCount.Value,
		CompletionTokens:      decoded.EvalCount.Value,
		EstimatedPromptTokens: estimatedPromptTokens,
		ReservedOutputTokens:  o.outputReserveTokens,
		ContextWindowTokens:   o.contextWindowTokens,
		ContextUsage:          usageFraction,
		ContextRisk:           o.contextRisk(usageFraction),
	}}, nil
}

func (o *Ollama) contextRisk(usage float64) ContextRisk {
	if usage < o.mediumRiskThreshold {
		return ContextRiskLow
	}
	if usage < o.highRiskThreshold {
		return ContextRiskMedium
	}
	return ContextRiskHigh
}

func validateMessages(messages []Message) error {
	if len(messages) == 0 {
		return errors.New("at least one AI message is required")
	}
	for i, message := range messages {
		if message.Role != RoleSystem && message.Role != RoleUser && message.Role != RoleAssistant {
			return fmt.Errorf("message %d has unsupported role %q", i, message.Role)
		}
		if strings.TrimSpace(message.Content) == "" {
			return fmt.Errorf("message %d has empty content", i)
		}
	}
	return nil
}

type ollamaRequest struct {
	Model    string        `json:"model"`
	Messages []Message     `json:"messages"`
	Stream   bool          `json:"stream"`
	Options  ollamaOptions `json:"options"`
}

type ollamaOptions struct {
	ContextWindowTokens int `json:"num_ctx"`
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

// EstimateLlama32PromptTokens returns a conservative preflight estimate for the
// configured local llama3.2:3b runtime. It counts UTF-8 bytes plus fixed chat
// framing, so it intentionally overestimates many English and Russian prompts.
// Ollama's prompt_eval_count remains the authoritative post-generation value.
func EstimateLlama32PromptTokens(messages []Message) int {
	tokens := 2 // chat start and generation prompt framing
	for _, message := range messages {
		tokens += 4 + len([]byte(message.Role)) + len([]byte(message.Content))
	}
	return tokens
}
