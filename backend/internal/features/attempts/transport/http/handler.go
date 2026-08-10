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
	"time"
)

type Handler struct{ service *service.GameService }

func New(service *service.GameService) *Handler { return &Handler{service: service} }
func (h *Handler) Routes() []router.Route {
	return []router.Route{{Path: "/training/levels", Handler: h.levels}, {Path: "/training/levels/", Handler: h.start}, {Path: "/training/free-play/start", Handler: h.startFreePlay}, {Path: "/attempts/", Handler: h.attempt}}
}

func gameIdentity(r *http.Request) (auth.Identity, bool) {
	return auth.IdentityFromContext(r.Context())
}
func gameRole(r *http.Request) (string, bool) {
	role := r.URL.Query().Get("role")
	return role, role == "buyer" || role == "seller"
}
func gameTopic(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.URL.Query().Get("topic_id"))
	return id, err == nil && id > 0
}

func (h *Handler) levels(w http.ResponseWriter, r *http.Request) {
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
	topicID, ok := gameTopic(r)
	if !ok {
		response.Error(w, "topic_id is required", http.StatusBadRequest)
		return
	}
	levels, err := h.service.Levels(identity.UserID, role, topicID)
	if err != nil {
		gameError(w, err)
		return
	}
	type dto struct {
		Number              int                 `json:"number"`
		Opened              bool                `json:"opened"`
		ScenarioID          int                 `json:"scenario_id"`
		ScenarioTitle       string              `json:"scenario_title"`
		ScenarioDescription string              `json:"scenario_description"`
		ResponseType        domain.ResponseType `json:"response_type"`
		InProgressAttemptID int                 `json:"in_progress_attempt_id,omitempty"`
	}
	result := make([]dto, len(levels))
	for i, level := range levels {
		result[i] = dto{Number: level.Level.Number, Opened: level.Opened, ScenarioID: level.ScenarioID, ScenarioTitle: level.ScenarioTitle, ScenarioDescription: level.ScenarioDescription, ResponseType: level.ResponseType, InProgressAttemptID: level.InProgressAttemptID}
	}
	response.JSON(w, result)
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
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
	topicID, ok := gameTopic(r)
	if !ok {
		response.Error(w, "topic_id is required", http.StatusBadRequest)
		return
	}
	state, err := h.service.Start(identity.UserID, level, role, topicID)
	if err != nil {
		gameError(w, err)
		return
	}
	response.JSON(w, gameStateDTO(state))
}

func (h *Handler) startFreePlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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
	state, err := h.service.StartFreePlay(r.Context(), identity.UserID, role)
	if err != nil {
		gameError(w, err)
		return
	}
	response.JSON(w, gameStateDTO(state))
}

