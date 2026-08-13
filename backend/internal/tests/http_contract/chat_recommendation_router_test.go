package http_contract_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/ratelimit"
	"anti-scam-trainer/backend/internal/core/server/router"
	authservice "anti-scam-trainer/backend/internal/features/auth/service"
	learningservice "anti-scam-trainer/backend/internal/features/learning/service"
	learninghttp "anti-scam-trainer/backend/internal/features/learning/transport/http"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPChatRecommendationReturnsRoleSpecificPublishedTopic(t *testing.T) {
	store := &chatRecommendationStore{topicsByRole: map[domain.UserRole][]domain.Topic{
		domain.UserRoleBuyer:  {{ID: 1, Slug: "buyer-phishing-links", UserRole: domain.UserRoleBuyer, Title: "Фишинговые ссылки", Description: "Описание", Status: domain.TopicStatusPublished}},
		domain.UserRoleSeller: {{ID: 2, Slug: "seller-external-links", UserRole: domain.UserRoleSeller, Title: "Внешние ссылки", Description: "Описание", Status: domain.TopicStatusPublished}},
	}}
	handler := router.New()
	handler.Register(router.V1, learninghttp.NewWithChatRecommendation(learningservice.New(store), nil).Routes())

	for _, role := range []string{"buyer", "seller"} {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/avito-chat/recommendations", strings.NewReader(`{"source":"avito_chat_demo","role":"`+role+`","messages":[{"role":"assistant","text":"Могу оформить заказ по внешней форме"},{"role":"user","text":"Проверю заказ внутри приложения"}],"risk_type":"phishing","risk_signals":["внешняя ссылка"]}`))
		request = request.WithContext(authservice.WithIdentity(request.Context(), authservice.Identity{UserID: 7}))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		wantTitle := "Фишинговые ссылки"
		if role == "seller" {
			wantTitle = "Внешние ссылки"
		}
		if body := recorder.Body.String(); recorder.Code != http.StatusOK || !strings.Contains(body, `"title":"`+wantTitle+`"`) || !strings.Contains(body, `"next_action"`) || strings.Contains(body, "policy") || strings.Contains(body, "rubric") {
			t.Fatalf("role=%s recommendation=(%d,%s)", role, recorder.Code, body)
		}
	}
}

func TestHTTPChatRecommendationRejectsSensitiveSnapshotAndRateLimits(t *testing.T) {
	store := &chatRecommendationStore{topicsByRole: map[domain.UserRole][]domain.Topic{
		domain.UserRoleBuyer: {{ID: 1, Slug: "buyer-phishing-links", UserRole: domain.UserRoleBuyer, Title: "Фишинговые ссылки", Description: "Описание", Status: domain.TopicStatusPublished}},
	}}
	limiter := ratelimit.New(ratelimit.Config{Limit: 1, Window: time.Minute, MaxBuckets: 10, IdleTTL: time.Minute}, time.Now)
	handler := router.New()
	handler.Register(router.V1, learninghttp.NewWithChatRecommendation(learningservice.New(store), limiter).Routes())
	call := func(body string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/avito-chat/recommendations", strings.NewReader(body))
		request = request.WithContext(authservice.WithIdentity(request.Context(), authservice.Identity{UserID: 7}))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder
	}
	unsafe := call(`{"source":"avito_chat_demo","role":"buyer","messages":[{"role":"assistant","text":"Откройте avito-order.example/74291"},{"role":"user","text":"Нет"}],"risk_type":"phishing"}`)
	if unsafe.Code != http.StatusBadRequest || !strings.Contains(unsafe.Body.String(), `"code":"VALIDATION_ERROR"`) || store.topicCalls != 0 {
		t.Fatalf("sensitive snapshot=(%d,%s), calls=%d", unsafe.Code, unsafe.Body.String(), store.topicCalls)
	}
	for _, sensitive := range []string{
		`Позвоните +7 (999) 123-45-67`,
		`Назовите код 123456`,
		`Введите карту 4111 1111 1111 1111`,
		`chat_id: 74291`,
	} {
		body := `{"source":"avito_chat_demo","role":"buyer","messages":[{"role":"assistant","text":"` + sensitive + `"},{"role":"user","text":"Нет"}],"risk_type":"phishing"}`
		if rejected := call(body); rejected.Code != http.StatusBadRequest || !strings.Contains(rejected.Body.String(), `"code":"VALIDATION_ERROR"`) || store.topicCalls != 0 {
			t.Fatalf("sensitive snapshot %q=(%d,%s), calls=%d", sensitive, rejected.Code, rejected.Body.String(), store.topicCalls)
		}
	}
	valid := `{"source":"avito_chat_demo","role":"buyer","messages":[{"role":"assistant","text":"Откройте внешнюю форму"},{"role":"user","text":"Сначала проверю заказ"}],"risk_type":"phishing"}`
	if first := call(valid); first.Code != http.StatusOK {
		t.Fatalf("first recommendation=(%d,%s)", first.Code, first.Body.String())
	}
	if second := call(valid); second.Code != http.StatusTooManyRequests || second.Header().Get("Retry-After") == "" || !strings.Contains(second.Body.String(), `"code":"RATE_LIMITED"`) {
		t.Fatalf("rate limit=(%d,%s)", second.Code, second.Body.String())
	}
}

