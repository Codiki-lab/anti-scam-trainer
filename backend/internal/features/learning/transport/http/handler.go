package http

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"anti-scam-trainer/backend/internal/core/server/request"
	"anti-scam-trainer/backend/internal/core/server/response"
	"anti-scam-trainer/backend/internal/core/server/router"
	auth "anti-scam-trainer/backend/internal/features/auth/service"
	"anti-scam-trainer/backend/internal/features/learning/service"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct{ service *service.Service }

func New(service *service.Service) *Handler { return &Handler{service: service} }
func (h *Handler) Routes() []router.Route {
	return []router.Route{{Path: "/topics", Handler: h.topics}, {Path: "/topics/", Handler: h.topic}, {Path: "/progress", Handler: h.progress}, {Path: "/achievements", Handler: h.achievements}, {Path: "/dashboard", Handler: h.dashboard}}
}

func identity(r *http.Request) (auth.Identity, bool) { return auth.IdentityFromContext(r.Context()) }
func role(r *http.Request) (domain.UserRole, bool) {
	value := domain.UserRole(r.URL.Query().Get("role"))
	return value, domain.ValidUserRole(value)
}

func (h *Handler) topics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, "method not allowed", 405)
		return
	}
	user, ok := identity(r)
	if !ok {
		response.Error(w, "unauthorized", 401)
		return
	}
	selected, ok := role(r)
	if !ok {
		response.Error(w, "role must be buyer or seller", 400)
		return
	}
	items, err := h.service.Topics(user.UserID, selected)
	if err != nil {
		learningError(w, err)
		return
	}
	response.JSON(w, topicsDTO(items))
}

func (h *Handler) topic(w http.ResponseWriter, r *http.Request) {
	user, ok := identity(r)
	if !ok {
		response.Error(w, "unauthorized", 401)
		return
	}
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/topics/"), "/")
	parts := strings.Split(path, "/")
	topicID, err := strconv.Atoi(parts[0])
	if err != nil || topicID < 1 {
		response.Error(w, "invalid topic id", 400)
		return
	}
	switch {
	case len(parts) == 1 && r.Method == http.MethodGet:
		item, err := h.service.Topic(user.UserID, topicID)
		if err != nil {
			learningError(w, err)
			return
		}
		response.JSON(w, topicDTO(item))
	case len(parts) == 2 && parts[1] == "theory" && r.Method == http.MethodGet:
		item, blocks, err := h.service.Theory(user.UserID, topicID)
		if err != nil {
			learningError(w, err)
			return
		}
		response.JSON(w, map[string]any{"topic": topicDTO(item), "blocks": blocksDTO(blocks)})
	case len(parts) == 3 && parts[1] == "theory" && parts[2] == "read" && r.Method == http.MethodPost:
		streak, newlyRead, err := h.service.MarkTheoryRead(user.UserID, topicID)
		if err != nil {
			learningError(w, err)
			return
		}
		response.JSON(w, map[string]any{"theory_read": true, "newly_read": newlyRead, "streak": streak})
	case len(parts) == 2 && parts[1] == "quiz" && r.Method == http.MethodGet:
		quiz, err := h.service.Quiz(user.UserID, topicID)
		if err != nil {
			learningError(w, err)
			return
		}
		response.JSON(w, quizDTO(quiz))
	case len(parts) == 3 && parts[1] == "quiz" && parts[2] == "attempts" && r.Method == http.MethodPost:
		var input struct {
			Answers []domain.QuizAnswer `json:"answers"`
		}
		if err := request.DecodeStrictJSON(r, &input); err != nil {
			response.Error(w, "invalid JSON", 400)
			return
		}
		result, err := h.service.SubmitQuiz(user.UserID, topicID, input.Answers)
		if err != nil {
			learningError(w, err)
			return
		}
		response.JSON(w, map[string]any{"score": result.Score, "passed": result.Passed, "best_score": result.BestScore, "newly_passed": result.NewlyPassed, "streak": result.Streak})
	default:
		response.Error(w, "method not allowed", 405)
	}
}

