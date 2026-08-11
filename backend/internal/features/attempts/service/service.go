package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"anti-scam-trainer/backend/internal/core/ratelimit"
	"context"
	cryptorand "crypto/rand"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type GameService struct {
	repository      GameRepository
	evaluator       Evaluator
	generator       ScammerGenerator
	selectScam      func() bool
	freeTextLimiter *ratelimit.Limiter
	freePlayLimiter *ratelimit.Limiter
	aiGate          *ratelimit.Gate
	contextMu       sync.Mutex
	lastContext     map[string]string
}

func NewGameWithRateLimits(repository GameRepository, evaluator Evaluator, generator ScammerGenerator, freeText, freePlay *ratelimit.Limiter, gate *ratelimit.Gate) *GameService {
	return &GameService{repository: repository, evaluator: evaluator, generator: generator, selectScam: randomScam, freeTextLimiter: freeText, freePlayLimiter: freePlay, aiGate: gate}
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
func NewGameWithAI(repository GameRepository, evaluator Evaluator, generator ScammerGenerator) *GameService {
	return &GameService{repository: repository, evaluator: evaluator, generator: generator, selectScam: randomScam}
}

func NewGameWithDependencies(repository GameRepository, evaluator Evaluator, generator ScammerGenerator, selectScam func() bool) *GameService {
	return &GameService{repository: repository, evaluator: evaluator, generator: generator, selectScam: selectScam}
}

func randomScam() bool {
	var value [1]byte
	if _, err := cryptorand.Read(value[:]); err != nil {
		return true
	}
	return value[0]%2 == 0
}

type OpenLevel struct {
	Level               domain.Level
	Opened              bool
	ScenarioID          int
	ScenarioTitle       string
	ScenarioDescription string
	ResponseType        domain.ResponseType
	InProgressAttemptID int
}

type GameState struct {
	Attempt        domain.Attempt
	Scenario       domain.Scenario
	Step           domain.ScenarioStep
	Answers        []domain.UserAnswer
	Messages       []domain.DialogueMessage
	TotalSteps     int
	CanFinishEarly bool
}

func (s *GameService) state(attempt domain.Attempt, scenario domain.Scenario, step domain.ScenarioStep, answers []domain.UserAnswer, messages []domain.DialogueMessage) GameState {
	total := 5
	if attempt.Mode != domain.AttemptModeFreePlay {
		total = 0
		for number := 1; number <= 100; number++ {
			if _, err := s.repository.Step(attempt.ScenarioID, number); err != nil {
				break
			}
			total = number
		}
	}
	canFinish := attempt.FreeTextCount >= 2 && (attempt.Mode == domain.AttemptModeFreePlay || scenario.Level == "4")
	return GameState{Attempt: attempt, Scenario: scenario, Step: step, Answers: answers, Messages: messages, TotalSteps: total, CanFinishEarly: canFinish}
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
		firstStep, _ := s.repository.Step(scenario.ID, 1)
		inProgressID := 0
		if attempt, findErr := s.repository.FindInProgress(userID, scenario.ID); findErr == nil {
			inProgressID = attempt.ID
		}
		result = append(result, OpenLevel{Level: level, Opened: opened, ScenarioID: scenario.ID, ScenarioTitle: scenario.Title, ScenarioDescription: scenario.Description, ResponseType: firstStep.ResponseType, InProgressAttemptID: inProgressID})
	}
	return result, nil
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
		return s.state(attempt, freePlayScenario(config, attempt.CompactSummary), freePlayStep(attempt.FreeTextCount), answers, messages), nil
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
	return s.state(attempt, scenario, step, answers, messages), nil
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
		return s.state(attempt, scenario, step, answers, messages), nil
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
	state := s.state(attempt, scenario, step, nil, nil)
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
		return s.state(attempt, freePlayScenario(config, attempt.CompactSummary), freePlayStep(attempt.FreeTextCount), answers, messages), nil
	}
	if s.generator == nil {
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
	contextKey, productContext := s.randomFreePlayContext(userID, role)
	attempt := domain.Attempt{UserID: userID, Mode: domain.AttemptModeFreePlay, UserRole: role, IsScam: &isScam, Status: domain.AttemptStatusInProgress, StartedAt: time.Now().UTC(), CompactSummary: contextMarker(contextKey)}
	freePlayScenario := domain.Scenario{ProductContext: productContext, AISystemPrompt: freePlayConfig.SystemPrompt, FinalRubric: freePlayConfig.FinalRubric}
	release, limitErr := s.beforeAI(userID, true)
	if limitErr != nil {
		return GameState{}, limitErr
	}
	defer release()
	initial, err := s.generateReply(ctx, attempt, freePlayScenario, domain.ScenarioStep{}, nil, "Начни разговор о сделке одной короткой репликой", "hook", "")
	if err != nil {
		return GameState{}, err
	}
	message := domain.DialogueMessage{Role: domain.MessageRoleAssistant, Text: initial.Message, CreatedAt: time.Now().UTC()}
	attempt, err = s.repository.StartFreePlay(attempt, message)
	if err != nil {
		return GameState{}, err
	}
	message.AttemptID = attempt.ID
	return s.state(attempt, freePlayScenario, freePlayStep(0), nil, []domain.DialogueMessage{message}), nil
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
		scenario = freePlayScenario(config, attempt.CompactSummary)
	}
	if finish && (attempt.FreeTextCount+1 < 3 || (attempt.Mode != domain.AttemptModeFreePlay && scenario.Level != "4")) {
		return GameState{}, nil, apperrors.ErrInvalidAnswer
	}
	if s.evaluator == nil || (usesGeneratedDialogue(attempt, scenario) && s.generator == nil) {
		return GameState{}, nil, ErrAIUnavailable
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
	evaluation, err := s.evaluateAnswer(ctx, attempt, scenario, step, messages, text)
	if err != nil {
		return GameState{}, nil, err
	}
	count := attempt.FreeTextCount + 1
	storedEvaluation := domain.AIEvaluation{Score: evaluation.Score, IsSafe: evaluation.IsSafe, RiskType: evaluation.RiskType, DetectedSignals: evaluation.DetectedSignals, Evaluation: evaluation.Evaluation, SafeAction: evaluation.SafeAction}
	stepID := step.ID
	if attempt.Mode == domain.AttemptModeFreePlay {
		stepID = 0
	}
	answer := domain.UserAnswer{AttemptID: attemptID, StepID: stepID, FreeText: text, AwardedPoints: PointsForEvaluatorScore(evaluation.Score), Explanation: evaluation.Evaluation, Evaluation: &storedEvaluation, TurnNumber: len(existingAnswers) + 1}
	userMessage := domain.DialogueMessage{AttemptID: attemptID, Role: domain.MessageRoleUser, Text: text, CreatedAt: time.Now().UTC()}
	replyText := step.FallbackMessage
	if usesGeneratedDialogue(attempt, scenario) {
		phase := phaseForTurn(count)
		generated, generateErr := s.generateReply(ctx, attempt, scenario, step, messages, text, phase, attempt.CompactSummary)
		if generateErr != nil {
			return GameState{}, nil, generateErr
		}
		replyText = generated.Message
	} else if next, nextErr := s.repository.Step(attempt.ScenarioID, attempt.CurrentStepNumber+1); nextErr == nil {
		replyText = next.CounterpartyMessage
		if strings.TrimSpace(replyText) == "" {
			replyText = next.FallbackMessage
		}
	}
	if strings.TrimSpace(replyText) == "" {
		replyText = "Диалог завершён. Перейдём к разбору ваших решений."
	}
	replyMessage := domain.DialogueMessage{AttemptID: attemptID, Role: domain.MessageRoleAssistant, Text: replyText, CreatedAt: time.Now().UTC()}
	phase := attempt.DialoguePhase
	summary := attempt.CompactSummary
	if usesGeneratedDialogue(attempt, scenario) {
		phase = phaseForTurn(count)
		if count >= 4 {
			summary = contextPrefix(attempt.CompactSummary) + compactDialogueSummary(append(append(append([]domain.DialogueMessage{}, messages...), userMessage), replyMessage))
		}
	}

	level, _ := strconv.Atoi(scenario.Level)
	complete := false
	if level == 3 {
		_, nextErr := s.repository.Step(attempt.ScenarioID, attempt.CurrentStepNumber+1)
		complete = nextErr != nil
	}
	if usesGeneratedDialogue(attempt, scenario) {
		complete = count >= 5 || finish
	}
	if !complete && attempt.Mode != domain.AttemptModeFreePlay && step.ResponseType == domain.ResponseTypeMixed {
		if _, nextErr := s.repository.Step(attempt.ScenarioID, attempt.CurrentStepNumber+1); nextErr != nil {
			complete = true
		}
	}
	if complete {
		attempt.DialoguePhase, attempt.CompactSummary = phase, summary
		return s.completeFreeText(attempt, scenario, answer, userMessage, replyMessage, count)
	}

	nextNumber := attempt.CurrentStepNumber
	if attempt.Mode != domain.AttemptModeFreePlay && (step.ResponseType == domain.ResponseTypeMixed || step.ResponseType == domain.ResponseTypeFreeText) {
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
		if err := store.UpdateDialogueState(attemptID, count, phase, summary); err != nil {
			return err
		}
		if nextNumber != attempt.CurrentStepNumber {
			return store.AdvanceAttempt(attemptID, nextNumber)
		}
		return nil
	}); err != nil {
		return GameState{}, nil, s.completionError(attempt, err)
	}
	attempt.FreeTextCount, attempt.CurrentStepNumber, attempt.DialoguePhase, attempt.CompactSummary = count, nextNumber, phase, summary
	nextStep := step
	if attempt.Mode == domain.AttemptModeFreePlay {
		nextStep = freePlayStep(count)
	}
	if attempt.Mode != domain.AttemptModeFreePlay && nextNumber != step.Number {
		nextStep, err = s.repository.Step(attempt.ScenarioID, nextNumber)
		if err != nil {
			return GameState{}, nil, err
		}
	}
	messages = append(messages, userMessage, replyMessage)
	return s.state(attempt, scenario, nextStep, append(existingAnswers, answer), messages), nil, nil
}

