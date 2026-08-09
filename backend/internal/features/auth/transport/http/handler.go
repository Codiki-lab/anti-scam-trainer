package http

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/ratelimit"
	"anti-scam-trainer/backend/internal/core/server/request"
	"anti-scam-trainer/backend/internal/core/server/response"
	"anti-scam-trainer/backend/internal/core/server/router"
	auth "anti-scam-trainer/backend/internal/features/auth/service"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const AccessTokenCookie = "access_token"

type Handler struct {
	service                           *auth.Service
	registrationLimiter, loginLimiter *ratelimit.Limiter
	clientIP                          *ratelimit.ClientIPResolver
}

type credentialsDTO struct {
	Username     string          `json:"username"`
	Password     string          `json:"password"`
	TrainingRole domain.UserRole `json:"training_role"`
}

type accountDTO struct {
	ID           int             `json:"id"`
	Username     string          `json:"username"`
	AccessRole   string          `json:"access_role"`
	TrainingRole domain.UserRole `json:"training_role"`
	Streak       domain.Streak   `json:"streak"`
}

func New(service *auth.Service) *Handler { return &Handler{service: service} }
func NewWithRateLimits(service *auth.Service, registration, login *ratelimit.Limiter, clientIP *ratelimit.ClientIPResolver) *Handler {
	return &Handler{service: service, registrationLimiter: registration, loginLimiter: login, clientIP: clientIP}
}

func (h *Handler) Routes() []router.Route {
	return []router.Route{
		{Path: "/auth/register", Handler: h.register},
		{Path: "/auth/login", Handler: h.login},
		{Path: "/auth/logout", Handler: h.logout},
		{Path: "/auth/me", Handler: h.me},
		{Path: "/profile/preferences", Handler: h.preferences},
	}
}

func (h *Handler) register(writer http.ResponseWriter, httpRequest *http.Request) {
	if httpRequest.Method != http.MethodPost {
		response.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var credentials credentialsDTO
	if err := request.DecodeStrictJSON(httpRequest, &credentials); err != nil {
		response.Error(writer, "invalid JSON", http.StatusBadRequest)
		return
	}
	if h.registrationLimiter != nil {
		if ok, retry := h.registrationLimiter.Allow(h.clientIP.ClientIP(httpRequest)); !ok {
			rateLimited(writer, retry)
			return
		}
	}
	user, err := h.service.Register(credentials.Username, credentials.Password, credentials.TrainingRole)
	if err != nil {
		handleCredentialsError(writer, err)
		return
	}
	response.JSONStatus(writer, accountFromDomain(user), http.StatusCreated)
}

func (h *Handler) preferences(writer http.ResponseWriter, httpRequest *http.Request) {
	if httpRequest.Method != http.MethodPatch {
		response.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	identity, ok := auth.IdentityFromContext(httpRequest.Context())
	if !ok {
		response.Error(writer, "unauthorized", http.StatusUnauthorized)
		return
	}
	var input struct {
		TrainingRole domain.UserRole `json:"training_role"`
	}
	if err := request.DecodeStrictJSON(httpRequest, &input); err != nil {
		response.Error(writer, "invalid JSON", http.StatusBadRequest)
		return
	}
	user, err := h.service.UpdateTrainingRole(identity, input.TrainingRole)
	if err != nil {
		response.Error(writer, "training_role must be buyer or seller", http.StatusBadRequest)
		return
	}
	response.JSON(writer, accountFromDomain(user))
}

func (h *Handler) login(writer http.ResponseWriter, httpRequest *http.Request) {
	if httpRequest.Method != http.MethodPost {
		response.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var credentials credentialsDTO
	if err := request.DecodeStrictJSON(httpRequest, &credentials); err != nil {
		response.Error(writer, "invalid JSON", http.StatusBadRequest)
		return
	}
	if h.loginLimiter != nil {
		key := h.clientIP.ClientIP(httpRequest) + ":" + strings.ToLower(strings.TrimSpace(credentials.Username))
		if ok, retry := h.loginLimiter.Allow(key); !ok {
			rateLimited(writer, retry)
			return
		}
	}
	token, err := h.service.Login(credentials.Username, credentials.Password)
	if err != nil {
		response.Error(writer, "invalid credentials", http.StatusUnauthorized)
		return
	}
	http.SetCookie(writer, &http.Cookie{Name: AccessTokenCookie, Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: httpRequest.TLS != nil, Expires: time.Now().Add(auth.TokenLifetime), MaxAge: int(auth.TokenLifetime.Seconds())})
	writer.WriteHeader(http.StatusNoContent)
}

func rateLimited(writer http.ResponseWriter, retry time.Duration) {
	seconds := int64((retry + time.Second - 1) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	writer.Header().Set("Retry-After", strconv.FormatInt(seconds, 10))
	response.ErrorCode(writer, "RATE_LIMITED", "too many requests", http.StatusTooManyRequests, map[string]any{"retry_after_seconds": seconds})
}

func (h *Handler) logout(writer http.ResponseWriter, httpRequest *http.Request) {
	if httpRequest.Method != http.MethodPost {
		response.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.SetCookie(writer, &http.Cookie{Name: AccessTokenCookie, Value: "", Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: httpRequest.TLS != nil, MaxAge: -1, Expires: time.Unix(1, 0)})
	writer.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(writer http.ResponseWriter, httpRequest *http.Request) {
	if httpRequest.Method != http.MethodGet {
		response.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	identity, ok := auth.IdentityFromContext(httpRequest.Context())
	if !ok {
		response.Error(writer, "unauthorized", http.StatusUnauthorized)
		return
	}
	user, err := h.service.CurrentUser(identity)
	if err != nil {
		response.Error(writer, "unauthorized", http.StatusUnauthorized)
		return
	}
	response.JSON(writer, accountFromDomain(user))
}

func handleCredentialsError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrUsernameTaken):
		response.Error(writer, "username already taken", http.StatusConflict)
	case errors.Is(err, auth.ErrInvalidCredentials):
		response.Error(writer, "username and password are required", http.StatusBadRequest)
	default:
		response.Error(writer, "could not register", http.StatusInternalServerError)
	}
}

func accountFromDomain(user domain.User) accountDTO {
	return accountDTO{ID: user.ID, Username: user.Username, AccessRole: string(user.AccessRole), TrainingRole: user.TrainingRole, Streak: user.Streak}
}

func RequireAuthentication(tokens auth.Tokens) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if isPublicPath(request.URL.Path) {
				next.ServeHTTP(writer, request)
				return
			}
			cookie, err := request.Cookie(AccessTokenCookie)
			if err != nil {
				response.Error(writer, "unauthorized", http.StatusUnauthorized)
				return
			}
			identity, err := tokens.Parse(cookie.Value)
			if err != nil {
				response.Error(writer, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(writer, request.WithContext(auth.WithIdentity(request.Context(), identity)))
		})
	}
}

func isPublicPath(path string) bool {
	return path == "/api/v1/health" || path == "/api/v1/auth/register" || path == "/api/v1/auth/login" || path == "/api/v1/auth/logout" || path == "/swagger" || path == "/swagger/" || strings.HasPrefix(path, "/openapi/")
}
