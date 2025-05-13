// Package genkithandler provides a simplified interface for integrating with Genkit.
package genkithandler

import (
"context"
// "fmt" // No longer needed directly for error formatting here

"file4you/internal/genkithandler/errors" // Import custom errors
"github.com/firebase/genkit/go/ai"
"github.com/firebase/genkit/go/genkit"
)

// DefineTool defines a new Genkit tool (action) and registers it with the provided Genkit instance.
// It's a wrapper around genkit.DefineTool.
// The fn parameter is the actual tool implementation.
// Its signature must match what genkit.DefineTool expects: func(ctx *ai.ToolContext, input In) (Out, error)
func DefineTool[In, Out any](
g *genkit.Genkit,
name, description string,
fn func(ctx *ai.ToolContext, input In) (Out, error),
) (ai.Tool, error) {
if g == nil {
return nil, errors.New("genkit instance is nil")
}
if name == "" {
return nil, errors.New("tool name cannot be empty")
}
if description == "" {
// Consider if description can be optional or if an error is appropriate.
// For now, maintaining previous logic.
return nil, errors.New("tool description cannot be empty")
}
if fn == nil {
return nil, errors.New("tool function (fn) cannot be nil")
}

// genkit.DefineTool takes the Genkit instance, name, description, and the function.
tool := genkit.DefineTool(g, name, description, fn)
if tool == nil {
// This condition might be hard to hit if DefineTool panics or always returns non-nil.
return nil, errors.Errorf("failed to define tool %s", name)
}
return tool, nil
}

// LookupTool retrieves a previously defined Genkit tool by its name from the Genkit instance.
// Returns nil if the tool is not found.
// Note: The original audit mentioned genkit.LookupTool returns (tool, error).
// The provided file content shows it returns ai.Tool directly. Adhering to file content.
func LookupTool(g *genkit.Genkit, name string) ai.Tool {
if g == nil || name == "" {
return nil
}
return genkit.LookupTool(g, name)
}

// ExecuteTool looks up a tool by name and executes it with the provided input.
// This is a convenience wrapper.
func ExecuteTool[In, Out any](
ctx context.Context,
g *genkit.Genkit,
toolName string,
input In,
) (Out, error) {
var zeroOut Out
if g == nil {
return zeroOut, errors.New("genkit instance is nil")
}
if toolName == "" {
return zeroOut, errors.New("tool name cannot be empty for execution")
}

tool := LookupTool(g, toolName)
if tool == nil {
return zeroOut, errors.NewToolNotFoundError(toolName, nil)
}

// ai.Tool has a Run method, but it's not generic. It's RunRaw(ctx context.Context, input any) (any, error)
outputRaw, err := tool.RunRaw(ctx, input)
if err != nil {
return zeroOut, errors.Wrapf(err, "tool '%s' execution failed", toolName)
}

output, ok := outputRaw.(Out)
if !ok {
// If outputRaw is nil and Out is a pointer type or interface, this might be a valid scenario.
// However, if Out is a non-pointer struct, and outputRaw is nil, this assertion fails.
// Or, the types simply mismatch.
typeErr := errors.Errorf("tool '%s' executed, but output type assertion to %T failed (actual type: %T)", toolName, zeroOut, outputRaw)
return zeroOut, errors.WithCode(typeErr, "TYPE_ASSERTION_FAILED")
}

return output, nil
}
