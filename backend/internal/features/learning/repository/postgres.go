package repository

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"anti-scam-trainer/backend/internal/features/learning/service"
	"encoding/json"
	"time"

	"github.com/go-pg/pg"
)

type PostgresRepository struct{ db *pg.DB }

func NewPostgres(db *pg.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) FindRecommendation(userID int, date time.Time, role domain.UserRole) (domain.ContinueAction, bool, error) {
	var encoded string
	_, err := r.db.QueryOne(pg.Scan(&encoded), `SELECT action::text FROM dashboard_recommendations WHERE user_id=? AND activity_date=?::date AND user_role=?`, userID, date.Format("2006-01-02"), role)
	if err == pg.ErrNoRows {
		return domain.ContinueAction{}, false, nil
	}
	if err != nil {
		return domain.ContinueAction{}, false, err
	}
	var action domain.ContinueAction
	if err := json.Unmarshal([]byte(encoded), &action); err != nil {
		return domain.ContinueAction{}, false, err
	}
	return action, true, nil
}

func (r *PostgresRepository) SaveRecommendation(userID int, date time.Time, role domain.UserRole, action domain.ContinueAction) error {
	encoded, err := json.Marshal(action)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`INSERT INTO dashboard_recommendations(user_id,activity_date,user_role,action) VALUES(?,?::date,?,?::jsonb) ON CONFLICT DO NOTHING`, userID, date.Format("2006-01-02"), role, string(encoded))
	return err
}

type topicRow struct {
	ID            int    `pg:"id"`
	Slug          string `pg:"slug"`
	UserRole      string `pg:"user_role"`
	Title         string `pg:"title"`
	Description   string `pg:"description"`
	SortOrder     int    `pg:"sort_order"`
	TheoryRead    bool   `pg:"theory_read"`
	QuizPassed    bool   `pg:"quiz_passed"`
	QuizScore     int    `pg:"quiz_score"`
	Completed     bool   `pg:"completed"`
	ContentStatus string
	ArchivedAt    time.Time `pg:"archived_at"`
}

func (r *PostgresRepository) Topics(userID int, role domain.UserRole) ([]domain.Topic, error) {
	var rows []topicRow
	_, err := r.db.Query(&rows, `SELECT t.id,t.slug,t.user_role,t.title,t.description,t.sort_order,
        (p.theory_read_at IS NOT NULL) theory_read,COALESCE(p.quiz_passed,FALSE) quiz_passed,COALESCE(p.quiz_best_score,0) quiz_score,
        (p.completed_at IS NOT NULL) completed
        FROM topics t LEFT JOIN user_topic_progress p ON p.topic_id=t.id AND p.user_id=?
		WHERE t.user_role=? AND t.content_status='published' ORDER BY t.sort_order`, userID, role)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Topic, len(rows))
	for i, row := range rows {
		result[i] = topicFromRow(row)
		result[i].Levels, err = r.levels(userID, row.ID, row.QuizPassed)
		if err != nil {
			return nil, err
		}
		result[i].Completed = domain.TopicComplete(row.TheoryRead, row.QuizPassed, result[i].Levels)
	}
	return result, nil
}

func (r *PostgresRepository) Topic(userID, topicID int) (domain.Topic, error) {
	var row topicRow
	_, err := r.db.QueryOne(&row, `SELECT t.id,t.slug,t.user_role,t.title,t.description,t.sort_order,
        (p.theory_read_at IS NOT NULL) theory_read,COALESCE(p.quiz_passed,FALSE) quiz_passed,COALESCE(p.quiz_best_score,0) quiz_score,
        (p.completed_at IS NOT NULL) completed
		FROM topics t LEFT JOIN user_topic_progress p ON p.topic_id=t.id AND p.user_id=? WHERE t.id=? AND t.content_status='published'`, userID, topicID)
	if err == pg.ErrNoRows {
		return domain.Topic{}, service.ErrTopicNotFound
	}
	if err != nil {
		return domain.Topic{}, err
	}
	result := topicFromRow(row)
	result.Levels, err = r.levels(userID, topicID, row.QuizPassed)
	result.Completed = domain.TopicComplete(row.TheoryRead, row.QuizPassed, result.Levels)
	return result, err
}

func topicFromRow(row topicRow) domain.Topic {
	return domain.Topic{ID: row.ID, Slug: row.Slug, UserRole: domain.UserRole(row.UserRole), Title: row.Title, Description: row.Description, SortOrder: row.SortOrder, Status: row.ContentStatus, ArchivedAt: row.ArchivedAt, TheoryRead: row.TheoryRead, QuizPassed: row.QuizPassed, QuizScore: row.QuizScore, Completed: row.Completed}
}

