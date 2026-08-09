package repository

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"time"

	"github.com/go-pg/pg"
)

type dailyTaskRow struct {
	ActivityDate string
	UserRole     string
	ActionType   string `pg:"action_type"`
	TopicID      int    `pg:"topic_id"`
	LevelNumber  int
	AttemptID    int        `pg:"attempt_id"`
	CompletedAt  *time.Time `pg:"completed_at"`
}

func (r *PostgresRepository) DailyTask(userID int, role domain.UserRole, date time.Time, recommendation domain.ContinueAction) (domain.DailyTask, error) {
	dateText := date.Format("2006-01-02")
	var row dailyTaskRow
	err := r.db.RunInTransaction(func(tx *pg.Tx) error {
		_, err := tx.QueryOne(&row, `SELECT activity_date::text,user_role,action_type,COALESCE(topic_id,0) topic_id,COALESCE(level_number,0) level_number,COALESCE(attempt_id,0) attempt_id,completed_at FROM daily_tasks WHERE user_id=? AND activity_date=?::date AND user_role=? FOR UPDATE`, userID, dateText, role)
		if err == pg.ErrNoRows {
			if _, err = tx.Exec(`INSERT INTO daily_tasks(user_id,activity_date,user_role,action_type,topic_id,level_number,attempt_id) VALUES(?,?::date,?,?,?,?,?) ON CONFLICT(user_id,activity_date,user_role) DO NOTHING`, userID, dateText, role, recommendation.Type, nullableInt(recommendation.TopicID), nullableInt(recommendation.Level), nullableInt(recommendation.AttemptID)); err != nil {
				return err
			}
			_, err = tx.QueryOne(&row, `SELECT activity_date::text,user_role,action_type,COALESCE(topic_id,0) topic_id,COALESCE(level_number,0) level_number,COALESCE(attempt_id,0) attempt_id,completed_at FROM daily_tasks WHERE user_id=? AND activity_date=?::date AND user_role=? FOR UPDATE`, userID, dateText, role)
			return err
		}
		if err != nil {
			return err
		}
		if row.CompletedAt == nil {
			valid, validErr := validDailyTarget(tx, row)
			if validErr != nil {
				return validErr
			}
			if !valid {
				_, err = tx.QueryOne(&row, `UPDATE daily_tasks SET action_type=?,topic_id=?,level_number=?,attempt_id=? WHERE user_id=? AND activity_date=?::date AND user_role=? RETURNING activity_date::text,user_role,action_type,COALESCE(topic_id,0) topic_id,COALESCE(level_number,0) level_number,COALESCE(attempt_id,0) attempt_id,completed_at`, recommendation.Type, nullableInt(recommendation.TopicID), nullableInt(recommendation.Level), nullableInt(recommendation.AttemptID), userID, dateText, role)
			}
		}
		return err
	})
	if err != nil {
		return domain.DailyTask{}, err
	}
	action := domain.ContinueAction{Type: row.ActionType, TopicID: row.TopicID, Level: row.LevelNumber, AttemptID: row.AttemptID}
	task := domain.DailyTask{Date: row.ActivityDate, Role: domain.UserRole(row.UserRole), Action: action, Completed: row.CompletedAt != nil}
	if task.Completed {
		task.CompletedAt = row.CompletedAt
	}
	return task, nil
}

func nullableInt(value int) any {
	if value == 0 {
		return nil
	}
	return value
}
func validDailyTarget(tx *pg.Tx, row dailyTaskRow) (bool, error) {
	var valid bool
	var err error
	switch row.ActionType {
	case "resume_attempt":
		_, err = tx.QueryOne(pg.Scan(&valid), `SELECT EXISTS(SELECT 1 FROM chat_sessions s JOIN chats c ON c.id=s.chat_id JOIN topics t ON t.id=c.topic_id WHERE s.id=? AND s.status='IN_PROGRESS' AND c.content_status='published' AND c.archived_at IS NULL AND t.content_status='published')`, row.AttemptID)
	case "read_theory", "take_quiz", "start_level":
		_, err = tx.QueryOne(pg.Scan(&valid), `SELECT EXISTS(SELECT 1 FROM topics WHERE id=? AND content_status='published')`, row.TopicID)
	case "start_free_play":
		valid = true
	}
	return valid, err
}
