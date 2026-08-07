package services

import (
	"anti-scam-trainer/backend/models"
	"anti-scam-trainer/backend/repositories"
)

func CreateChatSession(session *models.ChatSession) (*models.ChatSession, error) {
	db := InitDB()
	defer db.Close()
	return repositories.CreateChatSession(db, session)
}

func GetChatSessionByID(id int) (*models.ChatSession, error) {
	db := InitDB()
	defer db.Close()
	return repositories.GetChatSessionByID(db, id)
}

func UpdateChatSession(session *models.ChatSession) error {
	db := InitDB()
	defer db.Close()
	return repositories.UpdateChatSession(db, session)
}

func DeleteChatSession(id int) error {
	db := InitDB()
	defer db.Close()
	return repositories.DeleteChatSession(db, id)
}

func ListChatSessions() ([]models.ChatSession, error) {
	db := InitDB()
	defer db.Close()
	return repositories.ListChatSessions(db)
}
