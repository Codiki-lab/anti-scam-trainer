package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const chatRecommendationSource = "avito_chat_demo"

var (
	urlInSnapshot        = regexp.MustCompile(`(?i)\b(?:https?://|www\.|(?:[a-z0-9-]+\.)+[a-z]{2,})(?:[^\s]*)`)
	phoneInSnapshot      = regexp.MustCompile(`(?:\+?\d[\d\s()\-]{8,}\d)`)
	cardInSnapshot       = regexp.MustCompile(`(?:\d[\s-]*){12,19}`)
	confirmationCodeText = regexp.MustCompile(`(?i)(?:код|code|парол)[^\d]{0,20}\d{4,8}`)
	chatIdentifierText   = regexp.MustCompile(`(?i)(?:chat|чат|объявлен|listing|ad)[_\s-]*(?:id|№|number)?\s*[:#№-]?\s*\d{3,}`)
)

var (
	ErrTopicNotFound        = apperrors.ErrScenarioNotFound
	ErrInvalidQuiz          = errors.New("invalid quiz submission")
	ErrDailyTaskUnavailable = errors.New("no valid daily task is available")
	ErrDailyTaskAnswered    = errors.New("daily task is already answered")
	ErrInvalidDailyAnswer   = errors.New("invalid daily task answer")
	ErrInvalidChatReferral  = errors.New("invalid chat recommendation request")
)

type ChatRecommendationCommand struct {
	Source      string
	Role        domain.UserRole
	Messages    []domain.DialogueMessage
	RiskType    string
	RiskSignals []string
}

func ValidateChatRecommendation(command ChatRecommendationCommand) error {
	if command.Source != chatRecommendationSource || !domain.ValidUserRole(command.Role) || len(command.Messages) < 2 || len(command.Messages) > 6 || len(command.RiskSignals) > 3 || strings.TrimSpace(command.RiskType) == "" || len([]rune(command.RiskType)) > 80 {
		return ErrInvalidChatReferral
	}
	for _, message := range command.Messages {
		if (message.Role != domain.MessageRoleUser && message.Role != domain.MessageRoleAssistant) || strings.TrimSpace(message.Text) == "" || len([]rune(message.Text)) > 400 || containsSensitiveSnapshotData(message.Text) {
			return ErrInvalidChatReferral
		}
	}
	for _, signal := range command.RiskSignals {
		if strings.TrimSpace(signal) == "" || len([]rune(signal)) > 180 || containsSensitiveSnapshotData(signal) {
			return ErrInvalidChatReferral
		}
	}
	return nil
}

func containsSensitiveSnapshotData(value string) bool {
	return urlInSnapshot.MatchString(value) || phoneInSnapshot.MatchString(value) || cardInSnapshot.MatchString(value) || confirmationCodeText.MatchString(value) || chatIdentifierText.MatchString(value)
}

type DailyTaskGenerator interface {
	GenerateDailyTask(context.Context, DailyTaskProfile, domain.UserRole) (domain.DailyTask, error)
}
type DailyTaskProfile struct {
	UserID         int
	PreferredRole  domain.UserRole
	Topics         []domain.Topic
	RecentAttempts []domain.RecentAttempt
}

type Service struct {
	repository Repository
	now        func() time.Time
	generator  DailyTaskGenerator
	selectRole func() domain.UserRole
}

