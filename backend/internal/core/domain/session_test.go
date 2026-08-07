package domain

import "testing"

func TestCanTransitionAttemptStatus(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
		want bool
	}{
		{name: "keeps current status", from: AttemptStatusInProgress, to: AttemptStatusInProgress, want: true},
		{name: "completes an in-progress attempt", from: AttemptStatusInProgress, to: AttemptStatusCompleted, want: true},
		{name: "abandons an in-progress attempt", from: AttemptStatusInProgress, to: AttemptStatusAbandoned, want: true},
		{name: "does not reopen completed attempt", from: AttemptStatusCompleted, to: AttemptStatusInProgress, want: false},
		{name: "does not change completed attempt", from: AttemptStatusCompleted, to: AttemptStatusAbandoned, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanTransitionAttemptStatus(tt.from, tt.to); got != tt.want {
				t.Fatalf("CanTransitionAttemptStatus(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}
