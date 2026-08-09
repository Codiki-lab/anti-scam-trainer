package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"time"
)

type GameService struct{ repository GameRepository }

func NewGame(repository GameRepository) *GameService { return &GameService{repository: repository} }

type OpenLevel struct {
	Level      domain.Level
	Opened     bool
	ScenarioID int
}

type GameState struct {
	Attempt  domain.Attempt
	Step     domain.ScenarioStep
	Answers  []domain.UserAnswer
	Messages []domain.Message
}

type Completion struct {
	Attempt   domain.Attempt
	Stars     int
	Answers   []domain.UserAnswer
	Breakdown []AnswerBreakdown
}

type AnswerBreakdown struct {
	StepID, OptionID, Points int
	Explanation              string
	OptionText               string
}

func (s *GameService) Levels(userID int, role string) ([]OpenLevel, error) {
	levels, progress, err := s.repository.Levels(userID, role)
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
		scenario, scenarioErr := s.repository.PublishedScenario(level.Number, role)
		if scenarioErr != nil {
			continue
		}
		result = append(result, OpenLevel{Level: level, Opened: opened, ScenarioID: scenario.ID})
	}
	return result, nil
}

func (s *GameService) Start(userID, levelNumber int, role string) (GameState, error) {
	levels, err := s.Levels(userID, role)
	if err != nil {
		return GameState{}, err
	}
	var target OpenLevel
	found := false
	for _, level := range levels {
		if level.Level.Number == levelNumber {
			target, found = level, true
			break
		}
	}
	if !found {
		return GameState{}, apperrors.ErrScenarioNotFound
	}
	if !target.Opened {
		return GameState{}, apperrors.ErrForbidden
	}
	if attempt, err := s.repository.FindInProgress(userID, target.ScenarioID); err == nil {
		step, stepErr := s.repository.Step(attempt.ScenarioID, attempt.CurrentStepNumber)
		if stepErr != nil {
			return GameState{}, stepErr
		}
		answers, answersErr := s.repository.Answers(attempt.ID)
		if answersErr != nil {
			return GameState{}, answersErr
		}
		messages, messagesErr := s.repository.Messages(attempt.ID)
		if messagesErr != nil {
			return GameState{}, messagesErr
		}
		return GameState{Attempt: attempt, Step: step, Answers: answers, Messages: messages}, nil
	}
	step, err := s.repository.Step(target.ScenarioID, 1)
	if err != nil {
		return GameState{}, err
	}
	attempt, err := s.repository.CreateGameAttempt(
		domain.Attempt{UserID: userID, ScenarioID: target.ScenarioID, Status: domain.AttemptStatusInProgress, StartedAt: time.Now().UTC(), CurrentStepNumber: 1},
		domain.Message{Author: domain.MessageAuthorInterlocutor, Text: step.Goal},
	)
	if err != nil {
		return GameState{}, err
	}
	return GameState{Attempt: attempt, Step: step}, nil
}

