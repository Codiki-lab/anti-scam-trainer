package http_contract_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/server/middleware"
	"anti-scam-trainer/backend/internal/core/server/router"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	attemptshttp "anti-scam-trainer/backend/internal/features/attempts/transport/http"
	authservice "anti-scam-trainer/backend/internal/features/auth/service"
	authhttp "anti-scam-trainer/backend/internal/features/auth/transport/http"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestGameHTTPContractGuardsProgressionAndReturnsCompletionBreakdown(t *testing.T) {
	repository := newHTTPGameRepository()
	versionedRouter := router.New()
	versionedRouter.Register(router.V1, attemptshttp.NewGame(attemptsservice.NewGame(repository)).Routes())
	handler := middleware.RequestID()(authhttp.RequireAuthentication(contractTokens{})(versionedRouter))
	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/training/levels?role=buyer", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("missing cookie = %d, want 401", unauthorized.Code)
	}

	levels := serveGame(handler, 1, http.MethodGet, "/api/v1/training/levels?role=buyer", nil)
	if levels.Code != http.StatusOK || !bytes.Contains(levels.Body.Bytes(), []byte(`"opened":false`)) {
		t.Fatalf("levels = (%d, %s), want buyer level 2 closed", levels.Code, levels.Body.String())
	}
	if levels.Header().Get("X-Request-ID") == "" {
		t.Fatal("levels response has no X-Request-ID")
	}
	closed := serveGame(handler, 1, http.MethodPost, "/api/v1/training/levels/2/start?role=buyer", nil)
	if closed.Code != http.StatusForbidden {
		t.Fatalf("closed level status = %d, want %d", closed.Code, http.StatusForbidden)
	}

	started := serveGame(handler, 1, http.MethodPost, "/api/v1/training/levels/1/start?role=buyer", nil)
	if started.Code != http.StatusOK || !bytes.Contains(started.Body.Bytes(), []byte(`"number":1`)) {
		t.Fatalf("start = (%d, %s), want first step", started.Code, started.Body.String())
	}
	var state struct {
		AttemptID int `json:"attempt_id"`
	}
	if err := json.Unmarshal(started.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	skipped := serveGame(handler, 1, http.MethodPost, "/api/v1/attempts/1/answers", []byte(`{"option_id":21}`))
	if skipped.Code != http.StatusConflict {
		t.Fatalf("skipped answer = %d, want 409", skipped.Code)
	}

	first := serveGame(handler, 1, http.MethodPost, "/api/v1/attempts/1/answers", []byte(`{"option_id":11}`))
	if first.Code != http.StatusOK || !bytes.Contains(first.Body.Bytes(), []byte(`"number":2`)) || bytes.Contains(first.Body.Bytes(), []byte(`"score"`)) {
		t.Fatalf("intermediate answer = (%d, %s), want next step without score", first.Code, first.Body.String())
	}
	duplicate := serveGame(handler, 1, http.MethodPost, "/api/v1/attempts/1/answers", []byte(`{"option_id":11}`))
	if duplicate.Code != http.StatusConflict {
		t.Fatalf("duplicate answer = %d, want 409", duplicate.Code)
	}
	resumed := serveGame(handler, 1, http.MethodPost, "/api/v1/training/levels/1/start?role=buyer", nil)
	if resumed.Code != http.StatusOK || !bytes.Contains(resumed.Body.Bytes(), []byte(`"option_id":11`)) || !bytes.Contains(resumed.Body.Bytes(), []byte(`"author":"user"`)) || !bytes.Contains(resumed.Body.Bytes(), []byte(`"author":"interlocutor"`)) || !bytes.Contains(resumed.Body.Bytes(), []byte(`"text":"safe"`)) {
		t.Fatalf("resume = (%d, %s), want answer history", resumed.Code, resumed.Body.String())
	}
	foreign := serveGame(handler, 2, http.MethodPost, "/api/v1/attempts/1/answers", []byte(`{"option_id":21}`))
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign answer = %d, want 404", foreign.Code)
	}
	completed := serveGame(handler, 1, http.MethodPost, "/api/v1/attempts/1/answers", []byte(`{"option_id":21}`))
	if completed.Code != http.StatusOK || !bytes.Contains(completed.Body.Bytes(), []byte(`"score":100`)) || !bytes.Contains(completed.Body.Bytes(), []byte(`"option_text":"safe"`)) {
		t.Fatalf("completion = (%d, %s), want score and selected variant", completed.Code, completed.Body.String())
	}
	if state.AttemptID != 1 || repository.progressByRole["buyer"].Stars != 3 {
		t.Fatalf("completion did not persist buyer progress: attempt=%d progress=%#v", state.AttemptID, repository.progressByRole)
	}
}

