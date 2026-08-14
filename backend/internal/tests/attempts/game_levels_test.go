package attempts_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	attemptsai "anti-scam-trainer/backend/internal/features/attempts/aiprovider"
	"anti-scam-trainer/backend/internal/features/attempts/service"
	"context"
	"errors"
	"testing"
)

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

func TestLevelFourUsesOneModelCallAndKeepsScenarioReply(t *testing.T) {
	repo := newGameRepository()
	repo.progressByRole = map[string][]domain.Progress{"buyer": {{LevelID: 1, Stars: 1}, {LevelID: 2, Stars: 1}, {LevelID: 3, Stars: 1}}}
	repo.steps = map[int]domain.ScenarioStep{1: {ID: 41, ScenarioID: 4, Number: 1, ResponseType: domain.ResponseTypeFreeText, CounterpartyMessage: "Телефон в наличии.", AIInstruction: "Оценить безопасность", FallbackMessage: "Откройте присланную форму и оплатите заказ банковской картой."}}
	provider := &sequenceProvider{contents: []string{`{"score":3,"is_safe":true,"risk_type":"phishing","detected_signals":[],"evaluation":"Ответ снижает риск","safe_action":"Проверить заказ в приложении"}`}}
	ai := service.NewModelAI(attemptsai.New(provider))
	game := service.NewGameWithAI(repo, ai, ai)
	state, err := game.Start(1, 4, "buyer")
	if err != nil {
		t.Fatal(err)
	}
	answer := "Я подумаю над условиями"
	next, completed, err := game.SubmitAnswer(context.Background(), 1, state.Attempt.ID, service.AnswerCommand{FreeText: &answer})
	if err != nil || completed != nil {
		t.Fatalf("SubmitAnswer() = (%#v, %#v, %v)", next, completed, err)
	}
	if len(provider.requests) != 1 {
		t.Fatalf("model requests=%d, want evaluator only", len(provider.requests))
	}
	if got := next.Messages[len(next.Messages)-1].Text; got != "Откройте присланную форму и оплатите заказ банковской картой." {
		t.Fatalf("counterpart reply=%q, want scenario reply", got)
	}
}

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
