package content_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	learningrepository "anti-scam-trainer/backend/internal/features/learning/repository"
	learningservice "anti-scam-trainer/backend/internal/features/learning/service"
	scenariosrepository "anti-scam-trainer/backend/internal/features/scenarios/repository"
	"os"
	"testing"
	"time"

	"github.com/go-pg/pg"
)

// TestPublishedContentMatrix runs against the disposable migrated database used by acceptance.
func TestPublishedContentMatrix(t *testing.T) {
	database := os.Getenv("POSTGRES_TEST_NAME")
	if database == "" {
		t.Skip("POSTGRES_TEST_NAME is not set")
	}
	db := pg.Connect(&pg.Options{Addr: os.Getenv("POSTGRES_HOST") + ":" + os.Getenv("POSTGRES_PORT"), User: os.Getenv("POSTGRES_USER"), Password: os.Getenv("POSTGRES_PASSWORD"), Database: database})
	defer db.Close()
	var topics, theory, quiz, scenarios int
	_, err := db.QueryOne(pg.Scan(&topics, &theory, &quiz, &scenarios), `SELECT (SELECT COUNT(*) FROM topics),(SELECT COUNT(*) FROM theory_blocks),(SELECT COUNT(*) FROM quiz_questions),(SELECT COUNT(*) FROM chats WHERE content_status='published' AND archived_at IS NULL)`)
	if err != nil {
		t.Fatal(err)
	}
	if topics != 12 || theory != 60 || quiz != 60 || scenarios != 48 {
		t.Fatalf("content counts=(%d,%d,%d,%d), want (12,60,60,48)", topics, theory, quiz, scenarios)
	}
	var invalid int
	_, err = db.QueryOne(pg.Scan(&invalid), `SELECT COUNT(*) FROM (
		SELECT c.id,l.level_number,COUNT(DISTINCT s.id) steps,MIN(s.step_number) first_step,MAX(s.step_number) last_step,
			COUNT(DISTINCT o.id) options,COUNT(DISTINCT s.id) FILTER (WHERE s.response_type='mixed') mixed_steps,
			COUNT(DISTINCT s.id) FILTER (WHERE s.response_type='free_text') free_text_steps
		FROM chats c JOIN levels l ON l.id=c.level_id JOIN chat_steps s ON s.chat_id=c.id LEFT JOIN chat_options o ON o.step_id=s.id
		WHERE c.content_status='published' AND c.archived_at IS NULL GROUP BY c.id,l.level_number
		HAVING COUNT(DISTINCT s.id)<>CASE WHEN l.level_number=4 THEN 4 ELSE 3 END OR MIN(s.step_number)<>1
			OR MAX(s.step_number)<>CASE WHEN l.level_number=4 THEN 4 ELSE 3 END
			OR (l.level_number<=2 AND COUNT(DISTINCT o.id)<>12)
			OR (l.level_number=3 AND (COUNT(DISTINCT s.id) FILTER (WHERE s.response_type='mixed')<>2 OR COUNT(DISTINCT o.id)<>12))
			OR (l.level_number=4 AND COUNT(DISTINCT s.id) FILTER (WHERE s.response_type='free_text')<>4)
	) bad`)
	if err != nil {
		t.Fatal(err)
	}
	if invalid != 0 {
		t.Fatalf("invalid scenario structures=%d", invalid)
	}
	_, err = db.QueryOne(pg.Scan(&invalid), `SELECT COUNT(*) FROM (
		SELECT s.id FROM chat_steps s JOIN chats c ON c.id=s.chat_id JOIN levels l ON l.id=c.level_id LEFT JOIN chat_options o ON o.step_id=s.id
		WHERE c.content_status='published' AND l.level_number<=3 GROUP BY s.id HAVING COUNT(o.id)<>4
	) bad`)
	if err != nil {
		t.Fatal(err)
	}
	if invalid != 0 {
		t.Fatalf("steps without four options=%d", invalid)
	}
	_, err = db.QueryOne(pg.Scan(&invalid), `SELECT COUNT(*) FROM chat_options o JOIN chat_steps s ON s.id=o.step_id JOIN chats c ON c.id=s.chat_id WHERE c.content_status='published' AND (o.points NOT IN(0,25,50,75,100) OR char_length(o.option_text)>140)`)
	if err != nil {
		t.Fatal(err)
	}
	if invalid != 0 {
		t.Fatalf("invalid options=%d", invalid)
	}
	_, err = db.QueryOne(pg.Scan(&invalid), `SELECT COUNT(*) FROM chat_options l2o JOIN chat_steps l2s ON l2s.id=l2o.step_id JOIN chats l2c ON l2c.id=l2s.chat_id JOIN levels l2l ON l2l.id=l2c.level_id WHERE l2c.content_status='published' AND l2l.level_number=2 AND (l2o.points NOT IN(25,50,75,100) OR EXISTS(SELECT 1 FROM chat_options l1o JOIN chat_steps l1s ON l1s.id=l1o.step_id JOIN chats l1c ON l1c.id=l1s.chat_id JOIN levels l1l ON l1l.id=l1c.level_id WHERE l1c.topic_id=l2c.topic_id AND l1l.level_number=1 AND l1o.option_text=l2o.option_text))`)
	if err != nil {
		t.Fatal(err)
	}
	if invalid != 0 {
		t.Fatalf("level 2 has non-similar-choice scoring or duplicates level 1 options: %d", invalid)
	}
	_, err = db.QueryOne(pg.Scan(&invalid), `SELECT COUNT(*) FROM chat_steps s JOIN chats c ON c.id=s.chat_id WHERE c.content_status='published' AND (NULLIF(s.counterparty_message,'') IS NULL OR char_length(s.counterparty_message)>280 OR (s.response_type IN('mixed','free_text') AND (NULLIF(s.ai_instruction,'') IS NULL OR NULLIF(s.fallback_message,'') IS NULL)))`)
	if err != nil {
		t.Fatal(err)
	}
	if invalid != 0 {
		t.Fatalf("invalid steps=%d", invalid)
	}
	_, err = db.QueryOne(pg.Scan(&invalid), `SELECT COUNT(*) FROM (
		SELECT title text FROM topics UNION ALL SELECT description FROM topics UNION ALL SELECT body FROM theory_blocks
		UNION ALL SELECT text FROM quiz_questions UNION ALL SELECT text FROM quiz_options
		UNION ALL SELECT counterparty_message FROM chat_steps UNION ALL SELECT fallback_message FROM chat_steps
		UNION ALL SELECT option_text FROM chat_options
	) content WHERE text ~* '(https?://|www\.|[0-9]{10,})' OR text ~* '(avito доставка|безопасная сделка).{0,20}(опасн|мошенн)'`)
	if err != nil {
		t.Fatal(err)
	}
	if invalid != 0 {
		t.Fatalf("unsafe content fragments=%d", invalid)
	}
	var theorySignatures, quizSignatures, scenarioSignatures int
	_, err = db.QueryOne(pg.Scan(&theorySignatures, &quizSignatures, &scenarioSignatures), `SELECT
		(SELECT COUNT(DISTINCT signature) FROM (SELECT topic_id,string_agg(body,'|' ORDER BY sort_order) signature FROM theory_blocks GROUP BY topic_id) x),
		(SELECT COUNT(DISTINCT signature) FROM (SELECT topic_id,string_agg(text||' '||explanation,'|' ORDER BY sort_order) signature FROM quiz_questions GROUP BY topic_id) x),
		(SELECT COUNT(DISTINCT signature) FROM (SELECT c.topic_id,string_agg(s.counterparty_message,'|' ORDER BY l.level_number,s.step_number) signature FROM chats c JOIN levels l ON l.id=c.level_id JOIN chat_steps s ON s.chat_id=c.id WHERE c.content_status='published' GROUP BY c.topic_id) x)`)
	if err != nil {
		t.Fatal(err)
	}
	if theorySignatures != 12 || quizSignatures != 12 || scenarioSignatures != 12 {
		t.Fatalf("topic-specific signatures=(%d,%d,%d), want all 12", theorySignatures, quizSignatures, scenarioSignatures)
	}

	var scenarioIDs []int
	if _, err := db.Query(&scenarioIDs, `SELECT id FROM chats WHERE content_status='published' AND archived_at IS NULL ORDER BY id`); err != nil {
		t.Fatal(err)
	}
	validator := scenariosrepository.NewPostgres(db)
	for _, scenarioID := range scenarioIDs {
		valid, err := validator.ValidContent(scenarioID)
		if err != nil {
			t.Fatalf("validate scenario %d: %v", scenarioID, err)
		}
		if !valid {
			t.Fatalf("published scenario %d does not pass publication validator", scenarioID)
		}
	}
}

