package attempts_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	"context"
	"testing"
)

func TestGeneratorUsesScenarioReplyWithoutModelCall(t *testing.T) {
	model := &unavailableStructuredModel{}
	generator := attemptsservice.NewModelAI(model)

	result, err := generator.GenerateReply(context.Background(), attemptsservice.GenerationRequest{UserRole: "buyer", Phase: "escalation", AllowedTactics: []string{"urgency"}, Fallback: "Ссылка действует несколько минут, потом телефон заберёт другой покупатель."})

	if err != nil || result.Message != "Ссылка действует несколько минут, потом телефон заберёт другой покупатель." || result.Tactic != "urgency" || result.Phase != "escalation" || model.calls != 0 {
		t.Fatalf("GenerateReply() = (%#v, %v), model calls=%d", result, err, model.calls)
	}
	if metrics := generator.Metrics().Generator; metrics.Calls != 1 || metrics.Errors != 0 || metrics.Retries != 0 || metrics.Fallbacks != 0 {
		t.Fatalf("metrics = %#v", metrics)
	}
}

func TestGeneratorRejectsUnsafeScenarioReplyWithoutModelCall(t *testing.T) {
	model := &unavailableStructuredModel{}
	generator := attemptsservice.NewModelAI(model)

	result, err := generator.GenerateReply(context.Background(), attemptsservice.GenerationRequest{UserRole: "buyer", Phase: "escalation", AllowedTactics: []string{"urgency"}, Fallback: "Откройте https://unsafe.example"})

	if err != nil || result.Message != "Нужно решить сегодня, пожалуйста, не откладывайте." || result.Tactic != "urgency" || model.calls != 0 {
		t.Fatalf("GenerateReply() = (%#v, %v), model calls=%d", result, err, model.calls)
	}
}

func TestGeneratorDoesNotRepeatScenarioOrCuratedReply(t *testing.T) {
	model := &unavailableStructuredModel{}
	generator := attemptsservice.NewModelAI(model)
	history := []domain.DialogueMessage{
		{Role: domain.MessageRoleAssistant, Text: "Ссылка действует несколько минут."},
		{Role: domain.MessageRoleAssistant, Text: "Нужно решить сегодня, пожалуйста, не откладывайте."},
	}

	result, err := generator.GenerateReply(context.Background(), attemptsservice.GenerationRequest{UserRole: "buyer", Phase: "escalation", AllowedTactics: []string{"urgency"}, Fallback: "Ссылка действует несколько минут.", History: history})

	if err != nil || result.Message != "Давайте подтвердим решение прямо сейчас." || model.calls != 0 {
		t.Fatalf("GenerateReply() = (%#v, %v), model calls=%d", result, err, model.calls)
	}
}

func TestGeneratorKeepsCounterpartRoleInCuratedReply(t *testing.T) {
	model := &unavailableStructuredModel{}
	generator := attemptsservice.NewModelAI(model)

	honest, honestErr := generator.GenerateReply(context.Background(), attemptsservice.GenerationRequest{UserRole: "buyer", CounterpartKind: "обычный участник сделки", Phase: "hook", AllowedTactics: []string{"greeting"}})
	scam, scamErr := generator.GenerateReply(context.Background(), attemptsservice.GenerationRequest{UserRole: "seller", Phase: "hook", AllowedTactics: []string{"rapport"}})

	if honestErr != nil || honest.Message != "Здравствуйте. Давайте обсудим условия сделки здесь." {
		t.Fatalf("honest counterpart = (%#v, %v)", honest, honestErr)
	}
	if scamErr != nil || scam.Message != "Здравствуйте. Товар ещё в продаже?" {
		t.Fatalf("scam counterpart = (%#v, %v)", scam, scamErr)
	}
	if model.calls != 0 {
		t.Fatalf("model calls=%d, want 0", model.calls)
	}
}

func TestGeneratorRejectsMissingTactics(t *testing.T) {
	generator := attemptsservice.NewModelAI(&unavailableStructuredModel{})
	if _, err := generator.GenerateReply(context.Background(), attemptsservice.GenerationRequest{Phase: "hook"}); err == nil {
		t.Fatal("GenerateReply() succeeded without allowed tactics")
	}
}

func TestGeneratorUsesScenarioReplyWithoutCuratedTacticPool(t *testing.T) {
	generator := attemptsservice.NewModelAI(&unavailableStructuredModel{})
	result, err := generator.GenerateReply(context.Background(), attemptsservice.GenerationRequest{Phase: "hook", AllowedTactics: []string{"scenario_specific"}, Fallback: "Проверим условия текущего заказа."})
	if err != nil || result.Message != "Проверим условия текущего заказа." || result.Tactic != "scenario_specific" {
		t.Fatalf("GenerateReply() = (%#v, %v), want scenario fallback", result, err)
	}
}

func TestGeneratorSkipsMissingCuratedTacticPool(t *testing.T) {
	generator := attemptsservice.NewModelAI(&unavailableStructuredModel{})
	result, err := generator.GenerateReply(context.Background(), attemptsservice.GenerationRequest{UserRole: "buyer", Phase: "hook", AllowedTactics: []string{"scenario_specific", "urgency"}})
	if err != nil || result.Message != "Нужно решить сегодня, пожалуйста, не откладывайте." || result.Tactic != "urgency" {
		t.Fatalf("GenerateReply() = (%#v, %v), want next available tactic variation", result, err)
	}
}
