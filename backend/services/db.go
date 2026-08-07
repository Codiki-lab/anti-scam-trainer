package services

import (
	"fmt"
	"os"

	"github.com/go-pg/pg"
	"github.com/lpernett/godotenv"
)

func InitDB() *pg.DB {
	_ = godotenv.Load()

	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_NAME")
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")

	return pg.Connect(&pg.Options{
		Addr:     fmt.Sprintf("%s:%s", dbHost, dbPort),
		User:     dbUser,
		Password: dbPassword,
		Database: dbName,
	})
}
