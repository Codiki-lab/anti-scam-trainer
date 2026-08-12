package domain_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"testing"
)

func TestCanTransitionAttemptStatus(t *testing.T) {
	tests := []struct {
		name string
		from domain.AttemptStatus
		to   domain.AttemptStatus
		want bool
	}{
		{name: "keeps current status", from: domain.AttemptStatusInProgress, to: domain.AttemptStatusInProgress, want: true},
		{name: "completes an in-progress attempt", from: domain.AttemptStatusInProgress, to: domain.AttemptStatusCompleted, want: true},
		{name: "abandons an in-progress attempt", from: domain.AttemptStatusInProgress, to: domain.AttemptStatusAbandoned, want: true},
		{name: "does not reopen completed attempt", from: domain.AttemptStatusCompleted, to: domain.AttemptStatusInProgress, want: false},
		{name: "does not change completed attempt", from: domain.AttemptStatusCompleted, to: domain.AttemptStatusAbandoned, want: false},
		{name: "does not accept an unknown status", from: "UNKNOWN", to: "UNKNOWN", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := domain.CanTransitionAttemptStatus(tt.from, tt.to); got != tt.want {
				t.Fatalf("CanTransitionAttemptStatus(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestAttemptStatusTransitionRejectsUnknownCurrentStatus(t *testing.T) {
	attempt := domain.Attempt{Status: "UNKNOWN"}
	if attempt.TransitionTo(domain.AttemptStatusCompleted) {
		t.Fatal("unknown attempt status transitioned to completed")
	}
}
