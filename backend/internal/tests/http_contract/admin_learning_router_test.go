package http_contract_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/server/router"
	authservice "anti-scam-trainer/backend/internal/features/auth/service"
	learningservice "anti-scam-trainer/backend/internal/features/learning/service"
	learninghttp "anti-scam-trainer/backend/internal/features/learning/transport/http"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestHTTPAdminTopicLifecycleAndNestedContent(t *testing.T) {
	store := &adminLearningStore{topics: map[int]domain.TopicContent{}}
	handler := router.New()
	handler.Register(router.V1, learninghttp.NewAdmin(learningservice.NewContent(store)).Routes())
	call := func(role domain.AccessRole, method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req = req.WithContext(authservice.WithIdentity(req.Context(), authservice.Identity{UserID: 1, AccessRole: role}))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}
	if denied := call(domain.AccessRoleUser, http.MethodGet, "/api/v1/admin/topics", ""); denied.Code != http.StatusForbidden {
		t.Fatalf("user status=%d", denied.Code)
	}
	created := call(domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/topics", `{"slug":"buyer-new","role":"buyer","title":"Новая тема","description":"Описание","sort_order":6}`)
	if created.Code != http.StatusCreated || !strings.Contains(created.Body.String(), `"status":"draft"`) {
		t.Fatalf("create=(%d,%s)", created.Code, created.Body.String())
	}
	theory := call(domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/topics/1/theory-blocks", `{"sort_order":1,"kind":"intro","title":"Введение","body":"Текст"}`)
	if theory.Code != http.StatusCreated {
		t.Fatalf("theory=(%d,%s)", theory.Code, theory.Body.String())
	}
	question := call(domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/topics/1/quiz-questions", `{"sort_order":1,"text":"Вопрос","explanation":"Разбор"}`)
	if question.Code != http.StatusCreated {
		t.Fatalf("question=(%d,%s)", question.Code, question.Body.String())
	}
	option := call(domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/topics/1/quiz-questions/1/options", `{"sort_order":1,"text":"Ответ","is_correct":true}`)
	if option.Code != http.StatusCreated {
		t.Fatalf("option=(%d,%s)", option.Code, option.Body.String())
	}
	detail := call(domain.AccessRoleAdmin, http.MethodGet, "/api/v1/admin/topics/1", "")
	if detail.Code != http.StatusOK || !strings.Contains(detail.Body.String(), `"is_correct":true`) {
		t.Fatalf("detail=(%d,%s)", detail.Code, detail.Body.String())
	}
	if publish := call(domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/topics/1/publish", ""); publish.Code != http.StatusConflict || !strings.Contains(publish.Body.String(), `"code":"CONTENT_CONFLICT"`) {
		t.Fatalf("incomplete publish=(%d,%s)", publish.Code, publish.Body.String())
	}
	for optionOrder := 2; optionOrder <= 4; optionOrder++ {
		response := call(domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/topics/1/quiz-questions/1/options", `{"sort_order":`+strconv.Itoa(optionOrder)+`,"text":"Ответ","is_correct":false}`)
		if response.Code != http.StatusCreated {
			t.Fatalf("first question option %d=%d %s", optionOrder, response.Code, response.Body.String())
		}
	}
	kinds := []string{"intro", "risk", "example", "safe_action", "summary"}
	for order := 2; order <= 5; order++ {
		block := call(domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/topics/1/theory-blocks", `{"sort_order":`+strconv.Itoa(order)+`,"kind":"`+kinds[order-1]+`","title":"Блок","body":"Текст"}`)
		if block.Code != http.StatusCreated {
			t.Fatalf("block %d=%d %s", order, block.Code, block.Body.String())
		}
		question := call(domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/topics/1/quiz-questions", `{"sort_order":`+strconv.Itoa(order)+`,"text":"Вопрос","explanation":"Разбор"}`)
		if question.Code != http.StatusCreated {
			t.Fatalf("question %d=%d %s", order, question.Code, question.Body.String())
		}
		for optionOrder := 1; optionOrder <= 4; optionOrder++ {
			correct := "false"
			if optionOrder == 1 {
				correct = "true"
			}
			option := call(domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/topics/1/quiz-questions/"+strconv.Itoa(order)+"/options", `{"sort_order":`+strconv.Itoa(optionOrder)+`,"text":"Ответ","is_correct":`+correct+`}`)
			if option.Code != http.StatusCreated {
				t.Fatalf("question %d option %d=%d %s", order, optionOrder, option.Code, option.Body.String())
			}
		}
	}
	store.scenariosReady = true
	if publish := call(domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/topics/1/publish", ""); publish.Code != http.StatusNoContent {
		t.Fatalf("complete publish=(%d,%s)", publish.Code, publish.Body.String())
	}
	if mutation := call(domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/topics/1/theory-blocks", `{"sort_order":1,"kind":"intro","title":"Нельзя","body":"Нельзя"}`); mutation.Code != http.StatusConflict {
		t.Fatalf("published mutation=(%d,%s)", mutation.Code, mutation.Body.String())
	}
}

type adminLearningStore struct {
	topics         map[int]domain.TopicContent
	next           int
	scenariosReady bool
}

func (s *adminLearningStore) ListContent() ([]domain.Topic, error) {
	result := []domain.Topic{}
	for _, v := range s.topics {
		result = append(result, v.Topic)
	}
	return result, nil
}
func (s *adminLearningStore) Content(id int) (domain.TopicContent, error) {
	v, ok := s.topics[id]
	if !ok {
		return domain.TopicContent{}, learningservice.ErrTopicNotFound
	}
	return v, nil
}
func (s *adminLearningStore) CreateTopic(v domain.Topic) (domain.Topic, error) {
	s.next++
	v.ID = s.next
	v.Status = domain.TopicStatusDraft
	s.topics[v.ID] = domain.TopicContent{Topic: v}
	return v, nil
}
func (s *adminLearningStore) UpdateTopic(v domain.Topic) error {
	x := s.topics[v.ID]
	v.Status = x.Topic.Status
	x.Topic = v
	s.topics[v.ID] = x
	return nil
}
func (s *adminLearningStore) SetTopicStatus(id int, status string) error {
	x := s.topics[id]
	x.Topic.Status = status
	s.topics[id] = x
	return nil
}
func (s *adminLearningStore) CreateTheoryBlock(v domain.TheoryBlock) (domain.TheoryBlock, error) {
	x := s.topics[v.TopicID]
	v.ID = len(x.Theory) + 1
	x.Theory = append(x.Theory, v)
	s.topics[v.TopicID] = x
	return v, nil
}
func (s *adminLearningStore) UpdateTheoryBlock(domain.TheoryBlock) error { return nil }
func (s *adminLearningStore) DeleteTheoryBlock(int, int) error           { return nil }
func (s *adminLearningStore) CreateQuizQuestion(v domain.QuizQuestion) (domain.QuizQuestion, error) {
	x := s.topics[v.TopicID]
	v.ID = len(x.Quiz) + 1
	x.Quiz = append(x.Quiz, v)
	s.topics[v.TopicID] = x
	return v, nil
}
func (s *adminLearningStore) UpdateQuizQuestion(domain.QuizQuestion) error { return nil }
func (s *adminLearningStore) DeleteQuizQuestion(int, int) error            { return nil }
func (s *adminLearningStore) CreateQuizOption(v domain.QuizOption) (domain.QuizOption, error) {
	for id, x := range s.topics {
		for i, q := range x.Quiz {
			if q.ID == v.QuestionID {
				v.ID = len(q.Options) + 1
				x.Quiz[i].Options = append(x.Quiz[i].Options, v)
				s.topics[id] = x
				return v, nil
			}
		}
	}
	return domain.QuizOption{}, errors.New("missing question")
}
func (s *adminLearningStore) UpdateQuizOption(domain.QuizOption) error { return nil }
func (s *adminLearningStore) DeleteQuizOption(int, int) error          { return nil }
func (s *adminLearningStore) PublishTopic(id int) error {
	x := s.topics[id]
	if len(x.Theory) != 5 || len(x.Quiz) != 5 || !s.scenariosReady {
		return learningservice.ErrContentConflict
	}
	for _, question := range x.Quiz {
		correct := 0
		for _, option := range question.Options {
			if option.Correct {
				correct++
			}
		}
		if len(question.Options) != 4 || correct != 1 {
			return learningservice.ErrContentConflict
		}
	}
	x.Topic.Status = domain.TopicStatusPublished
	s.topics[id] = x
	return nil
}
