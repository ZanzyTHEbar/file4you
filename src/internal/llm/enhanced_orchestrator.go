package llm

import (
    "context"
    "fmt"
    "sync"
    "time"

    "golang.org/x/time/rate"
    lru "github.com/hashicorp/golang-lru/v2"
)

type EnhancedOrchestrator struct {
    *Orchestrator
    rateLimiter *rate.Limiter
    cache       *lru.Cache[string, DestinationDecision]
    metrics     *Metrics
    eventBus    *EventBus
    txManager   *TransactionManager
}

type Metrics struct {
    ProcessedFiles    int64
    CacheHits        int64
    CacheMisses      int64
    ProcessingErrors int64
    AvgProcessingTime float64
}

type EventBus struct {
    subscribers map[string][]chan Event
    mu          sync.RWMutex
}

type Event struct {
    Type    string
    Payload interface{}
    Time    time.Time
}

type TransactionManager struct {
    mu         sync.Mutex
    operations []Operation
}

type Operation struct {
    File     FileData
    Decision DestinationDecision
}

func NewEnhancedOrchestrator(orchestrator *Orchestrator, config *Config) (*EnhancedOrchestrator, error) {
    cache, err := lru.New[string, DestinationDecision](config.CacheSize)
    if err != nil {
        return nil, fmt.Errorf("failed to create cache: %w", err)
    }

    return &EnhancedOrchestrator{
        Orchestrator: orchestrator,
        rateLimiter: rate.NewLimiter(rate.Limit(config.RateLimit), config.BatchSize),
        cache:       cache,
        metrics:     &Metrics{},
        eventBus:    newEventBus(),
        txManager:   newTransactionManager(),
    }, nil
}

// ProcessBatch processes multiple files with rate limiting and transaction support
func (eo *EnhancedOrchestrator) ProcessBatch(ctx context.Context, batch *Batch) ([]DestinationDecision, error) {
    if err := eo.rateLimiter.Wait(ctx); err != nil {
        return nil, fmt.Errorf("rate limit exceeded: %w", err)
    }

    tx := eo.txManager.Begin()
    defer tx.Rollback()

    results := make([]DestinationDecision, 0, len(batch.Files))
    start := time.Now()

    for _, file := range batch.Files {
        decision, err := eo.processFile(ctx, file)
        if err != nil {
            return nil, err
        }
        results = append(results, decision)
        tx.AddOperation(Operation{File: file, Decision: decision})
    }

    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    eo.updateMetrics(len(batch.Files), time.Since(start))
    eo.emitEvent("BatchProcessed", results)

    return results, nil
}

func (eo *EnhancedOrchestrator) processFile(ctx context.Context, file FileData) (DestinationDecision, error) {
    cacheKey := fmt.Sprintf("%s-%s-%s", file.Name, file.Type, file.ModTime)
    
    // Check cache
    if decision, ok := eo.cache.Get(cacheKey); ok {
        atomic.AddInt64(&eo.metrics.CacheHits, 1)
        return decision, nil
    }
    atomic.AddInt64(&eo.metrics.CacheMisses, 1)

    // Process file
    decision, err := eo.Orchestrator.ProcessFile(ctx, file.Name, file.Type, file.Content, file.ModTime)
    if err != nil {
        atomic.AddInt64(&eo.metrics.ProcessingErrors, 1)
        return DestinationDecision{}, err
    }

    // Cache result
    eo.cache.Add(cacheKey, decision)
    return decision, nil
}

func (eo *EnhancedOrchestrator) updateMetrics(processedFiles int, duration time.Duration) {
    atomic.AddInt64(&eo.metrics.ProcessedFiles, int64(processedFiles))
    // Update average processing time using weighted average
    current := atomic.LoadInt64(&eo.metrics.ProcessedFiles)
    newAvg := (eo.metrics.AvgProcessingTime*float64(current-int64(processedFiles)) + 
        duration.Seconds()*float64(processedFiles)) / float64(current)
    atomic.StoreUint64((*uint64)(&eo.metrics.AvgProcessingTime), math.Float64bits(newAvg))
}

// Event handling methods
func (eo *EnhancedOrchestrator) Subscribe(eventType string, ch chan Event) {
    eo.eventBus.Subscribe(eventType, ch)
}

func (eo *EnhancedOrchestrator) emitEvent(eventType string, payload interface{}) {
    eo.eventBus.Emit(Event{
        Type:    eventType,
        Payload: payload,
        Time:    time.Now(),
    })
}

// Transaction management
func (tm *TransactionManager) Begin() *Transaction {
    return &Transaction{
        manager: tm,
        ops:     make([]Operation, 0),
    }
}

type Transaction struct {
    manager *TransactionManager
    ops     []Operation
}

func (t *Transaction) AddOperation(op Operation) {
    t.ops = append(t.ops, op)
}

func (t *Transaction) Commit() error {
    t.manager.mu.Lock()
    defer t.manager.mu.Unlock()
    
    t.manager.operations = append(t.manager.operations, t.ops...)
    return nil
}

func (t *Transaction) Rollback() {
    t.ops = nil
}

// Event bus implementation
func newEventBus() *EventBus {
    return &EventBus{
        subscribers: make(map[string][]chan Event),
    }
}

func (eb *EventBus) Subscribe(eventType string, ch chan Event) {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
}

func (eb *EventBus) Emit(event Event) {
    eb.mu.RLock()
    defer eb.mu.RUnlock()
    
    for _, ch := range eb.subscribers[event.Type] {
        select {
        case ch <- event:
        default:
            // Non-blocking send
        }
    }
}