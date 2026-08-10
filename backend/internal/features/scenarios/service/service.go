package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
)

type Service struct{ repository ContentRepository }

func New(repository ContentRepository) *Service {
	return &Service{repository: repository}
}
func (s *Service) Create(scenario domain.Scenario) (domain.Scenario, error) {
	scenario.Status = domain.ScenarioStatusDraft
	return s.repository.CreateContent(scenario)
}
func (s *Service) List() ([]domain.Scenario, error) { return s.repository.ListContent() }
func (s *Service) Update(scenario domain.Scenario) error {
	current, err := s.repository.ContentScenario(scenario.ID)
	if err != nil {
		return err
	}
	if current.Status != domain.ScenarioStatusDraft {
		return apperrors.ErrInvalidScenarioState
	}
	return s.repository.UpdateContent(scenario)
}
func (s *Service) Publish(id int) error {
	current, err := s.repository.ContentScenario(id)
	if err != nil {
		return err
	}
	if current.Status != domain.ScenarioStatusDraft {
		return apperrors.ErrInvalidScenarioState
	}
	if !domain.ValidRiskType(current.RiskType) || !domain.ValidProductContext(current.ProductContext) {
		return apperrors.ErrInvalidScenarioState
	}
	valid, err := s.repository.ValidContent(id)
	if err != nil {
		return err
	}
	if !valid {
		return apperrors.ErrInvalidScenarioState
	}
	return s.repository.SetContentStatus(id, domain.ScenarioStatusPublished, false)
}
func (s *Service) Deactivate(id int) error {
	current, err := s.repository.ContentScenario(id)
	if err != nil {
		return err
	}
	if current.Status != domain.ScenarioStatusPublished {
		return apperrors.ErrInvalidScenarioState
	}
	return s.repository.SetContentStatus(id, domain.ScenarioStatusDraft, false)
}
func (s *Service) Archive(id int) error {
	return s.repository.SetContentStatus(id, domain.ScenarioStatusArchived, true)
}
func (s *Service) Restore(id int) error {
	current, err := s.repository.ContentScenario(id)
	if err != nil {
		return err
	}
	if current.Status != domain.ScenarioStatusArchived {
		return apperrors.ErrInvalidScenarioState
	}
	return s.repository.SetContentStatus(id, domain.ScenarioStatusDraft, false)
}
func (s *Service) AddStep(step domain.ScenarioStep) (domain.ScenarioStep, error) {
	if len([]rune(step.CounterpartyMessage)) == 0 || len([]rune(step.CounterpartyMessage)) > 280 {
		return domain.ScenarioStep{}, apperrors.ErrInvalidAnswer
	}
	scenario, err := s.repository.ContentScenario(step.ScenarioID)
	if err != nil {
		return domain.ScenarioStep{}, err
	}
	if scenario.Status != domain.ScenarioStatusDraft {
		return domain.ScenarioStep{}, apperrors.ErrInvalidScenarioState
	}
	return s.repository.CreateStep(step)
}
func (s *Service) AddOption(option domain.ScenarioOption) (domain.ScenarioOption, error) {
	if !validOption(option) {
		return domain.ScenarioOption{}, apperrors.ErrInvalidAnswer
	}
	scenario, err := s.repository.StepScenario(option.StepID)
	if err != nil {
		return domain.ScenarioOption{}, err
	}
	if scenario.Status != domain.ScenarioStatusDraft {
		return domain.ScenarioOption{}, apperrors.ErrInvalidScenarioState
	}
	return s.repository.CreateOption(option)
}

func (s *Service) UpdateStep(step domain.ScenarioStep) error {
	if len([]rune(step.CounterpartyMessage)) == 0 || len([]rune(step.CounterpartyMessage)) > 280 {
		return apperrors.ErrInvalidAnswer
	}
	scenario, err := s.repository.StepScenario(step.ID)
	if err != nil {
		return err
	}
	if scenario.Status != domain.ScenarioStatusDraft {
		return apperrors.ErrInvalidScenarioState
	}
	return s.repository.UpdateStep(step)
}

func (s *Service) DeleteStep(id int) error {
	scenario, err := s.repository.StepScenario(id)
	if err != nil {
		return err
	}
	if scenario.Status != domain.ScenarioStatusDraft {
		return apperrors.ErrInvalidScenarioState
	}
	return s.repository.DeleteStep(id)
}

func (s *Service) UpdateOption(option domain.ScenarioOption) error {
	if !validOption(option) {
		return apperrors.ErrInvalidAnswer
	}
	scenario, err := s.repository.OptionScenario(option.ID)
	if err != nil {
		return err
	}
	if scenario.Status != domain.ScenarioStatusDraft {
		return apperrors.ErrInvalidScenarioState
	}
	return s.repository.UpdateOption(option)
}

func validOption(option domain.ScenarioOption) bool {
	return domain.ValidOptionPoints(option.Points) && len([]rune(option.Text)) > 0 && len([]rune(option.Text)) <= 140 && len([]rune(option.Reaction)) <= 280
}

func (s *Service) DeleteOption(id int) error {
	scenario, err := s.repository.OptionScenario(id)
	if err != nil {
		return err
	}
	if scenario.Status != domain.ScenarioStatusDraft {
		return apperrors.ErrInvalidScenarioState
	}
	return s.repository.DeleteOption(id)
}
