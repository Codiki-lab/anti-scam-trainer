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

func TestHTTPSkillCheckPersistsPairAndNeverTouchesProgression(t *testing.T) {
	store := &learningStore{}
	handler := router.New()
	handler.Register(router.V1, learninghttp.New(learningservice.New(store)).Routes())
	call := func(method, path, body string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		request = request.WithContext(authservice.WithIdentity(request.Context(), authservice.Identity{UserID: 7}))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder
	}
	beforeTheory, beforeActivity := store.theoryRead, store.activityCalls
	start := call(http.MethodPost, "/api/v1/topics/1/skill-check/start", "")
	if start.Code != http.StatusOK || !strings.Contains(start.Body.String(), `"phase":"before"`) || !strings.Contains(start.Body.String(), `"snapshot"`) {
		t.Fatalf("start = (%d,%s)", start.Code, start.Body.String())
	}
	pre := call(http.MethodPost, "/api/v1/skill-checks/9/answers", `{"answer":false}`)
	if pre.Code != http.StatusOK || !strings.Contains(pre.Body.String(), `"phase":"after_locked"`) || strings.Contains(pre.Body.String(), `"snapshot"`) {
		t.Fatalf("pre = (%d,%s)", pre.Code, pre.Body.String())
	}
	store.skillCheck.TopicComplete = true
	resume := call(http.MethodPost, "/api/v1/topics/1/skill-check/start", "")
	if resume.Code != http.StatusOK || !strings.Contains(resume.Body.String(), `"phase":"after"`) || !strings.Contains(resume.Body.String(), "Отправьте код возврата") {
		t.Fatalf("resume = (%d,%s)", resume.Code, resume.Body.String())
	}
	post := call(http.MethodPost, "/api/v1/skill-checks/9/answers", `{"answer":true}`)
	if post.Code != http.StatusOK || !strings.Contains(post.Body.String(), `"phase":"completed"`) || !strings.Contains(post.Body.String(), `"verdict_improved":true`) || !strings.Contains(post.Body.String(), `"pattern_improved":true`) || !strings.Contains(post.Body.String(), `"improved":true`) {
		t.Fatalf("post = (%d,%s)", post.Code, post.Body.String())
	}
	if store.theoryRead != beforeTheory || store.activityCalls != beforeActivity {
		t.Fatalf("skill check changed progression: theory=%v activity=%d", store.theoryRead, store.activityCalls)
	}
}

func TestHTTPPersonalRecommendationUsesStablePatternWithoutUnlockingLevel(t *testing.T) {
	store := &learningStore{stablePattern: "external_link", topics: []domain.Topic{{ID: 1, Slug: "buyer-phishing-links", UserRole: domain.UserRoleBuyer, Status: domain.TopicStatusPublished, TheoryRead: true, QuizPassed: true, Levels: []domain.TopicLevelProgress{{Number: 1, Opened: true, Stars: 1}, {Number: 2, Opened: true}, {Number: 3, Opened: false}, {Number: 4, Opened: false}}}}}
	handler := router.New()
	handler.Register(router.V1, learninghttp.New(learningservice.New(store)).Routes())
	request := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/next?role=buyer", nil)
	request = request.WithContext(authservice.WithIdentity(request.Context(), authservice.Identity{UserID: 7}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"type":"start_level"`) || !strings.Contains(recorder.Body.String(), `"level":2`) || strings.Contains(recorder.Body.String(), `"level":3`) {
		t.Fatalf("recommendation = (%d,%s)", recorder.Code, recorder.Body.String())
	}
}
