package services

import (
    "anti-scam-trainer/models"
    "anti-scam-trainer/repositories"
)

func CreateUser(user models.User) (*models.User, error) {
    db := InitDB()
    return repositories.CreateUser(db, user)
}
