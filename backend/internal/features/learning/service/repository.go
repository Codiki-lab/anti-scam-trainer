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

type DailyTaskRepository interface {
	FindDailyTask(userID int, activityDate time.Time) (domain.DailyTask, bool, error)
	DailyTask(userID int, activityDate time.Time, created domain.DailyTask) (domain.DailyTask, error)
	AnswerDailyTask(userID int, activityDate time.Time, answer bool) (domain.DailyTask, domain.Streak, error)
}

type RecommendationRepository interface {
	FindRecommendation(userID int, activityDate time.Time, role domain.UserRole) (domain.ContinueAction, bool, error)
	SaveRecommendation(userID int, activityDate time.Time, role domain.UserRole, action domain.ContinueAction) error
}

type SkillCheckRepository interface {
	StartSkillCheck(userID, topicID int) (domain.SkillCheck, error)
	SkillCheck(userID, checkID int) (domain.SkillCheck, error)
	AnswerSkillCheck(userID, checkID int, answer bool) (domain.SkillCheck, error)
}

type MistakeProfileRepository interface {
	MistakePatternStats(userID int, role domain.UserRole) ([]domain.MistakePatternStats, error)
}

type ContentRepository interface {
	ListContent() ([]domain.Topic, error)
	Content(id int) (domain.TopicContent, error)
	CreateTopic(domain.Topic) (domain.Topic, error)
	UpdateTopic(domain.Topic) error
	SetTopicStatus(id int, status string) error
	CreateTheoryBlock(domain.TheoryBlock) (domain.TheoryBlock, error)
	UpdateTheoryBlock(domain.TheoryBlock) error
	DeleteTheoryBlock(topicID, blockID int) error
	CreateQuizQuestion(domain.QuizQuestion) (domain.QuizQuestion, error)
	UpdateQuizQuestion(domain.QuizQuestion) error
	DeleteQuizQuestion(topicID, questionID int) error
	CreateQuizOption(domain.QuizOption) (domain.QuizOption, error)
	UpdateQuizOption(domain.QuizOption) error
	DeleteQuizOption(questionID, optionID int) error
	PublishTopic(id int) error
}
