package services

import (
	"anti-scam-trainer/backend/models"
	"anti-scam-trainer/backend/repositories"

	"github.com/go-pg/pg"
)

func CreateUser(db *pg.DB, user *models.User) (*models.User, error) {
	return repositories.CreateUser(db, user)
}

func GetUserByID(db *pg.DB, id int) (*models.User, error) {
	return repositories.GetUserByID(db, id)
}

func GetUserByExternalID(db *pg.DB, externalID string) (*models.User, error) {
	return repositories.GetUserByExternalID(db, externalID)
}

func UpdateUser(db *pg.DB, user *models.User) error {
	return repositories.UpdateUser(db, user)
}

func DeleteUser(db *pg.DB, id int) error {
	return repositories.DeleteUser(db, id)
}

func ListUsers(db *pg.DB) ([]models.User, error) {
	return repositories.ListUsers(db)
}
