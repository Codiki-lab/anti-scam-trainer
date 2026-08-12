package attempts_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/features/attempts/service"
	"context"
	"strings"
	"testing"
)

func TestFreePlayKeepsCounterpartTypeHiddenFromStateAndCompletesOnThirdRequestedAnswer(t *testing.T) {
	repo := newGameRepository()
	repo.progressByRole = map[string][]domain.Progress{"seller": {{LevelID: 4, Stars: 1}}}
	evaluation := service.EvaluatorResult{Score: 3, RiskType: "ordinary_transaction", Evaluation: "Осторожная стратегия", SafeAction: "Остаться в сервисе", DetectedSignals: []string{"давление"}}
	ai := &recordingAI{results: []service.EvaluatorResult{evaluation, evaluation, evaluation}}
	game := service.NewGameWithDependencies(repo, ai, ai, func() bool { return false })
	state, err := game.StartFreePlay(context.Background(), 1, "seller")
	if err != nil || state.Attempt.IsScam == nil || *state.Attempt.IsScam || len(state.Messages) != 1 {
		t.Fatalf("StartFreePlay() = (%#v, %v), want hidden honest counterpart and first message", state, err)
	}
	if state.Scenario.ProductContext.ItemTitle == "" || !strings.HasPrefix(state.Attempt.CompactSummary, "[free-play-context:") {
		t.Fatalf("free-play context was not selected and pinned: %#v", state)
	}
	for n := 1; n <= 2; n++ {
		text := "Продолжим безопасно в чате сервиса"
		next, completed, submitErr := game.SubmitAnswer(context.Background(), 1, state.Attempt.ID, service.AnswerCommand{FreeText: &text})
		if submitErr != nil || completed != nil || next.Attempt.FreeTextCount != n {
			t.Fatalf("turn %d = (%#v, %#v, %v), want continuation", n, next, completed, submitErr)
		}
		if len(repo.answers) != n || repo.answers[n-1].StepID != 0 {
			t.Fatalf("turn %d stored answer = %#v, want no scenario step", n, repo.answers)
		}
	}
	text := "Завершаю сделку только штатным способом"
	_, completed, err := game.SubmitAnswer(context.Background(), 1, state.Attempt.ID, service.AnswerCommand{FreeText: &text, Finish: true})
	if err != nil || completed == nil || completed.Attempt.Score != 75 || repo.progress.Stars != 0 {
		t.Fatalf("free play completion = (%#v, %v), want score without level progress", completed, err)
	}
	if len(completed.Result.RiskSignals) != 1 || completed.Result.RiskSignals[0].Code != "pressure" || completed.Result.RiskSignals[0].Label == "" {
		t.Fatalf("result risk signals=%v", completed.Result.RiskSignals)
	}
	for index, item := range completed.Result.DecisionReview {
		if item.StepID != index+1 {
			t.Fatalf("decision step %d has id %d", index+1, item.StepID)
		}
	}
	if len(ai.evaluations) != 3 || len(ai.generations) != 4 {
		t.Fatalf("honest counterpart AI calls=(eval=%d, gen=%d)", len(ai.evaluations), len(ai.generations))
	}
	for _, request := range ai.generations {
		if request.CounterpartKind != "обычный участник сделки" || request.RiskType != "ordinary_transaction" || containsString(request.AllowedTactics, "payment") || containsString(request.AllowedTactics, "credential_request") {
			t.Fatalf("honest counterpart received scam policy: %#v", request)
		}
	}
	nextState, nextErr := game.StartFreePlay(context.Background(), 1, "seller")
	if nextErr != nil || nextState.Scenario.ProductContext.ItemTitle == state.Scenario.ProductContext.ItemTitle {
		t.Fatalf("next free play repeated product context: first=%q next=%q err=%v", state.Scenario.ProductContext.ItemTitle, nextState.Scenario.ProductContext.ItemTitle, nextErr)
	}
}

func TestFreePlayRequiresCompletedRoleBranch(t *testing.T) {
	repo := newGameRepository()
	repo.freePlayUnlocked = false
	ai := fakeAI{generated: service.GeneratorResult{Message: "Первая реплика", Tactic: "rapport", Phase: "hook"}}
	game := service.NewGameWithDependencies(repo, ai, ai, func() bool { return true })

	_, err := game.StartFreePlay(context.Background(), 1, "buyer")
	if err != service.ErrFreePlayLocked {
		t.Fatalf("StartFreePlay() error = %v, want %v", err, service.ErrFreePlayLocked)
	}
}

