package http

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/server/request"
	"anti-scam-trainer/backend/internal/core/server/response"
	"anti-scam-trainer/backend/internal/core/server/router"
	"anti-scam-trainer/backend/internal/features/scenarios/service"
	"net/http"
)

type Handler struct{ service *service.Service }

type scenarioDTO struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Difficulty  string `json:"difficulty"`
	Role        string `json:"role"`
	IsActive    bool   `json:"is_active"`
}

func New(service *service.Service) *Handler { return &Handler{service: service} }

func (h *Handler) Routes() []router.Route {
	return []router.Route{{Path: "/scenarios", Handler: h.collection}, {Path: "/scenarios/", Handler: h.item}}
}

func (h *Handler) collection(writer http.ResponseWriter, httpRequest *http.Request) {
	switch httpRequest.Method {
	case http.MethodGet:
		scenarios, err := h.service.List()
		if err != nil {
			response.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		result := make([]scenarioDTO, len(scenarios))
		for index, scenario := range scenarios {
			result[index] = fromDomain(scenario)
		}
		response.JSON(writer, result)
	case http.MethodPost:
		var input scenarioDTO
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
	id, ok := request.PathID(httpRequest.URL.Path, "/api/v1/scenarios/")
	if !ok {
		response.Error(writer, "invalid scenario id", http.StatusBadRequest)
		return
	}
	switch httpRequest.Method {
	case http.MethodGet:
		scenario, err := h.service.GetByID(id)
		if err != nil {
			response.Error(writer, err.Error(), http.StatusNotFound)
			return
		}
		response.JSON(writer, fromDomain(scenario))
	case http.MethodPut:
		var input scenarioDTO
		if err := request.DecodeJSON(httpRequest, &input); err != nil {
			response.Error(writer, "invalid JSON", http.StatusBadRequest)
			return
		}
		scenario := toDomain(input)
		scenario.ID = id
		if err := h.service.Update(scenario); err != nil {
			response.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response.JSON(writer, fromDomain(scenario))
	case http.MethodDelete:
		if err := h.service.Delete(id); err != nil {
			response.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	default:
		response.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func toDomain(dto scenarioDTO) domain.Scenario {
	return domain.Scenario{ID: dto.ID, Title: dto.Title, Description: dto.Description, Level: dto.Difficulty, UserRole: dto.Role, IsActive: dto.IsActive}
}

func fromDomain(scenario domain.Scenario) scenarioDTO {
	return scenarioDTO{ID: scenario.ID, Title: scenario.Title, Description: scenario.Description, Difficulty: scenario.Level, Role: scenario.UserRole, IsActive: scenario.IsActive}
}
