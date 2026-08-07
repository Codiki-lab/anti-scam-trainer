package app_builder

import (
	"anti-scam-trainer/backend/internal/core/aiprovider"
	"anti-scam-trainer/backend/internal/core/config"
	"anti-scam-trainer/backend/internal/core/postgres"
	"anti-scam-trainer/backend/internal/core/transport/httpserver"
	chatshttp "anti-scam-trainer/backend/internal/features/chats/http"
	chatsrepository "anti-scam-trainer/backend/internal/features/chats/repository"
	chatsservice "anti-scam-trainer/backend/internal/features/chats/service"
	progressrepository "anti-scam-trainer/backend/internal/features/progress/repository"
	progressservice "anti-scam-trainer/backend/internal/features/progress/service"
	sessionshttp "anti-scam-trainer/backend/internal/features/sessions/http"
	sessionsrepository "anti-scam-trainer/backend/internal/features/sessions/repository"
	sessionsservice "anti-scam-trainer/backend/internal/features/sessions/service"
	usershttp "anti-scam-trainer/backend/internal/features/users/http"
	usersrepository "anti-scam-trainer/backend/internal/features/users/repository"
	usersservice "anti-scam-trainer/backend/internal/features/users/service"
	"fmt"
	"net/http"

	"github.com/go-pg/pg"
	"github.com/lpernett/godotenv"
)

type App struct {
	DB         *pg.DB
	Router     *http.ServeMux
	Port       string
	AIProvider aiprovider.Provider
}

func NewApp() (*App, error) {
	_ = godotenv.Load()
	cfg := config.Load()
	provider, err := aiprovider.NewOllama(aiprovider.Config{
		URL:                 cfg.OllamaURL,
		Model:               cfg.OllamaModel,
		RequestTimeout:      cfg.OllamaTimeout,
		ContextWindowTokens: cfg.OllamaContextWindowTokens,
		OutputReserveTokens: cfg.OllamaOutputReserveTokens,
		MediumRiskThreshold: cfg.OllamaMediumRiskThreshold,
		HighRiskThreshold:   cfg.OllamaHighRiskThreshold,
	})
	if err != nil {
		return nil, fmt.Errorf("create AI provider: %w", err)
	}
	db := postgres.Connect(cfg)
	if db == nil {
		return nil, fmt.Errorf("failed to connect to database")
	}

	users := usersservice.New(usersrepository.NewPostgres(db))
	chats := chatsservice.New(chatsrepository.NewPostgres(db))
	sessions := sessionsservice.New(sessionsrepository.NewPostgres(db))
	_ = progressservice.New(progressrepository.NewPostgres(db))
	return &App{DB: db, Router: httpserver.NewRouter(usershttp.New(users), chatshttp.New(chats), sessionshttp.New(sessions)), Port: cfg.Port, AIProvider: provider}, nil
}

func (a *App) Run() error { return http.ListenAndServe(":"+a.Port, a.Router) }
func (a *App) Close() error {
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}
