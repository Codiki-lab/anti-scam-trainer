package repository

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/features/attempts/service"
	"time"

	"github.com/go-pg/pg"
)

type PostgresRepository struct{ db *pg.DB }

type attemptRecord struct {
	tableName  struct{}  `sql:"chat_sessions"`
	ID         int       `pg:"id,pk"`
	UserID     int       `pg:"user_id,notnull"`
	ChatID     int       `pg:"chat_id,notnull"`
	Status     string    `pg:"status,notnull"`
	StartedAt  time.Time `pg:"started_at"`
	FinishedAt time.Time `pg:"finished_at"`
	Score      int       `pg:"score"`
}

type progressRecord struct {
	tableName struct{}  `sql:"user_level_progress"`
	ID        int       `pg:"id,pk"`
	UserID    int       `pg:"user_id"`
	LevelID   int       `pg:"level_id"`
	UserRole  string    `pg:"user_role"`
	BestScore int       `pg:"best_score"`
	Stars     int       `pg:"stars"`
	Attempts  int       `pg:"attempts"`
	PassedAt  time.Time `pg:"passed_at"`
}

func NewPostgres(db *pg.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) Create(attempt domain.Attempt) (domain.Attempt, error) {
	record := toAttemptRecord(attempt)
	if _, err := r.db.Model(&record).Insert(); err != nil {
		return domain.Attempt{}, err
	}
	return attemptFromRecord(record), nil
}

func (r *PostgresRepository) GetByID(id int) (domain.Attempt, error) {
	var record attemptRecord
	if err := r.db.Model(&record).Where("id = ?", id).Select(); err != nil {
		return domain.Attempt{}, err
	}
	return attemptFromRecord(record), nil
}

func (r *PostgresRepository) Update(attempt domain.Attempt) error {
	record := toAttemptRecord(attempt)
	_, err := r.db.Model(&record).Column("user_id", "chat_id", "status", "started_at", "finished_at", "score").WherePK().Update()
	return err
}

func (r *PostgresRepository) Delete(id int) error {
	_, err := r.db.Model(&attemptRecord{}).Where("id = ?", id).Delete()
	return err
}

func (r *PostgresRepository) List() ([]domain.Attempt, error) {
	var records []attemptRecord
	if err := r.db.Model(&records).Select(); err != nil {
		return nil, err
	}
	attempts := make([]domain.Attempt, len(records))
	for index, record := range records {
		attempts[index] = attemptFromRecord(record)
	}
	return attempts, nil
}

func (r *PostgresRepository) InTransaction(action func(service.CompletionStore) error) error {
	return r.db.RunInTransaction(func(tx *pg.Tx) error {
		return action(transactionStore{db: tx})
	})
}

type transactionStore struct{ db *pg.Tx }

func (store transactionStore) UpdateAttempt(attempt domain.Attempt) error {
	record := toAttemptRecord(attempt)
	_, err := store.db.Model(&record).Column("status", "finished_at", "score").WherePK().Update()
	return err
}

func (store transactionStore) SaveProgress(progress domain.Progress) error {
	record := toProgressRecord(progress)
	_, err := store.db.Model(&record).
		OnConflict("(user_id, user_role, level_id) DO UPDATE").
		Set("best_score = GREATEST(user_level_progress.best_score, EXCLUDED.best_score)").
		Set("stars = GREATEST(user_level_progress.stars, EXCLUDED.stars)").
		Set("attempts = user_level_progress.attempts + 1").
		Set("passed_at = COALESCE(user_level_progress.passed_at, EXCLUDED.passed_at)").
		Insert()
	return err
}

func toAttemptRecord(attempt domain.Attempt) attemptRecord {
	return attemptRecord{ID: attempt.ID, UserID: attempt.UserID, ChatID: attempt.ScenarioID, Status: attempt.Status, StartedAt: attempt.StartedAt, FinishedAt: attempt.FinishedAt, Score: attempt.Score}
}

func attemptFromRecord(record attemptRecord) domain.Attempt {
	return domain.Attempt{ID: record.ID, UserID: record.UserID, ScenarioID: record.ChatID, Status: record.Status, StartedAt: record.StartedAt, FinishedAt: record.FinishedAt, Score: record.Score}
}

func toProgressRecord(progress domain.Progress) progressRecord {
	return progressRecord{UserID: progress.UserID, LevelID: progress.LevelID, UserRole: progress.UserRole, BestScore: progress.BestScore, Stars: progress.Stars, Attempts: progress.Attempts, PassedAt: progress.PassedAt}
}
