package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func (s *GameService) SubmitAnswer(ctx context.Context, userID, attemptID int, command AnswerCommand) (GameState, *Completion, error) {
	if (command.OptionID == nil) == (command.FreeText == nil) {
		return GameState{}, nil, apperrors.ErrInvalidAnswer
	}
	if command.OptionID != nil {
		if command.Finish {
			return GameState{}, nil, apperrors.ErrInvalidAnswer
		}
		if command.StepID == nil {
			return s.Submit(userID, attemptID, *command.OptionID)
		}
		return s.Submit(userID, attemptID, *command.OptionID, *command.StepID)
	}
	text := strings.TrimSpace(*command.FreeText)
	if text == "" || len([]rune(text)) > 400 || approximateTokens(text) > 220 {
		return GameState{}, nil, apperrors.ErrInvalidAnswer
	}
	return s.submitFreeText(ctx, userID, attemptID, text, command.Finish, command.StepID)
}

func approximateTokens(text string) int {
	fields := strings.Fields(text)
	tokens := 0
	for _, field := range fields {
		runes := len([]rune(field))
		tokens += (runes + 3) / 4
	}
	return tokens
}

func (s *GameService) submitFreeText(ctx context.Context, userID, attemptID int, text string, finish bool, expectedStepID *int) (GameState, *Completion, error) {
	attempt, err := s.repository.GetGameAttempt(attemptID)
	if err != nil || attempt.UserID != userID {
		return GameState{}, nil, apperrors.ErrAttemptNotFound
	}
	if attempt.Status != domain.AttemptStatusInProgress {
		return GameState{}, nil, apperrors.ErrInvalidAttemptStatusTransition
	}
	if finish && attempt.FreeTextCount+1 < 3 {
		return GameState{}, nil, apperrors.ErrInvalidAnswer
	}
	if s.ai == nil {
		return GameState{}, nil, ErrAIUnavailable
	}

	var step domain.ScenarioStep
	var scenario domain.Scenario
	if attempt.Mode != domain.AttemptModeFreePlay {
		scenario, err = s.repository.Scenario(attempt.ScenarioID)
		if err != nil {
			return GameState{}, nil, err
		}
		step, err = s.repository.Step(attempt.ScenarioID, attempt.CurrentStepNumber)
		if err != nil {
			return GameState{}, nil, err
		}
		if expectedStepID != nil && *expectedStepID != step.ID {
			return GameState{}, nil, apperrors.ErrStaleStep
		}
		if step.ResponseType != domain.ResponseTypeMixed && step.ResponseType != domain.ResponseTypeFreeText {
			return GameState{}, nil, apperrors.ErrInvalidAnswer
		}
	} else {
		if expectedStepID != nil && *expectedStepID != attempt.FreeTextCount+1 {
			return GameState{}, nil, apperrors.ErrStaleStep
		}
		step = freePlayStep(attempt.FreeTextCount)
		config, configErr := s.repository.FreePlayConfig(attempt.UserRole)
		if configErr != nil {
			return GameState{}, nil, configErr
		}
		scenario = domain.Scenario{ProductContext: config.ProductContext, AISystemPrompt: config.SystemPrompt, FinalRubric: config.FinalRubric}
	}
	messages, err := s.repository.Messages(attemptID)
	if err != nil {
		return GameState{}, nil, err
	}
	existingAnswers, err := s.repository.Answers(attemptID)
	if err != nil {
		return GameState{}, nil, err
	}
	release, limitErr := s.beforeAI(userID, false)
	if limitErr != nil {
		return GameState{}, nil, limitErr
	}
	defer release()
	aiResult, err := s.evaluate(ctx, attempt, scenario, step, messages, text)
	if err != nil {
		return GameState{}, nil, err
	}
	count := attempt.FreeTextCount + 1
	answer := domain.UserAnswer{AttemptID: attemptID, StepID: step.ID, FreeText: text, AwardedPoints: aiResult.AwardedPoints, Explanation: aiResult.Explanation, Evaluation: &aiResult, TurnNumber: len(existingAnswers) + 1}
	userMessage := domain.DialogueMessage{AttemptID: attemptID, Role: domain.MessageRoleUser, Text: text, CreatedAt: time.Now().UTC()}
	replyMessage := domain.DialogueMessage{AttemptID: attemptID, Role: domain.MessageRoleAssistant, Text: aiResult.Reply, CreatedAt: time.Now().UTC()}

	complete := attempt.Mode == domain.AttemptModeFreePlay || step.ResponseType == domain.ResponseTypeFreeText
	complete = complete && CanFinishFreeText(count, finish)
	if !complete && attempt.Mode != domain.AttemptModeFreePlay && step.ResponseType == domain.ResponseTypeMixed {
		if _, nextErr := s.repository.Step(attempt.ScenarioID, attempt.CurrentStepNumber+1); nextErr != nil {
			complete = true
		}
	}
	if complete {
		return s.completeFreeText(attempt, scenario, answer, userMessage, replyMessage, count)
	}

	nextNumber := attempt.CurrentStepNumber
	if step.ResponseType == domain.ResponseTypeMixed || step.ResponseType == domain.ResponseTypeFreeText {
		if _, nextErr := s.repository.Step(attempt.ScenarioID, attempt.CurrentStepNumber+1); nextErr == nil {
			nextNumber++
		}
	}
	if err := s.repository.Complete(func(store GameCompletionStore) error {
		if err := store.SaveAnswer(answer); err != nil {
			return err
		}
		if err := store.SaveMessage(userMessage); err != nil {
			return err
		}
		if err := store.SaveMessage(replyMessage); err != nil {
			return err
		}
		if err := store.UpdateFreeTextCount(attemptID, count); err != nil {
			return err
		}
		if nextNumber != attempt.CurrentStepNumber {
			return store.AdvanceAttempt(attemptID, nextNumber)
		}
		return nil
	}); err != nil {
		return GameState{}, nil, s.completionError(attempt, err)
	}
	attempt.FreeTextCount, attempt.CurrentStepNumber = count, nextNumber
	nextStep := step
	if attempt.Mode == domain.AttemptModeFreePlay {
		nextStep = freePlayStep(count)
	}
	if nextNumber != step.Number {
		nextStep, err = s.repository.Step(attempt.ScenarioID, nextNumber)
		if err != nil {
			return GameState{}, nil, err
		}
	}
	messages = append(messages, userMessage, replyMessage)
	return GameState{Attempt: attempt, Scenario: scenario, Step: nextStep, Answers: append(existingAnswers, answer), Messages: messages, CanFinishEarly: count >= 2}, nil, nil
}

