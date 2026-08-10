package domain

import "time"

type Scenario struct {
	ID             int
	Title          string
	Description    string
	Level          string
	LevelID        int
	TopicID        int
	UserRole       string
	IsActive       bool
	Status         string
	Archived       bool
	ScamScheme     string
	ProductContext JSONObject
	AISystemPrompt string
	FinalRubric    JSONObject
}

type JSONObject map[string]any

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
	ProductContext JSONObject
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
