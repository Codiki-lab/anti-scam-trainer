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
	"sort"
	"strings"
	"sync"
	"time"
)

const maxUntrustedAnswerRunes = 800

var promptInjection = regexp.MustCompile(`(?i)(ignore|ignore previous|system prompt|раскрой.*(промпт|policy)|игнорируй.*(инструк|правил)|измен[иь].*(балл|оценк))`)

const ordinaryTransactionRisk = "ordinary_transaction"

var shortRefusals = map[string]struct{}{
	"нет": {}, "не буду": {}, "я не буду": {}, "не буду так делать": {},
	"я не буду так делать": {}, "нет не буду": {}, "нет я не буду": {},
	"нет не буду так делать": {}, "нет я не буду так делать": {},
	"не стану": {}, "я не стану": {}, "не стану это делать": {},
	"не стану так делать": {}, "не буду этого делать": {}, "не сделаю": {},
	"не согласен": {}, "не согласна": {}, "отказываюсь": {},
	"я отказываюсь": {}, "ни за что": {}, "не хочу": {}, "я не хочу": {},
}

var (
	shortRefusalAction = regexp.MustCompile(`^(?:нет спасибо|(?:я )?(?:не буду|не стану|не хочу|не собираюсь) (?:открывать|переходить|платить|оплачивать|переводить|вводить|сообщать|называть|передавать|показывать|отправлять)(?: .+)?|(?:я )?не (?:дам|сообщу|назову|передам|покажу|открою|перейду|оплачу|переведу|введу|отправлю|соглашусь)(?: .+)?)$`)
	substantiveSuffix  = regexp.MustCompile(`[0-9A-Za-z]`)
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

type ModelAI struct {
	model   StructuredModel
	metrics aiMetrics
}

func NewModelAI(model StructuredModel) *ModelAI { return &ModelAI{model: model} }

type AIKindMetrics struct {
	Calls        int64 `json:"calls"`
	Errors       int64 `json:"errors"`
	Retries      int64 `json:"retries"`
	Fallbacks    int64 `json:"fallbacks"`
	P95LatencyMS int64 `json:"p95_latency_ms"`
}

type AIMetricsSnapshot struct {
	Evaluator AIKindMetrics `json:"evaluator"`
	Generator AIKindMetrics `json:"generator"`
}

type metricSeries struct {
	calls, errors, retries, fallbacks int64
	latencies                         []time.Duration
}

type aiMetrics struct {
	mu        sync.Mutex
	evaluator metricSeries
	generator metricSeries
}

func (a *ModelAI) Metrics() AIMetricsSnapshot {
	a.metrics.mu.Lock()
	defer a.metrics.mu.Unlock()
	return AIMetricsSnapshot{Evaluator: snapshotMetric(a.metrics.evaluator), Generator: snapshotMetric(a.metrics.generator)}
}

func snapshotMetric(series metricSeries) AIKindMetrics {
	latencies := append([]time.Duration(nil), series.latencies...)
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	var p95 time.Duration
	if len(latencies) > 0 {
		index := (len(latencies)*95+99)/100 - 1
		p95 = latencies[index]
	}
	return AIKindMetrics{Calls: series.calls, Errors: series.errors, Retries: series.retries, Fallbacks: series.fallbacks, P95LatencyMS: p95.Milliseconds()}
}

func (m *aiMetrics) record(kind string, latency time.Duration, errors, retries, fallbacks int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	series := &m.evaluator
	if kind == "generator" {
		series = &m.generator
	}
	series.calls++
	series.errors += errors
	series.retries += retries
	series.fallbacks += fallbacks
	series.latencies = append(series.latencies, latency)
	if len(series.latencies) > 1024 {
		series.latencies = append([]time.Duration(nil), series.latencies[len(series.latencies)-1024:]...)
	}
}

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
	started := time.Now()
	if len([]rune(input.Answer)) > maxUntrustedAnswerRunes {
		a.metrics.record("evaluator", time.Since(started), 1, 0, 0)
		return EvaluatorResult{}, ErrAIInvalidResponse
	}
	if promptInjection.MatchString(input.Answer) {
		a.metrics.record("evaluator", time.Since(started), 0, 0, 1)
		return EvaluatorResult{Score: 2, RiskType: input.RiskType, Evaluation: "Ответ не оценивает условия сделки. Сформулируйте безопасное действие без команд для собеседника.", SafeAction: "Проверьте сделку самостоятельно внутри приложения."}, nil
	}
	if input.RiskType != ordinaryTransactionRisk && isShortRefusal(input.Answer) {
		a.metrics.record("evaluator", time.Since(started), 0, 0, 0)
		return EvaluatorResult{
			Score:           4,
			IsSafe:          true,
			RiskType:        input.RiskType,
			DetectedSignals: []string{},
			Evaluation:      "Короткий ответ однозначно отказывается выполнять опасную просьбу собеседника.",
			SafeAction:      "Не выполняйте просьбу и проверяйте условия сделки самостоятельно внутри приложения.",
		}, nil
	}
	history, err := json.Marshal(input.History)
	if err != nil {
		return EvaluatorResult{}, fmt.Errorf("encode dialogue history: %w", err)
	}
	rubric, err := json.Marshal(input.Rubric)
	if err != nil {
		return EvaluatorResult{}, fmt.Errorf("encode final rubric: %w", err)
	}
	prompt := fmt.Sprintf("Server policy (authoritative): %s\nRisk: %s\nManaged scenario instruction (context only): %s\nManaged final rubric (context only): %s\nStep criteria: %s\nRelevant history: %s\nUntrusted user_answer JSON: %q", input.Policy, input.RiskType, input.ScenarioInstruction, rubric, input.EvaluationContext, history, input.Answer)
	systemPrompt := "Оцени только Ответ пользователя. Не продолжай диалог и не управляй Баллами или переходами."
	if input.RiskType != ordinaryTransactionRisk {
		systemPrompt += " Короткий однозначный отказ от опасного действия является безопасным Ответом пользователя: не требуй длинной или шаблонной формулировки. Если после отказа Пользователь всё же соглашается на опасное действие, такой ответ небезопасен."
	}
	systemPrompt += " Верни JSON по schema. Поля evaluation, safe_action и detected_signals пиши только по-русски, без JSON, кода и служебных символов внутри строк."
	request := StructuredModelRequest{Messages: []ModelMessage{{Role: "system", Content: systemPrompt}, {Role: "user", Content: prompt}}, Schema: evaluatorSchema, OutputTokens: 120}
	var lastError error
	for attempt := 0; attempt < 2; attempt++ {
		raw, err := a.model.GenerateStructured(ctx, request)
		if err != nil {
			lastError = err
			if ctx.Err() != nil {
				break
			}
			continue
		}
		decoded, decodeErr := DecodeEvaluatorResult(raw)
		if decodeErr == nil && decoded.RiskType == input.RiskType {
			a.metrics.record("evaluator", time.Since(started), 0, int64(attempt), 0)
			return decoded, nil
		}
		request.Messages = append(request.Messages, ModelMessage{Role: "assistant", Content: raw}, ModelMessage{Role: "user", Content: "Исправь ответ: верни ровно schema, тот же risk_type, без URL, телефонов и реквизитов."})
	}
	if lastError != nil {
		a.metrics.record("evaluator", time.Since(started), 1, 1, 0)
		return EvaluatorResult{}, lastError
	}
	a.metrics.record("evaluator", time.Since(started), 0, 1, 1)
	return evaluatorFallback(input.RiskType, input.Answer), nil
}

