// Package response writes shared HTTP responses.
package response

import (
	"encoding/json"
	"net/http"
)

func JSON(writer http.ResponseWriter, payload any) {
	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(payload); err != nil {
		return
	}
}

func JSONStatus(writer http.ResponseWriter, payload any, status int) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	if err := json.NewEncoder(writer).Encode(payload); err != nil {
		return
	}
}
