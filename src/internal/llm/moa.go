package llm

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/teilomillet/gollm"
)

// GenerateWithMOA concurrently queries multiple LLM agents and aggregates a response.
// It returns the first valid non-empty response, or an error if all fail.
func GenerateWithMOA(ctx context.Context, prompt string, agents []gollm.LLM) (string, error) {
	type result struct {
		resp string
		err  error
	}
	resCh := make(chan result, len(agents))
	var wg sync.WaitGroup

	for _, agent := range agents {
		wg.Add(1)
		go func(llmAgent gollm.LLM) {
			defer wg.Done()
			// Use a timeout for each agent call.
			callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			resp, err := llmAgent.Generate(callCtx, gollm.NewPrompt(prompt))
			// Handle empty responses as errors.
			if err == nil && strings.TrimSpace(resp) == "" {
				err = fmt.Errorf("empty response")
			}
			resCh <- result{resp: resp, err: err}
		}(agent)
	}

	// Close the channel when all goroutines are done.
	go func() {
		wg.Wait()
		close(resCh)
	}()

	// Aggregate responses: return the first successful response.
	for res := range resCh {
		if res.err == nil {
			return res.resp, nil
		}
	}
	return "", fmt.Errorf("all agent calls failed")
}
