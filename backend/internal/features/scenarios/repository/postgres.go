package repository

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"encoding/json"

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
	_, err := r.db.QueryOne(pg.Scan(&id), `INSERT INTO chats (title, description, difficulty, role, is_active, level_id, user_role, content_status, scam_scheme, product_context, ai_system_prompt, final_rubric) VALUES (?, ?, ?, ?, false, ?, ?, 'draft', ?, ?::jsonb, ?, ?::jsonb) RETURNING id`, s.Title, s.Description, s.Level, s.UserRole, s.LevelID, s.UserRole, s.ScamScheme, encodeJSONObject(s.ProductContext), s.AISystemPrompt, encodeJSONObject(s.FinalRubric))
	s.ID = id
	s.Status = domain.ScenarioStatusDraft
	return s, err
}
func (r *PostgresRepository) ListContent() ([]domain.Scenario, error) {
	type row struct {
		ID             int    `pg:"id"`
		Title          string `pg:"title"`
		Description    string `pg:"description"`
		LevelID        int    `pg:"level_id"`
		UserRole       string `pg:"user_role"`
		Status         string `pg:"content_status"`
		ScamScheme     string `pg:"scam_scheme"`
		ProductContext string `pg:"product_context"`
		AISystemPrompt string `pg:"ai_system_prompt"`
		FinalRubric    string `pg:"final_rubric"`
	}
	var rows []row
	_, err := r.db.Query(&rows, `SELECT id,title,description,level_id,user_role,content_status,COALESCE(scam_scheme,'') AS scam_scheme,product_context::text AS product_context,COALESCE(ai_system_prompt,'') AS ai_system_prompt,final_rubric::text AS final_rubric FROM chats ORDER BY id`)
	result := make([]domain.Scenario, len(rows))
	for i, x := range rows {
		result[i] = domain.Scenario{ID: x.ID, Title: x.Title, Description: x.Description, LevelID: x.LevelID, UserRole: x.UserRole, Status: x.Status, ScamScheme: x.ScamScheme, ProductContext: decodeJSONObject(x.ProductContext), AISystemPrompt: x.AISystemPrompt, FinalRubric: decodeJSONObject(x.FinalRubric)}
	}
	return result, err
}
func (r *PostgresRepository) ContentScenario(id int) (domain.Scenario, error) {
	var s domain.Scenario
	type row struct {
		ID             int    `pg:"id"`
		Title          string `pg:"title"`
		Description    string `pg:"description"`
		LevelID        int    `pg:"level_id"`
		UserRole       string `pg:"user_role"`
		Status         string `pg:"content_status"`
		ScamScheme     string `pg:"scam_scheme"`
		ProductContext string `pg:"product_context"`
		AISystemPrompt string `pg:"ai_system_prompt"`
		FinalRubric    string `pg:"final_rubric"`
	}
	var x row
	_, err := r.db.QueryOne(&x, `SELECT id,title,description,level_id,user_role,content_status,COALESCE(scam_scheme,'') AS scam_scheme,product_context::text AS product_context,COALESCE(ai_system_prompt,'') AS ai_system_prompt,final_rubric::text AS final_rubric FROM chats WHERE id=?`, id)
	s = domain.Scenario{ID: x.ID, Title: x.Title, Description: x.Description, LevelID: x.LevelID, UserRole: x.UserRole, Status: x.Status, ScamScheme: x.ScamScheme, ProductContext: decodeJSONObject(x.ProductContext), AISystemPrompt: x.AISystemPrompt, FinalRubric: decodeJSONObject(x.FinalRubric)}
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
					(s.response_type IN ('multiple_choice','similar_choice') AND NOT EXISTS (SELECT 1 FROM chat_options o WHERE o.step_id = s.id))
					OR (EXISTS (SELECT 1 FROM chat_options o WHERE o.step_id = s.id) AND s.max_points <> (SELECT MAX(o.points) FROM chat_options o WHERE o.step_id = s.id))
					OR (s.response_type IN ('mixed','free_text') AND (NULLIF(s.ai_instruction,'') IS NULL OR NULLIF(s.fallback_message,'') IS NULL))
				  )
			)`, id, id)
	return valid, err
}
func (r *PostgresRepository) UpdateContent(s domain.Scenario) error {
	_, err := r.db.Exec(`UPDATE chats SET title=?, description=?, scam_scheme=?, product_context=?::jsonb, ai_system_prompt=?, final_rubric=?::jsonb WHERE id=?`, s.Title, s.Description, s.ScamScheme, encodeJSONObject(s.ProductContext), s.AISystemPrompt, encodeJSONObject(s.FinalRubric), s.ID)
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
	_, err := r.db.QueryOne(pg.Scan(&id), `INSERT INTO chat_steps (chat_id,step_number,response_type,step_goal,max_points,ai_instruction,fallback_message) VALUES (?,?,?,?,?,?,?) RETURNING id`, s.ScenarioID, s.Number, s.ResponseType, s.Goal, s.MaxPoints, s.AIInstruction, s.FallbackMessage)
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
	_, err := r.db.Exec(`UPDATE chat_steps SET step_number=?, response_type=?, step_goal=?, max_points=?, ai_instruction=?, fallback_message=? WHERE id=?`, s.Number, s.ResponseType, s.Goal, s.MaxPoints, s.AIInstruction, s.FallbackMessage, s.ID)
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

func encodeJSONObject(value domain.JSONObject) string {
	encoded, err := json.Marshal(value)
	if err != nil || value == nil {
		return "{}"
	}
	return string(encoded)
}

func decodeJSONObject(value string) domain.JSONObject {
	result := domain.JSONObject{}
	_ = json.Unmarshal([]byte(value), &result)
	return result
}

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
