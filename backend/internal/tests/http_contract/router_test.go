package http_contract_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/logger"
	"anti-scam-trainer/backend/internal/core/server/middleware"
	"anti-scam-trainer/backend/internal/core/server/router"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	attemptshttp "anti-scam-trainer/backend/internal/features/attempts/transport/http"
	scenariosservice "anti-scam-trainer/backend/internal/features/scenarios/service"
	scenarioshttp "anti-scam-trainer/backend/internal/features/scenarios/transport/http"
	usersservice "anti-scam-trainer/backend/internal/features/users/service"
	usershttp "anti-scam-trainer/backend/internal/features/users/transport/http"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestRouterServesVersionedProductRoutesWithoutPostgres(t *testing.T) {
	versionedRouter := router.New()
	versionedRouter.Register(router.V1, usershttp.New(usersservice.New(fakeUsers{})).Routes())
	versionedRouter.Register(router.V1, scenarioshttp.New(scenariosservice.New(fakeScenarios{})).Routes())
	versionedRouter.Register(router.V1, attemptshttp.New(attemptsservice.New(fakeAttempts{}, fakeCompletionRepository{})).Routes())
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{name: "lists users", method: http.MethodGet, path: "/api/v1/users", wantStatus: 200, wantBody: `[{"id":1,"user_id":"external-42","username":"alex","completed_chats":0}]`},
		{name: "creates scenario", method: http.MethodPost, path: "/api/v1/scenarios", body: `{"title":"Поддельная доставка","description":"Тренировка","difficulty":"easy","role":"seller","is_active":true}`, wantStatus: 200, wantBody: `{"id":2,"title":"Поддельная доставка","description":"Тренировка","difficulty":"easy","role":"seller","is_active":true}`},
		{name: "returns an attempt", method: http.MethodGet, path: "/api/v1/attempts/1", wantStatus: 200, wantBody: `{"id":1,"user_id":1,"scenario_id":1,"status":"IN_PROGRESS","started_at":"2026-08-07T18:00:00Z","finished_at":"0001-01-01T00:00:00Z","score":0}`},
		{name: "rejects invalid attempt transition", method: http.MethodPut, path: "/api/v1/attempts/2", body: `{"user_id":1,"scenario_id":1,"status":"IN_PROGRESS","started_at":"2026-08-07T18:00:00Z","finished_at":"0001-01-01T00:00:00Z","score":0}`, wantStatus: 500, wantBody: "invalid attempt status transition"},
		{name: "rejects invalid identifier", method: http.MethodGet, path: "/api/v1/users/nope", wantStatus: 400, wantBody: "invalid user id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()
			versionedRouter.ServeHTTP(recorder, request)
			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if got := strings.TrimSpace(recorder.Body.String()); got != tt.wantBody {
				t.Fatalf("body = %q, want %q", got, tt.wantBody)
			}
		})
	}
}

func TestVersionedRouterPreservesRequestID(t *testing.T) {
	versionedRouter := router.New()
	versionedRouter.Register(router.V1, []router.Route{{Path: "/health", Handler: func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}}})
	handler := middleware.Chain(versionedRouter, middleware.RequestID(), middleware.Logger(&logger.Logger{Logger: zap.NewNop()}), middleware.Panic(), middleware.Trace())

	for _, requestID := range []string{"", "request-42"} {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		if requestID != "" {
			request.Header.Set(middleware.RequestIDHeader, requestID)
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
		}
		got := recorder.Header().Get(middleware.RequestIDHeader)
		if got == "" {
			t.Fatal("response has no request ID")
		}
		if requestID != "" && got != requestID {
			t.Fatalf("response request ID = %q, want %q", got, requestID)
		}
	}
}

type fakeUsers struct{}

func (fakeUsers) Create(user domain.User) (domain.User, error) { user.ID = 1; return user, nil }
func (fakeUsers) GetByID(id int) (domain.User, error) {
	return domain.User{ID: id, ExternalID: "external-42", Username: "alex"}, nil
}
func (fakeUsers) GetByExternalID(string) (domain.User, error) {
	return domain.User{}, errors.New("not found")
}
func (fakeUsers) Update(domain.User) error { return nil }
func (fakeUsers) Delete(int) error         { return nil }
func (fakeUsers) List() ([]domain.User, error) {
	return []domain.User{{ID: 1, ExternalID: "external-42", Username: "alex"}}, nil
}

type fakeScenarios struct{}

func (fakeScenarios) Create(scenario domain.Scenario) (domain.Scenario, error) {
	scenario.ID = 2
	return scenario, nil
}
func (fakeScenarios) GetByID(id int) (domain.Scenario, error) { return domain.Scenario{ID: id}, nil }
func (fakeScenarios) Update(domain.Scenario) error            { return nil }
func (fakeScenarios) Delete(int) error                        { return nil }
func (fakeScenarios) List() ([]domain.Scenario, error)        { return nil, nil }

type fakeAttempts struct{}

func (fakeAttempts) Create(attempt domain.Attempt) (domain.Attempt, error) {
	attempt.ID = 1
	return attempt, nil
}
func (fakeAttempts) GetByID(id int) (domain.Attempt, error) {
	status := domain.AttemptStatusInProgress
	if id == 2 {
		status = domain.AttemptStatusCompleted
	}
	return domain.Attempt{ID: id, UserID: 1, ScenarioID: 1, Status: status, StartedAt: mustTime("2026-08-07T18:00:00Z")}, nil
}
func (fakeAttempts) Update(domain.Attempt) error     { return nil }
func (fakeAttempts) Delete(int) error                { return nil }
func (fakeAttempts) List() ([]domain.Attempt, error) { return nil, nil }

type fakeCompletionRepository struct{}

func (fakeCompletionRepository) InTransaction(func(attemptsservice.CompletionStore) error) error {
	return nil
}
func mustTime(value string) (parsedTime time.Time) {
	parsedTime, _ = time.Parse(time.RFC3339, value)
	return parsedTime
}
