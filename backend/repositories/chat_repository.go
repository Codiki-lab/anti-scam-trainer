package repositories

import (
	"anti-scam-trainer/backend/models"

	"github.com/go-pg/pg"
)

func CreateChat(db *pg.DB, chat *models.Chat) (*models.Chat, error) {
	_, err := db.Model(chat).Insert()
	if err != nil {
		return nil, err
	}
	return chat, nil
}

func GetChatByID(db *pg.DB, id int) (*models.Chat, error) {
	var chat models.Chat
	err := db.Model(&chat).Where("id = ?", id).Select()
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

func UpdateChat(db *pg.DB, chat *models.Chat) error {
	_, err := db.Model(chat).WherePK().Update()
	return err
}

func DeleteChat(db *pg.DB, id int) error {
	_, err := db.Model(&models.Chat{}).Where("id = ?", id).Delete()
	return err
}

func ListChats(db *pg.DB) ([]models.Chat, error) {
	var chats []models.Chat
	err := db.Model(&chats).Select()
	return chats, err
}
