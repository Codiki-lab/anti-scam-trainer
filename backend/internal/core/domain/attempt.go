package domain

import "time"

const (
	AttemptStatusInProgress = "IN_PROGRESS"
	AttemptStatusCompleted  = "COMPLETED"
	AttemptStatusAbandoned  = "ABANDONED"
)

type Attempt struct {
	ID                int
	UserID            int
	ScenarioID        int
	Mode              AttemptMode
	UserRole          string
	IsScam            *bool
	Status            string
	StartedAt         time.Time
	FinishedAt        time.Time
	Score             int
	MaxScore          int
	CurrentStepNumber int
	FreeTextCount     int
	FinalBreakdown    []AnswerBreakdown
}

type AttemptMode string

const (
	AttemptModeScenario AttemptMode = "scenario"
	AttemptModeFreePlay AttemptMode = "free_play"
)

func CanTransitionAttemptStatus(currentStatus, nextStatus string) bool {
	if currentStatus == nextStatus {
		return true
	}

	return currentStatus == AttemptStatusInProgress &&
		(nextStatus == AttemptStatusCompleted || nextStatus == AttemptStatusAbandoned)
}
