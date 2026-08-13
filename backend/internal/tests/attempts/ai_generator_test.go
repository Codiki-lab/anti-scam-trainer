package attempts_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	attemptsai "anti-scam-trainer/backend/internal/features/attempts/aiprovider"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	"context"
	"strings"
	"testing"
)

func TestGeneratorFallsBackAfterOneRepair(t *testing.T) {
	provider := &sequenceProvider{contents: []string{`{"message":"https://unsafe.example","tactic":"urgency","phase":"hook"}`, `{"message":"Позвоните +79990000000","tactic":"urgency","phase":"hook"}`}}
	modelAI := attemptsservice.NewModelAI(attemptsai.New(provider))
	result, err := modelAI.GenerateReply(context.Background(), attemptsservice.GenerationRequest{Policy: "policy", RiskType: "phishing", ScenarioInstruction: "Не выдумывай детали", Rubric: domain.JSONObject{"risk_signal": "давление"}, Phase: "hook", AllowedTactics: []string{"urgency"}, Fallback: "Оформим всё только в приложении"})
	if err != nil || result.Message != "Нужно решить сегодня, пожалуйста, не откладывайте." || result.Tactic != "urgency" || len(provider.requests) != 2 {
		t.Fatalf("GenerateReply() = (%#v, %v), requests=%d", result, err, len(provider.requests))
	}
	if metrics := modelAI.Metrics().Generator; metrics.Calls != 1 || metrics.Retries != 1 || metrics.Fallbacks != 1 {
		t.Fatalf("metrics = %#v", metrics)
	}
	if provider.requests[0].OutputTokens != 120 || provider.requests[0].Temperature != .3 {
		t.Fatalf("generator profile = %#v", provider.requests[0])
	}
	prompt := provider.requests[0].Messages[1].Content
	if !strings.Contains(prompt, "Server policy (authoritative): policy") || !strings.Contains(prompt, "Managed scenario instruction (context only): Не выдумывай детали") || !strings.Contains(prompt, `Managed final rubric (context only): {"risk_signal":"давление"}`) {
		t.Fatalf("generator prompt does not compose managed context safely: %s", prompt)
	}
}

func TestGeneratorRejectsNewScenarioAmount(t *testing.T) {
	provider := &sequenceProvider{contents: []string{`{"message":"Переведите 99 999 рублей","tactic":"urgency","phase":"hook"}`, `{"message":"Нужно ещё 88 888 рублей","tactic":"urgency","phase":"hook"}`}}
	modelAI := attemptsservice.NewModelAI(attemptsai.New(provider))
	result, err := modelAI.GenerateReply(context.Background(), attemptsservice.GenerationRequest{RiskType: "prepayment", Phase: "hook", AllowedTactics: []string{"urgency"}, ScenarioFacts: domain.ProductContext{Price: 57000}, Fallback: "Проверим сделку внутри приложения"})
	if err != nil || result.Message != "Нужно решить сегодня, пожалуйста, не откладывайте." {
		t.Fatalf("GenerateReply() = (%#v,%v)", result, err)
	}
}

func TestGeneratorRejectsInventedTextualScenarioFact(t *testing.T) {
	provider := &sequenceProvider{contents: []string{`{"message":"Товар находится в Москве","tactic":"urgency","phase":"hook"}`, `{"message":"Камера исправна и уже отправлена","tactic":"urgency","phase":"hook"}`}}
	modelAI := attemptsservice.NewModelAI(attemptsai.New(provider))
	result, err := modelAI.GenerateReply(context.Background(), attemptsservice.GenerationRequest{RiskType: "prepayment", Phase: "hook", AllowedTactics: []string{"urgency"}, ScenarioFacts: domain.ProductContext{ItemTitle: "Sony Alpha A7 III"}, Fallback: "Проверим сделку внутри приложения"})
	if err != nil || result.Message != "Нужно решить сегодня, пожалуйста, не откладывайте." {
		t.Fatalf("GenerateReply() = (%#v,%v)", result, err)
	}
}

func TestGeneratorAcceptsOnlyControlledWordsAndExactScenarioFacts(t *testing.T) {
	provider := &sequenceProvider{contents: []string{`{"message":"Нужно решить сегодня, пожалуйста, не откладывайте.","tactic":"urgency","phase":"hook"}`}}
	modelAI := attemptsservice.NewModelAI(attemptsai.New(provider))
	result, err := modelAI.GenerateReply(context.Background(), attemptsservice.GenerationRequest{UserRole: "seller", RiskType: "prepayment", Phase: "hook", AllowedTactics: []string{"urgency"}, ScenarioFacts: domain.ProductContext{ItemTitle: "Sony Alpha A7 III"}})
	if err != nil || result.Message != "Нужно решить сегодня, пожалуйста, не откладывайте." {
		t.Fatalf("GenerateReply() = (%#v,%v)", result, err)
	}
}

