package attempts_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"anti-scam-trainer/backend/internal/features/attempts/service"
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"
)

type fakeAI struct {
	evaluation service.EvaluatorResult
	generated  service.GeneratorResult
	err        error
}

func (a fakeAI) Evaluate(context.Context, service.EvaluationRequest) (service.EvaluatorResult, error) {
	if a.evaluation.Score == 0 {
		a.evaluation = service.EvaluatorResult{Score: 4, RiskType: "social_engineering", Evaluation: "Безопасный ответ", SafeAction: "Остаться в сервисе"}
	}
	return a.evaluation, a.err
}

func (a fakeAI) GenerateReply(_ context.Context, input service.GenerationRequest) (service.GeneratorResult, error) {
	if a.generated.Message == "" {
		a.generated = service.GeneratorResult{Message: "Продолжим", Tactic: input.AllowedTactics[0], Phase: input.Phase}
	}
	return a.generated, a.err
}

func TestLevelThreeFreeTextIsPersistedOnlyAfterValidAIResult(t *testing.T) {
	repo := newGameRepository()
	repo.progressByRole = map[string][]domain.Progress{"buyer": {{LevelID: 1, Stars: 1}, {LevelID: 2, Stars: 1}}}
	repo.steps = map[int]domain.ScenarioStep{1: {ID: 31, ScenarioID: 3, Number: 1, ResponseType: "mixed", MaxPoints: 100, AIInstruction: "Проверь отказ от внешней ссылки", FallbackMessage: "Оплатите доставку по ссылке"}}
	ai := fakeAI{evaluation: service.EvaluatorResult{Score: 4, IsSafe: true, RiskType: "social_engineering", Evaluation: "Безопасный отказ", SafeAction: "Остаться в сервисе", DetectedSignals: []string{"внешняя ссылка"}}}
	game := service.NewGameWithAI(repo, ai, ai)

	state, err := game.Start(1, 3, "buyer")
	if err != nil || state.Step.ResponseType != "mixed" || len(state.Messages) != 1 {
		t.Fatalf("Start(level 3) = (%#v, %v), want mixed state with opening message", state, err)
	}
	answer := "Я не перейду по ссылке и останусь в сервисе"
	_, completed, err := game.SubmitAnswer(context.Background(), 1, state.Attempt.ID, service.AnswerCommand{FreeText: &answer})
	if err != nil || completed == nil || completed.Attempt.Score != 100 {
		t.Fatalf("SubmitAnswer() = (%#v, %v), want completed score 100", completed, err)
	}
	if len(repo.answers) != 1 || repo.answers[0].FreeText != answer || len(repo.messages) != 3 {
		t.Fatalf("durable dialogue = answers %#v messages %#v", repo.answers, repo.messages)
	}
}

type recordingAI struct {
	evaluations []service.EvaluationRequest
	generations []service.GenerationRequest
	results     []service.EvaluatorResult
}

func (a *recordingAI) Evaluate(_ context.Context, input service.EvaluationRequest) (service.EvaluatorResult, error) {
	a.evaluations = append(a.evaluations, input)
	result := a.results[0]
	a.results = a.results[1:]
	return result, nil
}

func (a *recordingAI) GenerateReply(_ context.Context, input service.GenerationRequest) (service.GeneratorResult, error) {
	a.generations = append(a.generations, input)
	return service.GeneratorResult{Message: "Сгенерированная реплика", Tactic: input.AllowedTactics[0], Phase: input.Phase}, nil
}

