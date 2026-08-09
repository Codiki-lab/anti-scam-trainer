package domain

type Scenario struct {
	ID          int
	Title       string
	Description string
	Level       string
	LevelID     int
	UserRole    string
	IsActive    bool
	Status      string
	Archived    bool
}

const (
	ScenarioStatusDraft     = "draft"
	ScenarioStatusPublished = "published"
	ScenarioStatusArchived  = "archived"
)

type ScenarioStep struct {
	ID           int
	ScenarioID   int
	Number       int
	ResponseType string
	Goal         string
	MaxPoints    int
	Options      []ScenarioOption
}

type ScenarioOption struct {
	ID          int
	StepID      int
	Text        string
	Explanation string
	Points      int
	SortOrder   int
}

func ValidOptionPoints(points int) bool {
	return points == 0 || points == 25 || points == 50 || points == 75 || points == 100
}