func (s *GameService) evaluate(ctx context.Context, attempt domain.Attempt, scenario domain.Scenario, step domain.ScenarioStep, history []domain.DialogueMessage, text string) (AIResult, error) {
	system := scenario.AISystemPrompt
	if strings.TrimSpace(system) == "" {
		system = "Ты виртуальный собеседник тренажёра. Верни только JSON с awarded_points, explanation, reply и risk_signals."
	}
	if attempt.Mode == domain.AttemptModeFreePlay {
		kind := "обычный участник сделки"
		if attempt.IsScam != nil && *attempt.IsScam {
			kind = "мошенник"
		}
		base := scenario.AISystemPrompt
		if strings.TrimSpace(base) == "" {
			base = "Веди правдоподобный разговор о сделке."
		}
		system = fmt.Sprintf("%s\nТы %s. Не раскрывай свой тип. Данные сделки: %s. Рубрика: %s. Верни только строгий JSON с awarded_points (0,25,50,75,100), explanation, reply и risk_signals.", base, kind, jsonDocument(scenario.ProductContext), jsonDocument(scenario.FinalRubric))
	} else {
		system += "\nТы играешь роль мошенника в учебном диалоге. Не раскрывай роль напрямую.\nКонтекст: " + scenario.Description + "\nДанные сделки: " + jsonDocument(scenario.ProductContext) + "\nСхема риска: " + scenario.ScamScheme + "\nРубрика: " + jsonDocument(scenario.FinalRubric) + "\nКритерий шага: " + step.AIInstruction
	}
	messages := []AIMessage{{Role: "system", Content: system}}
	start := 0
	if len(history) > 8 {
		start = len(history) - 8
	}
	for _, message := range history[start:] {
		messages = append(messages, AIMessage{Role: string(message.Role), Content: message.Text})
	}
	messages = append(messages, AIMessage{Role: "user", Content: text})
	raw, err := s.ai.Generate(ctx, messages)
	if err != nil {
		return AIResult{}, err
	}
	return DecodeAIResult(raw)
}

