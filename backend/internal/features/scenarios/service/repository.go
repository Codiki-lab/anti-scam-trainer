package service

import "anti-scam-trainer/backend/internal/core/domain"

type Repository interface {
	Create(domain.Scenario) (domain.Scenario, error)
	GetByID(int) (domain.Scenario, error)
	Update(domain.Scenario) error
	Delete(int) error
	List() ([]domain.Scenario, error)
}
