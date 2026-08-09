package app

import (
	"anti-scam-trainer/backend/internal/core/aiprovider"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	"context"
	"errors"
)

type gameAIAdapter struct{ provider aiprovider.Provider }

func (a gameAIAdapter) Generate(ctx context.Context, input []attemptsservice.AIMessage) (string, error) {
	messages := make([]aiprovider.Message, len(input))
	for i, message := range input {
		messages[i] = aiprovider.Message{Role: aiprovider.Role(message.Role), Content: message.Content}
	}
	result, err := a.provider.Generate(ctx, messages)
	if err == nil {
		return result.Content, nil
	}
	var capacity *aiprovider.ContextCapacityError
	var transport *aiprovider.TransportError
	var runtime *aiprovider.OllamaError
	var malformed *aiprovider.MalformedResponseError
	switch {
	case errors.As(err, &capacity):
		return "", attemptsservice.ErrAIContextExhausted
	case errors.As(err, &transport), errors.As(err, &runtime):
		return "", attemptsservice.ErrAIUnavailable
	case errors.As(err, &malformed):
		return "", attemptsservice.ErrAIInvalidResponse
	default:
		return "", err
	}
}