func (r *PostgresRepository) levels(userID, topicID int, quizPassed bool) ([]domain.TopicLevelProgress, error) {
	type row struct {
		Number        int `pg:"number"`
		BestScore     int `pg:"best_score"`
		Stars         int `pg:"stars"`
		Attempts      int `pg:"attempts"`
		LastAttemptID int `pg:"last_attempt_id"`
	}
	var rows []row
	_, err := r.db.Query(&rows, `SELECT l.level_number number,COALESCE(p.best_score,0) best_score,COALESCE(p.stars,0) stars,COALESCE(p.attempts,0) attempts,
        COALESCE((SELECT s.id FROM chat_sessions s JOIN chats c ON c.id=s.chat_id WHERE s.user_id=? AND c.topic_id=? AND c.level_id=l.id AND s.status='COMPLETED' ORDER BY s.finished_at DESC LIMIT 1),0) last_attempt_id
        FROM levels l LEFT JOIN user_level_progress p ON p.level_id=l.id AND p.topic_id=? AND p.user_id=? ORDER BY l.level_number`, userID, topicID, topicID, userID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.TopicLevelProgress, len(rows))
	previousPassed := quizPassed
	for i, row := range rows {
		result[i] = domain.TopicLevelProgress{Number: row.Number, Opened: previousPassed, BestScore: row.BestScore, Stars: row.Stars, Attempts: row.Attempts, LastAttemptID: row.LastAttemptID}
		previousPassed = row.Stars > 0
	}
	return result, nil
}

func (r *PostgresRepository) Theory(topicID int) ([]domain.TheoryBlock, error) {
	var rows []struct {
		ID        int    `pg:"id"`
		TopicID   int    `pg:"topic_id"`
		SortOrder int    `pg:"sort_order"`
		Kind      string `pg:"kind"`
		Title     string `pg:"title"`
		Body      string `pg:"body"`
	}
	_, err := r.db.Query(&rows, `SELECT id,topic_id,sort_order,kind,title,body FROM theory_blocks WHERE topic_id=? ORDER BY sort_order`, topicID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.TheoryBlock, len(rows))
	for i, x := range rows {
		result[i] = domain.TheoryBlock{ID: x.ID, TopicID: x.TopicID, SortOrder: x.SortOrder, Kind: x.Kind, Title: x.Title, Body: x.Body}
	}
	return result, nil
}

func (r *PostgresRepository) MarkTheoryRead(userID, topicID int, activityDate time.Time) (domain.Streak, bool, error) {
	var streak domain.Streak
	newlyRead := false
	err := r.db.RunInTransaction(func(tx *pg.Tx) error {
		inserted, err := tx.Exec(`INSERT INTO user_topic_progress(user_id,topic_id,theory_read_at) VALUES(?,?,NOW()) ON CONFLICT(user_id,topic_id) DO NOTHING`, userID, topicID)
		if err != nil {
			return err
		}
		newlyRead = inserted.RowsAffected() > 0
		if !newlyRead {
			updated, updateErr := tx.Exec(`UPDATE user_topic_progress SET theory_read_at=NOW() WHERE user_id=? AND topic_id=? AND theory_read_at IS NULL`, userID, topicID)
			if updateErr != nil {
				return updateErr
			}
			newlyRead = updated.RowsAffected() > 0
		}
		if err = refreshTopicCompletion(tx, userID, topicID); err != nil {
			return err
		}
		streak, _, err = recordActivity(tx, userID, activityDate)
		return err
	})
	return streak, newlyRead, err
}

func (r *PostgresRepository) Quiz(topicID int) ([]domain.QuizQuestion, error) {
	return r.quiz(topicID, false)
}
func (r *PostgresRepository) quiz(topicID int, includeCorrect bool) ([]domain.QuizQuestion, error) {
	type qrow struct {
		ID          int    `pg:"id"`
		TopicID     int    `pg:"topic_id"`
		SortOrder   int    `pg:"sort_order"`
		Text        string `pg:"text"`
		Explanation string `pg:"explanation"`
	}
	var qs []qrow
	_, err := r.db.Query(&qs, `SELECT id,topic_id,sort_order,text,explanation FROM quiz_questions WHERE topic_id=? ORDER BY sort_order`, topicID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.QuizQuestion, len(qs))
	for i, q := range qs {
		result[i] = domain.QuizQuestion{ID: q.ID, TopicID: q.TopicID, SortOrder: q.SortOrder, Text: q.Text}
		if includeCorrect {
			result[i].Explanation = q.Explanation
		}
		var opts []struct {
			ID         int    `pg:"id"`
			QuestionID int    `pg:"question_id"`
			SortOrder  int    `pg:"sort_order"`
			Text       string `pg:"text"`
			IsCorrect  bool
		}
		_, err = r.db.Query(&opts, `SELECT id,question_id,sort_order,text,is_correct FROM quiz_options WHERE question_id=? ORDER BY sort_order`, q.ID)
		if err != nil {
			return nil, err
		}
		result[i].Options = make([]domain.QuizOption, len(opts))
		for j, o := range opts {
			result[i].Options[j] = domain.QuizOption{ID: o.ID, QuestionID: o.QuestionID, SortOrder: o.SortOrder, Text: o.Text, Correct: includeCorrect && o.IsCorrect}
		}
	}
	return result, nil
}

