package attempts_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/features/attempts/service"
	"context"
	"errors"
	"strconv"
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

func intPointer(value int) *int { return &value }

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

type gameRepository struct {
	attempts            map[int]domain.Attempt
	steps               map[int]domain.ScenarioStep
	answers             []domain.UserAnswer
	messages            []domain.DialogueMessage
	progress            domain.Progress
	progressByRole      map[string][]domain.Progress
	freePlayUnlocked    bool
	next                int
	failCompleteAttempt bool
	failStartFreePlay   bool
}

func newGameRepository() *gameRepository {
	return &gameRepository{attempts: map[int]domain.Attempt{}, next: 1, freePlayUnlocked: true, steps: map[int]domain.ScenarioStep{1: {ID: 1, ScenarioID: 1, Number: 1, MaxPoints: 100, FallbackMessage: "Первая реплика", Options: []domain.ScenarioOption{{ID: 11, Points: 100}}}, 2: {ID: 2, ScenarioID: 1, Number: 2, MaxPoints: 100, FallbackMessage: "Вторая реплика", Options: []domain.ScenarioOption{{ID: 21, Points: 100}}}}}
}

func (r *gameRepository) Levels(_ int, role domain.UserRole) ([]domain.Level, []domain.Progress, error) {
	return []domain.Level{{ID: 1, Number: 1}, {ID: 2, Number: 2}, {ID: 3, Number: 3}, {ID: 4, Number: 4}}, r.progressByRole[string(role)], nil
}

func (r *gameRepository) PublishedScenario(level int, role domain.UserRole) (domain.Scenario, error) {
	if (role == "buyer" || role == "seller") && level >= 1 && level <= 4 {
		id := level
		if role == "seller" {
			id += 2
		}
		return domain.Scenario{ID: id, LevelID: level, UserRole: role}, nil
	}
	return domain.Scenario{}, errors.New("missing")
}

func (r *gameRepository) FreePlayUnlocked(int, domain.UserRole) (bool, error) {
	return r.freePlayUnlocked, nil
}

func (r *gameRepository) FreePlayConfig(role domain.UserRole) (domain.FreePlayConfig, error) {
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

func (r *gameRepository) FindInProgressFreePlay(user int, role domain.UserRole) (domain.Attempt, error) {
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

func (r *gameRepository) SaveProgress(p domain.Progress) error {
	r.progress = p
	return nil
}

func (r *gameRepository) FinalizeLearning(*domain.AttemptResult) error { return nil }
func (r *gameRepository) RecordMistakePatternEvents(int, int, int, domain.UserRole, []domain.MistakePatternEvent) error {
	return nil
}
func (r *gameRepository) MistakePatternStats(int, domain.UserRole) ([]domain.MistakePatternStats, error) {
	return nil, nil
}
func (r *gameRepository) SaveResult(domain.AttemptResult) error { return nil }
