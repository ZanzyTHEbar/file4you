package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/teilomillet/gollm"
	"github.com/cenkalti/backoff/v4"
)

// DestinationDecision holds the structured decision.
type DestinationDecision struct {
	DestinationFolder string `json:"destinationFolder"`
	NewFileName       string `json:"newFileName"`
}

// Client wraps a gollm LLM instance and conversation memory.
type Client struct {
	LLM    gollm.LLM
	Memory Memory
}

// NewClient creates a new Client with the provided gollm options.
func NewClient(opts ...gollm.ConfigOption) (*Client, error) {
	llm, err := gollm.NewLLM(opts...)
	if err != nil {
		return nil, err
	}
	// Set up memory with a max of 10 messages.
	mem := NewBufferMemory(10)
	return &Client{
		LLM:    llm,
		Memory: mem,
	}, nil
}

// GenerateDecision builds a prompt (using memory context) and returns a DestinationDecision.
func (c *Client) GenerateDecision(ctx context.Context, filename, fileType, content string, modDateStr string) (DestinationDecision, error) {
	// Get memory context.
	memCtx, err := c.Memory.GetContext()
	if err != nil {
		return DestinationDecision{}, err
	}

	fileData := FileData{
		Filename:       filename,
		FileType:       fileType,
		ModDate:        modDateStr,
		ContentPreview: Truncate(content, 200),
	}

	basePrompt, err := BuildPrompt(fileData)
	if err != nil {
		return DestinationDecision{}, err
	}

	// Append memory context if available.
	if memCtx != "" {
		basePrompt += "\nContext:\n" + memCtx
	}

	// Generate a response using the LLM instance.
	resp, err := c.LLM.Generate(ctx, gollm.NewPrompt(basePrompt))
	if err != nil {
		return DestinationDecision{}, err
	}

	// Update memory with the prompt and response.
	c.Memory.AddMessage("Prompt: " + basePrompt)
	c.Memory.AddMessage("Response: " + resp)

	// Parse the JSON response.
	var decision DestinationDecision
	if err := json.Unmarshal([]byte(resp), &decision); err != nil {
		return DestinationDecision{}, fmt.Errorf("failed to parse LLM response: %v", err)
	}
	return decision, nil
}
