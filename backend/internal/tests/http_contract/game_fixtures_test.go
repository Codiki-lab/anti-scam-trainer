package http_contract_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	"context"
	"errors"
	"sync/atomic"
	"time"
)

type contractAI struct{}

func (contractAI) Evaluate(context.Context, attemptsservice.EvaluationRequest) (attemptsservice.EvaluatorResult, error) {
	return contractEvaluation(), nil
}

func (contractAI) GenerateReply(_ context.Context, input attemptsservice.GenerationRequest) (attemptsservice.GeneratorResult, error) {
	return contractGeneration(input), nil
}

type profileContractAI struct{ safe bool }

func (a profileContractAI) Evaluate(context.Context, attemptsservice.EvaluationRequest) (attemptsservice.EvaluatorResult, error) {
	if a.safe {
		return attemptsservice.EvaluatorResult{Score: 4, IsSafe: true, RiskType: "social_engineering", Evaluation: "безопасно", SafeAction: "проверить внутри приложения"}, nil
	}
	return attemptsservice.EvaluatorResult{Score: 1, IsSafe: false, RiskType: "social_engineering", Evaluation: "рискованно", SafeAction: "проверить внутри приложения"}, nil
}

func (a profileContractAI) GenerateReply(_ context.Context, input attemptsservice.GenerationRequest) (attemptsservice.GeneratorResult, error) {
	return contractGeneration(input), nil
}

type countingContractAI struct{ calls int }

func (a *countingContractAI) Evaluate(context.Context, attemptsservice.EvaluationRequest) (attemptsservice.EvaluatorResult, error) {
	a.calls++
	return contractEvaluation(), nil
}

func (a *countingContractAI) GenerateReply(_ context.Context, input attemptsservice.GenerationRequest) (attemptsservice.GeneratorResult, error) {
	a.calls++
	return contractGeneration(input), nil
}

type blockingContractAI struct {
	started chan struct{}
	release chan struct{}
	calls   atomic.Int32
}

func (a *blockingContractAI) Evaluate(context.Context, attemptsservice.EvaluationRequest) (attemptsservice.EvaluatorResult, error) {
	if a.calls.Add(1) == 1 {
		close(a.started)
	}
	<-a.release
	return contractEvaluation(), nil
}

func (a *blockingContractAI) GenerateReply(_ context.Context, input attemptsservice.GenerationRequest) (attemptsservice.GeneratorResult, error) {
	if a.calls.Add(1) == 1 {
		close(a.started)
	}
	<-a.release
	return contractGeneration(input), nil
}

type failingContractAI struct{}

func (failingContractAI) Evaluate(context.Context, attemptsservice.EvaluationRequest) (attemptsservice.EvaluatorResult, error) {
	return attemptsservice.EvaluatorResult{}, attemptsservice.ErrAIUnavailable
}

func (failingContractAI) GenerateReply(context.Context, attemptsservice.GenerationRequest) (attemptsservice.GeneratorResult, error) {
	return attemptsservice.GeneratorResult{}, attemptsservice.ErrAIUnavailable
}

type invalidContractAI struct{}

func (invalidContractAI) Evaluate(context.Context, attemptsservice.EvaluationRequest) (attemptsservice.EvaluatorResult, error) {
	return attemptsservice.EvaluatorResult{}, attemptsservice.ErrAIInvalidResponse
}

func (invalidContractAI) GenerateReply(context.Context, attemptsservice.GenerationRequest) (attemptsservice.GeneratorResult, error) {
	return attemptsservice.GeneratorResult{}, attemptsservice.ErrAIInvalidResponse
}

type sequenceContractAI struct{ calls int }

func (a *sequenceContractAI) Evaluate(context.Context, attemptsservice.EvaluationRequest) (attemptsservice.EvaluatorResult, error) {
	a.calls++
	if a.calls > 1 {
		return attemptsservice.EvaluatorResult{}, attemptsservice.ErrAIUnavailable
	}
	return contractEvaluation(), nil
}

