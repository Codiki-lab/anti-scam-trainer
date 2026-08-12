package config_test

import "testing"

func setRequiredEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("ADMIN_USERNAME", "admin")
	t.Setenv("ADMIN_PASSWORD", "admin-password")
	t.Setenv("SWAGGER_USERNAME", "docs-user")
	t.Setenv("SWAGGER_PASSWORD", "docs-password")
}
