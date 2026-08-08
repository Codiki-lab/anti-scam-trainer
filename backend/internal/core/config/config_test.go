package config

import (
	"strings"
	"testing"
)

func TestLoadRequiresSwaggerCredentials(t *testing.T) {
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("ADMIN_USERNAME", "admin")
	t.Setenv("ADMIN_PASSWORD", "admin-password")
	t.Setenv("SWAGGER_USERNAME", "")
	t.Setenv("SWAGGER_PASSWORD", "docs-password")

	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "SWAGGER_USERNAME and SWAGGER_PASSWORD are required") {
		t.Fatalf("Load() error = %v, want Swagger credentials error", err)
	}

	t.Setenv("SWAGGER_USERNAME", "docs-user")
	if cfg, err := Load(); err != nil || cfg.SwaggerUsername != "docs-user" || cfg.SwaggerPassword != "docs-password" {
		t.Fatalf("Load() = (%#v, %v), want Swagger credentials", cfg, err)
	}
}
