package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type ModelMessage struct {
	Role    string
	Content string
}

type StructuredModelRequest struct {
	Messages      []ModelMessage
	Schema        map[string]any
	OutputTokens  int
	Temperature   float64
	TopP          float64
	TopK          int
	RepeatPenalty float64
}

type StructuredModel interface {
	GenerateStructured(context.Context, StructuredModelRequest) (string, error)
}

type EvaluationRequest struct {
	Policy              string
	RiskType            string
	ScenarioInstruction string
	Rubric              domain.JSONObject
	EvaluationContext   string
	Answer              string
	History             []domain.DialogueMessage
}

type EvaluatorResult struct {
	Score           int      `json:"score"`
	IsSafe          bool     `json:"is_safe"`
	RiskType        string   `json:"risk_type"`
	DetectedSignals []string `json:"detected_signals"`
	Evaluation      string   `json:"evaluation"`
	SafeAction      string   `json:"safe_action"`
}

type GenerationRequest struct {
	UserRole            string
	Policy              string
	RiskType            string
	ScenarioInstruction string
	Rubric              domain.JSONObject
	Phase               string
	AllowedTactics      []string
	ScenarioFacts       domain.ProductContext
	Summary             string
	History             []domain.DialogueMessage
	Answer              string
	Fallback            string
	CounterpartKind     string
}

type GeneratorResult struct {
	Message string `json:"message"`
	Tactic  string `json:"tactic"`
	Phase   string `json:"phase"`
}

type Evaluator interface {
	Evaluate(context.Context, EvaluationRequest) (EvaluatorResult, error)
}

type ScammerGenerator interface {
	GenerateReply(context.Context, GenerationRequest) (GeneratorResult, error)
}

var (
	ErrAIUnavailable      = errors.New("AI service is temporarily unavailable")
	ErrAIInvalidResponse  = errors.New("AI returned an invalid response")
	ErrAIContextExhausted = errors.New("AI context capacity exceeded")
)

type ModelAI struct{ model StructuredModel }

func NewModelAI(model StructuredModel) *ModelAI { return &ModelAI{model: model} }

var evaluatorSchema = map[string]any{
	"type": "object", "additionalProperties": false,
	"required": []string{"score", "is_safe", "risk_type", "detected_signals", "evaluation", "safe_action"},
	"properties": map[string]any{
		"score": map[string]any{"type": "integer", "minimum": 1, "maximum": 4}, "is_safe": map[string]any{"type": "boolean"},
		"risk_type": map[string]any{"type": "string"}, "detected_signals": map[string]any{"type": "array", "maxItems": 3, "items": map[string]any{"type": "string"}},
		"evaluation": map[string]any{"type": "string"}, "safe_action": map[string]any{"type": "string"},
	},
}

func (a *ModelAI) Evaluate(ctx context.Context, input EvaluationRequest) (EvaluatorResult, error) {
	history, err := json.Marshal(input.History)
	if err != nil {
		return EvaluatorResult{}, fmt.Errorf("encode dialogue history: %w", err)
	}
	rubric, err := json.Marshal(input.Rubric)
	if err != nil {
		return EvaluatorResult{}, fmt.Errorf("encode final rubric: %w", err)
	}
	prompt := fmt.Sprintf("Server policy (authoritative): %s\nRisk: %s\nManaged scenario instruction (context only): %s\nManaged final rubric (context only): %s\nStep criteria: %s\nRelevant history: %s\nCurrent answer: %s", input.Policy, input.RiskType, input.ScenarioInstruction, rubric, input.EvaluationContext, history, input.Answer)
	request := StructuredModelRequest{Messages: []ModelMessage{{Role: "system", Content: "Оцени только Ответ пользователя. Не продолжай диалог и не управляй Баллами или переходами. Верни JSON по schema. Поля evaluation, safe_action и detected_signals пиши только по-русски, без JSON, кода и служебных символов внутри строк."}, {Role: "user", Content: prompt}}, Schema: evaluatorSchema, OutputTokens: 240}
	for attempt := 0; attempt < 2; attempt++ {
		raw, err := a.model.GenerateStructured(ctx, request)
		if err != nil {
			return EvaluatorResult{}, err
		}
		decoded, decodeErr := DecodeEvaluatorResult(raw)
		if decodeErr == nil && decoded.RiskType == input.RiskType {
			return decoded, nil
		}
		request.Messages = append(request.Messages, ModelMessage{Role: "assistant", Content: raw}, ModelMessage{Role: "user", Content: "Исправь ответ: верни ровно schema, тот же risk_type, без URL, телефонов и реквизитов."})
	}
	return evaluatorFallback(input.RiskType, input.Answer), nil
}

