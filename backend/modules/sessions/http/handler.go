package http

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/httputil"
	"encoding/json"
	"net/http"
	"time"
)

type Service interface {
	Create(domain.Attempt) (domain.Attempt, error)
	GetByID(int) (domain.Attempt, error)
	Update(domain.Attempt) error
	Delete(int) error
	List() ([]domain.Attempt, error)
}
type Handler struct{ service Service }
type sessionDTO struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	ChatID     int       `json:"chat_id"`
	Status     string    `json:"status"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	Score      int       `json:"score"`
}

func New(service Service) *Handler { return &Handler{service: service} }
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/chat-sessions", h.collection)
	mux.HandleFunc("/chat-sessions/", h.item)
}
func (h *Handler) collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sessions, err := h.service.List()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		result := make([]sessionDTO, len(sessions))
		for i, session := range sessions {
			result[i] = fromDomain(session)
		}
		httputil.WriteJSON(w, result)
	case http.MethodPost:
		var request sessionDTO
		if json.NewDecoder(r.Body).Decode(&request) != nil {
			http.Error(w, "invalid JSON", 400)
			return
		}
		created, err := h.service.Create(toDomain(request))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		httputil.WriteJSON(w, fromDomain(created))
	default:
		http.Error(w, "method not allowed", 405)
	}
}
func (h *Handler) item(w http.ResponseWriter, r *http.Request) {
	id, ok := httputil.PathID(w, r, "/chat-sessions/", "chat session")
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		session, err := h.service.GetByID(id)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		httputil.WriteJSON(w, fromDomain(session))
	case http.MethodPut:
		var request sessionDTO
		if json.NewDecoder(r.Body).Decode(&request) != nil {
			http.Error(w, "invalid JSON", 400)
			return
		}
		session := toDomain(request)
		session.ID = id
		if err := h.service.Update(session); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		httputil.WriteJSON(w, fromDomain(session))
	case http.MethodDelete:
		if err := h.service.Delete(id); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	default:
		http.Error(w, "method not allowed", 405)
	}
}
func toDomain(dto sessionDTO) domain.Attempt {
	return domain.Attempt{ID: dto.ID, UserID: dto.UserID, ChatID: dto.ChatID, Status: dto.Status, StartedAt: dto.StartedAt, FinishedAt: dto.FinishedAt, Score: dto.Score}
}
func fromDomain(attempt domain.Attempt) sessionDTO {
	return sessionDTO{ID: attempt.ID, UserID: attempt.UserID, ChatID: attempt.ChatID, Status: attempt.Status, StartedAt: attempt.StartedAt, FinishedAt: attempt.FinishedAt, Score: attempt.Score}
}
