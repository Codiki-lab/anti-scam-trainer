package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"anti-scam-trainer/backend/internal/core/ratelimit"
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
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

var (
	ErrAIUnavailable      = errors.New("AI service is temporarily unavailable")
	ErrAIInvalidResponse  = errors.New("AI returned an invalid response")
	ErrAIContextExhausted = errors.New("AI context capacity exceeded")
)

type AIMessage struct {
	Role    string
	Content string
}

type AIProvider interface {
	Generate(context.Context, []AIMessage) (string, error)
}

type AIResult = domain.AIEvaluation

func DecodeAIResult(raw string) (AIResult, error) {
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.DisallowUnknownFields()
	var result AIResult
	if err := decoder.Decode(&result); err != nil {
		return AIResult{}, fmt.Errorf("%w: %v", ErrAIInvalidResponse, err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return AIResult{}, fmt.Errorf("%w: %v", ErrAIInvalidResponse, err)
	}
	if !domain.ValidOptionPoints(result.AwardedPoints) || strings.TrimSpace(result.Explanation) == "" || strings.TrimSpace(result.Reply) == "" || result.RiskSignals == nil {
		return AIResult{}, ErrAIInvalidResponse
	}
	for _, signal := range result.RiskSignals {
		if strings.TrimSpace(signal) == "" {
			return AIResult{}, ErrAIInvalidResponse
		}
	}
	return result, nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func CanFinishFreeText(answerCount int, finishRequested bool) bool {
	return answerCount >= 5 || (finishRequested && answerCount >= 3)
}

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

func (s *GameService) Submit(userID, attemptID, optionID int, expectedStepID ...int) (GameState, *Completion, error) {
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
	if len(expectedStepID) > 0 && expectedStepID[0] != step.ID {
		return GameState{}, nil, apperrors.ErrStaleStep
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
	answer := domain.UserAnswer{AttemptID: attemptID, StepID: step.ID, OptionID: &optionID, OptionText: option.Text, TurnNumber: len(answers) + 1}
	userMessage := domain.DialogueMessage{AttemptID: attemptID, Role: domain.MessageRoleUser, Text: option.Text, CreatedAt: time.Now().UTC()}
	next, nextErr := s.repository.Step(attempt.ScenarioID, attempt.CurrentStepNumber+1)
	if nextErr == nil {
		var nextMessage *domain.DialogueMessage
		visibleMessage := next.CounterpartyMessage
		if strings.TrimSpace(visibleMessage) == "" {
			visibleMessage = next.FallbackMessage
		}
		if strings.TrimSpace(visibleMessage) != "" {
			message := domain.DialogueMessage{AttemptID: attemptID, Role: domain.MessageRoleAssistant, Text: visibleMessage, CreatedAt: time.Now().UTC()}
			nextMessage = &message
		}
		if err := s.repository.Complete(func(store GameCompletionStore) error {
			answer.AwardedPoints, answer.Explanation = option.Points, option.Explanation
			if err := store.SaveAnswer(answer); err != nil {
				return err
			}
			if err := store.SaveMessage(userMessage); err != nil {
				return err
			}
			if nextMessage != nil {
				if err := store.SaveMessage(*nextMessage); err != nil {
					return err
				}
			}
			return store.AdvanceAttempt(attemptID, next.Number)
		}); err != nil {
			return GameState{}, nil, s.completionError(attempt, err)
		}
		attempt.CurrentStepNumber = next.Number
		messages, messagesErr := s.repository.Messages(attemptID)
		if messagesErr != nil {
			return GameState{}, nil, messagesErr
		}
		scenario, _ := s.repository.Scenario(attempt.ScenarioID)
		return GameState{Attempt: attempt, Scenario: scenario, Step: next, Answers: append(answers, answer), Messages: messages}, nil, nil
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
	stars := domain.StarsFromScore(attempt.Score)
	passedAt := time.Time{}
	if stars > 0 {
		passedAt = attempt.FinishedAt
	}
	progress := domain.Progress{UserID: userID, LevelID: scenario.LevelID, TopicID: scenario.TopicID, UserRole: scenario.UserRole, BestScore: attempt.Score, Stars: stars, Attempts: 1, PassedAt: passedAt}
	result := domain.AttemptResult{AttemptID: attempt.ID, Score: attempt.Score, Stars: stars, TopicID: scenario.TopicID, RiskSignals: []string{}, SafeActions: []string{"Сохранять общение внутри сервиса", "Не передавать коды и данные карты", "Проверять статус сделки самостоятельно"}}
	if err := s.repository.Complete(func(store GameCompletionStore) error {
		answer.AwardedPoints, answer.Explanation = option.Points, option.Explanation
		if err := store.SaveAnswer(answer); err != nil {
			return err
		}
		if err := store.SaveMessage(userMessage); err != nil {
			return err
		}
		if err := store.CompleteAttempt(attempt); err != nil {
			return err
		}
		if err := store.SaveProgress(progress); err != nil {
			return err
		}
		result.DecisionReview = breakdownForAnswers(append(answers, answer))
		return store.FinalizeLearning(&result)
	}); err != nil {
		return GameState{}, nil, s.completionError(attempt, err)
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
	result.DecisionReview = breakdown
	return GameState{}, &Completion{Attempt: attempt, Stars: stars, Answers: answers, Breakdown: breakdown, Result: result}, nil
}

func breakdownForAnswers(answers []domain.UserAnswer) []domain.AnswerBreakdown {
	result := make([]domain.AnswerBreakdown, 0, len(answers))
	for _, item := range answers {
		optionID := 0
		if item.OptionID != nil {
			optionID = *item.OptionID
		}
		result = append(result, domain.AnswerBreakdown{StepID: item.StepID, OptionID: optionID, OptionText: item.OptionText, FreeText: item.FreeText, Points: item.AwardedPoints, Explanation: item.Explanation})
	}
	return result
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