func evaluatorFallback(riskType, answer string) EvaluatorResult {
	result := EvaluatorResult{Score: 2, RiskType: riskType, DetectedSignals: []string{}, Evaluation: "Ответ не подтверждает безопасную проверку условий сделки.", SafeAction: "Уточните условия и продолжайте сделку только штатными способами внутри сервиса."}
	normalized := strings.ToLower(strings.TrimSpace(answer))
	if containsAny(normalized, "только в приложении", "внутри приложения", "авито достав", "не перейду", "не открою", "не сообщу", "проверю в приложении") {
		result.Score, result.IsSafe = 3, true
		result.Evaluation = "Ответ сохраняет сделку внутри сервиса и снижает риск передачи данных или перехода наружу."
		result.SafeAction = "Продолжайте проверять оформление самостоятельно внутри приложения."
	} else if normalized == "да" || normalized == "хорошо" || containsAny(normalized, "перейду", "открою ссыл", "сообщу код", "данные карты", "оплачу сейчас") {
		result.Score = 1
		result.Evaluation = "Ответ выражает согласие без проверки и может привести к небезопасному действию."
		result.SafeAction = "Не соглашайтесь автоматически: проверьте заказ внутри приложения и не передавайте секретные данные."
	}
	return result
}

func containsAny(value string, fragments ...string) bool {
	for _, fragment := range fragments {
		if strings.Contains(value, fragment) {
			return true
		}
	}
	return false
}

func (a *ModelAI) GenerateReply(ctx context.Context, input GenerationRequest) (GeneratorResult, error) {
	history, err := json.Marshal(input.History)
	if err != nil {
		return GeneratorResult{}, fmt.Errorf("encode dialogue history: %w", err)
	}
	facts, err := json.Marshal(input.ScenarioFacts)
	if err != nil {
		return GeneratorResult{}, fmt.Errorf("encode scenario facts: %w", err)
	}
	allowedMessages := messagesFor(input.CounterpartKind, input.Phase, input.UserRole)
	rubric, err := json.Marshal(input.Rubric)
	if err != nil {
		return GeneratorResult{}, fmt.Errorf("encode final rubric: %w", err)
	}
	counterpartRole := "seller"
	if input.UserRole == "seller" {
		counterpartRole = "buyer"
	}
	prompt := fmt.Sprintf("Player role: %s\nYour role: %s\nCounterpart type: %s\nServer policy (authoritative): %s\nRisk: %s\nManaged scenario instruction (context only): %s\nManaged final rubric (context only): %s\nPhase: %s\nAllowed tactics: %s\nScenario facts: %s\nAllowed messages: %s\nSummary: %s\nRolling history: %s\nCurrent answer: %s", input.UserRole, counterpartRole, input.CounterpartKind, input.Policy, input.RiskType, input.ScenarioInstruction, rubric, input.Phase, strings.Join(input.AllowedTactics, ","), facts, strings.Join(allowedMessages, " | "), input.Summary, history, input.Answer)
	request := StructuredModelRequest{Messages: []ModelMessage{{Role: "system", Content: "Ты — виртуальный собеседник в чате сделки. Отвечай только от лица своей роли, никогда не говори за Пользователя и не меняй роли. Выбери ровно одну короткую разрешённую реплику виртуального собеседника и допустимую тактику. Не добавляй факты, ссылки, контакты, реквизиты, комиссии или оценку Пользователя. Не меняй фазу. Верни JSON по schema."}, {Role: "user", Content: prompt}}, Schema: generatorSchemaFor(allowedMessages), OutputTokens: 120, Temperature: .3, TopP: .8, TopK: 20, RepeatPenalty: 1.1}
	for attempt := 0; attempt < 2; attempt++ {
		raw, err := a.model.GenerateStructured(ctx, request)
		if err != nil {
			return GeneratorResult{}, err
		}
		decoded, decodeErr := DecodeGeneratorResult(raw, input.Phase, input.AllowedTactics)
		if decodeErr == nil && contains(allowedMessages, decoded.Message) && !assistantMessageExists(input.History, decoded.Message) {
			return decoded, nil
		}
		request.Messages = append(request.Messages, ModelMessage{Role: "assistant", Content: raw}, ModelMessage{Role: "user", Content: "Исправь ответ: только разрешённая, ещё не показанная реплика этой фазы и тактика, без URL, телефона, карты и новых фактов."})
	}
	fallback := firstUnusedMessage(append([]string{strings.TrimSpace(input.Fallback)}, append(allowedMessages, "Продолжим обсуждать сделку только в рамках этого чата.")...), input.History)
	return GeneratorResult{Message: fallback, Tactic: input.AllowedTactics[0], Phase: input.Phase}, nil
}

