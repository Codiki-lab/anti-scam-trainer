package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrAIUnavailable      = errors.New("AI service is temporarily unavailable")
	ErrAIInvalidResponse  = errors.New("AI returned an invalid response")
	ErrAIContextExhausted = errors.New("AI context capacity exceeded")
)

type AIMessage struct {
	Role    string
	Content string
}

type AIProvider interface {
	Generate(context.Context, []AIMessage) (string, error)
}

type AIResult = domain.AIEvaluation

func DecodeAIResult(raw string) (AIResult, error) {
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.DisallowUnknownFields()
	var result AIResult
	if err := decoder.Decode(&result); err != nil {
		return AIResult{}, fmt.Errorf("%w: %v", ErrAIInvalidResponse, err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return AIResult{}, fmt.Errorf("%w: %v", ErrAIInvalidResponse, err)
	}
	if !domain.ValidOptionPoints(result.AwardedPoints) || strings.TrimSpace(result.Explanation) == "" || strings.TrimSpace(result.Reply) == "" || result.RiskSignals == nil {
		return AIResult{}, ErrAIInvalidResponse
	}
	for _, signal := range result.RiskSignals {
		if strings.TrimSpace(signal) == "" {
			return AIResult{}, ErrAIInvalidResponse
		}
	}
	return result, nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func CanFinishFreeText(answerCount int, finishRequested bool) bool {
	return answerCount >= 5 || (finishRequested && answerCount >= 3)
}
