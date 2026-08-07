package main

import (
	"anti-scam-trainer/backend/models"
	"anti-scam-trainer/backend/services"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	db := services.InitDB()
	if db != nil {
		defer db.Close()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "AntiScamTrainer backend is running")
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			users, err := services.ListUsers()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, users)
		case http.MethodPost:
			var user models.User
			if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			created, err := services.CreateUser(&user)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, created)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/users/")
		if path == "" {
			http.NotFound(w, r)
			return
		}

		id, err := strconv.Atoi(path)
		if err != nil {
			http.Error(w, "invalid user id", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			user, err := services.GetUserByID(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			writeJSON(w, user)
		case http.MethodPut:
			var user models.User
			if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			user.ID = id
			if err := services.UpdateUser(&user); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, user)
		case http.MethodDelete:
			if err := services.DeleteUser(id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/chats", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			chats, err := services.ListChats()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, chats)
		case http.MethodPost:
			var chat models.Chat
			if err := json.NewDecoder(r.Body).Decode(&chat); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			created, err := services.CreateChat(&chat)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, created)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/chats/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/chats/")
		if path == "" {
			http.NotFound(w, r)
			return
		}

		id, err := strconv.Atoi(path)
		if err != nil {
			http.Error(w, "invalid chat id", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			chat, err := services.GetChatByID(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			writeJSON(w, chat)
		case http.MethodPut:
			var chat models.Chat
			if err := json.NewDecoder(r.Body).Decode(&chat); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			chat.ID = id
			if err := services.UpdateChat(&chat); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, chat)
		case http.MethodDelete:
			if err := services.DeleteChat(id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/chat-sessions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			sessions, err := services.ListChatSessions()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, sessions)
		case http.MethodPost:
			var session models.ChatSession
			if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			created, err := services.CreateChatSession(&session)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, created)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/chat-sessions/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/chat-sessions/")
		if path == "" {
			http.NotFound(w, r)
			return
		}

		id, err := strconv.Atoi(path)
		if err != nil {
			http.Error(w, "invalid chat session id", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			session, err := services.GetChatSessionByID(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			writeJSON(w, session)
		case http.MethodPut:
			var session models.ChatSession
			if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			session.ID = id
			if err := services.UpdateChatSession(&session); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, session)
		case http.MethodDelete:
			if err := services.DeleteChatSession(id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}
