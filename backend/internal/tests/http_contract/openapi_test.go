package http_contract_test

import (
	"anti-scam-trainer/backend/internal/core/server/middleware"
	"anti-scam-trainer/backend/internal/core/server/openapidocs"
	"anti-scam-trainer/backend/internal/core/server/router"
	authhttp "anti-scam-trainer/backend/internal/features/auth/transport/http"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pb33f/libopenapi"
)

func TestOpenAPIDocumentationRequiresSeparateBasicAuthentication(t *testing.T) {
	handler := middleware.RequireSwaggerAuthentication("docs-user", "docs-password")(openapidocs.NewHandler())

	for _, path := range []string{"/openapi/v1.yaml", "/swagger/"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("%s without credentials = %d, want %d", path, recorder.Code, http.StatusUnauthorized)
		}
		if got := recorder.Header().Get("WWW-Authenticate"); got != `Basic realm="Swagger"` {
			t.Fatalf("%s challenge = %q, want Basic challenge", path, got)
		}

		invalid := httptest.NewRecorder()
		invalidRequest := httptest.NewRequest(http.MethodGet, path, nil)
		invalidRequest.Header.Set("Authorization", basicAuthorization("docs-user", "wrong-password"))
		handler.ServeHTTP(invalid, invalidRequest)
		if invalid.Code != http.StatusUnauthorized || invalid.Header().Get("WWW-Authenticate") != `Basic realm="Swagger"` {
			t.Fatalf("%s with invalid credentials = (%d, %q), want Basic challenge", path, invalid.Code, invalid.Header().Get("WWW-Authenticate"))
		}
	}

	specification := httptest.NewRecorder()
	specificationRequest := httptest.NewRequest(http.MethodGet, "/openapi/v1.yaml", nil)
	specificationRequest.Header.Set("Authorization", basicAuthorization("docs-user", "docs-password"))
	handler.ServeHTTP(specification, specificationRequest)
	if specification.Code != http.StatusOK || !strings.Contains(specification.Header().Get("Content-Type"), "application/yaml") || !strings.Contains(specification.Body.String(), "openapi: 3.1.0") {
		t.Fatalf("specification = (%d, %q, %q), want OpenAPI YAML", specification.Code, specification.Header().Get("Content-Type"), specification.Body.String())
	}

	ui := httptest.NewRecorder()
	uiRequest := httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	uiRequest.Header.Set("Authorization", basicAuthorization("docs-user", "docs-password"))
	handler.ServeHTTP(ui, uiRequest)
	if ui.Code != http.StatusOK || !strings.Contains(ui.Body.String(), `name: "v1"`) || !strings.Contains(ui.Body.String(), `url: "/openapi/v1.yaml"`) {
		t.Fatalf("Swagger UI = (%d, %q), want v1 selector configuration", ui.Code, ui.Body.String())
	}
	if !strings.Contains(ui.Body.String(), "swagger-ui-standalone-preset.js") || !strings.Contains(ui.Body.String(), "SwaggerUIStandalonePreset") {
		t.Fatal("Swagger UI does not include the standalone preset required to select a specification from urls")
	}
}

func TestSwaggerAuthenticationIsIndependentFromUserCookieAuthentication(t *testing.T) {
	routes := http.NewServeMux()
	documentationHandler := middleware.RequireSwaggerAuthentication("docs-user", "docs-password")(openapidocs.NewHandler())
	routes.Handle("/swagger", http.RedirectHandler("/swagger/", http.StatusTemporaryRedirect))
	routes.Handle("/swagger/", documentationHandler)
	routes.Handle("/openapi/", documentationHandler)
	routes.Handle("/", router.New())
	handler := authhttp.RequireAuthentication(fakeTokens{})(routes)

	request := httptest.NewRequest(http.MethodGet, "/openapi/v1.yaml", nil)
	request.Header.Set("Authorization", basicAuthorization("docs-user", "docs-password"))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("OpenAPI endpoint behind application authentication = %d, want %d", recorder.Code, http.StatusOK)
	}

	swaggerRedirect := httptest.NewRecorder()
	handler.ServeHTTP(swaggerRedirect, httptest.NewRequest(http.MethodGet, "/swagger", nil))
	if swaggerRedirect.Code != http.StatusTemporaryRedirect || swaggerRedirect.Header().Get("Location") != "/swagger/" {
		t.Fatalf("Swagger URL without trailing slash = (%d, %q), want redirect to /swagger/", swaggerRedirect.Code, swaggerRedirect.Header().Get("Location"))
	}

	protected := httptest.NewRecorder()
	handler.ServeHTTP(protected, httptest.NewRequest(http.MethodGet, "/api/v1/scenarios", nil))
	if protected.Code != http.StatusUnauthorized {
		t.Fatalf("protected API without cookie = %d, want %d", protected.Code, http.StatusUnauthorized)
	}
}