func TestLearningActivityAwardsStreakAchievements(t *testing.T) {
	database := os.Getenv("POSTGRES_TEST_NAME")
	if database == "" {
		t.Skip("POSTGRES_TEST_NAME is not set")
	}
	db := pg.Connect(&pg.Options{Addr: os.Getenv("POSTGRES_HOST") + ":" + os.Getenv("POSTGRES_PORT"), User: os.Getenv("POSTGRES_USER"), Password: os.Getenv("POSTGRES_PASSWORD"), Database: database})
	defer db.Close()
	var userID, topicID int
	_, err := db.QueryOne(pg.Scan(&userID), `INSERT INTO users(username,password_hash,access_role,training_role,current_streak,longest_streak,last_activity_date) VALUES('streak-learning-test','hash','user','buyer',2,2,'2026-08-08') RETURNING id`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = db.Exec(`DELETE FROM users WHERE id=?`, userID) }()
	_, err = db.QueryOne(pg.Scan(&topicID), `SELECT id FROM topics WHERE slug='buyer-phishing-links'`)
	if err != nil {
		t.Fatal(err)
	}
	repository := learningrepository.NewPostgres(db)
	activityDate := time.Date(2026, 8, 9, 0, 0, 0, 0, time.FixedZone("Europe/Moscow", 3*60*60))
	streak, _, err := repository.MarkTheoryRead(userID, topicID, activityDate)
	if err != nil || streak.Current != 3 {
		t.Fatalf("theory streak = (%#v,%v)", streak, err)
	}
	var awarded int
	_, err = db.QueryOne(pg.Scan(&awarded), `SELECT COUNT(*) FROM user_achievements ua JOIN achievements a ON a.id=ua.achievement_id WHERE ua.user_id=? AND a.code='streak_3'`, userID)
	if err != nil || awarded != 1 {
		t.Fatalf("streak_3 awards=%d, err=%v", awarded, err)
	}

	_, err = db.Exec(`UPDATE users SET current_streak=6,longest_streak=6,last_activity_date='2026-08-09' WHERE id=?`, userID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO user_level_progress(user_id,level_id,user_role,topic_id,best_score,stars,attempts,passed_at) SELECT ?,id,'buyer',?,80,1,1,NOW() FROM levels`, userID, topicID)
	if err != nil {
		t.Fatal(err)
	}
	var answers []domain.QuizAnswer
	_, err = db.Query(&answers, `SELECT q.id question_id,o.id option_id FROM quiz_questions q JOIN quiz_options o ON o.question_id=q.id AND o.is_correct=TRUE WHERE q.topic_id=? ORDER BY q.sort_order`, topicID)
	if err != nil {
		t.Fatal(err)
	}
	quizDate := activityDate.AddDate(0, 0, 1)
	if _, err = repository.SubmitQuiz(userID, topicID, answers, quizDate); err != nil {
		t.Fatal(err)
	}
	_, err = db.QueryOne(pg.Scan(&awarded), `SELECT COUNT(*) FROM user_achievements ua JOIN achievements a ON a.id=ua.achievement_id WHERE ua.user_id=? AND a.code='streak_7'`, userID)
	if err != nil || awarded != 1 {
		t.Fatalf("streak_7 awards=%d, err=%v", awarded, err)
	}
	var completed bool
	_, err = db.QueryOne(pg.Scan(&completed), `SELECT completed_at IS NOT NULL FROM user_topic_progress WHERE user_id=? AND topic_id=?`, userID, topicID)
	if err != nil || !completed {
		t.Fatalf("quiz did not persist migrated topic completion: completed=%v err=%v", completed, err)
	}
	_, err = db.QueryOne(pg.Scan(&awarded), `SELECT COUNT(*) FROM user_achievements ua JOIN achievements a ON a.id=ua.achievement_id WHERE ua.user_id=? AND a.code='first_topic_completed'`, userID)
	if err != nil || awarded != 1 {
		t.Fatalf("first_topic_completed awards=%d, err=%v", awarded, err)
	}
}

func TestPostgresDailyTaskAndTopicLifecycle(t *testing.T) {
	database := os.Getenv("POSTGRES_TEST_NAME")
	if database == "" {
		t.Skip("POSTGRES_TEST_NAME is not set")
	}
	db := pg.Connect(&pg.Options{Addr: os.Getenv("POSTGRES_HOST") + ":" + os.Getenv("POSTGRES_PORT"), User: os.Getenv("POSTGRES_USER"), Password: os.Getenv("POSTGRES_PASSWORD"), Database: database})
	defer db.Close()
	var userID, topicID int
	if _, err := db.QueryOne(pg.Scan(&userID), `INSERT INTO users(username,password_hash,access_role,training_role) VALUES('daily-content-test','hash','user','buyer') RETURNING id`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Exec(`UPDATE topics SET content_status='published',archived_at=NULL WHERE id=?`, topicID)
		_, _ = db.Exec(`DELETE FROM users WHERE id=?`, userID)
	}()
	if _, err := db.QueryOne(pg.Scan(&topicID), `SELECT id FROM topics WHERE user_role='buyer' ORDER BY sort_order LIMIT 1`); err != nil {
		t.Fatal(err)
	}
	repository := learningrepository.NewPostgres(db)
	now := time.Date(2026, 8, 9, 9, 0, 0, 0, time.UTC)
	learning := learningservice.NewWithClock(repository, func() time.Time { return now })
	_, _, _, _, first, err := learning.Dashboard(userID, domain.UserRoleBuyer)
	if err != nil || first == nil || len(first.Messages) == 0 || first.Completed {
		t.Fatalf("first daily task=(%#v,%v)", first, err)
	}
	_, _, _, _, refresh, err := learning.Dashboard(userID, domain.UserRoleBuyer)
	if err != nil || refresh.Date != first.Date || refresh.Role != first.Role {
		t.Fatalf("refreshed daily task=(%#v,%v)", refresh, err)
	}
	content := learningservice.NewContent(repository)
	if err = content.Archive(topicID); err != nil {
		t.Fatal(err)
	}
	var progress int
	if _, err = db.QueryOne(pg.Scan(&progress), `SELECT COUNT(*) FROM user_topic_progress WHERE user_id=? AND topic_id=?`, userID, topicID); err != nil || progress != 1 {
		t.Fatalf("archived progress=(%d,%v)", progress, err)
	}
	if err = content.Restore(topicID); err != nil {
		t.Fatal(err)
	}
	if err = content.Publish(topicID); err != nil {
		t.Fatalf("republish complete seed topic: %v", err)
	}
}

func TestProgressStatsExcludeAbandonedAttempts(t *testing.T) {
	database := os.Getenv("POSTGRES_TEST_NAME")
	if database == "" {
		t.Skip("POSTGRES_TEST_NAME is not set")
	}
	db := pg.Connect(&pg.Options{Addr: os.Getenv("POSTGRES_HOST") + ":" + os.Getenv("POSTGRES_PORT"), User: os.Getenv("POSTGRES_USER"), Password: os.Getenv("POSTGRES_PASSWORD"), Database: database})
	defer db.Close()
	var userID, chatID int
	_, err := db.QueryOne(pg.Scan(&userID), `INSERT INTO users(username,password_hash,access_role,training_role) VALUES('progress-stats-test','hash','user','buyer') RETURNING id`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Exec(`DELETE FROM chat_sessions WHERE user_id=?`, userID)
		_, _ = db.Exec(`DELETE FROM users WHERE id=?`, userID)
	}()
	_, err = db.QueryOne(pg.Scan(&chatID), `SELECT c.id FROM chats c JOIN topics t ON t.id=c.topic_id JOIN levels l ON l.id=c.level_id WHERE t.slug='buyer-phishing-links' AND l.level_number=1 AND c.content_status='published'`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO chat_sessions(user_id,chat_id,status,mode,current_step_number,score,user_role,finished_at) VALUES
		(?,?,'COMPLETED','scenario',3,80,'buyer',NOW()-INTERVAL '1 minute'),
		(?,?,'ABANDONED','scenario',2,100,'buyer',NOW())`, userID, chatID, userID, chatID)
	if err != nil {
		t.Fatal(err)
	}
	recent, average, err := learningrepository.NewPostgres(db).RecentAttempts(userID, domain.UserRoleBuyer)
	if err != nil || len(recent) != 1 || recent[0].Score != 80 || average != 80 {
		t.Fatalf("progress stats = (%#v,%v,%v), want one completed score 80", recent, average, err)
	}
}
