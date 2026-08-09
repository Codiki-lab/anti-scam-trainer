package http

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"anti-scam-trainer/backend/internal/core/server/request"
	"anti-scam-trainer/backend/internal/core/server/response"
	"anti-scam-trainer/backend/internal/core/server/router"
	"anti-scam-trainer/backend/internal/features/attempts/service"
	auth "anti-scam-trainer/backend/internal/features/auth/service"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type GameHandler struct{ service *service.GameService }

func NewGame(service *service.GameService) *GameHandler { return &GameHandler{service: service} }
func (h *GameHandler) Routes() []router.Route {
	return []router.Route{{Path: "/training/levels", Handler: h.levels}, {Path: "/training/levels/", Handler: h.start}, {Path: "/attempts/", Handler: h.attempt}}
}

func gameIdentity(r *http.Request) (auth.Identity, bool) {
	return auth.IdentityFromContext(r.Context())
}
func gameRole(r *http.Request) (string, bool) {
	role := r.URL.Query().Get("role")
	return role, role == "buyer" || role == "seller"
}

func (h *GameHandler) levels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	identity, ok := gameIdentity(r)
	if !ok {
		response.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	role, ok := gameRole(r)
	if !ok {
		response.Error(w, "role must be buyer or seller", http.StatusBadRequest)
		return
	}
	levels, err := h.service.Levels(identity.UserID, role)
	if err != nil {
		gameError(w, err)
		return
	}
	type dto struct {
		Number     int  `json:"number"`
		Opened     bool `json:"opened"`
		ScenarioID int  `json:"scenario_id"`
	}
	result := make([]dto, len(levels))
	for i, level := range levels {
		result[i] = dto{Number: level.Level.Number, Opened: level.Opened, ScenarioID: level.ScenarioID}
	}
	response.JSON(w, result)
}

func (h *GameHandler) start(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/start") {
		response.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	identity, ok := gameIdentity(r)
	if !ok {
		response.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	role, ok := gameRole(r)
	if !ok {
		response.Error(w, "role must be buyer or seller", http.StatusBadRequest)
		return
	}
	path := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/training/levels/"), "/start")
	level, err := strconv.Atoi(path)
	if err != nil {
		response.Error(w, "invalid level", http.StatusBadRequest)
		return
	}
	state, err := h.service.Start(identity.UserID, level, role)
	if err != nil {
		gameError(w, err)
		return
	}
	response.JSON(w, gameStateDTO(state))
}

func (h *GameHandler) attempt(w http.ResponseWriter, r *http.Request) {
	identity, ok := gameIdentity(r)
	if !ok {
		response.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/v1/attempts/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 {
		response.Error(w, "not found", http.StatusNotFound)
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		response.Error(w, "invalid attempt", http.StatusBadRequest)
		return
	}
	switch {
	case r.Method == http.MethodPost && parts[1] == "answers":
		var input struct {
			OptionID int `json:"option_id"`
		}
		if err := request.DecodeJSON(r, &input); err != nil {
			response.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		state, completed, err := h.service.Submit(identity.UserID, id, input.OptionID)
		if err != nil {
			gameError(w, err)
			return
		}
		if completed != nil {
			breakdown := make([]map[string]interface{}, len(completed.Breakdown))
			for i, answer := range completed.Breakdown {
				breakdown[i] = map[string]interface{}{"step_id": answer.StepID, "option_id": answer.OptionID, "option_text": answer.OptionText, "points": answer.Points, "explanation": answer.Explanation}
			}
			response.JSON(w, map[string]interface{}{"attempt_id": completed.Attempt.ID, "score": completed.Attempt.Score, "stars": completed.Stars, "answers": breakdown})
			return
		}
		response.JSON(w, gameStateDTO(state))
	case r.Method == http.MethodPost && parts[1] == "abandon":
		if err := h.service.Abandon(identity.UserID, id); err != nil {
			gameError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		response.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func gameStateDTO(state service.GameState) map[string]interface{} {
	history := make([]map[string]interface{}, len(state.Answers))
	for i, answer := range state.Answers {
		optionID := 0
		if answer.OptionID != nil {
			optionID = *answer.OptionID
		}
		history[i] = map[string]interface{}{"step_id": answer.StepID, "option_id": optionID}
	}
	messages := make([]map[string]interface{}, len(state.Messages))
	for i, message := range state.Messages {
		messages[i] = map[string]interface{}{"id": message.ID, "author": message.Author, "text": message.Text, "created_at": message.CreatedAt}
	}
	return map[string]interface{}{"attempt_id": state.Attempt.ID, "scenario_id": state.Attempt.ScenarioID, "step": stepDTO(state.Step), "answers": history, "messages": messages}
}
func stepDTO(step domain.ScenarioStep) map[string]interface{} {
	options := make([]map[string]interface{}, len(step.Options))
	for i, option := range step.Options {
		options[i] = map[string]interface{}{"id": option.ID, "text": option.Text}
	}
	return map[string]interface{}{"id": step.ID, "number": step.Number, "goal": step.Goal, "options": options}
}
func gameError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apperrors.ErrForbidden):
		response.Error(w, "level is closed", http.StatusForbidden)
	case errors.Is(err, apperrors.ErrAttemptNotFound), errors.Is(err, apperrors.ErrScenarioNotFound):
		response.Error(w, "not found", http.StatusNotFound)
	case errors.Is(err, apperrors.ErrInvalidAnswer), errors.Is(err, apperrors.ErrInvalidAttemptStatusTransition):
		response.Error(w, "invalid game command", http.StatusConflict)
	default:
		response.Error(w, "could not process game command", http.StatusInternalServerError)
	}
}
