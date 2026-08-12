package http_contract_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	authservice "anti-scam-trainer/backend/internal/features/auth/service"
)

type countingAccounts struct{ created int }

func (s *countingAccounts) Create(user domain.User) (domain.User, error) {
	s.created++
	user.ID = s.created
	return user, nil
}

func (*countingAccounts) GetByID(int) (domain.User, error) {
	return domain.User{}, authservice.ErrNotFound
}

func (*countingAccounts) GetByUsername(string) (domain.User, error) {
	return domain.User{}, authservice.ErrNotFound
}

func (*countingAccounts) UpdateTrainingRole(int, domain.UserRole) (domain.User, error) {
	return domain.User{}, nil
}

type fakeAccounts struct{}

func (*fakeAccounts) Create(user domain.User) (domain.User, error) {
	user.ID = 1
	return user, nil
}

func (*fakeAccounts) GetByID(int) (domain.User, error) { return domain.User{}, nil }

func (*fakeAccounts) GetByUsername(string) (domain.User, error) {
	return domain.User{}, authservice.ErrNotFound
}

func (*fakeAccounts) UpdateTrainingRole(int, domain.UserRole) (domain.User, error) {
	return domain.User{}, nil
}

type accountStore struct{ users map[string]domain.User }

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
	return domain.User{}, authservice.ErrNotFound
}

func (s *accountStore) GetByUsername(username string) (domain.User, error) {
	user, ok := s.users[username]
	if !ok {
		return domain.User{}, authservice.ErrNotFound
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

type fakeTokens struct{}

func (fakeTokens) Issue(domain.User) (string, error) { return "token", nil }

func (fakeTokens) Parse(string) (authservice.Identity, error) {
	return authservice.Identity{}, nil
}
