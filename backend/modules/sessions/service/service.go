package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/errors"
)

type Repository interface {
	Create(domain.Attempt) (domain.Attempt, error)
	GetByID(int) (domain.Attempt, error)
	Update(domain.Attempt) error
	Delete(int) error
	List() ([]domain.Attempt, error)
}
type Service struct{ repository Repository }

func New(repository Repository) *Service { return &Service{repository: repository} }
func (s *Service) Create(attempt domain.Attempt) (domain.Attempt, error) {
	return s.repository.Create(attempt)
}
func (s *Service) GetByID(id int) (domain.Attempt, error) { return s.repository.GetByID(id) }
func (s *Service) Update(attempt domain.Attempt) error {
	current, err := s.repository.GetByID(attempt.ID)
	if err != nil {
		return err
	}
	if !domain.CanTransitionAttemptStatus(current.Status, attempt.Status) {
		return apperrors.ErrInvalidChatSessionStatusTransition
	}
	return s.repository.Update(attempt)
}
func (s *Service) Delete(id int) error             { return s.repository.Delete(id) }
func (s *Service) List() ([]domain.Attempt, error) { return s.repository.List() }
