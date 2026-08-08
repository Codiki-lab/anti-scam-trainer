package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrNotFound           = apperrors.ErrUserNotFound
)

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Register(username, password string) (domain.User, error) {
	username = normalizeUsername(username)
	if username == "" || password == "" {
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
	return s.repository.Create(domain.User{Username: username, PasswordHash: string(hash), AccessRole: domain.AccessRoleUser})
}

func (s *Service) Authenticate(username, password string) (domain.User, error) {
	if username == "" || password == "" {
		return domain.User{}, ErrInvalidCredentials
	}
	user, err := s.repository.GetByUsername(normalizeUsername(username))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return domain.User{}, ErrInvalidCredentials
	}
	return user, nil
}

func (s *Service) GetByID(id int) (domain.User, error) { return s.repository.GetByID(id) }

func (s *Service) EnsureAdmin(username, password string) (domain.User, error) {
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
	return s.repository.Create(domain.User{Username: username, PasswordHash: string(hash), AccessRole: domain.AccessRoleAdmin})
}

func normalizeUsername(username string) string { return strings.ToLower(username) }
