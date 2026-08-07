package domain

import "time"

type Level struct {
	ID     int
	Number int
}

type Progress struct {
	UserID    int
	LevelID   int
	UserRole  string
	BestScore int
	Stars     int
	Attempts  int
	PassedAt  time.Time
}

func StarsFromScore(score int) int {
	switch {
	case score >= 85:
		return 3
	case score >= 70:
		return 2
	case score >= 55:
		return 1
	default:
		return 0
	}
}
