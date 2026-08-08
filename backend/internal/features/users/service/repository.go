package service

import "anti-scam-trainer/backend/internal/core/domain"

// Repository is the users persistence port owned by the users service.
type Repository interface {
	Create(domain.User) (domain.User, error)
	GetByID(int) (domain.User, error)
	GetByUsername(string) (domain.User, error)
}
