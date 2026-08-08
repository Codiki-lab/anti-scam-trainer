package service

import "anti-scam-trainer/backend/internal/core/domain"

// CompletionStore exposes only the writes needed to finish one attempt.
type CompletionStore interface {
	UpdateAttempt(domain.Attempt) error
	SaveProgress(domain.Progress) error
}

// CompletionRepository runs an attempt completion function atomically.
type CompletionRepository interface {
	InTransaction(func(CompletionStore) error) error
}
