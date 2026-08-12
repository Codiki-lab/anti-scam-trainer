// Package app is the composition root of the backend application.
package app

import (
	"anti-scam-trainer/backend/internal/core/aiprovider"
	"anti-scam-trainer/backend/internal/core/config"
	"anti-scam-trainer/backend/internal/core/logger"
	serverruntime "anti-scam-trainer/backend/internal/core/server/runtime"
	"fmt"
	"net/http"

	"github.com/go-pg/pg"
	"github.com/lpernett/godotenv"
)

type App struct {
	DB         *pg.DB
	Log        *logger.Logger
	Handler    http.Handler
	Port       string
	AIProvider aiprovider.Provider
	server     *serverruntime.Server
}

// New loads configuration and composes the application dependencies.
func New() (*App, error) {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	log, err := logger.New(cfg.LogLevel, cfg.LogFolder)
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}
	application, err := compose(cfg, log)
	if err != nil {
		_ = log.Close()
		return nil, err
	}
	return application, nil
}