func (s *GameService) evaluateAnswer(ctx context.Context, attempt domain.Attempt, scenario domain.Scenario, step domain.ScenarioStep, history []domain.DialogueMessage, text string) (EvaluatorResult, error) {
	riskType := string(scenario.RiskType)
	if !domain.ValidRiskType(scenario.RiskType) {
		riskType = primaryRisk(scenario.ScamScheme)
	}
	if attempt.Mode == domain.AttemptModeFreePlay && attempt.IsScam != nil && !*attempt.IsScam {
		riskType = "ordinary_transaction"
	}
	return s.evaluator.Evaluate(ctx, EvaluationRequest{Policy: PolicyFor(attempt.UserRole, riskType), RiskType: riskType, ScenarioInstruction: scenario.AISystemPrompt, Rubric: scenario.FinalRubric, EvaluationContext: step.AIInstruction, Answer: text, History: tailMessages(history, 2)})
}

func (s *GameService) generateReply(ctx context.Context, attempt domain.Attempt, scenario domain.Scenario, step domain.ScenarioStep, history []domain.DialogueMessage, text, phase, summary string) (GeneratorResult, error) {
	riskType := string(scenario.RiskType)
	if !domain.ValidRiskType(scenario.RiskType) {
		riskType = primaryRisk(scenario.ScamScheme)
	}
	kind := "мошенник"
	if attempt.Mode == domain.AttemptModeFreePlay && attempt.IsScam != nil && !*attempt.IsScam {
		kind = "обычный участник сделки"
	}
	tactics := tacticsForPhase(phase)
	if kind == "обычный участник сделки" {
		tactics = honestTacticsForPhase(phase)
		riskType = "ordinary_transaction"
	}
	return s.generator.GenerateReply(ctx, GenerationRequest{UserRole: attempt.UserRole, Policy: PolicyFor(attempt.UserRole, riskType), RiskType: riskType, ScenarioInstruction: scenario.AISystemPrompt, Rubric: scenario.FinalRubric, Phase: phase, AllowedTactics: tactics, ScenarioFacts: scenario.ProductContext, Summary: summary, History: tailMessages(history, 6), Answer: text, Fallback: step.FallbackMessage, CounterpartKind: kind})
}

