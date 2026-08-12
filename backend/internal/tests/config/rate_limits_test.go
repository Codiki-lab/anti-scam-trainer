package config_test

import (
	"anti-scam-trainer/backend/internal/core/config"
	"strings"
	"testing"
)

func TestLoadRejectsInvalidRateLimitConfiguration(t *testing.T) {
	setRequiredEnvironment(t)
	t.Setenv("RATE_LIMIT_LOGIN_CAPACITY", "0")
	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "RATE_LIMIT_LOGIN_CAPACITY") {
		t.Fatalf("Load() error=%v", err)
	}
}
