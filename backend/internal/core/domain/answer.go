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
	AwardedPoints int      `json:"awarded_points"`
	Explanation   string   `json:"explanation"`
	Reply         string   `json:"reply"`
	RiskSignals   []string `json:"risk_signals"`
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