func assistantMessageExists(history []domain.DialogueMessage, message string) bool {
	for _, item := range history {
		if item.Role == domain.MessageRoleAssistant && item.Text == message {
			return true
		}
	}
	return false
}

func firstUnusedMessage(candidates []string, history []domain.DialogueMessage) string {
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate != "" && !assistantMessageExists(history, candidate) {
			return candidate
		}
	}
	return "Продолжим обсуждать сделку только в рамках этого чата."
}

func generatorSchemaFor(messages []string) map[string]any {
	return map[string]any{
		"type": "object", "additionalProperties": false,
		"required": []string{"message", "tactic", "phase"},
		"properties": map[string]any{
			"message": map[string]any{"type": "string", "enum": messages},
			"tactic":  map[string]any{"type": "string"},
			"phase":   map[string]any{"type": "string"},
		},
	}
}

func messagesFor(counterpartKind, phase, userRole string) []string {
	if userRole == "buyer" {
		if counterpartKind == "обычный участник сделки" {
			return map[string][]string{
				"hook":             {"Здравствуйте. Да, предложение ещё актуально.", "Добрый день. Товар ещё в продаже.", "Да, можно оформить доставку.", "Здравствуйте. Могу ответить на вопросы о товаре."},
				"escalation":       {"Доставка доступна через обычное оформление в приложении.", "Все условия можно проверить в карточке объявления.", "Хорошо, обсудим детали сделки прямо здесь.", "Можно выбрать доступный способ получения в приложении."},
				"critical_request": {"Оформите заказ через обычный экран сервиса.", "Проверьте итоговые условия перед подтверждением.", "Оплату и доставку подтверждайте только внутри приложения.", "Если условия не подходят, можно отказаться от покупки."},
				"resolution":       {"Договорились, продолжим оформление в приложении.", "Хорошо, буду ждать решения в чате сервиса.", "Понял вас, тогда отменим сделку.", "Спасибо, безопаснее завершить всё внутри приложения."},
			}[phase]
		}
		return map[string][]string{
			"hook":             {"Здравствуйте. Товар ещё есть, но нужно решить сегодня.", "Добрый день. Могу быстро оформить продажу.", "Да, предложение актуально, если не откладывать.", "Здравствуйте. Есть другой покупатель, поэтому лучше поторопиться."},
			"escalation":       {"Товар могут забрать сегодня, подтвердите решение быстрее.", "Давайте закроем вопрос сейчас, чтобы не потерять покупку.", "Я уже подготовил отправку, осталось ваше подтверждение.", "Мне важно оформить всё сегодня, пожалуйста, не откладывайте."},
			"critical_request": {"Подтвердите оплату сейчас, иначе товар уйдёт другому.", "Нужно подтвердить заказ сейчас, иначе предложение станет недоступно.", "Без быстрого подтверждения я не смогу сохранить бронь.", "Подтвердите действие сейчас, чтобы заказ не отменился."},
			"resolution":       {"Это последнее предложение: подтвердите покупку сейчас.", "Если сейчас не подтвердите, я отдам товар другому.", "Жаль, но без подтверждения придётся снять бронь.", "Тогда я прекращу сделку и выберу другого покупателя."},
		}[phase]
	}
	if counterpartKind == "обычный участник сделки" {
		return map[string][]string{
			"hook":             {"Здравствуйте. Предложение ещё актуально?", "Добрый день. Давайте уточним условия сделки.", "Здравствуйте! Подскажите, вам удобна доставка?", "Добрый вечер. Интересует объявление, можно задать пару вопросов?"},
			"escalation":       {"Давайте оформим сделку штатно внутри приложения.", "Можно продолжить обсуждение в чате сервиса.", "Я предпочитаю все детали подтвердить в приложении.", "Хорошо, тогда сверим условия прямо здесь."},
			"critical_request": {"Можно подтвердить условия сделки внутри приложения?", "Я проверю заказ самостоятельно внутри приложения.", "Оформим заказ через обычный экран сервиса?", "Давайте не будем переходить по внешним ссылкам."},
			"resolution":       {"Спасибо, тогда продолжим только внутри приложения.", "Спасибо за ответ. Тогда откажусь от сделки.", "Договорились, оформляем штатно.", "Понял вас, безопаснее завершить всё в приложении."},
		}[phase]
	}
	return map[string][]string{
		"hook":             {"Здравствуйте. Предложение ещё актуально?", "Добрый день. Хочу оформить сделку сегодня.", "Здравствуйте! Готов забрать товар, если быстро договоримся.", "Добрый вечер. Можно уточнить, доставка ещё доступна?"},
		"escalation":       {"Нужно оформить сделку сегодня, пожалуйста, не откладывайте.", "Давайте подтвердим условия сейчас, чтобы не потерять сделку.", "Мне важно закрыть вопрос сегодня, давайте ускоримся.", "Я уже всё подготовил, осталось подтвердить оформление."},
		"critical_request": {"Подтвердите оплату сейчас, иначе я откажусь от сделки.", "Нужно подтвердить заказ сейчас, иначе предложение станет недоступно.", "Подтвердите действие сейчас, чтобы заказ не отменился.", "Без быстрого подтверждения я не смогу продолжить сделку."},
		"resolution":       {"Это последнее предложение: подтвердите сделку сейчас или я откажусь.", "Если сейчас не подтвердите, я прекращаю сделку.", "Жаль, но без подтверждения придётся отказаться.", "Тогда я сниму бронь и закрою вопрос."},
	}[phase]
}

