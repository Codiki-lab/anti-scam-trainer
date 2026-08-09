// Package router provides versioned registration for HTTP product routes.
package router

import (
	"anti-scam-trainer/backend/internal/core/server/response"
	"net/http"
)

type Router struct {
	mux *http.ServeMux
}

func New() *Router {
	return &Router{mux: http.NewServeMux()}
}

func (r *Router) Register(version Version, routes []Route) {
	for _, route := range routes {
		r.mux.HandleFunc("/api/"+string(version)+route.Path, route.Handler)
	}
}

func (r *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	handler, pattern := r.mux.Handler(request)
	if pattern == "" {
		response.Error(writer, "not found", http.StatusNotFound)
		return
	}
	handler.ServeHTTP(writer, request)
}
