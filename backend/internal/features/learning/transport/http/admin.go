package http

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/server/request"
	"anti-scam-trainer/backend/internal/core/server/response"
	"anti-scam-trainer/backend/internal/core/server/router"
	auth "anti-scam-trainer/backend/internal/features/auth/service"
	"anti-scam-trainer/backend/internal/features/learning/service"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

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
		if parts[1] == "publish" {
			err = h.service.Publish(topicID)
		} else if parts[1] == "deactivate" {
			err = h.service.Deactivate(topicID)
		} else {
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
	if r.Method == http.MethodDelete {
		err = h.service.DeleteTheory(topicID, id)
	} else if r.Method == http.MethodPut {
		if request.DecodeStrictJSON(r, &input) != nil {
			response.Error(w, "invalid JSON", 400)
			return
		}
		err = h.service.UpdateTheory(domain.TheoryBlock{ID: id, TopicID: topicID, SortOrder: input.SortOrder, Kind: input.Kind, Title: input.Title, Body: input.Body})
	} else {
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
		if r.Method == http.MethodDelete {
			err = h.service.DeleteQuestion(topicID, questionID)
		} else if r.Method == http.MethodPut {
			if request.DecodeStrictJSON(r, &input) != nil {
				response.Error(w, "invalid JSON", 400)
				return
			}
			err = h.service.UpdateQuestion(domain.QuizQuestion{ID: questionID, TopicID: topicID, SortOrder: input.SortOrder, Text: input.Text, Explanation: input.Explanation})
		} else {
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
	if r.Method == http.MethodDelete {
		err = h.service.DeleteOption(topicID, questionID, id)
	} else if r.Method == http.MethodPut {
		if request.DecodeStrictJSON(r, &input) != nil {
			response.Error(w, "invalid JSON", 400)
			return
		}
		err = h.service.UpdateOption(topicID, domain.QuizOption{ID: id, QuestionID: questionID, SortOrder: input.SortOrder, Text: input.Text, Correct: input.IsCorrect})
	} else {
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
