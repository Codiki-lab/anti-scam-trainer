package domain

// UserAnswer records either a selected option or free text for one scenario step.
type UserAnswer struct {
	AttemptID int
	StepID    int
	OptionID  *int
	FreeText  string
}