func (a *sequenceContractAI) GenerateReply(_ context.Context, input attemptsservice.GenerationRequest) (attemptsservice.GeneratorResult, error) {
	a.calls++
	if a.calls > 1 {
		return attemptsservice.GeneratorResult{}, attemptsservice.ErrAIUnavailable
	}
	return contractGeneration(input), nil
}

func contractEvaluation() attemptsservice.EvaluatorResult {
	return attemptsservice.EvaluatorResult{Score: 4, RiskType: "social_engineering", Evaluation: "безопасно", SafeAction: "остаться в сервисе"}
}

func contractGeneration(input attemptsservice.GenerationRequest) attemptsservice.GeneratorResult {
	return attemptsservice.GeneratorResult{Message: "продолжим", Tactic: input.AllowedTactics[0], Phase: input.Phase}
}

type httpGameStore struct {
	attempts map[int]domain.Attempt
	answers  []domain.UserAnswer
	messages []domain.DialogueMessage
	result   domain.AttemptResult
	events   []domain.MistakePatternEvent
	quizBest int
	streak   domain.Streak
	levels   []domain.Level
}

func newHTTPGameStore() *httpGameStore {
	return &httpGameStore{attempts: map[int]domain.Attempt{}, quizBest: 80, streak: domain.Streak{Current: 3, Longest: 4, ActiveToday: true}, levels: []domain.Level{{ID: 1, Number: 1}, {ID: 2, Number: 2}, {ID: 3, Number: 3}, {ID: 4, Number: 4}}}
}

func (s *httpGameStore) Levels(int, domain.UserRole) ([]domain.Level, []domain.Progress, error) {
	return append([]domain.Level(nil), s.levels...), []domain.Progress{{LevelID: 1, Stars: 1}, {LevelID: 2, Stars: 1}, {LevelID: 4, Stars: 1}}, nil
}

func (s *httpGameStore) PublishedScenario(level int, role domain.UserRole) (domain.Scenario, error) {
	if level != 3 {
		return domain.Scenario{}, errors.New("missing")
	}
	return domain.Scenario{ID: 3, LevelID: 3, UserRole: role}, nil
}

func (s *httpGameStore) TopicLevels(userID int, role domain.UserRole, _ int) ([]domain.Level, []domain.Progress, bool, error) {
	levels, progress, err := s.Levels(userID, role)
	return levels, progress, true, err
}

func (s *httpGameStore) PublishedTopicScenario(level int, role domain.UserRole, _ int) (domain.Scenario, error) {
	return s.PublishedScenario(level, role)
}

func (s *httpGameStore) Result(int) (domain.AttemptResult, error) { return s.result, nil }

func (s *httpGameStore) FreePlayUnlocked(int, domain.UserRole) (bool, error) { return true, nil }

func (s *httpGameStore) FreePlayConfig(role domain.UserRole) (domain.FreePlayConfig, error) {
	return domain.FreePlayConfig{UserRole: role}, nil
}

func (s *httpGameStore) Scenario(int) (domain.Scenario, error) {
	return domain.Scenario{ID: 3, LevelID: 3, UserRole: "buyer"}, nil
}

func (s *httpGameStore) FindInProgress(userID, scenarioID int) (domain.Attempt, error) {
	for _, attempt := range s.attempts {
		if attempt.UserID == userID && attempt.ScenarioID == scenarioID {
			return attempt, nil
		}
	}
	return domain.Attempt{}, errors.New("missing")
}

func (s *httpGameStore) FindInProgressFreePlay(int, domain.UserRole) (domain.Attempt, error) {
	return domain.Attempt{}, errors.New("missing")
}

func (s *httpGameStore) CreateGameAttempt(attempt domain.Attempt) (domain.Attempt, error) {
	if previous, exists := s.attempts[1]; exists && previous.Status == domain.AttemptStatusCompleted {
		s.answers = nil
		s.messages = nil
	}
	attempt.ID = 1
	s.attempts[1] = attempt
	return attempt, nil
}

