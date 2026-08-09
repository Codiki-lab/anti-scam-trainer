package http_contract_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/server/middleware"
	"anti-scam-trainer/backend/internal/core/server/router"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	attemptshttp "anti-scam-trainer/backend/internal/features/attempts/transport/http"
	authhttp "anti-scam-trainer/backend/internal/features/auth/transport/http"
	scenariosservice "anti-scam-trainer/backend/internal/features/scenarios/service"
	scenarioshttp "anti-scam-trainer/backend/internal/features/scenarios/transport/http"
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminHTTPContractGuardsContentAndPublishesValidatedDraft(t *testing.T) {
	repository := &httpContentRepository{scenarios: map[int]domain.Scenario{1: {ID: 1, LevelID: 1, UserRole: "buyer", Status: domain.ScenarioStatusDraft}}}
	versionedRouter := router.New()
	versionedRouter.Register(router.V1, scenarioshttp.NewAdmin(scenariosservice.NewContent(repository)).Routes())
	gameRepository := &contentAwareGameRepository{httpGameRepository: newHTTPGameRepository(), content: repository}
	versionedRouter.Register(router.V1, attemptshttp.NewGame(attemptsservice.NewGame(gameRepository)).Routes())
	handler := middleware.RequestID()(authhttp.RequireAuthentication(contractTokens{})(versionedRouter))
	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/admin/scenarios", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("missing cookie = %d, want 401", unauthorized.Code)
	}
	draftHidden := serveGame(handler, 1, http.MethodPost, "/api/v1/training/levels/1/start?role=buyer", nil)
	if draftHidden.Code != http.StatusNotFound {
		t.Fatalf("draft scenario start = %d, want 404", draftHidden.Code)
	}

	denied := serveAdmin(handler, domain.AccessRoleUser, http.MethodPost, "/api/v1/admin/scenarios/1/publish", nil)
	if denied.Code != http.StatusForbidden {
		t.Fatalf("user publish = %d, want 403", denied.Code)
	}
	invalid := serveAdmin(handler, domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/scenarios/1/publish", nil)
	if invalid.Code != http.StatusConflict {
		t.Fatalf("invalid draft publish = %d, want 409", invalid.Code)
	}

	repository.valid = true
	published := serveAdmin(handler, domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/scenarios/1/publish", nil)
	if published.Code != http.StatusNoContent || repository.scenarios[1].Status != domain.ScenarioStatusPublished {
		t.Fatalf("publish = (%d, %#v), want published", published.Code, repository.scenarios[1])
	}
	if published.Header().Get("X-Request-ID") == "" {
		t.Fatal("admin response has no X-Request-ID")
	}
	visible := serveGame(handler, 1, http.MethodPost, "/api/v1/training/levels/1/start?role=buyer", nil)
	if visible.Code != http.StatusOK {
		t.Fatalf("published scenario start = %d, want 200", visible.Code)
	}
	blockedEdit := serveAdmin(handler, domain.AccessRoleAdmin, http.MethodPut, "/api/v1/admin/scenarios/1", []byte(`{"title":"changed","description":"d","level_id":1,"role":"buyer"}`))
	if blockedEdit.Code != http.StatusConflict {
		t.Fatalf("published edit = %d, want 409", blockedEdit.Code)
	}
	deactivated := serveAdmin(handler, domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/scenarios/1/deactivate", nil)
	if deactivated.Code != http.StatusNoContent || repository.scenarios[1].Status != domain.ScenarioStatusDraft {
		t.Fatalf("deactivate = (%d, %#v), want draft", deactivated.Code, repository.scenarios[1])
	}
	draftAgainHidden := serveGame(handler, 1, http.MethodPost, "/api/v1/training/levels/1/start?role=buyer", nil)
	if draftAgainHidden.Code != http.StatusNotFound {
		t.Fatalf("deactivated scenario start = %d, want 404", draftAgainHidden.Code)
	}
	archived := serveAdmin(handler, domain.AccessRoleAdmin, http.MethodDelete, "/api/v1/admin/scenarios/1", nil)
	if archived.Code != http.StatusNoContent || repository.scenarios[1].Status != domain.ScenarioStatusArchived {
		t.Fatalf("archive = (%d, %#v), want archived", archived.Code, repository.scenarios[1])
	}
	archivedHidden := serveGame(handler, 1, http.MethodPost, "/api/v1/training/levels/1/start?role=buyer", nil)
	if archivedHidden.Code != http.StatusNotFound {
		t.Fatalf("archived scenario start = %d, want 404", archivedHidden.Code)
	}
	restored := serveAdmin(handler, domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/scenarios/1/restore", nil)
	if restored.Code != http.StatusNoContent || repository.scenarios[1].Status != domain.ScenarioStatusDraft {
		t.Fatalf("restore = (%d, %#v), want draft", restored.Code, repository.scenarios[1])
	}
	step := serveAdmin(handler, domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/scenarios/1/steps", []byte(`{"number":1,"response_type":"multiple_choice","goal":"choose","max_points":100}`))
	if step.Code != http.StatusCreated || !bytes.Contains(step.Body.Bytes(), []byte(`"id":10`)) {
		t.Fatalf("create step = (%d, %s)", step.Code, step.Body.String())
	}
	option := serveAdmin(handler, domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/steps/10/options", []byte(`{"text":"safe","explanation":"inside service","points":100,"sort_order":1}`))
	if option.Code != http.StatusCreated || !bytes.Contains(option.Body.Bytes(), []byte(`"id":20`)) {
		t.Fatalf("create option = (%d, %s)", option.Code, option.Body.String())
	}
	updatedStep := serveAdmin(handler, domain.AccessRoleAdmin, http.MethodPut, "/api/v1/admin/scenarios/1/steps/10", []byte(`{"number":1,"response_type":"multiple_choice","goal":"updated","max_points":100}`))
	updatedOption := serveAdmin(handler, domain.AccessRoleAdmin, http.MethodPut, "/api/v1/admin/steps/10/options/20", []byte(`{"text":"safe","explanation":"updated","points":100,"sort_order":1}`))
	deletedOption := serveAdmin(handler, domain.AccessRoleAdmin, http.MethodDelete, "/api/v1/admin/steps/10/options/20", nil)
	deletedStep := serveAdmin(handler, domain.AccessRoleAdmin, http.MethodDelete, "/api/v1/admin/scenarios/1/steps/10", nil)
	for name, response := range map[string]*httptest.ResponseRecorder{"update step": updatedStep, "update option": updatedOption, "delete option": deletedOption, "delete step": deletedStep} {
		if response.Code != http.StatusNoContent {
			t.Fatalf("%s = %d, want 204", name, response.Code)
		}
	}
}

func serveAdmin(handler http.Handler, role domain.AccessRole, method, target string, body []byte) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, bytes.NewReader(body))
	request.AddCookie(&http.Cookie{Name: authhttp.AccessTokenCookie, Value: string(role) + "-1"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

type httpContentRepository struct {
	scenarios map[int]domain.Scenario
	valid     bool
}

type contentAwareGameRepository struct {
	*httpGameRepository
	content *httpContentRepository
}

func (r *contentAwareGameRepository) PublishedScenario(level int, role string) (domain.Scenario, error) {
	for _, scenario := range r.content.scenarios {
		if scenario.LevelID == level && scenario.UserRole == role && scenario.Status == domain.ScenarioStatusPublished {
			return scenario, nil
		}
	}
	return domain.Scenario{}, errors.New("missing")
}

func (r *httpContentRepository) CreateContent(s domain.Scenario) (domain.Scenario, error) {
	s.ID = len(r.scenarios) + 1
	r.scenarios[s.ID] = s
	return s, nil
}
func (r *httpContentRepository) ListContent() ([]domain.Scenario, error) {
	result := []domain.Scenario{}
	for _, scenario := range r.scenarios {
		result = append(result, scenario)
	}
	return result, nil
}
func (r *httpContentRepository) UpdateContent(s domain.Scenario) error {
	r.scenarios[s.ID] = s
	return nil
}
func (r *httpContentRepository) SetContentStatus(id int, status string, _ bool) error {
	scenario, ok := r.scenarios[id]
	if !ok {
		return errors.New("missing")
	}
	scenario.Status = status
	r.scenarios[id] = scenario
	return nil
}
func (r *httpContentRepository) CreateStep(s domain.ScenarioStep) (domain.ScenarioStep, error) {
	s.ID = 10
	return s, nil
}
func (r *httpContentRepository) CreateOption(o domain.ScenarioOption) (domain.ScenarioOption, error) {
	o.ID = 20
	return o, nil
}
func (r *httpContentRepository) ContentScenario(id int) (domain.Scenario, error) {
	scenario, ok := r.scenarios[id]
	if !ok {
		return domain.Scenario{}, errors.New("missing")
	}
	return scenario, nil
}
func (r *httpContentRepository) ValidContent(int) (bool, error) { return r.valid, nil }
func (r *httpContentRepository) StepScenario(int) (domain.Scenario, error) {
	return r.scenarios[1], nil
}
func (r *httpContentRepository) OptionScenario(int) (domain.Scenario, error) {
	return r.scenarios[1], nil
}
func (*httpContentRepository) UpdateStep(domain.ScenarioStep) error     { return nil }
func (*httpContentRepository) DeleteStep(int) error                     { return nil }
func (*httpContentRepository) UpdateOption(domain.ScenarioOption) error { return nil }
func (*httpContentRepository) DeleteOption(int) error                   { return nil }
