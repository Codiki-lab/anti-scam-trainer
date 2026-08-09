package repository

import (
	"anti-scam-trainer/backend/internal/core/domain"

	"github.com/go-pg/pg"
	"time"
)

type PostgresRepository struct{ db *pg.DB }

type scenarioRecord struct {
	tableName   struct{} `sql:"chats"`
	ID          int      `pg:"id,pk"`
	Title       string   `pg:"title,notnull"`
	Description string   `pg:"description"`
	Difficulty  string   `pg:"difficulty,notnull"`
	Role        string   `pg:"role,notnull"`
	IsActive    bool     `pg:"is_active"`
}

func (r *PostgresRepository) CreateContent(s domain.Scenario) (domain.Scenario, error) {
	var id int
	_, err := r.db.QueryOne(pg.Scan(&id), `INSERT INTO chats (title, description, difficulty, role, is_active, level_id, user_role, content_status) VALUES (?, ?, ?, ?, false, ?, ?, 'draft') RETURNING id`, s.Title, s.Description, s.Level, s.UserRole, s.LevelID, s.UserRole)
	s.ID = id
	s.Status = domain.ScenarioStatusDraft
	return s, err
}
func (r *PostgresRepository) ListContent() ([]domain.Scenario, error) {
	type row struct {
		ID          int    `pg:"id"`
		Title       string `pg:"title"`
		Description string `pg:"description"`
		LevelID     int    `pg:"level_id"`
		UserRole    string `pg:"user_role"`
		Status      string `pg:"content_status"`
	}
	var rows []row
	_, err := r.db.Query(&rows, `SELECT id,title,description,level_id,user_role,content_status FROM chats ORDER BY id`)
	result := make([]domain.Scenario, len(rows))
	for i, x := range rows {
		result[i] = domain.Scenario{ID: x.ID, Title: x.Title, Description: x.Description, LevelID: x.LevelID, UserRole: x.UserRole, Status: x.Status}
	}
	return result, err
}
func (r *PostgresRepository) ContentScenario(id int) (domain.Scenario, error) {
	var s domain.Scenario
	type row struct {
		ID          int    `pg:"id"`
		Title       string `pg:"title"`
		Description string `pg:"description"`
		LevelID     int    `pg:"level_id"`
		UserRole    string `pg:"user_role"`
		Status      string `pg:"content_status"`
	}
	var x row
	_, err := r.db.QueryOne(&x, `SELECT id,title,description,level_id,user_role,content_status FROM chats WHERE id=?`, id)
	s = domain.Scenario{ID: x.ID, Title: x.Title, Description: x.Description, LevelID: x.LevelID, UserRole: x.UserRole, Status: x.Status}
	return s, err
}

func (r *PostgresRepository) ValidContent(id int) (bool, error) {
	var valid bool
	_, err := r.db.QueryOne(pg.Scan(&valid), `
		SELECT EXISTS (SELECT 1 FROM chat_steps s WHERE s.chat_id = ?)
		   AND NOT EXISTS (
				SELECT 1 FROM chat_steps s
				WHERE s.chat_id = ?
				  AND (
					NOT EXISTS (SELECT 1 FROM chat_options o WHERE o.step_id = s.id)
					OR s.max_points <> (SELECT MAX(o.points) FROM chat_options o WHERE o.step_id = s.id)
				  )
			)`, id, id)
	return valid, err
}
func (r *PostgresRepository) UpdateContent(s domain.Scenario) error {
	_, err := r.db.Exec(`UPDATE chats SET title=?, description=? WHERE id=?`, s.Title, s.Description, s.ID)
	return err
}
func (r *PostgresRepository) SetContentStatus(id int, status string, archived bool) error {
	var at interface{}
	if archived {
		at = time.Now().UTC()
	}
	_, err := r.db.Exec(`UPDATE chats SET content_status=?, archived_at=?, is_active=? WHERE id=?`, status, at, status == domain.ScenarioStatusPublished, id)
	return err
}
func (r *PostgresRepository) CreateStep(s domain.ScenarioStep) (domain.ScenarioStep, error) {
	var id int
	_, err := r.db.QueryOne(pg.Scan(&id), `INSERT INTO chat_steps (chat_id,step_number,response_type,step_goal,max_points) VALUES (?,?,?,?,?) RETURNING id`, s.ScenarioID, s.Number, s.ResponseType, s.Goal, s.MaxPoints)
	s.ID = id
	return s, err
}
func (r *PostgresRepository) CreateOption(o domain.ScenarioOption) (domain.ScenarioOption, error) {
	var id int
	_, err := r.db.QueryOne(pg.Scan(&id), `INSERT INTO chat_options (step_id,option_text,explanation,points,sort_order) VALUES (?,?,?,?,?) RETURNING id`, o.StepID, o.Text, o.Explanation, o.Points, o.SortOrder)
	o.ID = id
	return o, err
}

func (r *PostgresRepository) StepScenario(stepID int) (domain.Scenario, error) {
	var scenario domain.Scenario
	_, err := r.db.QueryOne(&scenario, `SELECT c.id, c.content_status AS status FROM chat_steps s JOIN chats c ON c.id = s.chat_id WHERE s.id = ?`, stepID)
	return scenario, err
}

func (r *PostgresRepository) OptionScenario(optionID int) (domain.Scenario, error) {
	var scenario domain.Scenario
	_, err := r.db.QueryOne(&scenario, `SELECT c.id, c.content_status AS status FROM chat_options o JOIN chat_steps s ON s.id = o.step_id JOIN chats c ON c.id = s.chat_id WHERE o.id = ?`, optionID)
	return scenario, err
}

func (r *PostgresRepository) UpdateStep(s domain.ScenarioStep) error {
	_, err := r.db.Exec(`UPDATE chat_steps SET step_number=?, response_type=?, step_goal=?, max_points=? WHERE id=?`, s.Number, s.ResponseType, s.Goal, s.MaxPoints, s.ID)
	return err
}

func (r *PostgresRepository) DeleteStep(id int) error {
	_, err := r.db.Exec(`DELETE FROM chat_steps WHERE id=?`, id)
	return err
}

func (r *PostgresRepository) UpdateOption(o domain.ScenarioOption) error {
	_, err := r.db.Exec(`UPDATE chat_options SET option_text=?, explanation=?, points=?, sort_order=? WHERE id=?`, o.Text, o.Explanation, o.Points, o.SortOrder, o.ID)
	return err
}

func (r *PostgresRepository) DeleteOption(id int) error {
	_, err := r.db.Exec(`DELETE FROM chat_options WHERE id=?`, id)
	return err
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
	var record scenarioRecord
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
	_, err := r.db.Model(&scenarioRecord{}).Where("id = ?", id).Delete()
	return err
}

func (r *PostgresRepository) List() ([]domain.Scenario, error) {
	var records []scenarioRecord
	if err := r.db.Model(&records).Select(); err != nil {
		return nil, err
	}
	scenarios := make([]domain.Scenario, len(records))
	for index, record := range records {
		scenarios[index] = toDomain(record)
	}
	return scenarios, nil
}

func toRecord(scenario domain.Scenario) scenarioRecord {
	return scenarioRecord{ID: scenario.ID, Title: scenario.Title, Description: scenario.Description, Difficulty: scenario.Level, Role: scenario.UserRole, IsActive: scenario.IsActive}
}

func toDomain(record scenarioRecord) domain.Scenario {
	return domain.Scenario{ID: record.ID, Title: record.Title, Description: record.Description, Level: record.Difficulty, UserRole: record.Role, IsActive: record.IsActive}
}
