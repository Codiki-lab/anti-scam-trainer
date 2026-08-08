package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

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
