package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseAddress              string
	DatabaseUser                 string
	DatabasePassword             string
	DatabaseName                 string
	Port                         string
	LogLevel                     string
	LogFolder                    string
	OllamaURL                    string
	OllamaModel                  string
	OllamaTimeout                time.Duration
	OllamaContextWindowTokens    int
	OllamaOutputReserveTokens    int
	OllamaMediumRiskThreshold    float64
	OllamaHighRiskThreshold      float64
	JWTSecret                    string
	AdminUsername                string
	AdminPassword                string
	SwaggerUsername              string
	SwaggerPassword              string
	FrontendOrigins              []string
	TrustedProxyCIDRs            []string
	RegistrationRateLimit        int
	RegistrationRateWindow       time.Duration
	LoginRateLimit               int
	LoginRateWindow              time.Duration
	AIFreeTextRateLimit          int
	AIFreeTextRateWindow         time.Duration
	FreePlayRateLimit            int
	FreePlayRateWindow           time.Duration
	ChatRecommendationRateLimit  int
	ChatRecommendationRateWindow time.Duration
	RateLimitMaxBuckets          int
	RateLimitBucketTTL           time.Duration
}

func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	cfg := Config{
		DatabaseAddress:           fmt.Sprintf("%s:%s", os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT")),
		DatabaseUser:              os.Getenv("POSTGRES_USER"),
		DatabasePassword:          os.Getenv("POSTGRES_PASSWORD"),
		DatabaseName:              os.Getenv("POSTGRES_NAME"),
		Port:                      port,
		LogLevel:                  envString("LOG_LEVEL", "debug"),
		LogFolder:                 envString("LOG_FOLDER", "out/logs"),
		OllamaURL:                 envString("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:               envString("OLLAMA_MODEL", "qwen3:8b"),
		OllamaTimeout:             envDuration("OLLAMA_REQUEST_TIMEOUT", 30*time.Second),
		OllamaContextWindowTokens: envInt("OLLAMA_CONTEXT_WINDOW_TOKENS", 8192),
		OllamaOutputReserveTokens: envInt("OLLAMA_OUTPUT_RESERVE_TOKENS", 0),
		OllamaMediumRiskThreshold: envFloat("OLLAMA_MEDIUM_RISK_THRESHOLD", 0.60),
		OllamaHighRiskThreshold:   envFloat("OLLAMA_HIGH_RISK_THRESHOLD", 0.75),
		JWTSecret:                 os.Getenv("JWT_SECRET"),
		AdminUsername:             os.Getenv("ADMIN_USERNAME"),
		AdminPassword:             os.Getenv("ADMIN_PASSWORD"),
		SwaggerUsername:           os.Getenv("SWAGGER_USERNAME"),
		SwaggerPassword:           os.Getenv("SWAGGER_PASSWORD"),
		FrontendOrigins:           envCSV("FRONTEND_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173"),
		TrustedProxyCIDRs:         envCSV("TRUSTED_PROXY_CIDRS", ""),
	}
	var rateErr error
	if cfg.RegistrationRateLimit, rateErr = positiveEnvInt("RATE_LIMIT_REGISTER_CAPACITY", 5); rateErr != nil {
		return Config{}, rateErr
	}
	if cfg.RegistrationRateWindow, rateErr = positiveEnvDuration("RATE_LIMIT_REGISTER_WINDOW", 10*time.Minute); rateErr != nil {
		return Config{}, rateErr
	}
	if cfg.LoginRateLimit, rateErr = positiveEnvInt("RATE_LIMIT_LOGIN_CAPACITY", 10); rateErr != nil {
		return Config{}, rateErr
	}
	if cfg.LoginRateWindow, rateErr = positiveEnvDuration("RATE_LIMIT_LOGIN_WINDOW", 10*time.Minute); rateErr != nil {
		return Config{}, rateErr
	}
	if cfg.AIFreeTextRateLimit, rateErr = positiveEnvInt("RATE_LIMIT_AI_CAPACITY", 10); rateErr != nil {
		return Config{}, rateErr
	}
	if cfg.AIFreeTextRateWindow, rateErr = positiveEnvDuration("RATE_LIMIT_AI_WINDOW", time.Minute); rateErr != nil {
		return Config{}, rateErr
	}
	if cfg.FreePlayRateLimit, rateErr = positiveEnvInt("RATE_LIMIT_FREE_PLAY_CAPACITY", 5); rateErr != nil {
		return Config{}, rateErr
	}
	if cfg.FreePlayRateWindow, rateErr = positiveEnvDuration("RATE_LIMIT_FREE_PLAY_WINDOW", time.Minute); rateErr != nil {
		return Config{}, rateErr
	}
	if cfg.ChatRecommendationRateLimit, rateErr = positiveEnvInt("RATE_LIMIT_CHAT_RECOMMENDATION_CAPACITY", 20); rateErr != nil {
		return Config{}, rateErr
	}
	if cfg.ChatRecommendationRateWindow, rateErr = positiveEnvDuration("RATE_LIMIT_CHAT_RECOMMENDATION_WINDOW", time.Minute); rateErr != nil {
		return Config{}, rateErr
	}
	if cfg.RateLimitMaxBuckets, rateErr = positiveEnvInt("RATE_LIMIT_MAX_BUCKETS", 10000); rateErr != nil {
		return Config{}, rateErr
	}
	if cfg.RateLimitBucketTTL, rateErr = positiveEnvDuration("RATE_LIMIT_BUCKET_TTL", 20*time.Minute); rateErr != nil {
		return Config{}, rateErr
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.AdminUsername == "" || cfg.AdminPassword == "" {
		return Config{}, fmt.Errorf("ADMIN_USERNAME and ADMIN_PASSWORD are required")
	}
	if cfg.SwaggerUsername == "" || cfg.SwaggerPassword == "" {
		return Config{}, fmt.Errorf("SWAGGER_USERNAME and SWAGGER_PASSWORD are required")
	}
	return cfg, nil
}

func positiveEnvInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return parsed, nil
}
func positiveEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", key)
	}
	return parsed, nil
}

func envCSV(key, fallback string) []string {
	value := envString(key, fallback)
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if item := strings.TrimSpace(part); item != "" {
			result = append(result, item)
		}
	}
	return result
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(os.Getenv(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}

func envFloat(key string, fallback float64) float64 {
	value, err := strconv.ParseFloat(os.Getenv(key), 64)
	if err != nil {
		return fallback
	}
	return value
}
