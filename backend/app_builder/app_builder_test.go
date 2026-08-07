package app_builder

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/httpserver"
	chatshttp "anti-scam-trainer/backend/modules/chats/http"
	chatsservice "anti-scam-trainer/backend/modules/chats/service"
	sessionshttp "anti-scam-trainer/backend/modules/sessions/http"
	sessionsservice "anti-scam-trainer/backend/modules/sessions/service"
	usershttp "anti-scam-trainer/backend/modules/users/http"
	usersservice "anti-scam-trainer/backend/modules/users/service"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRouterPreservesCRUDHTTPContractWithoutPostgres(t *testing.T) {
	router := httpserver.NewRouter(
		usershttp.New(usersservice.New(fakeUsers{})),
		chatshttp.New(chatsservice.New(fakeChats{})),
		sessionshttp.New(sessionsservice.New(fakeSessions{})),
	)
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{name: "lists users", method: http.MethodGet, path: "/users", wantStatus: 200, wantBody: `[{"id":1,"user_id":"external-42","username":"alex","completed_chats":0}]`},
		{name: "creates chat", method: http.MethodPost, path: "/chats", body: `{"title":"Поддельная доставка","description":"Тренировка","difficulty":"easy","role":"seller","is_active":true}`, wantStatus: 200, wantBody: `{"id":2,"title":"Поддельная доставка","description":"Тренировка","difficulty":"easy","role":"seller","is_active":true}`},
		{name: "returns a session", method: http.MethodGet, path: "/chat-sessions/1", wantStatus: 200, wantBody: `{"id":1,"user_id":1,"chat_id":1,"status":"IN_PROGRESS","started_at":"2026-08-07T18:00:00Z","finished_at":"0001-01-01T00:00:00Z","score":0}`},
		{name: "rejects invalid session transition", method: http.MethodPut, path: "/chat-sessions/2", body: `{"user_id":1,"chat_id":1,"status":"IN_PROGRESS","started_at":"2026-08-07T18:00:00Z","finished_at":"0001-01-01T00:00:00Z","score":0}`, wantStatus: 500, wantBody: "invalid chat session status transition"},
		{name: "rejects invalid identifier", method: http.MethodGet, path: "/users/nope", wantStatus: 400, wantBody: "invalid user id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if got := strings.TrimSpace(recorder.Body.String()); got != tt.wantBody {
				t.Fatalf("body = %q, want %q", got, tt.wantBody)
			}
		})
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

type fakeChats struct{}

func (fakeChats) Create(scenario domain.Scenario) (domain.Scenario, error) {
	scenario.ID = 2
	return scenario, nil
}
func (fakeChats) GetByID(id int) (domain.Scenario, error) { return domain.Scenario{ID: id}, nil }
func (fakeChats) Update(domain.Scenario) error            { return nil }
func (fakeChats) Delete(int) error                        { return nil }
func (fakeChats) List() ([]domain.Scenario, error)        { return nil, nil }

type fakeSessions struct{}

func (fakeSessions) Create(attempt domain.Attempt) (domain.Attempt, error) {
	attempt.ID = 1
	return attempt, nil
}
func (fakeSessions) GetByID(id int) (domain.Attempt, error) {
	status := domain.AttemptStatusInProgress
	if id == 2 {
		status = domain.AttemptStatusCompleted
	}
	return domain.Attempt{ID: id, UserID: 1, ChatID: 1, Status: status, StartedAt: mustTime("2026-08-07T18:00:00Z")}, nil
}
func (fakeSessions) Update(domain.Attempt) error     { return nil }
func (fakeSessions) Delete(int) error                { return nil }
func (fakeSessions) List() ([]domain.Attempt, error) { return nil, nil }
func mustTime(value string) (parsedTime time.Time) {
	parsedTime, _ = time.Parse(time.RFC3339, value)
	return parsedTime
}
