package http_contract_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/server/router"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	attemptshttp "anti-scam-trainer/backend/internal/features/attempts/transport/http"
	authservice "anti-scam-trainer/backend/internal/features/auth/service"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestHTTPGameContractExposesMixedStateAndRejectsAmbiguousAnswer(t *testing.T) {
	store := newHTTPGameStore()
	game := attemptsservice.NewGameWithAI(store, contractAI{})
	handler := router.New()
	handler.Register(router.V1, attemptshttp.NewGame(game).Routes())

	start := httptest.NewRequest(http.MethodPost, "/api/v1/training/levels/3/start?role=buyer&topic_id=1", nil)
	start = start.WithContext(authservice.WithIdentity(start.Context(), authservice.Identity{UserID: 1}))
	startRecorder := httptest.NewRecorder()
	handler.ServeHTTP(startRecorder, start)
	if body := startRecorder.Body.String(); startRecorder.Code != http.StatusOK || !strings.Contains(body, `"mode":"mixed"`) || !strings.Contains(body, `"messages"`) {
		t.Fatalf("level 3 start = (%d, %s), want mixed game state", startRecorder.Code, body)
	}
	resumed := httptest.NewRequest(http.MethodPost, "/api/v1/training/levels/3/start?role=buyer&topic_id=1", nil)
	resumed = resumed.WithContext(authservice.WithIdentity(resumed.Context(), authservice.Identity{UserID: 1}))
	resumedRecorder := httptest.NewRecorder()
	handler.ServeHTTP(resumedRecorder, resumed)
	if resumedRecorder.Code != http.StatusOK || !strings.Contains(resumedRecorder.Body.String(), `"attempt_id":1`) {
		t.Fatalf("resume = (%d, %s)", resumedRecorder.Code, resumedRecorder.Body.String())
	}
	restored := httptest.NewRequest(http.MethodGet, "/api/v1/attempts/1", nil)
	restored = restored.WithContext(authservice.WithIdentity(restored.Context(), authservice.Identity{UserID: 1}))
	restoredRecorder := httptest.NewRecorder()
	handler.ServeHTTP(restoredRecorder, restored)
	if body := restoredRecorder.Body.String(); restoredRecorder.Code != http.StatusOK || !strings.Contains(body, `"step":{"counterparty_message"`) || strings.Contains(body, "step_goal") {
		t.Fatalf("restore = (%d, %s), want frontend-safe game state", restoredRecorder.Code, body)
	}

	foreign := httptest.NewRequest(http.MethodPost, "/api/v1/attempts/1/answers", strings.NewReader(`{"step_id":31,"option_id":11}`))
	foreign = foreign.WithContext(authservice.WithIdentity(foreign.Context(), authservice.Identity{UserID: 2}))
	foreignRecorder := httptest.NewRecorder()
	handler.ServeHTTP(foreignRecorder, foreign)
	if foreignRecorder.Code != http.StatusNotFound {
		t.Fatalf("foreign answer = %d, want 404", foreignRecorder.Code)
	}

	ambiguous := httptest.NewRequest(http.MethodPost, "/api/v1/attempts/1/answers", strings.NewReader(`{"step_id":31,"option_id":11,"free_text":"мой ответ"}`))
	ambiguous = ambiguous.WithContext(authservice.WithIdentity(ambiguous.Context(), authservice.Identity{UserID: 1}))
	ambiguousRecorder := httptest.NewRecorder()
	handler.ServeHTTP(ambiguousRecorder, ambiguous)
	if ambiguousRecorder.Code != http.StatusConflict || len(store.answers) != 0 {
		t.Fatalf("ambiguous answer = (%d, %s), want 409 without writes", ambiguousRecorder.Code, ambiguousRecorder.Body.String())
	}

	stale := httptest.NewRequest(http.MethodPost, "/api/v1/attempts/1/answers", strings.NewReader(`{"step_id":999,"option_id":11}`))
	stale = stale.WithContext(authservice.WithIdentity(stale.Context(), authservice.Identity{UserID: 1}))
	staleRecorder := httptest.NewRecorder()
	handler.ServeHTTP(staleRecorder, stale)
	if body := staleRecorder.Body.String(); staleRecorder.Code != http.StatusConflict || !strings.Contains(body, `"code":"STALE_STEP"`) || len(store.answers) != 0 {
		t.Fatalf("stale answer = (%d, %s), want STALE_STEP without writes", staleRecorder.Code, body)
	}

	optionFinish := httptest.NewRequest(http.MethodPost, "/api/v1/attempts/1/answers", strings.NewReader(`{"step_id":31,"option_id":11,"finish":true}`))
	optionFinish = optionFinish.WithContext(authservice.WithIdentity(optionFinish.Context(), authservice.Identity{UserID: 1}))
	optionFinishRecorder := httptest.NewRecorder()
	handler.ServeHTTP(optionFinishRecorder, optionFinish)
	if optionFinishRecorder.Code != http.StatusConflict || len(store.answers) != 0 {
		t.Fatalf("option with finish = (%d, %s), want 409 without writes", optionFinishRecorder.Code, optionFinishRecorder.Body.String())
	}

	for _, body := range []string{`{"step_id":31,"option_id":11,"unknown":true}`, `{"step_id":31,"option_id":11} {"option_id":11}`} {
		invalid := httptest.NewRequest(http.MethodPost, "/api/v1/attempts/1/answers", strings.NewReader(body))
		invalid = invalid.WithContext(authservice.WithIdentity(invalid.Context(), authservice.Identity{UserID: 1}))
		invalidRecorder := httptest.NewRecorder()
		handler.ServeHTTP(invalidRecorder, invalid)
		if invalidRecorder.Code != http.StatusBadRequest || len(store.answers) != 0 {
			t.Fatalf("strict JSON %q = (%d, %s), want 400 without writes", body, invalidRecorder.Code, invalidRecorder.Body.String())
		}
	}
}