func primaryRisk(value string) string {
	if field := strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' }); len(field) > 0 {
		return field[0]
	}
	return "social_engineering"
}

func tailMessages(history []domain.DialogueMessage, limit int) []domain.DialogueMessage {
	if len(history) <= limit {
		return history
	}
	return history[len(history)-limit:]
}

func phaseForTurn(turn int) string {
	switch {
	case turn <= 2:
		return "escalation"
	case turn <= 4:
		return "critical_request"
	default:
		return "resolution"
	}
}

func contextMarker(key string) string { return "[free-play-context:" + key + "]\n" }

func contextPrefix(summary string) string {
	if strings.HasPrefix(summary, "[free-play-context:") {
		if end := strings.Index(summary, "]\n"); end >= 0 {
			return summary[:end+2]
		}
	}
	return ""
}

func freePlayScenario(config domain.FreePlayConfig, summary string) domain.Scenario {
	context := config.ProductContext
	for key, candidate := range freePlayContexts(config.UserRole) {
		if strings.HasPrefix(summary, contextMarker(key)) {
			context = candidate
			break
		}
	}
	return domain.Scenario{ProductContext: context, AISystemPrompt: config.SystemPrompt, FinalRubric: config.FinalRubric}
}

func (s *GameService) randomFreePlayContext(userID int, role string) (string, domain.ProductContext) {
	contexts := freePlayContexts(role)
	keys := make([]string, 0, len(contexts))
	for key := range contexts {
		keys = append(keys, key)
	}
	var value [1]byte
	_, _ = cryptorand.Read(value[:])
	key := keys[int(value[0])%len(keys)]
	s.contextMu.Lock()
	defer s.contextMu.Unlock()
	if s.lastContext == nil {
		s.lastContext = make(map[string]string)
	}
	playerKey := strconv.Itoa(userID) + ":" + role
	if len(keys) > 1 && key == s.lastContext[playerKey] {
		for _, candidate := range keys {
			if candidate != key {
				key = candidate
				break
			}
		}
	}
	s.lastContext[playerKey] = key
	return key, contexts[key]
}

