package service

import "anti-scam-trainer/backend/internal/core/domain"

type Repository interface {
	Create(domain.User) (domain.User, error)
	GetByID(int) (domain.User, error)
	GetByExternalID(string) (domain.User, error)
	Update(domain.User) error
	Delete(int) error
	List() ([]domain.User, error)
}

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
