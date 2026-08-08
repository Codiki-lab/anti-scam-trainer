package middleware

import (
	"crypto/subtle"
	"net/http"
)

func RequireSwaggerAuthentication(username, password string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			requestUsername, requestPassword, ok := request.BasicAuth()
			if !ok || subtle.ConstantTimeCompare([]byte(requestUsername), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(requestPassword), []byte(password)) != 1 {
				writer.Header().Set("WWW-Authenticate", `Basic realm="Swagger"`)
				http.Error(writer, "Swagger authentication required", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(writer, request)
		})
	}
}
