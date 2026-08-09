package domain

import "time"

type Topic struct {
	ID          int
	Slug        string
	UserRole    UserRole
	Title       string
	Description string
	SortOrder   int
	Status      string
	ArchivedAt  time.Time
	TheoryRead  bool
	QuizPassed  bool
	QuizScore   int
	Completed   bool
	Levels      []TopicLevelProgress
}

const (
	TopicStatusDraft     = "draft"
	TopicStatusPublished = "published"
	TopicStatusArchived  = "archived"
)

type TopicContent struct {
	Topic  Topic
	Theory []TheoryBlock
	Quiz   []QuizQuestion
}

type DailyTask struct {
	Date        string
	Role        UserRole
	Messages    []DialogueMessage
	Completed   bool
	CompletedAt *time.Time
	Answer      *bool
	Correct     *bool
	Verdict     bool
	Signals     []string
	SafeAction  string
}

type TheoryBlock struct {
	ID        int
	TopicID   int
	SortOrder int
	Kind      string
	Title     string
	Body      string
}

type QuizQuestion struct {
	ID          int
	TopicID     int
	SortOrder   int
	Text        string
	Explanation string
	Options     []QuizOption
}

type QuizOption struct {
	ID         int
	QuestionID int
	SortOrder  int
	Text       string
	Correct    bool
}

type QuizAnswer struct {
	QuestionID int `json:"question_id"`
	OptionID   int `json:"option_id"`
}

type QuizResult struct {
	Score       int
	Passed      bool
	BestScore   int
	NewlyPassed bool
	Streak      Streak
}

type TopicLevelProgress struct {
	Number        int  `json:"number"`
	Opened        bool `json:"opened"`
	BestScore     int  `json:"best_score"`
	Stars         int  `json:"stars"`
	Attempts      int  `json:"attempts"`
	LastAttemptID int  `json:"last_attempt_id"`
}

type RecentAttempt struct {
	AttemptID int       `json:"attempt_id"`
	TopicID   int       `json:"topic_id"`
	Level     int       `json:"level"`
	Score     int       `json:"score"`
	Stars     int       `json:"stars"`
	Finished  time.Time `json:"finished_at"`
}

type Achievement struct {
	Code        string    `json:"code"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Earned      bool      `json:"earned"`
	EarnedAt    time.Time `json:"earned_at,omitempty"`
	Current     int       `json:"current"`
	Target      int       `json:"target"`
}

type ContinueAction struct {
	Type      string `json:"type"`
	TopicID   int    `json:"topic_id,omitempty"`
	Level     int    `json:"level,omitempty"`
	AttemptID int    `json:"attempt_id,omitempty"`
}

func QuizPassed(score int) bool { return score >= 80 }

func TopicComplete(theoryRead, quizPassed bool, levels []TopicLevelProgress) bool {
	if !theoryRead || !quizPassed || len(levels) != 4 {
		return false
	}
	for _, level := range levels {
		if level.Stars < 1 {
			return false
		}
	}
	return true
}

func NextStreak(current, longest int, lastActivity, activityDate time.Time) (int, int, bool) {
	if lastActivity.Equal(activityDate) {
		return current, longest, false
	}
	next := 1
	if !lastActivity.IsZero() && lastActivity.AddDate(0, 0, 1).Equal(activityDate) {
		next = current + 1
	}
	if next > longest {
		longest = next
	}
	return next, longest, true
}

type AchievementStats struct {
	CompletedAttempts int
	PerfectScore      int
	CompletedTopics   int
	BuyerTopics       int
	SellerTopics      int
	Streak            int
}

func EligibleAchievementCodes(stats AchievementStats) []string {
	result := []string{}
	checks := []struct {
		code string
		ok   bool
	}{
		{"first_training", stats.CompletedAttempts >= 1},
		{"five_trainings", stats.CompletedAttempts >= 5},
		{"perfect_score", stats.PerfectScore >= 100},
		{"first_topic_completed", stats.CompletedTopics >= 1},
		{"all_buyer_topics", stats.BuyerTopics >= 6},
		{"all_seller_topics", stats.SellerTopics >= 6},
		{"streak_3", stats.Streak >= 3},
		{"streak_7", stats.Streak >= 7},
	}
	for _, check := range checks {
		if check.ok {
			result = append(result, check.code)
		}
	}
	return result
}

func AchievementCurrent(code string, stats AchievementStats) int {
	switch code {
	case "perfect_score":
		return stats.PerfectScore
	case "first_topic_completed":
		return stats.CompletedTopics
	case "all_buyer_topics":
		return stats.BuyerTopics
	case "all_seller_topics":
		return stats.SellerTopics
	case "streak_3", "streak_7":
		return stats.Streak
	default:
		return stats.CompletedAttempts
	}
}
