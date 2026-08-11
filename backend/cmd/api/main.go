package main

import (
	"anti-scam-trainer/backend/internal/core/app"
	"log"
)

func main() {
	app, err := app.New()
	if err != nil {
		log.Fatalf("failed to build app: %v", err)
	}
	defer func() {
		if err := app.Close(); err != nil {
			log.Printf("failed to close app: %v", err)
		}
	}()

	if err := app.Run(); err != nil {
		log.Printf("server failed: %v", err)
	}
}
