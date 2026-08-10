package domain

// UserAnswer records either a selected option or free text for one scenario step.
type UserAnswer struct {
	AttemptID     int
	StepID        int
	OptionID      *int
	FreeText      string
	AwardedPoints int
	Explanation   string
	OptionText    string
	Evaluation    *AIEvaluation
	TurnNumber    int
}

type AIEvaluation struct {
	Score           int      `json:"score"`
	IsSafe          bool     `json:"is_safe"`
	RiskType        string   `json:"risk_type"`
	DetectedSignals []string `json:"detected_signals"`
	Evaluation      string   `json:"evaluation"`
	SafeAction      string   `json:"safe_action"`
}

type AnswerBreakdown struct {
	StepID      int          `json:"step_id"`
	StepNumber  int          `json:"step_number"`
	AnswerType  string       `json:"answer_type"`
	OptionID    int          `json:"option_id,omitempty"`
	Points      int          `json:"points"`
	Assessment  string       `json:"assessment"`
	Explanation string       `json:"explanation"`
	SafeAction  string       `json:"safe_action"`
	OptionText  string       `json:"option_text,omitempty"`
	FreeText    string       `json:"free_text,omitempty"`
	RiskSignals []RiskSignal `json:"risk_signals"`
}

type RiskSignal struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

func AssessmentForPoints(points int) string {
	switch {
	case points <= 25:
		return "unsafe"
	case points == 50:
		return "risky"
	case points == 75:
		return "mostly_safe"
	default:
		return "safe"
	}
}
