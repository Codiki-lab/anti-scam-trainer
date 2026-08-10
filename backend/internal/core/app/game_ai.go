package app

import (
	"anti-scam-trainer/backend/internal/core/aiprovider"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	"context"
	"errors"
)

type gameAIAdapter struct{ provider aiprovider.Provider }

func (a gameAIAdapter) GenerateStructured(ctx context.Context, input attemptsservice.StructuredModelRequest) (string, error) {
	messages := make([]aiprovider.Message, 0, len(input.Messages))
	for _, message := range input.Messages {
		messages = append(messages, aiprovider.Message{Role: aiprovider.Role(message.Role), Content: message.Content})
	}
	result, err := a.provider.GenerateStructured(ctx, aiprovider.StructuredRequest{
		Messages: messages, Schema: input.Schema, OutputTokens: input.OutputTokens,
		Temperature: input.Temperature, TopP: input.TopP, TopK: input.TopK, RepeatPenalty: input.RepeatPenalty,
	})
	if err != nil {
		return "", mapAIError(err)
	}
	return result.Content, nil
}

func mapAIError(err error) error {
	var capacity *aiprovider.ContextCapacityError
	var transport *aiprovider.TransportError
	var runtime *aiprovider.OllamaError
	var malformed *aiprovider.MalformedResponseError
	switch {
	case errors.As(err, &capacity):
		return attemptsservice.ErrAIContextExhausted
	case errors.As(err, &transport), errors.As(err, &runtime):
		return attemptsservice.ErrAIUnavailable
	case errors.As(err, &malformed):
		return attemptsservice.ErrAIInvalidResponse
	default:
		return err
	}
}
