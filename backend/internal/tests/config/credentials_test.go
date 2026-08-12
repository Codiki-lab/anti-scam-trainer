package config_test

import (
	"anti-scam-trainer/backend/internal/core/config"
	"strings"
	"testing"
)

func TestLoadRequiresSwaggerCredentials(t *testing.T) {
	setRequiredEnvironment(t)
	t.Setenv("SWAGGER_USERNAME", "")

	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "SWAGGER_USERNAME and SWAGGER_PASSWORD are required") {
		t.Fatalf("Load() error = %v, want Swagger credentials error", err)
	}

	t.Setenv("SWAGGER_USERNAME", "docs-user")
	if cfg, err := config.Load(); err != nil || cfg.SwaggerUsername != "docs-user" || cfg.SwaggerPassword != "docs-password" {
		t.Fatalf("Load() = (%#v, %v), want Swagger credentials", cfg, err)
	}
}
