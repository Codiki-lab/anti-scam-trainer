// Package server contains shared HTTP server configuration and infrastructure.
package server

import "net/http"

type Config struct {
	Addr    string
	Handler http.Handler
}
