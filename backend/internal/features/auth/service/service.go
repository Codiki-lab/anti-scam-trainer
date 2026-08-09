package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Identity struct {
	UserID     int
	AccessRole domain.AccessRole
}

type Tokens interface {
	Issue(domain.User) (string, error)
	Parse(string) (Identity, error)
}

type Service struct {
	accounts *Accounts
	tokens   Tokens
}

func New(accounts *Accounts, tokens Tokens) *Service {
	return &Service{accounts: accounts, tokens: tokens}
}

func (s *Service) Register(username, password string, trainingRole domain.UserRole) (domain.User, error) {
	return s.accounts.Register(username, password, trainingRole)
}

func (s *Service) UpdateTrainingRole(identity Identity, role domain.UserRole) (domain.User, error) {
	return s.accounts.UpdateTrainingRole(identity.UserID, role)
}

func (s *Service) Login(username, password string) (string, error) {
	user, err := s.accounts.Authenticate(username, password)
	if err != nil {
		return "", err
	}
	return s.tokens.Issue(user)
}

func (s *Service) CurrentUser(identity Identity) (domain.User, error) {
	return s.accounts.GetByID(identity.UserID)
}

func (s *Service) Parse(token string) (Identity, error) { return s.tokens.Parse(token) }

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrNotFound           = apperrors.ErrUserNotFound
)

// Accounts owns local account persistence within the auth feature.
type Accounts struct{ repository Repository }

func NewAccounts(repository Repository) *Accounts { return &Accounts{repository: repository} }

func (s *Accounts) Register(username, password string, trainingRole domain.UserRole) (domain.User, error) {
	username = normalizeUsername(username)
	if username == "" || password == "" || !domain.ValidUserRole(trainingRole) {
		return domain.User{}, ErrInvalidCredentials
	}
	if _, err := s.repository.GetByUsername(username); err == nil {
		return domain.User{}, ErrUsernameTaken
	} else if !errors.Is(err, ErrNotFound) {
		return domain.User{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}
	return s.repository.Create(domain.User{Username: username, PasswordHash: string(hash), AccessRole: domain.AccessRoleUser, TrainingRole: trainingRole})
}

func (s *Accounts) Authenticate(username, password string) (domain.User, error) {
	if username == "" || password == "" {
		return domain.User{}, ErrInvalidCredentials
	}
	user, err := s.repository.GetByUsername(normalizeUsername(username))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return domain.User{}, ErrInvalidCredentials
	}
	return user, nil
}
func (s *Accounts) GetByID(id int) (domain.User, error) { return s.repository.GetByID(id) }
func (s *Accounts) EnsureAdmin(username, password string) (domain.User, error) {
	username = normalizeUsername(username)
	if username == "" || password == "" {
		return domain.User{}, ErrInvalidCredentials
	}
	if user, err := s.repository.GetByUsername(username); err == nil {
		if user.AccessRole == domain.AccessRoleAdmin {
			return user, nil
		}
		return domain.User{}, errors.New("admin username is already assigned to a user")
	} else if !errors.Is(err, ErrNotFound) {
		return domain.User{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}
	return s.repository.Create(domain.User{Username: username, PasswordHash: string(hash), AccessRole: domain.AccessRoleAdmin, TrainingRole: domain.UserRoleBuyer})
}
func (s *Accounts) UpdateTrainingRole(userID int, role domain.UserRole) (domain.User, error) {
	if !domain.ValidUserRole(role) {
		return domain.User{}, ErrInvalidCredentials
	}
	return s.repository.UpdateTrainingRole(userID, role)
}
func normalizeUsername(username string) string { return strings.ToLower(username) }

type identityContextKey struct{}

func WithIdentity(requestContext context.Context, identity Identity) context.Context {
	return context.WithValue(requestContext, identityContextKey{}, identity)
}

func IdentityFromContext(requestContext context.Context) (Identity, bool) {
	identity, ok := requestContext.Value(identityContextKey{}).(Identity)
	return identity, ok
}

const TokenLifetime = 7 * 24 * time.Hour

type JWTManager struct{ secret []byte }

func NewJWTManager(secret string) (*JWTManager, error) {
	if secret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}
	return &JWTManager{secret: []byte(secret)}, nil
}

func (m *JWTManager) Issue(user domain.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":         strconv.Itoa(user.ID),
		"access_role": string(user.AccessRole),
		"iat":         time.Now().Unix(),
		"exp":         time.Now().Add(TokenLifetime).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *JWTManager) Parse(raw string) (Identity, error) {
	token, err := jwt.Parse(raw, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return Identity{}, errors.New("invalid access token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return Identity{}, errors.New("invalid access token claims")
	}
	if expiresAt, err := claims.GetExpirationTime(); err != nil || expiresAt == nil {
		return Identity{}, errors.New("invalid access token expiration")
	}
	if issuedAt, err := claims.GetIssuedAt(); err != nil || issuedAt == nil {
		return Identity{}, errors.New("invalid access token issue time")
	}
	subject, err := claims.GetSubject()
	if err != nil {
		return Identity{}, errors.New("invalid access token subject")
	}
	userID, err := strconv.Atoi(subject)
	if err != nil || userID < 1 {
		return Identity{}, errors.New("invalid access token subject")
	}
	role, ok := claims["access_role"].(string)
	if !ok || (domain.AccessRole(role) != domain.AccessRoleUser && domain.AccessRole(role) != domain.AccessRoleAdmin) {
		return Identity{}, fmt.Errorf("invalid access token role")
	}
	return Identity{UserID: userID, AccessRole: domain.AccessRole(role)}, nil
}
