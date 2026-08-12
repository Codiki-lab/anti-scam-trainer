package app

import (
	"anti-scam-trainer/backend/internal/core/aiprovider"
	"anti-scam-trainer/backend/internal/core/config"
	"anti-scam-trainer/backend/internal/core/logger"
	"anti-scam-trainer/backend/internal/core/postgres"
	"anti-scam-trainer/backend/internal/core/server"
	serverruntime "anti-scam-trainer/backend/internal/core/server/runtime"
	"fmt"
)

func compose(cfg config.Config, log *logger.Logger) (*App, error) {
	provider, err := newAIProvider(cfg)
	if err != nil {
		return nil, err
	}
	db := postgres.Connect(cfg)
	if db == nil {
		return nil, fmt.Errorf("connect PostgreSQL")
	}
	initialized := false
	defer func() {
		if !initialized {
			_ = db.Close()
		}
	}()

	dependencies, err := newDependencies(cfg, db, provider, log)
	if err != nil {
		return nil, err
	}
	handler := newHandler(cfg, log, dependencies)
	application := &App{
		DB: db, Log: log, Handler: handler, Port: cfg.Port, AIProvider: provider,
		server: serverruntime.New(server.Config{Addr: ":" + cfg.Port, Handler: handler}),
	}
	initialized = true
	return application, nil
}

func newAIProvider(cfg config.Config) (aiprovider.Provider, error) {
	provider, err := aiprovider.NewOllama(aiprovider.Config{
		URL: cfg.OllamaURL, Model: cfg.OllamaModel, RequestTimeout: cfg.OllamaTimeout,
		ContextWindowTokens: cfg.OllamaContextWindowTokens, OutputReserveTokens: cfg.OllamaOutputReserveTokens,
		MediumRiskThreshold: cfg.OllamaMediumRiskThreshold, HighRiskThreshold: cfg.OllamaHighRiskThreshold,
	})
	if err != nil {
		return nil, fmt.Errorf("create AI provider: %w", err)
	}
	return provider, nil
}
