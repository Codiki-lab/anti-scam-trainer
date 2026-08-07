package services

import (
	"anti-scam-trainer/backend/models"
	"anti-scam-trainer/backend/repositories"
)

func CreateUser(user *models.User) (*models.User, error) {
	db := InitDB()
	defer db.Close()
	return repositories.CreateUser(db, user)
}

func GetUserByID(id int) (*models.User, error) {
	db := InitDB()
	defer db.Close()
	return repositories.GetUserByID(db, id)
}

func GetUserByExternalID(externalID string) (*models.User, error) {
	db := InitDB()
	defer db.Close()
	return repositories.GetUserByExternalID(db, externalID)
}

func UpdateUser(user *models.User) error {
	db := InitDB()
	defer db.Close()
	return repositories.UpdateUser(db, user)
}

func DeleteUser(id int) error {
	db := InitDB()
	defer db.Close()
	return repositories.DeleteUser(db, id)
}

func ListUsers() ([]models.User, error) {
	db := InitDB()
	defer db.Close()
	return repositories.ListUsers(db)
}
