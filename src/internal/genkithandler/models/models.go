// Package models defines the domain models for the Genkit handler.
package models

import (
	"context"
	"time"
)

// FlowRequest represents a request to execute a flow.
type FlowRequest struct {
	// FlowName is the name of the flow to execute.
	FlowName string `json:"flowName"`

	// Input is the input to the flow.
	Input interface{} `json:"input"`

	// ModelName is the name of the model to use.
	ModelName string `json:"modelName,omitempty"`

	// Timeout is the timeout for the flow execution.
	Timeout time.Duration `json:"timeout,omitempty"`

	// Tags are key-value pairs for telemetry.
	Tags map[string]string `json:"tags,omitempty"`

	// Metadata is additional metadata for the request.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// FlowResponse represents a response from a flow execution.
type FlowResponse struct {
	// Output is the output from the flow.
	Output interface{} `json:"output"`

	// Duration is how long the flow execution took.
	Duration time.Duration `json:"duration,omitempty"`

	// Metadata is additional metadata from the flow execution.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ToolRequest represents a request to execute a tool.
type ToolRequest struct {
	// ToolName is the name of the tool to execute.
	ToolName string `json:"toolName"`

	// Input is the input to the tool.
	Input interface{} `json:"input"`

	// Timeout is the timeout for the tool execution.
	Timeout time.Duration `json:"timeout,omitempty"`

	// Tags are key-value pairs for telemetry.
	Tags map[string]string `json:"tags,omitempty"`

	// Metadata is additional metadata for the request.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ToolResponse represents a response from a tool execution.
type ToolResponse struct {
	// Output is the output from the tool.
	Output interface{} `json:"output"`

	// Duration is how long the tool execution took.
	Duration time.Duration `json:"duration,omitempty"`

	// Metadata is additional metadata from the tool execution.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// FlowHandler defines a function that handles a flow.
type FlowHandler[I any, O any] func(ctx context.Context, input I) (O, error)

// ToolHandler defines a function that handles a tool.
type ToolHandler[I any, O any] func(ctx context.Context, input I) (O, error)

// FlowRegistry manages flow registrations.
type FlowRegistry interface {
	// Register registers a flow.
	Register(name string, handler interface{}) error

	// Get retrieves a flow handler.
	Get(name string) (interface{}, error)

	// List lists all registered flows.
	List() []string
}

// ToolRegistry manages tool registrations.
type ToolRegistry interface {
	// Register registers a tool.
	Register(name string, handler interface{}) error

	// Get retrieves a tool handler.
	Get(name string) (interface{}, error)

	// List lists all registered tools.
	List() []string
}

// TenantInfo represents information about a tenant.
type TenantInfo struct {
	// ID is the unique identifier of the tenant.
	ID string `json:"id"`

	// Name is the name of the tenant.
	Name string `json:"name"`

	// Limits contains the resource limits for the tenant.
	Limits TenantLimits `json:"limits"`

	// Metadata is additional metadata for the tenant.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TenantLimits represents resource limits for a tenant.
type TenantLimits struct {
	// MaxRequestsPerMinute is the maximum number of requests allowed per minute.
	MaxRequestsPerMinute int `json:"maxRequestsPerMinute"`

	// MaxConcurrentRequests is the maximum number of concurrent requests allowed.
	MaxConcurrentRequests int `json:"maxConcurrentRequests"`

	// MaxRequestSize is the maximum size of a request in bytes.
	MaxRequestSize int64 `json:"maxRequestSize"`
}

// AuditEvent represents an audit event.
type AuditEvent struct {
	// EventType is the type of the event.
	EventType string `json:"eventType"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`

	// TenantID is the ID of the tenant.
	TenantID string `json:"tenantId,omitempty"`

	// UserID is the ID of the user.
	UserID string `json:"userId,omitempty"`

	// Resource is the resource being accessed.
	Resource string `json:"resource,omitempty"`

	// Action is the action being performed.
	Action string `json:"action,omitempty"`

	// Status is the status of the action.
	Status string `json:"status,omitempty"`

	// Details contains additional details about the event.
	Details map[string]interface{} `json:"details,omitempty"`
}

// ContextKey is a type for context keys.
type ContextKey string

const (
	// TenantIDKey is the key for tenant IDs in contexts.
	TenantIDKey ContextKey = "tenantId"

	// UserIDKey is the key for user IDs in contexts.
	UserIDKey ContextKey = "userId"

	// CorrelationIDKey is the key for correlation IDs in contexts.
	CorrelationIDKey ContextKey = "correlationId"

	// RequestIDKey is the key for request IDs in contexts.
	RequestIDKey ContextKey = "requestId"
)

// GetTenantID retrieves the tenant ID from a context.
func GetTenantID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(TenantIDKey).(string)
	return id, ok
}

// WithTenantID adds a tenant ID to a context.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// GetUserID retrieves the user ID from a context.
func GetUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(UserIDKey).(string)
	return id, ok
}

// WithUserID adds a user ID to a context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetCorrelationID retrieves the correlation ID from a context.
func GetCorrelationID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(CorrelationIDKey).(string)
	return id, ok
}

// WithCorrelationID adds a correlation ID to a context.
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// GetRequestID retrieves the request ID from a context.
func GetRequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(RequestIDKey).(string)
	return id, ok
}

// WithRequestID adds a request ID to a context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}
