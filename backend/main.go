package main

import (
	"fmt"
    "github.com/go-pg/pg"
    "github.com/lpernett/godotenv"
	"anti-scam-trainer/backend/services"
    "os"
)

func initDB() *pg.DB {
    err := godotenv.Load()
    if err != nil {
        panic("Ошибка загрузки .env файла")
    }

    dbUser := os.Getenv("POSTGRES_USER")
    dbPassword := os.Getenv("POSTGRES_PASSWORD")
    dbName := os.Getenv("POSTGRES_NAME")
    dbHost := os.Getenv("POSTGRES_HOST")
    dbPort := os.Getenv("POSTGRES_PORT")

    db := pg.Connect(&pg.Options{
        Addr:     fmt.Sprintf("%s:%s", dbHost, dbPort),
        User:     dbUser,
        Password: dbPassword,
        Database: dbName,
    })

    return db
}

func main() {
    db := services.InitDB()
	fmt.Println(db)
}
