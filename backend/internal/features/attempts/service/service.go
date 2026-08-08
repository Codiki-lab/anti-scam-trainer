package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"time"
)

var ErrAttemptNotFound = apperrors.ErrAttemptNotFound

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

func (s *Service) GetByIDForUser(userID, attemptID int) (domain.Attempt, error) {
	attempt, err := s.repository.GetByID(attemptID)
	if err != nil || attempt.UserID != userID {
		return domain.Attempt{}, ErrAttemptNotFound
	}
	return attempt, nil
}

func (s *Service) CreateForUser(userID int, attempt domain.Attempt) (domain.Attempt, error) {
	attempt.UserID = userID
	return s.repository.Create(attempt)
}

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

func (s *Service) UpdateForUser(userID int, attempt domain.Attempt) error {
	current, err := s.GetByIDForUser(userID, attempt.ID)
	if err != nil {
		return err
	}
	attempt.UserID = current.UserID
	return s.Update(attempt)
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

func (s *Service) Delete(id int) error { return s.repository.Delete(id) }

func (s *Service) DeleteForUser(userID, attemptID int) error {
	if _, err := s.GetByIDForUser(userID, attemptID); err != nil {
		return err
	}
	return s.repository.Delete(attemptID)
}

func (s *Service) ListForUser(userID int) ([]domain.Attempt, error) {
	return s.repository.ListByUserID(userID)
}
