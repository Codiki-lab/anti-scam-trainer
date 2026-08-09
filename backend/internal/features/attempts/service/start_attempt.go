package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"context"
	"strings"
	"time"
)

func (s *GameService) GetState(userID, attemptID int) (GameState, error) {
	attempt, err := s.repository.GetGameAttempt(attemptID)
	if err != nil || attempt.UserID != userID {
		return GameState{}, apperrors.ErrAttemptNotFound
	}
	messages, err := s.repository.Messages(attemptID)
	if err != nil {
		return GameState{}, err
	}
	answers, err := s.repository.Answers(attemptID)
	if err != nil {
		return GameState{}, err
	}
	if attempt.Mode == domain.AttemptModeFreePlay {
		config, configErr := s.repository.FreePlayConfig(attempt.UserRole)
		if configErr != nil {
			return GameState{}, configErr
		}
		return GameState{Attempt: attempt, Scenario: domain.Scenario{ProductContext: config.ProductContext}, Step: freePlayStep(attempt.FreeTextCount), Messages: messages, Answers: answers, CanFinishEarly: attempt.FreeTextCount >= 2}, nil
	}
	scenario, err := s.repository.Scenario(attempt.ScenarioID)
	if err != nil {
		return GameState{}, err
	}
	step := domain.ScenarioStep{}
	if attempt.CurrentStepNumber > 0 {
		step, err = s.repository.Step(attempt.ScenarioID, attempt.CurrentStepNumber)
		if err != nil {
			return GameState{}, err
		}
	}
	return GameState{Attempt: attempt, Scenario: scenario, Step: step, Messages: messages, Answers: answers, CanFinishEarly: attempt.FreeTextCount >= 2}, nil
}

func (s *GameService) Result(userID, attemptID int) (domain.AttemptResult, error) {
	attempt, err := s.repository.GetGameAttempt(attemptID)
	if err != nil || attempt.UserID != userID {
		return domain.AttemptResult{}, apperrors.ErrAttemptNotFound
	}
	if attempt.Status != domain.AttemptStatusCompleted {
		return domain.AttemptResult{}, apperrors.ErrInvalidAttemptStatusTransition
	}
	topical, ok := s.repository.(TopicGameRepository)
	if !ok {
		return domain.AttemptResult{}, apperrors.ErrAttemptNotFound
	}
	return topical.Result(attemptID)
}

func (s *GameService) Start(userID, levelNumber int, role string, topicID ...int) (GameState, error) {
	levels, err := s.Levels(userID, role, topicID...)
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
		scenario, scenarioErr := s.repository.Scenario(attempt.ScenarioID)
		if scenarioErr != nil {
			return GameState{}, scenarioErr
		}
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
		return GameState{Attempt: attempt, Scenario: scenario, Step: step, Answers: answers, Messages: messages, CanFinishEarly: attempt.FreeTextCount >= 2}, nil
	}
	step, err := s.repository.Step(target.ScenarioID, 1)
	if err != nil {
		return GameState{}, err
	}
	attempt, err := s.repository.CreateGameAttempt(domain.Attempt{UserID: userID, ScenarioID: target.ScenarioID, Mode: domain.AttemptModeScenario, UserRole: role, Status: domain.AttemptStatusInProgress, StartedAt: time.Now().UTC(), CurrentStepNumber: 1})
	if err != nil {
		return GameState{}, err
	}
	scenario, err := s.repository.Scenario(target.ScenarioID)
	if err != nil {
		return GameState{}, err
	}
	state := GameState{Attempt: attempt, Scenario: scenario, Step: step}
	visibleMessage := step.CounterpartyMessage
	if strings.TrimSpace(visibleMessage) == "" {
		visibleMessage = step.FallbackMessage
	}
	if strings.TrimSpace(visibleMessage) != "" {
		message := domain.DialogueMessage{AttemptID: attempt.ID, Role: domain.MessageRoleAssistant, Text: visibleMessage, CreatedAt: time.Now().UTC()}
		if err := s.repository.Complete(func(store GameCompletionStore) error { return store.SaveMessage(message) }); err != nil {
			return GameState{}, err
		}
		state.Messages = []domain.DialogueMessage{message}
	}
	return state, nil
}

func (s *GameService) StartFreePlay(ctx context.Context, userID int, role string) (GameState, error) {
	opened := false
	if topical, ok := s.repository.(TopicGameRepository); ok {
		var err error
		opened, err = topical.FreePlayUnlocked(userID, role)
		if err != nil {
			return GameState{}, err
		}
	} else {
		levels, progress, err := s.repository.Levels(userID, role)
		if err != nil {
			return GameState{}, err
		}
		level4ID := 0
		for _, level := range levels {
			if level.Number == 4 {
				level4ID = level.ID
			}
		}
		for _, item := range progress {
			if item.LevelID == level4ID && item.Stars > 0 {
				opened = true
			}
		}
	}
	if !opened {
		return GameState{}, apperrors.ErrForbidden
	}
	if attempt, findErr := s.repository.FindInProgressFreePlay(userID, role); findErr == nil {
		config, configErr := s.repository.FreePlayConfig(role)
		if configErr != nil {
			return GameState{}, configErr
		}
		messages, messagesErr := s.repository.Messages(attempt.ID)
		if messagesErr != nil {
			return GameState{}, messagesErr
		}
		answers, answersErr := s.repository.Answers(attempt.ID)
		if answersErr != nil {
			return GameState{}, answersErr
		}
		return GameState{Attempt: attempt, Scenario: domain.Scenario{ProductContext: config.ProductContext}, Step: freePlayStep(attempt.FreeTextCount), Answers: answers, Messages: messages, CanFinishEarly: attempt.FreeTextCount >= 2}, nil
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
	release, limitErr := s.beforeAI(userID, true)
	if limitErr != nil {
		return GameState{}, limitErr
	}
	defer release()
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
	return GameState{Attempt: attempt, Scenario: freePlayScenario, Step: freePlayStep(0), Messages: []domain.DialogueMessage{message}}, nil
}

func freePlayStep(answered int) domain.ScenarioStep {
	next := answered + 1
	return domain.ScenarioStep{ID: next, Number: next, ResponseType: domain.ResponseTypeFreeText}
}
