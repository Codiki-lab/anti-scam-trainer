package errors

import "errors"

var (
	ErrInvalidAttemptStatusTransition = errors.New("invalid attempt status transition")
	ErrAttemptNotFound                = errors.New("attempt not found")
	ErrUserNotFound                   = errors.New("user not found")
	ErrForbidden                      = errors.New("forbidden")
	ErrInvalidAnswer                  = errors.New("invalid answer")
	ErrStaleStep                      = errors.New("stale step")
	ErrScenarioNotFound               = errors.New("scenario not found")
	ErrInvalidScenarioState           = errors.New("invalid scenario state")
)
