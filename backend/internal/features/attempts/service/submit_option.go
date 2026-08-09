package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"strings"
	"time"
)

func (s *GameService) Submit(userID, attemptID, optionID int) (GameState, *Completion, error) {
	attempt, err := s.repository.GetGameAttempt(attemptID)
	if err != nil || attempt.UserID != userID {
		return GameState{}, nil, apperrors.ErrAttemptNotFound
	}
	if attempt.Status != domain.AttemptStatusInProgress {
		return GameState{}, nil, apperrors.ErrInvalidAttemptStatusTransition
	}
	if attempt.Mode == domain.AttemptModeFreePlay {
		return GameState{}, nil, apperrors.ErrInvalidAnswer
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
	if step.ResponseType == domain.ResponseTypeFreeText {
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
	answer := domain.UserAnswer{AttemptID: attemptID, StepID: step.ID, OptionID: &optionID, TurnNumber: len(answers) + 1}
	next, nextErr := s.repository.Step(attempt.ScenarioID, attempt.CurrentStepNumber+1)
	if nextErr == nil {
		var nextMessage *domain.DialogueMessage
		if strings.TrimSpace(next.FallbackMessage) != "" {
			message := domain.DialogueMessage{AttemptID: attemptID, Role: domain.MessageRoleAssistant, Text: next.FallbackMessage, CreatedAt: time.Now().UTC()}
			nextMessage = &message
		}
		if err := s.repository.Complete(func(store GameCompletionStore) error {
			answer.AwardedPoints, answer.Explanation = option.Points, option.Explanation
			if err := store.SaveAnswer(answer); err != nil {
				return err
			}
			if nextMessage != nil {
				if err := store.SaveMessage(*nextMessage); err != nil {
					return err
				}
			}
			return store.AdvanceAttempt(attemptID, next.Number)
		}); err != nil {
			return GameState{}, nil, err
		}
		attempt.CurrentStepNumber = next.Number
		messages, messagesErr := s.repository.Messages(attemptID)
		if messagesErr != nil {
			return GameState{}, nil, messagesErr
		}
		return GameState{Attempt: attempt, Step: next, Answers: append(answers, answer), Messages: messages}, nil, nil
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
		answer.AwardedPoints, answer.Explanation = option.Points, option.Explanation
		if err := store.SaveAnswer(answer); err != nil {
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
		breakdown = append(breakdown, AnswerBreakdown{StepID: item.StepID, OptionID: option, Points: item.AwardedPoints, Explanation: item.Explanation, OptionText: item.OptionText, RiskSignals: []string{}})
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
