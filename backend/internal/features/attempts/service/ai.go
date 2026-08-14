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

var (
	shortRefusalAction = regexp.MustCompile(`^(?:(?:я )?(?:не буду|не стану|не хочу|не собираюсь) (?:открывать ссылку|переходить по ссылке|платить|оплачивать|переводить|вводить данные|сообщать код|называть код|передавать данные|показывать код|отправлять код)|(?:я )?не (?:дам код|сообщу код|назову код|передам данные|покажу код|открою ссылку|перейду по ссылке|оплачу|переведу|введу данные|отправлю код))$`)
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
		a.metrics.record("evaluator", time.Since(started), 0, 0, 0)
		return EvaluatorResult{Score: 2, RiskType: input.RiskType, Evaluation: "Ответ не оценивает условия сделки. Сформулируйте безопасное действие без команд для собеседника.", SafeAction: "Проверьте сделку самостоятельно внутри приложения."}, nil
	}
	if input.RiskType != ordinaryTransactionRisk && isShortRefusal(input.RiskType, input.Answer) {
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
	if input.RiskType != ordinaryTransactionRisk {
		if local := evaluatorFallback(input.RiskType, input.Answer); local.IsSafe || local.Score == 1 {
			local.DetectedSignals = canonicalDetectedSignals(input.RiskType, local.IsSafe)
			a.metrics.record("evaluator", time.Since(started), 0, 0, 0)
			return local, nil
		}
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
	systemPrompt := "Оцени только Ответ пользователя и не продолжай диалог. Верни оценку по шкале: 1 — опасный ответ или согласие на рискованное действие; 2 — неоднозначный, недостаточный или не относящийся к текущей просьбе ответ; 3 — в основном безопасный ответ; 4 — безопасный ответ с явным отказом от риска или безопасной проверкой. Согласие относится к последней просьбе собеседника из Relevant history, даже если Ответ не повторяет само действие: если просьба рискованная, «Понял вас, поступлю именно так» — score 1, is_safe=false. Ответ, который после безопасной проверки оставляет возможность выполнить опасную просьбу, получит score 1 или 2. Уход от темы не подтверждает безопасное действие и получит score 2. Незначительные опечатки, сленг, эмодзи и разговорные частицы не меняют смысл Ответа. Примеры: «Сначала проверю в приложении, но потом могу сделать как вы просите» — score 1, is_safe=false; «Без штатного оформления предоплату не перевожу, ну» — score 4, is_safe=true; «харашо, зделаю как вы прасите» — score 1, is_safe=false; «А какая сегодня погода?» — score 2, is_safe=false. is_safe=true только для score 3 или 4, для score 1 или 2 верни is_safe=false. Не вычисляй итоговый Балл, Звёзды или переходы Прохождения."
	if input.RiskType != ordinaryTransactionRisk {
		systemPrompt += " Короткий однозначный отказ от опасного действия является безопасным Ответом пользователя: не требуй длинной или шаблонной формулировки. Если после отказа Пользователь всё же соглашается на опасное действие, такой ответ небезопасен."
	}
	systemPrompt += " Верни JSON по schema. Поля evaluation, safe_action и detected_signals пиши только по-русски, без JSON, кода и служебных символов внутри строк."
	request := StructuredModelRequest{Messages: []ModelMessage{{Role: "system", Content: systemPrompt}, {Role: "user", Content: prompt}}, Schema: evaluatorSchema, OutputTokens: 120}
	for attempt := 0; attempt < 2; attempt++ {
		raw, err := a.model.GenerateStructured(ctx, request)
		if err != nil {
			a.metrics.record("evaluator", time.Since(started), 1, int64(attempt), 0)
			return EvaluatorResult{}, err
		}
		decoded, decodeErr := DecodeEvaluatorResult(raw)
		if decodeErr == nil && decoded.RiskType == input.RiskType {
			a.metrics.record("evaluator", time.Since(started), 0, int64(attempt), 0)
			return decoded, nil
		}
		request.Messages = append(request.Messages, ModelMessage{Role: "assistant", Content: raw}, ModelMessage{Role: "user", Content: "Исправь ответ: верни ровно schema, тот же risk_type, is_safe=true только для score 3 или 4, без URL, телефонов и реквизитов."})
	}
	a.metrics.record("evaluator", time.Since(started), 0, 1, 1)
	result := evaluatorFallback(input.RiskType, input.Answer)
	if result.Score == 1 {
		result.DetectedSignals = canonicalDetectedSignals(input.RiskType, false)
	}
	return result, nil
}

func isShortRefusal(riskType, answer string) bool {
	if substantiveSuffix.MatchString(answer) {
		return false
	}
	normalized := normalizeEvaluatorAnswer(answer)
	padded := " " + normalized + " "
	if len(strings.Fields(normalized)) > 8 || containsAny(padded, " но ", " хотя ", " потом ", " ладно ", " всё таки ", " все таки ", " всё же ", " все же ", " и ", " а ", " затем ") {
		return false
	}
	return shortRefusalAction.MatchString(normalized) && matchesRiskAction(riskType, normalized)
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
	if explicitRiskyAnswer(riskType, normalized) {
		result.Score = 1
		result.Evaluation = "Ответ выражает согласие без проверки и может привести к небезопасному действию."
		result.SafeAction = "Не соглашайтесь автоматически: проверьте заказ внутри приложения и не передавайте секретные данные."
	} else if explicitSafeAnswer(riskType, normalized) {
		result.Score, result.IsSafe = 3, true
		result.Evaluation = "Ответ сохраняет сделку внутри сервиса и снижает риск передачи данных или перехода наружу."
		result.SafeAction = "Продолжайте проверять оформление самостоятельно внутри приложения."
	}
	return result
}

func explicitRiskyAnswer(riskType, normalized string) bool {
	padded := " " + normalized + " "
	unsafeContinuation := containsAny(padded, " но ", " потом ", " всё же ", " все же ") && containsAny(normalized, "как вы просите", "могу сделать", "сделаю", "выполню", "всё таки", "все таки")
	genericAgreement := normalized == "да" || normalized == "хорошо" || normalized == "сделаю" || normalized == "хорошо сделаю" || normalized == "ладно сделаю" || unsafeContinuation
	return genericAgreement || matchesRiskAction(riskType, normalized) && containsAffirmedAction(normalized,
		"перейду", "открою ссыл", "сообщу код", "перешлю код", "данные карты сообщу", "введу данные", "оплачу сейчас", "оплачу сразу", "оплачу страхов", "переведу деньги", "отдам товар", "сделаю как вы просите")
}

func containsAffirmedAction(value string, fragments ...string) bool {
	for _, fragment := range fragments {
		searchFrom := 0
		for {
			relative := strings.Index(value[searchFrom:], fragment)
			if relative < 0 {
				break
			}
			index := searchFrom + relative
			prefix := strings.TrimSpace(value[:index])
			if !strings.HasSuffix(prefix, "не") {
				return true
			}
			searchFrom = index + len(fragment)
		}
	}
	return false
}

func explicitSafeAnswer(riskType, normalized string) bool {
	padded := " " + normalized + " "
	if containsAny(padded, " но ", " потом ", " всё же ", " все же ") {
		return false
	}
	return matchesRiskAction(riskType, normalized) && containsAny(normalized,
		"только в приложении", "внутри приложения", "проверю в приложении", "проверю заказ в приложении", "авито достав",
		"не перейду", "не открою", "не открываю", "форму не открою", "не сообщу", "не сообщаю", "не перешлю",
		"не перевожу", "не оплачиваю", "внутри заказа", "в этом чате", "только в чате", "продолжу общение только",
		"проверю самостоятельно", "оплату проверю", "возьму паузу", "сначала проверю деньги")
}

func matchesRiskAction(riskType, normalized string) bool {
	switch riskType {
	case "phishing":
		return containsAny(normalized, "ссыл", "форм", "откро", "заказ в приложении", "внутри приложения")
	case "prepayment":
		return containsAny(normalized, "предоплат", "перев", "оплач", "плат", "брон", "штатного оформления")
	case "delivery":
		return containsAny(normalized, "достав", "курьер", "страхов", "внутри заказа")
	case "external_messenger":
		return containsAny(normalized, "мессендж", "телеграм", "ватсап", "в этом чате", "только в чате", "общение только")
	case "account_takeover", "sms_code":
		return containsAny(normalized, "код", "данн")
	case "fake_payment":
		return containsAny(normalized, "оплат", "деньг", "банк", "чек", "поступлен", "товар")
	case "social_engineering":
		return containsAny(normalized, "пауз", "спеш", "давлен", "тороп", "сразу", "пока", "предложение в приложении")
	default:
		return false
	}
}

func canonicalDetectedSignals(riskType string, isSafe bool) []string {
	if isSafe {
		return []string{}
	}
	signal := map[string]string{
		"phishing": "внешняя ссылка", "prepayment": "предоплата", "delivery": "сторонняя доставка",
		"external_messenger": "внешний мессенджер", "account_takeover": "код подтверждения", "sms_code": "код из сообщения",
		"fake_payment": "неподтверждённая оплата", "social_engineering": "давление",
	}[riskType]
	if signal == "" {
		return []string{}
	}
	return []string{signal}
}

func containsAny(value string, fragments ...string) bool {
	for _, fragment := range fragments {
		if strings.Contains(value, fragment) {
			return true
		}
	}
	return false
}

func (a *ModelAI) GenerateReply(_ context.Context, input GenerationRequest) (GeneratorResult, error) {
	started := time.Now()
	variations := tacticVariations(input.AllowedTactics, input.UserRole, input.CounterpartKind)
	if len(input.AllowedTactics) == 0 {
		a.metrics.record("generator", time.Since(started), 1, 0, 0)
		return GeneratorResult{}, ErrAIInvalidResponse
	}
	if fallback := strings.TrimSpace(input.Fallback); safeModelText(fallback) && !assistantMessageExists(input.History, fallback) {
		a.metrics.record("generator", time.Since(started), 0, 0, 0)
		return GeneratorResult{Message: fallback, Tactic: input.AllowedTactics[0], Phase: input.Phase}, nil
	}
	fallback, fallbackTactic := firstUnusedTacticMessage(input.AllowedTactics, variations, input.History)
	if fallback == "" {
		a.metrics.record("generator", time.Since(started), 1, 0, 0)
		return GeneratorResult{}, ErrAIInvalidResponse
	}
	a.metrics.record("generator", time.Since(started), 0, 0, 0)
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
	return "", ""
}

func assistantMessageExists(history []domain.DialogueMessage, message string) bool {
	for _, item := range history {
		if item.Role == domain.MessageRoleAssistant && item.Text == message {
			return true
		}
	}
	return false
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
	if result.Score < 1 || result.Score > 4 || result.IsSafe != (result.Score >= 3) || strings.TrimSpace(result.RiskType) == "" || len(result.DetectedSignals) > 3 || !safeRussianModelText(result.Evaluation) || !safeRussianModelText(result.SafeAction) {
		return EvaluatorResult{}, ErrAIInvalidResponse
	}
	for _, signal := range result.DetectedSignals {
		if !safeRussianModelText(signal) {
			return EvaluatorResult{}, ErrAIInvalidResponse
		}
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

func PointsForEvaluatorScore(score int) int {
	return map[int]int{1: 0, 2: 25, 3: 75, 4: 100}[score]
}
