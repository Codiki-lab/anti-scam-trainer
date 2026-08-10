// Package app is the composition root of the backend application.
package app

import (
	"anti-scam-trainer/backend/internal/core/aiprovider"
	"anti-scam-trainer/backend/internal/core/config"
	"anti-scam-trainer/backend/internal/core/logger"
	"anti-scam-trainer/backend/internal/core/postgres"
	"anti-scam-trainer/backend/internal/core/ratelimit"
	"anti-scam-trainer/backend/internal/core/server"
	"anti-scam-trainer/backend/internal/core/server/middleware"
	"anti-scam-trainer/backend/internal/core/server/openapidocs"
	"anti-scam-trainer/backend/internal/core/server/response"
	"anti-scam-trainer/backend/internal/core/server/router"
	serverruntime "anti-scam-trainer/backend/internal/core/server/runtime"
	attemptsrepository "anti-scam-trainer/backend/internal/features/attempts/repository"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	attemptshttp "anti-scam-trainer/backend/internal/features/attempts/transport/http"
	authrepository "anti-scam-trainer/backend/internal/features/auth/repository"
	authservice "anti-scam-trainer/backend/internal/features/auth/service"
	authhttp "anti-scam-trainer/backend/internal/features/auth/transport/http"
	learningrepository "anti-scam-trainer/backend/internal/features/learning/repository"
	learningservice "anti-scam-trainer/backend/internal/features/learning/service"
	learninghttp "anti-scam-trainer/backend/internal/features/learning/transport/http"
	scenariosrepository "anti-scam-trainer/backend/internal/features/scenarios/repository"
	scenariosservice "anti-scam-trainer/backend/internal/features/scenarios/service"
	scenarioshttp "anti-scam-trainer/backend/internal/features/scenarios/transport/http"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-pg/pg"
	"github.com/lpernett/godotenv"
	"go.uber.org/zap"
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
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
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

	accounts := authservice.NewAccounts(authrepository.NewPostgres(db))
	if _, err := accounts.EnsureAdmin(cfg.AdminUsername, cfg.AdminPassword); err != nil {
		return nil, fmt.Errorf("bootstrap admin: %w", err)
	}
	tokens, err := authservice.NewJWTManager(cfg.JWTSecret)
	if err != nil {
		return nil, err
	}
	authentication := authservice.New(accounts, tokens)
	clientIP, err := ratelimit.NewClientIPResolver(cfg.TrustedProxyCIDRs)
	if err != nil {
		return nil, err
	}
	registrationLimiter := ratelimit.New(ratelimit.Config{Limit: cfg.RegistrationRateLimit, Window: cfg.RegistrationRateWindow, MaxBuckets: cfg.RateLimitMaxBuckets, IdleTTL: cfg.RateLimitBucketTTL}, time.Now)
	loginLimiter := ratelimit.New(ratelimit.Config{Limit: cfg.LoginRateLimit, Window: cfg.LoginRateWindow, MaxBuckets: cfg.RateLimitMaxBuckets, IdleTTL: cfg.RateLimitBucketTTL}, time.Now)
	aiLimiter := ratelimit.New(ratelimit.Config{Limit: cfg.AIFreeTextRateLimit, Window: cfg.AIFreeTextRateWindow, MaxBuckets: cfg.RateLimitMaxBuckets, IdleTTL: cfg.RateLimitBucketTTL}, time.Now)
	freePlayLimiter := ratelimit.New(ratelimit.Config{Limit: cfg.FreePlayRateLimit, Window: cfg.FreePlayRateWindow, MaxBuckets: cfg.RateLimitMaxBuckets, IdleTTL: cfg.RateLimitBucketTTL}, time.Now)
	log.Info("rate limits configured", zap.Int("registration_capacity", cfg.RegistrationRateLimit), zap.Duration("registration_window", cfg.RegistrationRateWindow), zap.Int("login_capacity", cfg.LoginRateLimit), zap.Duration("login_window", cfg.LoginRateWindow), zap.Int("ai_capacity", cfg.AIFreeTextRateLimit), zap.Duration("ai_window", cfg.AIFreeTextRateWindow), zap.Int("free_play_capacity", cfg.FreePlayRateLimit), zap.Duration("free_play_window", cfg.FreePlayRateWindow), zap.Int("max_buckets", cfg.RateLimitMaxBuckets), zap.Strings("trusted_proxy_cidrs", cfg.TrustedProxyCIDRs))
	content := scenariosservice.New(scenariosrepository.NewPostgres(db))
	dailyAIGate := ratelimit.NewGate()
	learning := learningservice.NewWithDailyTaskGenerator(learningrepository.NewPostgres(db), dailyTaskAIAdapter{provider: provider, limiter: aiLimiter, gate: dailyAIGate})
	learningContent := learningservice.NewContent(learningrepository.NewPostgres(db))
	attemptRepository := attemptsrepository.NewPostgres(db)
	modelAI := attemptsservice.NewModelAI(gameAIAdapter{provider: provider})
	game := attemptsservice.NewGameWithRateLimits(attemptRepository, modelAI, modelAI, aiLimiter, freePlayLimiter, ratelimit.NewGate())
	versionedRouter := router.New()
	versionedRouter.Register(router.V1, []router.Route{{Path: "/health", Handler: health}})
	versionedRouter.Register(router.V1, authhttp.NewWithRateLimits(authentication, registrationLimiter, loginLimiter, clientIP).Routes())
	versionedRouter.Register(router.V1, learninghttp.New(learning).Routes())
	versionedRouter.Register(router.V1, learninghttp.NewAdmin(learningContent).Routes())
	versionedRouter.Register(router.V1, scenarioshttp.New(content).Routes())
	versionedRouter.Register(router.V1, attemptshttp.New(game).Routes())
	routes := http.NewServeMux()
	documentationHandler := middleware.RequireSwaggerAuthentication(cfg.SwaggerUsername, cfg.SwaggerPassword)(openapidocs.NewHandler())
	routes.Handle("/swagger", http.RedirectHandler("/swagger/", http.StatusTemporaryRedirect))
	routes.Handle("/swagger/", documentationHandler)
	routes.Handle("/openapi/", documentationHandler)
	routes.Handle("/", versionedRouter)
	handler := middleware.Chain(routes, middleware.RequestID(), middleware.CORS(cfg.FrontendOrigins), middleware.Logger(log), middleware.Panic(), middleware.Trace(), authhttp.RequireAuthentication(tokens))
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