func TestLevelThreeUsesEvaluatorForExactlyTwoFreeAnswers(t *testing.T) {
	repo := newGameRepository()
	repo.progressByRole = map[string][]domain.Progress{"buyer": {{LevelID: 1, Stars: 1}, {LevelID: 2, Stars: 1}}}
	repo.steps = map[int]domain.ScenarioStep{
		1: {ID: 31, ScenarioID: 3, Number: 1, ResponseType: domain.ResponseTypeMultipleChoice, CounterpartyMessage: "Курьер уже ждёт", Options: []domain.ScenarioOption{{ID: 311, Text: "Проверю оформление", Reaction: "Времени мало", Points: 100}}},
		2: {ID: 32, ScenarioID: 3, Number: 2, ResponseType: domain.ResponseTypeFreeText, CounterpartyMessage: "Внесите предоплату", AIInstruction: "Отказ от внешней формы", FallbackMessage: "Оформим внутри приложения"},
		3: {ID: 33, ScenarioID: 3, Number: 3, ResponseType: domain.ResponseTypeFreeText, CounterpartyMessage: "Введите данные карты", AIInstruction: "Отказ от карты и штрафа", FallbackMessage: "Диалог завершён"},
	}
	ai := &recordingAI{results: []service.EvaluatorResult{
		{Score: 2, RiskType: "phishing", Evaluation: "Сомнение сохраняет риск", SafeAction: "Не открывать форму"},
		{Score: 4, IsSafe: true, RiskType: "phishing", Evaluation: "Безопасный отказ", SafeAction: "Прекратить сделку"},
	}}
	game := service.NewGameWithAI(repo, ai, ai)
	state, err := game.Start(1, 3, "buyer")
	if err != nil {
		t.Fatal(err)
	}
	state, completed, err := game.Submit(1, state.Attempt.ID, 311, 31)
	if err != nil || completed != nil || state.Step.ID != 32 {
		t.Fatalf("scripted answer = (%#v,%#v,%v)", state, completed, err)
	}
	first := "Сначала посмотрю форму"
	state, completed, err = game.SubmitAnswer(context.Background(), 1, state.Attempt.ID, service.AnswerCommand{StepID: intPointer(32), FreeText: &first})
	if err != nil || completed != nil || state.Step.ID != 33 {
		t.Fatalf("first free answer = (%#v,%#v,%v)", state, completed, err)
	}
	second := "Ничего вводить не буду, прекращаю сделку"
	_, completed, err = game.SubmitAnswer(context.Background(), 1, state.Attempt.ID, service.AnswerCommand{StepID: intPointer(33), FreeText: &second})
	if err != nil || completed == nil || completed.Attempt.Score != 75 {
		t.Fatalf("second free answer = (%#v,%v), want score 75", completed, err)
	}
	if len(ai.evaluations) != 2 || len(ai.generations) != 0 || ai.evaluations[0].EvaluationContext != "Отказ от внешней формы" || ai.evaluations[0].ScenarioInstruction != "Верни JSON" || ai.evaluations[0].Rubric["safe_action"] != "Остаться в сервисе" || len(ai.evaluations[0].History) > 2 {
		t.Fatalf("AI calls = evaluations %#v generations %#v", ai.evaluations, ai.generations)
	}
}

func TestLevelFourUsesServerPhasesRollingHistoryAndFiveTurnLimit(t *testing.T) {
	repo := newGameRepository()
	repo.progressByRole = map[string][]domain.Progress{"buyer": {{LevelID: 1, Stars: 1}, {LevelID: 2, Stars: 1}, {LevelID: 3, Stars: 1}}}
	repo.steps = map[int]domain.ScenarioStep{}
	for number := 1; number <= 4; number++ {
		repo.steps[number] = domain.ScenarioStep{ID: 40 + number, ScenarioID: 4, Number: number, ResponseType: domain.ResponseTypeFreeText, CounterpartyMessage: "Реплика", AIInstruction: "Оценить безопасность", FallbackMessage: "Продолжим"}
	}
	results := make([]service.EvaluatorResult, 5)
	for index := range results {
		results[index] = service.EvaluatorResult{Score: 3, RiskType: "phishing", Evaluation: "Осторожный ответ", SafeAction: "Остаться в сервисе"}
	}
	ai := &recordingAI{results: results}
	game := service.NewGameWithAI(repo, ai, ai)
	state, err := game.Start(1, 4, "buyer")
	if err != nil {
		t.Fatal(err)
	}
	wantPhases := []string{"escalation", "escalation", "critical_request", "critical_request", "resolution"}
	for turn := 1; turn <= 5; turn++ {
		text := "Проверяю сделку только внутри приложения"
		next, completed, submitErr := game.SubmitAnswer(context.Background(), 1, state.Attempt.ID, service.AnswerCommand{FreeText: &text})
		if submitErr != nil {
			t.Fatalf("turn %d: %v", turn, submitErr)
		}
		if (turn < 5) != (completed == nil) {
			t.Fatalf("turn %d completion=%#v", turn, completed)
		}
		if turn < 5 {
			state = next
		}
	}
	if len(ai.evaluations) != 5 || len(ai.generations) != 5 {
		t.Fatalf("AI calls=(evaluations=%d generations=%d)", len(ai.evaluations), len(ai.generations))
	}
	for index, generation := range ai.generations {
		if generation.Phase != wantPhases[index] || generation.ScenarioInstruction != "Верни JSON" || generation.Rubric["safe_action"] != "Остаться в сервисе" || len(generation.History) > 6 {
			t.Fatalf("generation %d=%#v", index+1, generation)
		}
	}
	if ai.generations[4].Summary == "" || repo.attempts[state.Attempt.ID].CompactSummary == "" || repo.attempts[state.Attempt.ID].DialoguePhase != "resolution" {
		t.Fatalf("persisted dialogue state=%#v, fifth request=%#v", repo.attempts[state.Attempt.ID], ai.generations[4])
	}
}