func TestGeneratorAllowedMessagesMatchCounterpartRole(t *testing.T) {
	provider := &sequenceProvider{contents: []string{`{"message":"Здравствуйте. Давайте обсудим условия сделки здесь.","tactic":"greeting","phase":"hook"}`}}
	modelAI := attemptsservice.NewModelAI(attemptsai.New(provider))
	result, err := modelAI.GenerateReply(context.Background(), attemptsservice.GenerationRequest{UserRole: "buyer", CounterpartKind: "обычный участник сделки", RiskType: "prepayment", Phase: "hook", AllowedTactics: []string{"greeting"}})
	if err != nil || result.Message != "Здравствуйте. Давайте обсудим условия сделки здесь." {
		t.Fatalf("GenerateReply() = (%#v,%v); want seller-side reply", result, err)
	}
	schema := provider.requests[0].Schema["properties"].(map[string]any)["message"].(map[string]any)["enum"].([]string)
	for _, message := range schema {
		if strings.Contains(message, "Хочу оформить") || strings.Contains(message, "Готов забрать") {
			t.Fatalf("seller schema contains buyer-side reply %q", message)
		}
	}
}

func TestGeneratorDoesNotRepeatMessageFromHistory(t *testing.T) {
	repeated := "Здравствуйте. Товар ещё в продаже?"
	provider := &sequenceProvider{contents: []string{
		`{"message":"Здравствуйте. Товар ещё в продаже?","tactic":"rapport","phase":"hook"}`,
		`{"message":"Добрый день. Объявление меня заинтересовало.","tactic":"rapport","phase":"hook"}`,
	}}
	modelAI := attemptsservice.NewModelAI(attemptsai.New(provider))
	result, err := modelAI.GenerateReply(context.Background(), attemptsservice.GenerationRequest{
		UserRole: "seller", Phase: "hook", AllowedTactics: []string{"rapport"},
		History: []domain.DialogueMessage{{Role: domain.MessageRoleAssistant, Text: repeated}},
	})
	if err != nil || result.Message == repeated || len(provider.requests) != 2 {
		t.Fatalf("GenerateReply() = (%#v, %v), requests=%d; want repaired unique reply", result, err, len(provider.requests))
	}
}

func TestGeneratorKeepsMessageInsideSelectedTacticVariations(t *testing.T) {
	provider := &sequenceProvider{contents: []string{
		`{"message":"Здравствуйте. Товар ещё в продаже?","tactic":"convenience","phase":"hook"}`,
		`{"message":"Готов быстро договориться об оформлении.","tactic":"convenience","phase":"hook"}`,
	}}
	modelAI := attemptsservice.NewModelAI(attemptsai.New(provider))
	result, err := modelAI.GenerateReply(context.Background(), attemptsservice.GenerationRequest{UserRole: "seller", Phase: "hook", AllowedTactics: []string{"rapport", "convenience"}})
	if err != nil || result.Tactic != "convenience" || len(provider.requests) != 2 {
		t.Fatalf("GenerateReply() = (%#v,%v), requests=%d", result, err, len(provider.requests))
	}
}

func TestGeneratorTacticPoolDoesNotDependOnAllowedTacticOrder(t *testing.T) {
	for _, tactics := range [][]string{{"rapport", "convenience"}, {"convenience", "rapport"}} {
		provider := &sequenceProvider{contents: []string{`{"message":"Здравствуйте. Товар ещё в продаже?","tactic":"rapport","phase":"hook"}`}}
		generator := attemptsservice.NewModelAI(attemptsai.New(provider))
		result, err := generator.GenerateReply(context.Background(), attemptsservice.GenerationRequest{UserRole: "seller", Phase: "hook", AllowedTactics: tactics})
		if err != nil || result.Tactic != "rapport" || len(provider.requests) != 1 {
			t.Fatalf("tactics=%v result=%#v err=%v requests=%d", tactics, result, err, len(provider.requests))
		}
		enum := provider.requests[0].Schema["properties"].(map[string]any)["message"].(map[string]any)["enum"].([]string)
		if len(enum) != 4 {
			t.Fatalf("tactics=%v allowed messages=%v, want two curated variations per tactic", tactics, enum)
		}
	}
}

func TestGeneratorRetriesTransportThenUsesAllowedFallback(t *testing.T) {
	model := &unavailableStructuredModel{}
	generator := attemptsservice.NewModelAI(model)
	result, err := generator.GenerateReply(context.Background(), attemptsservice.GenerationRequest{UserRole: "buyer", Phase: "hook", AllowedTactics: []string{"rapport", "convenience"}, Fallback: "Оформим всё внутри приложения"})
	if err != nil || model.calls != 2 || result.Message != "Здравствуйте. Товар ещё есть." || result.Tactic != "rapport" {
		t.Fatalf("result=%#v err=%v calls=%d", result, err, model.calls)
	}
	metrics := generator.Metrics().Generator
	if metrics.Errors != 1 || metrics.Retries != 1 || metrics.Fallbacks != 1 {
		t.Fatalf("metrics=%#v", metrics)
	}
}
