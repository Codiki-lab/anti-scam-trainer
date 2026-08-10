package attempts_test

import (
	"anti-scam-trainer/backend/internal/features/attempts/service"
	"testing"
)

func TestDecodeEvaluatorResultAcceptsStrictTrainingEvaluation(t *testing.T) {
	result, err := service.DecodeEvaluatorResult(`{"score":3,"is_safe":true,"risk_type":"phishing","detected_signals":["внешняя ссылка"],"evaluation":"Ответ снижает риск","safe_action":"Проверить заказ в приложении"}`)
	if err != nil {
		t.Fatal(err)
	}
	if result.Score != 3 || len(result.DetectedSignals) != 1 || service.PointsForEvaluatorScore(result.Score) != 75 {
		t.Fatalf("result = %#v", result)
	}
}

func TestDecodeEvaluatorResultRejectsUnsafeOrExpandedContract(t *testing.T) {
	for _, raw := range []string{
		`{"score":4,"risk_type":"phishing","detected_signals":[],"evaluation":"x","safe_action":"y"}`,
		`{"score":5,"is_safe":true,"risk_type":"phishing","detected_signals":[],"evaluation":"x","safe_action":"y"}`,
		`{"score":4,"is_safe":true,"risk_type":"phishing","detected_signals":[],"evaluation":"x","safe_action":"y","next_step":2}`,
		`{"score":4,"is_safe":true,"risk_type":"phishing","detected_signals":[],"evaluation":"x","safe_action":"Откройте https://example.com"}`,
		`{"score":4,"is_safe":true,"risk_type":"phishing","detected_signals":[],"evaluation":"Safe JSON: безопасно","safe_action":"Проверить заказ в приложении"}`,
		`{"score":4,"is_safe":true,"risk_type":"phishing","detected_signals":[],"evaluation":"score=4, хорошо","safe_action":"Проверить заказ в приложении"}`,
	} {
		if _, err := service.DecodeEvaluatorResult(raw); err == nil {
			t.Fatalf("DecodeEvaluatorResult(%q) succeeded", raw)
		}
	}
}

func TestDecodeGeneratorResultRejectsUnsafeOrWrongPhaseOutput(t *testing.T) {
	for _, raw := range []string{
		`{"message":"Откройте https://example.com","tactic":"urgency","phase":"escalation"}`,
		`{"message":"Поторопитесь","tactic":"urgency","phase":"hook"}`,
	} {
		if _, err := service.DecodeGeneratorResult(raw, "escalation", []string{"urgency"}); err == nil {
			t.Fatalf("DecodeGeneratorResult(%q) succeeded", raw)
		}
	}
}
