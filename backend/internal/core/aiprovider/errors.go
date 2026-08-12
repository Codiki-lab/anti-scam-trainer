package aiprovider

import "fmt"

// TransportError is an unavailable runtime, network failure, or request timeout.
type TransportError struct {
	Err error
}

func (e *TransportError) Error() string { return "ollama transport failure: " + e.Err.Error() }
func (e *TransportError) Unwrap() error { return e.Err }

// OllamaError is a non-success response returned by the Ollama runtime.
type OllamaError struct {
	StatusCode int
	Body       string
}

func (e *OllamaError) Error() string {
	return fmt.Sprintf("ollama returned HTTP %d: %s", e.StatusCode, e.Body)
}

// MalformedResponseError means a success response cannot satisfy Provider's contract.
type MalformedResponseError struct {
	Err error
}

func (e *MalformedResponseError) Error() string { return "invalid ollama response: " + e.Err.Error() }
func (e *MalformedResponseError) Unwrap() error { return e.Err }

// ContextCapacityError is returned before a network call when the requested output
// reserve cannot fit alongside the estimated prompt in the configured context window.
type ContextCapacityError struct {
	EstimatedPromptTokens int
	ReservedOutputTokens  int
	ContextWindowTokens   int
}

func (e *ContextCapacityError) Error() string {
	return fmt.Sprintf("estimated prompt (%d) plus output reserve (%d) exceeds context window (%d)", e.EstimatedPromptTokens, e.ReservedOutputTokens, e.ContextWindowTokens)
}
