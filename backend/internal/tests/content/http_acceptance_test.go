package content_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/server/router"
	authservice "anti-scam-trainer/backend/internal/features/auth/service"
	learningrepository "anti-scam-trainer/backend/internal/features/learning/repository"
	learningservice "anti-scam-trainer/backend/internal/features/learning/service"
	learninghttp "anti-scam-trainer/backend/internal/features/learning/transport/http"
	scenariosrepository "anti-scam-trainer/backend/internal/features/scenarios/repository"
	scenariosservice "anti-scam-trainer/backend/internal/features/scenarios/service"
	scenarioshttp "anti-scam-trainer/backend/internal/features/scenarios/transport/http"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-pg/pg"
)

func TestPostgresHTTPPublishesAndServesCompleteTopicAggregate(t *testing.T) {
	database := os.Getenv("POSTGRES_TEST_NAME")
	if database == "" {
		t.Skip("POSTGRES_TEST_NAME is not set")
	}
	db := pg.Connect(&pg.Options{Addr: os.Getenv("POSTGRES_HOST") + ":" + os.Getenv("POSTGRES_PORT"), User: os.Getenv("POSTGRES_USER"), Password: os.Getenv("POSTGRES_PASSWORD"), Database: database})
	defer func() { _ = db.Close() }()

	learningRepo := learningrepository.NewPostgres(db)
	scenarioRepo := scenariosrepository.NewPostgres(db)
	handler := router.New()
	handler.Register(router.V1, learninghttp.NewAdmin(learningservice.NewContent(learningRepo)).Routes())
	handler.Register(router.V1, scenarioshttp.New(scenariosservice.New(scenarioRepo)).Routes())
	handler.Register(router.V1, learninghttp.New(learningservice.New(learningRepo)).Routes())

	call := func(identity authservice.Identity, method, path, body string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		request = request.WithContext(authservice.WithIdentity(request.Context(), identity))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder
	}
	admin := authservice.Identity{UserID: 1, AccessRole: domain.AccessRoleAdmin}
	var displacedTopicID int
	if _, err := db.QueryOne(pg.Scan(&displacedTopicID), `SELECT id FROM topics WHERE user_role='buyer' AND sort_order=6 AND content_status='published'`); err != nil {
		t.Fatal(err)
	}
	deactivated := call(admin, http.MethodPost, fmt.Sprintf("/api/v1/admin/topics/%d/deactivate", displacedTopicID), "")
	assertHTTPStatus(t, deactivated, http.StatusNoContent, "deactivate displaced topic")
	defer func() {
		_, _ = db.Exec(`UPDATE topics SET content_status='published',archived_at=NULL WHERE id=?`, displacedTopicID)
	}()
	created := call(admin, http.MethodPost, "/api/v1/admin/topics", `{"slug":"buyer-http-acceptance","role":"buyer","title":"HTTP acceptance","description":"Полный агрегат","sort_order":6}`)
	assertHTTPStatus(t, created, http.StatusCreated, "create topic")
	topicID := responseID(t, created)
	defer func() {
		_, _ = db.Exec(`DELETE FROM chats WHERE topic_id=?`, topicID)
		_, _ = db.Exec(`DELETE FROM topics WHERE id=?`, topicID)
	}()

	kinds := []string{"intro", "risk", "example", "safe_action", "summary"}
	answers := make([]string, 0, 5)
	for order := 1; order <= 5; order++ {
		block := call(admin, http.MethodPost, fmt.Sprintf("/api/v1/admin/topics/%d/theory-blocks", topicID), fmt.Sprintf(`{"sort_order":%d,"kind":%q,"title":"Блок","body":"Текст"}`, order, kinds[order-1]))
		assertHTTPStatus(t, block, http.StatusCreated, "create theory")
		question := call(admin, http.MethodPost, fmt.Sprintf("/api/v1/admin/topics/%d/quiz-questions", topicID), fmt.Sprintf(`{"sort_order":%d,"text":"Вопрос","explanation":"Разбор"}`, order))
		assertHTTPStatus(t, question, http.StatusCreated, "create question")
		questionID := responseID(t, question)
		for optionOrder := 1; optionOrder <= 4; optionOrder++ {
			correct := optionOrder == 1
			option := call(admin, http.MethodPost, fmt.Sprintf("/api/v1/admin/topics/%d/quiz-questions/%d/options", topicID, questionID), fmt.Sprintf(`{"sort_order":%d,"text":"Ответ","is_correct":%t}`, optionOrder, correct))
			assertHTTPStatus(t, option, http.StatusCreated, "create quiz option")
			if correct {
				answers = append(answers, fmt.Sprintf(`{"question_id":%d,"option_id":%d}`, questionID, responseID(t, option)))
			}
		}
	}

	for level := 1; level <= 4; level++ {
		scenario := call(admin, http.MethodPost, "/api/v1/admin/scenarios", fmt.Sprintf(`{"title":"Сценарий %d","description":"Описание","level_id":%d,"topic_id":%d,"role":"buyer","scam_scheme":"phishing","risk_type":"phishing","product_context":{"item_title":"Смартфон","category":"Электроника","deal_method":"delivery","price":42000,"currency":"RUB","image_key":"smartphone"},"ai_system_prompt":"prompt","final_rubric":{}}`, level, level, topicID))
		assertHTTPStatus(t, scenario, http.StatusCreated, "create scenario")
		scenarioID := responseID(t, scenario)
		storedScenario, err := scenarioRepo.ContentScenario(scenarioID)
		if err != nil || storedScenario.Status != domain.ScenarioStatusDraft {
			t.Fatalf("stored scenario %d = (%+v,%v)", scenarioID, storedScenario, err)
		}
		stepCount := map[int]int{1: 3, 2: 2, 3: 3, 4: 5}[level]
		for stepNumber := 1; stepNumber <= stepCount; stepNumber++ {
			responseType := "multiple_choice"
			aiInstruction, fallback := "", ""
			if level == 2 {
				responseType = "similar_choice"
			}
			if level == 3 && stepNumber >= 2 {
				responseType, aiInstruction, fallback = "free_text", "Оцени ответ", "Попробуйте ещё раз"
			}
			if level == 4 {
				responseType, aiInstruction, fallback = "free_text", "Оцени ответ", "Попробуйте ещё раз"
			}
			step := call(admin, http.MethodPost, fmt.Sprintf("/api/v1/admin/scenarios/%d/steps", scenarioID), fmt.Sprintf(`{"number":%d,"response_type":%q,"goal":"Цель","counterparty_message":"Сообщение","max_points":100,"ai_instruction":%q,"fallback_message":%q}`, stepNumber, responseType, aiInstruction, fallback))
			assertHTTPStatus(t, step, http.StatusCreated, fmt.Sprintf("create scenario step %d/%d (%s)", level, stepNumber, responseType))
			if level <= 2 || (level == 3 && stepNumber == 1) {
				stepID := responseID(t, step)
				for optionOrder, points := range []int{0, 50, 100} {
					option := call(admin, http.MethodPost, fmt.Sprintf("/api/v1/admin/steps/%d/options", stepID), fmt.Sprintf(`{"text":"Вариант %d","explanation":"Разбор","points":%d,"sort_order":%d}`, optionOrder+1, points, optionOrder+1))
					assertHTTPStatus(t, option, http.StatusCreated, "create scenario option")
				}
			}
		}
		published := call(admin, http.MethodPost, fmt.Sprintf("/api/v1/admin/scenarios/%d/publish", scenarioID), "")
		assertHTTPStatus(t, published, http.StatusNoContent, "publish scenario")
	}

	published := call(admin, http.MethodPost, fmt.Sprintf("/api/v1/admin/topics/%d/publish", topicID), "")
	assertHTTPStatus(t, published, http.StatusNoContent, "publish topic")
	var userID int
	if _, err := db.QueryOne(pg.Scan(&userID), `INSERT INTO users(username,password_hash,access_role,training_role) VALUES('http-content-acceptance','hash','user','buyer') RETURNING id`); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = db.Exec(`DELETE FROM users WHERE id=?`, userID) }()
	user := authservice.Identity{UserID: userID, AccessRole: domain.AccessRoleUser}
	topics := call(user, http.MethodGet, "/api/v1/topics?role=buyer", "")
	assertHTTPStatus(t, topics, http.StatusOK, "list published topics")
	if !strings.Contains(topics.Body.String(), `"slug":"buyer-http-acceptance"`) {
		t.Fatalf("published topic is absent: %s", topics.Body.String())
	}
	theory := call(user, http.MethodGet, fmt.Sprintf("/api/v1/topics/%d/theory", topicID), "")
	assertHTTPStatus(t, theory, http.StatusOK, "read theory")
	read := call(user, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/theory/read", topicID), "")
	assertHTTPStatus(t, read, http.StatusOK, "mark theory read")
	quiz := call(user, http.MethodGet, fmt.Sprintf("/api/v1/topics/%d/quiz", topicID), "")
	assertHTTPStatus(t, quiz, http.StatusOK, "read quiz")
	if strings.Contains(quiz.Body.String(), "is_correct") || strings.Contains(quiz.Body.String(), "explanation") {
		t.Fatalf("public quiz leaked answers: %s", quiz.Body.String())
	}
	attempt := call(user, http.MethodPost, fmt.Sprintf("/api/v1/topics/%d/quiz/attempts", topicID), `{"answers":[`+strings.Join(answers, ",")+`]}`)
	assertHTTPStatus(t, attempt, http.StatusOK, "submit quiz")
	if !strings.Contains(attempt.Body.String(), `"passed":true`) {
		t.Fatalf("quiz was not passed: %s", attempt.Body.String())
	}
}

