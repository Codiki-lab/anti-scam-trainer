package service

import "anti-scam-trainer/backend/internal/core/domain"

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(user domain.User) (domain.User, error) { return s.repository.Create(user) }
func (s *Service) GetByID(id int) (domain.User, error)          { return s.repository.GetByID(id) }
func (s *Service) GetByExternalID(id string) (domain.User, error) {
	return s.repository.GetByExternalID(id)
}
func (s *Service) Update(user domain.User) error { return s.repository.Update(user) }
func (s *Service) Delete(id int) error           { return s.repository.Delete(id) }
func (s *Service) List() ([]domain.User, error)  { return s.repository.List() }
