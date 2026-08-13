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
	Feedback        ResultFeedback     `json:"feedback"`
	MicroQuestion   *MicroQuestion     `json:"micro_question,omitempty"`
}

// ResultFeedback is deliberately compact: it is educational output, not a
// copy of evaluator policy, prompts, or raw model output.
type ResultFeedback struct {
	Reason          string
	RiskSignals     []RiskSignal
	SafeAlternative string
}

type MicroQuestion struct {
	PatternCode string
	Question    string
	Options     []string
	Correct     int
}

func MicroQuestionFor(pattern string) *MicroQuestion {
	questions := map[string]MicroQuestion{
		"external_link":      {Question: "Что делать, если собеседник прислал ссылку для оформления?", Options: []string{"Самостоятельно открыть заказ внутри приложения", "Перейти по ссылке из чата"}},
		"phishing":           {Question: "Что делать, если собеседник прислал ссылку для оформления?", Options: []string{"Самостоятельно открыть заказ внутри приложения", "Перейти по ссылке из чата"}},
		"credential_request": {Question: "Как поступить с кодом подтверждения из сообщения?", Options: []string{"Никому не сообщать код", "Передать код собеседнику для проверки"}},
		"sms_code":           {Question: "Как поступить с кодом подтверждения из сообщения?", Options: []string{"Никому не сообщать код", "Передать код собеседнику для проверки"}},
		"prepayment":         {Question: "Как проверить просьбу о предоплате?", Options: []string{"Проверить условия и оплату внутри приложения", "Сразу перевести деньги по реквизитам"}},
		"fake_payment":       {Question: "Как убедиться, что оплата действительно поступила?", Options: []string{"Проверить статус внутри приложения", "Довериться скриншоту собеседника"}},
		"external_messenger": {Question: "Где безопаснее продолжать обсуждение сделки?", Options: []string{"В чате сервиса", "В стороннем мессенджере"}},
		"fake_delivery":      {Question: "Как проверить просьбу об оплате доставки или страховки?", Options: []string{"Открыть заказ самостоятельно внутри приложения", "Оплатить комиссию по инструкции собеседника"}},
		"account_takeover":   {Question: "Что делать при просьбе подтвердить вход или действие в аккаунте?", Options: []string{"Самостоятельно проверить аккаунт и никому не передавать секреты", "Выполнить просьбу собеседника для ускорения сделки"}},
		"pressure":           {Question: "Что делать, когда собеседник торопит с решением?", Options: []string{"Остановиться и самостоятельно проверить условия", "Согласиться, чтобы не потерять сделку"}},
	}
	question, ok := questions[pattern]
	if !ok {
		question = MicroQuestion{Question: "Какое действие безопаснее в этой ситуации?", Options: []string{"Проверить условия самостоятельно внутри приложения", "Сразу выполнить просьбу собеседника"}}
	}
	question.PatternCode = pattern
	question.Correct = 0
	return &question
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