func TestHTTPChatRecommendationValidatesBoundaryAndAvailability(t *testing.T) {
	store := &chatRecommendationStore{topicsByRole: map[domain.UserRole][]domain.Topic{
		domain.UserRoleBuyer: {{ID: 1, Slug: "buyer-phishing-links", UserRole: domain.UserRoleBuyer, Title: "Фишинговые ссылки", Description: "Описание", Status: domain.TopicStatusPublished}},
	}}
	handler := router.New()
	handler.Register(router.V1, learninghttp.NewWithChatRecommendation(learningservice.New(store), nil).Routes())
	call := func(body string, authenticated bool) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/avito-chat/recommendations", strings.NewReader(body))
		if authenticated {
			request = request.WithContext(authservice.WithIdentity(request.Context(), authservice.Identity{UserID: 7}))
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder
	}
	valid := `{"source":"avito_chat_demo","role":"buyer","messages":[{"role":"assistant","text":"Внешняя форма"},{"role":"user","text":"Проверю заказ"}],"risk_type":"phishing","risk_signals":["внешняя форма"]}`
	if unauthenticated := call(valid, false); unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated=(%d,%s)", unauthenticated.Code, unauthenticated.Body.String())
	}
	for _, body := range []string{
		`{"source":"other","role":"buyer","messages":[{"role":"assistant","text":"x"},{"role":"user","text":"y"}],"risk_type":"phishing"}`,
		`{"source":"avito_chat_demo","role":"buyer","messages":[{"role":"assistant","text":"x"}],"risk_type":"phishing"}`,
		`{"source":"avito_chat_demo","role":"buyer","messages":[{"role":"assistant","text":"1"},{"role":"user","text":"2"},{"role":"assistant","text":"3"},{"role":"user","text":"4"},{"role":"assistant","text":"5"},{"role":"user","text":"6"},{"role":"assistant","text":"7"}],"risk_type":"phishing"}`,
		`{"source":"avito_chat_demo","role":"buyer","messages":[{"role":"assistant","text":"x"},{"role":"user","text":"y"}],"risk_type":"","risk_signals":["1","2","3","4"]}`,
	} {
		if invalid := call(body, true); invalid.Code != http.StatusBadRequest || !strings.Contains(invalid.Body.String(), `"code":"VALIDATION_ERROR"`) {
			t.Fatalf("invalid boundary=(%d,%s)", invalid.Code, invalid.Body.String())
		}
	}
	response := call(valid, true)
	for _, key := range []string{`"topic":`, `"explanation":"Внешняя форма`, `"next_action":`, `"fallback":false`} {
		if !strings.Contains(response.Body.String(), key) {
			t.Fatalf("response schema lacks %s: %s", key, response.Body.String())
		}
	}
	store.topicsByRole[domain.UserRoleBuyer] = nil
	if unavailable := call(valid, true); unavailable.Code != http.StatusNotFound {
		t.Fatalf("unavailable topic=(%d,%s)", unavailable.Code, unavailable.Body.String())
	}
}

