package app

import (
	"anti-scam-trainer/backend/internal/core/aiprovider"
	"anti-scam-trainer/backend/internal/core/domain"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	"context"
	"testing"
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

func TestEvaluatorRepairsOnceAndKeepsItsOwnProfile(t *testing.T) {
	provider := &sequenceProvider{contents: []string{`{"score":9}`, `{"score":4,"is_safe":true,"risk_type":"phishing","detected_signals":[],"evaluation":"Безопасный отказ","safe_action":"Проверить заказ в приложении"}`}}
	modelAI := attemptsservice.NewModelAI(gameAIAdapter{provider: provider})
	result, err := modelAI.Evaluate(context.Background(), attemptsservice.EvaluationRequest{Policy: "policy", RiskType: "phishing", EvaluationContext: "context", Answer: "Не перейду"})
	if err != nil || result.Score != 4 || len(provider.requests) != 2 {
		t.Fatalf("Evaluate() = (%#v, %v), requests=%d", result, err, len(provider.requests))
	}
	for _, request := range provider.requests {
		if request.OutputTokens != 240 || request.Temperature != 0 || request.Schema == nil {
			t.Fatalf("evaluator profile = %#v", request)
		}
	}
}

func TestGeneratorFallsBackAfterOneRepair(t *testing.T) {
	provider := &sequenceProvider{contents: []string{`{"message":"https://unsafe.example","tactic":"urgency","phase":"hook"}`, `{"message":"Позвоните +79990000000","tactic":"urgency","phase":"hook"}`}}
	modelAI := attemptsservice.NewModelAI(gameAIAdapter{provider: provider})
	result, err := modelAI.GenerateReply(context.Background(), attemptsservice.GenerationRequest{RiskType: "phishing", Phase: "hook", AllowedTactics: []string{"urgency"}, Fallback: "Оформим всё только в приложении"})
	if err != nil || result.Message != "Оформим всё только в приложении" || len(provider.requests) != 2 {
		t.Fatalf("GenerateReply() = (%#v, %v), requests=%d", result, err, len(provider.requests))
	}
	if provider.requests[0].OutputTokens != 120 || provider.requests[0].Temperature != .3 {
		t.Fatalf("generator profile = %#v", provider.requests[0])
	}
}

func TestGeneratorRejectsNewScenarioAmount(t *testing.T) {
	provider := &sequenceProvider{contents: []string{`{"message":"Переведите 99 999 рублей","tactic":"urgency","phase":"hook"}`, `{"message":"Нужно ещё 88 888 рублей","tactic":"urgency","phase":"hook"}`}}
	modelAI := attemptsservice.NewModelAI(gameAIAdapter{provider: provider})
	result, err := modelAI.GenerateReply(context.Background(), attemptsservice.GenerationRequest{RiskType: "prepayment", Phase: "hook", AllowedTactics: []string{"urgency"}, ScenarioFacts: domain.ProductContext{Price: 57000}, Fallback: "Проверим сделку внутри приложения"})
	if err != nil || result.Message != "Проверим сделку внутри приложения" {
		t.Fatalf("GenerateReply() = (%#v,%v)", result, err)
	}
}

func TestGeneratorRejectsInventedTextualScenarioFact(t *testing.T) {
	provider := &sequenceProvider{contents: []string{`{"message":"Товар находится в Москве","tactic":"urgency","phase":"hook"}`, `{"message":"Камера исправна и уже отправлена","tactic":"urgency","phase":"hook"}`}}
	modelAI := attemptsservice.NewModelAI(gameAIAdapter{provider: provider})
	result, err := modelAI.GenerateReply(context.Background(), attemptsservice.GenerationRequest{RiskType: "prepayment", Phase: "hook", AllowedTactics: []string{"urgency"}, ScenarioFacts: domain.ProductContext{ItemTitle: "Sony Alpha A7 III"}, Fallback: "Проверим сделку внутри приложения"})
	if err != nil || result.Message != "Проверим сделку внутри приложения" {
		t.Fatalf("GenerateReply() = (%#v,%v)", result, err)
	}
}

func TestGeneratorAcceptsOnlyControlledWordsAndExactScenarioFacts(t *testing.T) {
	provider := &sequenceProvider{contents: []string{`{"message":"Добрый день. Хочу оформить сделку сегодня.","tactic":"urgency","phase":"hook"}`}}
	modelAI := attemptsservice.NewModelAI(gameAIAdapter{provider: provider})
	result, err := modelAI.GenerateReply(context.Background(), attemptsservice.GenerationRequest{RiskType: "prepayment", Phase: "hook", AllowedTactics: []string{"urgency"}, ScenarioFacts: domain.ProductContext{ItemTitle: "Sony Alpha A7 III"}})
	if err != nil || result.Message != "Добрый день. Хочу оформить сделку сегодня." {
		t.Fatalf("GenerateReply() = (%#v,%v)", result, err)
	}
}
