package attempts_test

import (
	"anti-scam-trainer/backend/internal/core/aiprovider"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	"context"
	"errors"
)

type sequenceProvider struct {
	contents []string
	requests []aiprovider.StructuredRequest
}

type unavailableStructuredModel struct{ calls int }

func (m *unavailableStructuredModel) GenerateStructured(context.Context, attemptsservice.StructuredModelRequest) (string, error) {
	m.calls++
	return "", errors.New("transport unavailable")
}

func (p *sequenceProvider) Generate(context.Context, []aiprovider.Message) (aiprovider.Result, error) {
	return aiprovider.Result{}, nil
}

func (p *sequenceProvider) GenerateStructured(_ context.Context, request aiprovider.StructuredRequest) (aiprovider.Result, error) {
	p.requests = append(p.requests, request)
	content := p.contents[0]
	p.contents = p.contents[1:]
	return aiprovider.Result{Content: content}, nil
}
