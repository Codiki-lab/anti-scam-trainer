package http_contract_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/server/router"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	attemptshttp "anti-scam-trainer/backend/internal/features/attempts/transport/http"
	authservice "anti-scam-trainer/backend/internal/features/auth/service"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGameHTTPContractGuardsProgressionAndReturnsCompletionBreakdown(t *testing.T) {
	repository := newHTTPGameRepository()
	versionedRouter := router.New()
	versionedRouter.Register(router.V1, attemptshttp.NewGame(attemptsservice.NewGame(repository)).Routes())

	levels := serveGame(versionedRouter, 1, http.MethodGet, "/api/v1/training/levels?role=buyer", nil)
	if levels.Code != http.StatusOK || !bytes.Contains(levels.Body.Bytes(), []byte(`"opened":false`)) {
		t.Fatalf("levels = (%d, %s), want buyer level 2 closed", levels.Code, levels.Body.String())
	}
	closed := serveGame(versionedRouter, 1, http.MethodPost, "/api/v1/training/levels/2/start?role=buyer", nil)
	if closed.Code != http.StatusForbidden {
		t.Fatalf("closed level status = %d, want %d", closed.Code, http.StatusForbidden)
	}

	started := serveGame(versionedRouter, 1, http.MethodPost, "/api/v1/training/levels/1/start?role=buyer", nil)
	if started.Code != http.StatusOK || !bytes.Contains(started.Body.Bytes(), []byte(`"number":1`)) {
		t.Fatalf("start = (%d, %s), want first step", started.Code, started.Body.String())
	}
	var state struct {
		AttemptID int `json:"attempt_id"`
	}
	if err := json.Unmarshal(started.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}

	first := serveGame(versionedRouter, 1, http.MethodPost, "/api/v1/attempts/1/answers", []byte(`{"option_id":11}`))
	if first.Code != http.StatusOK || !bytes.Contains(first.Body.Bytes(), []byte(`"number":2`)) || bytes.Contains(first.Body.Bytes(), []byte(`"score"`)) {
		t.Fatalf("intermediate answer = (%d, %s), want next step without score", first.Code, first.Body.String())
	}
	resumed := serveGame(versionedRouter, 1, http.MethodPost, "/api/v1/training/levels/1/start?role=buyer", nil)
	if resumed.Code != http.StatusOK || !bytes.Contains(resumed.Body.Bytes(), []byte(`"option_id":11`)) {
		t.Fatalf("resume = (%d, %s), want answer history", resumed.Code, resumed.Body.String())
	}
	foreign := serveGame(versionedRouter, 2, http.MethodPost, "/api/v1/attempts/1/answers", []byte(`{"option_id":21}`))
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign answer = %d, want 404", foreign.Code)
	}
	completed := serveGame(versionedRouter, 1, http.MethodPost, "/api/v1/attempts/1/answers", []byte(`{"option_id":21}`))
	if completed.Code != http.StatusOK || !bytes.Contains(completed.Body.Bytes(), []byte(`"score":100`)) || !bytes.Contains(completed.Body.Bytes(), []byte(`"option_text":"safe"`)) {
		t.Fatalf("completion = (%d, %s), want score and selected variant", completed.Code, completed.Body.String())
	}
	if state.AttemptID != 1 || repository.progress.UserRole != "buyer" || repository.progress.Stars != 3 {
		t.Fatalf("completion did not persist buyer progress: attempt=%d progress=%#v", state.AttemptID, repository.progress)
	}
}

