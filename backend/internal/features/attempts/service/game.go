package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	cryptorand "crypto/rand"
)

type GameService struct {
	repository GameRepository
	ai         AIProvider
	selectScam func() bool
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
}

type AnswerBreakdown = domain.AnswerBreakdown

type AnswerCommand struct {
	OptionID *int
	FreeText *string
	Finish   bool
}

func (s *GameService) Levels(userID int, role string) ([]OpenLevel, error) {
	levels, progress, err := s.repository.Levels(userID, role)
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
		scenario, scenarioErr := s.repository.PublishedScenario(level.Number, role)
		if scenarioErr != nil {
			continue
		}
		result = append(result, OpenLevel{Level: level, Opened: opened, ScenarioID: scenario.ID})
	}
	return result, nil
}
