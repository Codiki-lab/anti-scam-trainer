package services

import (
	"anti-scam-trainer/backend/models"
	"anti-scam-trainer/backend/repositories"
	"errors"

	"github.com/go-pg/pg"
)

var ErrInvalidChatSessionStatusTransition = errors.New("invalid chat session status transition")

func CreateChatSession(db *pg.DB, session *models.ChatSession) (*models.ChatSession, error) {
	return repositories.CreateChatSession(db, session)
}

func GetChatSessionByID(db *pg.DB, id int) (*models.ChatSession, error) {
	return repositories.GetChatSessionByID(db, id)
}

func UpdateChatSession(db *pg.DB, session *models.ChatSession) error {
	current, err := repositories.GetChatSessionByID(db, session.ID)
	if err != nil {
		return err
	}

	if !isAllowedChatSessionStatusTransition(current.Status, session.Status) {
		return ErrInvalidChatSessionStatusTransition
	}

	return repositories.UpdateChatSession(db, session)
}

func isAllowedChatSessionStatusTransition(currentStatus, nextStatus string) bool {
	if currentStatus == nextStatus {
		return true
	}

	return currentStatus == "IN_PROGRESS" && (nextStatus == "COMPLETED" || nextStatus == "ABANDONED")
}

func DeleteChatSession(db *pg.DB, id int) error {
	return repositories.DeleteChatSession(db, id)
}

func ListChatSessions(db *pg.DB) ([]models.ChatSession, error) {
	return repositories.ListChatSessions(db)
}
