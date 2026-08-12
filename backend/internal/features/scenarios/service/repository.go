package service

import "anti-scam-trainer/backend/internal/core/domain"

type ContentRepository interface {
	CreateContent(domain.Scenario) (domain.Scenario, error)
	ListContent() ([]domain.Scenario, error)
	UpdateContent(domain.Scenario) error
	SetContentStatus(id int, status domain.ScenarioStatus, archived bool) error
	CreateStep(domain.ScenarioStep) (domain.ScenarioStep, error)
	CreateOption(domain.ScenarioOption) (domain.ScenarioOption, error)
	StepScenario(stepID int) (domain.Scenario, error)
	OptionScenario(optionID int) (domain.Scenario, error)
	UpdateStep(domain.ScenarioStep) error
	DeleteStep(id int) error
	UpdateOption(domain.ScenarioOption) error
	DeleteOption(id int) error
	ContentScenario(id int) (domain.Scenario, error)
	ValidContent(id int) (bool, error)
}
