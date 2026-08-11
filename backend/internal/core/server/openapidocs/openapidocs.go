// Package openapidocs provides HTTP publication of versioned API contracts.
package openapidocs

import (
	"anti-scam-trainer/backend/openapi"
	"net/http"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	switch request.URL.Path {
	case "/openapi/v1.yaml":
		writer.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		if _, err := writer.Write(openapi.V1()); err != nil {
			return
		}
	case "/swagger/":
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := writer.Write([]byte(swaggerUI)); err != nil {
			return
		}
	default:
		http.NotFound(writer, request)
	}
}

const swaggerUI = `<!doctype html>
<html lang="ru">
  <head>
    <meta charset="utf-8">
    <title>Anti-scam trainer API</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
    <script>
      window.ui = SwaggerUIBundle({
        urls: [{ name: "v1", url: "/openapi/v1.yaml" }],
        "urls.primaryName": "v1",
        dom_id: "#swagger-ui",
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
        layout: "StandaloneLayout"
      });
    </script>
  </body>
</html>`
