package domain

import "time"

const (
	AttemptStatusInProgress AttemptStatus = "IN_PROGRESS"
	AttemptStatusCompleted  AttemptStatus = "COMPLETED"
	AttemptStatusAbandoned  AttemptStatus = "ABANDONED"
)

type AttemptStatus string

type Attempt struct {
	ID                int
	UserID            int
	ScenarioID        int
	Mode              AttemptMode
	UserRole          UserRole
	IsScam            *bool
	Status            AttemptStatus
	StartedAt         time.Time
	FinishedAt        time.Time
	Score             int
	MaxScore          int
	CurrentStepNumber int
	FreeTextCount     int
	DialoguePhase     string
	CompactSummary    string
	FinalBreakdown    []AnswerBreakdown
}

type AttemptResult struct {
	AttemptID       int                `json:"attempt_id"`
	Score           int                `json:"score"`
	Stars           int                `json:"stars"`
	DecisionReview  []AnswerBreakdown  `json:"decision_review"`
	RiskSignals     []RiskSignal       `json:"risk_signals"`
	SafeActions     []string           `json:"safe_actions"`
	LevelProgress   TopicLevelProgress `json:"level_progress"`
	TopicID         int                `json:"topic_id"`
	TopicCompleted  bool               `json:"topic_completed"`
	NextAction      *ContinueAction    `json:"next_action"`
	NewAchievements []Achievement      `json:"new_achievements"`
	Streak          Streak             `json:"streak"`
	IsScam          *bool              `json:"is_scam,omitempty"`
}

type AttemptMode string

const (
	AttemptModeScenario AttemptMode = "scenario"
	AttemptModeFreePlay AttemptMode = "free_play"
)

func ValidAttemptStatus(status AttemptStatus) bool {
	return status == AttemptStatusInProgress || status == AttemptStatusCompleted || status == AttemptStatusAbandoned
}

func CanTransitionAttemptStatus(currentStatus, nextStatus AttemptStatus) bool {
	if !ValidAttemptStatus(currentStatus) || !ValidAttemptStatus(nextStatus) {
		return false
	}
	if currentStatus == nextStatus {
		return true
	}

	return currentStatus == AttemptStatusInProgress &&
		(nextStatus == AttemptStatusCompleted || nextStatus == AttemptStatusAbandoned)
}

func (attempt *Attempt) TransitionTo(nextStatus AttemptStatus) bool {
	if !CanTransitionAttemptStatus(attempt.Status, nextStatus) {
		return false
	}
	attempt.Status = nextStatus
	return true
}
