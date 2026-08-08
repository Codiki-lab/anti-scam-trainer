package repository

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"

	"github.com/go-pg/pg"
)

type PostgresRepository struct{ db *pg.DB }

type userRecord struct {
	tableName    struct{} `sql:"users"`
	ID           int      `pg:"id,pk"`
	Username     string   `pg:"username,notnull"`
	PasswordHash string   `pg:"password_hash,notnull"`
	AccessRole   string   `pg:"access_role,notnull"`
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
		if err == pg.ErrNoRows {
			return domain.User{}, apperrors.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return toDomain(record), nil
}

func (r *PostgresRepository) GetByUsername(username string) (domain.User, error) {
	var record userRecord
	if err := r.db.Model(&record).Where("username = ?", username).Select(); err != nil {
		if err == pg.ErrNoRows {
			return domain.User{}, apperrors.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return toDomain(record), nil
}

func toRecord(user domain.User) userRecord {
	return userRecord{ID: user.ID, Username: user.Username, PasswordHash: user.PasswordHash, AccessRole: string(user.AccessRole)}
}

func toDomain(record userRecord) domain.User {
	return domain.User{ID: record.ID, Username: record.Username, PasswordHash: record.PasswordHash, AccessRole: domain.AccessRole(record.AccessRole)}
}
