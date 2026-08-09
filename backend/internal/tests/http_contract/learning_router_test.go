package http_contract_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/server/router"
	authservice "anti-scam-trainer/backend/internal/features/auth/service"
	learningservice "anti-scam-trainer/backend/internal/features/learning/service"
	learninghttp "anti-scam-trainer/backend/internal/features/learning/transport/http"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPLearningContractKeepsTheoryIdempotentAndQuizAnswersPrivate(t *testing.T) {
	store := &learningStore{}
	handler := router.New()
	handler.Register(router.V1, learninghttp.New(learningservice.New(store)).Routes())

	markTheory := func() *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/topics/1/theory/read", nil)
		request = request.WithContext(authservice.WithIdentity(request.Context(), authservice.Identity{UserID: 7}))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder
	}
	first, second := markTheory(), markTheory()
	if first.Code != http.StatusOK || !strings.Contains(first.Body.String(), `"newly_read":true`) {
		t.Fatalf("first theory read = (%d,%s)", first.Code, first.Body.String())
	}
	if second.Code != http.StatusOK || !strings.Contains(second.Body.String(), `"newly_read":false`) || store.activityCalls != 2 {
		t.Fatalf("repeated theory read = (%d,%s), calls=%d", second.Code, second.Body.String(), store.activityCalls)
	}

	quiz := httptest.NewRequest(http.MethodGet, "/api/v1/topics/1/quiz", nil)
	quiz = quiz.WithContext(authservice.WithIdentity(quiz.Context(), authservice.Identity{UserID: 7}))
	quizRecorder := httptest.NewRecorder()
	handler.ServeHTTP(quizRecorder, quiz)
	if body := quizRecorder.Body.String(); quizRecorder.Code != http.StatusOK || strings.Contains(body, "correct") || strings.Contains(body, "explanation") || !strings.Contains(body, `"pass_threshold":80`) {
		t.Fatalf("quiz = (%d,%s), want private answers", quizRecorder.Code, body)
	}

	answers := `{"answers":[{"question_id":1,"option_id":11},{"question_id":2,"option_id":21},{"question_id":3,"option_id":31},{"question_id":4,"option_id":41},{"question_id":5,"option_id":51}]}`
	attempt := httptest.NewRequest(http.MethodPost, "/api/v1/topics/1/quiz/attempts", strings.NewReader(answers))
	attempt = attempt.WithContext(authservice.WithIdentity(attempt.Context(), authservice.Identity{UserID: 7}))
	attemptRecorder := httptest.NewRecorder()
	handler.ServeHTTP(attemptRecorder, attempt)
	if body := attemptRecorder.Body.String(); attemptRecorder.Code != http.StatusOK || !strings.Contains(body, `"passed":true`) || !strings.Contains(body, `"score":80`) {
		t.Fatalf("quiz attempt = (%d,%s)", attemptRecorder.Code, body)
	}
}

func TestHTTPDashboardUsesServerContinuePriorityAndRoleIsolation(t *testing.T) {
	store := &learningStore{attemptID: 42}
	handler := router.New()
	handler.Register(router.V1, learninghttp.New(learningservice.New(store)).Routes())
	request := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard?role=seller", nil)
	request = request.WithContext(authservice.WithIdentity(request.Context(), authservice.Identity{UserID: 7}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if body := recorder.Body.String(); recorder.Code != http.StatusOK || !strings.Contains(body, `"type":"resume_attempt"`) || !strings.Contains(body, `"attempt_id":42`) || store.lastRole != domain.UserRoleSeller || !strings.Contains(body, `"daily_task":null`) {
		t.Fatalf("dashboard = (%d,%s), role=%s", recorder.Code, body, store.lastRole)
	}
}

type learningStore struct {
	theoryRead    bool
	activityCalls int
	attemptID     int
	lastRole      domain.UserRole
}

func (s *learningStore) Topics(_ int, role domain.UserRole) ([]domain.Topic, error) {
	s.lastRole = role
	return []domain.Topic{{ID: 1, Slug: string(role) + "-topic", UserRole: role, Title: "Тема", Description: "Описание", SortOrder: 1, TheoryRead: s.theoryRead, Levels: []domain.TopicLevelProgress{{Number: 1, Opened: true}, {Number: 2}, {Number: 3}, {Number: 4}}}}, nil
}

func (s *learningStore) Topic(_ int, topicID int) (domain.Topic, error) {
	return domain.Topic{ID: topicID, UserRole: domain.UserRoleBuyer, TheoryRead: s.theoryRead}, nil
}

func (s *learningStore) Theory(topicID int) ([]domain.TheoryBlock, error) {
	return []domain.TheoryBlock{{ID: 1, TopicID: topicID, SortOrder: 1, Kind: "intro", Title: "Введение", Body: "Текст"}}, nil
}

func (s *learningStore) MarkTheoryRead(_ int, _ int, _ time.Time) (domain.Streak, bool, error) {
	s.activityCalls++
	newlyRead := !s.theoryRead
	s.theoryRead = true
	return domain.Streak{Current: 1, Longest: 1, ActiveToday: true, LastActivityDate: "2026-08-09"}, newlyRead, nil
}

func (s *learningStore) Quiz(int) ([]domain.QuizQuestion, error) {
	questions := make([]domain.QuizQuestion, 5)
	for i := range questions {
		questions[i] = domain.QuizQuestion{ID: i + 1, SortOrder: i + 1, Text: "Вопрос", Explanation: "Скрыто", Options: []domain.QuizOption{{ID: i*10 + 1, Text: "Вариант", Correct: true}, {ID: i*10 + 2, Text: "Вариант"}, {ID: i*10 + 3, Text: "Вариант"}, {ID: i*10 + 4, Text: "Вариант"}}}
	}
	return questions, nil
}

func (s *learningStore) SubmitQuiz(_ int, _ int, _ []domain.QuizAnswer, _ time.Time) (domain.QuizResult, error) {
	return domain.QuizResult{Score: 80, Passed: true, BestScore: 80, NewlyPassed: true, Streak: domain.Streak{Current: 1, Longest: 1, ActiveToday: true}}, nil
}

func (s *learningStore) Achievements(int) ([]domain.Achievement, error) {
	return []domain.Achievement{}, nil
}
func (s *learningStore) User(int) (domain.User, error) {
	return domain.User{ID: 7, Username: "alex", TrainingRole: domain.UserRoleBuyer}, nil
}
func (s *learningStore) InProgressAttempt(_ int, role domain.UserRole) (int, int, int, error) {
	s.lastRole = role
	return s.attemptID, 1, 2, nil
}
