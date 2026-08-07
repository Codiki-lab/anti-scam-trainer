package httpserver

import (
	"anti-scam-trainer/backend/internal/core/transport/httputil"
	chatshttp "anti-scam-trainer/backend/internal/features/chats/http"
	sessionshttp "anti-scam-trainer/backend/internal/features/sessions/http"
	usershttp "anti-scam-trainer/backend/internal/features/users/http"
	"fmt"
	"net/http"
)

func NewRouter(users *usershttp.Handler, chats *chatshttp.Handler, sessions *sessionshttp.Handler) *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "AntiScamTrainer backend is running") })
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { httputil.WriteJSON(w, map[string]string{"status": "ok"}) })
	users.RegisterRoutes(router)
	chats.RegisterRoutes(router)
	sessions.RegisterRoutes(router)
	return router
}