func New(repository Repository) *Service {
	return &Service{repository: repository, now: time.Now, selectRole: randomDailyRole}
}
func NewWithClock(repository Repository, now func() time.Time) *Service {
	return &Service{repository: repository, now: now, selectRole: randomDailyRole}
}
func NewWithDailyTaskGenerator(repository Repository, generator DailyTaskGenerator) *Service {
	s := New(repository)
	s.generator = generator
	return s
}
func (s *Service) WithDailyRoleSelector(selectRole func() domain.UserRole) *Service {
	s.selectRole = selectRole
	return s
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

func (s *Service) StartSkillCheck(userID, topicID int) (domain.SkillCheck, error) {
	if _, err := s.repository.Topic(userID, topicID); err != nil {
		return domain.SkillCheck{}, err
	}
	repository, ok := s.repository.(SkillCheckRepository)
	if !ok {
		return domain.SkillCheck{}, ErrTopicNotFound
	}
	return repository.StartSkillCheck(userID, topicID)
}

func (s *Service) SkillCheck(userID, checkID int) (domain.SkillCheck, error) {
	repository, ok := s.repository.(SkillCheckRepository)
	if !ok {
		return domain.SkillCheck{}, ErrTopicNotFound
	}
	return repository.SkillCheck(userID, checkID)
}

func (s *Service) AnswerSkillCheck(userID, checkID int, answer bool) (domain.SkillCheck, error) {
	repository, ok := s.repository.(SkillCheckRepository)
	if !ok {
		return domain.SkillCheck{}, ErrTopicNotFound
	}
	return repository.AnswerSkillCheck(userID, checkID, answer)
}

// PersonalRecommendation only selects content already available in the normal
// learning path. The Profile is a hint, never an unlock mechanism.
func (s *Service) PersonalRecommendation(userID int, role domain.UserRole) (domain.ChatRecommendation, error) {
	profile, ok := s.repository.(MistakeProfileRepository)
	if !ok {
		return domain.ChatRecommendation{}, ErrTopicNotFound
	}
	stats, err := profile.MistakePatternStats(userID, role)
	if err != nil {
		return domain.ChatRecommendation{}, err
	}
	pattern, stable := "", false
	for _, item := range stats {
		if item.Stable() {
			pattern, stable = item.PatternCode, true
			break
		}
	}
	topics, err := s.Topics(userID, role)
	if err != nil {
		return domain.ChatRecommendation{}, err
	}
	slug := generalSafetyTopicSlug(role)
	if stable {
		if mapped, exact := chatRecommendationSlug(role, pattern); exact {
			slug = mapped
		}
	}
	for _, topic := range topics {
		if topic.Slug == slug {
			return domain.ChatRecommendation{Topic: topic, Explanation: "Следующее действие выбрано по повторяющемуся паттерну и не меняет доступность контента.", NextAction: nextTopicAction(topic), IsFallback: !stable || topic.Slug != slug}, nil
		}
	}
	if len(topics) == 0 {
		return domain.ChatRecommendation{}, ErrTopicNotFound
	}
	return domain.ChatRecommendation{Topic: topics[0], Explanation: "Продолжайте доступную последовательность обучения.", NextAction: nextTopicAction(topics[0]), IsFallback: true}, nil
}

func (s *Service) RecommendFromChat(userID int, command ChatRecommendationCommand) (domain.ChatRecommendation, error) {
	if err := ValidateChatRecommendation(command); err != nil {
		return domain.ChatRecommendation{}, err
	}
	topics, err := s.Topics(userID, command.Role)
	if err != nil {
		return domain.ChatRecommendation{}, err
	}
	slug, exact := chatRecommendationSlug(command.Role, command.RiskType)
	if !exact {
		slug = generalSafetyTopicSlug(command.Role)
	}
	for _, topic := range topics {
		if topic.Slug == slug && topic.UserRole == command.Role && topic.Status == domain.TopicStatusPublished {
			return s.chatRecommendationForTopic(userID, command.Role, topic, command.RiskType, exact), nil
		}
	}
	if exact {
		fallbackSlug := generalSafetyTopicSlug(command.Role)
		for _, topic := range topics {
			if topic.Slug == fallbackSlug && topic.UserRole == command.Role && topic.Status == domain.TopicStatusPublished {
				return s.chatRecommendationForTopic(userID, command.Role, topic, command.RiskType, false), nil
			}
		}
	}
	return domain.ChatRecommendation{}, fmt.Errorf("%w: mapped topic is unavailable", ErrTopicNotFound)
}

func (s *Service) chatRecommendationForTopic(userID int, role domain.UserRole, topic domain.Topic, riskType string, exact bool) domain.ChatRecommendation {
	action := nextTopicAction(topic)
	if attemptID, topicID, _, err := s.repository.InProgressAttempt(userID, role); err == nil && attemptID != 0 && topicID == topic.ID {
		action = domain.ContinueAction{Type: "resume_attempt", TopicID: topic.ID, AttemptID: attemptID}
	}
	return domain.ChatRecommendation{Topic: topic, Explanation: chatRecommendationExplanation(riskType, exact), NextAction: action, IsFallback: !exact}
}

func chatRecommendationSlug(role domain.UserRole, riskType string) (string, bool) {
	if canonical, ok := map[string]string{"external_link": "phishing", "fake_delivery": "delivery", "sms_code": "account_takeover", "pressure": "social_engineering"}[riskType]; ok {
		riskType = canonical
	}
	mapping := map[domain.UserRole]map[string]string{
		domain.UserRoleBuyer:  {"phishing": "buyer-phishing-links", "prepayment": "buyer-prepayment", "delivery": "buyer-fake-delivery", "external_messenger": "buyer-off-platform", "account_takeover": "buyer-sms-codes", "fake_payment": "buyer-phishing-links", "social_engineering": "buyer-too-good-offer"},
		domain.UserRoleSeller: {"phishing": "seller-external-links", "prepayment": "seller-fake-payment", "delivery": "seller-fake-delivery", "external_messenger": "seller-off-platform", "account_takeover": "seller-confirmation-codes", "fake_payment": "seller-fake-payment", "social_engineering": "seller-pressure"},
	}
	slug, ok := mapping[role][riskType]
	return slug, ok
}

func generalSafetyTopicSlug(role domain.UserRole) string {
	if role == domain.UserRoleSeller {
		return "seller-pressure"
	}
	return "buyer-too-good-offer"
}

func chatRecommendationExplanation(riskType string, exact bool) string {
	if !exact {
		return "Для этого сигнала нет точного учебного соответствия. Начните с общей Темы о безопасном порядке сделки."
	}
	explanations := map[string]string{
		"phishing":           "Внешняя форма или ссылка может подменить штатное оформление. Разберите, как проверить сделку внутри сервиса.",
		"prepayment":         "Предоплата до штатного оформления не подтверждает сделку. Разберите безопасный порядок проверки и оплаты.",
		"delivery":           "Неожиданные платежи и условия доставки нужно проверять только внутри сервиса. Эта Тема поможет распознать подмену.",
		"external_messenger": "Перенос общения во внешний мессенджер лишает сделку проверяемого контекста. Разберите безопасные действия в этой Теме.",
		"account_takeover":   "Код подтверждения предназначен только владельцу аккаунта. Эта Тема объяснит, как защитить доступ к нему.",
		"fake_payment":       "Чек, сообщение или форма собеседника не подтверждают оплату. Разберите, как проверить её самостоятельно.",
		"social_engineering": "Срочность и давление не меняют безопасный порядок сделки. Эта Тема поможет вовремя взять паузу и проверить условия.",
	}
	return explanations[riskType]
}

func nextTopicAction(topic domain.Topic) domain.ContinueAction {
	if !topic.TheoryRead {
		return domain.ContinueAction{Type: "read_theory", TopicID: topic.ID}
	}
	if !topic.QuizPassed {
		return domain.ContinueAction{Type: "take_quiz", TopicID: topic.ID}
	}
	for _, level := range topic.Levels {
		if level.Opened && level.Stars == 0 {
			return domain.ContinueAction{Type: "start_level", TopicID: topic.ID, Level: level.Number}
		}
	}
	return domain.ContinueAction{Type: "read_theory", TopicID: topic.ID}
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
	action := s.stableContinueAction(userID, role, topics)
	profileTopics := append([]domain.Topic{}, topics...)
	otherRole := domain.UserRoleBuyer
	if role == domain.UserRoleBuyer {
		otherRole = domain.UserRoleSeller
	}
	if otherTopics, otherErr := s.Topics(userID, otherRole); otherErr == nil {
		profileTopics = append(profileTopics, otherTopics...)
	}
	var task *domain.DailyTask
	if daily, ok := s.repository.(DailyTaskRepository); ok {
		value, exists, taskErr := daily.FindDailyTask(userID, s.activityDate())
		if taskErr != nil {
			return domain.User{}, nil, nil, nil, nil, taskErr
		}
		if !exists {
			value, taskErr = daily.DailyTask(userID, s.activityDate(), s.newDailyTask(user, profileTopics))
		}
		if taskErr != nil {
			return domain.User{}, nil, nil, nil, nil, taskErr
		}
		task = &value
	} else {
		value := s.newDailyTask(user, profileTopics)
		task = &value
	}
	return user, topics, achievements, action, task, nil
}

func (s *Service) stableContinueAction(userID int, role domain.UserRole, topics []domain.Topic) *domain.ContinueAction {
	recommendations, ok := s.repository.(RecommendationRepository)
	if !ok {
		return s.continueAction(userID, role, topics)
	}
	date := s.activityDate()
	if saved, exists, err := recommendations.FindRecommendation(userID, date, role); err == nil && exists {
		return &saved
	}
	action := s.continueAction(userID, role, topics)
	if action != nil {
		_ = recommendations.SaveRecommendation(userID, date, role, *action)
	}
	return action
}
func (s *Service) AnswerDailyTask(userID int, answer *bool) (domain.DailyTask, domain.Streak, error) {
	if answer == nil {
		return domain.DailyTask{}, domain.Streak{}, ErrInvalidDailyAnswer
	}
	daily, ok := s.repository.(DailyTaskRepository)
	if !ok {
		return domain.DailyTask{}, domain.Streak{}, ErrDailyTaskUnavailable
	}
	return daily.AnswerDailyTask(userID, s.activityDate(), *answer)
}
func (s *Service) newDailyTask(user domain.User, topics []domain.Topic) domain.DailyTask {
	role := s.selectRole()
	if !domain.ValidUserRole(role) {
		role = domain.UserRoleBuyer
	}
	fallback := fallbackDailyTask(s.activityDate(), role)
	if s.generator == nil {
		return fallback
	}
	recent, _, err := s.repository.RecentAttempts(user.ID, user.TrainingRole)
	if err != nil {
		return fallback
	}
	generated, err := s.generator.GenerateDailyTask(context.Background(), DailyTaskProfile{UserID: user.ID, PreferredRole: user.TrainingRole, Topics: topics, RecentAttempts: recent}, role)
	generated.Role = role
	if err != nil || !validDailyTask(generated) {
		return fallback
	}
	generated.Date = s.activityDate().Format("2006-01-02")
	return generated
}
func randomDailyRole() domain.UserRole {
	var b [1]byte
	if _, err := rand.Read(b[:]); err != nil || b[0]%2 == 0 {
		return domain.UserRoleBuyer
	}
	return domain.UserRoleSeller
}
func fallbackDailyTask(date time.Time, role domain.UserRole) domain.DailyTask {
	return domain.DailyTask{Date: date.Format("2006-01-02"), Role: role, Messages: []domain.DialogueMessage{{Role: "assistant", Text: "Покупатель просит продолжить общение и оплату только внутри сервиса объявлений."}, {Role: "user", Text: "Давайте так и сделаем, без ссылок и переводов вне сервиса."}}, Verdict: false, Signals: []string{"Нет просьбы перейти по внешней ссылке", "Оплата остаётся внутри сервиса"}, SafeAction: "Продолжайте использовать защищённые инструменты сервиса и не сообщайте коды."}
}
func validDailyTask(task domain.DailyTask) bool {
	if !domain.ValidUserRole(task.Role) || len(task.Messages) < 2 || len(task.Messages) > 6 || len(task.Signals) > 3 || strings.TrimSpace(task.SafeAction) == "" || len([]rune(task.SafeAction)) > 300 {
		return false
	}
	for _, m := range task.Messages {
		if (m.Role != "user" && m.Role != "assistant") || strings.TrimSpace(m.Text) == "" || len([]rune(m.Text)) > 400 {
			return false
		}
	}
	for _, signal := range task.Signals {
		if strings.TrimSpace(signal) == "" || len([]rune(signal)) > 180 {
			return false
		}
	}
	return true
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
