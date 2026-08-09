package http

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/server/request"
	"anti-scam-trainer/backend/internal/core/server/response"
	"anti-scam-trainer/backend/internal/core/server/router"
	auth "anti-scam-trainer/backend/internal/features/auth/service"
	users "anti-scam-trainer/backend/internal/features/users/service"
	"errors"
	"net/http"
	"time"
)

const AccessTokenCookie = "access_token"

type Handler struct{ service *auth.Service }

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
	token, err := h.service.Login(credentials.Username, credentials.Password)
	if err != nil {
		response.Error(writer, "invalid credentials", http.StatusUnauthorized)
		return
	}
	http.SetCookie(writer, &http.Cookie{Name: AccessTokenCookie, Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: httpRequest.TLS != nil, Expires: time.Now().Add(auth.TokenLifetime), MaxAge: int(auth.TokenLifetime.Seconds())})
	writer.WriteHeader(http.StatusNoContent)
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
	case errors.Is(err, users.ErrUsernameTaken):
		response.Error(writer, "username already taken", http.StatusConflict)
	case errors.Is(err, users.ErrInvalidCredentials):
		response.Error(writer, "username and password are required", http.StatusBadRequest)
	default:
		response.Error(writer, "could not register", http.StatusInternalServerError)
	}
}

func accountFromDomain(user domain.User) accountDTO {
	return accountDTO{ID: user.ID, Username: user.Username, AccessRole: string(user.AccessRole), TrainingRole: user.TrainingRole, Streak: user.Streak}
}
