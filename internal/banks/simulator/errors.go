package simulator

import (
	"errors"
)

var (
	// Client-side / internal failures
	ErrAuthorizationInternal = errors.New("authorization internal error")

	// Business / API outcomes
	ErrAuthorizationRejected    = errors.New("authorization rejected")
	ErrAuthorizationUnavailable = errors.New("authorization service unavailable")
	ErrAuthorizationUnexpected  = errors.New("unexpected authorization error")
)
