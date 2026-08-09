package app

import (
	"anti-scam-trainer/backend/internal/core/aiprovider"
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/core/ratelimit"
	learning "anti-scam-trainer/backend/internal/features/learning/service"
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type dailyTaskAIAdapter struct {
	provider aiprovider.Provider
	limiter  *ratelimit.Limiter
	gate     *ratelimit.Gate
}
type dailyTaskAIResponse struct {
	Messages []struct {
		Role string `json:"role"`
		Text string `json:"text"`
	} `json:"messages"`
	Verdict    *bool     `json:"verdict"`
	Signals    *[]string `json:"signals"`
	SafeAction *string   `json:"safe_action"`
}

func (a dailyTaskAIAdapter) GenerateDailyTask(ctx context.Context, profile learning.DailyTaskProfile, role domain.UserRole) (domain.DailyTask, error) {
	key := fmt.Sprintf("user:%d", profile.UserID)
	if a.limiter != nil {
		if ok, _ := a.limiter.Allow(key); !ok {
			return domain.DailyTask{}, errors.New("daily task AI limited")
		}
	}
	release := func() {}
	if a.gate != nil {
		var ok bool
		release, ok = a.gate.TryEnter(key)
		if !ok {
			return domain.DailyTask{}, errors.New("daily task AI busy")
		}
	}
	defer release()
	topics := make([]map[string]any, len(profile.Topics))
	for i, topic := range profile.Topics {
		topics[i] = map[string]any{"role": topic.UserRole, "opened": topic.QuizPassed, "completed": topic.Completed, "quiz_best_score": topic.QuizScore}
	}
	recent := make([]map[string]any, len(profile.RecentAttempts))
	for i, attempt := range profile.RecentAttempts {
		recent[i] = map[string]any{"topic_id": attempt.TopicID, "level": attempt.Level, "score": attempt.Score, "stars": attempt.Stars}
	}
	input, _ := json.Marshal(map[string]any{"preferred_role": profile.PreferredRole, "topics": topics, "recent_completed_attempts": recent, "task_role": role})
	result, err := a.provider.Generate(ctx, []aiprovider.Message{{Role: aiprovider.RoleSystem, Content: "Generate one Russian anti-scam dialogue snapshot. Return JSON only: {messages:[{role:user|assistant,text:string}],verdict:boolean,signals:[string],safe_action:string}. Include 2-6 messages, at most 3 concise signals and one safe action. Do not include personal data."}, {Role: aiprovider.RoleUser, Content: string(input)}})
	if err != nil {
		return domain.DailyTask{}, err
	}
	var decoded dailyTaskAIResponse
	if err := json.Unmarshal([]byte(result.Content), &decoded); err != nil {
		return domain.DailyTask{}, err
	}
	if decoded.Verdict == nil || decoded.Signals == nil || decoded.SafeAction == nil {
		return domain.DailyTask{}, errors.New("daily task AI response is incomplete")
	}
	messages := make([]domain.DialogueMessage, len(decoded.Messages))
	for i, message := range decoded.Messages {
		messages[i] = domain.DialogueMessage{Role: domain.MessageRole(message.Role), Text: message.Text}
	}
	task := domain.DailyTask{Role: role, Messages: messages, Verdict: *decoded.Verdict, Signals: *decoded.Signals, SafeAction: *decoded.SafeAction}
	return task, nil
}