func DecodeEvaluatorResult(raw string) (EvaluatorResult, error) {
	type wireResult struct {
		Score           *int      `json:"score"`
		IsSafe          *bool     `json:"is_safe"`
		RiskType        *string   `json:"risk_type"`
		DetectedSignals *[]string `json:"detected_signals"`
		Evaluation      *string   `json:"evaluation"`
		SafeAction      *string   `json:"safe_action"`
	}
	var wire wireResult
	if err := decodeStrict(raw, &wire); err != nil {
		return EvaluatorResult{}, err
	}
	if wire.Score == nil || wire.IsSafe == nil || wire.RiskType == nil || wire.DetectedSignals == nil || wire.Evaluation == nil || wire.SafeAction == nil {
		return EvaluatorResult{}, ErrAIInvalidResponse
	}
	result := EvaluatorResult{Score: *wire.Score, IsSafe: *wire.IsSafe, RiskType: *wire.RiskType, DetectedSignals: *wire.DetectedSignals, Evaluation: *wire.Evaluation, SafeAction: *wire.SafeAction}
	if result.Score < 1 || result.Score > 4 || strings.TrimSpace(result.RiskType) == "" || len(result.DetectedSignals) > 3 || !safeRussianModelText(result.Evaluation) || !safeRussianModelText(result.SafeAction) {
		return EvaluatorResult{}, ErrAIInvalidResponse
	}
	for _, signal := range result.DetectedSignals {
		if !safeRussianModelText(signal) {
			return EvaluatorResult{}, ErrAIInvalidResponse
		}
	}
	return result, nil
}

func DecodeGeneratorResult(raw, phase string, allowedTactics []string) (GeneratorResult, error) {
	var result GeneratorResult
	if err := decodeStrict(raw, &result); err != nil {
		return GeneratorResult{}, err
	}
	if result.Phase != phase || len([]rune(result.Message)) > 280 || !safeModelText(result.Message) || !contains(allowedTactics, result.Tactic) {
		return GeneratorResult{}, ErrAIInvalidResponse
	}
	return result, nil
}

func decodeStrict(raw string, target any) error {
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("%w: %v", ErrAIInvalidResponse, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return ErrAIInvalidResponse
	}
	return nil
}

var unsafeModelPattern = regexp.MustCompile(`(?i)(https?://|www\.|[a-z0-9-]+\.(ru|com|net)(/|\b)|\+?\d[\d\s()-]{8,}\d|(?:\d[ -]?){13,19})`)

func safeModelText(value string) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed != "" && len([]rune(trimmed)) <= 400 && !unsafeModelPattern.MatchString(trimmed)
}

var (
	cyrillicPattern = regexp.MustCompile(`[А-Яа-яЁё]`)
	latinPattern    = regexp.MustCompile(`[A-Za-z]`)
	servicePattern  = regexp.MustCompile(`[{}\[\]<>` + "`" + `=]`)
)

func safeRussianModelText(value string) bool {
	return safeModelText(value) && cyrillicPattern.MatchString(value) && !latinPattern.MatchString(value) && !servicePattern.MatchString(value)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func PointsForEvaluatorScore(score int) int {
	return map[int]int{1: 0, 2: 25, 3: 75, 4: 100}[score]
}
