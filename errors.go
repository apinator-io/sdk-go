package apinator

import "fmt"

// Error represents a generic error from the Realtime API.
type Error struct {
	Message string
	Status  int
	Body    string
}

func (e *Error) Error() string {
	if e.Status > 0 {
		return fmt.Sprintf("%s (status: %d)", e.Message, e.Status)
	}
	return e.Message
}

// AuthenticationError represents an authentication error (401).
type AuthenticationError struct {
	Message string
	Status  int
	Body    string
}

func (e *AuthenticationError) Error() string {
	if e.Status > 0 {
		return fmt.Sprintf("%s (status: %d)", e.Message, e.Status)
	}
	return e.Message
}

// ValidationError represents a validation error (400, 422).
type ValidationError struct {
	Message string
	Status  int
	Body    string
}

func (e *ValidationError) Error() string {
	if e.Status > 0 {
		return fmt.Sprintf("%s (status: %d)", e.Message, e.Status)
	}
	return e.Message
}

// ApiError represents a general API error.
type ApiError struct {
	Message string
	Status  int
	Body    string
}

func (e *ApiError) Error() string {
	if e.Status > 0 {
		return fmt.Sprintf("%s (status: %d)", e.Message, e.Status)
	}
	return e.Message
}
