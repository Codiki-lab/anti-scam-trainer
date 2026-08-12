package http_contract_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/ratelimit"
	"anti-scam-trainer/backend/internal/core/server/middleware"
	"anti-scam-trainer/backend/internal/core/server/router"
	authservice "anti-scam-trainer/backend/internal/features/auth/service"
	authhttp "anti-scam-trainer/backend/internal/features/auth/transport/http"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestRouterRegistersUserWithoutExposingCredentialsOrUsersCRUD(t *testing.T) {
	accounts := authservice.NewAccounts(&fakeAccounts{})
	versionedRouter := router.New()
	versionedRouter.Register(router.V1, authhttp.New(authservice.New(accounts, fakeTokens{})).Routes())

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"username":"Alex","password":"secret","training_role":"buyer"}`))
	recorder := httptest.NewRecorder()
	versionedRouter.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if got, want := strings.TrimSpace(recorder.Body.String()), `{"id":1,"username":"alex","access_role":"user","training_role":"buyer","streak":{"current":0,"longest":0,"active_today":false}}`; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
	if strings.Contains(recorder.Body.String(), "password") {
		t.Fatalf("registration response leaks credentials: %q", recorder.Body.String())
	}

	oldUsersRequest := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	oldUsersRecorder := httptest.NewRecorder()
	versionedRouter.ServeHTTP(oldUsersRecorder, oldUsersRequest)
	if oldUsersRecorder.Code != http.StatusNotFound {
		t.Fatalf("old users endpoint status = %d, want %d", oldUsersRecorder.Code, http.StatusNotFound)
	}
}

func TestCredentialEndpointsRejectUnknownFieldsAndTrailingJSON(t *testing.T) {
	accounts := authservice.NewAccounts(&fakeAccounts{})
	r := router.New()
	r.Register(router.V1, authhttp.New(authservice.New(accounts, fakeTokens{})).Routes())
	for _, test := range []struct{ path, body string }{
		{"/api/v1/auth/register", `{"username":"Alex","password":"secret","training_role":"buyer","admin":true}`},
		{"/api/v1/auth/login", `{"username":"Alex","password":"secret"} {"extra":true}`},
	} {
		recorder := httptest.NewRecorder()
		r.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body)))
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s status=%d, want 400", test.path, recorder.Code)
		}
	}
}

func TestHTTPRegistrationRateLimitUsesRetryableEnvelopeBeforeAccountCreation(t *testing.T) {
	store := &countingAccounts{}
	accounts := authservice.NewAccounts(store)
	now := time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)
	limit := ratelimit.New(ratelimit.Config{Limit: 1, Window: time.Minute, MaxBuckets: 10, IdleTTL: time.Minute}, func() time.Time { return now })
	resolver, _ := ratelimit.NewClientIPResolver(nil)
	r := router.New()
	r.Register(router.V1, authhttp.NewWithRateLimits(authservice.New(accounts, fakeTokens{}), limit, limit, resolver).Routes())
	handler := middleware.RequestID()(r)
	request := func(username string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"username":"`+username+`","password":"secret","training_role":"buyer"}`))
		req.RemoteAddr = "192.0.2.1:9000"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}
	if first := request("first"); first.Code != http.StatusCreated {
		t.Fatalf("first=%d %s", first.Code, first.Body.String())
	}
	second := request("second")
	if second.Code != http.StatusTooManyRequests || second.Header().Get("Retry-After") == "" || second.Header().Get("X-Request-ID") == "" || !strings.Contains(second.Body.String(), `"code":"RATE_LIMITED"`) || store.created != 1 {
		t.Fatalf("limited=(%d,%s,created=%d)", second.Code, second.Body.String(), store.created)
	}
}

func TestAuthRoutesUseCookieIdentityForProfileAndPreferences(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	store := &accountStore{users: map[string]domain.User{"alex": {ID: 1, Username: "alex", PasswordHash: string(hash), AccessRole: domain.AccessRoleUser, TrainingRole: domain.UserRoleBuyer}}}
	accounts := authservice.NewAccounts(store)
	tokens, err := authservice.NewJWTManager("test-secret")
	if err != nil {
		t.Fatal(err)
	}
	routes := router.New()
	routes.Register(router.V1, authhttp.New(authservice.New(accounts, tokens)).Routes())
	handler := authhttp.RequireAuthentication(tokens)(routes)

	login := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"Alex","password":"secret"}`))
	loginRecorder := httptest.NewRecorder()
	handler.ServeHTTP(loginRecorder, login)
	if loginRecorder.Code != http.StatusNoContent {
		t.Fatalf("login status = %d, want %d", loginRecorder.Code, http.StatusNoContent)
	}
	cookie := loginRecorder.Result().Cookies()[0]
	me := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	me.AddCookie(cookie)
	meRecorder := httptest.NewRecorder()
	handler.ServeHTTP(meRecorder, me)
	if meRecorder.Code != http.StatusOK || !strings.Contains(meRecorder.Body.String(), `"username":"alex"`) {
		t.Fatalf("me = (%d, %q)", meRecorder.Code, meRecorder.Body.String())
	}

	preferences := httptest.NewRequest(http.MethodPatch, "/api/v1/profile/preferences", strings.NewReader(`{"training_role":"seller"}`))
	preferences.AddCookie(cookie)
	preferencesRecorder := httptest.NewRecorder()
	handler.ServeHTTP(preferencesRecorder, preferences)
	if preferencesRecorder.Code != http.StatusOK || !strings.Contains(preferencesRecorder.Body.String(), `"training_role":"seller"`) {
		t.Fatalf("preferences = (%d, %q)", preferencesRecorder.Code, preferencesRecorder.Body.String())
	}
}
