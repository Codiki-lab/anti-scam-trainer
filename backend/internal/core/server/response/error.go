package response

import (
	"net/http"
	"strings"
)

const requestIDHeader = "X-Request-ID"

// ErrorEnvelope is the single error shape exposed by API v1.
type ErrorEnvelope struct {
	Error APIError `json:"error"`
}

type APIError struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details"`
	RequestID string         `json:"request_id"`
}

func Error(writer http.ResponseWriter, message string, status int) {
	ErrorCode(writer, defaultErrorCode(message, status), message, status, nil)
}

func ErrorCode(writer http.ResponseWriter, code, message string, status int, details map[string]any) {
	if details == nil {
		details = map[string]any{}
	}
	JSONStatus(writer, ErrorEnvelope{Error: APIError{
		Code: code, Message: message, Details: details,
		RequestID: writer.Header().Get(requestIDHeader),
	}}, status)
}

func defaultErrorCode(message string, status int) string {
	if strings.EqualFold(message, "stale step") {
		return "STALE_STEP"
	}
	switch status {
	case http.StatusBadRequest:
		return "VALIDATION_ERROR"
	case http.StatusUnauthorized:
		return "AUTHENTICATION_REQUIRED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "STATE_CONFLICT"
	case http.StatusMethodNotAllowed:
		return "METHOD_NOT_ALLOWED"
	case http.StatusBadGateway:
		return "AI_INVALID_RESPONSE"
	case http.StatusServiceUnavailable:
		return "AI_UNAVAILABLE"
	default:
		return "INTERNAL_ERROR"
	}
}
