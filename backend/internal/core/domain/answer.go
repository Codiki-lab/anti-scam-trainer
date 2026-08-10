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
	StepID      int      `json:"step_id"`
	OptionID    int      `json:"option_id"`
	Points      int      `json:"points"`
	Explanation string   `json:"explanation"`
	OptionText  string   `json:"option_text"`
	FreeText    string   `json:"free_text"`
	RiskSignals []string `json:"risk_signals"`
}
