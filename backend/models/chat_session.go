package models

import "time"

type ChatSession struct {
	ID         int       `json:"id" pg:"id,pk"`
	UserID     int       `json:"user_id" pg:"user_id,notnull"`
	ChatID     int       `json:"chat_id" pg:"chat_id,notnull"`
	Status     string    `json:"status" pg:"status,notnull"`
	StartedAt  time.Time `json:"started_at" pg:"started_at"`
	FinishedAt time.Time `json:"finished_at" pg:"finished_at"`
	Score      int       `json:"score" pg:"score"`
}
