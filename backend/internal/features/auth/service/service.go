package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"errors"
	"strings"

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
