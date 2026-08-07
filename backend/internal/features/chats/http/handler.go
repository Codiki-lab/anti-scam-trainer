package http

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/transport/httputil"
	"encoding/json"
	"net/http"
)

type Service interface {
	Create(domain.Scenario) (domain.Scenario, error)
	GetByID(int) (domain.Scenario, error)
	Update(domain.Scenario) error
	Delete(int) error
	List() ([]domain.Scenario, error)
}
type Handler struct{ service Service }
type chatDTO struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Difficulty  string `json:"difficulty"`
	Role        string `json:"role"`
	IsActive    bool   `json:"is_active"`
}

func New(service Service) *Handler { return &Handler{service: service} }
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/chats", h.collection)
	mux.HandleFunc("/chats/", h.item)
}
func (h *Handler) collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		chats, err := h.service.List()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		result := make([]chatDTO, len(chats))
		for i, chat := range chats {
			result[i] = fromDomain(chat)
		}
		httputil.WriteJSON(w, result)
	case http.MethodPost:
		var request chatDTO
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
	id, ok := httputil.PathID(w, r, "/chats/", "chat")
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		chat, err := h.service.GetByID(id)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		httputil.WriteJSON(w, fromDomain(chat))
	case http.MethodPut:
		var request chatDTO
		if json.NewDecoder(r.Body).Decode(&request) != nil {
			http.Error(w, "invalid JSON", 400)
			return
		}
		chat := toDomain(request)
		chat.ID = id
		if err := h.service.Update(chat); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		httputil.WriteJSON(w, fromDomain(chat))
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
func toDomain(dto chatDTO) domain.Scenario {
	return domain.Scenario{ID: dto.ID, Title: dto.Title, Description: dto.Description, Level: dto.Difficulty, UserRole: dto.Role, IsActive: dto.IsActive}
}
func fromDomain(scenario domain.Scenario) chatDTO {
	return chatDTO{ID: scenario.ID, Title: scenario.Title, Description: scenario.Description, Difficulty: scenario.Level, Role: scenario.UserRole, IsActive: scenario.IsActive}
}
