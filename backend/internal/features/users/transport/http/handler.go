package http

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/server/request"
	"anti-scam-trainer/backend/internal/core/server/response"
	"anti-scam-trainer/backend/internal/core/server/router"
	"anti-scam-trainer/backend/internal/features/users/service"
	"net/http"
)

type Handler struct{ service *service.Service }

type userDTO struct {
	ID             int    `json:"id"`
	ExternalID     string `json:"user_id"`
	Username       string `json:"username"`
	CompletedChats int    `json:"completed_chats"`
}

func New(service *service.Service) *Handler { return &Handler{service: service} }

func (h *Handler) Routes() []router.Route {
	return []router.Route{{Path: "/users", Handler: h.collection}, {Path: "/users/", Handler: h.item}}
}

func (h *Handler) collection(writer http.ResponseWriter, httpRequest *http.Request) {
	switch httpRequest.Method {
	case http.MethodGet:
		users, err := h.service.List()
		if err != nil {
			response.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		result := make([]userDTO, len(users))
		for index, user := range users {
			result[index] = fromDomain(user)
		}
		response.JSON(writer, result)
	case http.MethodPost:
		var input userDTO
		if err := request.DecodeJSON(httpRequest, &input); err != nil {
			response.Error(writer, "invalid JSON", http.StatusBadRequest)
			return
		}
		created, err := h.service.Create(toDomain(input))
		if err != nil {
			response.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.JSON(writer, fromDomain(created))
	default:
		response.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) item(writer http.ResponseWriter, httpRequest *http.Request) {
	id, ok := request.PathID(httpRequest.URL.Path, "/api/v1/users/")
	if !ok {
		response.Error(writer, "invalid user id", http.StatusBadRequest)
		return
	}
	switch httpRequest.Method {
	case http.MethodGet:
		user, err := h.service.GetByID(id)
		if err != nil {
			response.Error(writer, err.Error(), http.StatusNotFound)
			return
		}
		response.JSON(writer, fromDomain(user))
	case http.MethodPut:
		var input userDTO
		if err := request.DecodeJSON(httpRequest, &input); err != nil {
			response.Error(writer, "invalid JSON", http.StatusBadRequest)
			return
		}
		user := toDomain(input)
		user.ID = id
		if err := h.service.Update(user); err != nil {
			response.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.JSON(writer, fromDomain(user))
	case http.MethodDelete:
		if err := h.service.Delete(id); err != nil {
			response.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	default:
		response.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func toDomain(dto userDTO) domain.User {
	return domain.User{ID: dto.ID, ExternalID: dto.ExternalID, Username: dto.Username, CompletedChats: dto.CompletedChats}
}

func fromDomain(user domain.User) userDTO {
	return userDTO{ID: user.ID, ExternalID: user.ExternalID, Username: user.Username, CompletedChats: user.CompletedChats}
}
