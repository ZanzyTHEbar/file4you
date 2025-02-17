package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/teilomillet/gollm"
)

// Orchestrator defines the interface for processing file metadata.
type Orchestrator struct {
	LLMClient *Client
	// MOAAgents can be a slice of different gollm.LLM instances.
	MOAAgents []Client
}

// NewOrchestrator creates a new Orchestrator.
func NewOrchestrator(client *Client, moaAgents []Client) *Orchestrator {
	return &Orchestrator{
		LLMClient: client,
		MOAAgents: moaAgents,
	}
}

// ProcessFile uses the LLMClient and MOA module to process file metadata.
// It returns the DestinationDecision.
func (o *Orchestrator) ProcessFile(ctx context.Context, filename, fileType, content, modDate string) (DestinationDecision, error) {
	// First, generate a decision using the primary LLM client.
	decision, err := o.LLMClient.GenerateDecision(ctx, filename, fileType, content, modDate)
	if err == nil {
		return decision, nil
	}

	// If the primary decision fails, fall back to MOA.
	// Build the prompt as in the
	// (For brevity, reusing the same method; in practice you might adjust parameters.)
	// For MOA, gather the underlying gollm.LLM instances from the MOAAgents.
	var agents []gollm.LLM
	for _, c := range o.MOAAgents {
		agents = append(agents, c.LLM)
	}

	// Construct the prompt as done in the
	// (You could extract this to a shared utility.)
	// Here we simply assume the prompt built by the
	primaryDecision, err := o.LLMClient.GenerateDecision(ctx, filename, fileType, content, modDate)
	if err != nil {
		return DestinationDecision{}, fmt.Errorf("primary LLM call failed: %v", err)
	}

	// Use the MOA module for parallel processing.
	resp, err := GenerateWithMOA(ctx, primaryDecision.DestinationFolder+":"+primaryDecision.NewFileName, agents)
	if err != nil {
		return DestinationDecision{}, fmt.Errorf("MOA call failed: %v", err)
	}

	var finalDecision DestinationDecision
	if err := json.Unmarshal([]byte(resp), &finalDecision); err != nil {
		return DestinationDecision{}, fmt.Errorf("failed to parse MOA response: %v", err)
	}
	return finalDecision, nil
}

// ProcessFileWithRetry adds retry logic with exponential backoff
func (o *Orchestrator) ProcessFileWithRetry(ctx context.Context, filename, fileType, content, modDate string) (DestinationDecision, error) {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 2 * time.Minute

	var decision DestinationDecision
	operation := func() error {
		var err error
		decision, err = o.ProcessFile(ctx, filename, fileType, content, modDate)
		if err != nil {
			return fmt.Errorf("processing failed: %w", err)
		}
		return nil
	}

	err := backoff.Retry(operation, b)
	return decision, err
}
