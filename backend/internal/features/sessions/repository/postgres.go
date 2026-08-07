package repository

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"github.com/go-pg/pg"
	"time"
)

type PostgresRepository struct{ db *pg.DB }
type sessionRecord struct {
	ID         int       `pg:"id,pk"`
	UserID     int       `pg:"user_id,notnull"`
	ChatID     int       `pg:"chat_id,notnull"`
	Status     string    `pg:"status,notnull"`
	StartedAt  time.Time `pg:"started_at"`
	FinishedAt time.Time `pg:"finished_at"`
	Score      int       `pg:"score"`
}

func NewPostgres(db *pg.DB) *PostgresRepository { return &PostgresRepository{db: db} }
func (r *PostgresRepository) Create(attempt domain.Attempt) (domain.Attempt, error) {
	record := toRecord(attempt)
	if _, err := r.db.Model(&record).Insert(); err != nil {
		return domain.Attempt{}, err
	}
	return toDomain(record), nil
}
func (r *PostgresRepository) GetByID(id int) (domain.Attempt, error) {
	var record sessionRecord
	if err := r.db.Model(&record).Where("id = ?", id).Select(); err != nil {
		return domain.Attempt{}, err
	}
	return toDomain(record), nil
}
func (r *PostgresRepository) Update(attempt domain.Attempt) error {
	record := toRecord(attempt)
	_, err := r.db.Model(&record).Column("user_id", "chat_id", "status", "started_at", "finished_at", "score").WherePK().Update()
	return err
}
func (r *PostgresRepository) Delete(id int) error {
	_, err := r.db.Model(&sessionRecord{}).Where("id = ?", id).Delete()
	return err
}
func (r *PostgresRepository) List() ([]domain.Attempt, error) {
	var records []sessionRecord
	if err := r.db.Model(&records).Select(); err != nil {
		return nil, err
	}
	attempts := make([]domain.Attempt, len(records))
	for i, record := range records {
		attempts[i] = toDomain(record)
	}
	return attempts, nil
}
func toRecord(attempt domain.Attempt) sessionRecord {
	return sessionRecord{ID: attempt.ID, UserID: attempt.UserID, ChatID: attempt.ChatID, Status: attempt.Status, StartedAt: attempt.StartedAt, FinishedAt: attempt.FinishedAt, Score: attempt.Score}
}
func toDomain(record sessionRecord) domain.Attempt {
	return domain.Attempt{ID: record.ID, UserID: record.UserID, ChatID: record.ChatID, Status: record.Status, StartedAt: record.StartedAt, FinishedAt: record.FinishedAt, Score: record.Score}
}
