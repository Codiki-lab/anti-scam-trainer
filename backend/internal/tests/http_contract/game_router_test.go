package http_contract_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/ratelimit"
	"anti-scam-trainer/backend/internal/core/server/middleware"
	"anti-scam-trainer/backend/internal/core/server/router"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	attemptshttp "anti-scam-trainer/backend/internal/features/attempts/transport/http"
	authservice "anti-scam-trainer/backend/internal/features/auth/service"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

type contractAIProvider interface {
	attemptsservice.Evaluator
	attemptsservice.ScammerGenerator
}

func TestHTTPGameContractExposesMixedStateAndRejectsAmbiguousAnswer(t *testing.T) {
	store := newHTTPGameStore()
	game := attemptsservice.NewGameWithAI(store, contractAI{}, contractAI{})
	handler := router.New()
	handler.Register(router.V1, attemptshttp.New(game).Routes())

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
		ai     contractAIProvider
		status int
	}{
		{name: "timeout", ai: failingContractAI{}, status: http.StatusServiceUnavailable},
		{name: "invalid JSON", ai: invalidContractAI{}, status: http.StatusBadGateway},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			store := newHTTPGameStore()
			game := attemptsservice.NewGameWithAI(store, test.ai, test.ai)
			handler := router.New()
			handler.Register(router.V1, attemptshttp.New(game).Routes())
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

func TestHTTPAIRateLimitRejectsBeforeProviderAndStateMutation(t *testing.T) {
	store := newHTTPGameStore()
	provider := &countingContractAI{}
	now := time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)
	limit := ratelimit.New(ratelimit.Config{Limit: 1, Window: time.Minute, MaxBuckets: 10, IdleTTL: time.Minute}, func() time.Time { return now })
	freePlayLimit := ratelimit.New(ratelimit.Config{Limit: 10, Window: time.Minute, MaxBuckets: 10, IdleTTL: time.Minute}, func() time.Time { return now })
	game := attemptsservice.NewGameWithRateLimits(store, provider, provider, limit, freePlayLimit, ratelimit.NewGate())
	r := router.New()
	r.Register(router.V1, attemptshttp.New(game).Routes())
	handler := middleware.RequestID()(r)
	start := httptest.NewRequest(http.MethodPost, "/api/v1/training/free-play/start?role=buyer", nil)
	start = start.WithContext(authservice.WithIdentity(start.Context(), authservice.Identity{UserID: 1}))
	handler.ServeHTTP(httptest.NewRecorder(), start)
	answer := func(step int) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/attempts/1/answers", strings.NewReader(`{"step_id":`+strconv.Itoa(step)+`,"free_text":"Останусь в сервисе"}`))
		req = req.WithContext(authservice.WithIdentity(req.Context(), authservice.Identity{UserID: 1}))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}
	if first := answer(1); first.Code != http.StatusOK {
		t.Fatalf("first=%d %s", first.Code, first.Body.String())
	}
	answersBefore := len(store.answers)
	messagesBefore := len(store.messages)
	second := answer(2)
	if second.Code != http.StatusTooManyRequests || !strings.Contains(second.Body.String(), `"code":"RATE_LIMITED"`) || second.Header().Get("Retry-After") == "" || provider.calls != 3 || len(store.answers) != answersBefore || len(store.messages) != messagesBefore {
		t.Fatalf("limited=(%d,%s,calls=%d,answers=%d)", second.Code, second.Body.String(), provider.calls, len(store.answers))
	}
}

func TestHTTPMicroQuestionAnswerIsOptionalAndDoesNotChangeResult(t *testing.T) {
	store := newHTTPGameStore()
	store.attempts[1] = domain.Attempt{ID: 1, UserID: 1, Status: domain.AttemptStatusCompleted}
	store.result = domain.AttemptResult{AttemptID: 1, Score: 75, Stars: 2, MicroQuestion: &domain.MicroQuestion{PatternCode: "phishing", Question: "Безопасное действие?", Options: []string{"Проверить", "Выполнить"}, Correct: 0}}
	game := attemptsservice.NewGame(store)
	handler := router.New()
	handler.Register(router.V1, attemptshttp.New(game).Routes())
	call := func(body string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/attempts/1/micro-question/answer", strings.NewReader(body))
		request = request.WithContext(authservice.WithIdentity(request.Context(), authservice.Identity{UserID: 1}))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder
	}
	if missing := call(`{}`); missing.Code != http.StatusConflict {
		t.Fatalf("missing answer = (%d,%s)", missing.Code, missing.Body.String())
	}
	before := store.result
	answered := call(`{"answer_index":0}`)
	if answered.Code != http.StatusOK || !strings.Contains(answered.Body.String(), `"correct":true`) || !reflect.DeepEqual(store.result, before) {
		t.Fatalf("answer = (%d,%s), result=%#v", answered.Code, answered.Body.String(), store.result)
	}
}