func freePlayContexts(role string) map[string]domain.ProductContext {
	if role == "seller" {
		return map[string]domain.ProductContext{
			"phone":  {ItemTitle: "Смартфон Samsung Galaxy S24", Category: "Смартфоны", DealMethod: "delivery", Price: 59000, Currency: "RUB", Location: "Москва", ImageKey: "smartphone"},
			"camera": {ItemTitle: "Фотоаппарат Sony Alpha", Category: "Фототехника", DealMethod: "delivery", Price: 48000, Currency: "RUB", Location: "Казань", ImageKey: "camera"},
			"bike":   {ItemTitle: "Городской велосипед", Category: "Велосипеды", DealMethod: "meetup", Price: 23000, Currency: "RUB", Location: "Санкт-Петербург", ImageKey: "bicycle"},
		}
	}
	return map[string]domain.ProductContext{
		"console": {ItemTitle: "Nintendo Switch OLED", Category: "Игровые приставки", DealMethod: "delivery", Price: 28000, Currency: "RUB", Location: "Москва", ImageKey: "console"},
		"laptop":  {ItemTitle: "Ноутбук ASUS Vivobook", Category: "Ноутбуки", DealMethod: "delivery", Price: 52000, Currency: "RUB", Location: "Тула", ImageKey: "laptop"},
		"phones":  {ItemTitle: "Наушники Sony WH-1000XM5", Category: "Аудио", DealMethod: "pickup", Price: 24000, Currency: "RUB", Location: "Москва", ImageKey: "headphones"},
	}
}

