package services

import (
	"anti-scam-trainer/backend/models"
	"anti-scam-trainer/backend/repositories"
)

func CreateChat(chat *models.Chat) (*models.Chat, error) {
	db := InitDB()
	defer db.Close()
	return repositories.CreateChat(db, chat)
}

func GetChatByID(id int) (*models.Chat, error) {
	db := InitDB()
	defer db.Close()
	return repositories.GetChatByID(db, id)
}

func UpdateChat(chat *models.Chat) error {
	db := InitDB()
	defer db.Close()
	return repositories.UpdateChat(db, chat)
}

func DeleteChat(id int) error {
	db := InitDB()
	defer db.Close()
	return repositories.DeleteChat(db, id)
}

func ListChats() ([]models.Chat, error) {
	db := InitDB()
	defer db.Close()
	return repositories.ListChats(db)
}
