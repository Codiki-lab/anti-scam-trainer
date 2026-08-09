package repository

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/features/attempts/service"
	"encoding/json"
	"strconv"
	"time"

	"github.com/go-pg/pg"
)

type PostgresRepository struct{ db *pg.DB }

type attemptRecord struct {
	tableName         struct{}  `sql:"chat_sessions"`
	ID                int       `pg:"id,pk"`
	UserID            int       `pg:"user_id,notnull"`
	ChatID            int       `pg:"chat_id,notnull"`
	Status            string    `pg:"status,notnull"`
	StartedAt         time.Time `pg:"started_at"`
	FinishedAt        time.Time `pg:"finished_at"`
	Score             int       `pg:"score"`
	MaxScore          int       `pg:"max_score"`
	CurrentStepNumber int       `pg:"current_step_number"`
	Mode              string    `pg:"mode"`
	UserRole          string    `pg:"user_role"`
	IsScam            *bool     `pg:"is_scam"`
	FreeTextCount     int       `pg:"free_text_count"`
	FinalBreakdown    string    `pg:"final_breakdown"`
}

type gameScenarioRecord struct {
	ID             int    `pg:"id"`
	Title          string `pg:"title"`
	Description    string `pg:"description"`
	LevelID        int    `pg:"level_id"`
	LevelNumber    int    `pg:"level_number"`
	UserRole       string `pg:"user_role"`
	ContentStatus  string `pg:"content_status"`
	ScamScheme     string `pg:"scam_scheme"`
	ProductContext string `pg:"product_context"`
	AISystemPrompt string `pg:"ai_system_prompt"`
	FinalRubric    string `pg:"final_rubric"`
}

type gameStepRecord struct {
	ID              int    `pg:"id"`
	ChatID          int    `pg:"chat_id"`
	StepNumber      int    `pg:"step_number"`
	ResponseType    string `pg:"response_type"`
	StepGoal        string `pg:"step_goal"`
	MaxPoints       int    `pg:"max_points"`
	AIInstruction   string `pg:"ai_instruction"`
	FallbackMessage string `pg:"fallback_message"`
}

type gameOptionRecord struct {
	ID          int    `pg:"id"`
	StepID      int    `pg:"step_id"`
	OptionText  string `pg:"option_text"`
	Explanation string `pg:"explanation"`
	Points      int    `pg:"points"`
	SortOrder   int    `pg:"sort_order"`
}