func intPointer(value int) *int { return &value }

func TestAIFailureLeavesAnswerAndDialogueUnchanged(t *testing.T) {
	cases := []struct {
		name string
		ai   fakeAI
		want error
	}{
		{name: "timeout", ai: fakeAI{err: service.ErrAIUnavailable}, want: service.ErrAIUnavailable},
		{name: "invalid response", ai: fakeAI{err: service.ErrAIInvalidResponse}, want: service.ErrAIInvalidResponse},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			repo := newGameRepository()
			repo.progressByRole = map[string][]domain.Progress{"buyer": {{LevelID: 1, Stars: 1}, {LevelID: 2, Stars: 1}}}
			repo.steps = map[int]domain.ScenarioStep{1: {ID: 31, ScenarioID: 3, Number: 1, ResponseType: "mixed", MaxPoints: 100, FallbackMessage: "Начальная реплика"}}
			game := service.NewGameWithAI(repo, test.ai, test.ai)
			state, err := game.Start(1, 3, "buyer")
			if err != nil {
				t.Fatal(err)
			}
			beforeMessages := len(repo.messages)
			answer := "Безопасный ответ"
			_, _, err = game.SubmitAnswer(context.Background(), 1, state.Attempt.ID, service.AnswerCommand{FreeText: &answer})
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
			if len(repo.answers) != 0 || len(repo.messages) != beforeMessages || repo.attempts[state.Attempt.ID].CurrentStepNumber != 1 {
				t.Fatalf("AI failure changed state: %#v", repo)
			}
		})
	}
}

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