func TestHTTPAIFailureIsRetryableAndHasNoSideEffects(t *testing.T) {
	cases := []struct {
		name   string
		ai     attemptsservice.AIProvider
		status int
	}{
		{name: "timeout", ai: failingContractAI{}, status: http.StatusServiceUnavailable},
		{name: "invalid JSON", ai: invalidContractAI{}, status: http.StatusBadGateway},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			store := newHTTPGameStore()
			game := attemptsservice.NewGameWithAI(store, test.ai)
			handler := router.New()
			handler.Register(router.V1, attemptshttp.NewGame(game).Routes())
			start := httptest.NewRequest(http.MethodPost, "/api/v1/training/levels/3/start?role=buyer&topic_id=1", nil)
			start = start.WithContext(authservice.WithIdentity(start.Context(), authservice.Identity{UserID: 1}))
			handler.ServeHTTP(httptest.NewRecorder(), start)
			beforeMessages := len(store.messages)
			answer := httptest.NewRequest(http.MethodPost, "/api/v1/attempts/1/answers", strings.NewReader(`{"step_id":31,"free_text":"Останусь в сервисе"}`))
			answer = answer.WithContext(authservice.WithIdentity(answer.Context(), authservice.Identity{UserID: 1}))
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, answer)
			if recorder.Code != test.status || len(store.answers) != 0 || len(store.messages) != beforeMessages {
				t.Fatalf("AI failure = (%d, answers=%d, messages=%d), want %d without changes", recorder.Code, len(store.answers), len(store.messages), test.status)
			}
		})
	}
}

