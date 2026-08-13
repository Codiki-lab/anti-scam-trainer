package http_contract_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	learningservice "anti-scam-trainer/backend/internal/features/learning/service"
	"time"
)

type learningStore struct {
	theoryRead    bool
	activityCalls int
	attemptID     int
	lastRole      domain.UserRole
	daily         map[string]domain.DailyTask
	topics        []domain.Topic
	stablePattern string
	skillCheck    domain.SkillCheck
}

type chatRecommendationStore struct {
	learningStore
	topicsByRole map[domain.UserRole][]domain.Topic
	topicCalls   int
}

func (s *chatRecommendationStore) Topics(_ int, role domain.UserRole) ([]domain.Topic, error) {
	s.topicCalls++
	return s.topicsByRole[role], nil
}

type stableLearningStore struct {
	*learningStore
	recommendations map[string]domain.ContinueAction
}

func (s *stableLearningStore) FindRecommendation(_ int, date time.Time, role domain.UserRole) (domain.ContinueAction, bool, error) {
	action, ok := s.recommendations[date.Format("2006-01-02")+":"+string(role)]
	return action, ok, nil
}

func (s *stableLearningStore) SaveRecommendation(_ int, date time.Time, role domain.UserRole, action domain.ContinueAction) error {
	s.recommendations[date.Format("2006-01-02")+":"+string(role)] = action
	return nil
}

func (s *learningStore) FindDailyTask(_ int, date time.Time) (domain.DailyTask, bool, error) {
	task, ok := s.daily[date.Format("2006-01-02")]
	return task, ok, nil
}

func (s *learningStore) DailyTask(_ int, date time.Time, created domain.DailyTask) (domain.DailyTask, error) {
	if s.daily == nil {
		s.daily = map[string]domain.DailyTask{}
	}
	key := date.Format("2006-01-02")
	if task, ok := s.daily[key]; ok {
		return task, nil
	}
	s.daily[key] = created
	return created, nil
}

func (s *learningStore) AnswerDailyTask(_ int, date time.Time, answer bool) (domain.DailyTask, domain.Streak, error) {
	key := date.Format("2006-01-02")
	task, ok := s.daily[key]
	if !ok {
		return domain.DailyTask{}, domain.Streak{}, learningservice.ErrDailyTaskUnavailable
	}
	if task.Completed {
		return domain.DailyTask{}, domain.Streak{}, learningservice.ErrDailyTaskAnswered
	}
	correct := answer == task.Verdict
	task.Answer, task.Correct, task.Completed = &answer, &correct, true
	now := time.Now()
	task.CompletedAt = &now
	s.daily[key] = task
	return task, domain.Streak{Current: 1, Longest: 1, ActiveToday: true}, nil
}

func (s *learningStore) Topics(_ int, role domain.UserRole) ([]domain.Topic, error) {
	s.lastRole = role
	if s.topics != nil {
		return s.topics, nil
	}
	return []domain.Topic{{ID: 1, Slug: string(role) + "-topic", UserRole: role, Title: "Тема", Description: "Описание", SortOrder: 1, TheoryRead: s.theoryRead, Levels: []domain.TopicLevelProgress{{Number: 1, Opened: true}, {Number: 2}, {Number: 3}, {Number: 4}}}}, nil
}

func (s *learningStore) Topic(_ int, topicID int) (domain.Topic, error) {
	return domain.Topic{ID: topicID, UserRole: domain.UserRoleBuyer, TheoryRead: s.theoryRead}, nil
}

func (s *learningStore) Theory(topicID int) ([]domain.TheoryBlock, error) {
	return []domain.TheoryBlock{{ID: 1, TopicID: topicID, SortOrder: 1, Kind: "intro", Title: "Введение", Body: "Текст"}}, nil
}

func (s *learningStore) MarkTheoryRead(_ int, _ int, _ time.Time) (domain.Streak, bool, error) {
	s.activityCalls++
	newlyRead := !s.theoryRead
	s.theoryRead = true
	return domain.Streak{Current: 1, Longest: 1, ActiveToday: true, LastActivityDate: "2026-08-09"}, newlyRead, nil
}

func (s *learningStore) Quiz(int) ([]domain.QuizQuestion, error) {
	questions := make([]domain.QuizQuestion, 5)
	for i := range questions {
		questions[i] = domain.QuizQuestion{ID: i + 1, SortOrder: i + 1, Text: "Вопрос", Explanation: "Скрыто", Options: []domain.QuizOption{{ID: i*10 + 1, Text: "Вариант", Correct: true}, {ID: i*10 + 2, Text: "Вариант"}, {ID: i*10 + 3, Text: "Вариант"}, {ID: i*10 + 4, Text: "Вариант"}}}
	}
	return questions, nil
}

func (s *learningStore) SubmitQuiz(_ int, _ int, _ []domain.QuizAnswer, _ time.Time) (domain.QuizResult, error) {
	return domain.QuizResult{Score: 80, Passed: true, BestScore: 80, NewlyPassed: true, Streak: domain.Streak{Current: 1, Longest: 1, ActiveToday: true}}, nil
}

func (s *learningStore) RecentAttempts(int, domain.UserRole) ([]domain.RecentAttempt, float64, error) {
	return []domain.RecentAttempt{}, 0, nil
}

func (s *learningStore) Achievements(int) ([]domain.Achievement, error) {
	return []domain.Achievement{}, nil
}

func (s *learningStore) User(int) (domain.User, error) {
	return domain.User{ID: 7, Username: "alex", TrainingRole: domain.UserRoleBuyer}, nil
}

func (s *learningStore) InProgressAttempt(_ int, role domain.UserRole) (int, int, int, error) {
	s.lastRole = role
	return s.attemptID, 1, 2, nil
}

func (s *learningStore) MistakePatternStats(int, domain.UserRole) ([]domain.MistakePatternStats, error) {
	if s.stablePattern == "" {
		return nil, nil
	}
	return []domain.MistakePatternStats{{PatternCode: s.stablePattern, UnsafeCount: 3, RecentUnsafe: 2}}, nil
}

func (s *learningStore) StartSkillCheck(_ int, topicID int) (domain.SkillCheck, error) {
	if s.skillCheck.ID == 0 {
		s.skillCheck = domain.SkillCheck{ID: 9, TopicID: topicID, Before: domain.DialogueSnapshot{Messages: []domain.DialogueMessage{{Role: domain.MessageRoleAssistant, Text: "Откройте форму оплаты"}}, IsScam: true, PatternCode: "external_link"}, After: domain.DialogueSnapshot{Messages: []domain.DialogueMessage{{Role: domain.MessageRoleAssistant, Text: "Отправьте код возврата"}}, IsScam: true, PatternCode: "external_link"}}
	}
	return s.skillCheck, nil
}

func (s *learningStore) SkillCheck(_ int, _ int) (domain.SkillCheck, error) { return s.skillCheck, nil }

func (s *learningStore) AnswerSkillCheck(_ int, _ int, answer bool) (domain.SkillCheck, error) {
	switch {
	case s.skillCheck.BeforeAnswer == nil:
		s.skillCheck.BeforeAnswer = &answer
	case s.skillCheck.TopicComplete && s.skillCheck.AfterAnswer == nil:
		s.skillCheck.AfterAnswer = &answer
	default:
		return domain.SkillCheck{}, learningservice.ErrInvalidQuiz
	}
	return s.skillCheck, nil
}