func TestHTTPCompletedAttemptsDrivePatternThresholdAndSafeDecay(t *testing.T) {
	store := newHTTPGameStore()
	complete := func(safe bool) (*router.Router, string) {
		game := attemptsservice.NewGameWithDependencies(store, profileContractAI{safe: safe}, profileContractAI{safe: safe}, func() bool { return true })
		handler := router.New()
		handler.Register(router.V1, attemptshttp.New(game).Routes())
		start := httptest.NewRequest(http.MethodPost, "/api/v1/training/free-play/start?role=buyer", nil)
		start = start.WithContext(authservice.WithIdentity(start.Context(), authservice.Identity{UserID: 1}))
		startRecorder := httptest.NewRecorder()
		handler.ServeHTTP(startRecorder, start)
		if startRecorder.Code != http.StatusOK {
			t.Fatalf("start = (%d,%s)", startRecorder.Code, startRecorder.Body.String())
		}
		body := ""
		for turn := 1; turn <= 3; turn++ {
			finish := ""
			if turn == 3 {
				finish = `,"finish":true`
			}
			request := httptest.NewRequest(http.MethodPost, "/api/v1/attempts/1/answers", strings.NewReader(`{"step_id":`+strconv.Itoa(turn)+`,"free_text":"ответ"`+finish+`}`))
			request = request.WithContext(authservice.WithIdentity(request.Context(), authservice.Identity{UserID: 1}))
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("turn %d = (%d,%s)", turn, recorder.Code, recorder.Body.String())
			}
			body = recorder.Body.String()
		}
		return handler, body
	}

	_, first := complete(false)
	if strings.Contains(first, `"micro_question"`) {
		t.Fatalf("one risky completion must not create a stable pattern: %s", first)
	}
	secondHandler, second := complete(false)
	if !strings.Contains(second, `"micro_question"`) || !strings.Contains(second, "social_engineering") {
		t.Fatalf("second risky completion must expose the pattern question: %s", second)
	}
	beforeResult := store.result
	beforeAttempt := store.attempts[1]
	beforeAnswers, beforeMessages := len(store.answers), len(store.messages)
	beforeQuiz, beforeStreak, beforeLevels := store.quizBest, store.streak, append([]domain.Level(nil), store.levels...)
	answer := httptest.NewRequest(http.MethodPost, "/api/v1/attempts/1/micro-question/answer", strings.NewReader(`{"answer_index":0}`))
	answer = answer.WithContext(authservice.WithIdentity(answer.Context(), authservice.Identity{UserID: 1}))
	answerRecorder := httptest.NewRecorder()
	secondHandler.ServeHTTP(answerRecorder, answer)
	if answerRecorder.Code != http.StatusOK || !reflect.DeepEqual(store.result, beforeResult) || !reflect.DeepEqual(store.attempts[1], beforeAttempt) || len(store.answers) != beforeAnswers || len(store.messages) != beforeMessages || store.quizBest != beforeQuiz || !reflect.DeepEqual(store.streak, beforeStreak) || !reflect.DeepEqual(store.levels, beforeLevels) {
		t.Fatalf("micro answer changed progression: code=%d body=%s", answerRecorder.Code, answerRecorder.Body.String())
	}

	_, third := complete(true)
	if !strings.Contains(third, `"micro_question"`) {
		t.Fatalf("one safe completion must reduce but not yet clear priority: %s", third)
	}
	_, fourth := complete(true)
	if strings.Contains(fourth, `"micro_question"`) {
		t.Fatalf("two safe completions must clear the stable pattern: %s", fourth)
	}
}

