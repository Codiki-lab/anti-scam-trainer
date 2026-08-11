package aiprovider_test

import (
	aiprovider "anti-scam-trainer/backend/internal/core/aiprovider"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestOllamaGeneratesVirtualInterlocutorMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Fatalf("path = %q, want /api/chat", r.URL.Path)
		}
		var request struct {
			Model    string               `json:"model"`
			Messages []aiprovider.Message `json:"messages"`
			Stream   bool                 `json:"stream"`
			Options  struct {
				ContextWindowTokens int `json:"num_ctx"`
			} `json:"options"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request.Model != "llama3.2:3b" || request.Stream || request.Options.ContextWindowTokens != 1024 {
			t.Fatalf("unexpected request: %#v", request)
		}
		if got := request.Messages; len(got) != 2 || got[0].Role != aiprovider.RoleSystem || got[1].Content != "Здравствуйте" {
			t.Fatalf("messages = %#v", got)
		}
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"Чем могу помочь?"},"done":true,"prompt_eval_count":18,"eval_count":7}`))
	}))
	defer server.Close()

	provider := mustProvider(t, aiprovider.Config{URL: server.URL, Model: "llama3.2:3b", ContextWindowTokens: 1024, OutputReserveTokens: 256})
	result, err := provider.Generate(context.Background(), []aiprovider.Message{{Role: aiprovider.RoleSystem, Content: "Ты виртуальный собеседник"}, {Role: aiprovider.RoleUser, Content: "Здравствуйте"}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Content != "Чем могу помочь?" {
		t.Fatalf("content = %q", result.Content)
	}
	if result.Usage.PromptTokens != 18 || result.Usage.CompletionTokens != 7 {
		t.Fatalf("actual usage = %#v", result.Usage)
	}
	if result.Usage.ContextRisk != aiprovider.ContextRiskLow || result.Usage.ReservedOutputTokens != 256 {
		t.Fatalf("context usage = %#v", result.Usage)
	}
}

func TestOllamaUsesSeparateStructuredProfiles(t *testing.T) {
	requests := make([]map[string]any, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		requests = append(requests, request)
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"{}"},"done":true,"prompt_eval_count":10,"eval_count":2}`))
	}))
	defer server.Close()
	provider := mustProvider(t, aiprovider.Config{URL: server.URL, Model: "qwen3:8b", ContextWindowTokens: 8192})
	schema := map[string]any{"type": "object", "additionalProperties": false}
	_, err := provider.GenerateStructured(context.Background(), aiprovider.StructuredRequest{Messages: []aiprovider.Message{{Role: aiprovider.RoleUser, Content: "evaluate"}}, Schema: schema, OutputTokens: 240, Temperature: 0})
	if err != nil {
		t.Fatal(err)
	}
	_, err = provider.GenerateStructured(context.Background(), aiprovider.StructuredRequest{Messages: []aiprovider.Message{{Role: aiprovider.RoleUser, Content: "generate"}}, Schema: schema, OutputTokens: 120, Temperature: .3, TopP: .8, TopK: 20, RepeatPenalty: 1.1})
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) != 2 || requests[0]["think"] != false || requests[1]["think"] != false {
		t.Fatalf("requests = %#v", requests)
	}
	firstOptions := requests[0]["options"].(map[string]any)
	secondOptions := requests[1]["options"].(map[string]any)
	if firstOptions["num_predict"] != float64(240) || secondOptions["num_predict"] != float64(120) || secondOptions["temperature"] != .3 || requests[0]["format"] == nil {
		t.Fatalf("structured profiles = %#v", requests)
	}
}

func TestOllamaDistinguishesFailures(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		matches func(error) bool
	}{
		{name: "non-success response", handler: func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "model missing", http.StatusNotFound) }, matches: func(err error) bool { var target *aiprovider.OllamaError; return errors.As(err, &target) }},
		{name: "malformed response", handler: func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"message":`)) }, matches: func(err error) bool { var target *aiprovider.MalformedResponseError; return errors.As(err, &target) }},
		{name: "missing required response fields", handler: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"ok"},"done":true}`))
		}, matches: func(err error) bool { var target *aiprovider.MalformedResponseError; return errors.As(err, &target) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()
			_, err := mustProvider(t, aiprovider.Config{URL: server.URL, Model: "llama3.2:3b", ContextWindowTokens: 1024}).Generate(context.Background(), []aiprovider.Message{{Role: aiprovider.RoleUser, Content: "hi"}})
			if !tt.matches(err) {
				t.Fatalf("error = %T (%v), has unexpected type", err, err)
			}
		})
	}
}

func TestOllamaReturnsCallerCancellation(t *testing.T) {
	provider := mustProvider(t, aiprovider.Config{URL: "http://127.0.0.1:1", Model: "llama3.2:3b", ContextWindowTokens: 1024})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := provider.Generate(ctx, []aiprovider.Message{{Role: aiprovider.RoleUser, Content: "hi"}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %T (%v), want context cancellation", err, err)
	}
}

func TestOllamaTimeoutIsTransportErrorAndDoesNotRetry(t *testing.T) {
	calls := 0
	var callsMu sync.Mutex
	requestStarted := make(chan struct{})
	var startedOnce sync.Once
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		callsMu.Lock()
		calls++
		callsMu.Unlock()
		startedOnce.Do(func() { close(requestStarted) })
		time.Sleep(50 * time.Millisecond)
	}))
	defer server.Close()
	provider := mustProvider(t, aiprovider.Config{URL: server.URL, Model: "llama3.2:3b", RequestTimeout: time.Millisecond, ContextWindowTokens: 1024})
	_, err := provider.Generate(context.Background(), []aiprovider.Message{{Role: aiprovider.RoleUser, Content: "hi"}})
	var transportError *aiprovider.TransportError
	if !errors.As(err, &transportError) {
		t.Fatalf("error = %T (%v), want TransportError", err, err)
	}
	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for the request to reach Ollama")
	}
	callsMu.Lock()
	defer callsMu.Unlock()
	if calls != 1 {
		t.Fatalf("calls = %d, want exactly one request", calls)
	}
}

func TestOllamaRejectsHistoryThatDoesNotLeaveOutputReserve(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	defer server.Close()
	provider := mustProvider(t, aiprovider.Config{URL: server.URL, Model: "llama3.2:3b", ContextWindowTokens: 32, OutputReserveTokens: 16})
	_, err := provider.Generate(context.Background(), []aiprovider.Message{{Role: aiprovider.RoleUser, Content: "this history is too long for the configured context"}})
	var capacityError *aiprovider.ContextCapacityError
	if !errors.As(err, &capacityError) {
		t.Fatalf("error = %T (%v), want ContextCapacityError", err, err)
	}
	if called {
		t.Fatal("provider sent an oversized history")
	}
}

func TestContextRiskBoundariesAndDefaultReserve(t *testing.T) {
	defaultConfig := aiprovider.Config{Model: "llama3.2:3b", ContextWindowTokens: 2000}
	for _, tt := range []struct {
		name    string
		content string
		want    aiprovider.ContextRisk
	}{
		{name: "below 60 percent", content: strings.Repeat("a", 789), want: aiprovider.ContextRiskLow},
		{name: "at 60 percent", content: strings.Repeat("a", 790), want: aiprovider.ContextRiskMedium},
		{name: "at 75 percent", content: strings.Repeat("a", 1090), want: aiprovider.ContextRiskHigh},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := generateResult(t, defaultConfig, tt.content)
			if result.Usage.ReservedOutputTokens != 400 || result.Usage.ContextRisk != tt.want {
				t.Fatalf("usage = %#v", result.Usage)
			}
		})
	}

	minimum := generateResult(t, aiprovider.Config{Model: "llama3.2:3b", ContextWindowTokens: 1000}, "hi")
	if minimum.Usage.ReservedOutputTokens != 256 {
		t.Fatalf("reserve floor = %d, want 256", minimum.Usage.ReservedOutputTokens)
	}

	customConfig := aiprovider.Config{Model: "llama3.2:3b", ContextWindowTokens: 1000, MediumRiskThreshold: .30, HighRiskThreshold: .50}
	if got := generateResult(t, customConfig, strings.Repeat("a", 34)).Usage.ContextRisk; got != aiprovider.ContextRiskMedium {
		t.Fatalf("custom medium boundary risk = %s", got)
	}
	if got := generateResult(t, customConfig, strings.Repeat("a", 234)).Usage.ContextRisk; got != aiprovider.ContextRiskHigh {
		t.Fatalf("custom high boundary risk = %s", got)
	}
}

func TestQwen3PreflightEstimateMatchesLocalOllamaCalibration(t *testing.T) {
	if os.Getenv("OLLAMA_PREFLIGHT_TEST") != "1" {
		t.Skip("set OLLAMA_PREFLIGHT_TEST=1 after pulling qwen3:8b to calibrate against a real Ollama runtime")
	}
	url := os.Getenv("OLLAMA_URL")
	if url == "" {
		url = "http://localhost:11434"
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "qwen3:8b"
	}
	provider := mustProvider(t, aiprovider.Config{URL: url, Model: model, RequestTimeout: 2 * time.Minute, ContextWindowTokens: 8192})

	tests := []struct {
		name     string
		messages []aiprovider.Message
	}{
		{name: "short english", messages: []aiprovider.Message{{Role: aiprovider.RoleUser, Content: "Hello"}}},
		{name: "russian dialogue", messages: []aiprovider.Message{{Role: aiprovider.RoleSystem, Content: "Ты помощник по безопасным сделкам."}, {Role: aiprovider.RoleUser, Content: "Продавец просит перейти по ссылке для доставки."}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := provider.Generate(context.Background(), tt.messages)
			if err != nil {
				t.Fatal(err)
			}
			if result.Usage.EstimatedPromptTokens < result.Usage.PromptTokens {
				t.Fatalf("estimate = %d, actual prompt_eval_count = %d", result.Usage.EstimatedPromptTokens, result.Usage.PromptTokens)
			}
			if result.Usage.EstimatedPromptTokens > max(32, result.Usage.PromptTokens*2) {
				t.Fatalf("estimate = %d exceeds calibration tolerance for actual prompt_eval_count = %d", result.Usage.EstimatedPromptTokens, result.Usage.PromptTokens)
			}
		})
	}
}

func generateResult(t *testing.T, config aiprovider.Config, content string) aiprovider.Result {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"ok"},"done":true,"prompt_eval_count":1,"eval_count":1}`))
	}))
	defer server.Close()
	config.URL = server.URL
	result, err := mustProvider(t, config).Generate(context.Background(), []aiprovider.Message{{Role: aiprovider.RoleUser, Content: content}})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func mustProvider(t *testing.T, config aiprovider.Config) aiprovider.Provider {
	t.Helper()
	provider, err := aiprovider.NewOllama(config)
	if err != nil {
		t.Fatal(err)
	}
	return provider
}
