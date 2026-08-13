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
	Cases                 int     `json:"cases"`
	StructuredJSONRate    float64 `json:"structured_json_rate"`
	RiskRecognitionRate   float64 `json:"risk_recognition_rate"`
	SafeFalsePositiveRate float64 `json:"safe_false_positive_rate"`
	P95LatencyMS          int64   `json:"p95_latency_ms"`
	Retries               int64   `json:"retries"`
	FallbackRate          float64 `json:"fallback_rate"`
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
	evaluator := attemptsservice.NewModelAI(attemptsai.New(provider))
	cases := ClosedCases()
	latencies := make([]time.Duration, 0, len(cases))
	risky, riskyRecognized, safe, safeFalsePositive := 0, 0, 0, 0
	for _, item := range cases {
		started := time.Now()
		result, evaluateErr := evaluator.Evaluate(context.Background(), attemptsservice.EvaluationRequest{RiskType: item.RiskType, Answer: item.Answer})
		latencies = append(latencies, time.Since(started))
		if evaluateErr != nil {
			t.Fatalf("case %s: %v", item.ID, evaluateErr)
		}
		normalizedSignals := attemptsservice.NormalizeRiskSignalCodes(result.DetectedSignals, domain.RiskType(item.RiskType))
		if !item.ExpectedSafe && !containsSignal(normalizedSignals, item.ExpectedSignal) {
			t.Errorf("case %s: normalized signals=%v, want %q", item.ID, normalizedSignals, item.ExpectedSignal)
		}
		if result.Score < item.MinScore || result.Score > item.MaxScore {
			t.Errorf("case %s: score=%d, want %d..%d", item.ID, result.Score, item.MinScore, item.MaxScore)
		}
		if item.ExpectedSafe {
			safe++
			if !result.IsSafe {
				safeFalsePositive++
			}
		} else {
			risky++
			if !result.IsSafe {
				riskyRecognized++
			}
		}
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	metrics := evaluator.Metrics().Evaluator
	report := evaluationReport{Cases: len(cases), StructuredJSONRate: 1 - float64(metrics.Fallbacks)/float64(len(cases)), RiskRecognitionRate: float64(riskyRecognized) / float64(risky), SafeFalsePositiveRate: float64(safeFalsePositive) / float64(safe), P95LatencyMS: latencies[(len(latencies)*95+99)/100-1].Milliseconds(), Retries: metrics.Retries, FallbackRate: float64(metrics.Fallbacks) / float64(len(cases))}
	encoded, _ := json.Marshal(report)
	t.Logf("evaluator report: %s", encoded)
	if t.Failed() || report.StructuredJSONRate < .90 || report.RiskRecognitionRate < .85 || report.SafeFalsePositiveRate > .20 || report.P95LatencyMS > 30000 || report.FallbackRate > .10 {
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