func (r *PostgresRepository) SubmitQuiz(userID, topicID int, answers []domain.QuizAnswer, activityDate time.Time) (domain.QuizResult, error) {
	questions, err := r.quiz(topicID, true)
	if err != nil {
		return domain.QuizResult{}, err
	}
	selected := map[int]int{}
	for _, a := range answers {
		if _, exists := selected[a.QuestionID]; exists {
			return domain.QuizResult{}, service.ErrInvalidQuiz
		}
		selected[a.QuestionID] = a.OptionID
	}
	correct := 0
	for _, q := range questions {
		optionID, ok := selected[q.ID]
		if !ok {
			return domain.QuizResult{}, service.ErrInvalidQuiz
		}
		valid := false
		for _, o := range q.Options {
			if o.ID == optionID {
				valid = true
				if o.Correct {
					correct++
				}
			}
		}
		if !valid {
			return domain.QuizResult{}, service.ErrInvalidQuiz
		}
	}
	score := correct * 20
	passed := domain.QuizPassed(score)
	result := domain.QuizResult{Score: score, Passed: passed}
	err = r.db.RunInTransaction(func(tx *pg.Tx) error {
		var previous bool
		_, _ = tx.QueryOne(pg.Scan(&previous), `SELECT quiz_passed FROM user_topic_progress WHERE user_id=? AND topic_id=?`, userID, topicID)
		result.NewlyPassed = passed && !previous
		if _, err := tx.Exec(`INSERT INTO quiz_attempts(user_id,topic_id,score,passed) VALUES(?,?,?,?)`, userID, topicID, score, passed); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO user_topic_progress(user_id,topic_id,quiz_passed,quiz_best_score) VALUES(?,?,?,?) ON CONFLICT(user_id,topic_id) DO UPDATE SET quiz_passed=user_topic_progress.quiz_passed OR EXCLUDED.quiz_passed,quiz_best_score=GREATEST(user_topic_progress.quiz_best_score,EXCLUDED.quiz_best_score)`, userID, topicID, passed, score); err != nil {
			return err
		}
		if err := refreshTopicCompletion(tx, userID, topicID); err != nil {
			return err
		}
		result.Streak, _, err = recordActivity(tx, userID, activityDate)
		return err
	})
	if err != nil {
		return domain.QuizResult{}, err
	}
	_, err = r.db.QueryOne(pg.Scan(&result.BestScore), `SELECT quiz_best_score FROM user_topic_progress WHERE user_id=? AND topic_id=?`, userID, topicID)
	return result, err
}

func (r *PostgresRepository) RecentAttempts(userID int, role domain.UserRole) ([]domain.RecentAttempt, float64, error) {
	var rows []struct {
		AttemptID  int       `pg:"attempt_id"`
		TopicID    int       `pg:"topic_id"`
		Level      int       `pg:"level"`
		Score      int       `pg:"score"`
		FinishedAt time.Time `pg:"finished_at"`
	}
	_, err := r.db.Query(&rows, `SELECT s.id attempt_id,c.topic_id,l.level_number level,s.score,s.finished_at
		FROM chat_sessions s JOIN chats c ON c.id=s.chat_id JOIN levels l ON l.id=c.level_id
		WHERE s.user_id=? AND c.user_role=? AND s.status='COMPLETED'
		ORDER BY s.finished_at DESC,s.id DESC LIMIT 10`, userID, role)
	if err != nil {
		return nil, 0, err
	}
	result := make([]domain.RecentAttempt, len(rows))
	for i, row := range rows {
		result[i] = domain.RecentAttempt{AttemptID: row.AttemptID, TopicID: row.TopicID, Level: row.Level, Score: row.Score, Stars: domain.StarsFromScore(row.Score), Finished: row.FinishedAt}
	}
	var average float64
	_, err = r.db.QueryOne(pg.Scan(&average), `SELECT COALESCE(AVG(s.score),0)::float8 FROM chat_sessions s JOIN chats c ON c.id=s.chat_id WHERE s.user_id=? AND c.user_role=? AND s.status='COMPLETED'`, userID, role)
	return result, average, err
}

func recordActivity(tx *pg.Tx, userID int, date time.Time) (domain.Streak, bool, error) {
	dateText := date.Format("2006-01-02")
	res, err := tx.Exec(`INSERT INTO daily_activity(user_id,activity_date) VALUES(?,?::date) ON CONFLICT DO NOTHING`, userID, dateText)
	if err != nil {
		return domain.Streak{}, false, err
	}
	added := res.RowsAffected() > 0
	var row struct {
		CurrentStreak    int    `pg:"current_streak"`
		LongestStreak    int    `pg:"longest_streak"`
		LastActivityDate string `pg:"last_activity_date"`
	}
	_, err = tx.QueryOne(&row, `SELECT current_streak,longest_streak,COALESCE(last_activity_date::text,'') last_activity_date FROM users WHERE id=? FOR UPDATE`, userID)
	if err != nil {
		return domain.Streak{}, false, err
	}
	location := date.Location()
	var lastActivity time.Time
	if row.LastActivityDate != "" {
		lastActivity, _ = time.ParseInLocation("2006-01-02", row.LastActivityDate, location)
	}
	current, longest := row.CurrentStreak, row.LongestStreak
	if added {
		var changed bool
		current, longest, changed = domain.NextStreak(current, longest, lastActivity, date)
		if changed {
			if _, err = tx.Exec(`UPDATE users SET current_streak=?,longest_streak=?,last_activity_date=?::date WHERE id=?`, current, longest, dateText, userID); err != nil {
				return domain.Streak{}, false, err
			}
			row.LastActivityDate = dateText
		}
	}
	streak := domain.Streak{Current: current, Longest: longest, ActiveToday: row.LastActivityDate == dateText, LastActivityDate: row.LastActivityDate}
	var stats domain.AchievementStats
	_, err = tx.QueryOne(&stats, `SELECT
		(SELECT COUNT(*) FROM chat_sessions WHERE user_id=? AND status='COMPLETED') completed_attempts,
		(SELECT COALESCE(MAX(score),0) FROM chat_sessions WHERE user_id=? AND status='COMPLETED') perfect_score,
		(SELECT COUNT(*) FROM user_topic_progress WHERE user_id=? AND completed_at IS NOT NULL) completed_topics,
		(SELECT COUNT(*) FROM user_topic_progress p JOIN topics t ON t.id=p.topic_id WHERE p.user_id=? AND p.completed_at IS NOT NULL AND t.user_role='buyer') buyer_topics,
		(SELECT COUNT(*) FROM user_topic_progress p JOIN topics t ON t.id=p.topic_id WHERE p.user_id=? AND p.completed_at IS NOT NULL AND t.user_role='seller') seller_topics,
		? streak`, userID, userID, userID, userID, userID, streak.Current)
	if err != nil {
		return domain.Streak{}, false, err
	}
	codes := domain.EligibleAchievementCodes(stats)
	if len(codes) > 0 {
		_, err = tx.Exec(`INSERT INTO user_achievements(user_id,achievement_id) SELECT ?,id FROM achievements WHERE code IN (?) ON CONFLICT DO NOTHING`, userID, pg.In(codes))
	}
	return streak, added, err
}

func refreshTopicCompletion(tx *pg.Tx, userID, topicID int) error {
	var state struct {
		TheoryRead bool `pg:"theory_read"`
		QuizPassed bool `pg:"quiz_passed"`
	}
	_, err := tx.QueryOne(&state, `SELECT theory_read_at IS NOT NULL theory_read,quiz_passed FROM user_topic_progress WHERE user_id=? AND topic_id=?`, userID, topicID)
	if err != nil {
		return err
	}
	var rows []struct {
		Number int `pg:"number"`
		Stars  int `pg:"stars"`
	}
	_, err = tx.Query(&rows, `SELECT l.level_number number,p.stars FROM user_level_progress p JOIN levels l ON l.id=p.level_id WHERE p.user_id=? AND p.topic_id=? ORDER BY l.level_number`, userID, topicID)
	if err != nil {
		return err
	}
	levels := make([]domain.TopicLevelProgress, len(rows))
	for i, row := range rows {
		levels[i] = domain.TopicLevelProgress{Number: row.Number, Stars: row.Stars}
	}
	if domain.TopicComplete(state.TheoryRead, state.QuizPassed, levels) {
		_, err = tx.Exec(`UPDATE user_topic_progress SET completed_at=COALESCE(completed_at,NOW()) WHERE user_id=? AND topic_id=?`, userID, topicID)
	}
	return err
}

func (r *PostgresRepository) Achievements(userID int) ([]domain.Achievement, error) {
	type row struct {
		Code        string    `pg:"code"`
		Title       string    `pg:"title"`
		Description string    `pg:"description"`
		Icon        string    `pg:"icon"`
		Earned      bool      `pg:"earned"`
		EarnedAt    time.Time `pg:"earned_at"`
		Target      int       `pg:"target"`
	}
	var rows []row
	_, err := r.db.Query(&rows, `SELECT a.code,a.title,a.description,COALESCE(a.icon,'') icon,(ua.id IS NOT NULL) earned,ua.received_at earned_at,a.condition_value::int target FROM achievements a LEFT JOIN user_achievements ua ON ua.achievement_id=a.id AND ua.user_id=? WHERE a.code NOT LIKE 'legacy-%' ORDER BY a.id`, userID)
	if err != nil {
		return nil, err
	}
	var stats struct {
		CompletedAttempts int
		PerfectScore      int
		CompletedTopics   int
		BuyerTopics       int
		SellerTopics      int
		Streak            int
	}
	_, _ = r.db.QueryOne(&stats, `SELECT (SELECT COUNT(*) FROM chat_sessions WHERE user_id=? AND status='COMPLETED') completed_attempts,(SELECT COALESCE(MAX(score),0) FROM chat_sessions WHERE user_id=? AND status='COMPLETED') perfect_score,(SELECT COUNT(*) FROM user_topic_progress WHERE user_id=? AND completed_at IS NOT NULL) completed_topics,(SELECT COUNT(*) FROM user_topic_progress p JOIN topics t ON t.id=p.topic_id WHERE p.user_id=? AND p.completed_at IS NOT NULL AND t.user_role='buyer') buyer_topics,(SELECT COUNT(*) FROM user_topic_progress p JOIN topics t ON t.id=p.topic_id WHERE p.user_id=? AND p.completed_at IS NOT NULL AND t.user_role='seller') seller_topics,(SELECT current_streak FROM users WHERE id=?) streak`, userID, userID, userID, userID, userID, userID)
	result := make([]domain.Achievement, len(rows))
	for i, x := range rows {
		current := domain.AchievementCurrent(x.Code, stats)
		result[i] = domain.Achievement{Code: x.Code, Title: x.Title, Description: x.Description, Icon: x.Icon, Earned: x.Earned, EarnedAt: x.EarnedAt, Current: current, Target: x.Target}
	}
	return result, nil
}

func (r *PostgresRepository) User(userID int) (domain.User, error) {
	type row struct {
		ID               int    `pg:"id"`
		Username         string `pg:"username"`
		AccessRole       string `pg:"access_role"`
		TrainingRole     string `pg:"training_role"`
		CurrentStreak    int
		LongestStreak    int
		LastActivityDate string
		ActiveToday      bool
	}
	var x row
	_, err := r.db.QueryOne(&x, `SELECT id,username,access_role,training_role,current_streak,longest_streak,COALESCE(last_activity_date::text,'') last_activity_date,last_activity_date=(NOW() AT TIME ZONE 'Europe/Moscow')::date active_today FROM users WHERE id=?`, userID)
	if err == pg.ErrNoRows {
		return domain.User{}, apperrors.ErrUserNotFound
	}
	return domain.User{ID: x.ID, Username: x.Username, AccessRole: domain.AccessRole(x.AccessRole), TrainingRole: domain.UserRole(x.TrainingRole), Streak: domain.Streak{Current: x.CurrentStreak, Longest: x.LongestStreak, ActiveToday: x.ActiveToday, LastActivityDate: x.LastActivityDate}}, err
}

func (r *PostgresRepository) InProgressAttempt(userID int, role domain.UserRole) (int, int, int, error) {
	var row struct {
		AttemptID int `pg:"attempt_id"`
		TopicID   int `pg:"topic_id"`
		Level     int `pg:"level"`
	}
	_, err := r.db.QueryOne(&row, `SELECT s.id attempt_id,c.topic_id,l.level_number level FROM chat_sessions s JOIN chats c ON c.id=s.chat_id JOIN topics t ON t.id=c.topic_id JOIN levels l ON l.id=c.level_id WHERE s.user_id=? AND c.user_role=? AND s.status='IN_PROGRESS' AND c.content_status='published' AND c.archived_at IS NULL AND t.content_status='published' ORDER BY s.started_at DESC LIMIT 1`, userID, role)
	if err == pg.ErrNoRows {
		return 0, 0, 0, nil
	}
	return row.AttemptID, row.TopicID, row.Level, err
}

func (r *PostgresRepository) ListContent() ([]domain.Topic, error) {
	var rows []topicRow
	_, err := r.db.Query(&rows, `SELECT id,slug,user_role,title,description,sort_order,content_status,archived_at FROM topics ORDER BY user_role,sort_order,id`)
	result := make([]domain.Topic, len(rows))
	for i, row := range rows {
		result[i] = topicFromRow(row)
	}
	return result, mapContentError(err)
}

func (r *PostgresRepository) Content(id int) (domain.TopicContent, error) {
	var row topicRow
	_, err := r.db.QueryOne(&row, `SELECT id,slug,user_role,title,description,sort_order,content_status,archived_at FROM topics WHERE id=?`, id)
	if err != nil {
		return domain.TopicContent{}, mapContentError(err)
	}
	theory, err := r.Theory(id)
	if err != nil {
		return domain.TopicContent{}, err
	}
	quiz, err := r.quiz(id, true)
	if err != nil {
		return domain.TopicContent{}, err
	}
	return domain.TopicContent{Topic: topicFromRow(row), Theory: theory, Quiz: quiz}, nil
}

func (r *PostgresRepository) CreateTopic(topic domain.Topic) (domain.Topic, error) {
	_, err := r.db.QueryOne(pg.Scan(&topic.ID), `INSERT INTO topics(slug,user_role,title,description,sort_order,content_status) VALUES(?,?,?,?,?,'draft') RETURNING id`, topic.Slug, topic.UserRole, topic.Title, topic.Description, topic.SortOrder)
	topic.Status = domain.TopicStatusDraft
	return topic, mapContentError(err)
}
func (r *PostgresRepository) UpdateTopic(topic domain.Topic) error {
	result, err := r.db.Exec(`UPDATE topics SET slug=?,user_role=?,title=?,description=?,sort_order=? WHERE id=? AND content_status='draft'`, topic.Slug, topic.UserRole, topic.Title, topic.Description, topic.SortOrder, topic.ID)
	if err == nil && result.RowsAffected() == 0 {
		err = service.ErrContentConflict
	}
	return mapContentError(err)
}
func (r *PostgresRepository) SetTopicStatus(id int, status string) error {
	var archived any
	if status == domain.TopicStatusArchived {
		archived = time.Now().UTC()
	}
	_, err := r.db.Exec(`UPDATE topics SET content_status=?,archived_at=? WHERE id=?`, status, archived, id)
	return mapContentError(err)
}
func (r *PostgresRepository) CreateTheoryBlock(block domain.TheoryBlock) (domain.TheoryBlock, error) {
	err := r.withDraftTopic(block.TopicID, func(tx *pg.Tx) error {
		_, err := tx.QueryOne(pg.Scan(&block.ID), `INSERT INTO theory_blocks(topic_id,sort_order,kind,title,body) VALUES(?,?,?,?,?) RETURNING id`, block.TopicID, block.SortOrder, block.Kind, block.Title, block.Body)
		return err
	})
	return block, mapContentError(err)
}
func (r *PostgresRepository) UpdateTheoryBlock(block domain.TheoryBlock) error {
	err := r.withDraftTopic(block.TopicID, func(tx *pg.Tx) error {
		res, err := tx.Exec(`UPDATE theory_blocks SET sort_order=?,kind=?,title=?,body=? WHERE id=? AND topic_id=?`, block.SortOrder, block.Kind, block.Title, block.Body, block.ID, block.TopicID)
		if err == nil && res.RowsAffected() == 0 {
			err = service.ErrTopicNotFound
		}
		return err
	})
	return mapContentError(err)
}
func (r *PostgresRepository) DeleteTheoryBlock(topicID, blockID int) error {
	err := r.withDraftTopic(topicID, func(tx *pg.Tx) error {
		res, err := tx.Exec(`DELETE FROM theory_blocks WHERE id=? AND topic_id=?`, blockID, topicID)
		if err == nil && res.RowsAffected() == 0 {
			err = service.ErrTopicNotFound
		}
		return err
	})
	return mapContentError(err)
}
func (r *PostgresRepository) CreateQuizQuestion(q domain.QuizQuestion) (domain.QuizQuestion, error) {
	err := r.withDraftTopic(q.TopicID, func(tx *pg.Tx) error {
		_, err := tx.QueryOne(pg.Scan(&q.ID), `INSERT INTO quiz_questions(topic_id,sort_order,text,explanation) VALUES(?,?,?,?) RETURNING id`, q.TopicID, q.SortOrder, q.Text, q.Explanation)
		return err
	})
	return q, mapContentError(err)
}
func (r *PostgresRepository) UpdateQuizQuestion(q domain.QuizQuestion) error {
	err := r.withDraftTopic(q.TopicID, func(tx *pg.Tx) error {
		res, err := tx.Exec(`UPDATE quiz_questions SET sort_order=?,text=?,explanation=? WHERE id=? AND topic_id=?`, q.SortOrder, q.Text, q.Explanation, q.ID, q.TopicID)
		if err == nil && res.RowsAffected() == 0 {
			err = service.ErrTopicNotFound
		}
		return err
	})
	return mapContentError(err)
}
func (r *PostgresRepository) DeleteQuizQuestion(topicID, questionID int) error {
	err := r.withDraftTopic(topicID, func(tx *pg.Tx) error {
		res, err := tx.Exec(`DELETE FROM quiz_questions WHERE id=? AND topic_id=?`, questionID, topicID)
		if err == nil && res.RowsAffected() == 0 {
			err = service.ErrTopicNotFound
		}
		return err
	})
	return mapContentError(err)
}
func (r *PostgresRepository) CreateQuizOption(o domain.QuizOption) (domain.QuizOption, error) {
	var topicID int
	if _, err := r.db.QueryOne(pg.Scan(&topicID), `SELECT topic_id FROM quiz_questions WHERE id=?`, o.QuestionID); err != nil {
		return domain.QuizOption{}, mapContentError(err)
	}
	err := r.withDraftTopic(topicID, func(tx *pg.Tx) error {
		_, err := tx.QueryOne(pg.Scan(&o.ID), `INSERT INTO quiz_options(question_id,sort_order,text,is_correct) VALUES(?,?,?,?) RETURNING id`, o.QuestionID, o.SortOrder, o.Text, o.Correct)
		return err
	})
	return o, mapContentError(err)
}
func (r *PostgresRepository) UpdateQuizOption(o domain.QuizOption) error {
	var topicID int
	if _, err := r.db.QueryOne(pg.Scan(&topicID), `SELECT topic_id FROM quiz_questions WHERE id=?`, o.QuestionID); err != nil {
		return mapContentError(err)
	}
	err := r.withDraftTopic(topicID, func(tx *pg.Tx) error {
		res, err := tx.Exec(`UPDATE quiz_options SET sort_order=?,text=?,is_correct=? WHERE id=? AND question_id=?`, o.SortOrder, o.Text, o.Correct, o.ID, o.QuestionID)
		if err == nil && res.RowsAffected() == 0 {
			err = service.ErrTopicNotFound
		}
		return err
	})
	return mapContentError(err)
}
func (r *PostgresRepository) DeleteQuizOption(questionID, optionID int) error {
	var topicID int
	if _, err := r.db.QueryOne(pg.Scan(&topicID), `SELECT topic_id FROM quiz_questions WHERE id=?`, questionID); err != nil {
		return mapContentError(err)
	}
	err := r.withDraftTopic(topicID, func(tx *pg.Tx) error {
		res, err := tx.Exec(`DELETE FROM quiz_options WHERE id=? AND question_id=?`, optionID, questionID)
		if err == nil && res.RowsAffected() == 0 {
			err = service.ErrTopicNotFound
		}
		return err
	})
	return mapContentError(err)
}

func (r *PostgresRepository) PublishTopic(id int) error {
	return r.db.RunInTransaction(func(tx *pg.Tx) error {
		var status string
		if _, err := tx.QueryOne(pg.Scan(&status), `SELECT content_status FROM topics WHERE id=? FOR UPDATE`, id); err != nil {
			return mapContentError(err)
		}
		if status != domain.TopicStatusDraft {
			return service.ErrContentConflict
		}
		var valid bool
		_, err := tx.QueryOne(pg.Scan(&valid), `SELECT
			(SELECT COUNT(*)=5 AND MIN(sort_order)=1 AND MAX(sort_order)=5 FROM theory_blocks WHERE topic_id=t.id)
			AND (SELECT COUNT(*)=5 AND MIN(sort_order)=1 AND MAX(sort_order)=5 FROM quiz_questions WHERE topic_id=t.id)
			AND NOT EXISTS (SELECT 1 FROM quiz_questions q WHERE q.topic_id=t.id AND ((SELECT COUNT(*) FROM quiz_options o WHERE o.question_id=q.id)<>4 OR (SELECT COUNT(*) FROM quiz_options o WHERE o.question_id=q.id AND o.is_correct)<>1 OR (SELECT MIN(sort_order)=1 AND MAX(sort_order)=4 FROM quiz_options o WHERE o.question_id=q.id) IS NOT TRUE))
			AND (SELECT COUNT(DISTINCT l.level_number)=4 FROM chats c JOIN levels l ON l.id=c.level_id WHERE c.topic_id=t.id AND c.user_role=t.user_role AND c.content_status='published' AND c.archived_at IS NULL)
			FROM topics t WHERE t.id=? AND t.content_status='draft'`, id)
		if err != nil {
			return mapContentError(err)
		}
		if !valid {
			return service.ErrContentConflict
		}
		_, err = tx.Exec(`UPDATE topics SET content_status='published',archived_at=NULL WHERE id=?`, id)
		return mapContentError(err)
	})
}

func (r *PostgresRepository) withDraftTopic(topicID int, action func(*pg.Tx) error) error {
	return r.db.RunInTransaction(func(tx *pg.Tx) error {
		var status string
		if _, err := tx.QueryOne(pg.Scan(&status), `SELECT content_status FROM topics WHERE id=? FOR UPDATE`, topicID); err != nil {
			return mapContentError(err)
		}
		if status != domain.TopicStatusDraft {
			return service.ErrContentConflict
		}
		return mapContentError(action(tx))
	})
}

func mapContentError(err error) error {
	if err == pg.ErrNoRows {
		return service.ErrTopicNotFound
	}
	if pgErr, ok := err.(pg.Error); ok && pgErr.IntegrityViolation() {
		return service.ErrContentConflict
	}
	return err
}

type dailyTaskRow struct {
	ActivityDate string                   `sql:"activity_date"`
	UserRole     string                   `sql:"user_role"`
	Messages     []domain.DialogueMessage `sql:"messages"`
	Verdict      bool                     `sql:"verdict"`
	Signals      []string                 `sql:"signals"`
	SafeAction   string                   `sql:"safe_action"`
	Answer       *bool                    `sql:"user_answer"`
	Correct      *bool                    `sql:"is_correct"`
	CompletedAt  *time.Time               `sql:"completed_at"`
}

func (r *PostgresRepository) FindDailyTask(userID int, date time.Time) (domain.DailyTask, bool, error) {
	var row dailyTaskRow
	_, err := r.db.QueryOne(&row, `SELECT activity_date::text,user_role,messages,verdict,signals,safe_action,user_answer,is_correct,completed_at FROM daily_tasks WHERE user_id=? AND activity_date=?::date`, userID, date.Format("2006-01-02"))
	if err == pg.ErrNoRows {
		return domain.DailyTask{}, false, nil
	}
	if err != nil {
		return domain.DailyTask{}, false, err
	}
	return taskFromRow(row), true, nil
}

func (r *PostgresRepository) DailyTask(userID int, date time.Time, created domain.DailyTask) (domain.DailyTask, error) {
	dateText := date.Format("2006-01-02")
	var row dailyTaskRow
	err := r.db.RunInTransaction(func(tx *pg.Tx) error {
		messages, _ := json.Marshal(created.Messages)
		signals, _ := json.Marshal(created.Signals)
		if _, err := tx.Exec(`INSERT INTO daily_tasks(user_id,activity_date,user_role,messages,verdict,signals,safe_action) VALUES(?,?::date,?,?::jsonb,?,?::jsonb,?) ON CONFLICT(user_id,activity_date) DO NOTHING`, userID, dateText, created.Role, string(messages), created.Verdict, string(signals), created.SafeAction); err != nil {
			return err
		}
		_, err := tx.QueryOne(&row, `SELECT activity_date::text,user_role,messages,verdict,signals,safe_action,user_answer,is_correct,completed_at FROM daily_tasks WHERE user_id=? AND activity_date=?::date FOR UPDATE`, userID, dateText)
		return err
	})
	if err != nil {
		return domain.DailyTask{}, err
	}
	return taskFromRow(row), nil
}
func (r *PostgresRepository) AnswerDailyTask(userID int, date time.Time, answer bool) (domain.DailyTask, domain.Streak, error) {
	dateText := date.Format("2006-01-02")
	var row dailyTaskRow
	var streak domain.Streak
	err := r.db.RunInTransaction(func(tx *pg.Tx) error {
		_, err := tx.QueryOne(&row, `SELECT activity_date::text,user_role,messages,verdict,signals,safe_action,user_answer,is_correct,completed_at FROM daily_tasks WHERE user_id=? AND activity_date=?::date FOR UPDATE`, userID, dateText)
		if err == pg.ErrNoRows {
			return service.ErrDailyTaskUnavailable
		}
		if err != nil {
			return err
		}
		if row.CompletedAt != nil {
			return service.ErrDailyTaskAnswered
		}
		correct := answer == row.Verdict
		if _, err = tx.Exec(`UPDATE daily_tasks SET user_answer=?,is_correct=?,completed_at=NOW() WHERE user_id=? AND activity_date=?::date`, answer, correct, userID, dateText); err != nil {
			return err
		}
		row.Answer, row.Correct = &answer, &correct
		now := time.Now()
		row.CompletedAt = &now
		streak, _, err = recordActivity(tx, userID, date)
		return err
	})
	if err != nil {
		return domain.DailyTask{}, domain.Streak{}, err
	}
	return taskFromRow(row), streak, nil
}
func taskFromRow(row dailyTaskRow) domain.DailyTask {
	return domain.DailyTask{Date: row.ActivityDate, Role: domain.UserRole(row.UserRole), Messages: row.Messages, Verdict: row.Verdict, Signals: row.Signals, SafeAction: row.SafeAction, Answer: row.Answer, Correct: row.Correct, Completed: row.CompletedAt != nil, CompletedAt: row.CompletedAt}
}
