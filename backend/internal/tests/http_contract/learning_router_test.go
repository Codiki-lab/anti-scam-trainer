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
	lockedQuiz := httptest.NewRequest(http.MethodGet, "/api/v1/topics/1/quiz", nil)
	lockedQuiz = lockedQuiz.WithContext(authservice.WithIdentity(lockedQuiz.Context(), authservice.Identity{UserID: 7}))
	lockedRecorder := httptest.NewRecorder()
	handler.ServeHTTP(lockedRecorder, lockedQuiz)
	if body := lockedRecorder.Body.String(); lockedRecorder.Code != http.StatusForbidden || !strings.Contains(body, `"code":"CONTENT_UNAVAILABLE"`) {
		t.Fatalf("locked quiz = (%d,%s)", lockedRecorder.Code, body)
	}

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
	if body := recorder.Body.String(); recorder.Code != http.StatusOK || !strings.Contains(body, `"type":"resume_attempt"`) || !strings.Contains(body, `"attempt_id":42`) || !strings.Contains(body, `"role":"seller"`) || !strings.Contains(body, `"daily_task":{"date":`) || !strings.Contains(body, `"messages":`) {
		t.Fatalf("dashboard = (%d,%s), role=%s", recorder.Code, body, store.lastRole)
	}
}

func TestHTTPDashboardRecommendationIsStableForMoscowDateAndRole(t *testing.T) {
	now := time.Date(2026, 8, 9, 9, 0, 0, 0, time.UTC)
	store := &stableLearningStore{learningStore: &learningStore{}, recommendations: map[string]domain.ContinueAction{}}
	service := learningservice.NewWithClock(store, func() time.Time { return now })
	first, _, _, action, _, err := service.Dashboard(7, domain.UserRoleBuyer)
	if err != nil || first.ID != 7 || action == nil || action.Type != "read_theory" {
		t.Fatalf("first recommendation=(%+v,%+v,%v)", first, action, err)
	}
	store.theoryRead = true
	_, _, _, refresh, _, err := service.Dashboard(7, domain.UserRoleBuyer)
	if err != nil || refresh == nil || refresh.Type != "read_theory" {
		t.Fatalf("same-day recommendation changed: %+v %v", refresh, err)
	}
	now = now.Add(24 * time.Hour)
	_, _, _, nextDay, _, err := service.Dashboard(7, domain.UserRoleBuyer)
	if err != nil || nextDay == nil || nextDay.Type != "take_quiz" {
		t.Fatalf("next-day recommendation=(%+v,%v), want quiz", nextDay, err)
	}
}

func TestHTTPDailyTaskIsStableByMoscowDateAndRole(t *testing.T) {
	now := time.Date(2026, 8, 9, 20, 59, 0, 0, time.UTC)
	store := &learningStore{daily: map[string]domain.DailyTask{}}
	learning := learningservice.NewWithClock(store, func() time.Time { return now })
	handler := router.New()
	handler.Register(router.V1, learninghttp.New(learning).Routes())
	get := func(role string) string {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard?role="+role, nil)
		request = request.WithContext(authservice.WithIdentity(request.Context(), authservice.Identity{UserID: 7}))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, request)
		if rec.Code != http.StatusOK {
			t.Fatalf("dashboard=%d %s", rec.Code, rec.Body.String())
		}
		return rec.Body.String()
	}
	first, refresh := get("buyer"), get("buyer")
	if first != refresh || !strings.Contains(first, `"date":"2026-08-09"`) || len(store.daily) != 1 {
		t.Fatalf("unstable task: first=%s refresh=%s assignments=%d", first, refresh, len(store.daily))
	}
	get("seller")
	if len(store.daily) != 1 {
		t.Fatalf("role query created another task: %d", len(store.daily))
	}
	now = now.Add(2 * time.Minute)
	next := get("buyer")
	if !strings.Contains(next, `"date":"2026-08-10"`) || len(store.daily) != 2 {
		t.Fatalf("midnight task=%s assignments=%d", next, len(store.daily))
	}
}