func TestHTTPFreePlayCoversBothRolesAndHidesCounterpartUntilCompletion(t *testing.T) {
	for _, role := range []string{"buyer", "seller"} {
		t.Run(role, func(t *testing.T) {
			store := newHTTPGameStore()
			game := attemptsservice.NewGameWithDependencies(store, contractAI{}, func() bool { return role == "buyer" })
			handler := router.New()
			handler.Register(router.V1, attemptshttp.NewGame(game).Routes())
			start := httptest.NewRequest(http.MethodPost, "/api/v1/training/free-play/start?role="+role, nil)
			start = start.WithContext(authservice.WithIdentity(start.Context(), authservice.Identity{UserID: 1}))
			startRecorder := httptest.NewRecorder()
			handler.ServeHTTP(startRecorder, start)
			if startRecorder.Code != http.StatusOK || strings.Contains(startRecorder.Body.String(), "is_scam") {
				t.Fatalf("start = (%d,%s), type must stay hidden", startRecorder.Code, startRecorder.Body.String())
			}
			wrong := httptest.NewRequest(http.MethodPost, "/api/v1/attempts/1/answers", strings.NewReader(`{"step_id":1,"option_id":11}`))
			wrong = wrong.WithContext(authservice.WithIdentity(wrong.Context(), authservice.Identity{UserID: 1}))
			wrongRecorder := httptest.NewRecorder()
			handler.ServeHTTP(wrongRecorder, wrong)
			if wrongRecorder.Code != http.StatusConflict {
				t.Fatalf("option in free play = %d, want 409", wrongRecorder.Code)
			}
			foreign := httptest.NewRequest(http.MethodPost, "/api/v1/attempts/1/answers", strings.NewReader(`{"step_id":1,"free_text":"Чужой ответ"}`))
			foreign = foreign.WithContext(authservice.WithIdentity(foreign.Context(), authservice.Identity{UserID: 2}))
			foreignRecorder := httptest.NewRecorder()
			handler.ServeHTTP(foreignRecorder, foreign)
			if foreignRecorder.Code != http.StatusNotFound {
				t.Fatalf("foreign free play answer = %d, want 404", foreignRecorder.Code)
			}
			for turn := 1; turn <= 3; turn++ {
				body := `{"step_id":` + strconv.Itoa(turn) + `,"free_text":"Безопасный ответ"}`
				if turn == 3 {
					body = `{"step_id":3,"free_text":"Безопасный ответ","finish":true}`
				}
				request := httptest.NewRequest(http.MethodPost, "/api/v1/attempts/1/answers", strings.NewReader(body))
				request = request.WithContext(authservice.WithIdentity(request.Context(), authservice.Identity{UserID: 1}))
				recorder := httptest.NewRecorder()
				handler.ServeHTTP(recorder, request)
				if recorder.Code != http.StatusOK {
					t.Fatalf("turn %d = (%d,%s)", turn, recorder.Code, recorder.Body.String())
				}
				if turn < 3 && strings.Contains(recorder.Body.String(), "is_scam") {
					t.Fatalf("turn %d revealed type", turn)
				}
				if turn == 1 {
					stale := httptest.NewRequest(http.MethodPost, "/api/v1/attempts/1/answers", strings.NewReader(`{"step_id":1,"free_text":"Повтор"}`))
					stale = stale.WithContext(authservice.WithIdentity(stale.Context(), authservice.Identity{UserID: 1}))
					staleRecorder := httptest.NewRecorder()
					handler.ServeHTTP(staleRecorder, stale)
					if staleRecorder.Code != http.StatusConflict || !strings.Contains(staleRecorder.Body.String(), `"code":"STALE_STEP"`) {
						t.Fatalf("repeated free-play step = (%d,%s)", staleRecorder.Code, staleRecorder.Body.String())
					}
				}
				if turn == 3 && !strings.Contains(recorder.Body.String(), `"is_scam"`) {
					t.Fatalf("completion lacks reveal: %s", recorder.Body.String())
				}
			}
		})
	}
}

