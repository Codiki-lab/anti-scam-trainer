package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"errors"
	"strings"
	"time"
)

var (
	ErrTopicNotFound        = apperrors.ErrScenarioNotFound
	ErrInvalidQuiz          = errors.New("invalid quiz submission")
	ErrDailyTaskUnavailable = errors.New("no valid daily task is available")
)

type Service struct {
	repository Repository
	now        func() time.Time
}

func New(repository Repository) *Service { return &Service{repository: repository, now: time.Now} }
func NewWithClock(repository Repository, now func() time.Time) *Service {
	return &Service{repository: repository, now: now}
}

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
func (s *Service) Dashboard(userID int, role domain.UserRole) (domain.User, []domain.Topic, []domain.Achievement, *domain.ContinueAction, *domain.DailyTask, error) {
	user, err := s.repository.User(userID)
	if err != nil {
		return domain.User{}, nil, nil, nil, nil, err
	}
	topics, err := s.Topics(userID, role)
	if err != nil {
		return domain.User{}, nil, nil, nil, nil, err
	}
	achievements, err := s.repository.Achievements(userID)
	if err != nil {
		return domain.User{}, nil, nil, nil, nil, err
	}
	action := s.continueAction(userID, role, topics)
	if action == nil {
		return domain.User{}, nil, nil, nil, nil, ErrDailyTaskUnavailable
	}
	var task *domain.DailyTask
	if daily, ok := s.repository.(DailyTaskRepository); ok {
		value, taskErr := daily.DailyTask(userID, role, s.activityDate(), *action)
		if taskErr != nil {
			return domain.User{}, nil, nil, nil, nil, taskErr
		}
		task = &value
	} else {
		value := domain.DailyTask{Date: s.activityDate().Format("2006-01-02"), Role: role, Action: *action}
		task = &value
	}
	return user, topics, achievements, action, task, nil
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
	if len(topics) == 6 {
		allComplete := true
		for _, topic := range topics {
			if !topic.Completed {
				allComplete = false
				break
			}
		}
		if allComplete {
			return &domain.ContinueAction{Type: "start_free_play"}
		}
	}
	return nil
}
func (s *Service) activityDate() time.Time {
	location, _ := time.LoadLocation("Europe/Moscow")
	now := s.now().In(location)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
}

var (
	ErrContentConflict = errors.New("content conflict")
	ErrInvalidContent  = errors.New("invalid content")
)

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
