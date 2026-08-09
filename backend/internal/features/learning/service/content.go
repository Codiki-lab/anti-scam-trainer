package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"errors"
	"strings"
)

var (
	ErrContentConflict = errors.New("content conflict")
	ErrInvalidContent  = errors.New("invalid content")
)

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

type ContentService struct{ repository ContentRepository }

func NewContent(repository ContentRepository) *ContentService {
	return &ContentService{repository: repository}
}

func (s *ContentService) List() ([]domain.Topic, error)           { return s.repository.ListContent() }
func (s *ContentService) Get(id int) (domain.TopicContent, error) { return s.repository.Content(id) }

func validateTopic(topic domain.Topic) error {
	if strings.TrimSpace(topic.Slug) == "" || strings.TrimSpace(topic.Title) == "" || strings.TrimSpace(topic.Description) == "" || !domain.ValidUserRole(topic.UserRole) || topic.SortOrder < 1 || topic.SortOrder > 6 {
		return ErrInvalidContent
	}
	return nil
}

func (s *ContentService) Create(topic domain.Topic) (domain.Topic, error) {
	if err := validateTopic(topic); err != nil {
		return domain.Topic{}, err
	}
	topic.Status = domain.TopicStatusDraft
	return s.repository.CreateTopic(topic)
}

func (s *ContentService) Update(topic domain.Topic) error {
	if err := validateTopic(topic); err != nil {
		return err
	}
	current, err := s.repository.Content(topic.ID)
	if err != nil {
		return err
	}
	if current.Topic.Status != domain.TopicStatusDraft {
		return ErrContentConflict
	}
	return s.repository.UpdateTopic(topic)
}

func (s *ContentService) Deactivate(id int) error {
	current, err := s.repository.Content(id)
	if err != nil {
		return err
	}
	if current.Topic.Status != domain.TopicStatusPublished {
		return ErrContentConflict
	}
	return s.repository.SetTopicStatus(id, domain.TopicStatusDraft)
}
func (s *ContentService) Archive(id int) error {
	current, err := s.repository.Content(id)
	if err != nil {
		return err
	}
	if current.Topic.Status == domain.TopicStatusArchived {
		return ErrContentConflict
	}
	return s.repository.SetTopicStatus(id, domain.TopicStatusArchived)
}
func (s *ContentService) Restore(id int) error {
	current, err := s.repository.Content(id)
	if err != nil {
		return err
	}
	if current.Topic.Status != domain.TopicStatusArchived {
		return ErrContentConflict
	}
	return s.repository.SetTopicStatus(id, domain.TopicStatusDraft)
}
func (s *ContentService) Publish(id int) error {
	current, err := s.repository.Content(id)
	if err != nil {
		return err
	}
	if current.Topic.Status != domain.TopicStatusDraft {
		return ErrContentConflict
	}
	return s.repository.PublishTopic(id)
}

func validTheory(block domain.TheoryBlock) bool {
	if block.SortOrder < 1 || block.SortOrder > 5 || strings.TrimSpace(block.Title) == "" || strings.TrimSpace(block.Body) == "" {
		return false
	}
	switch block.Kind {
	case "intro", "risk", "example", "safe_action", "summary":
		return true
	}
	return false
}
func (s *ContentService) AddTheory(block domain.TheoryBlock) (domain.TheoryBlock, error) {
	if !validTheory(block) {
		return domain.TheoryBlock{}, ErrInvalidContent
	}
	if err := s.requireDraft(block.TopicID); err != nil {
		return domain.TheoryBlock{}, err
	}
	return s.repository.CreateTheoryBlock(block)
}
func (s *ContentService) UpdateTheory(block domain.TheoryBlock) error {
	if !validTheory(block) {
		return ErrInvalidContent
	}
	if err := s.requireDraft(block.TopicID); err != nil {
		return err
	}
	return s.repository.UpdateTheoryBlock(block)
}
func (s *ContentService) DeleteTheory(topicID, blockID int) error {
	if err := s.requireDraft(topicID); err != nil {
		return err
	}
	return s.repository.DeleteTheoryBlock(topicID, blockID)
}

func validQuestion(question domain.QuizQuestion) bool {
	return question.SortOrder >= 1 && question.SortOrder <= 5 && strings.TrimSpace(question.Text) != "" && strings.TrimSpace(question.Explanation) != ""
}
func (s *ContentService) AddQuestion(question domain.QuizQuestion) (domain.QuizQuestion, error) {
	if !validQuestion(question) {
		return domain.QuizQuestion{}, ErrInvalidContent
	}
	if err := s.requireDraft(question.TopicID); err != nil {
		return domain.QuizQuestion{}, err
	}
	return s.repository.CreateQuizQuestion(question)
}
func (s *ContentService) UpdateQuestion(question domain.QuizQuestion) error {
	if !validQuestion(question) {
		return ErrInvalidContent
	}
	if err := s.requireDraft(question.TopicID); err != nil {
		return err
	}
	return s.repository.UpdateQuizQuestion(question)
}
func (s *ContentService) DeleteQuestion(topicID, questionID int) error {
	if err := s.requireDraft(topicID); err != nil {
		return err
	}
	return s.repository.DeleteQuizQuestion(topicID, questionID)
}
func validOption(option domain.QuizOption) bool {
	return option.SortOrder >= 1 && option.SortOrder <= 4 && strings.TrimSpace(option.Text) != ""
}
func (s *ContentService) AddOption(topicID int, option domain.QuizOption) (domain.QuizOption, error) {
	if !validOption(option) {
		return domain.QuizOption{}, ErrInvalidContent
	}
	if err := s.requireDraft(topicID); err != nil {
		return domain.QuizOption{}, err
	}
	if err := s.requireQuestion(topicID, option.QuestionID); err != nil {
		return domain.QuizOption{}, err
	}
	return s.repository.CreateQuizOption(option)
}
func (s *ContentService) UpdateOption(topicID int, option domain.QuizOption) error {
	if !validOption(option) {
		return ErrInvalidContent
	}
	if err := s.requireDraft(topicID); err != nil {
		return err
	}
	if err := s.requireQuestion(topicID, option.QuestionID); err != nil {
		return err
	}
	return s.repository.UpdateQuizOption(option)
}
func (s *ContentService) DeleteOption(topicID, questionID, optionID int) error {
	if err := s.requireDraft(topicID); err != nil {
		return err
	}
	if err := s.requireQuestion(topicID, questionID); err != nil {
		return err
	}
	return s.repository.DeleteQuizOption(questionID, optionID)
}
func (s *ContentService) requireQuestion(topicID, questionID int) error {
	content, err := s.repository.Content(topicID)
	if err != nil {
		return err
	}
	for _, question := range content.Quiz {
		if question.ID == questionID {
			return nil
		}
	}
	return ErrTopicNotFound
}
func (s *ContentService) requireDraft(topicID int) error {
	content, err := s.repository.Content(topicID)
	if err != nil {
		return err
	}
	if content.Topic.Status != domain.TopicStatusDraft {
		return ErrContentConflict
	}
	return nil
}
