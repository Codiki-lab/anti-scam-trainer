package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"time"
)

type GameRepository interface {
	Levels(userID int, userRole string) ([]domain.Level, []domain.Progress, error)
	PublishedScenario(levelNumber int, userRole string) (domain.Scenario, error)
	FreePlayConfig(userRole string) (domain.FreePlayConfig, error)
	Scenario(id int) (domain.Scenario, error)
	FindInProgress(userID, scenarioID int) (domain.Attempt, error)
	FindInProgressFreePlay(userID int, userRole string) (domain.Attempt, error)
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
	TopicLevels(userID int, userRole string, topicID int) ([]domain.Level, []domain.Progress, bool, error)
	PublishedTopicScenario(levelNumber int, userRole string, topicID int) (domain.Scenario, error)
	FreePlayUnlocked(userID int, userRole string) (bool, error)
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
}
