// user_repository.go
package repositories

import (
    "github.com/go-pg/pg"
    "backend/models"
)

func CreateUser(db *pg.DB, user models.User) error {
    _, err := db.Model(&user).Insert()
    return err
}

func getUser(db *pg.DB, id int) (models.User, error) {
    var user User
    err := db.Model(&user).Where("id = ?", id).Select()
    return &user, err
}

func updateUser(db *pg.DB, user models.User) error {
    _, err := db.Model(&user).Where("id = ?", user.ID).Update()
    return err
}

func deleteUser(db *pg.DB, id int) error {
    _, err := db.Model(&model.User{}).Where("id = ?", id).Delete()
    return err
}