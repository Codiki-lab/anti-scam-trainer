package aiprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func (o *Ollama) Generate(ctx context.Context, messages []Message) (Result, error) {
	return o.generate(ctx, StructuredRequest{Messages: messages, OutputTokens: o.outputReserveTokens})
}

func (o *Ollama) GenerateStructured(ctx context.Context, input StructuredRequest) (Result, error) {
	if len(input.Schema) == 0 {
		return Result{}, errors.New("structured JSON schema is required")
	}
	if input.OutputTokens <= 0 || input.OutputTokens >= o.contextWindowTokens {
		return Result{}, errors.New("structured output token limit is invalid")
	}
	return o.generate(ctx, input)
}

func (o *Ollama) generate(ctx context.Context, input StructuredRequest) (Result, error) {
	messages := input.Messages
	if err := validateMessages(messages); err != nil {
		return Result{}, err
	}
	outputTokens := input.OutputTokens
	if outputTokens == 0 {
		outputTokens = o.outputReserveTokens
	}
	estimatedPromptTokens := EstimatePromptTokens(messages)
	if estimatedPromptTokens+outputTokens > int(float64(o.contextWindowTokens)*o.highRiskThreshold) {
		return Result{}, &ContextCapacityError{EstimatedPromptTokens: estimatedPromptTokens, ReservedOutputTokens: outputTokens, ContextWindowTokens: o.contextWindowTokens}
	}
	requestBody, err := json.Marshal(ollamaRequest{Model: o.model, Messages: messages, Stream: false, Format: input.Schema, Think: false, Options: ollamaOptions{ContextWindowTokens: o.contextWindowTokens, OutputTokens: outputTokens, Temperature: input.Temperature, TopP: input.TopP, TopK: input.TopK, RepeatPenalty: input.RepeatPenalty}})
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
	responseBody, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if readErr != nil {
		return Result{}, &TransportError{Err: fmt.Errorf("read ollama response: %w", readErr)}
	}
	if closeErr != nil {
		return Result{}, &TransportError{Err: fmt.Errorf("close ollama response: %w", closeErr)}
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
	usageFraction := float64(estimatedPromptTokens+outputTokens) / float64(o.contextWindowTokens)
	return Result{Content: *decoded.Message.Content, Usage: Usage{
		PromptTokens:          decoded.PromptEvalCount.Value,
		CompletionTokens:      decoded.EvalCount.Value,
		EstimatedPromptTokens: estimatedPromptTokens,
		ReservedOutputTokens:  outputTokens,
		ContextWindowTokens:   o.contextWindowTokens,
		ContextUsage:          usageFraction,
		ContextRisk:           o.contextRisk(usageFraction),
	}}, nil
}