func tacticsForPhase(phase string) []string {
	return map[string][]string{"hook": {"rapport", "convenience"}, "escalation": {"urgency", "authority"}, "critical_request": {"payment", "credential_request"}, "resolution": {"last_chance", "withdrawal"}}[phase]
}

func honestTacticsForPhase(phase string) []string {
	return map[string][]string{"hook": {"greeting", "product_question"}, "escalation": {"clarification", "in_service_details"}, "critical_request": {"in_service_offer", "safety_confirmation"}, "resolution": {"agreement", "polite_withdrawal"}}[phase]
}

func usesGeneratedDialogue(attempt domain.Attempt, scenario domain.Scenario) bool {
	return attempt.Mode == domain.AttemptModeFreePlay || scenario.Level == "4"
}

func compactDialogueSummary(messages []domain.DialogueMessage) string {
	cutoff := len(messages) - 6
	if cutoff <= 0 {
		return ""
	}
	var summary strings.Builder
	for _, message := range messages[:cutoff] {
		line := fmt.Sprintf("%s: %s; ", message.Role, strings.TrimSpace(message.Text))
		if summary.Len()+len([]byte(line)) > 1100 {
			break
		}
		summary.WriteString(line)
	}
	return strings.TrimSpace(summary.String())
}

func (s *GameService) completeFreeText(attempt domain.Attempt, scenario domain.Scenario, answer domain.UserAnswer, userMessage, replyMessage domain.DialogueMessage, count int) (GameState, *Completion, error) {
	raw, err := s.repository.AwardedPoints(attempt.ID)
	if err != nil {
		return GameState{}, nil, err
	}
	raw += answer.AwardedPoints
	previousAnswers, err := s.repository.Answers(attempt.ID)
	if err != nil {
		return GameState{}, nil, err
	}
	allAnswers := append(append([]domain.UserAnswer{}, previousAnswers...), answer)
	attempt.Score = domain.NormalizedScore(raw, len(allAnswers)*100)
	attempt.Status = domain.AttemptStatusCompleted
	attempt.FinishedAt = time.Now().UTC()
	attempt.FreeTextCount = count
	breakdown := make([]AnswerBreakdown, 0, len(allAnswers))
	for _, item := range allAnswers {
		entry := breakdownEntry(item, scenario)
		if attempt.Mode == domain.AttemptModeFreePlay {
			entry.StepID = item.TurnNumber
		}
		if item.OptionID != nil {
			entry.OptionID = *item.OptionID
		}
		breakdown = append(breakdown, entry)
	}
	attempt.FinalBreakdown = breakdown
	riskSignals := make([]domain.RiskSignal, 0)
	seenRisk := make(map[string]struct{})
	safeActions := make([]string, 0)
	seenAction := make(map[string]struct{})
	for _, item := range breakdown {
		for _, signal := range item.RiskSignals {
			if _, exists := seenRisk[signal.Code]; !exists {
				seenRisk[signal.Code] = struct{}{}
				riskSignals = append(riskSignals, signal)
			}
		}
	}
	for _, item := range allAnswers {
		if item.Evaluation != nil && strings.TrimSpace(item.Evaluation.SafeAction) != "" {
			if _, exists := seenAction[item.Evaluation.SafeAction]; !exists {
				seenAction[item.Evaluation.SafeAction] = struct{}{}
				safeActions = append(safeActions, item.Evaluation.SafeAction)
			}
		}
	}
	if len(safeActions) == 0 {
		safeActions = []string{"Сохранять общение внутри сервиса", "Не передавать секретные данные", "Остановиться при давлении"}
	}
	result := domain.AttemptResult{AttemptID: attempt.ID, Score: attempt.Score, Stars: domain.StarsFromScore(attempt.Score), DecisionReview: breakdown, RiskSignals: riskSignals, TopicID: scenario.TopicID, IsScam: attempt.IsScam, SafeActions: safeActions}
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
		if err := store.UpdateDialogueState(attempt.ID, count, attempt.DialoguePhase, attempt.CompactSummary); err != nil {
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
		var reactionMessage *domain.DialogueMessage
		if strings.TrimSpace(option.Reaction) != "" {
			message := domain.DialogueMessage{AttemptID: attemptID, Role: domain.MessageRoleAssistant, Text: option.Reaction, CreatedAt: time.Now().UTC()}
			reactionMessage = &message
		}
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
			if reactionMessage != nil {
				if err := store.SaveMessage(*reactionMessage); err != nil {
					return err
				}
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
		return s.state(attempt, scenario, next, append(answers, answer), messages), nil, nil
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
	result := domain.AttemptResult{AttemptID: attempt.ID, Score: attempt.Score, Stars: stars, TopicID: scenario.TopicID, RiskSignals: []domain.RiskSignal{}, SafeActions: []string{"Сохранять общение внутри сервиса", "Не передавать коды и данные карты", "Проверять статус сделки самостоятельно"}}
	var reactionMessage *domain.DialogueMessage
	if strings.TrimSpace(option.Reaction) != "" {
		message := domain.DialogueMessage{AttemptID: attemptID, Role: domain.MessageRoleAssistant, Text: option.Reaction, CreatedAt: time.Now().UTC()}
		reactionMessage = &message
	}
	if err := s.repository.Complete(func(store GameCompletionStore) error {
		answer.AwardedPoints, answer.Explanation = option.Points, option.Explanation
		if err := store.SaveAnswer(answer); err != nil {
			return err
		}
		if err := store.SaveMessage(userMessage); err != nil {
			return err
		}
		if reactionMessage != nil {
			if err := store.SaveMessage(*reactionMessage); err != nil {
				return err
			}
		}
		if err := store.CompleteAttempt(attempt); err != nil {
			return err
		}
		if err := store.SaveProgress(progress); err != nil {
			return err
		}
		result.DecisionReview = breakdownForAnswers(append(answers, answer), scenario)
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
		entry := breakdownEntry(item, scenario)
		entry.OptionID = option
		breakdown = append(breakdown, entry)
	}
	breakdown[len(breakdown)-1].Points, breakdown[len(breakdown)-1].Explanation, breakdown[len(breakdown)-1].OptionText = option.Points, option.Explanation, option.Text
	result.DecisionReview = breakdown
	return GameState{}, &Completion{Attempt: attempt, Stars: stars, Answers: answers, Breakdown: breakdown, Result: result}, nil
}

func breakdownForAnswers(answers []domain.UserAnswer, scenario domain.Scenario) []domain.AnswerBreakdown {
	result := make([]domain.AnswerBreakdown, 0, len(answers))
	for _, item := range answers {
		optionID := 0
		if item.OptionID != nil {
			optionID = *item.OptionID
		}
		entry := breakdownEntry(item, scenario)
		entry.OptionID = optionID
		result = append(result, entry)
	}
	return result
}

func breakdownEntry(item domain.UserAnswer, scenario domain.Scenario) domain.AnswerBreakdown {
	answerType := "free_text"
	if item.OptionID != nil {
		answerType = "option"
	}
	safeAction := rubricString(scenario.FinalRubric, "safe_action")
	rawSignals := []string{rubricString(scenario.FinalRubric, "risk_signal")}
	if item.Evaluation != nil {
		if strings.TrimSpace(item.Evaluation.SafeAction) != "" {
			safeAction = item.Evaluation.SafeAction
		}
		rawSignals = item.Evaluation.DetectedSignals
	}
	if strings.TrimSpace(safeAction) == "" {
		safeAction = "Проверить сделку самостоятельно внутри приложения и не выполнять просьбу собеседника."
	}
	return domain.AnswerBreakdown{StepID: item.StepID, StepNumber: item.TurnNumber, AnswerType: answerType, OptionText: item.OptionText, FreeText: item.FreeText, Points: item.AwardedPoints, Assessment: domain.AssessmentForPoints(item.AwardedPoints), Explanation: item.Explanation, SafeAction: safeAction, RiskSignals: normalizeRiskSignals(rawSignals, scenario.RiskType)}
}

func rubricString(rubric domain.JSONObject, key string) string {
	value, _ := rubric[key].(string)
	return strings.TrimSpace(value)
}

var signalLabels = map[string]string{
	"external_link":      "Внешняя ссылка или форма вместо штатного экрана",
	"prepayment":         "Предоплата незнакомому человеку до проверки",
	"fake_payment":       "Оплата подтверждается скриншотом, сообщением или чужой формой",
	"fake_delivery":      "Выдуманная комиссия, страховка или активация доставки",
	"external_messenger": "Перевод сделки во внешний канал без проверяемой истории",
	"account_takeover":   "Запрос секрета, который подтверждает действие в аккаунте",
	"sms_code":           "Просьба передать код из SMS",
	"pressure":           "Срочность и давление мешают самостоятельной проверке",
}

func normalizeRiskSignals(values []string, riskType domain.RiskType) []domain.RiskSignal {
	result := make([]domain.RiskSignal, 0, 3)
	seen := map[string]struct{}{}
	for _, value := range values {
		lower := strings.ToLower(value)
		code := "pressure"
		switch {
		case strings.Contains(lower, "ссыл") || strings.Contains(lower, "форм"):
			code = "external_link"
		case strings.Contains(lower, "предоплат") || strings.Contains(lower, "брон"):
			code = "prepayment"
		case strings.Contains(lower, "скрин") || strings.Contains(lower, "поступлен"):
			code = "fake_payment"
		case strings.Contains(lower, "достав") || strings.Contains(lower, "страхов") || strings.Contains(lower, "комисс"):
			code = "fake_delivery"
		case strings.Contains(lower, "мессендж") || strings.Contains(lower, "внешн"):
			code = "external_messenger"
		case strings.Contains(lower, "sms") || strings.Contains(lower, "смс") || strings.Contains(lower, "код"):
			code = "sms_code"
		case strings.Contains(lower, "аккаунт") || strings.Contains(lower, "вход"):
			code = "account_takeover"
		default:
			if normalized := signalCodeForRisk(riskType); normalized != "" {
				code = normalized
			}
		}
		if _, exists := seen[code]; exists {
			continue
		}
		seen[code] = struct{}{}
		result = append(result, domain.RiskSignal{Code: code, Label: signalLabels[code]})
		if len(result) == 3 {
			break
		}
	}
	return result
}

func signalCodeForRisk(riskType domain.RiskType) string {
	return map[domain.RiskType]string{
		domain.RiskPhishing: "external_link", domain.RiskPrepayment: "prepayment", domain.RiskFakePayment: "fake_payment",
		domain.RiskDelivery: "fake_delivery", domain.RiskExternalMessenger: "external_messenger", domain.RiskAccountTakeover: "account_takeover",
		domain.RiskSMSCode: "sms_code", domain.RiskSocialEngineering: "pressure",
	}[riskType]
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
