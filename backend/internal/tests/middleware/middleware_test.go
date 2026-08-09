package middleware_test

import (
	"anti-scam-trainer/backend/internal/core/logger"
	"anti-scam-trainer/backend/internal/core/server/middleware"
	"anti-scam-trainer/backend/internal/core/server/response"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestAPIErrorEnvelopeUsesResponseRequestID(t *testing.T) {
	handler := middleware.RequestID()(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		response.Error(writer, "invalid payload", http.StatusBadRequest)
	}))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	request.Header.Set(middleware.RequestIDHeader, "request-123")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if got, want := strings.TrimSpace(recorder.Body.String()), `{"error":{"code":"VALIDATION_ERROR","message":"invalid payload","details":{},"request_id":"request-123"}}`; got != want {
		t.Fatalf("body = %s, want %s", got, want)
	}
	if got := recorder.Header().Get(middleware.RequestIDHeader); got != "request-123" {
		t.Fatalf("X-Request-ID = %q", got)
	}
}

func TestChainPreservesCallerRequestIDAndProvidesRequestLogger(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	request.Header.Set(middleware.RequestIDHeader, "request-42")
	recorder := httptest.NewRecorder()
	handler := middleware.Chain(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if got := request.Header.Get(middleware.RequestIDHeader); got != "request-42" {
			t.Fatalf("request ID = %q", got)
		}
		logger.FromContext(request.Context()).Debug("request logger is available")
		writer.WriteHeader(http.StatusNoContent)
	}), middleware.RequestID(), middleware.Logger(testLogger()), middleware.Panic(), middleware.Trace())

	handler.ServeHTTP(recorder, request)
	if got := recorder.Header().Get(middleware.RequestIDHeader); got != "request-42" {
		t.Fatalf("response request ID = %q", got)
	}
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestChainCreatesRequestIDAndRecoversPanic(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler := middleware.Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("unexpected")
	}), middleware.RequestID(), middleware.Logger(testLogger()), middleware.Panic(), middleware.Trace())

	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if recorder.Header().Get(middleware.RequestIDHeader) == "" {
		t.Fatal("response has no generated request ID")
	}
}

func testLogger() *logger.Logger {
	return &logger.Logger{Logger: zap.NewNop()}
}
