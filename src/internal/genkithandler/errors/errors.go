// Package errors provides error handling utilities for the Genkit handler.
package errors

import (
	"errors"
	"fmt"
)

// Error types for categorization and handling.
var (
	// ErrNotFound indicates that a requested resource was not found.
	ErrNotFound = errors.New("not found")

	// ErrInvalidInput indicates that the input was invalid.
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized indicates that the request was not authorized.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates that the request is forbidden.
	ErrForbidden = errors.New("forbidden")

	// ErrTimeout indicates that a request timed out.
	ErrTimeout = errors.New("timeout")

	// ErrRateLimit indicates that a rate limit was exceeded.
	ErrRateLimit = errors.New("rate limit exceeded")

	// ErrInternal indicates an internal error.
	ErrInternal = errors.New("internal error")

	// ErrUnavailable indicates that a service is unavailable.
	ErrUnavailable = errors.New("service unavailable")
)

// GenkitError is a custom error type for Genkit handler errors.
type GenkitError struct {
	// Err is the underlying error.
	Err error

	// Code is an error code.
	Code string

	// Message is a human-readable error message.
	Message string

	// Details contains additional error details.
	Details map[string]interface{}

	// Retriable indicates whether the operation can be retried.
	Retriable bool
}

// Error implements the error interface.
func (e *GenkitError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown error"
}

// Unwrap implements the errors.Unwrap interface.
func (e *GenkitError) Unwrap() error {
	return e.Err
}

// New creates a new error with the given message.
func New(message string) error {
	return &GenkitError{
		Message: message,
	}
}

// Errorf creates a new error with the given format and arguments.
func Errorf(format string, args ...interface{}) error {
	return &GenkitError{
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an error with an additional message.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	// If it's already a GenkitError, keep its properties.
	if genkitErr, ok := err.(*GenkitError); ok {
		return &GenkitError{
			Err:       genkitErr.Err,
			Code:      genkitErr.Code,
			Message:   message + ": " + genkitErr.Error(),
			Details:   genkitErr.Details,
			Retriable: genkitErr.Retriable,
		}
	}

	return &GenkitError{
		Err:     err,
		Message: message + ": " + err.Error(),
	}
}

// Wrapf wraps an error with an additional formatted message.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return Wrap(err, fmt.Sprintf(format, args...))
}

// WithCode adds an error code to an error.
func WithCode(err error, code string) error {
	if err == nil {
		return nil
	}

	// If it's already a GenkitError, keep its properties.
	if genkitErr, ok := err.(*GenkitError); ok {
		genkitErr.Code = code
		return genkitErr
	}

	return &GenkitError{
		Err:  err,
		Code: code,
	}
}

// WithDetails adds details to an error.
func WithDetails(err error, details map[string]interface{}) error {
	if err == nil {
		return nil
	}

	// If it's already a GenkitError, keep its properties.
	if genkitErr, ok := err.(*GenkitError); ok {
		if genkitErr.Details == nil {
			genkitErr.Details = details
		} else {
			// Merge details.
			for k, v := range details {
				genkitErr.Details[k] = v
			}
		}
		return genkitErr
	}

	return &GenkitError{
		Err:     err,
		Details: details,
	}
}

// WithRetriable marks an error as retriable or not.
func WithRetriable(err error, retriable bool) error {
	if err == nil {
		return nil
	}

	// If it's already a GenkitError, keep its properties.
	if genkitErr, ok := err.(*GenkitError); ok {
		genkitErr.Retriable = retriable
		return genkitErr
	}

	return &GenkitError{
		Err:       err,
		Retriable: retriable,
	}
}

// IsRetriable returns whether an error is retriable.
func IsRetriable(err error) bool {
	if err == nil {
		return false
	}

	// If it's a GenkitError, check its Retriable field.
	if genkitErr, ok := err.(*GenkitError); ok {
		return genkitErr.Retriable
	}

	// Check for specific error types that are generally retriable.
	if errors.Is(err, ErrTimeout) || errors.Is(err, ErrUnavailable) {
		return true
	}

	// By default, errors are not retriable.
	return false
}

// Is implements the errors.Is interface.
func (e *GenkitError) Is(target error) bool {
	if e.Err == nil {
		return e == target
	}
	return errors.Is(e.Err, target)
}

// NewFlowNotFoundError creates a new error for when a flow is not found.
func NewFlowNotFoundError(flowName string, cause error) error {
	err := &GenkitError{
		Err:     cause,
		Code:    "FLOW_NOT_FOUND",
		Message: fmt.Sprintf("flow %s not found", flowName),
	}
	if cause != nil {
		err.Message = fmt.Sprintf("flow %s not found: %v", flowName, cause)
	}
	return err
}

// NewToolNotFoundError creates a new error for when a tool is not found.
// (Adding this proactively as it's a common pattern)
func NewToolNotFoundError(toolName string, cause error) error {
	err := &GenkitError{
		Err:     cause,
		Code:    "TOOL_NOT_FOUND",
		Message: fmt.Sprintf("tool %s not found", toolName),
	}
	if cause != nil {
		err.Message = fmt.Sprintf("tool %s not found: %v", toolName, cause)
	}
	return err
}
