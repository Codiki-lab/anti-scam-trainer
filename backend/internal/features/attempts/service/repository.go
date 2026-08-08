package service

import "anti-scam-trainer/backend/internal/core/domain"

type Repository interface {
	Create(domain.Attempt) (domain.Attempt, error)
	GetByID(int) (domain.Attempt, error)
	Update(domain.Attempt) error
	Delete(int) error
	List() ([]domain.Attempt, error)
}
