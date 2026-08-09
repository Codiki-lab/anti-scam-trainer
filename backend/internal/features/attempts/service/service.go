package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"anti-scam-trainer/backend/internal/core/ratelimit"
	cryptorand "crypto/rand"
	"fmt"
	"time"
)

type GameService struct {
	repository      GameRepository
	ai              AIProvider
	selectScam      func() bool
	freeTextLimiter *ratelimit.Limiter
	freePlayLimiter *ratelimit.Limiter
	aiGate          *ratelimit.Gate
}

func NewGameWithRateLimits(repository GameRepository, ai AIProvider, freeText, freePlay *ratelimit.Limiter, gate *ratelimit.Gate) *GameService {
	return &GameService{repository: repository, ai: ai, selectScam: randomScam, freeTextLimiter: freeText, freePlayLimiter: freePlay, aiGate: gate}
}

type RateLimitError struct{ RetryAfter time.Duration }

func (e *RateLimitError) Error() string { return fmt.Sprintf("rate limited for %s", e.RetryAfter) }
func (s *GameService) beforeAI(userID int, freePlay bool) (func(), error) {
	key := fmt.Sprintf("user:%d", userID)
	limiter := s.freeTextLimiter
	if freePlay {
		limiter = s.freePlayLimiter
	}
	release := func() {}
	if s.aiGate != nil {
		var ok bool
		release, ok = s.aiGate.TryEnter(key)
		if !ok {
			return nil, &RateLimitError{RetryAfter: time.Second}
		}
	}
	if limiter != nil {
		if ok, retry := limiter.Allow(key); !ok {
			release()
			return nil, &RateLimitError{RetryAfter: retry}
		}
	}
	return release, nil
}

func NewGame(repository GameRepository) *GameService { return &GameService{repository: repository} }
func NewGameWithAI(repository GameRepository, ai AIProvider) *GameService {
	return &GameService{repository: repository, ai: ai, selectScam: randomScam}
}

func NewGameWithDependencies(repository GameRepository, ai AIProvider, selectScam func() bool) *GameService {
	return &GameService{repository: repository, ai: ai, selectScam: selectScam}
}

func randomScam() bool {
	var value [1]byte
	if _, err := cryptorand.Read(value[:]); err != nil {
		return true
	}
	return value[0]%2 == 0
}

type OpenLevel struct {
	Level      domain.Level
	Opened     bool
	ScenarioID int
}

type GameState struct {
	Attempt        domain.Attempt
	Scenario       domain.Scenario
	Step           domain.ScenarioStep
	Answers        []domain.UserAnswer
	Messages       []domain.DialogueMessage
	CanFinishEarly bool
}

type Completion struct {
	Attempt   domain.Attempt
	Stars     int
	Answers   []domain.UserAnswer
	Breakdown []AnswerBreakdown
	Result    domain.AttemptResult
}

type AnswerBreakdown = domain.AnswerBreakdown

type AnswerCommand struct {
	StepID   *int
	OptionID *int
	FreeText *string
	Finish   bool
}

func (s *GameService) completionError(attempt domain.Attempt, err error) error {
	refreshed, refreshErr := s.repository.GetGameAttempt(attempt.ID)
	if refreshErr == nil && (refreshed.Status != domain.AttemptStatusInProgress || refreshed.CurrentStepNumber != attempt.CurrentStepNumber || refreshed.FreeTextCount != attempt.FreeTextCount) {
		return apperrors.ErrStaleStep
	}
	return err
}

func (s *GameService) Levels(userID int, role string, topicID ...int) ([]OpenLevel, error) {
	levels, progress, err := s.repository.Levels(userID, role)
	quizPassed := true
	if len(topicID) > 0 {
		if topical, ok := s.repository.(TopicGameRepository); ok {
			levels, progress, quizPassed, err = topical.TopicLevels(userID, role, topicID[0])
		}
	}
	if err != nil {
		return nil, err
	}
	stars := map[int]int{}
	for _, item := range progress {
		stars[item.LevelID] = item.Stars
	}
	result := make([]OpenLevel, 0, len(levels))
	for _, level := range levels {
		opened := level.Number == 1
		if level.Number > 1 {
			for _, previous := range levels {
				if previous.Number == level.Number-1 {
					opened = stars[previous.ID] > 0
					break
				}
			}
		}
		if level.Number == 1 {
			opened = quizPassed
		}
		var scenario domain.Scenario
		var scenarioErr error
		if len(topicID) > 0 {
			if topical, ok := s.repository.(TopicGameRepository); ok {
				scenario, scenarioErr = topical.PublishedTopicScenario(level.Number, role, topicID[0])
			} else {
				scenario, scenarioErr = s.repository.PublishedScenario(level.Number, role)
			}
		} else {
			scenario, scenarioErr = s.repository.PublishedScenario(level.Number, role)
		}
		if scenarioErr != nil {
			continue
		}
		result = append(result, OpenLevel{Level: level, Opened: opened, ScenarioID: scenario.ID})
	}
	return result, nil
}
