package llm

import (
	"sync"
	"time"
)

type Metrics struct {
	mu              sync.RWMutex
	successfulCalls int64
	failedCalls     int64
	latencies       []time.Duration
}

func (m *Metrics) RecordSuccess(latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.successfulCalls++
	m.latencies = append(m.latencies, latency)
}