func (s *httpGameStore) StartFreePlay(attempt domain.Attempt, message domain.DialogueMessage) (domain.Attempt, error) {
	created, err := s.CreateGameAttempt(attempt)
	if err != nil {
		return domain.Attempt{}, err
	}
	message.AttemptID = created.ID
	s.messages = append(s.messages, message)
	return created, nil
}

func (s *httpGameStore) GetGameAttempt(id int) (domain.Attempt, error) {
	a, ok := s.attempts[id]
	if !ok {
		return domain.Attempt{}, errors.New("missing")
	}
	return a, nil
}

func (s *httpGameStore) Step(_ int, number int) (domain.ScenarioStep, error) {
	if number != 1 {
		return domain.ScenarioStep{}, errors.New("missing")
	}
	return domain.ScenarioStep{ID: 31, ScenarioID: 3, Number: 1, ResponseType: "mixed", FallbackMessage: "Начальная реплика", Options: []domain.ScenarioOption{{ID: 11, Points: 100}}}, nil
}

func (s *httpGameStore) Answers(int) ([]domain.UserAnswer, error) {
	return append([]domain.UserAnswer(nil), s.answers...), nil
}

func (s *httpGameStore) Messages(int) ([]domain.DialogueMessage, error) {
	return append([]domain.DialogueMessage(nil), s.messages...), nil
}

func (s *httpGameStore) AwardedPoints(int) (int, error) {
	total := 0
	for _, a := range s.answers {
		total += a.AwardedPoints
	}
	return total, nil
}

func (s *httpGameStore) Advance(id, next int) error {
	a := s.attempts[id]
	a.CurrentStepNumber = next
	s.attempts[id] = a
	return nil
}

func (s *httpGameStore) Abandon(id int, _ time.Time) error {
	a := s.attempts[id]
	a.Status = domain.AttemptStatusAbandoned
	s.attempts[id] = a
	return nil
}

func (s *httpGameStore) Complete(action func(attemptsservice.GameCompletionStore) error) error {
	return action(s)
}

func (s *httpGameStore) SaveAnswer(answer domain.UserAnswer) error {
	s.answers = append(s.answers, answer)
	return nil
}

func (s *httpGameStore) SaveMessage(message domain.DialogueMessage) error {
	s.messages = append(s.messages, message)
	return nil
}

func (s *httpGameStore) AdvanceAttempt(id, next int) error { return s.Advance(id, next) }

func (s *httpGameStore) UpdateDialogueState(id, count int, phase, summary string) error {
	a := s.attempts[id]
	a.FreeTextCount = count
	a.DialoguePhase = phase
	a.CompactSummary = summary
	s.attempts[id] = a
	return nil
}

func (s *httpGameStore) CompleteAttempt(attempt domain.Attempt) error {
	s.attempts[attempt.ID] = attempt
	return nil
}

func (s *httpGameStore) SaveProgress(domain.Progress) error { return nil }
func (s *httpGameStore) FinalizeLearning(result *domain.AttemptResult) error {
	s.result = *result
	return nil
}
func (s *httpGameStore) RecordMistakePatternEvents(_ int, _ int, _ int, _ domain.UserRole, events []domain.MistakePatternEvent) error {
	s.events = append(s.events, events...)
	return nil
}
func (s *httpGameStore) MistakePatternStats(int, domain.UserRole) ([]domain.MistakePatternStats, error) {
	if len(s.events) == 0 && s.result.MicroQuestion != nil {
		return []domain.MistakePatternStats{{PatternCode: s.result.MicroQuestion.PatternCode, UnsafeCount: 3, RecentUnsafe: 2}}, nil
	}
	if len(s.events) == 0 {
		return nil, nil
	}
	stats := domain.MistakePatternStats{PatternCode: s.events[0].PatternCode}
	for _, event := range s.events {
		if event.IsSafe {
			stats.SafeCount++
		} else {
			stats.UnsafeCount++
			stats.RecentUnsafe++
		}
	}
	return []domain.MistakePatternStats{stats}, nil
}
func (s *httpGameStore) SaveResult(result domain.AttemptResult) error {
	s.result = result
	return nil
}
