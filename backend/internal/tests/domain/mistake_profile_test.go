package domain_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"testing"
)

func TestMistakePatternStabilityAndSafeAnswerDecay(t *testing.T) {
	tests := []struct {
		name     string
		stats    domain.MistakePatternStats
		stable   bool
		priority int
	}{
		{"one mistake is noise", domain.MistakePatternStats{UnsafeCount: 1, RecentUnsafe: 1}, false, 1},
		{"two among latest five are stable", domain.MistakePatternStats{UnsafeCount: 2, RecentUnsafe: 2}, true, 2},
		{"three all time are stable", domain.MistakePatternStats{UnsafeCount: 3, RecentUnsafe: 1}, true, 3},
		{"two safe answers lower and fade a two-mistake pattern", domain.MistakePatternStats{UnsafeCount: 2, SafeCount: 2, RecentUnsafe: 2}, false, 1},
		{"four safe answers fade a three-mistake pattern", domain.MistakePatternStats{UnsafeCount: 3, SafeCount: 4, RecentUnsafe: 1}, false, 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.stats.Stable() != test.stable || test.stats.Priority() != test.priority {
				t.Fatalf("stats=%#v stable=%v priority=%d", test.stats, test.stats.Stable(), test.stats.Priority())
			}
		})
	}
}

func TestMicroQuestionIsSpecificToRecurringPattern(t *testing.T) {
	phishing := domain.MicroQuestionFor("phishing")
	code := domain.MicroQuestionFor("sms_code")
	prepayment := domain.MicroQuestionFor("prepayment")
	if phishing.Question == code.Question || code.Question == prepayment.Question || phishing.Question == prepayment.Question {
		t.Fatalf("pattern questions must differ: phishing=%q code=%q prepayment=%q", phishing.Question, code.Question, prepayment.Question)
	}
	for _, question := range []*domain.MicroQuestion{phishing, code, prepayment} {
		if question.PatternCode == "" || len(question.Options) != 2 || question.Correct != 0 {
			t.Fatalf("invalid micro-question %#v", question)
		}
	}
}

func TestSkillCheckOutcomeComparesVerdictAndRecurringPattern(t *testing.T) {
	before, after := false, true
	check := domain.SkillCheck{Before: domain.DialogueSnapshot{IsScam: true, PatternCode: "external_link"}, After: domain.DialogueSnapshot{IsScam: true, PatternCode: "external_link"}, BeforeAnswer: &before, AfterAnswer: &after}
	outcome, completed := check.Outcome()
	if !completed || outcome.BeforeCorrect || !outcome.AfterCorrect || !outcome.VerdictImproved || !outcome.PatternImproved || !outcome.Improved {
		t.Fatalf("outcome=%#v completed=%v", outcome, completed)
	}
	check.After.PatternCode = "account_takeover"
	outcome, _ = check.Outcome()
	if outcome.PatternImproved || outcome.Improved {
		t.Fatalf("different patterns must not count as recurring improvement: %#v", outcome)
	}
}
