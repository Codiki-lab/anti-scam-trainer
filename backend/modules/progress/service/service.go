package service

import "anti-scam-trainer/backend/internal/core/domain"

// Repository is the future persistence seam for role-specific level progress.
// The current HTTP contract exposes no progress operations yet.
type Repository interface {
	Get(userID, levelID int, userRole string) (domain.Progress, error)
	Save(domain.Progress) error
}

type Service struct{ repository Repository }

func New(repository Repository) *Service { return &Service{repository: repository} }
