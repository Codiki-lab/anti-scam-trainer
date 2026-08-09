package attempts_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"anti-scam-trainer/backend/internal/features/attempts/service"
	"context"
	"errors"
	"testing"
	"time"
)

type fakeAI struct {
	result string
	err    error
}

func (a fakeAI) Generate(context.Context, []service.AIMessage) (string, error) {
	return a.result, a.err
}

func TestLevelThreeFreeTextIsPersistedOnlyAfterValidAIResult(t *testing.T) {
	repo := newGameRepository()
	repo.progressByRole = map[string][]domain.Progress{"buyer": {{LevelID: 1, Stars: 1}, {LevelID: 2, Stars: 1}}}
	repo.steps = map[int]domain.ScenarioStep{1: {ID: 31, ScenarioID: 3, Number: 1, ResponseType: "mixed", MaxPoints: 100, AIInstruction: "Проверь отказ от внешней ссылки", FallbackMessage: "Оплатите доставку по ссылке"}}
	game := service.NewGameWithAI(repo, fakeAI{result: `{"awarded_points":100,"explanation":"Безопасный отказ","reply":"Почему вы не доверяете ссылке?","risk_signals":["внешняя ссылка"]}`})

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

func TestAIFailureLeavesAnswerAndDialogueUnchanged(t *testing.T) {
	cases := []struct {
		name string
		ai   fakeAI
		want error
	}{
		{name: "timeout", ai: fakeAI{err: service.ErrAIUnavailable}, want: service.ErrAIUnavailable},
		{name: "invalid JSON", ai: fakeAI{result: `{"awarded_points":100`}, want: service.ErrAIInvalidResponse},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			repo := newGameRepository()
			repo.progressByRole = map[string][]domain.Progress{"buyer": {{LevelID: 1, Stars: 1}, {LevelID: 2, Stars: 1}}}
			repo.steps = map[int]domain.ScenarioStep{1: {ID: 31, ScenarioID: 3, Number: 1, ResponseType: "mixed", MaxPoints: 100, FallbackMessage: "Начальная реплика"}}
			game := service.NewGameWithAI(repo, test.ai)
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
	ai := fakeAI{result: `{"awarded_points":75,"explanation":"Осторожная стратегия","reply":"Хорошо, продолжим в сервисе","risk_signals":[]}`}
	game := service.NewGameWithDependencies(repo, ai, func() bool { return false })
	state, err := game.StartFreePlay(context.Background(), 1, "seller")
	if err != nil || state.Attempt.IsScam == nil || *state.Attempt.IsScam || len(state.Messages) != 1 {
		t.Fatalf("StartFreePlay() = (%#v, %v), want hidden honest counterpart and first message", state, err)
	}
	for n := 1; n <= 2; n++ {
		text := "Продолжим безопасно в чате сервиса"
		next, completed, submitErr := game.SubmitAnswer(context.Background(), 1, state.Attempt.ID, service.AnswerCommand{FreeText: &text})
		if submitErr != nil || completed != nil || next.Attempt.FreeTextCount != n {
			t.Fatalf("turn %d = (%#v, %#v, %v), want continuation", n, next, completed, submitErr)
		}
	}
	text := "Завершаю сделку только штатным способом"
	_, completed, err := game.SubmitAnswer(context.Background(), 1, state.Attempt.ID, service.AnswerCommand{FreeText: &text, Finish: true})
	if err != nil || completed == nil || completed.Attempt.Score != 75 || repo.progress.Stars != 0 {
		t.Fatalf("free play completion = (%#v, %v), want score without level progress", completed, err)
	}
}

func TestFreePlayCompletesAutomaticallyOnFifthAnswer(t *testing.T) {
	repo := newGameRepository()
	repo.progressByRole = map[string][]domain.Progress{"buyer": {{LevelID: 4, Stars: 1}}}
	game := service.NewGameWithDependencies(repo, fakeAI{result: `{"awarded_points":75,"explanation":"Безопасно","reply":"Продолжим","risk_signals":[]}`}, func() bool { return true })
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
	game := service.NewGameWithDependencies(repo, fakeAI{result: `{"awarded_points":100,"explanation":"Безопасно","reply":"Продолжим","risk_signals":[]}`}, func() bool { return true })
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
	game := service.NewGameWithDependencies(repo, fakeAI{result: `{"awarded_points":0,"explanation":"Старт","reply":"Первая реплика","risk_signals":[]}`}, func() bool { return true })
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
	return domain.FreePlayConfig{UserRole: role, ProductContext: domain.JSONObject{"item": "товар"}, SystemPrompt: "Веди диалог", FinalRubric: domain.JSONObject{"safe": 100}}, nil
}
func (r *gameRepository) Scenario(id int) (domain.Scenario, error) {
	return domain.Scenario{ID: id, LevelID: id, UserRole: "buyer", AISystemPrompt: "Верни JSON"}, nil
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
func (r *gameRepository) Step(_ int, n int) (domain.ScenarioStep, error) {
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
func (r *gameRepository) UpdateFreeTextCount(id, count int) error {
	a := r.attempts[id]
	a.FreeTextCount = count
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
