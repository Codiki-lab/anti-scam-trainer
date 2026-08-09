package repository

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/features/attempts/service"
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
}

type gameScenarioRecord struct {
	ID            int    `pg:"id"`
	Title         string `pg:"title"`
	Description   string `pg:"description"`
	LevelID       int    `pg:"level_id"`
	LevelNumber   int    `pg:"level_number"`
	UserRole      string `pg:"user_role"`
	ContentStatus string `pg:"content_status"`
}

type gameStepRecord struct {
	ID           int    `pg:"id"`
	ChatID       int    `pg:"chat_id"`
	StepNumber   int    `pg:"step_number"`
	ResponseType string `pg:"response_type"`
	StepGoal     string `pg:"step_goal"`
	MaxPoints    int    `pg:"max_points"`
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
	if err := r.db.Model(&levels).TableExpr("levels").Where("level_number IN (1, 2)").Order("level_number").Select(); err != nil {
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
	query := `SELECT c.id, c.title, c.description, c.level_id, l.level_number, c.user_role, c.content_status FROM chats c JOIN levels l ON l.id = c.level_id WHERE c.content_status = 'published' AND c.archived_at IS NULL`
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

func (r *PostgresRepository) Scenario(id int) (domain.Scenario, error) {
	var record gameScenarioRecord
	_, err := r.db.QueryOne(&record, `SELECT c.id, c.title, c.description, c.level_id, l.level_number, c.user_role, c.content_status FROM chats c JOIN levels l ON l.id = c.level_id WHERE c.id = ?`, id)
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

func (r *PostgresRepository) CreateGameAttempt(attempt domain.Attempt) (domain.Attempt, error) {
	record := toAttemptRecord(attempt)
	record.MaxScore = 0
	if _, err := r.db.Model(&record).Insert(); err != nil {
		return domain.Attempt{}, err
	}
	return attemptFromRecord(record), nil
}

func (r *PostgresRepository) GetGameAttempt(id int) (domain.Attempt, error) { return r.GetByID(id) }

func (r *PostgresRepository) Step(scenarioID, number int) (domain.ScenarioStep, error) {
	var step gameStepRecord
	if _, err := r.db.QueryOne(&step, `SELECT id, chat_id, step_number, response_type, step_goal, max_points FROM chat_steps WHERE chat_id = ? AND step_number = ?`, scenarioID, number); err != nil {
		return domain.ScenarioStep{}, err
	}
	var options []gameOptionRecord
	if _, err := r.db.Query(&options, `SELECT id, step_id, option_text, explanation, points, sort_order FROM chat_options WHERE step_id = ? ORDER BY sort_order`, step.ID); err != nil {
		return domain.ScenarioStep{}, err
	}
	result := domain.ScenarioStep{ID: step.ID, ScenarioID: step.ChatID, Number: step.StepNumber, ResponseType: step.ResponseType, Goal: step.StepGoal, MaxPoints: step.MaxPoints, Options: make([]domain.ScenarioOption, len(options))}
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
	}
	var rows []row
	if _, err := r.db.Query(&rows, `SELECT a.session_id, a.step_id, a.option_id, a.free_text, a.awarded_points, a.explanation, o.option_text FROM session_answers a LEFT JOIN chat_options o ON o.id = a.option_id WHERE a.session_id = ? ORDER BY a.created_at`, attemptID); err != nil {
		return nil, err
	}
	answers := make([]domain.UserAnswer, len(rows))
	for i, row := range rows {
		id := row.OptionID
		answers[i] = domain.UserAnswer{AttemptID: row.SessionID, StepID: row.StepID, OptionID: &id, FreeText: row.FreeText, AwardedPoints: row.AwardedPoints, Explanation: row.Explanation, OptionText: row.OptionText}
	}
	return answers, nil
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

func (s gameTransactionStore) SaveAnswer(answer domain.UserAnswer, points int, explanation string) error {
	_, err := s.db.Exec(`INSERT INTO session_answers (session_id, step_id, option_id, awarded_points, explanation) VALUES (?, ?, ?, ?, ?)`, answer.AttemptID, answer.StepID, answer.OptionID, points, explanation)
	return err
}
func (s gameTransactionStore) AdvanceAttempt(id, next int) error {
	_, err := s.db.Model(&attemptRecord{}).Set("current_step_number = ?", next).Where("id = ?", id).Update()
	return err
}
func (s gameTransactionStore) CompleteAttempt(attempt domain.Attempt) error {
	_, err := s.db.Model(&attemptRecord{}).Set("status = ?", attempt.Status).Set("finished_at = ?", attempt.FinishedAt).Set("score = ?", attempt.Score).Where("id = ?", attempt.ID).Update()
	return err
}
func (s gameTransactionStore) SaveProgress(progress domain.Progress) error {
	record := toProgressRecord(progress)
	_, err := s.db.Model(&record).OnConflict("(user_id, user_role, level_id) DO UPDATE").Set("best_score = GREATEST(user_level_progress.best_score, EXCLUDED.best_score)").Set("stars = GREATEST(user_level_progress.stars, EXCLUDED.stars)").Set("attempts = user_level_progress.attempts + 1").Set("passed_at = COALESCE(user_level_progress.passed_at, EXCLUDED.passed_at)").Insert()
	return err
}

func scenarioFromGameRecord(record gameScenarioRecord) domain.Scenario {
	return domain.Scenario{ID: record.ID, Title: record.Title, Description: record.Description, Level: strconv.Itoa(record.LevelNumber), LevelID: record.LevelID, UserRole: record.UserRole, Status: record.ContentStatus}
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
	return attemptRecord{ID: attempt.ID, UserID: attempt.UserID, ChatID: attempt.ScenarioID, Status: attempt.Status, StartedAt: attempt.StartedAt, FinishedAt: attempt.FinishedAt, Score: attempt.Score, MaxScore: attempt.MaxScore, CurrentStepNumber: attempt.CurrentStepNumber}
}

func attemptFromRecord(record attemptRecord) domain.Attempt {
	return domain.Attempt{ID: record.ID, UserID: record.UserID, ScenarioID: record.ChatID, Status: record.Status, StartedAt: record.StartedAt, FinishedAt: record.FinishedAt, Score: record.Score, MaxScore: record.MaxScore, CurrentStepNumber: record.CurrentStepNumber}
}

func toProgressRecord(progress domain.Progress) progressRecord {
	return progressRecord{UserID: progress.UserID, LevelID: progress.LevelID, UserRole: progress.UserRole, BestScore: progress.BestScore, Stars: progress.Stars, Attempts: progress.Attempts, PassedAt: progress.PassedAt}
}
