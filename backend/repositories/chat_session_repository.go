package repositories

import (
	"anti-scam-trainer/backend/models"

	"github.com/go-pg/pg"
)

func CreateChatSession(db *pg.DB, session *models.ChatSession) (*models.ChatSession, error) {
	_, err := db.Model(session).Insert()
	if err != nil {
		return nil, err
	}
	return session, nil
}

func GetChatSessionByID(db *pg.DB, id int) (*models.ChatSession, error) {
	var session models.ChatSession
	err := db.Model(&session).Where("id = ?", id).Select()
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func UpdateChatSession(db *pg.DB, session *models.ChatSession) error {
	_, err := db.Model(session).
		Column("user_id", "chat_id", "status", "started_at", "finished_at", "score").
		WherePK().
		Update()
	return err
}

func DeleteChatSession(db *pg.DB, id int) error {
	_, err := db.Model(&models.ChatSession{}).Where("id = ?", id).Delete()
	return err
}

func ListChatSessions(db *pg.DB) ([]models.ChatSession, error) {
	var sessions []models.ChatSession
	err := db.Model(&sessions).Select()
	return sessions, err
}