func TestFreePlayStartsBeforeTrainingIsCompleted(t *testing.T) {
	repo := newGameRepository()
	ai := fakeAI{generated: service.GeneratorResult{Message: "Первая реплика", Tactic: "rapport", Phase: "hook"}}
	game := service.NewGameWithDependencies(repo, ai, ai, func() bool { return true })

	state, err := game.StartFreePlay(context.Background(), 1, "buyer")
	if err != nil || state.Attempt.Mode != domain.AttemptModeFreePlay || len(state.Messages) != 1 {
		t.Fatalf("StartFreePlay() = (%#v, %v), want a free-play attempt with opening message", state, err)
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
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

func TestGameStartRejectsClosedSecondLevel(t *testing.T) {
	repo := newGameRepository()
	game := service.NewGame(repo)
	_, err := game.Start(1, 2, "buyer")
	if !errors.Is(err, apperrors.ErrForbidden) {
		t.Fatalf("Start() error = %v, want forbidden", err)
	}
}

func TestGameCompletesOnlyAfterLastAnswer(t *testing.T) {
	repo := newGameRepository()
	game := service.NewGame(repo)
	state, err := game.Start(1, 1, "buyer")
	if err != nil {
		t.Fatal(err)
	}
	next, finished, err := game.Submit(1, state.Attempt.ID, 11)
	if err != nil || finished != nil || next.Step.Number != 2 || len(next.Messages) != 3 {
		t.Fatalf("first answer = (%#v,%#v,%v), want next step", next, finished, err)
	}
	_, finished, err = game.Submit(1, state.Attempt.ID, 21)
	if err != nil || finished == nil || finished.Attempt.Score != 100 || finished.Stars != 3 {
		t.Fatalf("final answer = (%#v,%v), want completed 100/3", finished, err)
	}
	if repo.progress.Stars != 3 || repo.progress.UserRole != "buyer" {
		t.Fatalf("progress=%#v, want buyer three stars", repo.progress)
	}
}

func TestOptionReactionPrecedesNextScriptedMessage(t *testing.T) {
	repo := newGameRepository()
	repo.steps[1].Options[0].Reaction = "Поторопитесь, предложение скоро исчезнет"
	game := service.NewGame(repo)
	state, err := game.Start(1, 1, "buyer")
	if err != nil {
		t.Fatal(err)
	}

	next, completed, err := game.Submit(1, state.Attempt.ID, 11)
	if err != nil || completed != nil {
		t.Fatalf("Submit() = (%#v, %#v, %v)", next, completed, err)
	}
	want := []string{"Первая реплика", "", "Поторопитесь, предложение скоро исчезнет", "Вторая реплика"}
	if len(next.Messages) != len(want) {
		t.Fatalf("messages = %#v, want %d ordered messages", next.Messages, len(want))
	}
	want[1] = next.Step.Options[0].Text
	for index, text := range want {
		if next.Messages[index].Text != text {
			t.Fatalf("message %d = %q, want %q", index, next.Messages[index].Text, text)
		}
	}
}

func TestEmptyOptionReactionDoesNotCreateMessage(t *testing.T) {
	repo := newGameRepository()
	game := service.NewGame(repo)
	state, err := game.Start(1, 1, "buyer")
	if err != nil {
		t.Fatal(err)
	}
	next, _, err := game.Submit(1, state.Attempt.ID, 11)
	if err != nil {
		t.Fatal(err)
	}
	if len(next.Messages) != 3 {
		t.Fatalf("messages = %#v, want opening, answer and next scripted message", next.Messages)
	}
}

func TestFinalOptionReactionIsPersisted(t *testing.T) {
	repo := newGameRepository()
	step := repo.steps[2]
	step.Options[0].Reaction = "Последняя реакция собеседника"
	repo.steps[2] = step
	game := service.NewGame(repo)
	state, err := game.Start(1, 1, "buyer")
	if err != nil {
		t.Fatal(err)
	}
	state, _, err = game.Submit(1, state.Attempt.ID, 11)
	if err != nil {
		t.Fatal(err)
	}
	_, completed, err := game.Submit(1, state.Attempt.ID, 21)
	if err != nil || completed == nil {
		t.Fatalf("final Submit()=(%#v,%v)", completed, err)
	}
	if got := repo.messages[len(repo.messages)-1].Text; got != "Последняя реакция собеседника" {
		t.Fatalf("last message=%q", got)
	}
}

func TestFailedGameDoesNotSetPassedAt(t *testing.T) {
	repo := newGameRepository()
	repo.steps[1].Options[0].Points = 0
	repo.steps[2].Options[0].Points = 0
	game := service.NewGame(repo)
	state, err := game.Start(1, 1, "buyer")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = game.Submit(1, state.Attempt.ID, 11); err != nil {
		t.Fatal(err)
	}
	if _, completed, err := game.Submit(1, state.Attempt.ID, 21); err != nil || completed == nil {
		t.Fatalf("completion = (%#v,%v)", completed, err)
	}
	if repo.progress.Stars != 0 || !repo.progress.PassedAt.IsZero() {
		t.Fatalf("failed progress = %#v, want zero stars and no passed_at", repo.progress)
	}
}

func TestGameStartResumesOwnedAttemptAndRejectsForeignAnswer(t *testing.T) {
	repo := newGameRepository()
	game := service.NewGame(repo)
	started, err := game.Start(1, 1, "buyer")
	if err != nil {
		t.Fatal(err)
	}
	resumed, err := game.Start(1, 1, "buyer")
	if err != nil || resumed.Attempt.ID != started.Attempt.ID {
		t.Fatalf("resume = (%#v, %v), want existing attempt %d", resumed, err, started.Attempt.ID)
	}
	_, _, err = game.Submit(2, started.Attempt.ID, 11)
	if !errors.Is(err, apperrors.ErrAttemptNotFound) {
		t.Fatalf("foreign Submit() error = %v, want attempt not found", err)
	}
}

func TestGameOpensRoleBranchesIndependently(t *testing.T) {
	repo := newGameRepository()
	repo.progressByRole = map[string][]domain.Progress{
		"buyer":  {{UserID: 1, LevelID: 1, UserRole: "buyer", Stars: 1}},
		"seller": {},
	}
	game := service.NewGame(repo)

	buyerLevels, err := game.Levels(1, "buyer")
	if err != nil || !buyerLevels[1].Opened {
		t.Fatalf("buyer levels = %#v, %v; want level 2 open", buyerLevels, err)
	}
	sellerLevels, err := game.Levels(1, "seller")
	if err != nil || sellerLevels[1].Opened {
		t.Fatalf("seller levels = %#v, %v; want level 2 closed", sellerLevels, err)
	}
}

type gameRepository struct {
	attempts            map[int]domain.Attempt
	steps               map[int]domain.ScenarioStep
	answers             []domain.UserAnswer
	messages            []domain.DialogueMessage
	progress            domain.Progress
	progressByRole      map[string][]domain.Progress
	next                int
	failCompleteAttempt bool
	failStartFreePlay   bool
}

func newGameRepository() *gameRepository {
	return &gameRepository{attempts: map[int]domain.Attempt{}, next: 1, steps: map[int]domain.ScenarioStep{1: {ID: 1, ScenarioID: 1, Number: 1, MaxPoints: 100, FallbackMessage: "Первая реплика", Options: []domain.ScenarioOption{{ID: 11, Points: 100}}}, 2: {ID: 2, ScenarioID: 1, Number: 2, MaxPoints: 100, FallbackMessage: "Вторая реплика", Options: []domain.ScenarioOption{{ID: 21, Points: 100}}}}}
}
func (r *gameRepository) Levels(_ int, role string) ([]domain.Level, []domain.Progress, error) {
	return []domain.Level{{ID: 1, Number: 1}, {ID: 2, Number: 2}, {ID: 3, Number: 3}, {ID: 4, Number: 4}}, r.progressByRole[role], nil
}
func (r *gameRepository) PublishedScenario(level int, role string) (domain.Scenario, error) {
	if (role == "buyer" || role == "seller") && level >= 1 && level <= 4 {
		id := level
		if role == "seller" {
			id += 2
		}
		return domain.Scenario{ID: id, LevelID: level, UserRole: role}, nil
	}
	return domain.Scenario{}, errors.New("missing")
}
func (r *gameRepository) FreePlayConfig(role string) (domain.FreePlayConfig, error) {
	return domain.FreePlayConfig{UserRole: role, ProductContext: domain.ProductContext{ItemTitle: "Товар", Category: "Другое", DealMethod: "delivery"}, SystemPrompt: "Веди диалог", FinalRubric: domain.JSONObject{"safe": 100}}, nil
}
func (r *gameRepository) Scenario(id int) (domain.Scenario, error) {
	return domain.Scenario{ID: id, Level: strconv.Itoa(id), LevelID: id, UserRole: "buyer", ScamScheme: "phishing", AISystemPrompt: "Верни JSON", FinalRubric: domain.JSONObject{"safe_action": "Остаться в сервисе"}}, nil
}
func (r *gameRepository) FindInProgress(user, scenario int) (domain.Attempt, error) {
	for _, a := range r.attempts {
		if a.UserID == user && a.ScenarioID == scenario && a.Status == domain.AttemptStatusInProgress {
			return a, nil
		}
	}
	return domain.Attempt{}, errors.New("missing")
}
func (r *gameRepository) FindInProgressFreePlay(user int, role string) (domain.Attempt, error) {
	for _, a := range r.attempts {
		if a.UserID == user && a.Mode == domain.AttemptModeFreePlay && a.UserRole == role && a.Status == domain.AttemptStatusInProgress {
			return a, nil
		}
	}
	return domain.Attempt{}, errors.New("missing")
}
func (r *gameRepository) CreateGameAttempt(a domain.Attempt) (domain.Attempt, error) {
	a.ID = r.next
	r.next++
	r.attempts[a.ID] = a
	return a, nil
}
func (r *gameRepository) StartFreePlay(a domain.Attempt, message domain.DialogueMessage) (domain.Attempt, error) {
	if r.failStartFreePlay {
		return domain.Attempt{}, errors.New("opening message write failed")
	}
	created, err := r.CreateGameAttempt(a)
	if err != nil {
		return domain.Attempt{}, err
	}
	message.AttemptID = created.ID
	r.messages = append(r.messages, message)
	return created, nil
}
func (r *gameRepository) GetGameAttempt(id int) (domain.Attempt, error) {
	a, ok := r.attempts[id]
	if !ok {
		return domain.Attempt{}, errors.New("missing")
	}
	return a, nil
}
func (r *gameRepository) Step(scenarioID, n int) (domain.ScenarioStep, error) {
	if scenarioID == 0 {
		return domain.ScenarioStep{}, errors.New("free play has no scenario steps")
	}
	v, ok := r.steps[n]
	if !ok {
		return domain.ScenarioStep{}, errors.New("missing")
	}
	return v, nil
}
func (r *gameRepository) Answers(id int) ([]domain.UserAnswer, error) {
	var out []domain.UserAnswer
	for _, a := range r.answers {
		if a.AttemptID == id {
			out = append(out, a)
		}
	}
	return out, nil
}
func (r *gameRepository) Messages(id int) ([]domain.DialogueMessage, error) {
	var out []domain.DialogueMessage
	for _, message := range r.messages {
		if message.AttemptID == id {
			out = append(out, message)
		}
	}
	return out, nil
}
func (r *gameRepository) AwardedPoints(int) (int, error) {
	total := 0
	for _, a := range r.answers {
		total += a.AwardedPoints
	}
	return total, nil
}
func (r *gameRepository) Advance(id, next int) error {
	a := r.attempts[id]
	a.CurrentStepNumber = next
	r.attempts[id] = a
	return nil
}
func (r *gameRepository) Abandon(id int, _ time.Time) error {
	a := r.attempts[id]
	a.Status = domain.AttemptStatusAbandoned
	r.attempts[id] = a
	return nil
}
func (r *gameRepository) Complete(action func(service.GameCompletionStore) error) error {
	clone := *r
	clone.attempts = make(map[int]domain.Attempt, len(r.attempts))
	for id, attempt := range r.attempts {
		clone.attempts[id] = attempt
	}
	clone.answers = append([]domain.UserAnswer(nil), r.answers...)
	clone.messages = append([]domain.DialogueMessage(nil), r.messages...)
	if err := action(&clone); err != nil {
		return err
	}
	*r = clone
	return nil
}
func (r *gameRepository) SaveAnswer(a domain.UserAnswer) error {
	r.answers = append(r.answers, a)
	return nil
}
func (r *gameRepository) SaveMessage(message domain.DialogueMessage) error {
	r.messages = append(r.messages, message)
	return nil
}
func (r *gameRepository) UpdateDialogueState(id, count int, phase, summary string) error {
	a := r.attempts[id]
	a.FreeTextCount = count
	a.DialoguePhase = phase
	a.CompactSummary = summary
	r.attempts[id] = a
	return nil
}
func (r *gameRepository) AdvanceAttempt(id, n int) error { return r.Advance(id, n) }
func (r *gameRepository) CompleteAttempt(a domain.Attempt) error {
	if r.failCompleteAttempt {
		return errors.New("completion write failed")
	}
	r.attempts[a.ID] = a
	return nil
}
func (r *gameRepository) SaveProgress(p domain.Progress) error         { r.progress = p; return nil }
func (r *gameRepository) FinalizeLearning(*domain.AttemptResult) error { return nil }