func (h *Handler) attempt(w http.ResponseWriter, r *http.Request) {
	identity, ok := gameIdentity(r)
	if !ok {
		response.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/v1/attempts/")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) < 1 || len(parts) > 2 {
		response.Error(w, "not found", http.StatusNotFound)
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		response.Error(w, "invalid attempt", http.StatusBadRequest)
		return
	}
	switch {
	case r.Method == http.MethodGet && len(parts) == 1:
		state, err := h.service.GetState(identity.UserID, id)
		if err != nil {
			gameError(w, err)
			return
		}
		response.JSON(w, gameStateDTO(state))
	case r.Method == http.MethodPost && len(parts) == 2 && parts[1] == "answers":
		var input struct {
			StepID   *int    `json:"step_id"`
			OptionID *int    `json:"option_id"`
			FreeText *string `json:"free_text"`
			Finish   bool    `json:"finish"`
		}
		if err := request.DecodeStrictJSON(r, &input); err != nil {
			response.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if input.StepID == nil {
			response.Error(w, "step_id is required", http.StatusBadRequest)
			return
		}
		state, completed, err := h.service.SubmitAnswer(r.Context(), identity.UserID, id, service.AnswerCommand{StepID: input.StepID, OptionID: input.OptionID, FreeText: input.FreeText, Finish: input.Finish})
		if err != nil {
			gameError(w, err)
			return
		}
		if completed != nil {
			breakdown := make([]map[string]interface{}, len(completed.Breakdown))
			for i, answer := range completed.Breakdown {
				breakdown[i] = map[string]interface{}{"step_id": answer.StepID, "step_number": answer.StepNumber, "answer_type": answer.AnswerType, "option_id": answer.OptionID, "option_text": answer.OptionText, "free_text": answer.FreeText, "points": answer.Points, "assessment": answer.Assessment, "explanation": answer.Explanation, "safe_action": answer.SafeAction, "risk_signals": answer.RiskSignals}
			}
			result := map[string]interface{}{"attempt_id": completed.Attempt.ID, "score": completed.Attempt.Score, "stars": completed.Stars, "answers": breakdown}
			if completed.Attempt.Mode == domain.AttemptModeFreePlay && completed.Attempt.IsScam != nil {
				result["is_scam"] = *completed.Attempt.IsScam
			}
			for key, value := range resultDTO(completed.Result) {
				result[key] = value
			}
			response.JSON(w, result)
			return
		}
		response.JSON(w, gameStateDTO(state))
	case r.Method == http.MethodPost && len(parts) == 2 && parts[1] == "abandon":
		if err := h.service.Abandon(identity.UserID, id); err != nil {
			gameError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodGet && len(parts) == 2 && parts[1] == "result":
		result, err := h.service.Result(identity.UserID, id)
		if err != nil {
			gameError(w, err)
			return
		}
		response.JSON(w, resultDTO(result))
	default:
		response.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func gameStateDTO(state service.GameState) map[string]interface{} {
	history := make([]map[string]interface{}, len(state.Answers))
	for i, answer := range state.Answers {
		stepID := answer.StepID
		if state.Attempt.Mode == domain.AttemptModeFreePlay {
			stepID = answer.TurnNumber
		}
		history[i] = map[string]interface{}{"step_id": stepID, "points": answer.AwardedPoints}
		if answer.OptionID != nil {
			history[i]["answer_type"] = "option"
			history[i]["option_id"] = *answer.OptionID
			history[i]["option_text"] = answer.OptionText
		} else {
			history[i]["answer_type"] = "free_text"
			history[i]["free_text"] = answer.FreeText
		}
	}
	messages := make([]map[string]interface{}, len(state.Messages))
	for i, message := range state.Messages {
		messages[i] = map[string]interface{}{"role": message.Role, "text": message.Text}
	}
	scenarioTitle := state.Scenario.Title
	scenarioDescription := state.Scenario.Description
	level, _ := strconv.Atoi(state.Scenario.Level)
	topicTitle := state.Scenario.TopicTitle
	if state.Attempt.Mode == domain.AttemptModeFreePlay {
		scenarioTitle, scenarioDescription, level, topicTitle = "Свободная игра", "Непредсказуемая тренировка безопасной сделки", 0, "Все изученные Темы"
	}
	counterpartyRole := "seller"
	if state.Attempt.UserRole == "seller" {
		counterpartyRole = "buyer"
	}
	return map[string]interface{}{"attempt_id": state.Attempt.ID, "status": state.Attempt.Status, "scenario_id": state.Attempt.ScenarioID, "scenario_title": scenarioTitle, "scenario_description": scenarioDescription, "topic_id": state.Scenario.TopicID, "topic_title": topicTitle, "level": level, "user_role": state.Attempt.UserRole, "counterparty_role": counterpartyRole, "product_context": state.Scenario.ProductContext, "mode": state.Step.ResponseType, "step_progress": map[string]int{"current": state.Step.Number, "answered": len(state.Answers), "total": state.TotalSteps}, "step": stepDTO(state.Step), "answers": history, "messages": messages, "can_finish_early": state.CanFinishEarly}
}
func stepDTO(step domain.ScenarioStep) map[string]interface{} {
	options := make([]map[string]interface{}, len(step.Options))
	for i, option := range step.Options {
		options[i] = map[string]interface{}{"id": option.ID, "text": option.Text}
	}
	return map[string]interface{}{"id": step.ID, "number": step.Number, "counterparty_message": step.CounterpartyMessage, "options": options}
}
func resultDTO(result domain.AttemptResult) map[string]interface{} {
	achievements := make([]map[string]interface{}, len(result.NewAchievements))
	for i, achievement := range result.NewAchievements {
		achievements[i] = map[string]interface{}{"code": achievement.Code, "title": achievement.Title, "description": achievement.Description, "icon": achievement.Icon, "earned": achievement.Earned, "earned_at": achievement.EarnedAt, "progress": map[string]int{"current": achievement.Current, "target": achievement.Target}}
	}
	dto := map[string]interface{}{"attempt_id": result.AttemptID, "score": result.Score, "stars": result.Stars, "decision_review": result.DecisionReview, "risk_signals": result.RiskSignals, "safe_actions": result.SafeActions, "level_progress": result.LevelProgress, "topic_id": result.TopicID, "topic_completed": result.TopicCompleted, "next_action": result.NextAction, "new_achievements": achievements, "streak": result.Streak}
	if result.IsScam != nil {
		dto["is_scam"] = *result.IsScam
	}
	return dto
}
func gameError(w http.ResponseWriter, err error) {
	var limited *service.RateLimitError
	switch {
	case errors.As(err, &limited):
		seconds := int64((limited.RetryAfter + time.Second - 1) / time.Second)
		if seconds < 1 {
			seconds = 1
		}
		w.Header().Set("Retry-After", strconv.FormatInt(seconds, 10))
		response.ErrorCode(w, "RATE_LIMITED", "too many AI requests", http.StatusTooManyRequests, map[string]any{"retry_after_seconds": seconds})
	case errors.Is(err, apperrors.ErrForbidden):
		response.ErrorCode(w, "CONTENT_UNAVAILABLE", "level is closed", http.StatusForbidden, nil)
	case errors.Is(err, apperrors.ErrAttemptNotFound), errors.Is(err, apperrors.ErrScenarioNotFound):
		response.Error(w, "not found", http.StatusNotFound)
	case errors.Is(err, apperrors.ErrInvalidAnswer), errors.Is(err, apperrors.ErrInvalidAttemptStatusTransition):
		response.Error(w, "invalid game command", http.StatusConflict)
	case errors.Is(err, apperrors.ErrStaleStep):
		response.ErrorCode(w, "STALE_STEP", "stale step", http.StatusConflict, nil)
	case errors.Is(err, service.ErrAIUnavailable):
		response.Error(w, "AI is temporarily unavailable; retry the answer", http.StatusServiceUnavailable)
	case errors.Is(err, service.ErrAIInvalidResponse), errors.Is(err, service.ErrAIContextExhausted):
		response.Error(w, "AI could not safely process the answer; retry later", http.StatusBadGateway)
	default:
		response.Error(w, "could not process game command", http.StatusInternalServerError)
	}
}