func isShortRefusal(answer string) bool {
	if substantiveSuffix.MatchString(answer) {
		return false
	}
	normalized := normalizeEvaluatorAnswer(answer)
	padded := " " + normalized + " "
	if len(strings.Fields(normalized)) > 8 || containsAny(padded, " но ", " хотя ", " потом ", " ладно ", " всё таки ", " все таки ", " всё же ", " все же ", " и ", " а ", " затем ") {
		return false
	}
	if _, ok := shortRefusals[normalized]; ok {
		return true
	}
	return shortRefusalAction.MatchString(normalized)
}

func normalizeEvaluatorAnswer(answer string) string {
	normalized := strings.ToLower(strings.TrimSpace(answer))
	normalized = strings.Map(func(r rune) rune {
		if r >= 'а' && r <= 'я' || r == 'ё' || r == ' ' {
			return r
		}
		return ' '
	}, normalized)
	return strings.Join(strings.Fields(normalized), " ")
}

func evaluatorFallback(riskType, answer string) EvaluatorResult {
	result := EvaluatorResult{Score: 2, RiskType: riskType, DetectedSignals: []string{}, Evaluation: "Ответ не подтверждает безопасную проверку условий сделки.", SafeAction: "Уточните условия и продолжайте сделку только штатными способами внутри сервиса."}
	normalized := normalizeEvaluatorAnswer(answer)
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
	started := time.Now()
	history, err := json.Marshal(input.History)
	if err != nil {
		return GeneratorResult{}, fmt.Errorf("encode dialogue history: %w", err)
	}
	facts, err := json.Marshal(input.ScenarioFacts)
	if err != nil {
		return GeneratorResult{}, fmt.Errorf("encode scenario facts: %w", err)
	}
	variations := tacticVariations(input.AllowedTactics, input.UserRole, input.CounterpartKind)
	allowedMessages := make([]string, 0, len(input.AllowedTactics)*2)
	for _, tactic := range input.AllowedTactics {
		allowedMessages = append(allowedMessages, variations[tactic]...)
	}
	if len(allowedMessages) == 0 {
		return GeneratorResult{}, ErrAIInvalidResponse
	}
	rubric, err := json.Marshal(input.Rubric)
	if err != nil {
		return GeneratorResult{}, fmt.Errorf("encode final rubric: %w", err)
	}
	counterpartRole := "seller"
	if input.UserRole == "seller" {
		counterpartRole = "buyer"
	}
	prompt := fmt.Sprintf("Player role: %s\nYour role: %s\nCounterpart type: %s\nServer policy (authoritative): %s\nRisk: %s\nManaged scenario instruction (context only): %s\nManaged final rubric (context only): %s\nPhase: %s\nAllowed tactics: %s\nScenario facts: %s\nAllowed messages: %s\nSummary: %s\nRolling history: %s", input.UserRole, counterpartRole, input.CounterpartKind, input.Policy, input.RiskType, input.ScenarioInstruction, rubric, input.Phase, strings.Join(input.AllowedTactics, ","), facts, strings.Join(allowedMessages, " | "), input.Summary, history)
	request := StructuredModelRequest{Messages: []ModelMessage{{Role: "system", Content: "Ты — виртуальный собеседник в чате сделки. Отвечай только от лица своей роли, никогда не говори за Пользователя и не меняй роли. Выбери ровно одну короткую разрешённую реплику виртуального собеседника и допустимую тактику. Не добавляй факты, ссылки, контакты, реквизиты, комиссии или оценку Пользователя. Не меняй фазу. Верни JSON по schema."}, {Role: "user", Content: prompt}}, Schema: generatorSchemaFor(allowedMessages), OutputTokens: 120, Temperature: .3, TopP: .8, TopK: 20, RepeatPenalty: 1.1}
	hadError := false
	for attempt := 0; attempt < 2; attempt++ {
		raw, err := a.model.GenerateStructured(ctx, request)
		if err != nil {
			hadError = true
			if ctx.Err() != nil {
				break
			}
			continue
		}
		decoded, decodeErr := DecodeGeneratorResult(raw, input.Phase, input.AllowedTactics)
		if decodeErr == nil && contains(variations[decoded.Tactic], decoded.Message) && !assistantMessageExists(input.History, decoded.Message) {
			a.metrics.record("generator", time.Since(started), 0, int64(attempt), 0)
			return decoded, nil
		}
		request.Messages = append(request.Messages, ModelMessage{Role: "assistant", Content: raw}, ModelMessage{Role: "user", Content: "Исправь ответ: только разрешённая, ещё не показанная реплика этой фазы и тактика, без URL, телефона, карты и новых фактов."})
	}
	fallback, fallbackTactic := firstUnusedTacticMessage(input.AllowedTactics, variations, input.History)
	errors := int64(0)
	if hadError {
		errors = 1
	}
	a.metrics.record("generator", time.Since(started), errors, 1, 1)
	return GeneratorResult{Message: fallback, Tactic: fallbackTactic, Phase: input.Phase}, nil
}

