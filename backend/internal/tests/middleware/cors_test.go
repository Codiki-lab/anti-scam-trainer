package middleware_test

import (
	"anti-scam-trainer/backend/internal/core/server/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCredentialedCORSAllowsOnlyConfiguredFrontendOrigins(t *testing.T) {
	handler := middleware.RequestID()(middleware.CORS([]string{"http://localhost:5173"})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })))

	allowedRequest := httptest.NewRequest(http.MethodOptions, "/api/v1/training/levels", nil)
	allowedRequest.Header.Set("Origin", "http://localhost:5173")
	allowedRequest.Header.Set("Access-Control-Request-Method", http.MethodGet)
	allowed := httptest.NewRecorder()
	handler.ServeHTTP(allowed, allowedRequest)
	if allowed.Code != http.StatusNoContent || allowed.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" || allowed.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatalf("allowed preflight = (%d, %#v)", allowed.Code, allowed.Header())
	}
	if allowed.Header().Get("X-Request-ID") == "" {
		t.Fatal("allowed preflight is missing X-Request-ID")
	}

	blockedRequest := httptest.NewRequest(http.MethodGet, "/api/v1/training/levels", nil)
	blockedRequest.Header.Set("Origin", "https://malicious.example")
	blocked := httptest.NewRecorder()
	handler.ServeHTTP(blocked, blockedRequest)
	if blocked.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("unknown origin received CORS access: %#v", blocked.Header())
	}
}
