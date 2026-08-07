package httputil

import (
	"net/http"
	"strconv"
	"strings"
)

func PathID(w http.ResponseWriter, r *http.Request, prefix, resource string) (int, bool) {
	path := strings.TrimPrefix(r.URL.Path, prefix)
	if path == "" {
		http.NotFound(w, r)
		return 0, false
	}
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "invalid "+resource+" id", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}
