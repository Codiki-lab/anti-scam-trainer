package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"time"
)

type Repository interface {
	Topics(userID int, role domain.UserRole) ([]domain.Topic, error)
	Topic(userID, topicID int) (domain.Topic, error)
	Theory(topicID int) ([]domain.TheoryBlock, error)
	MarkTheoryRead(userID, topicID int, activityDate time.Time) (domain.Streak, bool, error)
	Quiz(topicID int) ([]domain.QuizQuestion, error)
	SubmitQuiz(userID, topicID int, answers []domain.QuizAnswer, activityDate time.Time) (domain.QuizResult, error)
	RecentAttempts(userID int, role domain.UserRole) ([]domain.RecentAttempt, float64, error)
	Achievements(userID int) ([]domain.Achievement, error)
	User(userID int) (domain.User, error)
	InProgressAttempt(userID int, role domain.UserRole) (int, int, int, error)
}
