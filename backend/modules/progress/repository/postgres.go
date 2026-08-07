package repository

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"time"

	"github.com/go-pg/pg"
)

type PostgresRepository struct{ db *pg.DB }

type progressRecord struct {
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

func (r *PostgresRepository) Get(userID, levelID int, userRole string) (domain.Progress, error) {
	var record progressRecord
	if err := r.db.Model(&record).Where("user_id = ? AND level_id = ? AND user_role = ?", userID, levelID, userRole).Select(); err != nil {
		return domain.Progress{}, err
	}
	return toDomain(record), nil
}

func (r *PostgresRepository) Save(progress domain.Progress) error {
	record := toRecord(progress)
	_, err := r.db.Model(&record).
		OnConflict("(user_id, user_role, level_id) DO UPDATE").
		Set("best_score = EXCLUDED.best_score").
		Set("stars = EXCLUDED.stars").
		Set("attempts = EXCLUDED.attempts").
		Set("passed_at = EXCLUDED.passed_at").
		Insert()
	return err
}

func toRecord(progress domain.Progress) progressRecord {
	return progressRecord{UserID: progress.UserID, LevelID: progress.LevelID, UserRole: progress.UserRole, BestScore: progress.BestScore, Stars: progress.Stars, Attempts: progress.Attempts, PassedAt: progress.PassedAt}
}

func toDomain(record progressRecord) domain.Progress {
	return domain.Progress{UserID: record.UserID, LevelID: record.LevelID, UserRole: record.UserRole, BestScore: record.BestScore, Stars: record.Stars, Attempts: record.Attempts, PassedAt: record.PassedAt}
}
