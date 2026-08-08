package errors

import "errors"

var (
	ErrInvalidAttemptStatusTransition = errors.New("invalid attempt status transition")
	ErrAttemptNotFound                = errors.New("attempt not found")
	ErrUserNotFound                   = errors.New("user not found")
)
