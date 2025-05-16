// Package genkithandler provides a simplified interface for integrating with Genkit.
package genkithandler

import (
	"context"
	"fmt"

	"file4you/internal/genkithandler/errors"

	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

// DefineFlow defines a new Genkit flow and registers it with the provided Genkit instance.
// It's a wrapper around genkit.DefineFlow.
func DefineFlow[In, Out any](g *genkit.Genkit, name string, fn core.Func[In, Out]) (*core.Flow[In, Out, struct{}], error) {
	if g == nil {
		return nil, errors.New("genkit instance is nil")
	}
	if name == "" {
		return nil, errors.New("flow name cannot be empty")
	}
	if fn == nil {
		return nil, errors.New("flow function (fn) cannot be nil")
	}

	flow := genkit.DefineFlow(g, name, fn)
	if flow == nil {
		// This condition might be hard to hit if DefineFlow panics or always returns non-nil.
		// Depending on genkit's behavior, error handling might need adjustment.
		return nil, errors.Errorf("failed to define flow %s", name)
	}
	return flow, nil
}

// DefineStreamingFlow defines a new Genkit streaming flow.
func DefineStreamingFlow[In, Out, Stream any](g *genkit.Genkit, name string, fn core.StreamingFunc[In, Out, Stream]) (*core.Flow[In, Out, Stream], error) {
	if g == nil {
		return nil, errors.New("genkit instance is nil")
	}
	if name == "" {
		return nil, errors.New("streaming flow name cannot be empty")
	}
	if fn == nil {
		return nil, errors.New("streaming flow function (fn) cannot be nil")
	}
	flow := genkit.DefineStreamingFlow(g, name, fn)
	if flow == nil {
		return nil, errors.Errorf("failed to define streaming flow %s", name)
	}
	return flow, nil
}

// ExecuteFlow executes a previously defined Genkit flow.
// It looks up the flow by name from the provided Genkit instance and runs it.
func ExecuteFlow[In, Out any](ctx context.Context, g *genkit.Genkit, flowName string, input In) (Out, error) {
	var zeroOut Out
	if g == nil {
		return zeroOut, errors.New("genkit instance is nil")
	}
	if flowName == "" {
		return zeroOut, errors.New("flow name cannot be empty for execution")
	}

	var targetAction core.Action
	flows := genkit.ListFlows(g)
	for _, f := range flows {
		if f.Name() == flowName {
			targetAction = f
			break
		}
	}

	if targetAction == nil {
		return zeroOut, errors.NewFlowNotFoundError(flowName, nil)
	}

	// DEBUG: Print the type of the targetAction for troubleshooting
	fmt.Printf("DEBUG: Found flow '%s' with Go type: %T\n", flowName, targetAction)

	typedFlow, ok := targetAction.(*core.ActionDef[In, Out, struct{}])
	if !ok {
		err := errors.Errorf("flow '%s' found, but it is not a non-streaming flow with the expected input/output types, or type assertion failed (actual type: %T)", flowName, targetAction)
		return zeroOut, errors.WithCode(err, "TYPE_ASSERTION_FAILED")
	}

	output, runErr := typedFlow.Run(ctx, input, nil)
	if runErr != nil {
		return zeroOut, errors.Wrapf(runErr, "error running flow '%s'", flowName)
	}
	return output, nil
}

// ExecuteStreamingFlow executes a previously defined Genkit streaming flow.
// It looks up the flow by name from the provided Genkit instance,
// starts streaming it, and calls the userCallback with each streamed chunk.
// It returns the final output of the flow.
func ExecuteStreamingFlow[In, Out, StreamChunk any](
	ctx context.Context,
	g *genkit.Genkit,
	flowName string,
	input In,
	userCallback core.StreamCallback[StreamChunk],
) (Out, error) {
	var zeroOut Out
	if g == nil {
		return zeroOut, errors.New("genkit instance is nil")
	}
	if flowName == "" {
		return zeroOut, errors.New("streaming flow name cannot be empty for execution")
	}

	var targetAction core.Action
	flows := genkit.ListFlows(g)
	for _, f := range flows {
		if f.Name() == flowName {
			targetAction = f
			break
		}
	}

	if targetAction == nil {
		return zeroOut, errors.NewFlowNotFoundError(flowName, nil)
	}

	typedFlow, ok := targetAction.(*core.Flow[In, Out, StreamChunk])
	if !ok {
		err := errors.Errorf("flow '%s' found, but it is not a streaming flow with the expected input/output/stream types, or type assertion failed", flowName)
		return zeroOut, errors.WithCode(err, "TYPE_ASSERTION_FAILED")
	}

	// Use the .Stream() method to get a channel of results
	// Assuming the channel yields *core.StreamingFlowResult[Out, StreamChunk]
	// based on typical Genkit patterns and examples.
	// Corrected based on compiler error: Stream() returns only one value (the channel).
	streamCh := typedFlow.Stream(ctx, input)
	// Removed error check for typedFlow.Stream() as it only returns the channel.

	var finalOutput Out // To store the 'Out' when result.Done is true

	for result := range streamCh { // result is of type *core.StreamingFlowValue[Out, StreamChunk]
		// Assuming core.StreamingFlowValue has fields: Done (bool), Output (Out), Stream (StreamChunk)
		// Based on compiler feedback, result.Err is not available.

		if !result.Done {
			// This is a stream chunk of type StreamChunk
			if userCallback != nil {
				// Call the user-provided callback with the chunk
				if cbErr := userCallback(ctx, result.Stream); cbErr != nil {
					// User callback indicated an error, stop processing.
					// It's important to decide if the flow's finalOutput (if any was received before this point)
					// should be returned or if the callback error takes precedence.
					// For now, callback error stops everything and returns zeroOut for Out.
					return zeroOut, errors.Wrapf(cbErr, "user callback for streaming flow '%s' failed", flowName)
				}
			}
		} else {
			// The flow has completed. result.Output contains the final 'Out'.
			finalOutput = result.Output
			break // Exit loop as flow is done
		}
	}
	// The loop finishes when the channel is closed or explicitly broken out of.
	// If the stream terminated due to an internal flow error not caught by the userCallback,
	// finalOutput might be its zero value or incomplete. This function cannot distinguish
	// that from a flow that legitimately finishes with a zero/empty Out value without an Err field on result.
	return finalOutput, nil
}

/*
RegisterCoreFlows registers core flows (greetingFlow, backupFlow) using the new Genkit API.
Call this during Genkit initialization.
*/
func RegisterCoreFlows(g *genkit.Genkit) error {
	// Greeting flow
	greetingHandler := func(ctx context.Context, name string) (string, error) {
		return "Hello, " + name, nil
	}
	if _, err := DefineFlow[string, string](g, "greetingFlow", greetingHandler); err != nil {
		return fmt.Errorf("failed to register greetingFlow: %w", err)
	}

	// Backup flow
	backupHandler := func(ctx context.Context, input BackupToolInput) (*BackupToolOutput, error) {
		// This should call the performBackup tool using ExecuteTool
		output, toolErr := ExecuteTool[BackupToolInput, *BackupToolOutput](ctx, g, "performBackup", input)
		if toolErr != nil {
			return nil, fmt.Errorf("error executing performBackup tool in backupFlow: %w", toolErr)
		}
		if output == nil {
			return nil, fmt.Errorf("performBackup tool returned nil output")
		}
		return output, nil
	}
	if _, err := DefineFlow(g, "backupFlow", backupHandler); err != nil {
		return fmt.Errorf("failed to register backupFlow: %w", err)
	}

return nil
}

// The `DefineFlow` and `DefineStreamingFlow` functions return the created flow
// as per Genkit's pattern, allowing users to also call .Run() directly on the flow object.
// The `ExecuteFlow` and `ExecuteStreamingFlow` functions are convenience wrappers for running by name.
