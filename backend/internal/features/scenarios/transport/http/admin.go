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

type AdminHandler struct{ service *service.ContentService }

type adminScenarioDTO struct {
	ID             int            `json:"id"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	LevelID        int            `json:"level_id"`
	TopicID        int            `json:"topic_id"`
	Role           string         `json:"role"`
	Status         string         `json:"status"`
	ScamScheme     string         `json:"scam_scheme"`
	ProductContext map[string]any `json:"product_context"`
	AISystemPrompt string         `json:"ai_system_prompt"`
	FinalRubric    map[string]any `json:"final_rubric"`
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
	Explanation string `json:"explanation"`
	Points      int    `json:"points"`
	SortOrder   int    `json:"sort_order"`
}

func scenarioFromDTO(v adminScenarioDTO) domain.Scenario {
	return domain.Scenario{ID: v.ID, Title: v.Title, Description: v.Description, LevelID: v.LevelID, TopicID: v.TopicID, UserRole: v.Role, Status: v.Status, ScamScheme: v.ScamScheme, ProductContext: domain.JSONObject(v.ProductContext), AISystemPrompt: v.AISystemPrompt, FinalRubric: domain.JSONObject(v.FinalRubric)}
}
func scenarioToDTO(v domain.Scenario) adminScenarioDTO {
	return adminScenarioDTO{ID: v.ID, Title: v.Title, Description: v.Description, LevelID: v.LevelID, TopicID: v.TopicID, Role: v.UserRole, Status: v.Status, ScamScheme: v.ScamScheme, ProductContext: map[string]any(v.ProductContext), AISystemPrompt: v.AISystemPrompt, FinalRubric: map[string]any(v.FinalRubric)}
}
func stepFromDTO(v adminStepDTO) domain.ScenarioStep {
	return domain.ScenarioStep{ID: v.ID, Number: v.Number, ResponseType: domain.ResponseType(v.ResponseType), Goal: v.Goal, CounterpartyMessage: v.CounterpartyMessage, MaxPoints: v.MaxPoints, AIInstruction: v.AIInstruction, FallbackMessage: v.FallbackMessage}
}
func stepToDTO(v domain.ScenarioStep) adminStepDTO {
	return adminStepDTO{ID: v.ID, Number: v.Number, ResponseType: string(v.ResponseType), Goal: v.Goal, CounterpartyMessage: v.CounterpartyMessage, MaxPoints: v.MaxPoints, AIInstruction: v.AIInstruction, FallbackMessage: v.FallbackMessage}
}
func optionFromDTO(v adminOptionDTO) domain.ScenarioOption {
	return domain.ScenarioOption{ID: v.ID, Text: v.Text, Explanation: v.Explanation, Points: v.Points, SortOrder: v.SortOrder}
}
func optionToDTO(v domain.ScenarioOption) adminOptionDTO {
	return adminOptionDTO{ID: v.ID, Text: v.Text, Explanation: v.Explanation, Points: v.Points, SortOrder: v.SortOrder}
}

func NewAdmin(s *service.ContentService) *AdminHandler { return &AdminHandler{service: s} }
func (h *AdminHandler) Routes() []router.Route {
	return []router.Route{{Path: "/admin/scenarios", Handler: h.scenarios}, {Path: "/admin/scenarios/", Handler: h.scenario}, {Path: "/admin/steps/", Handler: h.step}}
}
func admin(r *http.Request) bool {
	identity, ok := auth.IdentityFromContext(r.Context())
	return ok && identity.AccessRole == domain.AccessRoleAdmin
}
func (h *AdminHandler) scenarios(w http.ResponseWriter, r *http.Request) {
	if !admin(r) {
		response.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if r.Method == http.MethodGet {
		v, e := h.service.List()
		if e != nil {
			response.Error(w, "could not list scenarios", 500)
			return
		}
		result := make([]adminScenarioDTO, len(v))
		for i, scenario := range v {
			result[i] = scenarioToDTO(scenario)
		}
		response.JSON(w, result)
		return
	}
	if r.Method != http.MethodPost {
		response.Error(w, "method not allowed", 405)
		return
	}
	var v adminScenarioDTO
	if request.DecodeJSON(r, &v) != nil {
		response.Error(w, "invalid JSON", 400)
		return
	}
	created, e := h.service.Create(scenarioFromDTO(v))
	if e != nil {
		response.Error(w, "could not create scenario", 400)
		return
	}
	response.JSONStatus(w, scenarioToDTO(created), 201)
}
func (h *AdminHandler) scenario(w http.ResponseWriter, r *http.Request) {
	if !admin(r) {
		response.Error(w, "forbidden", 403)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/scenarios/")
	parts := strings.Split(path, "/")
	id, e := strconv.Atoi(parts[0])
	if e != nil {
		response.Error(w, "invalid scenario", 400)
		return
	}
	if len(parts) == 1 {
		if r.Method == http.MethodDelete {
			e = h.service.Archive(id)
		} else if r.Method == http.MethodPut {
			var v adminScenarioDTO
			if request.DecodeJSON(r, &v) != nil {
				response.Error(w, "invalid JSON", 400)
				return
			}
			v.ID = id
			e = h.service.Update(scenarioFromDTO(v))
		} else {
			response.Error(w, "method not allowed", 405)
			return
		}
	} else if parts[1] == "publish" && r.Method == http.MethodPost {
		e = h.service.Publish(id)
	} else if parts[1] == "deactivate" && r.Method == http.MethodPost {
		e = h.service.Deactivate(id)
	} else if parts[1] == "restore" && r.Method == http.MethodPost {
		e = h.service.Restore(id)
	} else if parts[1] == "steps" && r.Method == http.MethodPost {
		var v adminStepDTO
		if request.DecodeJSON(r, &v) != nil {
			response.Error(w, "invalid JSON", 400)
			return
		}
		step := stepFromDTO(v)
		step.ScenarioID = id
		created, x := h.service.AddStep(step)
		if x != nil {
			response.Error(w, "invalid content state", 409)
			return
		}
		response.JSONStatus(w, stepToDTO(created), 201)
		return
	} else if len(parts) == 3 && parts[1] == "steps" {
		stepID, parseErr := strconv.Atoi(parts[2])
		if parseErr != nil {
			response.Error(w, "invalid step", http.StatusBadRequest)
			return
		}
		if r.Method == http.MethodDelete {
			e = h.service.DeleteStep(stepID)
		} else if r.Method == http.MethodPut {
			var v adminStepDTO
			if request.DecodeJSON(r, &v) != nil {
				response.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			step := stepFromDTO(v)
			step.ID, step.ScenarioID = stepID, id
			e = h.service.UpdateStep(step)
		} else {
			response.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	} else {
		response.Error(w, "method not allowed", 405)
		return
	}
	if e != nil {
		response.Error(w, "invalid content state", 409)
		return
	}
	w.WriteHeader(204)
}
func (h *AdminHandler) step(w http.ResponseWriter, r *http.Request) {
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
	id, e := strconv.Atoi(parts[0])
	if e != nil {
		response.Error(w, "invalid step", 400)
		return
	}
	if len(parts) == 2 && r.Method == http.MethodPost {
		var v adminOptionDTO
		if request.DecodeJSON(r, &v) != nil {
			response.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		option := optionFromDTO(v)
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
	if r.Method == http.MethodDelete {
		err = h.service.DeleteOption(optionID)
	} else if r.Method == http.MethodPut {
		var v adminOptionDTO
		if request.DecodeJSON(r, &v) != nil {
			response.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		option := optionFromDTO(v)
		option.ID, option.StepID = optionID, id
		err = h.service.UpdateOption(option)
	} else {
		response.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err != nil {
		response.Error(w, "invalid content state", http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
