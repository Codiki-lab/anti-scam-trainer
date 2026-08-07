package aiprovider

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestOllamaGeneratesVirtualInterlocutorMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Fatalf("path = %q, want /api/chat", r.URL.Path)
		}
		var request struct {
			Model    string    `json:"model"`
			Messages []Message `json:"messages"`
			Stream   bool      `json:"stream"`
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
		if got := request.Messages; len(got) != 2 || got[0].Role != RoleSystem || got[1].Content != "Здравствуйте" {
			t.Fatalf("messages = %#v", got)
		}
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"Чем могу помочь?"},"done":true,"prompt_eval_count":18,"eval_count":7}`))
	}))
	defer server.Close()

	provider := mustProvider(t, Config{URL: server.URL, Model: "llama3.2:3b", ContextWindowTokens: 1024, OutputReserveTokens: 256})
	result, err := provider.Generate(context.Background(), []Message{{Role: RoleSystem, Content: "Ты виртуальный собеседник"}, {Role: RoleUser, Content: "Здравствуйте"}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Content != "Чем могу помочь?" {
		t.Fatalf("content = %q", result.Content)
	}
	if result.Usage.PromptTokens != 18 || result.Usage.CompletionTokens != 7 {
		t.Fatalf("actual usage = %#v", result.Usage)
	}
	if result.Usage.ContextRisk != ContextRiskLow || result.Usage.ReservedOutputTokens != 256 {
		t.Fatalf("context usage = %#v", result.Usage)
	}
}

func TestOllamaDistinguishesFailures(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		matches func(error) bool
	}{
		{name: "non-success response", handler: func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "model missing", http.StatusNotFound) }, matches: func(err error) bool { var target *OllamaError; return errors.As(err, &target) }},
		{name: "malformed response", handler: func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"message":`)) }, matches: func(err error) bool { var target *MalformedResponseError; return errors.As(err, &target) }},
		{name: "missing required response fields", handler: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"ok"},"done":true}`))
		}, matches: func(err error) bool { var target *MalformedResponseError; return errors.As(err, &target) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()
			_, err := mustProvider(t, Config{URL: server.URL, Model: "llama3.2:3b", ContextWindowTokens: 1024}).Generate(context.Background(), []Message{{Role: RoleUser, Content: "hi"}})
			if !tt.matches(err) {
				t.Fatalf("error = %T (%v), has unexpected type", err, err)
			}
		})
	}
}

func TestOllamaReturnsCallerCancellation(t *testing.T) {
	provider := mustProvider(t, Config{URL: "http://127.0.0.1:1", Model: "llama3.2:3b", ContextWindowTokens: 1024})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := provider.Generate(ctx, []Message{{Role: RoleUser, Content: "hi"}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %T (%v), want context cancellation", err, err)
	}
}

func TestOllamaTimeoutIsTransportErrorAndDoesNotRetry(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		calls++
		time.Sleep(50 * time.Millisecond)
	}))
	defer server.Close()
	provider := mustProvider(t, Config{URL: server.URL, Model: "llama3.2:3b", RequestTimeout: time.Millisecond, ContextWindowTokens: 1024})
	_, err := provider.Generate(context.Background(), []Message{{Role: RoleUser, Content: "hi"}})
	var transportError *TransportError
	if !errors.As(err, &transportError) {
		t.Fatalf("error = %T (%v), want TransportError", err, err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want exactly one request", calls)
	}
}

func TestOllamaRejectsHistoryThatDoesNotLeaveOutputReserve(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	defer server.Close()
	provider := mustProvider(t, Config{URL: server.URL, Model: "llama3.2:3b", ContextWindowTokens: 32, OutputReserveTokens: 16})
	_, err := provider.Generate(context.Background(), []Message{{Role: RoleUser, Content: "this history is too long for the configured context"}})
	var capacityError *ContextCapacityError
	if !errors.As(err, &capacityError) {
		t.Fatalf("error = %T (%v), want ContextCapacityError", err, err)
	}
	if called {
		t.Fatal("provider sent an oversized history")
	}
}

func TestContextRiskBoundariesAndDefaultReserve(t *testing.T) {
	config := Config{URL: "http://example.invalid", Model: "llama3.2:3b", ContextWindowTokens: 2000}
	provider := mustProvider(t, config).(*Ollama)
	if provider.outputReserveTokens != 400 {
		t.Fatalf("reserve = %d, want 400", provider.outputReserveTokens)
	}
	for _, tt := range []struct {
		name string
		used int
		want ContextRisk
	}{
		{name: "below 60 percent", used: 1199, want: ContextRiskLow},
		{name: "at 60 percent", used: 1200, want: ContextRiskMedium},
		{name: "at 75 percent", used: 1500, want: ContextRiskHigh},
		{name: "at 90 percent", used: 1800, want: ContextRiskHigh},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := provider.contextRisk(float64(tt.used) / 2000); got != tt.want {
				t.Fatalf("risk = %s, want %s", got, tt.want)
			}
		})
	}

	minimum := mustProvider(t, Config{URL: "http://example.invalid", Model: "llama3.2:3b", ContextWindowTokens: 1000}).(*Ollama)
	if minimum.outputReserveTokens != 256 {
		t.Fatalf("reserve floor = %d, want 256", minimum.outputReserveTokens)
	}

	custom := mustProvider(t, Config{URL: "http://example.invalid", Model: "llama3.2:3b", ContextWindowTokens: 2000, MediumRiskThreshold: .30, HighRiskThreshold: .50}).(*Ollama)
	if got := custom.contextRisk(.30); got != ContextRiskMedium {
		t.Fatalf("custom medium boundary risk = %s", got)
	}
	if got := custom.contextRisk(.50); got != ContextRiskHigh {
		t.Fatalf("custom high boundary risk = %s", got)
	}
}

func TestLlama32PreflightEstimateMatchesLocalOllamaCalibration(t *testing.T) {
	if os.Getenv("OLLAMA_PREFLIGHT_TEST") != "1" {
		t.Skip("set OLLAMA_PREFLIGHT_TEST=1 after pulling llama3.2:3b to calibrate against a real Ollama runtime")
	}
	url := os.Getenv("OLLAMA_URL")
	if url == "" {
		url = "http://localhost:11434"
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama3.2:3b"
	}
	provider := mustProvider(t, Config{URL: url, Model: model, RequestTimeout: 2 * time.Minute, ContextWindowTokens: 8192})

	tests := []struct {
		name     string
		messages []Message
	}{
		{name: "short english", messages: []Message{{Role: RoleUser, Content: "Hello"}}},
		{name: "russian dialogue", messages: []Message{{Role: RoleSystem, Content: "Ты помощник по безопасным сделкам."}, {Role: RoleUser, Content: "Продавец просит перейти по ссылке для доставки."}}},
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

func mustProvider(t *testing.T, config Config) Provider {
	t.Helper()
	provider, err := NewOllama(config)
	if err != nil {
		t.Fatal(err)
	}
	return provider
}
