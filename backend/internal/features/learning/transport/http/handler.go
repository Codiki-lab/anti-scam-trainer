package http

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"anti-scam-trainer/backend/internal/core/ratelimit"
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

type Handler struct {
	service                     *service.Service
	chatRecommendationRateLimit *ratelimit.Limiter
}

func New(service *service.Service) *Handler { return &Handler{service: service} }
func NewWithChatRecommendation(service *service.Service, limiter *ratelimit.Limiter) *Handler {
	return &Handler{service: service, chatRecommendationRateLimit: limiter}
}
func (h *Handler) Routes() []router.Route {
	return []router.Route{{Path: "/topics", Handler: h.topics}, {Path: "/topics/", Handler: h.topic}, {Path: "/skill-checks/", Handler: h.skillCheck}, {Path: "/recommendations/next", Handler: h.personalRecommendation}, {Path: "/progress", Handler: h.progress}, {Path: "/achievements", Handler: h.achievements}, {Path: "/dashboard", Handler: h.dashboard}, {Path: "/daily-tasks/answer", Handler: h.answerDailyTask}, {Path: "/integrations/avito-chat/recommendations", Handler: h.chatRecommendation}}
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

func (h *Handler) personalRecommendation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, ok := identity(r)
	if !ok {
		response.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	selected, ok := role(r)
	if !ok {
		response.Error(w, "role must be buyer or seller", http.StatusBadRequest)
		return
	}
	recommendation, err := h.service.PersonalRecommendation(user.UserID, selected)
	if err != nil {
		learningError(w, err)
		return
	}
	response.JSON(w, map[string]any{"topic": topicDTO(recommendation.Topic), "explanation": recommendation.Explanation, "next_action": recommendation.NextAction, "fallback": recommendation.IsFallback})
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
	case len(parts) == 3 && parts[1] == "skill-check" && parts[2] == "start" && r.Method == http.MethodPost:
		check, err := h.service.StartSkillCheck(user.UserID, topicID)
		if err != nil {
			learningError(w, err)
			return
		}
		response.JSON(w, skillCheckDTO(check))
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

func (h *Handler) skillCheck(w http.ResponseWriter, r *http.Request) {
	user, ok := identity(r)
	if !ok {
		response.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/skill-checks/"), "/"), "/")
	checkID, err := strconv.Atoi(parts[0])
	if err != nil || checkID < 1 {
		response.Error(w, "invalid skill check", http.StatusBadRequest)
		return
	}
	switch {
	case len(parts) == 1 && r.Method == http.MethodGet:
		check, err := h.service.SkillCheck(user.UserID, checkID)
		if err != nil {
			learningError(w, err)
			return
		}
		response.JSON(w, skillCheckDTO(check))
	case len(parts) == 2 && parts[1] == "answers" && r.Method == http.MethodPost:
		var input struct {
			Answer *bool `json:"answer"`
		}
		if err := request.DecodeStrictJSON(r, &input); err != nil || input.Answer == nil {
			response.Error(w, "answer is required", http.StatusBadRequest)
			return
		}
		check, err := h.service.AnswerSkillCheck(user.UserID, checkID, *input.Answer)
		if err != nil {
			learningError(w, err)
			return
		}
		response.JSON(w, skillCheckDTO(check))
	default:
		response.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func skillCheckDTO(check domain.SkillCheck) map[string]any {
	phase := check.Phase()
	dto := map[string]any{"id": check.ID, "topic_id": check.TopicID, "phase": phase}
	if phase == "before" {
		dto["snapshot"] = check.Before.Messages
	}
	if phase == "after" {
		dto["snapshot"] = check.After.Messages
	}
	if phase == "completed" {
		outcome, _ := check.Outcome()
		dto["before_correct"] = outcome.BeforeCorrect
		dto["after_correct"] = outcome.AfterCorrect
		dto["verdict_improved"] = outcome.VerdictImproved
		dto["before_pattern"] = outcome.BeforePattern
		dto["after_pattern"] = outcome.AfterPattern
		dto["pattern_improved"] = outcome.PatternImproved
		dto["improved"] = outcome.Improved
	}
	return dto
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
	Messages    []dailyMessageDTO `json:"messages"`
	Completed   bool              `json:"completed"`
	CompletedAt *time.Time        `json:"completed_at"`
	Answer      *bool             `json:"answer,omitempty"`
	Correct     *bool             `json:"correct,omitempty"`
	Verdict     *bool             `json:"verdict,omitempty"`
	Signals     []string          `json:"signals,omitempty"`
	SafeAction  string            `json:"safe_action,omitempty"`
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
	messages := make([]dailyMessageDTO, len(task.Messages))
	for i, message := range task.Messages {
		messages[i] = dailyMessageDTO{Role: string(message.Role), Text: message.Text}
	}
	dto := &dailyTaskDTO{Date: task.Date, Role: task.Role, Messages: messages, Completed: task.Completed, CompletedAt: task.CompletedAt}
	if task.Completed {
		dto.Answer, dto.Correct, dto.Verdict, dto.Signals, dto.SafeAction = task.Answer, task.Correct, &task.Verdict, task.Signals, task.SafeAction
	}
	return dto
}

type dailyMessageDTO struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

func (h *Handler) answerDailyTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, ok := identity(r)
	if !ok {
		response.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var input struct {
		Answer *bool `json:"answer"`
	}
	if err := request.DecodeStrictJSON(r, &input); err != nil {
		response.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	task, streak, err := h.service.AnswerDailyTask(user.UserID, input.Answer)
	if err != nil {
		learningError(w, err)
		return
	}
	response.JSON(w, map[string]any{"daily_task": dailyTaskDTOFrom(&task), "streak": streak})
}

func (h *Handler) chatRecommendation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, ok := identity(r)
	if !ok {
		response.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var input struct {
		Source      string            `json:"source"`
		Role        domain.UserRole   `json:"role"`
		Messages    []dailyMessageDTO `json:"messages"`
		RiskType    string            `json:"risk_type"`
		RiskSignals []string          `json:"risk_signals"`
	}
	if err := request.DecodeStrictJSON(r, &input); err != nil {
		response.ErrorCode(w, "VALIDATION_ERROR", "invalid chat recommendation request", http.StatusBadRequest, nil)
		return
	}
	messages := make([]domain.DialogueMessage, len(input.Messages))
	for i, message := range input.Messages {
		messages[i] = domain.DialogueMessage{Role: domain.MessageRole(message.Role), Text: message.Text}
	}
	command := service.ChatRecommendationCommand{Source: input.Source, Role: input.Role, Messages: messages, RiskType: input.RiskType, RiskSignals: input.RiskSignals}
	if err := service.ValidateChatRecommendation(command); err != nil {
		response.ErrorCode(w, "VALIDATION_ERROR", "chat snapshot must be anonymized", http.StatusBadRequest, nil)
		return
	}
	if h.chatRecommendationRateLimit != nil {
		if allowed, retry := h.chatRecommendationRateLimit.Allow(strconv.Itoa(user.UserID)); !allowed {
			seconds := int64((retry + time.Second - 1) / time.Second)
			if seconds < 1 {
				seconds = 1
			}
			w.Header().Set("Retry-After", strconv.FormatInt(seconds, 10))
			response.ErrorCode(w, "RATE_LIMITED", "too many chat recommendation requests", http.StatusTooManyRequests, map[string]any{"retry_after_seconds": seconds})
			return
		}
	}
	recommendation, err := h.service.RecommendFromChat(user.UserID, command)
	if err != nil {
		learningError(w, err)
		return
	}
	response.JSON(w, map[string]any{"topic": topicDTO(recommendation.Topic), "explanation": recommendation.Explanation, "next_action": continueActionDTOFrom(&recommendation.NextAction), "fallback": recommendation.IsFallback})
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
	case errors.Is(err, service.ErrDailyTaskAnswered):
		response.ErrorCode(w, "STATE_CONFLICT", "daily task is already answered", http.StatusConflict, nil)
	case errors.Is(err, service.ErrInvalidDailyAnswer):
		response.ErrorCode(w, "VALIDATION_ERROR", "answer must be a boolean", http.StatusBadRequest, nil)
	case errors.Is(err, service.ErrInvalidChatReferral):
		response.ErrorCode(w, "VALIDATION_ERROR", "chat snapshot must be anonymized", http.StatusBadRequest, nil)
	default:
		response.Error(w, "could not process learning request", 500)
	}
}

type AdminHandler struct{ service *service.ContentService }

func NewAdmin(service *service.ContentService) *AdminHandler { return &AdminHandler{service: service} }
func (h *AdminHandler) Routes() []router.Route {
	return []router.Route{{Path: "/admin/topics", Handler: h.topics}, {Path: "/admin/topics/", Handler: h.topic}}
}

type adminTopicDTO struct {
	ID          int                `json:"id"`
	Slug        string             `json:"slug"`
	Role        domain.UserRole    `json:"role"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	SortOrder   int                `json:"sort_order"`
	Status      string             `json:"status"`
	Theory      []adminTheoryDTO   `json:"theory,omitempty"`
	Quiz        []adminQuestionDTO `json:"quiz,omitempty"`
}
type adminTheoryDTO struct {
	ID        int    `json:"id"`
	SortOrder int    `json:"sort_order"`
	Kind      string `json:"kind"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}
type adminQuestionDTO struct {
	ID          int                  `json:"id"`
	SortOrder   int                  `json:"sort_order"`
	Text        string               `json:"text"`
	Explanation string               `json:"explanation"`
	Options     []adminQuizOptionDTO `json:"options,omitempty"`
}
type adminQuizOptionDTO struct {
	ID        int    `json:"id"`
	SortOrder int    `json:"sort_order"`
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
}

func requireAdmin(r *http.Request) bool {
	identity, ok := auth.IdentityFromContext(r.Context())
	return ok && identity.AccessRole == domain.AccessRoleAdmin
}
func topicFromAdmin(v adminTopicDTO) domain.Topic {
	return domain.Topic{ID: v.ID, Slug: v.Slug, UserRole: v.Role, Title: v.Title, Description: v.Description, SortOrder: v.SortOrder, Status: v.Status}
}
func topicToAdmin(v domain.Topic) adminTopicDTO {
	return adminTopicDTO{ID: v.ID, Slug: v.Slug, Role: v.UserRole, Title: v.Title, Description: v.Description, SortOrder: v.SortOrder, Status: v.Status}
}
func contentToAdmin(v domain.TopicContent) adminTopicDTO {
	dto := topicToAdmin(v.Topic)
	dto.Theory = make([]adminTheoryDTO, len(v.Theory))
	for i, x := range v.Theory {
		dto.Theory[i] = adminTheoryDTO{ID: x.ID, SortOrder: x.SortOrder, Kind: x.Kind, Title: x.Title, Body: x.Body}
	}
	dto.Quiz = make([]adminQuestionDTO, len(v.Quiz))
	for i, q := range v.Quiz {
		opts := make([]adminQuizOptionDTO, len(q.Options))
		for j, o := range q.Options {
			opts[j] = adminQuizOptionDTO{ID: o.ID, SortOrder: o.SortOrder, Text: o.Text, IsCorrect: o.Correct}
		}
		dto.Quiz[i] = adminQuestionDTO{ID: q.ID, SortOrder: q.SortOrder, Text: q.Text, Explanation: q.Explanation, Options: opts}
	}
	return dto
}

func (h *AdminHandler) topics(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(r) {
		response.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := h.service.List()
		if err != nil {
			adminContentError(w, err)
			return
		}
		result := make([]adminTopicDTO, len(items))
		for i, item := range items {
			result[i] = topicToAdmin(item)
		}
		response.JSON(w, result)
	case http.MethodPost:
		var input adminTopicDTO
		if request.DecodeStrictJSON(r, &input) != nil {
			response.Error(w, "invalid JSON", 400)
			return
		}
		created, err := h.service.Create(topicFromAdmin(input))
		if err != nil {
			adminContentError(w, err)
			return
		}
		response.JSONStatus(w, topicToAdmin(created), http.StatusCreated)
	default:
		response.Error(w, "method not allowed", 405)
	}
}
func (h *AdminHandler) topic(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(r) {
		response.Error(w, "forbidden", 403)
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/admin/topics/"), "/"), "/")
	if len(parts) == 0 {
		return
	}
	topicID, err := strconv.Atoi(parts[0])
	if err != nil || topicID < 1 {
		response.Error(w, "invalid topic", 400)
		return
	}
	if len(parts) == 1 {
		h.topicRoot(w, r, topicID)
		return
	}
	switch parts[1] {
	case "publish", "deactivate", "restore":
		if len(parts) != 2 || r.Method != http.MethodPost {
			response.Error(w, "method not allowed", 405)
			return
		}
		switch parts[1] {
		case "publish":
			err = h.service.Publish(topicID)
		case "deactivate":
			err = h.service.Deactivate(topicID)
		default:
			err = h.service.Restore(topicID)
		}
		if err != nil {
			adminContentError(w, err)
			return
		}
		w.WriteHeader(204)
	case "theory-blocks":
		h.theory(w, r, topicID, parts[2:])
	case "quiz-questions":
		h.questions(w, r, topicID, parts[2:])
	default:
		response.Error(w, "not found", 404)
	}
}
func (h *AdminHandler) topicRoot(w http.ResponseWriter, r *http.Request, id int) {
	switch r.Method {
	case http.MethodGet:
		v, err := h.service.Get(id)
		if err != nil {
			adminContentError(w, err)
			return
		}
		response.JSON(w, contentToAdmin(v))
	case http.MethodPut:
		var input adminTopicDTO
		if request.DecodeStrictJSON(r, &input) != nil {
			response.Error(w, "invalid JSON", 400)
			return
		}
		input.ID = id
		if err := h.service.Update(topicFromAdmin(input)); err != nil {
			adminContentError(w, err)
			return
		}
		w.WriteHeader(204)
	case http.MethodDelete:
		if err := h.service.Archive(id); err != nil {
			adminContentError(w, err)
			return
		}
		w.WriteHeader(204)
	default:
		response.Error(w, "method not allowed", 405)
	}
}
func (h *AdminHandler) theory(w http.ResponseWriter, r *http.Request, topicID int, rest []string) {
	var input adminTheoryDTO
	if r.Method == http.MethodPost && len(rest) == 0 {
		if request.DecodeStrictJSON(r, &input) != nil {
			response.Error(w, "invalid JSON", 400)
			return
		}
		created, err := h.service.AddTheory(domain.TheoryBlock{TopicID: topicID, SortOrder: input.SortOrder, Kind: input.Kind, Title: input.Title, Body: input.Body})
		if err != nil {
			adminContentError(w, err)
			return
		}
		input.ID = created.ID
		response.JSONStatus(w, input, 201)
		return
	}
	if len(rest) != 1 {
		response.Error(w, "method not allowed", 405)
		return
	}
	id, err := strconv.Atoi(rest[0])
	if err != nil {
		response.Error(w, "invalid theory block", 400)
		return
	}
	switch r.Method {
	case http.MethodDelete:
		err = h.service.DeleteTheory(topicID, id)
	case http.MethodPut:
		if request.DecodeStrictJSON(r, &input) != nil {
			response.Error(w, "invalid JSON", 400)
			return
		}
		err = h.service.UpdateTheory(domain.TheoryBlock{ID: id, TopicID: topicID, SortOrder: input.SortOrder, Kind: input.Kind, Title: input.Title, Body: input.Body})
	default:
		response.Error(w, "method not allowed", 405)
		return
	}
	if err != nil {
		adminContentError(w, err)
		return
	}
	w.WriteHeader(204)
}
func (h *AdminHandler) questions(w http.ResponseWriter, r *http.Request, topicID int, rest []string) {
	var input adminQuestionDTO
	if r.Method == http.MethodPost && len(rest) == 0 {
		if request.DecodeStrictJSON(r, &input) != nil {
			response.Error(w, "invalid JSON", 400)
			return
		}
		created, err := h.service.AddQuestion(domain.QuizQuestion{TopicID: topicID, SortOrder: input.SortOrder, Text: input.Text, Explanation: input.Explanation})
		if err != nil {
			adminContentError(w, err)
			return
		}
		input.ID = created.ID
		response.JSONStatus(w, input, 201)
		return
	}
	if len(rest) < 1 {
		response.Error(w, "method not allowed", 405)
		return
	}
	questionID, err := strconv.Atoi(rest[0])
	if err != nil {
		response.Error(w, "invalid quiz question", 400)
		return
	}
	if len(rest) == 1 {
		switch r.Method {
		case http.MethodDelete:
			err = h.service.DeleteQuestion(topicID, questionID)
		case http.MethodPut:
			if request.DecodeStrictJSON(r, &input) != nil {
				response.Error(w, "invalid JSON", 400)
				return
			}
			err = h.service.UpdateQuestion(domain.QuizQuestion{ID: questionID, TopicID: topicID, SortOrder: input.SortOrder, Text: input.Text, Explanation: input.Explanation})
		default:
			response.Error(w, "method not allowed", 405)
			return
		}
		if err != nil {
			adminContentError(w, err)
			return
		}
		w.WriteHeader(204)
		return
	}
	if rest[1] != "options" {
		response.Error(w, "not found", 404)
		return
	}
	h.options(w, r, topicID, questionID, rest[2:])
}
func (h *AdminHandler) options(w http.ResponseWriter, r *http.Request, topicID, questionID int, rest []string) {
	var input adminQuizOptionDTO
	if r.Method == http.MethodPost && len(rest) == 0 {
		if request.DecodeStrictJSON(r, &input) != nil {
			response.Error(w, "invalid JSON", 400)
			return
		}
		created, err := h.service.AddOption(topicID, domain.QuizOption{QuestionID: questionID, SortOrder: input.SortOrder, Text: input.Text, Correct: input.IsCorrect})
		if err != nil {
			adminContentError(w, err)
			return
		}
		input.ID = created.ID
		response.JSONStatus(w, input, 201)
		return
	}
	if len(rest) != 1 {
		response.Error(w, "method not allowed", 405)
		return
	}
	id, err := strconv.Atoi(rest[0])
	if err != nil {
		response.Error(w, "invalid quiz option", 400)
		return
	}
	switch r.Method {
	case http.MethodDelete:
		err = h.service.DeleteOption(topicID, questionID, id)
	case http.MethodPut:
		if request.DecodeStrictJSON(r, &input) != nil {
			response.Error(w, "invalid JSON", 400)
			return
		}
		err = h.service.UpdateOption(topicID, domain.QuizOption{ID: id, QuestionID: questionID, SortOrder: input.SortOrder, Text: input.Text, Correct: input.IsCorrect})
	default:
		response.Error(w, "method not allowed", 405)
		return
	}
	if err != nil {
		adminContentError(w, err)
		return
	}
	w.WriteHeader(204)
}
func adminContentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidContent):
		response.Error(w, "invalid content", 400)
	case errors.Is(err, service.ErrTopicNotFound):
		response.Error(w, "topic not found", 404)
	case errors.Is(err, service.ErrContentConflict):
		response.ErrorCode(w, "CONTENT_CONFLICT", "content lifecycle or invariant conflict", 409, map[string]any{"reason": "content_invariant"})
	default:
		response.Error(w, "could not process content", 500)
	}
}
