package http_contract_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/server/router"
	authservice "anti-scam-trainer/backend/internal/features/auth/service"
	scenariosservice "anti-scam-trainer/backend/internal/features/scenarios/service"
	scenarioshttp "anti-scam-trainer/backend/internal/features/scenarios/transport/http"
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminHTTPContractGuardsContentAndPublishesValidatedDraft(t *testing.T) {
	repository := &httpContentRepository{scenarios: map[int]domain.Scenario{1: {ID: 1, Status: domain.ScenarioStatusDraft}}}
	versionedRouter := router.New()
	versionedRouter.Register(router.V1, scenarioshttp.NewAdmin(scenariosservice.NewContent(repository)).Routes())

	denied := serveAdmin(versionedRouter, domain.AccessRoleUser, http.MethodPost, "/api/v1/admin/scenarios/1/publish", nil)
	if denied.Code != http.StatusForbidden {
		t.Fatalf("user publish = %d, want 403", denied.Code)
	}
	invalid := serveAdmin(versionedRouter, domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/scenarios/1/publish", nil)
	if invalid.Code != http.StatusConflict {
		t.Fatalf("invalid draft publish = %d, want 409", invalid.Code)
	}

	repository.valid = true
	published := serveAdmin(versionedRouter, domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/scenarios/1/publish", nil)
	if published.Code != http.StatusNoContent || repository.scenarios[1].Status != domain.ScenarioStatusPublished {
		t.Fatalf("publish = (%d, %#v), want published", published.Code, repository.scenarios[1])
	}
	blockedEdit := serveAdmin(versionedRouter, domain.AccessRoleAdmin, http.MethodPut, "/api/v1/admin/scenarios/1", []byte(`{"title":"changed","description":"d","level_id":1,"role":"buyer"}`))
	if blockedEdit.Code != http.StatusConflict {
		t.Fatalf("published edit = %d, want 409", blockedEdit.Code)
	}
	deactivated := serveAdmin(versionedRouter, domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/scenarios/1/deactivate", nil)
	if deactivated.Code != http.StatusNoContent || repository.scenarios[1].Status != domain.ScenarioStatusDraft {
		t.Fatalf("deactivate = (%d, %#v), want draft", deactivated.Code, repository.scenarios[1])
	}
	archived := serveAdmin(versionedRouter, domain.AccessRoleAdmin, http.MethodDelete, "/api/v1/admin/scenarios/1", nil)
	if archived.Code != http.StatusNoContent || repository.scenarios[1].Status != domain.ScenarioStatusArchived {
		t.Fatalf("archive = (%d, %#v), want archived", archived.Code, repository.scenarios[1])
	}
	restored := serveAdmin(versionedRouter, domain.AccessRoleAdmin, http.MethodPost, "/api/v1/admin/scenarios/1/restore", nil)
	if restored.Code != http.StatusNoContent || repository.scenarios[1].Status != domain.ScenarioStatusDraft {
		t.Fatalf("restore = (%d, %#v), want draft", restored.Code, repository.scenarios[1])
	}
}

func serveAdmin(handler http.Handler, role domain.AccessRole, method, target string, body []byte) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, bytes.NewReader(body))
	request = request.WithContext(authservice.WithIdentity(request.Context(), authservice.Identity{UserID: 1, AccessRole: role}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

type httpContentRepository struct {
	scenarios map[int]domain.Scenario
	valid     bool
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
	return s, nil
}
func (r *httpContentRepository) CreateOption(o domain.ScenarioOption) (domain.ScenarioOption, error) {
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
