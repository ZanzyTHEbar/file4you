package llm

import (
	"context"
	"time"
)

type Logger struct {
	logChan chan LogEntry
}

type LogEntry struct {
	Timestamp time.Time
	Operation string
	Input     string
	Output    DestinationDecision
	Duration  time.Duration
	Error     string
}

func (l *Logger) Log(ctx context.Context, entry LogEntry) {
	select {
	case l.logChan <- entry:
	case <-ctx.Done():
		// Handle context cancellation
	default:
		// Handle channel full
	}
}
