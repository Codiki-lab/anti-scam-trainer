package domain

import (
	"strings"
	"time"
)

type Scenario struct {
	ID             int
	Title          string
	Description    string
	Level          string
	LevelID        int
	TopicID        int
	TopicTitle     string
	UserRole       string
	IsActive       bool
	Status         string
	Archived       bool
	ScamScheme     string
	RiskType       RiskType
	ProductContext ProductContext
	AISystemPrompt string
	FinalRubric    JSONObject
}

type JSONObject map[string]any

type RiskType string

const (
	RiskPhishing          RiskType = "phishing"
	RiskPrepayment        RiskType = "prepayment"
	RiskFakePayment       RiskType = "fake_payment"
	RiskDelivery          RiskType = "delivery"
	RiskExternalMessenger RiskType = "external_messenger"
	RiskAccountTakeover   RiskType = "account_takeover"
	RiskSMSCode           RiskType = "sms_code"
	RiskSocialEngineering RiskType = "social_engineering"
)

type ProductContext struct {
	ItemTitle  string `json:"item_title"`
	Category   string `json:"category"`
	DealMethod string `json:"deal_method"`
	Price      int    `json:"price,omitempty"`
	Currency   string `json:"currency,omitempty"`
	Location   string `json:"location,omitempty"`
	ImageKey   string `json:"image_key,omitempty"`
}

func ValidRiskType(value RiskType) bool {
	switch value {
	case RiskPhishing, RiskPrepayment, RiskFakePayment, RiskDelivery,
		RiskExternalMessenger, RiskAccountTakeover, RiskSMSCode, RiskSocialEngineering:
		return true
	default:
		return false
	}
}

func ValidProductContext(value ProductContext) bool {
	if strings.TrimSpace(value.ItemTitle) == "" || strings.TrimSpace(value.Category) == "" {
		return false
	}
	if value.DealMethod != "delivery" && value.DealMethod != "meetup" && value.DealMethod != "pickup" {
		return false
	}
	if value.Price < 0 || (value.Price > 0 && value.Currency != "RUB") || (value.Price == 0 && value.Currency != "") {
		return false
	}
	if value.ImageKey != "" {
		for _, allowed := range []string{"smartphone", "electronics", "appliance", "camera", "bicycle", "laptop", "headphones", "console"} {
			if value.ImageKey == allowed {
				return true
			}
		}
		return false
	}
	return true
}

type ResponseType string

type MessageRole string

const (
	ScenarioStatusDraft     = "draft"
	ScenarioStatusPublished = "published"
	ScenarioStatusArchived  = "archived"
)

const (
	ResponseTypeMultipleChoice ResponseType = "multiple_choice"
	ResponseTypeSimilarChoice  ResponseType = "similar_choice"
	ResponseTypeMixed          ResponseType = "mixed"
	ResponseTypeFreeText       ResponseType = "free_text"
	MessageRoleUser            MessageRole  = "user"
	MessageRoleAssistant       MessageRole  = "assistant"
)

type ScenarioStep struct {
	ID                  int
	ScenarioID          int
	Number              int
	ResponseType        ResponseType
	Goal                string
	CounterpartyMessage string
	MaxPoints           int
	AIInstruction       string
	FallbackMessage     string
	Options             []ScenarioOption
}

type DialogueMessage struct {
	ID        int
	AttemptID int
	Role      MessageRole
	Text      string
	CreatedAt time.Time
}

type FreePlayConfig struct {
	UserRole       string
	ProductContext ProductContext
	SystemPrompt   string
	FinalRubric    JSONObject
}

type ScenarioOption struct {
	ID          int
	StepID      int
	Text        string
	Reaction    string
	Explanation string
	Points      int
	SortOrder   int
}

func ValidOptionPoints(points int) bool {
	return points == 0 || points == 25 || points == 50 || points == 75 || points == 100
}
