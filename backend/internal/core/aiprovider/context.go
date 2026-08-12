package aiprovider

import (
	"errors"
	"fmt"
	"strings"
)

func (o *Ollama) contextRisk(usage float64) ContextRisk {
	if usage < o.mediumRiskThreshold {
		return ContextRiskLow
	}
	if usage < o.highRiskThreshold {
		return ContextRiskMedium
	}
	return ContextRiskHigh
}

func validateMessages(messages []Message) error {
	if len(messages) == 0 {
		return errors.New("at least one AI message is required")
	}
	for i, message := range messages {
		if message.Role != RoleSystem && message.Role != RoleUser && message.Role != RoleAssistant {
			return fmt.Errorf("message %d has unsupported role %q", i, message.Role)
		}
		if strings.TrimSpace(message.Content) == "" {
			return fmt.Errorf("message %d has empty content", i)
		}
	}
	return nil
}

// EstimatePromptTokens returns a conservative, tokenizer-independent preflight
// estimate. It counts UTF-8 bytes plus fixed chat framing; Ollama's
// prompt_eval_count remains the authoritative post-generation value.
func EstimatePromptTokens(messages []Message) int {
	tokens := 2 // chat start and generation prompt framing
	for _, message := range messages {
		tokens += 4 + len([]byte(message.Role)) + len([]byte(message.Content))
	}
	return tokens
}