func TestGameHTTPContractKeepsRolesIndependentAndSupportsAbandon(t *testing.T) {
	repository := newHTTPGameRepository()
	versionedRouter := router.New()
	versionedRouter.Register(router.V1, attemptshttp.NewGame(attemptsservice.NewGame(repository)).Routes())
	handler := middleware.RequestID()(authhttp.RequireAuthentication(contractTokens{})(versionedRouter))

	repository.progressByRole["buyer"] = domain.Progress{UserID: 1, LevelID: 1, UserRole: "buyer", Stars: 1}
	buyer := serveGame(handler, 1, http.MethodGet, "/api/v1/training/levels?role=buyer", nil)
	seller := serveGame(handler, 1, http.MethodGet, "/api/v1/training/levels?role=seller", nil)
	if !bytes.Contains(buyer.Body.Bytes(), []byte(`"number":2,"opened":true`)) || !bytes.Contains(seller.Body.Bytes(), []byte(`"number":2,"opened":false`)) {
		t.Fatalf("role levels buyer=%s seller=%s", buyer.Body.String(), seller.Body.String())
	}

	started := serveGame(handler, 1, http.MethodPost, "/api/v1/training/levels/1/start?role=seller", nil)
	if started.Code != http.StatusOK {
		t.Fatalf("seller start = %d", started.Code)
	}
	foreign := serveGame(handler, 2, http.MethodPost, "/api/v1/attempts/1/abandon", nil)
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign abandon = %d, want 404", foreign.Code)
	}
	abandoned := serveGame(handler, 1, http.MethodPost, "/api/v1/attempts/1/abandon", nil)
	if abandoned.Code != http.StatusNoContent || repository.attempts[1].Status != domain.AttemptStatusAbandoned {
		t.Fatalf("abandon = (%d, %#v)", abandoned.Code, repository.attempts[1])
	}

	repository.published["buyer"][1] = false
	hidden := serveGame(handler, 1, http.MethodPost, "/api/v1/training/levels/1/start?role=buyer", nil)
	if hidden.Code != http.StatusNotFound {
		t.Fatalf("draft/archived scenario start = %d, want 404", hidden.Code)
	}
}

func TestGameHTTPContractStartsAllFourPublishedScenarios(t *testing.T) {
	repository := newHTTPGameRepository()
	repository.progressByRole["buyer"] = domain.Progress{UserID: 1, LevelID: 1, UserRole: "buyer", Stars: 1}
	repository.progressByRole["seller"] = domain.Progress{UserID: 1, LevelID: 1, UserRole: "seller", Stars: 1}
	versionedRouter := router.New()
	versionedRouter.Register(router.V1, attemptshttp.NewGame(attemptsservice.NewGame(repository)).Routes())
	handler := middleware.RequestID()(authhttp.RequireAuthentication(contractTokens{})(versionedRouter))

	paths := []string{
		"/api/v1/training/levels/1/start?role=buyer",
		"/api/v1/training/levels/2/start?role=buyer",
		"/api/v1/training/levels/1/start?role=seller",
		"/api/v1/training/levels/2/start?role=seller",
	}
	for index, path := range paths {
		response := serveGame(handler, index+1, http.MethodPost, path, nil)
		if response.Code != http.StatusOK {
			t.Fatalf("start %s = %d, want 200: %s", path, response.Code, response.Body.String())
		}
	}
}

func serveGame(handler http.Handler, userID int, method, target string, body []byte) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, bytes.NewReader(body))
	request.AddCookie(&http.Cookie{Name: authhttp.AccessTokenCookie, Value: "user-" + strconv.Itoa(userID)})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