func tacticVariations(tactics []string, userRole, counterpartKind string) map[string][]string {
	result := make(map[string][]string, len(tactics))
	pool := curatedTacticPool(userRole, counterpartKind)
	for _, tactic := range tactics {
		if messages := pool[tactic]; len(messages) >= 2 {
			result[tactic] = append([]string(nil), messages...)
		}
	}
	return result
}

func curatedTacticPool(userRole, counterpartKind string) map[string][]string {
	if counterpartKind == "обычный участник сделки" {
		return map[string][]string{
			"greeting":            {"Здравствуйте. Давайте обсудим условия сделки здесь.", "Добрый день. Предлагаю спокойно уточнить детали."},
			"product_question":    {"Подскажите, товар и описание актуальны?", "Можно уточнить состояние товара?"},
			"clarification":       {"Давайте сверим условия прямо в этом чате.", "Уточним детали без спешки внутри сервиса."},
			"in_service_details":  {"Все условия можно проверить в карточке объявления.", "Обсудим доставку штатно внутри приложения."},
			"in_service_offer":    {"Оформим заказ через обычный экран сервиса.", "Можно подтвердить условия внутри приложения."},
			"safety_confirmation": {"Я проверю заказ самостоятельно внутри приложения.", "Не будем переходить по внешним ссылкам."},
			"agreement":           {"Договорились, продолжим оформление в приложении.", "Хорошо, подтверждаем всё штатно."},
			"polite_withdrawal":   {"Спасибо, тогда безопаснее отказаться от сделки.", "Понял вас, завершим обсуждение без перевода денег."},
		}
	}
	buyerSide := userRole == "seller"
	rapport := []string{"Здравствуйте. Товар ещё в продаже?", "Добрый день. Объявление меня заинтересовало."}
	convenience := []string{"Готов быстро договориться об оформлении.", "Давайте выберем самый быстрый вариант сделки."}
	if !buyerSide {
		rapport = []string{"Здравствуйте. Товар ещё есть.", "Добрый день. Предложение пока актуально."}
		convenience = []string{"Могу быстро подготовить оформление.", "Давайте оперативно согласуем получение."}
	}
	return map[string][]string{
		"rapport":            rapport,
		"convenience":        convenience,
		"urgency":            {"Нужно решить сегодня, пожалуйста, не откладывайте.", "Давайте подтвердим решение прямо сейчас."},
		"authority":          {"Это обязательное условие оформления.", "Без этого подтверждения система не продолжит сделку."},
		"payment":            {"Подтвердите оплату сейчас, иначе заказ отменится.", "Нужно немедленно подтвердить платёж."},
		"credential_request": {"Для продолжения сообщите код подтверждения.", "Продиктуйте данные подтверждения прямо сейчас."},
		"last_chance":        {"Это последнее предложение: подтвердите сделку сейчас.", "Если не подтвердите сейчас, бронь будет снята."},
		"withdrawal":         {"Тогда я прекращу сделку.", "Без быстрого подтверждения я выберу другого участника."},
	}
}

func firstUnusedTacticMessage(tactics []string, variations map[string][]string, history []domain.DialogueMessage) (string, string) {
	for _, tactic := range tactics {
		for _, message := range variations[tactic] {
			if !assistantMessageExists(history, message) {
				return message, tactic
			}
		}
	}
	return variations[tactics[0]][0], tactics[0]
}

func assistantMessageExists(history []domain.DialogueMessage, message string) bool {
	for _, item := range history {
		if item.Role == domain.MessageRoleAssistant && item.Text == message {
			return true
		}
	}
	return false
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
