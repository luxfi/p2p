// Copyright (C) 2019-2025, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package p2p

import "fmt"

// Error represents an application-level error for peer messaging
type Error struct {
	Code    int32
	Message string
}

// Error implements the error interface
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("app error %d: %s", e.Code, e.Message)
}

var (
	// ErrUnexpected should be used to indicate that a request failed due to a
	// generic error
	ErrUnexpected = &Error{
		Code:    -1,
		Message: "unexpected error",
	}
	// ErrUnregisteredHandler should be used to indicate that a request failed
	// due to it not matching a registered handler
	ErrUnregisteredHandler = &Error{
		Code:    -2,
		Message: "unregistered handler",
	}
	// ErrNotValidator should be used to indicate that a request failed due to
	// the requesting peer not being a validator
	ErrNotValidator = &Error{
		Code:    -3,
		Message: "not a validator",
	}
	// ErrThrottled should be used to indicate that a request failed due to the
	// requesting peer exceeding a rate limit
	ErrThrottled = &Error{
		Code:    -4,
		Message: "throttled",
	}
)
