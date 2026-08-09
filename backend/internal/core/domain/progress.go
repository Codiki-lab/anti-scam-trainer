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
	TopicID   int
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

func NormalizedScore(rawScore, maximumScore int) int {
	if rawScore <= 0 || maximumScore <= 0 {
		return 0
	}
	score := (rawScore*100 + maximumScore/2) / maximumScore
	if score > 100 {
		return 100
	}
	return score
}
