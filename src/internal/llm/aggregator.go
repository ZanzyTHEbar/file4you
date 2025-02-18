package llm

import (
	"go.uber.org/zap"
)

// Aggregator collects responses from agents and aggregates them.
type Aggregator struct {
	responseChan chan ResponseMessage
	logger       *zap.SugaredLogger
}

// NewAggregator creates a new Aggregator.
func NewAggregator(resChan chan ResponseMessage, logger *zap.SugaredLogger) *Aggregator {
	return &Aggregator{
		responseChan: resChan,
		logger:       logger,
	}
}

// Run listens on the response channel and aggregates responses.
// For this basic example, it logs each response. In future, ensemble logic can be added.
func (agg *Aggregator) Run() {
	for res := range agg.responseChan {
		if res.Err != nil {
			agg.logger.Errorf("Aggregator received error from %s: %v", res.AgentName, res.Err)
		} else {
			agg.logger.Infof("Aggregator received response from %s: %s", res.AgentName, res.Output)
		}
		// Aggregation logic (e.g., weighted voting, consensus) can be implemented here.
	}
}
