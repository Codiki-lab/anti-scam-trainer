// Package response writes shared HTTP responses.
package response

import (
	"encoding/json"
	"net/http"
)

func JSON(writer http.ResponseWriter, payload any) {
	writer.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(writer).Encode(payload)
}
