package attempts_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	attemptsai "anti-scam-trainer/backend/internal/features/attempts/aiprovider"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	"context"
	"strings"
	"testing"
)

func TestEvaluatorRepairsOnceAndKeepsItsOwnProfile(t *testing.T) {
	provider := &sequenceProvider{contents: []string{`{"score":9}`, `{"score":4,"is_safe":true,"risk_type":"phishing","detected_signals":[],"evaluation":"Безопасный отказ","safe_action":"Проверить заказ в приложении"}`}}
	modelAI := attemptsservice.NewModelAI(attemptsai.New(provider))
	result, err := modelAI.Evaluate(context.Background(), attemptsservice.EvaluationRequest{Policy: "policy", RiskType: "phishing", ScenarioInstruction: "Сохраняй факты Сценария", Rubric: domain.JSONObject{"safe_action": "Остаться в сервисе"}, EvaluationContext: "context", Answer: "Не перейду"})
	if err != nil || result.Score != 4 || len(provider.requests) != 2 {
		t.Fatalf("Evaluate() = (%#v, %v), requests=%d", result, err, len(provider.requests))
	}
	for _, request := range provider.requests {
		if request.OutputTokens != 240 || request.Temperature != 0 || request.Schema == nil {
			t.Fatalf("evaluator profile = %#v", request)
		}
		prompt := request.Messages[1].Content
		if !strings.Contains(prompt, "Server policy (authoritative): policy") || !strings.Contains(prompt, "Managed scenario instruction (context only): Сохраняй факты Сценария") || !strings.Contains(prompt, `Managed final rubric (context only): {"safe_action":"Остаться в сервисе"}`) {
			t.Fatalf("evaluator prompt does not compose managed context safely: %s", prompt)
		}
	}
}

func TestEvaluatorFallsBackWhenOllamaReturnsInvalidJSONTwice(t *testing.T) {
	provider := &sequenceProvider{contents: []string{`{"score":9}`, `{"score":9}`}}
	modelAI := attemptsservice.NewModelAI(attemptsai.New(provider))

	result, err := modelAI.Evaluate(context.Background(), attemptsservice.EvaluationRequest{RiskType: "phishing", Answer: "Перейду по ссылке"})
	if err != nil || result.Score != 1 || result.IsSafe || result.RiskType != "phishing" || result.Evaluation == "" || result.SafeAction == "" || len(provider.requests) != 2 {
		t.Fatalf("Evaluate() = (%#v, %v), requests=%d; want safe fallback", result, err, len(provider.requests))
	}
}

func TestEvaluatorFallbackRecognizesExplicitSafeAnswer(t *testing.T) {
	provider := &sequenceProvider{contents: []string{`not-json`, `still-not-json`}}
	modelAI := attemptsservice.NewModelAI(attemptsai.New(provider))

	result, err := modelAI.Evaluate(context.Background(), attemptsservice.EvaluationRequest{RiskType: "phishing", Answer: "Проверю заказ только внутри приложения"})
	if err != nil || result.Score != 3 || !result.IsSafe || !strings.Contains(result.Evaluation, "внутри сервиса") {
		t.Fatalf("Evaluate() = (%#v, %v); want Russian safe fallback", result, err)
	}
}
