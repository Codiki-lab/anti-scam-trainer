package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"time"
)

type GameRepository interface {
	Levels(userID int, userRole domain.UserRole) ([]domain.Level, []domain.Progress, error)
	PublishedScenario(levelNumber int, userRole domain.UserRole) (domain.Scenario, error)
	FreePlayUnlocked(userID int, userRole domain.UserRole) (bool, error)
	FreePlayConfig(userRole domain.UserRole) (domain.FreePlayConfig, error)
	Scenario(id int) (domain.Scenario, error)
	FindInProgress(userID, scenarioID int) (domain.Attempt, error)
	FindInProgressFreePlay(userID int, userRole domain.UserRole) (domain.Attempt, error)
	CreateGameAttempt(domain.Attempt) (domain.Attempt, error)
	StartFreePlay(domain.Attempt, domain.DialogueMessage) (domain.Attempt, error)
	GetGameAttempt(id int) (domain.Attempt, error)
	Step(scenarioID, number int) (domain.ScenarioStep, error)
	Answers(attemptID int) ([]domain.UserAnswer, error)
	Messages(attemptID int) ([]domain.DialogueMessage, error)
	AwardedPoints(attemptID int) (int, error)
	Advance(attemptID, nextStepNumber int) error
	Abandon(attemptID int, finishedAt time.Time) error
	Complete(func(GameCompletionStore) error) error
}

type TopicGameRepository interface {
	TopicLevels(userID int, userRole domain.UserRole, topicID int) ([]domain.Level, []domain.Progress, bool, error)
	PublishedTopicScenario(levelNumber int, userRole domain.UserRole, topicID int) (domain.Scenario, error)
	Result(attemptID int) (domain.AttemptResult, error)
}

type GameCompletionStore interface {
	SaveAnswer(domain.UserAnswer) error
	SaveMessage(domain.DialogueMessage) error
	AdvanceAttempt(attemptID, nextStepNumber int) error
	UpdateDialogueState(attemptID, count int, phase, summary string) error
	CompleteAttempt(domain.Attempt) error
	SaveProgress(domain.Progress) error
	FinalizeLearning(*domain.AttemptResult) error
	RecordMistakePatternEvents(userID, attemptID, topicID int, role domain.UserRole, events []domain.MistakePatternEvent) error
	MistakePatternStats(userID int, role domain.UserRole) ([]domain.MistakePatternStats, error)
	SaveResult(domain.AttemptResult) error
}
