package aiprovider

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	URL                 string
	Model               string
	RequestTimeout      time.Duration
	ContextWindowTokens int
	OutputReserveTokens int
	MediumRiskThreshold float64
	HighRiskThreshold   float64
}

type Ollama struct {
	url                 string
	model               string
	client              *http.Client
	contextWindowTokens int
	outputReserveTokens int
	mediumRiskThreshold float64
	highRiskThreshold   float64
}

func NewOllama(config Config) (Provider, error) {
	if config.URL == "" {
		return nil, errors.New("ollama URL is required")
	}
	parsedURL, err := url.Parse(config.URL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("invalid ollama URL: %q", config.URL)
	}
	if config.Model == "" {
		return nil, errors.New("ollama model is required")
	}
	if config.ContextWindowTokens <= 0 {
		return nil, errors.New("ollama context window must be positive")
	}
	if config.RequestTimeout <= 0 {
		config.RequestTimeout = 30 * time.Second
	}
	if config.MediumRiskThreshold == 0 {
		config.MediumRiskThreshold = 0.60
	}
	if config.HighRiskThreshold == 0 {
		config.HighRiskThreshold = 0.75
	}
	if config.MediumRiskThreshold <= 0 || config.MediumRiskThreshold >= config.HighRiskThreshold || config.HighRiskThreshold >= 1 {
		return nil, errors.New("ollama context-risk thresholds must be ordered between 0 and 1")
	}
	if config.OutputReserveTokens == 0 {
		config.OutputReserveTokens = max(256, int(math.Ceil(float64(config.ContextWindowTokens)*0.20)))
	}
	if config.OutputReserveTokens <= 0 || config.OutputReserveTokens >= config.ContextWindowTokens {
		return nil, errors.New("ollama output reserve must be positive and smaller than the context window")
	}
	return &Ollama{
		url:                 strings.TrimRight(config.URL, "/"),
		model:               config.Model,
		client:              &http.Client{Timeout: config.RequestTimeout},
		contextWindowTokens: config.ContextWindowTokens,
		outputReserveTokens: config.OutputReserveTokens,
		mediumRiskThreshold: config.MediumRiskThreshold,
		highRiskThreshold:   config.HighRiskThreshold,
	}, nil
}