func TestHTTPFreePlayAIFailureHasNoSideEffects(t *testing.T) {
	store := newHTTPGameStore()
	ai := &sequenceContractAI{}
	game := attemptsservice.NewGameWithDependencies(store, ai, func() bool { return true })
	handler := router.New()
	handler.Register(router.V1, attemptshttp.NewGame(game).Routes())
	start := httptest.NewRequest(http.MethodPost, "/api/v1/training/free-play/start?role=buyer", nil)
	start = start.WithContext(authservice.WithIdentity(start.Context(), authservice.Identity{UserID: 1}))
	handler.ServeHTTP(httptest.NewRecorder(), start)
	beforeMessages := len(store.messages)
	answer := httptest.NewRequest(http.MethodPost, "/api/v1/attempts/1/answers", strings.NewReader(`{"step_id":1,"free_text":"Безопасный ответ"}`))
	answer = answer.WithContext(authservice.WithIdentity(answer.Context(), authservice.Identity{UserID: 1}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, answer)
	if recorder.Code != http.StatusServiceUnavailable || len(store.answers) != 0 || len(store.messages) != beforeMessages || store.attempts[1].FreeTextCount != 0 {
		t.Fatalf("free play AI failure = (%d, answers=%d, messages=%d), want 503 without changes", recorder.Code, len(store.answers), len(store.messages))
	}
}

type contractAI struct{}

func (contractAI) Generate(context.Context, []attemptsservice.AIMessage) (string, error) {
	return `{"awarded_points":100,"explanation":"безопасно","reply":"продолжим","risk_signals":[]}`, nil
}

type failingContractAI struct{}

func (failingContractAI) Generate(context.Context, []attemptsservice.AIMessage) (string, error) {
	return "", attemptsservice.ErrAIUnavailable
}

type invalidContractAI struct{}

func (invalidContractAI) Generate(context.Context, []attemptsservice.AIMessage) (string, error) {
	return `{"awarded_points":100`, nil
}

type sequenceContractAI struct{ calls int }

func (a *sequenceContractAI) Generate(context.Context, []attemptsservice.AIMessage) (string, error) {
	a.calls++
	if a.calls > 1 {
		return "", attemptsservice.ErrAIUnavailable
	}
	return `{"awarded_points":0,"explanation":"старт","reply":"начнём","risk_signals":[]}`, nil
}

type httpGameStore struct {
	attempts map[int]domain.Attempt
	answers  []domain.UserAnswer
	messages []domain.DialogueMessage
}

func newHTTPGameStore() *httpGameStore { return &httpGameStore{attempts: map[int]domain.Attempt{}} }
func (s *httpGameStore) Levels(int, string) ([]domain.Level, []domain.Progress, error) {
	return []domain.Level{{ID: 1, Number: 1}, {ID: 2, Number: 2}, {ID: 3, Number: 3}, {ID: 4, Number: 4}}, []domain.Progress{{LevelID: 1, Stars: 1}, {LevelID: 2, Stars: 1}, {LevelID: 4, Stars: 1}}, nil
}
func (s *httpGameStore) PublishedScenario(level int, role string) (domain.Scenario, error) {
	if level != 3 {
		return domain.Scenario{}, errors.New("missing")
	}
	return domain.Scenario{ID: 3, LevelID: 3, UserRole: role}, nil
}
func (s *httpGameStore) FreePlayConfig(role string) (domain.FreePlayConfig, error) {
	return domain.FreePlayConfig{UserRole: role}, nil
}
func (s *httpGameStore) Scenario(int) (domain.Scenario, error) {
	return domain.Scenario{ID: 3, LevelID: 3, UserRole: "buyer"}, nil
}
func (s *httpGameStore) FindInProgress(userID, scenarioID int) (domain.Attempt, error) {
	for _, attempt := range s.attempts {
		if attempt.UserID == userID && attempt.ScenarioID == scenarioID {
			return attempt, nil
		}
	}
	return domain.Attempt{}, errors.New("missing")
}
func (s *httpGameStore) FindInProgressFreePlay(int, string) (domain.Attempt, error) {
	return domain.Attempt{}, errors.New("missing")
}
func (s *httpGameStore) CreateGameAttempt(attempt domain.Attempt) (domain.Attempt, error) {
	attempt.ID = 1
	s.attempts[1] = attempt
	return attempt, nil
}
func (s *httpGameStore) StartFreePlay(attempt domain.Attempt, message domain.DialogueMessage) (domain.Attempt, error) {
	created, err := s.CreateGameAttempt(attempt)
	if err != nil {
		return domain.Attempt{}, err
	}
	message.AttemptID = created.ID
	s.messages = append(s.messages, message)
	return created, nil
}
func (s *httpGameStore) GetGameAttempt(id int) (domain.Attempt, error) {
	a, ok := s.attempts[id]
	if !ok {
		return domain.Attempt{}, errors.New("missing")
	}
	return a, nil
}
func (s *httpGameStore) Step(_ int, number int) (domain.ScenarioStep, error) {
	if number != 1 {
		return domain.ScenarioStep{}, errors.New("missing")
	}
	return domain.ScenarioStep{ID: 31, ScenarioID: 3, Number: 1, ResponseType: "mixed", FallbackMessage: "Начальная реплика", Options: []domain.ScenarioOption{{ID: 11, Points: 100}}}, nil
}
func (s *httpGameStore) Answers(int) ([]domain.UserAnswer, error) {
	return append([]domain.UserAnswer(nil), s.answers...), nil
}
func (s *httpGameStore) Messages(int) ([]domain.DialogueMessage, error) {
	return append([]domain.DialogueMessage(nil), s.messages...), nil
}
func (s *httpGameStore) AwardedPoints(int) (int, error) {
	total := 0
	for _, a := range s.answers {
		total += a.AwardedPoints
	}
	return total, nil
}
func (s *httpGameStore) Advance(id, next int) error {
	a := s.attempts[id]
	a.CurrentStepNumber = next
	s.attempts[id] = a
	return nil
}
func (s *httpGameStore) Abandon(id int, _ time.Time) error {
	a := s.attempts[id]
	a.Status = domain.AttemptStatusAbandoned
	s.attempts[id] = a
	return nil
}
func (s *httpGameStore) Complete(action func(attemptsservice.GameCompletionStore) error) error {
	return action(s)
}
func (s *httpGameStore) SaveAnswer(answer domain.UserAnswer) error {
	s.answers = append(s.answers, answer)
	return nil
}
func (s *httpGameStore) SaveMessage(message domain.DialogueMessage) error {
	s.messages = append(s.messages, message)
	return nil
}
func (s *httpGameStore) AdvanceAttempt(id, next int) error { return s.Advance(id, next) }
func (s *httpGameStore) UpdateFreeTextCount(id, count int) error {
	a := s.attempts[id]
	a.FreeTextCount = count
	s.attempts[id] = a
	return nil
}
func (s *httpGameStore) CompleteAttempt(attempt domain.Attempt) error {
	s.attempts[attempt.ID] = attempt
	return nil
}
func (s *httpGameStore) SaveProgress(domain.Progress) error           { return nil }
func (s *httpGameStore) FinalizeLearning(*domain.AttemptResult) error { return nil }
