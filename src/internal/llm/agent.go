package llm

import (
	"context"

	"go.uber.org/zap"
)

// Agent defines the interface that every expert agent must implement.
type Agent interface {
	// Process takes an input string and returns an output string or an error.
	Process(ctx context.Context, input string) (string, error)
}

// SimpleAgent is a basic agent that uses the LLM client to process inputs.
type SimpleAgent struct {
	llmClient *LLMClient
	name      string
	logger    *zap.SugaredLogger
}

// NewSimpleAgent creates a new SimpleAgent.
func NewSimpleAgent(name string, client *LLMClient, logger *zap.SugaredLogger) *SimpleAgent {
	return &SimpleAgent{
		name:      name,
		llmClient: client,
		logger:    logger,
	}
}

func (a *SimpleAgent) Name() string {
	return a.name
}

// Process calls the LLM client to generate a response for the given input.
func (a *SimpleAgent) Process(ctx context.Context, input string) (string, error) {
	a.logger.Infof("%s processing input: %s", a.name, input)
	response, err := a.llmClient.GenerateText(ctx, input)
	if err != nil {
		a.logger.Errorf("%s encountered error: %v", a.name, err)
		return "", err
	}
	a.logger.Infof("%s produced response: %s", a.name, response)
	return response, nil
}