func (h *Handler) progress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, "method not allowed", 405)
		return
	}
	user, ok := identity(r)
	if !ok {
		response.Error(w, "unauthorized", 401)
		return
	}
	selected, ok := role(r)
	if !ok {
		response.Error(w, "role must be buyer or seller", 400)
		return
	}
	items, recent, average, err := h.service.Progress(user.UserID, selected)
	if err != nil {
		learningError(w, err)
		return
	}
	completed := 0
	completedLevels := 0
	stars := 0
	for _, t := range items {
		if t.Completed {
			completed++
		}
		for _, l := range t.Levels {
			stars += l.Stars
			if l.Stars > 0 {
				completedLevels++
			}
		}
	}
	response.JSON(w, map[string]any{"role": selected, "summary": map[string]any{"completed_topics": completed, "total_topics": len(items), "completed_levels": completedLevels, "total_levels": len(items) * 4, "stars": stars, "average_score": average}, "topics": topicsDTO(items), "recent_attempts": recent})
}
func (h *Handler) achievements(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, "method not allowed", 405)
		return
	}
	user, ok := identity(r)
	if !ok {
		response.Error(w, "unauthorized", 401)
		return
	}
	items, err := h.service.Achievements(user.UserID)
	if err != nil {
		learningError(w, err)
		return
	}
	earned := []map[string]any{}
	available := []map[string]any{}
	for _, item := range items {
		dto := achievementDTO(item)
		if item.Earned {
			earned = append(earned, dto)
		} else {
			available = append(available, dto)
		}
	}
	response.JSON(w, map[string]any{"earned": earned, "available": available})
}
func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, "method not allowed", 405)
		return
	}
	identity, ok := identity(r)
	if !ok {
		response.Error(w, "unauthorized", 401)
		return
	}
	selected, ok := role(r)
	if !ok {
		response.Error(w, "role must be buyer or seller", 400)
		return
	}
	user, topics, achievements, action, dailyTask, err := h.service.Dashboard(identity.UserID, selected)
	if err != nil {
		learningError(w, err)
		return
	}
	preview := []map[string]any{}
	for i, item := range achievements {
		if i == 3 {
			break
		}
		preview = append(preview, achievementDTO(item))
	}
	response.JSON(w, map[string]any{"profile": map[string]any{"id": user.ID, "username": user.Username, "training_role": user.TrainingRole}, "streak": user.Streak, "topics": topicsDTO(topics), "achievements": preview, "continue_action": continueActionDTOFrom(action), "daily_task": dailyTaskDTOFrom(dailyTask)})
}

type continueActionDTO struct {
	Type      string `json:"type"`
	TopicID   int    `json:"topic_id,omitempty"`
	Level     int    `json:"level,omitempty"`
	AttemptID int    `json:"attempt_id,omitempty"`
}
type dailyTaskDTO struct {
	Date        string            `json:"date"`
	Role        domain.UserRole   `json:"role"`
	Completed   bool              `json:"completed"`
	CompletedAt *time.Time        `json:"completed_at"`
	Action      continueActionDTO `json:"action"`
}

func continueActionDTOFrom(action *domain.ContinueAction) *continueActionDTO {
	if action == nil {
		return nil
	}
	return &continueActionDTO{Type: action.Type, TopicID: action.TopicID, Level: action.Level, AttemptID: action.AttemptID}
}
func dailyTaskDTOFrom(task *domain.DailyTask) *dailyTaskDTO {
	if task == nil {
		return nil
	}
	action := continueActionDTOFrom(&task.Action)
	return &dailyTaskDTO{Date: task.Date, Role: task.Role, Completed: task.Completed, CompletedAt: task.CompletedAt, Action: *action}
}

func topicDTO(t domain.Topic) map[string]any {
	levels := make([]map[string]any, len(t.Levels))
	for i, l := range t.Levels {
		levels[i] = map[string]any{"number": l.Number, "opened": l.Opened, "best_score": l.BestScore, "stars": l.Stars, "attempts": l.Attempts, "last_attempt_id": l.LastAttemptID}
	}
	return map[string]any{"id": t.ID, "slug": t.Slug, "role": t.UserRole, "title": t.Title, "description": t.Description, "sort_order": t.SortOrder, "theory_read": t.TheoryRead, "quiz_passed": t.QuizPassed, "quiz_best_score": t.QuizScore, "completed": t.Completed, "levels": levels}
}
func topicsDTO(items []domain.Topic) []map[string]any {
	result := make([]map[string]any, len(items))
	for i, item := range items {
		result[i] = topicDTO(item)
	}
	return result
}
func blocksDTO(items []domain.TheoryBlock) []map[string]any {
	result := make([]map[string]any, len(items))
	for i, x := range items {
		result[i] = map[string]any{"id": x.ID, "sort_order": x.SortOrder, "kind": x.Kind, "title": x.Title, "body": x.Body}
	}
	return result
}
func quizDTO(items []domain.QuizQuestion) map[string]any {
	questions := make([]map[string]any, len(items))
	for i, q := range items {
		options := make([]map[string]any, len(q.Options))
		for j, o := range q.Options {
			options[j] = map[string]any{"id": o.ID, "text": o.Text}
		}
		questions[i] = map[string]any{"id": q.ID, "sort_order": q.SortOrder, "text": q.Text, "options": options}
	}
	return map[string]any{"questions": questions, "pass_threshold": 80}
}
func achievementDTO(x domain.Achievement) map[string]any {
	result := map[string]any{"code": x.Code, "title": x.Title, "description": x.Description, "icon": x.Icon, "earned": x.Earned, "progress": map[string]int{"current": x.Current, "target": x.Target}}
	if x.Earned {
		result["earned_at"] = x.EarnedAt
	}
	return result
}
func learningError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidQuiz):
		response.Error(w, "invalid quiz submission", 400)
	case errors.Is(err, apperrors.ErrForbidden):
		response.ErrorCode(w, "CONTENT_UNAVAILABLE", "content is not available", 403, nil)
	case errors.Is(err, service.ErrTopicNotFound):
		response.Error(w, "topic not found", 404)
	case errors.Is(err, service.ErrDailyTaskUnavailable):
		response.ErrorCode(w, "CONTENT_UNAVAILABLE", "no valid daily task is available", http.StatusConflict, nil)
	default:
		response.Error(w, "could not process learning request", 500)
	}
}
