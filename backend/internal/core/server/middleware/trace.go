package middleware

import (
	"anti-scam-trainer/backend/internal/core/logger"
	"anti-scam-trainer/backend/internal/core/server/response"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func Trace() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			startedAt := time.Now()
			log := logger.FromContext(request.Context())
			log.Debug(">>> incoming HTTP request", zap.Time("time", startedAt.UTC()))
			responseWriter := response.NewWriter(writer)
			next.ServeHTTP(responseWriter, request)
			log.Debug("<<< done HTTP request", zap.Int("status_code", responseWriter.Status()), zap.Duration("latency", time.Since(startedAt)))
		})
	}
}
