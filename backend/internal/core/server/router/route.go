package router

import "net/http"

type Route struct {
	Path    string
	Handler http.HandlerFunc
}
