package repository

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"github.com/go-pg/pg"
)

type PostgresRepository struct{ db *pg.DB }
type chatRecord struct {
	ID          int    `pg:"id,pk"`
	Title       string `pg:"title,notnull"`
	Description string `pg:"description"`
	Difficulty  string `pg:"difficulty,notnull"`
	Role        string `pg:"role,notnull"`
	IsActive    bool   `pg:"is_active"`
}

func NewPostgres(db *pg.DB) *PostgresRepository { return &PostgresRepository{db: db} }
func (r *PostgresRepository) Create(scenario domain.Scenario) (domain.Scenario, error) {
	record := toRecord(scenario)
	if _, err := r.db.Model(&record).Insert(); err != nil {
		return domain.Scenario{}, err
	}
	return toDomain(record), nil
}
func (r *PostgresRepository) GetByID(id int) (domain.Scenario, error) {
	var record chatRecord
	if err := r.db.Model(&record).Where("id = ?", id).Select(); err != nil {
		return domain.Scenario{}, err
	}
	return toDomain(record), nil
}
func (r *PostgresRepository) Update(scenario domain.Scenario) error {
	record := toRecord(scenario)
	_, err := r.db.Model(&record).Column("title", "description", "difficulty", "role", "is_active").WherePK().Update()
	return err
}
func (r *PostgresRepository) Delete(id int) error {
	_, err := r.db.Model(&chatRecord{}).Where("id = ?", id).Delete()
	return err
}
func (r *PostgresRepository) List() ([]domain.Scenario, error) {
	var records []chatRecord
	if err := r.db.Model(&records).Select(); err != nil {
		return nil, err
	}
	scenarios := make([]domain.Scenario, len(records))
	for i, record := range records {
		scenarios[i] = toDomain(record)
	}
	return scenarios, nil
}
func toRecord(scenario domain.Scenario) chatRecord {
	return chatRecord{ID: scenario.ID, Title: scenario.Title, Description: scenario.Description, Difficulty: scenario.Level, Role: scenario.UserRole, IsActive: scenario.IsActive}
}
func toDomain(record chatRecord) domain.Scenario {
	return domain.Scenario{ID: record.ID, Title: record.Title, Description: record.Description, Level: record.Difficulty, UserRole: record.Role, IsActive: record.IsActive}
}
