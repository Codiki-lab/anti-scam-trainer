// Package runtime owns the net/http server lifecycle.
package runtime

import (
	"anti-scam-trainer/backend/internal/core/server"
	"context"
	"errors"
	"net/http"
)

type Server struct {
	httpServer *http.Server
}

func New(config server.Config) *Server {
	return &Server{httpServer: &http.Server{Addr: config.Addr, Handler: config.Handler}}
}

func (s *Server) Run() error {
	err := s.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(context context.Context) error {
	return s.httpServer.Shutdown(context)
}
