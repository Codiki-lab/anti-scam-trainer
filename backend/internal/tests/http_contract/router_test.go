package http_contract_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/ratelimit"
	"anti-scam-trainer/backend/internal/core/server/middleware"
	"anti-scam-trainer/backend/internal/core/server/router"
	attemptsservice "anti-scam-trainer/backend/internal/features/attempts/service"
	attemptshttp "anti-scam-trainer/backend/internal/features/attempts/transport/http"
	authservice "anti-scam-trainer/backend/internal/features/auth/service"
	authhttp "anti-scam-trainer/backend/internal/features/auth/transport/http"
	usersservice "anti-scam-trainer/backend/internal/features/users/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func TestRouterRegistersUserWithoutExposingCredentialsOrUsersCRUD(t *testing.T) {
	accounts := usersservice.New(&fakeAccounts{})
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
	accounts := usersservice.New(&fakeAccounts{})
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
	accounts := usersservice.New(store)
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

type countingAccounts struct{ created int }

func (s *countingAccounts) Create(user domain.User) (domain.User, error) {
	s.created++
	user.ID = s.created
	return user, nil
}
func (*countingAccounts) GetByID(int) (domain.User, error) {
	return domain.User{}, usersservice.ErrNotFound
}
func (*countingAccounts) GetByUsername(string) (domain.User, error) {
	return domain.User{}, usersservice.ErrNotFound
}
func (*countingAccounts) UpdateTrainingRole(int, domain.UserRole) (domain.User, error) {
	return domain.User{}, nil
}