func TestOpenAPIV1SpecificationIsValid(t *testing.T) {
	handler := middleware.RequireSwaggerAuthentication("docs-user", "docs-password")(openapidocs.NewHandler())
	request := httptest.NewRequest(http.MethodGet, "/openapi/v1.yaml", nil)
	request.Header.Set("Authorization", basicAuthorization("docs-user", "docs-password"))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	document, err := libopenapi.NewDocument(recorder.Body.Bytes())
	if err != nil {
		t.Fatalf("parse OpenAPI v1 specification: %v", err)
	}
	if document.GetVersion() != "3.1.0" {
		t.Fatalf("OpenAPI version = %q, want 3.1.0", document.GetVersion())
	}
	if _, err := document.BuildV3Model(); err != nil {
		t.Fatalf("validate OpenAPI v1 specification: %v", err)
	}

	specification := recorder.Body.String()
	for _, path := range []string{"/api/v1/health:", "/api/v1/auth/register:", "/api/v1/profile/preferences:", "/api/v1/topics:", "/api/v1/topics/{id}/theory:", "/api/v1/topics/{id}/quiz/attempts:", "/api/v1/training/levels:", "/api/v1/training/levels/{level}/start:", "/api/v1/training/free-play/start:", "/api/v1/attempts/{id}:", "/api/v1/attempts/{id}/answers:", "/api/v1/attempts/{id}/result:", "/api/v1/attempts/{id}/abandon:", "/api/v1/progress:", "/api/v1/achievements:", "/api/v1/dashboard:", "/api/v1/admin/topics:", "/api/v1/admin/topics/{id}/theory-blocks:", "/api/v1/admin/topics/{id}/quiz-questions:", "/api/v1/admin/scenarios:", "/api/v1/admin/scenarios/{id}/publish:"} {
		if !strings.Contains(specification, "  "+path) {
			t.Fatalf("OpenAPI v1 does not document registered path %s", path)
		}
	}
	if !strings.Contains(specification, "name: access_token") || !strings.Contains(specification, "in: cookie") {
		t.Fatal("OpenAPI v1 does not declare the access_token cookie security scheme")
	}
	if !strings.Contains(specification, "X-Request-ID: { $ref: '#/components/headers/RequestID' }") {
		t.Fatal("OpenAPI v1 does not attach X-Request-ID to responses")
	}
	for _, schema := range []string{"AdminTopic:", "AdminTheoryBlock:", "AdminQuizQuestion:", "AdminQuizOption:", "DailyTask:", "AdminScenario:", "AdminStep:", "AdminOption:", "RecentAttempt:"} {
		if !strings.Contains(specification, "    "+schema) {
			t.Fatalf("OpenAPI v1 does not define %s", schema)
		}
	}
	for _, fragment := range []string{"RateLimited:", "RATE_LIMITED", "CONTENT_CONFLICT", "Retry-After:"} {
		if !strings.Contains(specification, fragment) {
			t.Fatalf("OpenAPI v1 does not document %q", fragment)
		}
	}
	if !strings.Contains(specification, "requestBody: { required: true, content: { application/json: { schema: { $ref: '#/components/schemas/AdminScenario' }") || !strings.Contains(specification, "counterparty_message: { type: string, minLength: 1, maxLength: 280 }") || !strings.Contains(specification, "counterparty_reaction: { type: string, minLength: 1, maxLength: 280") {
		t.Fatal("OpenAPI v1 does not fully document admin topic_id/counterparty_message DTOs")
	}
	for _, tag := range []string{"Состояние сервиса", "Аутентификация", "Сценарии", "Прохождения"} {
		if !strings.Contains(specification, "- name: "+tag) || !strings.Contains(specification, "tags: ["+tag+"]") {
			t.Fatalf("OpenAPI v1 does not declare and assign the %q Swagger tag", tag)
		}
	}
}

func basicAuthorization(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}
