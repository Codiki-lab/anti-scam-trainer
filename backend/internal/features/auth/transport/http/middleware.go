package http

import (
	"anti-scam-trainer/backend/internal/core/server/response"
	auth "anti-scam-trainer/backend/internal/features/auth/service"
	"net/http"
	"strings"
)

func RequireAuthentication(tokens auth.Tokens) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if isPublicPath(request.URL.Path) {
				next.ServeHTTP(writer, request)
				return
			}
			cookie, err := request.Cookie(AccessTokenCookie)
			if err != nil {
				response.Error(writer, "unauthorized", http.StatusUnauthorized)
				return
			}
			identity, err := tokens.Parse(cookie.Value)
			if err != nil {
				response.Error(writer, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(writer, request.WithContext(auth.WithIdentity(request.Context(), identity)))
		})
	}
}

func isPublicPath(path string) bool {
	return path == "/api/v1/health" || path == "/api/v1/auth/register" || path == "/api/v1/auth/login" || path == "/api/v1/auth/logout" || path == "/swagger" || path == "/swagger/" || strings.HasPrefix(path, "/openapi/")
}
