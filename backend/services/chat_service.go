package services

import (
	"anti-scam-trainer/backend/models"
	"anti-scam-trainer/backend/repositories"

	"github.com/go-pg/pg"
)

func CreateChat(db *pg.DB, chat *models.Chat) (*models.Chat, error) {
	return repositories.CreateChat(db, chat)
}

func GetChatByID(db *pg.DB, id int) (*models.Chat, error) {
	return repositories.GetChatByID(db, id)
}

func UpdateChat(db *pg.DB, chat *models.Chat) error {
	return repositories.UpdateChat(db, chat)
}

func DeleteChat(db *pg.DB, id int) error {
	return repositories.DeleteChat(db, id)
}

func ListChats(db *pg.DB) ([]models.Chat, error) {
	return repositories.ListChats(db)
}
