package app

import (
	"anti-scam-trainer/backend/internal/core/aiprovider"
	"anti-scam-trainer/backend/internal/core/config"
	"anti-scam-trainer/backend/internal/core/logger"
	"anti-scam-trainer/backend/internal/core/ratelimit"
	attemptsai "anti-scam-trainer/backend/internal/features/attempts/aiprovider"
	attemptsrepository "anti-scam-trainer/backend/internal/features/attempts/repository"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	authrepository "anti-scam-trainer/backend/internal/features/auth/repository"
	authservice "anti-scam-trainer/backend/internal/features/auth/service"
	learningai "anti-scam-trainer/backend/internal/features/learning/aiprovider"
	learningrepository "anti-scam-trainer/backend/internal/features/learning/repository"
	learningservice "anti-scam-trainer/backend/internal/features/learning/service"
	scenariosrepository "anti-scam-trainer/backend/internal/features/scenarios/repository"
	scenariosservice "anti-scam-trainer/backend/internal/features/scenarios/service"
	"fmt"
	"time"

	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type dependencies struct {
	authentication     *authservice.Service
	tokens             *authservice.JWTManager
	learning           *learningservice.Service
	learningContent    *learningservice.ContentService
	content            *scenariosservice.Service
	game               *attemptsservice.GameService
	registration       *ratelimit.Limiter
	login              *ratelimit.Limiter
	chatRecommendation *ratelimit.Limiter
	clientIP           *ratelimit.ClientIPResolver
}

func newDependencies(cfg config.Config, db *pg.DB, provider aiprovider.Provider, log *logger.Logger) (dependencies, error) {
	accounts := authservice.NewAccounts(authrepository.NewPostgres(db))
	if _, err := accounts.EnsureAdmin(cfg.AdminUsername, cfg.AdminPassword); err != nil {
		return dependencies{}, fmt.Errorf("bootstrap admin: %w", err)
	}
	tokens, err := authservice.NewJWTManager(cfg.JWTSecret)
	if err != nil {
		return dependencies{}, err
	}
	clientIP, err := ratelimit.NewClientIPResolver(cfg.TrustedProxyCIDRs)
	if err != nil {
		return dependencies{}, err
	}
	limits := newRateLimiters(cfg, log)
	localProviderGate := ratelimit.NewGate()
	learning := learningservice.NewWithDailyTaskGenerator(
		learningrepository.NewPostgres(db),
		learningai.NewDailyTaskGenerator(provider, limits.ai, localProviderGate),
	)
	modelAI := attemptsservice.NewModelAI(attemptsai.New(provider))
	game := attemptsservice.NewGameWithRateLimits(
		attemptsrepository.NewPostgres(db), modelAI, modelAI,
		limits.ai, limits.freePlay, localProviderGate,
	)
	return dependencies{
		authentication: authservice.New(accounts, tokens), tokens: tokens,
		learning: learning, learningContent: learningservice.NewContent(learningrepository.NewPostgres(db)),
		content: scenariosservice.New(scenariosrepository.NewPostgres(db)), game: game,
		registration: limits.registration, login: limits.login, chatRecommendation: limits.chatRecommendation, clientIP: clientIP,
	}, nil
}

type rateLimiters struct {
	registration       *ratelimit.Limiter
	login              *ratelimit.Limiter
	ai                 *ratelimit.Limiter
	freePlay           *ratelimit.Limiter
	chatRecommendation *ratelimit.Limiter
}

func newRateLimiters(cfg config.Config, log *logger.Logger) rateLimiters {
	limits := rateLimiters{
		registration:       ratelimit.New(ratelimit.Config{Limit: cfg.RegistrationRateLimit, Window: cfg.RegistrationRateWindow, MaxBuckets: cfg.RateLimitMaxBuckets, IdleTTL: cfg.RateLimitBucketTTL}, time.Now),
		login:              ratelimit.New(ratelimit.Config{Limit: cfg.LoginRateLimit, Window: cfg.LoginRateWindow, MaxBuckets: cfg.RateLimitMaxBuckets, IdleTTL: cfg.RateLimitBucketTTL}, time.Now),
		ai:                 ratelimit.New(ratelimit.Config{Limit: cfg.AIFreeTextRateLimit, Window: cfg.AIFreeTextRateWindow, MaxBuckets: cfg.RateLimitMaxBuckets, IdleTTL: cfg.RateLimitBucketTTL}, time.Now),
		freePlay:           ratelimit.New(ratelimit.Config{Limit: cfg.FreePlayRateLimit, Window: cfg.FreePlayRateWindow, MaxBuckets: cfg.RateLimitMaxBuckets, IdleTTL: cfg.RateLimitBucketTTL}, time.Now),
		chatRecommendation: ratelimit.New(ratelimit.Config{Limit: cfg.ChatRecommendationRateLimit, Window: cfg.ChatRecommendationRateWindow, MaxBuckets: cfg.RateLimitMaxBuckets, IdleTTL: cfg.RateLimitBucketTTL}, time.Now),
	}
	log.Info("rate limits configured", zap.Int("registration_capacity", cfg.RegistrationRateLimit), zap.Duration("registration_window", cfg.RegistrationRateWindow), zap.Int("login_capacity", cfg.LoginRateLimit), zap.Duration("login_window", cfg.LoginRateWindow), zap.Int("ai_capacity", cfg.AIFreeTextRateLimit), zap.Duration("ai_window", cfg.AIFreeTextRateWindow), zap.Int("free_play_capacity", cfg.FreePlayRateLimit), zap.Duration("free_play_window", cfg.FreePlayRateWindow), zap.Int("chat_recommendation_capacity", cfg.ChatRecommendationRateLimit), zap.Duration("chat_recommendation_window", cfg.ChatRecommendationRateWindow), zap.Int("max_buckets", cfg.RateLimitMaxBuckets), zap.Strings("trusted_proxy_cidrs", cfg.TrustedProxyCIDRs))
	return limits
}
