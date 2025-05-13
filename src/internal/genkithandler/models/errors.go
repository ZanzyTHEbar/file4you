// Package models defines the domain models for the Genkit handler.
package models

// ContextError represents an error related to context operations.
type ContextError struct {
	// Message is a human-readable error message.
	Message string

	// Code is an error code.
	Code string
}

// Error implements the error interface.
func (e ContextError) Error() string {
	return e.Message
}

// Common context errors.
var (
	// ErrTenantNotFound indicates that a tenant ID was not found in the context.
	ErrTenantNotFound = ContextError{
		Message: "tenant not found in context",
		Code:    "tenant_not_found",
	}

	// ErrUserNotFound indicates that a user ID was not found in the context.
	ErrUserNotFound = ContextError{
		Message: "user not found in context",
		Code:    "user_not_found",
	}

	// ErrCorrelationIDNotFound indicates that a correlation ID was not found in the context.
	ErrCorrelationIDNotFound = ContextError{
		Message: "correlation ID not found in context",
		Code:    "correlation_id_not_found",
	}

	// ErrRequestIDNotFound indicates that a request ID was not found in the context.
	ErrRequestIDNotFound = ContextError{
		Message: "request ID not found in context",
		Code:    "request_id_not_found",
	}
)
