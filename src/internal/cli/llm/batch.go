package llm

import (
	"context"
	"file4you/internal/llm"
	"sync"
)

// BatchProcessor handles concurrent processing of files
type BatchProcessor struct {
	orchestrator *llm.Orchestrator
	batchSize    int
}

func NewBatchProcessor(orchestrator *llm.Orchestrator, batchSize int) *BatchProcessor {
	return &BatchProcessor{
		orchestrator: orchestrator,
		batchSize:    batchSize,
	}
}

// ProcessFiles processes files in batches
func (bp *BatchProcessor) ProcessFiles(ctx context.Context, files []FileInfo) []llm.DestinationDecision {
	var (
		decisions []llm.DestinationDecision
		mu        sync.Mutex
		wg        sync.WaitGroup
		sem       = make(chan struct{}, bp.batchSize)
	)

	for _, file := range files {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(f FileInfo) {
			defer func() {
				<-sem // Release semaphore
				wg.Done()
			}()

			decision, err := bp.orchestrator.ProcessFile(ctx,
				f.Name(),
				f.Type(),
				f.Preview(),
				f.ModTime().Format("2006-01-02"))

			if err != nil {
				return
			}

			mu.Lock()
			decisions = append(decisions, decision)
			mu.Unlock()
		}(file)
	}

	wg.Wait()
	return decisions
}