type levelRecord struct {
	ID          int `pg:"id"`
	LevelNumber int `pg:"level_number"`
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

func (r *PostgresRepository) ListByUserID(userID int) ([]domain.Attempt, error) {
	var records []attemptRecord
	if err := r.db.Model(&records).Where("user_id = ?", userID).Select(); err != nil {
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

func (r *PostgresRepository) Levels(userID int, userRole string) ([]domain.Level, []domain.Progress, error) {
	var levels []levelRecord
	if _, err := r.db.Query(&levels, `SELECT id, level_number FROM levels WHERE is_active = TRUE ORDER BY level_number`); err != nil {
		return nil, nil, err
	}
	result := make([]domain.Level, len(levels))
	for i, row := range levels {
		result[i] = domain.Level{ID: row.ID, Number: row.LevelNumber}
	}
	var records []progressRecord
	if err := r.db.Model(&records).Where("user_id = ? AND user_role = ?", userID, userRole).Select(); err != nil && err != pg.ErrNoRows {
		return nil, nil, err
	}
	progress := make([]domain.Progress, len(records))
	for i, record := range records {
		progress[i] = domain.Progress{UserID: record.UserID, LevelID: record.LevelID, UserRole: record.UserRole, BestScore: record.BestScore, Stars: record.Stars, Attempts: record.Attempts, PassedAt: record.PassedAt}
	}
	return result, progress, nil
}

func (r *PostgresRepository) PublishedScenario(levelNumber int, userRole string) (domain.Scenario, error) {
	var record gameScenarioRecord
	query := `SELECT c.id, c.title, c.description, c.level_id, l.level_number, c.user_role, c.content_status, COALESCE(c.scam_scheme,'') AS scam_scheme, c.product_context::text AS product_context, COALESCE(c.ai_system_prompt,'') AS ai_system_prompt, c.final_rubric::text AS final_rubric FROM chats c JOIN levels l ON l.id = c.level_id WHERE c.content_status = 'published' AND c.archived_at IS NULL`
	args := []interface{}{}
	if levelNumber > 0 {
		query += ` AND l.level_number = ?`
		args = append(args, levelNumber)
	}
	if userRole != "" {
		query += ` AND c.user_role = ?`
		args = append(args, userRole)
	}
	if _, err := r.db.QueryOne(&record, query, args...); err != nil {
		return domain.Scenario{}, err
	}
	return scenarioFromGameRecord(record), nil
}

func (r *PostgresRepository) FreePlayConfig(userRole string) (domain.FreePlayConfig, error) {
	type row struct {
		UserRole       string `pg:"user_role"`
		ProductContext string `pg:"product_context"`
		SystemPrompt   string `pg:"system_prompt"`
		FinalRubric    string `pg:"final_rubric"`
	}
	var item row
	_, err := r.db.QueryOne(&item, `SELECT user_role, product_context::text AS product_context, system_prompt, final_rubric::text AS final_rubric FROM free_play_configs WHERE user_role = ?`, userRole)
	if err != nil {
		return domain.FreePlayConfig{}, err
	}
	return domain.FreePlayConfig{UserRole: item.UserRole, ProductContext: decodeJSONObject(item.ProductContext), SystemPrompt: item.SystemPrompt, FinalRubric: decodeJSONObject(item.FinalRubric)}, nil
}

func (r *PostgresRepository) Scenario(id int) (domain.Scenario, error) {
	var record gameScenarioRecord
	_, err := r.db.QueryOne(&record, `SELECT c.id, c.title, c.description, c.level_id, l.level_number, c.user_role, c.content_status, COALESCE(c.scam_scheme,'') AS scam_scheme, c.product_context::text AS product_context, COALESCE(c.ai_system_prompt,'') AS ai_system_prompt, c.final_rubric::text AS final_rubric FROM chats c JOIN levels l ON l.id = c.level_id WHERE c.id = ?`, id)
	if err != nil {
		return domain.Scenario{}, err
	}
	return scenarioFromGameRecord(record), nil
}

func (r *PostgresRepository) FindInProgress(userID, scenarioID int) (domain.Attempt, error) {
	var record attemptRecord
	if err := r.db.Model(&record).Where("user_id = ? AND chat_id = ? AND status = ?", userID, scenarioID, domain.AttemptStatusInProgress).Select(); err != nil {
		return domain.Attempt{}, err
	}
	return attemptFromRecord(record), nil
}

func (r *PostgresRepository) FindInProgressFreePlay(userID int, userRole string) (domain.Attempt, error) {
	var record attemptRecord
	if err := r.db.Model(&record).Where("user_id = ? AND mode = ? AND user_role = ? AND status = ?", userID, domain.AttemptModeFreePlay, userRole, domain.AttemptStatusInProgress).Select(); err != nil {
		return domain.Attempt{}, err
	}
	return attemptFromRecord(record), nil
}

func (r *PostgresRepository) CreateGameAttempt(attempt domain.Attempt) (domain.Attempt, error) {
	record := toAttemptRecord(attempt)
	record.MaxScore = 0
	if _, err := r.db.Model(&record).Insert(); err != nil {
		return domain.Attempt{}, err
	}
	return attemptFromRecord(record), nil
}

func (r *PostgresRepository) StartFreePlay(attempt domain.Attempt, message domain.DialogueMessage) (domain.Attempt, error) {
	var created domain.Attempt
	err := r.db.RunInTransaction(func(tx *pg.Tx) error {
		var record attemptRecord
		_, err := tx.QueryOne(&record, `INSERT INTO chat_sessions (user_id, chat_id, mode, user_role, is_scam, status, started_at, current_step_number, free_text_count) VALUES (?, NULL, 'free_play', ?, ?, ?, ?, 0, 0) RETURNING *`, attempt.UserID, attempt.UserRole, attempt.IsScam, attempt.Status, attempt.StartedAt)
		if err != nil {
			return err
		}
		created = attemptFromRecord(record)
		message.AttemptID = created.ID
		_, err = tx.Exec(`INSERT INTO messages (session_id, role, message, created_at) VALUES (?, ?, ?, ?)`, message.AttemptID, message.Role, message.Text, message.CreatedAt)
		return err
	})
	return created, err
}

func (r *PostgresRepository) GetGameAttempt(id int) (domain.Attempt, error) { return r.GetByID(id) }

func (r *PostgresRepository) Step(scenarioID, number int) (domain.ScenarioStep, error) {
	var step gameStepRecord
	if _, err := r.db.QueryOne(&step, `SELECT id, chat_id, step_number, response_type, step_goal, max_points, COALESCE(ai_instruction,'') AS ai_instruction, COALESCE(fallback_message,'') AS fallback_message FROM chat_steps WHERE chat_id = ? AND step_number = ?`, scenarioID, number); err != nil {
		return domain.ScenarioStep{}, err
	}
	var options []gameOptionRecord
	if _, err := r.db.Query(&options, `SELECT id, step_id, option_text, explanation, points, sort_order FROM chat_options WHERE step_id = ? ORDER BY sort_order`, step.ID); err != nil {
		return domain.ScenarioStep{}, err
	}
	result := domain.ScenarioStep{ID: step.ID, ScenarioID: step.ChatID, Number: step.StepNumber, ResponseType: domain.ResponseType(step.ResponseType), Goal: step.StepGoal, MaxPoints: step.MaxPoints, AIInstruction: step.AIInstruction, FallbackMessage: step.FallbackMessage, Options: make([]domain.ScenarioOption, len(options))}
	for i, option := range options {
		result.Options[i] = domain.ScenarioOption{ID: option.ID, StepID: option.StepID, Text: option.OptionText, Explanation: option.Explanation, Points: option.Points, SortOrder: option.SortOrder}
	}
	return result, nil
}

func (r *PostgresRepository) Answers(attemptID int) ([]domain.UserAnswer, error) {
	type row struct {
		SessionID     int    `pg:"session_id"`
		StepID        int    `pg:"step_id"`
		OptionID      int    `pg:"option_id"`
		FreeText      string `pg:"free_text"`
		AwardedPoints int    `pg:"awarded_points"`
		Explanation   string `pg:"explanation"`
		OptionText    string `pg:"option_text"`
		AIEvaluation  string `pg:"ai_evaluation"`
		TurnNumber    int    `pg:"turn_number"`
	}
	var rows []row
	if _, err := r.db.Query(&rows, `SELECT a.session_id, COALESCE(a.step_id,0) AS step_id, COALESCE(a.option_id,0) AS option_id, COALESCE(a.free_text,'') AS free_text, a.awarded_points, COALESCE(a.explanation,'') AS explanation, COALESCE(o.option_text,'') AS option_text, COALESCE(a.ai_evaluation::text,'') AS ai_evaluation, a.turn_number FROM session_answers a LEFT JOIN chat_options o ON o.id = a.option_id WHERE a.session_id = ? ORDER BY a.turn_number, a.created_at`, attemptID); err != nil {
		return nil, err
	}
	answers := make([]domain.UserAnswer, len(rows))
	for i, row := range rows {
		var optionID *int
		if row.OptionID != 0 {
			id := row.OptionID
			optionID = &id
		}
		var evaluation *domain.AIEvaluation
		if row.AIEvaluation != "" {
			var decoded domain.AIEvaluation
			if json.Unmarshal([]byte(row.AIEvaluation), &decoded) == nil {
				evaluation = &decoded
			}
		}
		answers[i] = domain.UserAnswer{AttemptID: row.SessionID, StepID: row.StepID, OptionID: optionID, FreeText: row.FreeText, AwardedPoints: row.AwardedPoints, Explanation: row.Explanation, OptionText: row.OptionText, Evaluation: evaluation, TurnNumber: row.TurnNumber}
	}
	return answers, nil
}

func (r *PostgresRepository) Messages(attemptID int) ([]domain.DialogueMessage, error) {
	type row struct {
		ID        int       `pg:"id"`
		SessionID int       `pg:"session_id"`
		Role      string    `pg:"role"`
		Message   string    `pg:"message"`
		CreatedAt time.Time `pg:"created_at"`
	}
	var rows []row
	if _, err := r.db.Query(&rows, `SELECT id, session_id, role, message, created_at FROM messages WHERE session_id = ? ORDER BY created_at, id`, attemptID); err != nil {
		return nil, err
	}
	result := make([]domain.DialogueMessage, len(rows))
	for i, item := range rows {
		result[i] = domain.DialogueMessage{ID: item.ID, AttemptID: item.SessionID, Role: domain.MessageRole(item.Role), Text: item.Message, CreatedAt: item.CreatedAt}
	}
	return result, nil
}

func (r *PostgresRepository) AwardedPoints(attemptID int) (int, error) {
	var total int
	_, err := r.db.QueryOne(pg.Scan(&total), `SELECT COALESCE(SUM(awarded_points), 0) FROM session_answers WHERE session_id = ?`, attemptID)
	return total, err
}
func (r *PostgresRepository) Advance(attemptID, nextStepNumber int) error {
	_, err := r.db.Model(&attemptRecord{}).Set("current_step_number = ?", nextStepNumber).Where("id = ?", attemptID).Update()
	return err
}
func (r *PostgresRepository) Abandon(attemptID int, finishedAt time.Time) error {
	_, err := r.db.Model(&attemptRecord{}).Set("status = ?", domain.AttemptStatusAbandoned).Set("finished_at = ?", finishedAt).Where("id = ?", attemptID).Update()
	return err
}
func (r *PostgresRepository) Complete(action func(service.GameCompletionStore) error) error {
	return r.db.RunInTransaction(func(tx *pg.Tx) error { return action(gameTransactionStore{db: tx}) })
}

type gameTransactionStore struct{ db *pg.Tx }

func (s gameTransactionStore) SaveAnswer(answer domain.UserAnswer) error {
	var evaluation string
	if answer.Evaluation != nil {
		if encoded, err := json.Marshal(answer.Evaluation); err == nil {
			evaluation = string(encoded)
		}
	}
	_, err := s.db.Exec(`INSERT INTO session_answers (session_id, step_id, option_id, free_text, awarded_points, ai_evaluation, explanation, turn_number) VALUES (?, NULLIF(?,0), ?, NULLIF(?,''), ?, NULLIF(?,''), ?, ?)`, answer.AttemptID, answer.StepID, answer.OptionID, answer.FreeText, answer.AwardedPoints, evaluation, answer.Explanation, answer.TurnNumber)
	return err
}
func (s gameTransactionStore) SaveMessage(message domain.DialogueMessage) error {
	_, err := s.db.Exec(`INSERT INTO messages (session_id, role, message, created_at) VALUES (?, ?, ?, ?)`, message.AttemptID, message.Role, message.Text, message.CreatedAt)
	return err
}
func (s gameTransactionStore) AdvanceAttempt(id, next int) error {
	_, err := s.db.Model(&attemptRecord{}).Set("current_step_number = ?", next).Where("id = ?", id).Update()
	return err
}
func (s gameTransactionStore) CompleteAttempt(attempt domain.Attempt) error {
	encodedBreakdown, err := json.Marshal(attempt.FinalBreakdown)
	if err != nil {
		return err
	}
	_, err = s.db.Model(&attemptRecord{}).Set("status = ?", attempt.Status).Set("finished_at = ?", attempt.FinishedAt).Set("score = ?", attempt.Score).Set("free_text_count = ?", attempt.FreeTextCount).Set("final_breakdown = ?::jsonb", string(encodedBreakdown)).Where("id = ?", attempt.ID).Update()
	return err
}
func (s gameTransactionStore) UpdateFreeTextCount(id, count int) error {
	_, err := s.db.Model(&attemptRecord{}).Set("free_text_count = ?", count).Where("id = ?", id).Update()
	return err
}
func (s gameTransactionStore) SaveProgress(progress domain.Progress) error {
	record := toProgressRecord(progress)
	_, err := s.db.Model(&record).OnConflict("(user_id, user_role, level_id) DO UPDATE").Set("best_score = GREATEST(progress_record.best_score, EXCLUDED.best_score)").Set("stars = GREATEST(progress_record.stars, EXCLUDED.stars)").Set("attempts = progress_record.attempts + 1").Set("passed_at = COALESCE(progress_record.passed_at, EXCLUDED.passed_at)").Insert()
	return err
}

func scenarioFromGameRecord(record gameScenarioRecord) domain.Scenario {
	return domain.Scenario{ID: record.ID, Title: record.Title, Description: record.Description, Level: strconv.Itoa(record.LevelNumber), LevelID: record.LevelID, UserRole: record.UserRole, Status: record.ContentStatus, ScamScheme: record.ScamScheme, ProductContext: decodeJSONObject(record.ProductContext), AISystemPrompt: record.AISystemPrompt, FinalRubric: decodeJSONObject(record.FinalRubric)}
}

func decodeJSONObject(value string) domain.JSONObject {
	result := domain.JSONObject{}
	_ = json.Unmarshal([]byte(value), &result)
	return result
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
		Set("best_score = GREATEST(progress_record.best_score, EXCLUDED.best_score)").
		Set("stars = GREATEST(progress_record.stars, EXCLUDED.stars)").
		Set("attempts = progress_record.attempts + 1").
		Set("passed_at = COALESCE(progress_record.passed_at, EXCLUDED.passed_at)").
		Insert()
	return err
}

func toAttemptRecord(attempt domain.Attempt) attemptRecord {
	encodedBreakdown, _ := json.Marshal(attempt.FinalBreakdown)
	return attemptRecord{ID: attempt.ID, UserID: attempt.UserID, ChatID: attempt.ScenarioID, Mode: string(attempt.Mode), UserRole: attempt.UserRole, IsScam: attempt.IsScam, Status: attempt.Status, StartedAt: attempt.StartedAt, FinishedAt: attempt.FinishedAt, Score: attempt.Score, MaxScore: attempt.MaxScore, CurrentStepNumber: attempt.CurrentStepNumber, FreeTextCount: attempt.FreeTextCount, FinalBreakdown: string(encodedBreakdown)}
}

func attemptFromRecord(record attemptRecord) domain.Attempt {
	var breakdown []domain.AnswerBreakdown
	if record.FinalBreakdown != "" {
		_ = json.Unmarshal([]byte(record.FinalBreakdown), &breakdown)
	}
	return domain.Attempt{ID: record.ID, UserID: record.UserID, ScenarioID: record.ChatID, Mode: domain.AttemptMode(record.Mode), UserRole: record.UserRole, IsScam: record.IsScam, Status: record.Status, StartedAt: record.StartedAt, FinishedAt: record.FinishedAt, Score: record.Score, MaxScore: record.MaxScore, CurrentStepNumber: record.CurrentStepNumber, FreeTextCount: record.FreeTextCount, FinalBreakdown: breakdown}
}

func toProgressRecord(progress domain.Progress) progressRecord {
	return progressRecord{UserID: progress.UserID, LevelID: progress.LevelID, UserRole: progress.UserRole, BestScore: progress.BestScore, Stars: progress.Stars, Attempts: progress.Attempts, PassedAt: progress.PassedAt}
}
