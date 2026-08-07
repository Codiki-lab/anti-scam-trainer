package repository

import (
	"anti-scam-trainer/backend/internal/core/domain"

	"github.com/go-pg/pg"
)

type PostgresRepository struct{ db *pg.DB }

type userRecord struct {
	ID             int    `pg:"id,pk"`
	ExternalID     string `pg:"user_id,notnull"`
	Username       string `pg:"username,notnull"`
	CompletedChats int    `pg:"completed_chats"`
}

func NewPostgres(db *pg.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) Create(user domain.User) (domain.User, error) {
	record := toRecord(user)
	if _, err := r.db.Model(&record).Insert(); err != nil {
		return domain.User{}, err
	}
	return toDomain(record), nil
}
func (r *PostgresRepository) GetByID(id int) (domain.User, error) {
	var record userRecord
	if err := r.db.Model(&record).Where("id = ?", id).Select(); err != nil {
		return domain.User{}, err
	}
	return toDomain(record), nil
}
func (r *PostgresRepository) GetByExternalID(id string) (domain.User, error) {
	var record userRecord
	if err := r.db.Model(&record).Where("user_id = ?", id).Select(); err != nil {
		return domain.User{}, err
	}
	return toDomain(record), nil
}
func (r *PostgresRepository) Update(user domain.User) error {
	record := toRecord(user)
	_, err := r.db.Model(&record).Column("user_id", "username", "completed_chats").WherePK().Update()
	return err
}
func (r *PostgresRepository) Delete(id int) error {
	_, err := r.db.Model(&userRecord{}).Where("id = ?", id).Delete()
	return err
}
func (r *PostgresRepository) List() ([]domain.User, error) {
	var records []userRecord
	if err := r.db.Model(&records).Select(); err != nil {
		return nil, err
	}
	users := make([]domain.User, len(records))
	for i, record := range records {
		users[i] = toDomain(record)
	}
	return users, nil
}
func toRecord(user domain.User) userRecord {
	return userRecord{ID: user.ID, ExternalID: user.ExternalID, Username: user.Username, CompletedChats: user.CompletedChats}
}
func toDomain(record userRecord) domain.User {
	return domain.User{ID: record.ID, ExternalID: record.ExternalID, Username: record.Username, CompletedChats: record.CompletedChats}
}
