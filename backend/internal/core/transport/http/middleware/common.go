package middleware

import (
	"anti-scam-trainer/backend/internal/core/logger"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const RequestIDHeader = "X-Request-ID"

// RequestID preserves a caller-provided request ID or creates one for the request.
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			requestID := request.Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = uuid.NewString()
			}
			request.Header.Set(RequestIDHeader, requestID)
			writer.Header().Set(RequestIDHeader, requestID)
			next.ServeHTTP(writer, request)
		})
	}
}

// Logger attaches request metadata to a structured logger in the request context.
func Logger(log *logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			requestLog := log.With(
				zap.String("request_id", request.Header.Get(RequestIDHeader)),
				zap.String("url", request.URL.String()),
			)
			next.ServeHTTP(writer, request.WithContext(logger.WithContext(request.Context(), requestLog)))
		})
	}
}

// Panic converts an unexpected panic into a 500 response and records it.
func Panic() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.FromContext(request.Context()).Error("panic while handling HTTP request", zap.Any("panic", recovered))
					http.Error(writer, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(writer, request)
		})
	}
}

// Trace records the request start, completion status, and duration.
func Trace() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			startedAt := time.Now()
			log := logger.FromContext(request.Context())
			log.Debug(">>> incoming HTTP request", zap.Time("time", startedAt.UTC()))

			responseWriter := &statusWriter{ResponseWriter: writer}
			next.ServeHTTP(responseWriter, request)

			log.Debug("<<< done HTTP request", zap.Int("status_code", responseWriter.status()), zap.Duration("latency", time.Since(startedAt)))
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (writer *statusWriter) WriteHeader(statusCode int) {
	writer.statusCode = statusCode
	writer.ResponseWriter.WriteHeader(statusCode)
}

func (writer *statusWriter) Write(body []byte) (int, error) {
	if writer.statusCode == 0 {
		writer.WriteHeader(http.StatusOK)
	}
	return writer.ResponseWriter.Write(body)
}

func (writer *statusWriter) status() int {
	if writer.statusCode == 0 {
		return http.StatusOK
	}
	return writer.statusCode
}