func TestChatRecommendationResumesOnlyAttemptInRecommendedTopic(t *testing.T) {
	store := &chatRecommendationStore{learningStore: learningStore{attemptID: 42}, topicsByRole: map[domain.UserRole][]domain.Topic{
		domain.UserRoleBuyer: {{ID: 1, Slug: "buyer-phishing-links", UserRole: domain.UserRoleBuyer, Status: domain.TopicStatusPublished, TheoryRead: true, QuizPassed: true, Levels: []domain.TopicLevelProgress{{Number: 1, Opened: true}}}},
	}}
	service := learningservice.New(store)
	command := learningservice.ChatRecommendationCommand{Source: "avito_chat_demo", Role: domain.UserRoleBuyer, Messages: []domain.DialogueMessage{{Role: domain.MessageRoleAssistant, Text: "Внешняя форма"}, {Role: domain.MessageRoleUser, Text: "Проверю заказ"}}, RiskType: "phishing"}
	recommendation, err := service.RecommendFromChat(7, command)
	if err != nil || recommendation.NextAction.Type != "resume_attempt" || recommendation.NextAction.AttemptID != 42 {
		t.Fatalf("resume=(%+v,%v)", recommendation, err)
	}
	store.attemptID = 0
	recommendation, err = service.RecommendFromChat(7, command)
	if err != nil || recommendation.NextAction.Type != "start_level" {
		t.Fatalf("topic action=(%+v,%v)", recommendation, err)
	}
}

func TestChatRecommendationUsesServerFallbackForUnmatchedRisk(t *testing.T) {
	store := &chatRecommendationStore{topicsByRole: map[domain.UserRole][]domain.Topic{
		domain.UserRoleBuyer: {{ID: 3, Slug: "buyer-too-good-offer", UserRole: domain.UserRoleBuyer, Title: "Слишком выгодное предложение", Description: "Описание", Status: domain.TopicStatusPublished}},
	}}
	service := learningservice.New(store)
	recommendation, err := service.RecommendFromChat(7, learningservice.ChatRecommendationCommand{Source: "avito_chat_demo", Role: domain.UserRoleBuyer, Messages: []domain.DialogueMessage{{Role: domain.MessageRoleAssistant, Text: "Торопят с решением"}, {Role: domain.MessageRoleUser, Text: "Возьму паузу"}}, RiskType: "unknown_risk"})
	if err != nil || !recommendation.IsFallback || recommendation.Topic.Slug != "buyer-too-good-offer" || recommendation.NextAction.Type != "read_theory" {
		t.Fatalf("fallback=(%+v,%v)", recommendation, err)
	}
}

