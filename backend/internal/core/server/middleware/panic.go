package middleware

import (
	"anti-scam-trainer/backend/internal/core/logger"
	"anti-scam-trainer/backend/internal/core/server/response"
	"net/http"

	"go.uber.org/zap"
)

func Panic() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.FromContext(request.Context()).Error("panic while handling HTTP request", zap.Any("panic", recovered))
					response.Error(writer, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(writer, request)
		})
	}
}
