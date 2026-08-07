package main

import (
	"anti-scam-trainer/backend/app_builder"
	"log"
)

func main() {
	app, err := app_builder.NewApp()
	if err != nil {
		log.Fatalf("failed to build app: %v", err)
	}
	defer func() {
		if err := app.Close(); err != nil {
			log.Printf("failed to close app: %v", err)
		}
	}()

	if err := app.Run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