func TestHTTPAIMetricsExposeOnlyAggregates(t *testing.T) {
	store := newHTTPGameStore()
	game := attemptsservice.NewGameWithAI(store, contractAI{}, contractAI{})
	handler := router.New()
	handler.Register(router.V1, attemptshttp.New(game).Routes())
	request := httptest.NewRequest(http.MethodGet, "/api/v1/ai/metrics", nil)
	request = request.WithContext(authservice.WithIdentity(request.Context(), authservice.Identity{UserID: 1}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"evaluator"`) || strings.Contains(recorder.Body.String(), "policy") {
		t.Fatalf("metrics = (%d,%s)", recorder.Code, recorder.Body.String())
	}
}

func TestHTTPBoundsSharedProviderConcurrencyAcrossUsers(t *testing.T) {
	store := newHTTPGameStore()
	provider := &blockingContractAI{started: make(chan struct{}), release: make(chan struct{})}
	limit := ratelimit.New(ratelimit.Config{Limit: 10, Window: time.Minute, MaxBuckets: 10, IdleTTL: time.Minute}, time.Now)
	game := attemptsservice.NewGameWithRateLimits(store, provider, provider, limit, limit, ratelimit.NewGate())
	r := router.New()
	r.Register(router.V1, attemptshttp.New(game).Routes())
	handler := middleware.RequestID()(r)
	start := func(userID int) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/training/free-play/start?role=buyer", nil)
		request = request.WithContext(authservice.WithIdentity(request.Context(), authservice.Identity{UserID: userID}))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, request)
		return rec
	}
	firstDone := make(chan *httptest.ResponseRecorder, 1)
	go func() { firstDone <- start(1) }()
	<-provider.started
	second := start(2)
	if second.Code != http.StatusTooManyRequests || !strings.Contains(second.Body.String(), `"code":"RATE_LIMITED"`) {
		t.Fatalf("concurrent=(%d,%s)", second.Code, second.Body.String())
	}
	close(provider.release)
	first := <-firstDone
	if first.Code != http.StatusOK || provider.calls.Load() != 1 {
		t.Fatalf("first=(%d,%s),calls=%d", first.Code, first.Body.String(), provider.calls.Load())
	}
}

func TestHTTPFreePlayCoversBothRolesAndHidesCounterpartUntilCompletion(t *testing.T) {
	for _, role := range []string{"buyer", "seller"} {
		t.Run(role, func(t *testing.T) {
			store := newHTTPGameStore()
			game := attemptsservice.NewGameWithDependencies(store, contractAI{}, contractAI{}, func() bool { return role == "buyer" })
			handler := router.New()
			handler.Register(router.V1, attemptshttp.New(game).Routes())
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
	game := attemptsservice.NewGameWithDependencies(store, ai, ai, func() bool { return true })
	handler := router.New()
	handler.Register(router.V1, attemptshttp.New(game).Routes())
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

func TestHTTPFreePlayCompletesAtFiveAnswersWithoutLeakingAIState(t *testing.T) {
	store := newHTTPGameStore()
	game := attemptsservice.NewGameWithDependencies(store, contractAI{}, contractAI{}, func() bool { return true })
	handler := router.New()
	handler.Register(router.V1, attemptshttp.New(game).Routes())
	start := httptest.NewRequest(http.MethodPost, "/api/v1/training/free-play/start?role=buyer", nil)
	start = start.WithContext(authservice.WithIdentity(start.Context(), authservice.Identity{UserID: 1}))
	handler.ServeHTTP(httptest.NewRecorder(), start)
	for turn := 1; turn <= 5; turn++ {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/attempts/1/answers", strings.NewReader(`{"step_id":`+strconv.Itoa(turn)+`,"free_text":"Проверю сделку внутри приложения"}`))
		request = request.WithContext(authservice.WithIdentity(request.Context(), authservice.Identity{UserID: 1}))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		body := recorder.Body.String()
		if recorder.Code != http.StatusOK || strings.Contains(body, "policy") || strings.Contains(body, "compact_summary") || strings.Contains(body, "dialogue_phase") {
			t.Fatalf("turn %d=(%d,%s)", turn, recorder.Code, body)
		}
		if turn == 5 && !strings.Contains(body, `"decision_review"`) {
			t.Fatalf("fifth turn did not return Result: %s", body)
		}
	}
	sixth := httptest.NewRequest(http.MethodPost, "/api/v1/attempts/1/answers", strings.NewReader(`{"step_id":6,"free_text":"Ещё ответ"}`))
	sixth = sixth.WithContext(authservice.WithIdentity(sixth.Context(), authservice.Identity{UserID: 1}))
	sixthRecorder := httptest.NewRecorder()
	handler.ServeHTTP(sixthRecorder, sixth)
	if sixthRecorder.Code != http.StatusConflict || len(store.answers) != 5 {
		t.Fatalf("sixth=(%d,%s), answers=%d", sixthRecorder.Code, sixthRecorder.Body.String(), len(store.answers))
	}
}