func serveGame(handler http.Handler, userID int, method, target string, body []byte) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, bytes.NewReader(body))
	request = request.WithContext(authservice.WithIdentity(request.Context(), authservice.Identity{UserID: userID, AccessRole: domain.AccessRoleUser}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

type httpGameRepository struct {
	attempts map[int]domain.Attempt
	answers  []domain.UserAnswer
	progress domain.Progress
}

func newHTTPGameRepository() *httpGameRepository {
	return &httpGameRepository{attempts: map[int]domain.Attempt{}}
}
func (*httpGameRepository) Levels(int, string) ([]domain.Level, []domain.Progress, error) {
	return []domain.Level{{ID: 1, Number: 1}, {ID: 2, Number: 2}}, nil, nil
}
func (*httpGameRepository) PublishedScenario(level int, role string) (domain.Scenario, error) {
	if level > 0 && level < 3 && (role == "buyer" || role == "seller") {
		return domain.Scenario{ID: level, LevelID: level, UserRole: role}, nil
	}
	return domain.Scenario{}, errors.New("missing")
}
func (*httpGameRepository) Scenario(id int) (domain.Scenario, error) {
	return domain.Scenario{ID: id, LevelID: id, UserRole: "buyer"}, nil
}
func (r *httpGameRepository) FindInProgress(userID, scenarioID int) (domain.Attempt, error) {
	for _, attempt := range r.attempts {
		if attempt.UserID == userID && attempt.ScenarioID == scenarioID && attempt.Status == domain.AttemptStatusInProgress {
			return attempt, nil
		}
	}
	return domain.Attempt{}, errors.New("missing")
}
func (r *httpGameRepository) CreateGameAttempt(attempt domain.Attempt) (domain.Attempt, error) {
	attempt.ID = len(r.attempts) + 1
	r.attempts[attempt.ID] = attempt
	return attempt, nil
}
func (r *httpGameRepository) GetGameAttempt(id int) (domain.Attempt, error) {
	attempt, ok := r.attempts[id]
	if !ok {
		return domain.Attempt{}, errors.New("missing")
	}
	return attempt, nil
}
func (*httpGameRepository) Step(scenarioID, number int) (domain.ScenarioStep, error) {
	if number < 1 || number > 2 {
		return domain.ScenarioStep{}, errors.New("missing")
	}
	optionID := number*10 + 1
	return domain.ScenarioStep{ID: number, ScenarioID: scenarioID, Number: number, Goal: "choose", MaxPoints: 100, Options: []domain.ScenarioOption{{ID: optionID, Text: "safe", Explanation: "safe action", Points: 100}}}, nil
}
func (r *httpGameRepository) Answers(attemptID int) ([]domain.UserAnswer, error) {
	result := []domain.UserAnswer{}
	for _, answer := range r.answers {
		if answer.AttemptID == attemptID {
			result = append(result, answer)
		}
	}
	return result, nil
}
func (r *httpGameRepository) AwardedPoints(attemptID int) (int, error) {
	total := 0
	for _, answer := range r.answers {
		if answer.AttemptID == attemptID {
			total += answer.AwardedPoints
		}
	}
	return total, nil
}
func (r *httpGameRepository) Advance(id, next int) error {
	attempt := r.attempts[id]
	attempt.CurrentStepNumber = next
	r.attempts[id] = attempt
	return nil
}
func (r *httpGameRepository) Abandon(id int, _ time.Time) error {
	attempt := r.attempts[id]
	attempt.Status = domain.AttemptStatusAbandoned
	r.attempts[id] = attempt
	return nil
}
func (r *httpGameRepository) Complete(action func(attemptsservice.GameCompletionStore) error) error {
	return action(r)
}
func (r *httpGameRepository) SaveAnswer(answer domain.UserAnswer, points int, explanation string) error {
	answer.AwardedPoints, answer.Explanation, answer.OptionText = points, explanation, "safe"
	r.answers = append(r.answers, answer)
	return nil
}
func (r *httpGameRepository) AdvanceAttempt(id, next int) error { return r.Advance(id, next) }
func (r *httpGameRepository) CompleteAttempt(attempt domain.Attempt) error {
	r.attempts[attempt.ID] = attempt
	return nil
}
func (r *httpGameRepository) SaveProgress(progress domain.Progress) error {
	r.progress = progress
	return nil
}
