package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-pg/pg"
	"golang.org/x/crypto/bcrypt"
)

const (
	demoUsername = "demo-seller"
	demoPassword = "demo1234"
)

func main() {
	db := pg.Connect(&pg.Options{
		Addr:     env("POSTGRES_HOST", "localhost") + ":" + env("POSTGRES_PORT", "5432"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Database: env("POSTGRES_NAME", os.Getenv("POSTGRES_DB")),
	})
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("demo reset: close database: %v", err)
		}
	}()
	if _, err := db.Exec(`SELECT 1`); err != nil {
		log.Printf("demo reset: database is unavailable: %v", err)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(demoPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("demo reset: hash password: %v", err)
		return
	}
	if err := reset(db, string(hash)); err != nil {
		log.Printf("demo reset: %v", err)
		return
	}
	fmt.Printf("Demo account is ready: %s / %s\n", demoUsername, demoPassword)
}

func reset(db *pg.DB, passwordHash string) error {
	return db.RunInTransaction(func(tx *pg.Tx) error {
		if _, err := tx.Exec(`DELETE FROM chat_sessions WHERE user_id=(SELECT id FROM users WHERE username=?)`, demoUsername); err != nil {
			return fmt.Errorf("remove previous attempts: %w", err)
		}
		if _, err := tx.Exec(`DELETE FROM users WHERE username = ?`, demoUsername); err != nil {
			return fmt.Errorf("remove previous account: %w", err)
		}

		var userID, topicID, levelOneID, levelTwoID, levelOneScenarioID, levelTwoScenarioID int
		if _, err := tx.QueryOne(pg.Scan(&userID), `
			INSERT INTO users(username,password_hash,access_role,training_role,current_streak,longest_streak,last_activity_date)
			VALUES(? ,? ,'user','seller',2,3,(CURRENT_TIMESTAMP AT TIME ZONE 'Europe/Moscow')::date-1)
			RETURNING id`, demoUsername, passwordHash); err != nil {
			return fmt.Errorf("create account: %w", err)
		}
		if _, err := tx.QueryOne(pg.Scan(&topicID), `SELECT id FROM topics WHERE slug='seller-fake-payment' AND content_status='published'`); err != nil {
			return fmt.Errorf("find demo topic: %w", err)
		}
		if _, err := tx.QueryOne(pg.Scan(&levelOneID), `SELECT id FROM levels WHERE level_number=1`); err != nil {
			return fmt.Errorf("find level 1: %w", err)
		}
		if _, err := tx.QueryOne(pg.Scan(&levelTwoID), `SELECT id FROM levels WHERE level_number=2`); err != nil {
			return fmt.Errorf("find level 2: %w", err)
		}
		if _, err := tx.QueryOne(pg.Scan(&levelOneScenarioID), `SELECT id FROM chats WHERE topic_id=? AND level_id=? AND content_status='published' AND archived_at IS NULL`, topicID, levelOneID); err != nil {
			return fmt.Errorf("find level 1 scenario: %w", err)
		}
		if _, err := tx.QueryOne(pg.Scan(&levelTwoScenarioID), `SELECT id FROM chats WHERE topic_id=? AND level_id=? AND content_status='published' AND archived_at IS NULL`, topicID, levelTwoID); err != nil {
			return fmt.Errorf("find level 2 scenario: %w", err)
		}

		if _, err := tx.Exec(`
			INSERT INTO user_topic_progress(user_id,topic_id,theory_read_at,quiz_passed,quiz_best_score)
			VALUES(?,?,CURRENT_TIMESTAMP,TRUE,100)`, userID, topicID); err != nil {
			return fmt.Errorf("seed topic progress: %w", err)
		}
		if _, err := tx.Exec(`INSERT INTO quiz_attempts(user_id,topic_id,score,passed) VALUES(?,?,100,TRUE)`, userID, topicID); err != nil {
			return fmt.Errorf("seed quiz: %w", err)
		}
		if _, err := tx.Exec(`
			INSERT INTO daily_tasks(user_id,activity_date,user_role,messages,verdict,signals,safe_action)
			VALUES(?,(CURRENT_TIMESTAMP AT TIME ZONE 'Europe/Moscow')::date,'seller',
				'[{"role":"assistant","text":"Я уже оплатил смартфон. Отдайте товар курьеру по скриншоту."},{"role":"user","text":"В приложении поступление денег пока не отображается."}]'::jsonb,
				TRUE,'["Скриншот не подтверждает оплату","Давление с передачей товара"]'::jsonb,
				'Проверить оплату в собственном интерфейсе банка и заказа.')`, userID); err != nil {
			return fmt.Errorf("seed daily task: %w", err)
		}

		var completedAttemptID int
		if _, err := tx.QueryOne(pg.Scan(&completedAttemptID), `
			INSERT INTO chat_sessions(user_id,chat_id,status,started_at,finished_at,score,mode,current_step_number,max_score,user_role,free_text_count)
			VALUES(?,?,'COMPLETED',CURRENT_TIMESTAMP-INTERVAL '1 day',CURRENT_TIMESTAMP-INTERVAL '1 day',75,'scenario',3,100,'seller',0)
			RETURNING id`, userID, levelOneScenarioID); err != nil {
			return fmt.Errorf("seed completed attempt: %w", err)
		}
		if _, err := tx.Exec(`
			INSERT INTO user_level_progress(user_id,level_id,user_role,best_score,stars,attempts,passed_at,topic_id)
			VALUES(?,?,'seller',75,2,1,CURRENT_TIMESTAMP-INTERVAL '1 day',?)`, userID, levelOneID, topicID); err != nil {
			return fmt.Errorf("seed level progress: %w", err)
		}
		if _, err := tx.Exec(`
			INSERT INTO user_achievements(user_id,achievement_id,received_at)
			SELECT ?,id,CURRENT_TIMESTAMP-INTERVAL '1 day' FROM achievements WHERE code='first_training'`, userID); err != nil {
			return fmt.Errorf("seed achievement: %w", err)
		}
		if _, err := tx.Exec(`
			INSERT INTO attempt_results(attempt_id,result)
			VALUES(?,jsonb_build_object(
				'attempt_id', ?, 'score', 75, 'stars', 2,
				'decision_review', '[]'::jsonb, 'risk_signals', '[]'::jsonb,
				'safe_actions', jsonb_build_array('Проверять оплату в собственном интерфейсе сделки.'),
				'level_progress', jsonb_build_object('number',1,'opened',true,'best_score',75,'stars',2,'attempts',1,'last_attempt_id',?),
				'topic_id', ?, 'topic_completed', false,
				'next_action', jsonb_build_object('type','start_level','topic_id',?,'level',2),
				'new_achievements', jsonb_build_array(jsonb_build_object('code','first_training','title','Первое прохождение','description','Завершить первое Прохождение.','icon','star','earned',true,'current',1,'target',1)),
				'streak', jsonb_build_object('current',2,'longest',3,'active_today',true)
			))`, completedAttemptID, completedAttemptID, completedAttemptID, topicID, topicID); err != nil {
			return fmt.Errorf("seed result: %w", err)
		}

		if _, err := tx.Exec(`
			INSERT INTO chat_sessions(user_id,chat_id,status,started_at,score,mode,current_step_number,max_score,user_role,free_text_count)
			VALUES(?,?,'IN_PROGRESS',CURRENT_TIMESTAMP,0,'scenario',1,0,'seller',0)`, userID, levelTwoScenarioID); err != nil {
			return fmt.Errorf("seed unfinished level 2 attempt: %w", err)
		}
		return nil
	})
}

func env(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
