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
	OllamaURL                 string
	OllamaModel               string
	OllamaTimeout             time.Duration
	OllamaContextWindowTokens int
	OllamaOutputReserveTokens int
	OllamaMediumRiskThreshold float64
	OllamaHighRiskThreshold   float64
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		DatabaseAddress:           fmt.Sprintf("%s:%s", os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT")),
		DatabaseUser:              os.Getenv("POSTGRES_USER"),
		DatabasePassword:          os.Getenv("POSTGRES_PASSWORD"),
		DatabaseName:              os.Getenv("POSTGRES_NAME"),
		Port:                      port,
		OllamaURL:                 envString("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:               envString("OLLAMA_MODEL", "llama3.2:3b"),
		OllamaTimeout:             envDuration("OLLAMA_REQUEST_TIMEOUT", 30*time.Second),
		OllamaContextWindowTokens: envInt("OLLAMA_CONTEXT_WINDOW_TOKENS", 8192),
		OllamaOutputReserveTokens: envInt("OLLAMA_OUTPUT_RESERVE_TOKENS", 0),
		OllamaMediumRiskThreshold: envFloat("OLLAMA_MEDIUM_RISK_THRESHOLD", 0.60),
		OllamaHighRiskThreshold:   envFloat("OLLAMA_HIGH_RISK_THRESHOLD", 0.75),
	}
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
