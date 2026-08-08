package middleware

import (
	"anti-scam-trainer/backend/internal/core/logger"
	"net/http"

	"go.uber.org/zap"
)

func Logger(log *logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			requestLog := log.With(zap.String("request_id", request.Header.Get(RequestIDHeader)), zap.String("url", request.URL.String()))
			next.ServeHTTP(writer, request.WithContext(logger.WithContext(request.Context(), requestLog)))
		})
	}
}
