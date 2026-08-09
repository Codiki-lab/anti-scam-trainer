package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"errors"
	"time"
)

var (
	ErrTopicNotFound = apperrors.ErrScenarioNotFound
	ErrInvalidQuiz   = errors.New("invalid quiz submission")
)

type Service struct {
	repository Repository
	now        func() time.Time
}

func New(repository Repository) *Service { return &Service{repository: repository, now: time.Now} }

func (s *Service) Topics(userID int, role domain.UserRole) ([]domain.Topic, error) {
	if !domain.ValidUserRole(role) {
		return nil, ErrTopicNotFound
	}
	return s.repository.Topics(userID, role)
}
func (s *Service) Topic(userID, topicID int) (domain.Topic, error) {
	return s.repository.Topic(userID, topicID)
}
func (s *Service) Theory(userID, topicID int) (domain.Topic, []domain.TheoryBlock, error) {
	topic, err := s.repository.Topic(userID, topicID)
	if err != nil {
		return domain.Topic{}, nil, err
	}
	blocks, err := s.repository.Theory(topicID)
	return topic, blocks, err
}
func (s *Service) MarkTheoryRead(userID, topicID int) (domain.Streak, bool, error) {
	if _, err := s.repository.Topic(userID, topicID); err != nil {
		return domain.Streak{}, false, err
	}
	return s.repository.MarkTheoryRead(userID, topicID, s.activityDate())
}
func (s *Service) Quiz(userID, topicID int) ([]domain.QuizQuestion, error) {
	topic, err := s.repository.Topic(userID, topicID)
	if err != nil {
		return nil, err
	}
	if !topic.TheoryRead {
		return nil, apperrors.ErrForbidden
	}
	return s.repository.Quiz(topicID)
}
func (s *Service) SubmitQuiz(userID, topicID int, answers []domain.QuizAnswer) (domain.QuizResult, error) {
	if len(answers) != 5 {
		return domain.QuizResult{}, ErrInvalidQuiz
	}
	topic, err := s.repository.Topic(userID, topicID)
	if err != nil {
		return domain.QuizResult{}, err
	}
	if !topic.TheoryRead {
		return domain.QuizResult{}, apperrors.ErrForbidden
	}
	return s.repository.SubmitQuiz(userID, topicID, answers, s.activityDate())
}
func (s *Service) Progress(userID int, role domain.UserRole) ([]domain.Topic, []domain.RecentAttempt, float64, error) {
	topics, err := s.Topics(userID, role)
	if err != nil {
		return nil, nil, 0, err
	}
	recent, average, err := s.repository.RecentAttempts(userID, role)
	return topics, recent, average, err
}
func (s *Service) Achievements(userID int) ([]domain.Achievement, error) {
	return s.repository.Achievements(userID)
}
func (s *Service) Dashboard(userID int, role domain.UserRole) (domain.User, []domain.Topic, []domain.Achievement, *domain.ContinueAction, error) {
	user, err := s.repository.User(userID)
	if err != nil {
		return domain.User{}, nil, nil, nil, err
	}
	topics, err := s.Topics(userID, role)
	if err != nil {
		return domain.User{}, nil, nil, nil, err
	}
	achievements, err := s.repository.Achievements(userID)
	if err != nil {
		return domain.User{}, nil, nil, nil, err
	}
	action := s.continueAction(userID, role, topics)
	return user, topics, achievements, action, nil
}
func (s *Service) continueAction(userID int, role domain.UserRole, topics []domain.Topic) *domain.ContinueAction {
	attemptID, topicID, level, err := s.repository.InProgressAttempt(userID, role)
	if err == nil && attemptID != 0 {
		return &domain.ContinueAction{Type: "resume_attempt", AttemptID: attemptID, TopicID: topicID, Level: level}
	}
	for _, topic := range topics {
		if !topic.TheoryRead {
			return &domain.ContinueAction{Type: "read_theory", TopicID: topic.ID}
		}
	}
	for _, topic := range topics {
		if !topic.QuizPassed {
			return &domain.ContinueAction{Type: "take_quiz", TopicID: topic.ID}
		}
	}
	for _, topic := range topics {
		for _, item := range topic.Levels {
			if item.Opened && item.Stars == 0 {
				return &domain.ContinueAction{Type: "start_level", TopicID: topic.ID, Level: item.Number}
			}
		}
	}
	return nil
}
func (s *Service) activityDate() time.Time {
	location, _ := time.LoadLocation("Europe/Moscow")
	now := s.now().In(location)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
}
