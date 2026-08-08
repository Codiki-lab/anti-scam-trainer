package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"time"
)

type Service struct {
	repository Repository
	completion CompletionRepository
}

func New(repository Repository, completion CompletionRepository) *Service {
	return &Service{repository: repository, completion: completion}
}

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
		return apperrors.ErrInvalidAttemptStatusTransition
	}
	return s.repository.Update(attempt)
}

// Finish records a completed attempt and its resulting level progress atomically.
func (s *Service) Finish(attempt domain.Attempt, progress domain.Progress) error {
	current, err := s.repository.GetByID(attempt.ID)
	if err != nil {
		return err
	}
	if current.Status != domain.AttemptStatusInProgress {
		return apperrors.ErrInvalidAttemptStatusTransition
	}

	attempt.UserID = current.UserID
	attempt.ScenarioID = current.ScenarioID
	attempt.Status = domain.AttemptStatusCompleted
	if attempt.FinishedAt.IsZero() {
		attempt.FinishedAt = time.Now().UTC()
	}
	progress.UserID = attempt.UserID
	progress.Attempts++
	if attempt.Score > progress.BestScore {
		progress.BestScore = attempt.Score
		progress.Stars = domain.StarsFromScore(attempt.Score)
	}
	if progress.Stars > 0 && progress.PassedAt.IsZero() {
		progress.PassedAt = attempt.FinishedAt
	}

	return s.completion.InTransaction(func(store CompletionStore) error {
		if err := store.UpdateAttempt(attempt); err != nil {
			return err
		}
		return store.SaveProgress(progress)
	})
}

func (s *Service) Delete(id int) error             { return s.repository.Delete(id) }
func (s *Service) List() ([]domain.Attempt, error) { return s.repository.List() }
