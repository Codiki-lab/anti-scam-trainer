package repositories

import (
	"anti-scam-trainer/backend/models"

	"github.com/go-pg/pg"
)

func CreateUser(db *pg.DB, user *models.User) (*models.User, error) {
	_, err := db.Model(user).Insert()
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByID(db *pg.DB, id int) (*models.User, error) {
	var user models.User
	err := db.Model(&user).Where("id = ?", id).Select()
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByExternalID(db *pg.DB, externalID string) (*models.User, error) {
	var user models.User
	err := db.Model(&user).Where("user_id = ?", externalID).Select()
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUser(db *pg.DB, user *models.User) error {
	_, err := db.Model(user).WherePK().Update()
	return err
}

func DeleteUser(db *pg.DB, id int) error {
	_, err := db.Model(&models.User{}).Where("id = ?", id).Delete()
	return err
}

func ListUsers(db *pg.DB) ([]models.User, error) {
	var users []models.User
	err := db.Model(&users).Select()
	return users, err
}