func TestRouterUsesCookieIdentityForCurrentUserAndAttempts(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	accountsStore := &accountStore{users: map[string]domain.User{"alex": {ID: 1, Username: "alex", PasswordHash: string(hash), AccessRole: domain.AccessRoleUser, TrainingRole: domain.UserRoleBuyer}}}
	accounts := usersservice.New(accountsStore)
	tokens, err := authservice.NewJWTManager("test-secret")
	if err != nil {
		t.Fatal(err)
	}
	versionedRouter := router.New()
	versionedRouter.Register(router.V1, authhttp.New(authservice.New(accounts, tokens)).Routes())
	attemptStore := &attemptStore{attempts: map[int]domain.Attempt{9: {ID: 9, UserID: 2, ScenarioID: 1, Status: domain.AttemptStatusInProgress}}}
	versionedRouter.Register(router.V1, attemptshttp.New(attemptsservice.New(attemptStore, noCompletion{})).Routes())
	handler := authhttp.RequireAuthentication(tokens)(versionedRouter)

	login := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"Alex","password":"secret"}`))
	loginRecorder := httptest.NewRecorder()
	handler.ServeHTTP(loginRecorder, login)
	if loginRecorder.Code != http.StatusNoContent {
		t.Fatalf("login status = %d, want %d", loginRecorder.Code, http.StatusNoContent)
	}
	loginCookie := loginRecorder.Result().Cookies()[0]
	if !loginCookie.HttpOnly || loginCookie.MaxAge != int(authservice.TokenLifetime.Seconds()) {
		t.Fatalf("login cookie = %#v, want an HttpOnly seven-day cookie", loginCookie)
	}

	me := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	me.AddCookie(loginCookie)
	meRecorder := httptest.NewRecorder()
	handler.ServeHTTP(meRecorder, me)
	if got, want := strings.TrimSpace(meRecorder.Body.String()), `{"id":1,"username":"alex","access_role":"user","training_role":"buyer","streak":{"current":0,"longest":0,"active_today":false}}`; meRecorder.Code != http.StatusOK || got != want {
		t.Fatalf("me = (%d, %q), want (%d, %q)", meRecorder.Code, got, http.StatusOK, want)
	}

	preferences := httptest.NewRequest(http.MethodPatch, "/api/v1/profile/preferences", strings.NewReader(`{"training_role":"seller"}`))
	preferences.AddCookie(loginCookie)
	preferencesRecorder := httptest.NewRecorder()
	handler.ServeHTTP(preferencesRecorder, preferences)
	if body := strings.TrimSpace(preferencesRecorder.Body.String()); preferencesRecorder.Code != http.StatusOK || !strings.Contains(body, `"training_role":"seller"`) {
		t.Fatalf("preferences = (%d, %q), want saved seller role", preferencesRecorder.Code, body)
	}

	invalidPreferences := httptest.NewRequest(http.MethodPatch, "/api/v1/profile/preferences", strings.NewReader(`{"training_role":"admin"}`))
	invalidPreferences.AddCookie(loginCookie)
	invalidPreferencesRecorder := httptest.NewRecorder()
	handler.ServeHTTP(invalidPreferencesRecorder, invalidPreferences)
	if invalidPreferencesRecorder.Code != http.StatusBadRequest {
		t.Fatalf("invalid preferences = %d, want %d", invalidPreferencesRecorder.Code, http.StatusBadRequest)
	}

	withoutCookie := httptest.NewRecorder()
	handler.ServeHTTP(withoutCookie, httptest.NewRequest(http.MethodGet, "/api/v1/attempts", nil))
	if withoutCookie.Code != http.StatusUnauthorized {
		t.Fatalf("attempts without cookie = %d, want %d", withoutCookie.Code, http.StatusUnauthorized)
	}

	createAttempt := httptest.NewRequest(http.MethodPost, "/api/v1/attempts", strings.NewReader(`{"user_id":99,"scenario_id":1,"status":"IN_PROGRESS"}`))
	createAttempt.AddCookie(loginCookie)
	createdRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createdRecorder, createAttempt)
	if got := strings.TrimSpace(createdRecorder.Body.String()); createdRecorder.Code != http.StatusOK || strings.Contains(got, "user_id") || attemptStore.attempts[11].UserID != 1 {
		t.Fatalf("created attempt = (%d, %q), want owner from JWT without user_id in API", createdRecorder.Code, got)
	}

	foreignAttempt := httptest.NewRequest(http.MethodGet, "/api/v1/attempts/9", nil)
	foreignAttempt.AddCookie(loginCookie)
	foreignRecorder := httptest.NewRecorder()
	handler.ServeHTTP(foreignRecorder, foreignAttempt)
	if foreignRecorder.Code != http.StatusNotFound {
		t.Fatalf("foreign attempt = %d, want %d", foreignRecorder.Code, http.StatusNotFound)
	}

	expiredToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "1", "access_role": "user", "iat": time.Now().Add(-2 * time.Hour).Unix(), "exp": time.Now().Add(-time.Hour).Unix()}).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	expiredRequest := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	expiredRequest.AddCookie(&http.Cookie{Name: authhttp.AccessTokenCookie, Value: expiredToken})
	expiredRecorder := httptest.NewRecorder()
	handler.ServeHTTP(expiredRecorder, expiredRequest)
	if expiredRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("expired token = %d, want %d", expiredRecorder.Code, http.StatusUnauthorized)
	}

	invalidRequest := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	invalidRequest.AddCookie(&http.Cookie{Name: authhttp.AccessTokenCookie, Value: "not-a-jwt"})
	invalidRecorder := httptest.NewRecorder()
	handler.ServeHTTP(invalidRecorder, invalidRequest)
	if invalidRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("invalid token = %d, want %d", invalidRecorder.Code, http.StatusUnauthorized)
	}

	logout := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutRecorder := httptest.NewRecorder()
	handler.ServeHTTP(logoutRecorder, logout)
	if logoutRecorder.Code != http.StatusNoContent || logoutRecorder.Result().Cookies()[0].MaxAge >= 0 {
		t.Fatalf("logout = (%d, %#v), want expired cookie", logoutRecorder.Code, logoutRecorder.Result().Cookies())
	}
}

type fakeAccounts struct{}

func (*fakeAccounts) Create(user domain.User) (domain.User, error) {
	user.ID = 1
	return user, nil
}

func (*fakeAccounts) GetByID(int) (domain.User, error) { return domain.User{}, nil }
func (*fakeAccounts) GetByUsername(string) (domain.User, error) {
	return domain.User{}, usersservice.ErrNotFound
}
func (*fakeAccounts) UpdateTrainingRole(int, domain.UserRole) (domain.User, error) {
	return domain.User{}, nil
}

type fakeTokens struct{}

func (fakeTokens) Issue(domain.User) (string, error) { return "token", nil }
func (fakeTokens) Parse(string) (authservice.Identity, error) {
	return authservice.Identity{}, nil
}

type accountStore struct {
	users map[string]domain.User
}

func (s *accountStore) Create(user domain.User) (domain.User, error) {
	user.ID = len(s.users) + 1
	s.users[user.Username] = user
	return user, nil
}
func (s *accountStore) GetByID(id int) (domain.User, error) {
	for _, user := range s.users {
		if user.ID == id {
			return user, nil
		}
	}
	return domain.User{}, usersservice.ErrNotFound
}
func (s *accountStore) GetByUsername(username string) (domain.User, error) {
	user, ok := s.users[username]
	if !ok {
		return domain.User{}, usersservice.ErrNotFound
	}
	return user, nil
}
func (s *accountStore) UpdateTrainingRole(id int, role domain.UserRole) (domain.User, error) {
	user, err := s.GetByID(id)
	if err != nil {
		return domain.User{}, err
	}
	user.TrainingRole = role
	s.users[user.Username] = user
	return user, nil
}

type attemptStore struct {
	attempts map[int]domain.Attempt
}

func (s *attemptStore) Create(attempt domain.Attempt) (domain.Attempt, error) {
	attempt.ID = len(s.attempts) + 10
	s.attempts[attempt.ID] = attempt
	return attempt, nil
}
func (s *attemptStore) GetByID(id int) (domain.Attempt, error) {
	attempt, ok := s.attempts[id]
	if !ok {
		return domain.Attempt{}, attemptsservice.ErrAttemptNotFound
	}
	return attempt, nil
}
func (s *attemptStore) Update(attempt domain.Attempt) error {
	s.attempts[attempt.ID] = attempt
	return nil
}
func (s *attemptStore) Delete(id int) error { delete(s.attempts, id); return nil }
func (s *attemptStore) ListByUserID(userID int) ([]domain.Attempt, error) {
	result := make([]domain.Attempt, 0, len(s.attempts))
	for _, attempt := range s.attempts {
		if attempt.UserID == userID {
			result = append(result, attempt)
		}
	}
	return result, nil
}

type noCompletion struct{}

func (noCompletion) InTransaction(func(attemptsservice.CompletionStore) error) error { return nil }