func TestPostgresTopicsPreservePublishedStatusForChatRecommendations(t *testing.T) {
	database := os.Getenv("POSTGRES_TEST_NAME")
	if database == "" {
		t.Skip("POSTGRES_TEST_NAME is not set")
	}
	db := pg.Connect(&pg.Options{Addr: os.Getenv("POSTGRES_HOST") + ":" + os.Getenv("POSTGRES_PORT"), User: os.Getenv("POSTGRES_USER"), Password: os.Getenv("POSTGRES_PASSWORD"), Database: database})
	defer func() { _ = db.Close() }()

	topics, err := learningrepository.NewPostgres(db).Topics(1, domain.UserRoleSeller)
	if err != nil {
		t.Fatal(err)
	}
	for _, topic := range topics {
		if topic.Slug == "seller-external-links" {
			if topic.Status != domain.TopicStatusPublished {
				t.Fatalf("seller-external-links status = %q, want %q", topic.Status, domain.TopicStatusPublished)
			}
			return
		}
	}
	t.Fatal("seller-external-links is absent from published seller topics")
}

func responseID(t *testing.T, recorder *httptest.ResponseRecorder) int {
	t.Helper()
	var response struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil || response.ID == 0 {
		t.Fatalf("response ID: body=%s err=%v", recorder.Body.String(), err)
	}
	return response.ID
}

func assertHTTPStatus(t *testing.T, recorder *httptest.ResponseRecorder, want int, operation string) {
	t.Helper()
	if recorder.Code != want {
		t.Fatalf("%s = (%d,%s), want %d", operation, recorder.Code, recorder.Body.String(), want)
	}
}
