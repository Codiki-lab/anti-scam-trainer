package http

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/server/request"
	"anti-scam-trainer/backend/internal/core/server/response"
	"anti-scam-trainer/backend/internal/core/server/router"
	"anti-scam-trainer/backend/internal/features/attempts/service"
	auth "anti-scam-trainer/backend/internal/features/auth/service"
	"net/http"
	"time"
)

type Handler struct{ service *service.Service }

type attemptDTO struct {
	ID         int       `json:"id"`
	UserID     int       `json:"-"`
	ScenarioID int       `json:"scenario_id"`
	Status     string    `json:"status"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	Score      int       `json:"score"`
}

func New(service *service.Service) *Handler { return &Handler{service: service} }

func (h *Handler) Routes() []router.Route {
	return []router.Route{{Path: "/attempts", Handler: h.collection}, {Path: "/attempts/", Handler: h.item}}
}

func (h *Handler) collection(writer http.ResponseWriter, httpRequest *http.Request) {
	switch httpRequest.Method {
	case http.MethodGet:
		identity, ok := auth.IdentityFromContext(httpRequest.Context())
		if !ok {
			response.Error(writer, "unauthorized", http.StatusUnauthorized)
			return
		}
		attempts, err := h.service.ListForUser(identity.UserID)
		if err != nil {
			response.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		result := make([]attemptDTO, len(attempts))
		for index, attempt := range attempts {
			result[index] = fromDomain(attempt)
		}
		response.JSON(writer, result)
	case http.MethodPost:
		identity, ok := auth.IdentityFromContext(httpRequest.Context())
		if !ok {
			response.Error(writer, "unauthorized", http.StatusUnauthorized)
			return
		}
		var input attemptDTO
		if err := request.DecodeJSON(httpRequest, &input); err != nil {
			response.Error(writer, "invalid JSON", http.StatusBadRequest)
			return
		}
		created, err := h.service.CreateForUser(identity.UserID, toDomain(input))
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
	id, ok := request.PathID(httpRequest.URL.Path, "/api/v1/attempts/")
	if !ok {
		response.Error(writer, "invalid attempt id", http.StatusBadRequest)
		return
	}
	switch httpRequest.Method {
	case http.MethodGet:
		identity, ok := auth.IdentityFromContext(httpRequest.Context())
		if !ok {
			response.Error(writer, "unauthorized", http.StatusUnauthorized)
			return
		}
		attempt, err := h.service.GetByIDForUser(identity.UserID, id)
		if err != nil {
			response.Error(writer, err.Error(), http.StatusNotFound)
			return
		}
		response.JSON(writer, fromDomain(attempt))
	case http.MethodPut:
		identity, ok := auth.IdentityFromContext(httpRequest.Context())
		if !ok {
			response.Error(writer, "unauthorized", http.StatusUnauthorized)
			return
		}
		var input attemptDTO
		if err := request.DecodeJSON(httpRequest, &input); err != nil {
			response.Error(writer, "invalid JSON", http.StatusBadRequest)
			return
		}
		attempt := toDomain(input)
		attempt.ID = id
		if err := h.service.UpdateForUser(identity.UserID, attempt); err != nil {
			response.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.JSON(writer, fromDomain(attempt))
	case http.MethodDelete:
		identity, ok := auth.IdentityFromContext(httpRequest.Context())
		if !ok {
			response.Error(writer, "unauthorized", http.StatusUnauthorized)
			return
		}
		if err := h.service.DeleteForUser(identity.UserID, id); err != nil {
			response.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	default:
		response.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func toDomain(dto attemptDTO) domain.Attempt {
	return domain.Attempt{ID: dto.ID, UserID: dto.UserID, ScenarioID: dto.ScenarioID, Status: dto.Status, StartedAt: dto.StartedAt, FinishedAt: dto.FinishedAt, Score: dto.Score}
}

func fromDomain(attempt domain.Attempt) attemptDTO {
	return attemptDTO{ID: attempt.ID, UserID: attempt.UserID, ScenarioID: attempt.ScenarioID, Status: attempt.Status, StartedAt: attempt.StartedAt, FinishedAt: attempt.FinishedAt, Score: attempt.Score}
}
