package service

import "anti-scam-trainer/backend/internal/core/domain"

// Repository is the persistence port for independent progress read scenarios.
type Repository interface {
	Get(userID, levelID int, userRole string) (domain.Progress, error)
}
