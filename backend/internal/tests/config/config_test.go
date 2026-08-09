package config_test

import (
	"anti-scam-trainer/backend/internal/core/config"
	"strings"
	"testing"
)

func TestLoadRejectsInvalidRateLimitConfiguration(t *testing.T) {
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("ADMIN_USERNAME", "admin")
	t.Setenv("ADMIN_PASSWORD", "admin-password")
	t.Setenv("SWAGGER_USERNAME", "docs")
	t.Setenv("SWAGGER_PASSWORD", "docs-password")
	t.Setenv("RATE_LIMIT_LOGIN_CAPACITY", "0")
	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "RATE_LIMIT_LOGIN_CAPACITY") {
		t.Fatalf("Load() error=%v", err)
	}
}