func jsonDocument(value domain.JSONObject) string {
	encoded, err := json.Marshal(value)
	if err != nil || value == nil {
		return "{}"
	}
	return string(encoded)
}

func (s *GameService) completeFreeText(attempt domain.Attempt, scenario domain.Scenario, answer domain.UserAnswer, userMessage, replyMessage domain.DialogueMessage, count int) (GameState, *Completion, error) {
	raw, err := s.repository.AwardedPoints(attempt.ID)
	if err != nil {
		return GameState{}, nil, err
	}
	raw += answer.AwardedPoints
	attempt.Score = domain.NormalizedScore(raw, count*100)
	attempt.Status = domain.AttemptStatusCompleted
	attempt.FinishedAt = time.Now().UTC()
	attempt.FreeTextCount = count
	previousAnswers, err := s.repository.Answers(attempt.ID)
	if err != nil {
		return GameState{}, nil, err
	}
	allAnswers := append(previousAnswers, answer)
	breakdown := make([]AnswerBreakdown, 0, len(allAnswers))
	for _, item := range allAnswers {
		entry := AnswerBreakdown{StepID: item.StepID, Points: item.AwardedPoints, Explanation: item.Explanation, FreeText: item.FreeText}
		if item.OptionID != nil {
			entry.OptionID = *item.OptionID
		}
		if item.Evaluation != nil {
			entry.RiskSignals = item.Evaluation.RiskSignals
		}
		breakdown = append(breakdown, entry)
	}
	attempt.FinalBreakdown = breakdown
	riskSignals := make([]string, 0)
	seenRisk := make(map[string]struct{})
	for _, item := range breakdown {
		for _, signal := range item.RiskSignals {
			if _, exists := seenRisk[signal]; !exists {
				seenRisk[signal] = struct{}{}
				riskSignals = append(riskSignals, signal)
			}
		}
	}
	result := domain.AttemptResult{AttemptID: attempt.ID, Score: attempt.Score, Stars: domain.StarsFromScore(attempt.Score), DecisionReview: breakdown, RiskSignals: riskSignals, TopicID: scenario.TopicID, IsScam: attempt.IsScam, SafeActions: []string{"Сохранять общение внутри сервиса", "Не передавать секретные данные", "Остановиться при давлении"}}
	if err := s.repository.Complete(func(store GameCompletionStore) error {
		if err := store.SaveAnswer(answer); err != nil {
			return err
		}
		if err := store.SaveMessage(userMessage); err != nil {
			return err
		}
		if err := store.SaveMessage(replyMessage); err != nil {
			return err
		}
		if err := store.UpdateFreeTextCount(attempt.ID, count); err != nil {
			return err
		}
		if err := store.CompleteAttempt(attempt); err != nil {
			return err
		}
		if attempt.Mode != domain.AttemptModeFreePlay {
			passedAt := time.Time{}
			if result.Stars > 0 {
				passedAt = attempt.FinishedAt
			}
			progress := domain.Progress{UserID: attempt.UserID, LevelID: scenario.LevelID, TopicID: scenario.TopicID, UserRole: scenario.UserRole, BestScore: attempt.Score, Stars: result.Stars, Attempts: 1, PassedAt: passedAt}
			if err := store.SaveProgress(progress); err != nil {
				return err
			}
		}
		return store.FinalizeLearning(&result)
	}); err != nil {
		return GameState{}, nil, s.completionError(attempt, err)
	}
	return GameState{}, &Completion{Attempt: attempt, Stars: domain.StarsFromScore(attempt.Score), Answers: allAnswers, Breakdown: breakdown, Result: result}, nil
}
