package attempts_test

import (
	"anti-scam-trainer/backend/internal/core/aiprovider"
	"context"
)

type sequenceProvider struct {
	contents []string
	requests []aiprovider.StructuredRequest
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
