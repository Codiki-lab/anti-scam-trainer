package domain

import "testing"

func TestStarsFromScore(t *testing.T) {
	tests := []struct {
		score int
		want  int
	}{
		{score: 54, want: 0},
		{score: 55, want: 1},
		{score: 69, want: 1},
		{score: 70, want: 2},
		{score: 84, want: 2},
		{score: 85, want: 3},
	}

	for _, tt := range tests {
		t.Run("score", func(t *testing.T) {
			if got := StarsFromScore(tt.score); got != tt.want {
				t.Fatalf("StarsFromScore(%d) = %d, want %d", tt.score, got, tt.want)
			}
		})
	}
}
