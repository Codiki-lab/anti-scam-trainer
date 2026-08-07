package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseAddress  string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	Port             string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		DatabaseAddress:  fmt.Sprintf("%s:%s", os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT")),
		DatabaseUser:     os.Getenv("POSTGRES_USER"),
		DatabasePassword: os.Getenv("POSTGRES_PASSWORD"),
		DatabaseName:     os.Getenv("POSTGRES_NAME"),
		Port:             port,
	}
}
