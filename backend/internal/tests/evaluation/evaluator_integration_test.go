package evaluation

import (
	"anti-scam-trainer/backend/internal/core/aiprovider"
	"anti-scam-trainer/backend/internal/core/domain"
	attemptsai "anti-scam-trainer/backend/internal/features/attempts/aiprovider"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	"context"
	"encoding/json"
	"os"
	"sort"
	"testing"
	"time"
)

type evaluationReport struct {
	Cases               int     `json:"cases"`
	ModelCases          int     `json:"model_cases"`
	StructuredJSONRate  float64 `json:"structured_json_rate"`
	RiskRecognitionRate float64 `json:"risk_recognition_rate"`
	SafeMissRate        float64 `json:"safe_miss_rate"`
	P95LatencyMS        int64   `json:"p95_latency_ms"`
	Retries             int64   `json:"retries"`
	FallbackRate        float64 `json:"fallback_rate"`
}

type measuredStructuredModel struct {
	delegate        attemptsservice.StructuredModel
	calls           int
	validStructured int
}

func (m *measuredStructuredModel) GenerateStructured(ctx context.Context, request attemptsservice.StructuredModelRequest) (string, error) {
	m.calls++
	raw, err := m.delegate.GenerateStructured(ctx, request)
	if err == nil {
		if _, decodeErr := attemptsservice.DecodeEvaluatorResult(raw); decodeErr == nil {
			m.validStructured++
		}
	}
	return raw, err
}

func TestConfiguredEvaluatorMeetsReleaseThresholds(t *testing.T) {
	if os.Getenv("AI_EVALUATION_TEST") != "1" {
		t.Skip("set AI_EVALUATION_TEST=1 to run the closed set against configured Ollama")
	}
	url := os.Getenv("OLLAMA_URL")
	if url == "" {
		url = "http://localhost:11434"
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "qwen3:8b"
	}
	provider, err := aiprovider.NewOllama(aiprovider.Config{URL: url, Model: model, RequestTimeout: 30 * time.Second, ContextWindowTokens: 8192})
	if err != nil {
		t.Fatal(err)
	}
	modelProbe := &measuredStructuredModel{delegate: attemptsai.New(provider)}
	evaluator := attemptsservice.NewModelAI(modelProbe)
	cases := ClosedCases()
	modelLatencies := make([]time.Duration, 0, len(cases))
	modelCases, modelRisky, modelRiskyRecognized, modelSafe, modelSafeMisses := 0, 0, 0, 0, 0
	for _, item := range cases {
		callsBefore := modelProbe.calls
		started := time.Now()
		result, evaluateErr := evaluator.Evaluate(context.Background(), attemptsservice.EvaluationRequest{
			Policy: attemptsservice.PolicyFor(item.Role, item.RiskType), RiskType: item.RiskType,
			ScenarioInstruction: "Оцени текущий Ответ пользователя в рамках управляемого Сценария и не меняй его факты.",
			Rubric:              domain.JSONObject{"safe_action": item.EvaluationContext}, EvaluationContext: item.EvaluationContext,
			History: []domain.DialogueMessage{{Role: domain.MessageRoleAssistant, Text: item.CounterpartyMessage}}, Answer: item.Answer,
		})
		elapsed := time.Since(started)
		usedModel := modelProbe.calls > callsBefore
		if usedModel {
			modelCases++
			modelLatencies = append(modelLatencies, elapsed)
		}
		if evaluateErr != nil {
			t.Fatalf("case %s: %v", item.ID, evaluateErr)
		}
		normalizedSignals := attemptsservice.NormalizeRiskSignalCodes(result.DetectedSignals, domain.RiskType(item.RiskType))
		if !item.ExpectedSafe && item.ExpectedSignal != "" && !containsSignal(normalizedSignals, item.ExpectedSignal) {
			t.Errorf("case %s: normalized signals=%v, want %q", item.ID, normalizedSignals, item.ExpectedSignal)
		}
		if result.Score < item.MinScore || result.Score > item.MaxScore {
			t.Errorf("case %s: score=%d, want %d..%d", item.ID, result.Score, item.MinScore, item.MaxScore)
		}
		if item.ExpectedSafe {
			if usedModel {
				modelSafe++
				if !result.IsSafe {
					modelSafeMisses++
				}
			}
		} else if usedModel {
			modelRisky++
			if !result.IsSafe {
				modelRiskyRecognized++
			}
		}
	}
	if modelCases == 0 || modelSafe == 0 || modelRisky == 0 || modelProbe.calls == 0 {
		t.Fatalf("closed set model coverage: cases=%d safe=%d risky=%d calls=%d", modelCases, modelSafe, modelRisky, modelProbe.calls)
	}
	sort.Slice(modelLatencies, func(i, j int) bool { return modelLatencies[i] < modelLatencies[j] })
	metrics := evaluator.Metrics().Evaluator
	report := evaluationReport{Cases: len(cases), ModelCases: modelCases, StructuredJSONRate: float64(modelProbe.validStructured) / float64(modelProbe.calls), RiskRecognitionRate: float64(modelRiskyRecognized) / float64(modelRisky), SafeMissRate: float64(modelSafeMisses) / float64(modelSafe), P95LatencyMS: modelLatencies[(len(modelLatencies)*95+99)/100-1].Milliseconds(), Retries: metrics.Retries, FallbackRate: float64(metrics.Fallbacks) / float64(modelCases)}
	encoded, _ := json.Marshal(report)
	t.Logf("evaluator report: %s", encoded)
	if t.Failed() || report.StructuredJSONRate < .90 || report.RiskRecognitionRate < .85 || report.SafeMissRate > .20 || report.P95LatencyMS > 30000 || report.FallbackRate > .10 {
		t.Fatalf("release thresholds failed: %s", encoded)
	}
}

func containsSignal(signals []string, expected string) bool {
	for _, signal := range signals {
		if signal == expected {
			return true
		}
	}
	return false
}
