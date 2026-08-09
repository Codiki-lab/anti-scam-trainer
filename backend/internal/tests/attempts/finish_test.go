package attempts_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"anti-scam-trainer/backend/internal/features/attempts/service"
	"errors"
	"testing"
)

func TestFinishCommitsAttemptAndProgressTogether(t *testing.T) {
	repository := &attemptRepository{attempt: domain.Attempt{ID: 8, UserID: 3, Status: domain.AttemptStatusInProgress}}
	completion := &transactionalCompletionRepository{}
	attempts := service.New(repository, completion)

	err := attempts.Finish(
		domain.Attempt{ID: 8, UserID: 3, Score: 72},
		domain.Progress{LevelID: 4, UserRole: "seller"},
	)
	if err != nil {
		t.Fatalf("Finish() error = %v", err)
	}
	if completion.attempt.Status != domain.AttemptStatusCompleted {
		t.Fatalf("attempt status = %q, want %q", completion.attempt.Status, domain.AttemptStatusCompleted)
	}
	if completion.progress.UserID != 3 || completion.progress.BestScore != 72 || completion.progress.Stars != 2 || completion.progress.Attempts != 1 {
		t.Fatalf("progress = %#v, want user 3, best score 72, 2 stars, 1 attempt", completion.progress)
	}
	if completion.attempt.FinishedAt.IsZero() {
		t.Fatal("completed attempt has no finished time")
	}
}

func TestFinishRollsBackWhenProgressCannotBeSaved(t *testing.T) {
	repository := &attemptRepository{attempt: domain.Attempt{ID: 8, UserID: 3, Status: domain.AttemptStatusInProgress}}
	completion := &transactionalCompletionRepository{saveProgressError: errors.New("progress unavailable")}
	attempts := service.New(repository, completion)

	err := attempts.Finish(domain.Attempt{ID: 8, UserID: 3, Score: 72}, domain.Progress{LevelID: 4, UserRole: "seller"})
	if err == nil {
		t.Fatal("Finish() error = nil, want progress error")
	}
	if completion.committed {
		t.Fatal("transaction committed despite progress save error")
	}
}

func TestFinishRejectsAnAlreadyCompletedAttempt(t *testing.T) {
	repository := &attemptRepository{attempt: domain.Attempt{ID: 8, UserID: 3, Status: domain.AttemptStatusCompleted}}
	completion := &transactionalCompletionRepository{}
	attempts := service.New(repository, completion)

	err := attempts.Finish(domain.Attempt{ID: 8, Score: 72}, domain.Progress{LevelID: 4, UserRole: "seller"})
	if !errors.Is(err, apperrors.ErrInvalidAttemptStatusTransition) {
		t.Fatalf("Finish() error = %v, want invalid status transition", err)
	}
	if completion.committed {
		t.Fatal("transaction committed for an already completed attempt")
	}
}

type attemptRepository struct{ attempt domain.Attempt }

func (r *attemptRepository) Create(domain.Attempt) (domain.Attempt, error) {
	return domain.Attempt{}, nil
}
func (r *attemptRepository) GetByID(int) (domain.Attempt, error)        { return r.attempt, nil }
func (r *attemptRepository) Update(domain.Attempt) error                { return nil }
func (r *attemptRepository) Delete(int) error                           { return nil }
func (r *attemptRepository) ListByUserID(int) ([]domain.Attempt, error) { return nil, nil }

type transactionalCompletionRepository struct {
	attempt           domain.Attempt
	progress          domain.Progress
	committed         bool
	saveProgressError error
}

func (r *transactionalCompletionRepository) InTransaction(action func(service.CompletionStore) error) error {
	staged := transactionStore{saveProgressError: r.saveProgressError}
	if err := action(&staged); err != nil {
		return err
	}
	r.attempt = staged.attempt
	r.progress = staged.progress
	r.committed = true
	return nil
}

type transactionStore struct {
	attempt           domain.Attempt
	progress          domain.Progress
	saveProgressError error
}

func (store *transactionStore) UpdateAttempt(attempt domain.Attempt) error {
	store.attempt = attempt
	return nil
}

func (store *transactionStore) SaveProgress(progress domain.Progress) error {
	if store.saveProgressError != nil {
		return store.saveProgressError
	}
	store.progress = progress
	return nil
}
func (store *transactionStore) FinalizeLearning(*domain.AttemptResult) error { return nil }
