// Package app is the composition root of the backend application.
package app

import (
	"anti-scam-trainer/backend/internal/core/aiprovider"
	"anti-scam-trainer/backend/internal/core/config"
	"anti-scam-trainer/backend/internal/core/logger"
	"anti-scam-trainer/backend/internal/core/postgres"
	"anti-scam-trainer/backend/internal/core/server"
	"anti-scam-trainer/backend/internal/core/server/middleware"
	"anti-scam-trainer/backend/internal/core/server/response"
	"anti-scam-trainer/backend/internal/core/server/router"
	serverruntime "anti-scam-trainer/backend/internal/core/server/runtime"
	attemptsrepository "anti-scam-trainer/backend/internal/features/attempts/repository"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	attemptshttp "anti-scam-trainer/backend/internal/features/attempts/transport/http"
	scenariosrepository "anti-scam-trainer/backend/internal/features/scenarios/repository"
	scenariosservice "anti-scam-trainer/backend/internal/features/scenarios/service"
	scenarioshttp "anti-scam-trainer/backend/internal/features/scenarios/transport/http"
	usersrepository "anti-scam-trainer/backend/internal/features/users/repository"
	usersservice "anti-scam-trainer/backend/internal/features/users/service"
	usershttp "anti-scam-trainer/backend/internal/features/users/transport/http"
	"context"
	"fmt"
	"net/http"
	"time"

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

func New() (*App, error) {
	_ = godotenv.Load()
	cfg := config.Load()
	log, err := logger.New(cfg.LogLevel, cfg.LogFolder)
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}
	initialized := false
	defer func() {
		if !initialized {
			_ = log.Close()
		}
	}()
	provider, err := aiprovider.NewOllama(aiprovider.Config{
		URL: cfg.OllamaURL, Model: cfg.OllamaModel, RequestTimeout: cfg.OllamaTimeout,
		ContextWindowTokens: cfg.OllamaContextWindowTokens, OutputReserveTokens: cfg.OllamaOutputReserveTokens,
		MediumRiskThreshold: cfg.OllamaMediumRiskThreshold, HighRiskThreshold: cfg.OllamaHighRiskThreshold,
	})
	if err != nil {
		return nil, fmt.Errorf("create AI provider: %w", err)
	}
	db := postgres.Connect(cfg)
	if db == nil {
		return nil, fmt.Errorf("connect PostgreSQL")
	}

	users := usersservice.New(usersrepository.NewPostgres(db))
	scenarios := scenariosservice.New(scenariosrepository.NewPostgres(db))
	attemptRepository := attemptsrepository.NewPostgres(db)
	attempts := attemptsservice.New(attemptRepository, attemptRepository)
	versionedRouter := router.New()
	versionedRouter.Register(router.V1, []router.Route{{Path: "/health", Handler: health}})
	versionedRouter.Register(router.V1, usershttp.New(users).Routes())
	versionedRouter.Register(router.V1, scenarioshttp.New(scenarios).Routes())
	versionedRouter.Register(router.V1, attemptshttp.New(attempts).Routes())
	handler := middleware.Chain(versionedRouter, middleware.RequestID(), middleware.Logger(log), middleware.Panic(), middleware.Trace())
	app := &App{DB: db, Log: log, Handler: handler, Port: cfg.Port, AIProvider: provider}
	app.server = serverruntime.New(server.Config{Addr: ":" + cfg.Port, Handler: handler})
	initialized = true
	return app, nil
}

func (a *App) Run() error { return a.server.Run() }

func (a *App) Close() error {
	shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if a.server != nil {
		if err := a.server.Shutdown(shutdownContext); err != nil {
			return err
		}
	}
	if a.Log != nil {
		if err := a.Log.Close(); err != nil {
			return err
		}
	}
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}

func health(writer http.ResponseWriter, _ *http.Request) {
	response.JSON(writer, map[string]string{"status": "ok"})
}
