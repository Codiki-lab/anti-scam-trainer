package http

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/httputil"
	"encoding/json"
	"net/http"
)

type Service interface {
	Create(domain.User) (domain.User, error)
	GetByID(int) (domain.User, error)
	Update(domain.User) error
	Delete(int) error
	List() ([]domain.User, error)
}
type Handler struct{ service Service }
type userDTO struct {
	ID             int    `json:"id"`
	ExternalID     string `json:"user_id"`
	Username       string `json:"username"`
	CompletedChats int    `json:"completed_chats"`
}

func New(service Service) *Handler { return &Handler{service: service} }
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/users", h.collection)
	mux.HandleFunc("/users/", h.item)
}
func (h *Handler) collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		users, err := h.service.List()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		result := make([]userDTO, len(users))
		for i, user := range users {
			result[i] = fromDomain(user)
		}
		httputil.WriteJSON(w, result)
	case http.MethodPost:
		var request userDTO
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
	id, ok := httputil.PathID(w, r, "/users/", "user")
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		user, err := h.service.GetByID(id)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		httputil.WriteJSON(w, fromDomain(user))
	case http.MethodPut:
		var request userDTO
		if json.NewDecoder(r.Body).Decode(&request) != nil {
			http.Error(w, "invalid JSON", 400)
			return
		}
		user := toDomain(request)
		user.ID = id
		if err := h.service.Update(user); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		httputil.WriteJSON(w, fromDomain(user))
	case http.MethodDelete:
		if err := h.service.Delete(id); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(200)
	default:
		http.Error(w, "method not allowed", 405)
	}
}
func toDomain(dto userDTO) domain.User {
	return domain.User{ID: dto.ID, ExternalID: dto.ExternalID, Username: dto.Username, CompletedChats: dto.CompletedChats}
}
func fromDomain(user domain.User) userDTO {
	return userDTO{ID: user.ID, ExternalID: user.ExternalID, Username: user.Username, CompletedChats: user.CompletedChats}
}