func (s *GameService) Submit(userID, attemptID, optionID int) (GameState, *Completion, error) {
	attempt, err := s.repository.GetGameAttempt(attemptID)
	if err != nil || attempt.UserID != userID {
		return GameState{}, nil, apperrors.ErrAttemptNotFound
	}
	if attempt.Status != domain.AttemptStatusInProgress {
		return GameState{}, nil, apperrors.ErrInvalidAttemptStatusTransition
	}
	step, err := s.repository.Step(attempt.ScenarioID, attempt.CurrentStepNumber)
	if err != nil {
		return GameState{}, nil, err
	}
	var option domain.ScenarioOption
	found := false
	for _, candidate := range step.Options {
		if candidate.ID == optionID {
			option, found = candidate, true
			break
		}
	}
	if !found {
		return GameState{}, nil, apperrors.ErrInvalidAnswer
	}
	answers, err := s.repository.Answers(attemptID)
	if err != nil {
		return GameState{}, nil, err
	}
	for _, answer := range answers {
		if answer.StepID == step.ID {
			return GameState{}, nil, apperrors.ErrInvalidAnswer
		}
	}
	answer := domain.UserAnswer{AttemptID: attemptID, StepID: step.ID, OptionID: &optionID, OptionText: option.Text}
	next, nextErr := s.repository.Step(attempt.ScenarioID, attempt.CurrentStepNumber+1)
	if nextErr == nil {
		if err := s.repository.Complete(func(store GameCompletionStore) error {
			if err := store.SaveAnswer(answer, option.Points, option.Explanation); err != nil {
				return err
			}
			if err := store.SaveMessage(domain.Message{AttemptID: attemptID, Author: domain.MessageAuthorUser, Text: option.Text}); err != nil {
				return err
			}
			if err := store.SaveMessage(domain.Message{AttemptID: attemptID, Author: domain.MessageAuthorInterlocutor, Text: next.Goal}); err != nil {
				return err
			}
			return store.AdvanceAttempt(attemptID, next.Number)
		}); err != nil {
			return GameState{}, nil, err
		}
		attempt.CurrentStepNumber = next.Number
		return GameState{Attempt: attempt, Step: next}, nil, nil
	}
	raw, err := s.repository.AwardedPoints(attemptID)
	if err != nil {
		return GameState{}, nil, err
	}
	raw += option.Points
	maximum := 0
	for number := 1; ; number++ {
		current, stepErr := s.repository.Step(attempt.ScenarioID, number)
		if stepErr != nil {
			break
		}
		maximum += current.MaxPoints
	}
	attempt.Score = domain.NormalizedScore(raw, maximum)
	attempt.Status = domain.AttemptStatusCompleted
	attempt.FinishedAt = time.Now().UTC()
	scenario, scenarioErr := s.repository.Scenario(attempt.ScenarioID)
	if scenarioErr != nil {
		return GameState{}, nil, scenarioErr
	}
	progress := domain.Progress{UserID: userID, LevelID: scenario.LevelID, UserRole: scenario.UserRole, BestScore: attempt.Score, Stars: domain.StarsFromScore(attempt.Score), Attempts: 1, PassedAt: attempt.FinishedAt}
	if err := s.repository.Complete(func(store GameCompletionStore) error {
		if err := store.SaveAnswer(answer, option.Points, option.Explanation); err != nil {
			return err
		}
		if err := store.SaveMessage(domain.Message{AttemptID: attemptID, Author: domain.MessageAuthorUser, Text: option.Text}); err != nil {
			return err
		}
		if err := store.CompleteAttempt(attempt); err != nil {
			return err
		}
		return store.SaveProgress(progress)
	}); err != nil {
		return GameState{}, nil, err
	}
	answers = append(answers, answer)
	breakdown := make([]AnswerBreakdown, 0, len(answers))
	for _, item := range answers {
		option := 0
		if item.OptionID != nil {
			option = *item.OptionID
		}
		breakdown = append(breakdown, AnswerBreakdown{StepID: item.StepID, OptionID: option, Points: item.AwardedPoints, Explanation: item.Explanation, OptionText: item.OptionText})
	}
	breakdown[len(breakdown)-1].Points, breakdown[len(breakdown)-1].Explanation, breakdown[len(breakdown)-1].OptionText = option.Points, option.Explanation, option.Text
	return GameState{}, &Completion{Attempt: attempt, Stars: domain.StarsFromScore(attempt.Score), Answers: answers, Breakdown: breakdown}, nil
}

func (s *GameService) Abandon(userID, attemptID int) error {
	attempt, err := s.repository.GetGameAttempt(attemptID)
	if err != nil || attempt.UserID != userID {
		return apperrors.ErrAttemptNotFound
	}
	if attempt.Status != domain.AttemptStatusInProgress {
		return apperrors.ErrInvalidAttemptStatusTransition
	}
	return s.repository.Abandon(attemptID, time.Now().UTC())
}
