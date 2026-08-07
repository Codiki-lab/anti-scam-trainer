package services

import (
	"anti-scam-trainer/backend/models"
	"anti-scam-trainer/backend/repositories"

	"github.com/go-pg/pg"
)

func CreateChatSession(db *pg.DB, session *models.ChatSession) (*models.ChatSession, error) {
	return repositories.CreateChatSession(db, session)
}

func GetChatSessionByID(db *pg.DB, id int) (*models.ChatSession, error) {
	return repositories.GetChatSessionByID(db, id)
}

func UpdateChatSession(db *pg.DB, session *models.ChatSession) error {
	return repositories.UpdateChatSession(db, session)
}

func DeleteChatSession(db *pg.DB, id int) error {
	return repositories.DeleteChatSession(db, id)
}

func ListChatSessions(db *pg.DB) ([]models.ChatSession, error) {
	return repositories.ListChatSessions(db)
}
