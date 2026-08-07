package httputil

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}
