package domain_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"testing"
)

func TestNormalizedScoreRoundsHalfUp(t *testing.T) {
	if got := domain.NormalizedScore(5, 8); got != 63 {
		t.Fatalf("NormalizedScore(5, 8) = %d, want 63", got)
	}
	if got := domain.NormalizedScore(0, 0); got != 0 {
		t.Fatalf("NormalizedScore(0, 0) = %d, want 0", got)
	}
	if got := domain.NormalizedScore(200, 100); got != 100 {
		t.Fatalf("NormalizedScore(200, 100) = %d, want 100", got)
	}
}
