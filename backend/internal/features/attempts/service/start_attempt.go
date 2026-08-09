package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"context"
	"strings"
	"time"
)

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
		return GameState{Attempt: attempt, Step: step, Answers: answers, Messages: messages, CanFinishEarly: attempt.FreeTextCount >= 2}, nil
	}
	step, err := s.repository.Step(target.ScenarioID, 1)
	if err != nil {
		return GameState{}, err
	}
	attempt, err := s.repository.CreateGameAttempt(domain.Attempt{UserID: userID, ScenarioID: target.ScenarioID, Mode: domain.AttemptModeScenario, UserRole: role, Status: domain.AttemptStatusInProgress, StartedAt: time.Now().UTC(), CurrentStepNumber: 1})
	if err != nil {
		return GameState{}, err
	}
	state := GameState{Attempt: attempt, Step: step}
	if strings.TrimSpace(step.FallbackMessage) != "" {
		message := domain.DialogueMessage{AttemptID: attempt.ID, Role: domain.MessageRoleAssistant, Text: step.FallbackMessage, CreatedAt: time.Now().UTC()}
		if err := s.repository.Complete(func(store GameCompletionStore) error { return store.SaveMessage(message) }); err != nil {
			return GameState{}, err
		}
		state.Messages = []domain.DialogueMessage{message}
	}
	return state, nil
}

func (s *GameService) StartFreePlay(ctx context.Context, userID int, role string) (GameState, error) {
	levels, progress, err := s.repository.Levels(userID, role)
	if err != nil {
		return GameState{}, err
	}
	level4ID := 0
	for _, level := range levels {
		if level.Number == 4 {
			level4ID = level.ID
			break
		}
	}
	opened := false
	for _, item := range progress {
		if item.LevelID == level4ID && item.Stars > 0 {
			opened = true
			break
		}
	}
	if !opened {
		return GameState{}, apperrors.ErrForbidden
	}
	if attempt, findErr := s.repository.FindInProgressFreePlay(userID, role); findErr == nil {
		messages, messagesErr := s.repository.Messages(attempt.ID)
		if messagesErr != nil {
			return GameState{}, messagesErr
		}
		return GameState{Attempt: attempt, Step: domain.ScenarioStep{ResponseType: domain.ResponseTypeFreeText}, Messages: messages, CanFinishEarly: attempt.FreeTextCount >= 2}, nil
	}
	if s.ai == nil {
		return GameState{}, ErrAIUnavailable
	}
	freePlayConfig, err := s.repository.FreePlayConfig(role)
	if err != nil {
		return GameState{}, err
	}
	isScam := true
	if s.selectScam != nil {
		isScam = s.selectScam()
	}
	attempt := domain.Attempt{UserID: userID, Mode: domain.AttemptModeFreePlay, UserRole: role, IsScam: &isScam, Status: domain.AttemptStatusInProgress, StartedAt: time.Now().UTC()}
	freePlayScenario := domain.Scenario{ProductContext: freePlayConfig.ProductContext, AISystemPrompt: freePlayConfig.SystemPrompt, FinalRubric: freePlayConfig.FinalRubric}
	initial, err := s.evaluate(ctx, attempt, freePlayScenario, domain.ScenarioStep{}, nil, "Начни разговор о сделке одной короткой репликой")
	if err != nil {
		return GameState{}, err
	}
	message := domain.DialogueMessage{Role: domain.MessageRoleAssistant, Text: initial.Reply, CreatedAt: time.Now().UTC()}
	attempt, err = s.repository.StartFreePlay(attempt, message)
	if err != nil {
		return GameState{}, err
	}
	message.AttemptID = attempt.ID
	return GameState{Attempt: attempt, Step: domain.ScenarioStep{ResponseType: domain.ResponseTypeFreeText}, Messages: []domain.DialogueMessage{message}}, nil
}
