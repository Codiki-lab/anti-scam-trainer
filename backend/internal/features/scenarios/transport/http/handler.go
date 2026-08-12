package http

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/server/request"
	"anti-scam-trainer/backend/internal/core/server/response"
	"anti-scam-trainer/backend/internal/core/server/router"
	auth "anti-scam-trainer/backend/internal/features/auth/service"
	"anti-scam-trainer/backend/internal/features/scenarios/service"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct{ service *service.Service }

type adminScenarioDTO struct {
	ID             int                   `json:"id"`
	Title          string                `json:"title"`
	Description    string                `json:"description"`
	LevelID        int                   `json:"level_id"`
	TopicID        int                   `json:"topic_id"`
	Role           string                `json:"role"`
	Status         string                `json:"status"`
	ScamScheme     string                `json:"scam_scheme"`
	RiskType       string                `json:"risk_type"`
	ProductContext domain.ProductContext `json:"product_context"`
	AISystemPrompt string                `json:"ai_system_prompt"`
	FinalRubric    map[string]any        `json:"final_rubric"`
}
type adminStepDTO struct {
	ID                  int    `json:"id"`
	Number              int    `json:"number"`
	ResponseType        string `json:"response_type"`
	Goal                string `json:"goal"`
	CounterpartyMessage string `json:"counterparty_message"`
	MaxPoints           int    `json:"max_points"`
	AIInstruction       string `json:"ai_instruction"`
	FallbackMessage     string `json:"fallback_message"`
}
type adminOptionDTO struct {
	ID          int    `json:"id"`
	Text        string `json:"text"`
	Reaction    string `json:"counterparty_reaction,omitempty"`
	Explanation string `json:"explanation"`
	Points      int    `json:"points"`
	SortOrder   int    `json:"sort_order"`
}

func scenarioFromDTO(v adminScenarioDTO) domain.Scenario {
	return domain.Scenario{ID: v.ID, Title: v.Title, Description: v.Description, LevelID: v.LevelID, TopicID: v.TopicID, UserRole: domain.UserRole(v.Role), Status: domain.ScenarioStatus(v.Status), ScamScheme: v.ScamScheme, RiskType: domain.RiskType(v.RiskType), ProductContext: v.ProductContext, AISystemPrompt: v.AISystemPrompt, FinalRubric: domain.JSONObject(v.FinalRubric)}
}
func scenarioToDTO(v domain.Scenario) adminScenarioDTO {
	return adminScenarioDTO{ID: v.ID, Title: v.Title, Description: v.Description, LevelID: v.LevelID, TopicID: v.TopicID, Role: string(v.UserRole), Status: string(v.Status), ScamScheme: v.ScamScheme, RiskType: string(v.RiskType), ProductContext: v.ProductContext, AISystemPrompt: v.AISystemPrompt, FinalRubric: map[string]any(v.FinalRubric)}
}
func stepFromDTO(v adminStepDTO) domain.ScenarioStep {
	return domain.ScenarioStep{ID: v.ID, Number: v.Number, ResponseType: domain.ResponseType(v.ResponseType), Goal: v.Goal, CounterpartyMessage: v.CounterpartyMessage, MaxPoints: v.MaxPoints, AIInstruction: v.AIInstruction, FallbackMessage: v.FallbackMessage}
}
func stepToDTO(v domain.ScenarioStep) adminStepDTO {
	return adminStepDTO{ID: v.ID, Number: v.Number, ResponseType: string(v.ResponseType), Goal: v.Goal, CounterpartyMessage: v.CounterpartyMessage, MaxPoints: v.MaxPoints, AIInstruction: v.AIInstruction, FallbackMessage: v.FallbackMessage}
}
func optionFromDTO(v adminOptionDTO) domain.ScenarioOption {
	return domain.ScenarioOption{ID: v.ID, Text: v.Text, Reaction: v.Reaction, Explanation: v.Explanation, Points: v.Points, SortOrder: v.SortOrder}
}
func optionToDTO(v domain.ScenarioOption) adminOptionDTO {
	return adminOptionDTO{ID: v.ID, Text: v.Text, Reaction: v.Reaction, Explanation: v.Explanation, Points: v.Points, SortOrder: v.SortOrder}
}

