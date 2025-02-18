package llm

import (
	"context"

	"github.com/sourcegraph/conc"
	"go.uber.org/zap"
)

// RequestMessage represents a request to process input.
type RequestMessage struct {
	ID      string
	Payload string
}

// ResponseMessage represents a response from an agent.
type ResponseMessage struct {
	RequestID string
	AgentName string
	Output    string
	Err       error
}

// Router dispatches requests to a list of agents and forwards their responses.
type Router struct {
	requestChan  chan RequestMessage
	responseChan chan ResponseMessage
	agents       []Agent
	logger       *zap.SugaredLogger
}

// NewRouter creates a new Router instance.
func NewRouter(reqChan chan RequestMessage, resChan chan ResponseMessage, agents []Agent, logger *zap.SugaredLogger) *Router {
	return &Router{
		requestChan:  reqChan,
		responseChan: resChan,
		agents:       agents,
		logger:       logger,
	}
}

// Run listens for incoming requests and dispatches them to all agents concurrently using conc.WaitGroup.
func (r *Router) Run() {
	for req := range r.requestChan {
		r.logger.Infof("Router received request: %s", req.ID)
		var g conc.WaitGroup
		// Dispatch the request to all agents concurrently.
		for _, agent := range r.agents {
			agent := agent // capture loop variable
			g.Go(func() {
				// TODO: Create a fresh background context (can be extended to pass proper context)
				ctx := context.Background()
				output, err := agent.Process(ctx, req.Payload)
				res := ResponseMessage{
					RequestID: req.ID,
					AgentName: getAgentName(agent),
					Output:    output,
					Err:       err,
				}
				// Send the result via the response channel.
				r.responseChan <- res
			})
		}
		// Wait for all agent processing to finish.
		g.Wait()
	}
}

// getAgentName is a helper to extract the agent's name.
// For our SimpleAgent, we can type assert.
func getAgentName(a Agent) string {
	if sa, ok := a.(*SimpleAgent); ok {
		return sa.Name()
	}
	return "unknown"
}
