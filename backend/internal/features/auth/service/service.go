package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	users "anti-scam-trainer/backend/internal/features/users/service"
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
	accounts *users.Service
	tokens   Tokens
}

func New(accounts *users.Service, tokens Tokens) *Service {
	return &Service{accounts: accounts, tokens: tokens}
}

func (s *Service) Register(username, password string) (domain.User, error) {
	return s.accounts.Register(username, password)
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
