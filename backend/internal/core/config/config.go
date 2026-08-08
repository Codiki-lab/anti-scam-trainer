package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseAddress           string
	DatabaseUser              string
	DatabasePassword          string
	DatabaseName              string
	Port                      string
	LogLevel                  string
	LogFolder                 string
	OllamaURL                 string
	OllamaModel               string
	OllamaTimeout             time.Duration
	OllamaContextWindowTokens int
	OllamaOutputReserveTokens int
	OllamaMediumRiskThreshold float64
	OllamaHighRiskThreshold   float64
	JWTSecret                 string
	AdminUsername             string
	AdminPassword             string
	SwaggerUsername           string
	SwaggerPassword           string
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
		OllamaModel:               envString("OLLAMA_MODEL", "llama3.2:3b"),
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