func TestHTTPDailyTaskHidesVerdictUntilOneFinalAnswer(t *testing.T) {
	store := &learningStore{daily: map[string]domain.DailyTask{}}
	service := learningservice.NewWithClock(store, func() time.Time { return time.Date(2026, 8, 9, 9, 0, 0, 0, time.UTC) })
	handler := router.New()
	handler.Register(router.V1, learninghttp.New(service).Routes())
	dashboard := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard?role=buyer", nil)
	dashboard = dashboard.WithContext(authservice.WithIdentity(dashboard.Context(), authservice.Identity{UserID: 7}))
	before := httptest.NewRecorder()
	handler.ServeHTTP(before, dashboard)
	if before.Code != http.StatusOK || strings.Contains(before.Body.String(), `"verdict"`) || strings.Contains(before.Body.String(), `"safe_action"`) {
		t.Fatalf("incomplete task leaked feedback: %d %s", before.Code, before.Body.String())
	}
	answer := httptest.NewRequest(http.MethodPost, "/api/v1/daily-tasks/answer", strings.NewReader(`{"answer":false}`))
	answer = answer.WithContext(authservice.WithIdentity(answer.Context(), authservice.Identity{UserID: 7}))
	completed := httptest.NewRecorder()
	handler.ServeHTTP(completed, answer)
	if completed.Code != http.StatusOK || !strings.Contains(completed.Body.String(), `"verdict"`) || !strings.Contains(completed.Body.String(), `"safe_action"`) {
		t.Fatalf("completion = %d %s", completed.Code, completed.Body.String())
	}
	retryRequest := httptest.NewRequest(http.MethodPost, "/api/v1/daily-tasks/answer", strings.NewReader(`{"answer":false}`))
	retryRequest = retryRequest.WithContext(authservice.WithIdentity(retryRequest.Context(), authservice.Identity{UserID: 7}))
	repeat := httptest.NewRecorder()
	handler.ServeHTTP(repeat, retryRequest)
	if repeat.Code != http.StatusConflict || !strings.Contains(repeat.Body.String(), `"STATE_CONFLICT"`) {
		t.Fatalf("repeat = %d %s", repeat.Code, repeat.Body.String())
	}
}

func TestHTTPDashboardOffersFreePlayOnlyAfterAllSixTopicsAreCompleted(t *testing.T) {
	for count := 5; count <= 6; count++ {
		store := &learningStore{topics: make([]domain.Topic, count)}
		for i := range store.topics {
			store.topics[i] = domain.Topic{ID: i + 1, UserRole: domain.UserRoleBuyer, TheoryRead: true, QuizPassed: true, Completed: true}
		}
		handler := router.New()
		handler.Register(router.V1, learninghttp.New(learningservice.New(store)).Routes())
		request := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard?role=buyer", nil)
		request = request.WithContext(authservice.WithIdentity(request.Context(), authservice.Identity{UserID: 7}))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if count == 5 && recorder.Code != http.StatusOK {
			t.Fatalf("five topics = (%d,%s), want dashboard with no continue action", recorder.Code, recorder.Body.String())
		}
		if count == 6 && (recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"type":"start_free_play"`)) {
			t.Fatalf("six topics = (%d,%s), want free play", recorder.Code, recorder.Body.String())
		}
	}
}

type learningStore struct {
	theoryRead    bool
	activityCalls int
	attemptID     int
	lastRole      domain.UserRole
	daily         map[string]domain.DailyTask
	topics        []domain.Topic
}

type stableLearningStore struct {
	*learningStore
	recommendations map[string]domain.ContinueAction
}

func (s *stableLearningStore) FindRecommendation(_ int, date time.Time, role domain.UserRole) (domain.ContinueAction, bool, error) {
	action, ok := s.recommendations[date.Format("2006-01-02")+":"+string(role)]
	return action, ok, nil
}

func (s *stableLearningStore) SaveRecommendation(_ int, date time.Time, role domain.UserRole, action domain.ContinueAction) error {
	s.recommendations[date.Format("2006-01-02")+":"+string(role)] = action
	return nil
}

func (s *learningStore) FindDailyTask(_ int, date time.Time) (domain.DailyTask, bool, error) {
	task, ok := s.daily[date.Format("2006-01-02")]
	return task, ok, nil
}

func (s *learningStore) DailyTask(_ int, date time.Time, created domain.DailyTask) (domain.DailyTask, error) {
	if s.daily == nil {
		s.daily = map[string]domain.DailyTask{}
	}
	key := date.Format("2006-01-02")
	if task, ok := s.daily[key]; ok {
		return task, nil
	}
	task := created
	s.daily[key] = task
	return task, nil
}
func (s *learningStore) AnswerDailyTask(_ int, date time.Time, answer bool) (domain.DailyTask, domain.Streak, error) {
	key := date.Format("2006-01-02")
	task, ok := s.daily[key]
	if !ok {
		return domain.DailyTask{}, domain.Streak{}, learningservice.ErrDailyTaskUnavailable
	}
	if task.Completed {
		return domain.DailyTask{}, domain.Streak{}, learningservice.ErrDailyTaskAnswered
	}
	correct := answer == task.Verdict
	task.Answer, task.Correct, task.Completed = &answer, &correct, true
	now := time.Now()
	task.CompletedAt = &now
	s.daily[key] = task
	return task, domain.Streak{Current: 1, Longest: 1, ActiveToday: true}, nil
}

func (s *learningStore) Topics(_ int, role domain.UserRole) ([]domain.Topic, error) {
	s.lastRole = role
	if s.topics != nil {
		return s.topics, nil
	}
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
func (s *learningStore) RecentAttempts(int, domain.UserRole) ([]domain.RecentAttempt, float64, error) {
	return []domain.RecentAttempt{}, 0, nil
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
