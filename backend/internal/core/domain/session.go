package domain

import "time"

const (
	AttemptStatusInProgress = "IN_PROGRESS"
	AttemptStatusCompleted  = "COMPLETED"
	AttemptStatusAbandoned  = "ABANDONED"
)

type Attempt struct {
	ID         int
	UserID     int
	ChatID     int
	Status     string
	StartedAt  time.Time
	FinishedAt time.Time
	Score      int
}

func CanTransitionAttemptStatus(currentStatus, nextStatus string) bool {
	if currentStatus == nextStatus {
		return true
	}

	return currentStatus == AttemptStatusInProgress &&
		(nextStatus == AttemptStatusCompleted || nextStatus == AttemptStatusAbandoned)
}
