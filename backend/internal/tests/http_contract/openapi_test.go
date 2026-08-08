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
}

func TestSwaggerAuthenticationIsIndependentFromUserCookieAuthentication(t *testing.T) {
	routes := http.NewServeMux()
	documentationHandler := middleware.RequireSwaggerAuthentication("docs-user", "docs-password")(openapidocs.NewHandler())
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
	for _, path := range []string{"/api/v1/health:", "/api/v1/auth/register:", "/api/v1/auth/login:", "/api/v1/auth/logout:", "/api/v1/auth/me:", "/api/v1/scenarios:", "/api/v1/scenarios/{id}:", "/api/v1/attempts:", "/api/v1/attempts/{id}:"} {
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
}

func basicAuthorization(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}
