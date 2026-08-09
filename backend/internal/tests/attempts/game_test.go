package attempts_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	apperrors "anti-scam-trainer/backend/internal/core/errors"
	"anti-scam-trainer/backend/internal/features/attempts/service"
	"errors"
	"testing"
	"time"
)

func TestGameStartRejectsClosedSecondLevel(t *testing.T) {
	repo := newGameRepository()
	game := service.NewGame(repo)
	_, err := game.Start(1, 2, "buyer")
	if !errors.Is(err, apperrors.ErrForbidden) {
		t.Fatalf("Start() error = %v, want forbidden", err)
	}
}

func TestGameCompletesOnlyAfterLastAnswer(t *testing.T) {
	repo := newGameRepository()
	game := service.NewGame(repo)
	state, err := game.Start(1, 1, "buyer")
	if err != nil {
		t.Fatal(err)
	}
	next, finished, err := game.Submit(1, state.Attempt.ID, 11)
	if err != nil || finished != nil || next.Step.Number != 2 {
		t.Fatalf("first answer = (%#v,%#v,%v), want next step", next, finished, err)
	}
	_, finished, err = game.Submit(1, state.Attempt.ID, 21)
	if err != nil || finished == nil || finished.Attempt.Score != 100 || finished.Stars != 3 {
		t.Fatalf("final answer = (%#v,%v), want completed 100/3", finished, err)
	}
	if repo.progress.Stars != 3 || repo.progress.UserRole != "buyer" {
		t.Fatalf("progress=%#v, want buyer three stars", repo.progress)
	}
}

func TestGameStartResumesOwnedAttemptAndRejectsForeignAnswer(t *testing.T) {
	repo := newGameRepository()
	game := service.NewGame(repo)
	started, err := game.Start(1, 1, "buyer")
	if err != nil {
		t.Fatal(err)
	}
	resumed, err := game.Start(1, 1, "buyer")
	if err != nil || resumed.Attempt.ID != started.Attempt.ID {
		t.Fatalf("resume = (%#v, %v), want existing attempt %d", resumed, err, started.Attempt.ID)
	}
	_, _, err = game.Submit(2, started.Attempt.ID, 11)
	if !errors.Is(err, apperrors.ErrAttemptNotFound) {
		t.Fatalf("foreign Submit() error = %v, want attempt not found", err)
	}
}

func TestGameOpensRoleBranchesIndependently(t *testing.T) {
	repo := newGameRepository()
	repo.progressByRole = map[string][]domain.Progress{
		"buyer":  {{UserID: 1, LevelID: 1, UserRole: "buyer", Stars: 1}},
		"seller": {},
	}
	game := service.NewGame(repo)

	buyerLevels, err := game.Levels(1, "buyer")
	if err != nil || !buyerLevels[1].Opened {
		t.Fatalf("buyer levels = %#v, %v; want level 2 open", buyerLevels, err)
	}
	sellerLevels, err := game.Levels(1, "seller")
	if err != nil || sellerLevels[1].Opened {
		t.Fatalf("seller levels = %#v, %v; want level 2 closed", sellerLevels, err)
	}
}

type gameRepository struct {
	attempts       map[int]domain.Attempt
	steps          map[int]domain.ScenarioStep
	answers        []domain.UserAnswer
	progress       domain.Progress
	progressByRole map[string][]domain.Progress
	next           int
}

func newGameRepository() *gameRepository {
	return &gameRepository{attempts: map[int]domain.Attempt{}, next: 1, steps: map[int]domain.ScenarioStep{1: {ID: 1, ScenarioID: 1, Number: 1, MaxPoints: 100, Options: []domain.ScenarioOption{{ID: 11, Points: 100}}}, 2: {ID: 2, ScenarioID: 1, Number: 2, MaxPoints: 100, Options: []domain.ScenarioOption{{ID: 21, Points: 100}}}}}
}
func (r *gameRepository) Levels(_ int, role string) ([]domain.Level, []domain.Progress, error) {
	return []domain.Level{{ID: 1, Number: 1}, {ID: 2, Number: 2}}, r.progressByRole[role], nil
}
func (r *gameRepository) PublishedScenario(level int, role string) (domain.Scenario, error) {
	if (role == "buyer" || role == "seller") && (level == 1 || level == 2) {
		id := level
		if role == "seller" {
			id += 2
		}
		return domain.Scenario{ID: id, LevelID: level, UserRole: role}, nil
	}
	return domain.Scenario{}, errors.New("missing")
}
func (r *gameRepository) Scenario(int) (domain.Scenario, error) {
	return domain.Scenario{ID: 1, LevelID: 1, UserRole: "buyer"}, nil
}
func (r *gameRepository) FindInProgress(user, scenario int) (domain.Attempt, error) {
	for _, a := range r.attempts {
		if a.UserID == user && a.ScenarioID == scenario && a.Status == domain.AttemptStatusInProgress {
			return a, nil
		}
	}
	return domain.Attempt{}, errors.New("missing")
}
func (r *gameRepository) CreateGameAttempt(a domain.Attempt, _ domain.Message) (domain.Attempt, error) {
	a.ID = r.next
	r.next++
	r.attempts[a.ID] = a
	return a, nil
}
func (r *gameRepository) GetGameAttempt(id int) (domain.Attempt, error) {
	a, ok := r.attempts[id]
	if !ok {
		return domain.Attempt{}, errors.New("missing")
	}
	return a, nil
}
func (r *gameRepository) Step(_ int, n int) (domain.ScenarioStep, error) {
	v, ok := r.steps[n]
	if !ok {
		return domain.ScenarioStep{}, errors.New("missing")
	}
	return v, nil
}
func (r *gameRepository) Answers(id int) ([]domain.UserAnswer, error) {
	var out []domain.UserAnswer
	for _, a := range r.answers {
		if a.AttemptID == id {
			out = append(out, a)
		}
	}
	return out, nil
}
func (*gameRepository) Messages(int) ([]domain.Message, error) { return nil, nil }
func (r *gameRepository) AwardedPoints(int) (int, error) {
	total := 0
	for _, a := range r.answers {
		total += a.AwardedPoints
	}
	return total, nil
}
func (r *gameRepository) Advance(id, next int) error {
	a := r.attempts[id]
	a.CurrentStepNumber = next
	r.attempts[id] = a
	return nil
}
func (r *gameRepository) Abandon(id int, _ time.Time) error {
	a := r.attempts[id]
	a.Status = domain.AttemptStatusAbandoned
	r.attempts[id] = a
	return nil
}
func (r *gameRepository) Complete(action func(service.GameCompletionStore) error) error {
	return action(r)
}
func (r *gameRepository) SaveAnswer(a domain.UserAnswer, p int, e string) error {
	a.AwardedPoints = p
	a.Explanation = e
	r.answers = append(r.answers, a)
	return nil
}
func (*gameRepository) SaveMessage(domain.Message) error         { return nil }
func (r *gameRepository) AdvanceAttempt(id, n int) error         { return r.Advance(id, n) }
func (r *gameRepository) CompleteAttempt(a domain.Attempt) error { r.attempts[a.ID] = a; return nil }
func (r *gameRepository) SaveProgress(p domain.Progress) error   { r.progress = p; return nil }
