package service

import "anti-scam-trainer/backend/internal/core/domain"

type Repository interface {
	Create(domain.Scenario) (domain.Scenario, error)
	GetByID(int) (domain.Scenario, error)
	Update(domain.Scenario) error
	Delete(int) error
	List() ([]domain.Scenario, error)
}
type Service struct{ repository Repository }

func New(repository Repository) *Service { return &Service{repository: repository} }
func (s *Service) Create(scenario domain.Scenario) (domain.Scenario, error) {
	return s.repository.Create(scenario)
}
func (s *Service) GetByID(id int) (domain.Scenario, error) { return s.repository.GetByID(id) }
func (s *Service) Update(scenario domain.Scenario) error   { return s.repository.Update(scenario) }
func (s *Service) Delete(id int) error                     { return s.repository.Delete(id) }
func (s *Service) List() ([]domain.Scenario, error)        { return s.repository.List() }
