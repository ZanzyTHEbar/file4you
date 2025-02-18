// pkg/llm/llm.go
package llm

import (
	"context"
	"time"

	"github.com/ZanzyTHEbar/assert-lib"
	"github.com/teilomillet/gollm"
	"go.uber.org/zap"
)

type LLMInterface interface {
	GenerateText(ctx context.Context, prompt string) (string, error)
}

type LLMClient struct {
	LLM           gollm.LLM
	logger        *zap.SugaredLogger
	assertHandler *assert.AssertHandler
}

func NewClient(ctx context.Context, logger *zap.SugaredLogger, assertHandler *assert.AssertHandler) (*LLMClient, error) {
	// Initialize gollm with the provided configuration.
	llmClient, err := gollm.NewLLM(
		gollm.SetProvider(cfg.LLMProvider),
		gollm.SetModel(cfg.LLMModel),
		gollm.SetAPIKey(cfg.APIKey),
		gollm.SetMaxTokens(cfg.MaxTokens),
		gollm.SetTimeout(time.Duration(cfg.TimeoutSec)*time.Second),
		gollm.SetEnableCaching(true),
		gollm.SetMemory(4096),
		gollm.SetRetryDelay(5*time.Second),
		gollm.SetMaxRetries(3),
		gollm.SetLogLevel(gollm.LogLevelDebug),
	)

	assertHandler.NotNil(ctx, err, "Failed to create LLM client")

	return &LLMClient{
		LLM:           llmClient,
		logger:        logger,
		assertHandler: assertHandler,
	}, nil
}

// GenerateText generates text based on a given prompt.
func (c *LLMClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	p := gollm.NewPrompt(prompt)
	response, err := c.LLM.Generate(ctx, p)
	if err != nil {
		c.logger.Errorf("LLM Generate error: %v", err)
		return "", err
	}
	return response, nil
}