func TestFreePlayDoesNotResumeWhenRoleBranchIsNoLongerCompleted(t *testing.T) {
	repo := newGameRepository()
	isScam := true
	repo.attempts[1] = domain.Attempt{ID: 1, UserID: 1, Mode: domain.AttemptModeFreePlay, UserRole: domain.UserRoleBuyer, IsScam: &isScam, Status: domain.AttemptStatusInProgress}
	repo.freePlayUnlocked = false
	game := service.NewGameWithDependencies(repo, fakeAI{}, fakeAI{}, func() bool { return true })

	_, err := game.StartFreePlay(context.Background(), 1, domain.UserRoleBuyer)
	if err != service.ErrFreePlayLocked {
		t.Fatalf("StartFreePlay() error = %v, want %v", err, service.ErrFreePlayLocked)
	}
}

func TestFreePlayCompletesAutomaticallyOnFifthAnswer(t *testing.T) {
	repo := newGameRepository()
	repo.progressByRole = map[string][]domain.Progress{"buyer": {{LevelID: 4, Stars: 1}}}
	ai := fakeAI{evaluation: service.EvaluatorResult{Score: 3, RiskType: "social_engineering", Evaluation: "Безопасно", SafeAction: "Остаться в сервисе"}}
	game := service.NewGameWithDependencies(repo, ai, ai, func() bool { return true })
	state, err := game.StartFreePlay(context.Background(), 1, "buyer")
	if err != nil {
		t.Fatal(err)
	}
	for turn := 1; turn <= 5; turn++ {
		text := "Проверяю условия сделки в сервисе"
		_, completed, submitErr := game.SubmitAnswer(context.Background(), 1, state.Attempt.ID, service.AnswerCommand{FreeText: &text})
		if submitErr != nil {
			t.Fatalf("turn %d: %v", turn, submitErr)
		}
		if (turn < 5) != (completed == nil) {
			t.Fatalf("turn %d completion = %#v", turn, completed)
		}
	}
}

func TestFinalFreeTextWriteRollsBackAtomically(t *testing.T) {
	repo := newGameRepository()
	repo.progressByRole = map[string][]domain.Progress{"seller": {{LevelID: 4, Stars: 1}}}
	ai := fakeAI{}
	game := service.NewGameWithDependencies(repo, ai, ai, func() bool { return true })
	state, err := game.StartFreePlay(context.Background(), 1, "seller")
	if err != nil {
		t.Fatal(err)
	}
	for turn := 1; turn <= 2; turn++ {
		text := "Безопасный ответ"
		if _, _, err = game.SubmitAnswer(context.Background(), 1, state.Attempt.ID, service.AnswerCommand{FreeText: &text}); err != nil {
			t.Fatal(err)
		}
	}
	repo.failCompleteAttempt = true
	text := "Завершаю безопасно"
	_, _, err = game.SubmitAnswer(context.Background(), 1, state.Attempt.ID, service.AnswerCommand{FreeText: &text, Finish: true})
	if err == nil {
		t.Fatal("completion succeeded, want storage failure")
	}
	if len(repo.answers) != 2 || repo.attempts[state.Attempt.ID].Status != domain.AttemptStatusInProgress || repo.progress.Stars != 0 {
		t.Fatalf("partial completion persisted: %#v", repo)
	}
}

func TestFreePlayStartDoesNotLeaveAttemptWithoutOpeningMessage(t *testing.T) {
	repo := newGameRepository()
	repo.progressByRole = map[string][]domain.Progress{"buyer": {{LevelID: 4, Stars: 1}}}
	repo.failStartFreePlay = true
	ai := fakeAI{generated: service.GeneratorResult{Message: "Первая реплика", Tactic: "rapport", Phase: "hook"}}
	game := service.NewGameWithDependencies(repo, ai, ai, func() bool { return true })
	if _, err := game.StartFreePlay(context.Background(), 1, "buyer"); err == nil {
		t.Fatal("start succeeded, want storage failure")
	}
	if len(repo.attempts) != 0 || len(repo.messages) != 0 {
		t.Fatalf("partial free play start persisted: %#v", repo)
	}
}
