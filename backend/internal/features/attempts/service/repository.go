package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"time"
)

type Repository interface {
	Create(domain.Attempt) (domain.Attempt, error)
	GetByID(int) (domain.Attempt, error)
	Update(domain.Attempt) error
	Delete(int) error
	ListByUserID(int) ([]domain.Attempt, error)
}

type GameRepository interface {
	Levels(userID int, userRole string) ([]domain.Level, []domain.Progress, error)
	PublishedScenario(levelNumber int, userRole string) (domain.Scenario, error)
	Scenario(id int) (domain.Scenario, error)
	FindInProgress(userID, scenarioID int) (domain.Attempt, error)
	CreateGameAttempt(domain.Attempt, domain.Message) (domain.Attempt, error)
	GetGameAttempt(id int) (domain.Attempt, error)
	Step(scenarioID, number int) (domain.ScenarioStep, error)
	Answers(attemptID int) ([]domain.UserAnswer, error)
	Messages(attemptID int) ([]domain.Message, error)
	AwardedPoints(attemptID int) (int, error)
	Advance(attemptID, nextStepNumber int) error
	Abandon(attemptID int, finishedAt time.Time) error
	Complete(func(GameCompletionStore) error) error
}

type GameCompletionStore interface {
	SaveAnswer(domain.UserAnswer, int, string) error
	SaveMessage(domain.Message) error
	AdvanceAttempt(attemptID, nextStepNumber int) error
	CompleteAttempt(domain.Attempt) error
	SaveProgress(domain.Progress) error
}