type contractTokens struct{}

func (contractTokens) Issue(domain.User) (string, error) { return "", nil }
func (contractTokens) Parse(raw string) (authservice.Identity, error) {
	parts := strings.Split(raw, "-")
	if len(parts) != 2 {
		return authservice.Identity{}, errors.New("invalid")
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return authservice.Identity{}, err
	}
	role := domain.AccessRole(parts[0])
	if role != domain.AccessRoleUser && role != domain.AccessRoleAdmin {
		return authservice.Identity{}, errors.New("invalid role")
	}
	return authservice.Identity{UserID: id, AccessRole: role}, nil
}

type httpGameRepository struct {
	attempts       map[int]domain.Attempt
	answers        []domain.UserAnswer
	messages       map[int][]domain.Message
	progressByRole map[string]domain.Progress
	published      map[string]map[int]bool
}

func newHTTPGameRepository() *httpGameRepository {
	return &httpGameRepository{
		attempts: map[int]domain.Attempt{}, messages: map[int][]domain.Message{}, progressByRole: map[string]domain.Progress{},
		published: map[string]map[int]bool{"buyer": {1: true, 2: true}, "seller": {1: true, 2: true}},
	}
}
func (r *httpGameRepository) Levels(_ int, role string) ([]domain.Level, []domain.Progress, error) {
	progress := []domain.Progress{}
	if item, ok := r.progressByRole[role]; ok {
		progress = append(progress, item)
	}
	return []domain.Level{{ID: 1, Number: 1}, {ID: 2, Number: 2}}, progress, nil
}
func (r *httpGameRepository) PublishedScenario(level int, role string) (domain.Scenario, error) {
	if r.published[role][level] {
		id := level
		if role == "seller" {
			id += 2
		}
		return domain.Scenario{ID: id, LevelID: level, UserRole: role}, nil
	}
	return domain.Scenario{}, errors.New("missing")
}
func (*httpGameRepository) Scenario(id int) (domain.Scenario, error) {
	role, level := "buyer", id
	if id > 2 {
		role, level = "seller", id-2
	}
	return domain.Scenario{ID: id, LevelID: level, UserRole: role}, nil
}
func (r *httpGameRepository) FindInProgress(userID, scenarioID int) (domain.Attempt, error) {
	for _, attempt := range r.attempts {
		if attempt.UserID == userID && attempt.ScenarioID == scenarioID && attempt.Status == domain.AttemptStatusInProgress {
			return attempt, nil
		}
	}
	return domain.Attempt{}, errors.New("missing")
}
func (r *httpGameRepository) CreateGameAttempt(attempt domain.Attempt, initialMessage domain.Message) (domain.Attempt, error) {
	attempt.ID = len(r.attempts) + 1
	r.attempts[attempt.ID] = attempt
	initialMessage.ID, initialMessage.AttemptID, initialMessage.CreatedAt = 1, attempt.ID, time.Unix(1, 0).UTC()
	r.messages[attempt.ID] = append(r.messages[attempt.ID], initialMessage)
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
func (r *httpGameRepository) Messages(attemptID int) ([]domain.Message, error) {
	return r.messages[attemptID], nil
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
	answer.AwardedPoints, answer.Explanation = points, explanation
	r.answers = append(r.answers, answer)
	return nil
}
func (r *httpGameRepository) SaveMessage(message domain.Message) error {
	message.ID = len(r.messages[message.AttemptID]) + 1
	message.CreatedAt = time.Unix(int64(message.ID), 0).UTC()
	r.messages[message.AttemptID] = append(r.messages[message.AttemptID], message)
	return nil
}
func (r *httpGameRepository) AdvanceAttempt(id, next int) error { return r.Advance(id, next) }
func (r *httpGameRepository) CompleteAttempt(attempt domain.Attempt) error {
	r.attempts[attempt.ID] = attempt
	return nil
}
func (r *httpGameRepository) SaveProgress(progress domain.Progress) error {
	r.progressByRole[progress.UserRole] = progress
	return nil
}