func TestChatRecommendationMapsEveryExactRiskAndExcludesUnpublishedTopic(t *testing.T) {
	buyer := []domain.Topic{
		{ID: 1, Slug: "buyer-phishing-links", UserRole: domain.UserRoleBuyer, Status: domain.TopicStatusPublished}, {ID: 2, Slug: "buyer-prepayment", UserRole: domain.UserRoleBuyer, Status: domain.TopicStatusPublished}, {ID: 3, Slug: "buyer-fake-delivery", UserRole: domain.UserRoleBuyer, Status: domain.TopicStatusPublished}, {ID: 4, Slug: "buyer-off-platform", UserRole: domain.UserRoleBuyer, Status: domain.TopicStatusPublished}, {ID: 5, Slug: "buyer-sms-codes", UserRole: domain.UserRoleBuyer, Status: domain.TopicStatusPublished}, {ID: 6, Slug: "buyer-too-good-offer", UserRole: domain.UserRoleBuyer, Status: domain.TopicStatusPublished},
	}
	seller := []domain.Topic{
		{ID: 7, Slug: "seller-external-links", UserRole: domain.UserRoleSeller, Status: domain.TopicStatusPublished}, {ID: 8, Slug: "seller-fake-payment", UserRole: domain.UserRoleSeller, Status: domain.TopicStatusPublished}, {ID: 9, Slug: "seller-fake-delivery", UserRole: domain.UserRoleSeller, Status: domain.TopicStatusPublished}, {ID: 10, Slug: "seller-off-platform", UserRole: domain.UserRoleSeller, Status: domain.TopicStatusPublished}, {ID: 11, Slug: "seller-confirmation-codes", UserRole: domain.UserRoleSeller, Status: domain.TopicStatusPublished}, {ID: 12, Slug: "seller-pressure", UserRole: domain.UserRoleSeller, Status: domain.TopicStatusPublished},
	}
	store := &chatRecommendationStore{topicsByRole: map[domain.UserRole][]domain.Topic{domain.UserRoleBuyer: buyer, domain.UserRoleSeller: seller}}
	service := learningservice.New(store)
	for _, test := range []struct {
		role domain.UserRole
		risk string
		slug string
	}{
		{domain.UserRoleBuyer, "phishing", "buyer-phishing-links"}, {domain.UserRoleBuyer, "prepayment", "buyer-prepayment"}, {domain.UserRoleBuyer, "delivery", "buyer-fake-delivery"}, {domain.UserRoleBuyer, "external_messenger", "buyer-off-platform"}, {domain.UserRoleBuyer, "account_takeover", "buyer-sms-codes"}, {domain.UserRoleBuyer, "social_engineering", "buyer-too-good-offer"},
		{domain.UserRoleBuyer, "fake_payment", "buyer-phishing-links"},
		{domain.UserRoleSeller, "phishing", "seller-external-links"}, {domain.UserRoleSeller, "prepayment", "seller-fake-payment"}, {domain.UserRoleSeller, "fake_payment", "seller-fake-payment"}, {domain.UserRoleSeller, "delivery", "seller-fake-delivery"}, {domain.UserRoleSeller, "external_messenger", "seller-off-platform"}, {domain.UserRoleSeller, "account_takeover", "seller-confirmation-codes"}, {domain.UserRoleSeller, "social_engineering", "seller-pressure"},
	} {
		recommendation, err := service.RecommendFromChat(7, learningservice.ChatRecommendationCommand{Source: "avito_chat_demo", Role: test.role, Messages: []domain.DialogueMessage{{Role: domain.MessageRoleAssistant, Text: "Опасное предложение"}, {Role: domain.MessageRoleUser, Text: "Проверю условия"}}, RiskType: test.risk})
		if err != nil || recommendation.IsFallback || recommendation.Topic.Slug != test.slug {
			t.Fatalf("%s/%s=(%+v,%v)", test.role, test.risk, recommendation, err)
		}
	}
	store.topicsByRole[domain.UserRoleBuyer][0].Status = domain.TopicStatusArchived
	recommendation, err := service.RecommendFromChat(7, learningservice.ChatRecommendationCommand{Source: "avito_chat_demo", Role: domain.UserRoleBuyer, Messages: []domain.DialogueMessage{{Role: domain.MessageRoleAssistant, Text: "Опасное предложение"}, {Role: domain.MessageRoleUser, Text: "Проверю условия"}}, RiskType: "phishing"})
	if err != nil || !recommendation.IsFallback || recommendation.Topic.Slug != "buyer-too-good-offer" {
		t.Fatalf("archived topic fallback=(%+v,%v)", recommendation, err)
	}
}