func New(s *service.Service) *Handler { return &Handler{service: s} }
func (h *Handler) Routes() []router.Route {
	return []router.Route{{Path: "/admin/scenarios", Handler: h.scenarios}, {Path: "/admin/scenarios/", Handler: h.scenario}, {Path: "/admin/steps/", Handler: h.step}}
}
func admin(r *http.Request) bool {
	identity, ok := auth.IdentityFromContext(r.Context())
	return ok && identity.AccessRole == domain.AccessRoleAdmin
}
func (h *Handler) scenarios(w http.ResponseWriter, r *http.Request) {
	if !admin(r) {
		response.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if r.Method == http.MethodGet {
		scenarios, err := h.service.List()
		if err != nil {
			response.Error(w, "could not list scenarios", 500)
			return
		}
		result := make([]adminScenarioDTO, len(scenarios))
		for i, scenario := range scenarios {
			result[i] = scenarioToDTO(scenario)
		}
		response.JSON(w, result)
		return
	}
	if r.Method != http.MethodPost {
		response.Error(w, "method not allowed", 405)
		return
	}
	var input adminScenarioDTO
	if request.DecodeJSON(r, &input) != nil {
		response.Error(w, "invalid JSON", 400)
		return
	}
	created, err := h.service.Create(scenarioFromDTO(input))
	if err != nil {
		response.Error(w, "could not create scenario", 400)
		return
	}
	response.JSONStatus(w, scenarioToDTO(created), 201)
}
func (h *Handler) scenario(w http.ResponseWriter, r *http.Request) {
	if !admin(r) {
		response.Error(w, "forbidden", 403)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/scenarios/")
	parts := strings.Split(path, "/")
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		response.Error(w, "invalid scenario", 400)
		return
	}
	switch {
	case len(parts) == 1:
		switch r.Method {
		case http.MethodDelete:
			err = h.service.Archive(id)
		case http.MethodPut:
			var input adminScenarioDTO
			if request.DecodeJSON(r, &input) != nil {
				response.Error(w, "invalid JSON", 400)
				return
			}
			input.ID = id
			err = h.service.Update(scenarioFromDTO(input))
		default:
			response.Error(w, "method not allowed", 405)
			return
		}
	case len(parts) > 1 && parts[1] == "publish" && r.Method == http.MethodPost:
		err = h.service.Publish(id)
	case len(parts) > 1 && parts[1] == "deactivate" && r.Method == http.MethodPost:
		err = h.service.Deactivate(id)
	case len(parts) > 1 && parts[1] == "restore" && r.Method == http.MethodPost:
		err = h.service.Restore(id)
	case len(parts) > 1 && parts[1] == "steps" && r.Method == http.MethodPost:
		var input adminStepDTO
		if request.DecodeJSON(r, &input) != nil {
			response.Error(w, "invalid JSON", 400)
			return
		}
		step := stepFromDTO(input)
		step.ScenarioID = id
		created, createErr := h.service.AddStep(step)
		if createErr != nil {
			response.Error(w, "invalid content state", 409)
			return
		}
		response.JSONStatus(w, stepToDTO(created), 201)
		return
	case len(parts) == 3 && parts[1] == "steps":
		stepID, parseErr := strconv.Atoi(parts[2])
		if parseErr != nil {
			response.Error(w, "invalid step", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodDelete:
			err = h.service.DeleteStep(stepID)
		case http.MethodPut:
			var input adminStepDTO
			if request.DecodeJSON(r, &input) != nil {
				response.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			step := stepFromDTO(input)
			step.ID, step.ScenarioID = stepID, id
			err = h.service.UpdateStep(step)
		default:
			response.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	default:
		response.Error(w, "method not allowed", 405)
		return
	}
	if err != nil {
		response.Error(w, "invalid content state", 409)
		return
	}
	w.WriteHeader(204)
}
func (h *Handler) step(w http.ResponseWriter, r *http.Request) {
	if !admin(r) {
		response.Error(w, "forbidden", 403)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/steps/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "options" {
		response.Error(w, "method not allowed", 405)
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		response.Error(w, "invalid step", 400)
		return
	}
	if len(parts) == 2 && r.Method == http.MethodPost {
		var input adminOptionDTO
		if request.DecodeJSON(r, &input) != nil {
			response.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		option := optionFromDTO(input)
		option.StepID = id
		created, err := h.service.AddOption(option)
		if err != nil {
			response.Error(w, "invalid option", http.StatusConflict)
			return
		}
		response.JSONStatus(w, optionToDTO(created), http.StatusCreated)
		return
	}
	if len(parts) != 3 {
		response.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	optionID, err := strconv.Atoi(parts[2])
	if err != nil {
		response.Error(w, "invalid option", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodDelete:
		err = h.service.DeleteOption(optionID)
	case http.MethodPut:
		var input adminOptionDTO
		if request.DecodeJSON(r, &input) != nil {
			response.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		option := optionFromDTO(input)
		option.ID, option.StepID = optionID, id
		err = h.service.UpdateOption(option)
	default:
		response.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err != nil {
		response.Error(w, "invalid content state", http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
