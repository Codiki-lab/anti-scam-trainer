package domain_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"testing"
	"time"
)

func TestQuizRequiresFourOfFiveCorrectAnswers(t *testing.T) {
	for score, want := range map[int]bool{60: false, 79: false, 80: true, 100: true} {
		if got := domain.QuizPassed(score); got != want {
			t.Fatalf("QuizPassed(%d)=%v, want %v", score, got, want)
		}
	}
}

func TestTopicRequiresTheoryQuizAndAllFourLevels(t *testing.T) {
	levels := []domain.TopicLevelProgress{{Number: 1, Stars: 1}, {Number: 2, Stars: 1}, {Number: 3, Stars: 1}, {Number: 4, Stars: 1}}
	if !domain.TopicComplete(true, true, levels) {
		t.Fatal("complete topic was not recognized")
	}
	levels[3].Stars = 0
	if domain.TopicComplete(true, true, levels) {
		t.Fatal("topic completed without passing level 4")
	}
}

func TestStreakUsesDistinctConsecutiveCalendarDates(t *testing.T) {
	location, _ := time.LoadLocation("Europe/Moscow")
	day := func(n int) time.Time { return time.Date(2026, 8, n, 0, 0, 0, 0, location) }
	current, longest, changed := domain.NextStreak(2, 5, day(8), day(9))
	if current != 3 || longest != 5 || !changed {
		t.Fatalf("consecutive=(%d,%d,%v)", current, longest, changed)
	}
	current, longest, changed = domain.NextStreak(current, longest, day(9), day(9))
	if current != 3 || longest != 5 || changed {
		t.Fatalf("same day=(%d,%d,%v)", current, longest, changed)
	}
	current, longest, changed = domain.NextStreak(current, longest, day(9), day(11))
	if current != 1 || longest != 5 || !changed {
		t.Fatalf("gap=(%d,%d,%v)", current, longest, changed)
	}
}

func TestAllEightAchievementConditionsAreDeterministic(t *testing.T){codes:=domain.EligibleAchievementCodes(domain.AchievementStats{CompletedAttempts:5,PerfectScore:100,CompletedTopics:1,BuyerTopics:6,SellerTopics:6,Streak:7});if len(codes)!=8{t.Fatalf("codes=%v, want all eight",codes)};none:=domain.EligibleAchievementCodes(domain.AchievementStats{});if len(none)!=0{t.Fatalf("empty stats awarded %v",none)}}
