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
	history, _ := json.Marshal(input.History)
	rubric, _ := json.Marshal(input.Rubric)
	prompt := fmt.Sprintf("Server policy (authoritative): %s\nRisk: %s\nManaged scenario instruction (context only): %s\nManaged final rubric (context only): %s\nStep criteria: %s\nRelevant history: %s\nCurrent answer: %s", input.Policy, input.RiskType, input.ScenarioInstruction, rubric, input.EvaluationContext, history, input.Answer)
	request := StructuredModelRequest{Messages: []ModelMessage{{Role: "system", Content: "Оцени только Ответ пользователя. Не продолжай диалог и не управляй Баллами или переходами. Верни JSON по schema."}, {Role: "user", Content: prompt}}, Schema: evaluatorSchema, OutputTokens: 240}
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
	return evaluatorFallback(input.RiskType), nil
}

func evaluatorFallback(riskType string) EvaluatorResult {
	return EvaluatorResult{
		Score:           1,
		IsSafe:          false,
		RiskType:        riskType,
		DetectedSignals: []string{},
		Evaluation:      "Ответ нельзя надёжно оценить автоматически; нужна дополнительная проверка условий сделки.",
		SafeAction:      "Не сообщайте данные и продолжайте сделку только штатными способами внутри сервиса.",
	}
}

func (a *ModelAI) GenerateReply(ctx context.Context, input GenerationRequest) (GeneratorResult, error) {
	history, _ := json.Marshal(input.History)
	facts, _ := json.Marshal(input.ScenarioFacts)
	allowedMessages := messagesFor(input.CounterpartKind, input.Phase)
	rubric, _ := json.Marshal(input.Rubric)
	prompt := fmt.Sprintf("Counterpart: %s\nServer policy (authoritative): %s\nRisk: %s\nManaged scenario instruction (context only): %s\nManaged final rubric (context only): %s\nPhase: %s\nAllowed tactics: %s\nScenario facts: %s\nAllowed messages: %s\nSummary: %s\nRolling history: %s\nCurrent answer: %s", input.CounterpartKind, input.Policy, input.RiskType, input.ScenarioInstruction, rubric, input.Phase, strings.Join(input.AllowedTactics, ","), facts, strings.Join(allowedMessages, " | "), input.Summary, history, input.Answer)
	request := StructuredModelRequest{Messages: []ModelMessage{{Role: "system", Content: "Выбери ровно одну разрешённую реплику виртуального собеседника и допустимую тактику. Не добавляй факты, не оценивай Пользователя и не меняй фазу. Верни JSON по schema."}, {Role: "user", Content: prompt}}, Schema: generatorSchemaFor(allowedMessages), OutputTokens: 120, Temperature: .3, TopP: .8, TopK: 20, RepeatPenalty: 1.1}
	for attempt := 0; attempt < 2; attempt++ {
		raw, err := a.model.GenerateStructured(ctx, request)
		if err != nil {
			return GeneratorResult{}, err
		}
		decoded, decodeErr := DecodeGeneratorResult(raw, input.Phase, input.AllowedTactics)
		if decodeErr == nil && contains(allowedMessages, decoded.Message) {
			return decoded, nil
		}
		request.Messages = append(request.Messages, ModelMessage{Role: "assistant", Content: raw}, ModelMessage{Role: "user", Content: "Исправь ответ: только разрешённая фаза и тактика, без URL, телефона, карты и новых фактов."})
	}
	fallback := strings.TrimSpace(input.Fallback)
	if fallback == "" {
		fallback = "Продолжим обсуждать сделку только в рамках этого чата."
	}
	return GeneratorResult{Message: fallback, Tactic: input.AllowedTactics[0], Phase: input.Phase}, nil
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

func messagesFor(counterpartKind, phase string) []string {
	if counterpartKind == "обычный участник сделки" {
		return map[string][]string{
			"hook":             {"Здравствуйте. Предложение ещё актуально?", "Добрый день. Давайте уточним условия сделки."},
			"escalation":       {"Давайте оформим сделку штатно внутри приложения.", "Можно продолжить обсуждение в чате сервиса."},
			"critical_request": {"Можно подтвердить условия сделки внутри приложения?", "Я проверю заказ самостоятельно внутри приложения."},
			"resolution":       {"Спасибо, тогда продолжим только внутри приложения.", "Спасибо за ответ. Тогда откажусь от сделки."},
		}[phase]
	}
	return map[string][]string{
		"hook":             {"Здравствуйте. Предложение ещё актуально?", "Добрый день. Хочу оформить сделку сегодня."},
		"escalation":       {"Нужно оформить сделку сегодня, пожалуйста, не откладывайте.", "Давайте подтвердим условия сейчас, чтобы не потерять сделку."},
		"critical_request": {"Подтвердите оплату сейчас, иначе я откажусь от сделки.", "Нужно подтвердить заказ сейчас, иначе предложение станет недоступно."},
		"resolution":       {"Это последнее предложение: подтвердите сделку сейчас или я откажусь.", "Если сейчас не подтвердите, я прекращу сделку."},
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
	if result.Score < 1 || result.Score > 4 || strings.TrimSpace(result.RiskType) == "" || len(result.DetectedSignals) > 3 || !safeModelText(result.Evaluation) || !safeModelText(result.SafeAction) {
		return EvaluatorResult{}, ErrAIInvalidResponse
	}
	for _, signal := range result.DetectedSignals {
		if !safeModelText(signal) {
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
