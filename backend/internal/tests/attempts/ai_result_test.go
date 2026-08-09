package attempts_test

import (
	"anti-scam-trainer/backend/internal/features/attempts/service"
	"testing"
)

func TestDecodeAIResultAcceptsStrictTrainingEvaluation(t *testing.T) {
	result, err := service.DecodeAIResult(`{"awarded_points":75,"explanation":"Ответ снижает риск","reply":"Оплатите по ссылке","risk_signals":["внешняя ссылка"]}`)
	if err != nil {
		t.Fatal(err)
	}
	if result.AwardedPoints != 75 || result.Explanation == "" || result.Reply == "" || len(result.RiskSignals) != 1 {
		t.Fatalf("DecodeAIResult() = %#v, want complete validated result", result)
	}
}

func TestDecodeAIResultRejectsUnknownFieldsAndInvalidPoints(t *testing.T) {
	for _, raw := range []string{
		`{"awarded_points":10,"explanation":"x","reply":"y","risk_signals":[]}`,
		`{"awarded_points":75,"explanation":"x","reply":"y","risk_signals":[],"finish":true}`,
		`{"awarded_points":75,"explanation":"","reply":"y","risk_signals":[]}`,
		`{"awarded_points":75,"explanation":"x","reply":"y","risk_signals":[]} trailing`,
	} {
		if _, err := service.DecodeAIResult(raw); err == nil {
			t.Fatalf("DecodeAIResult(%q) succeeded, want strict validation error", raw)
		}
	}
}

func TestFreeTextCompletionRules(t *testing.T) {
	if service.CanFinishFreeText(2, true) {
		t.Fatal("finish=true before the third free-text answer must be rejected")
	}
	if !service.CanFinishFreeText(3, true) {
		t.Fatal("finish=true on the third free-text answer must complete")
	}
	if !service.CanFinishFreeText(5, false) {
		t.Fatal("the fifth free-text answer must complete automatically")
	}
	if service.CanFinishFreeText(4, false) {
		t.Fatal("the fourth free-text answer without finish must continue")
	}
}
