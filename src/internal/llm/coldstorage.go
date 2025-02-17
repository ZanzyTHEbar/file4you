package llm

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// ColdStorageMemory implements Memory and additionally persists older messages to SQLite.
type ColdStorageMemory struct {
	buffer      Memory  // underlying in-memory buffer (e.g., BufferMemory)
	db          *sql.DB // database handle for persistent storage
	flushPeriod time.Duration
	lastFlush   time.Time
	mu          sync.Mutex
	// Optionally, you can embed observers if needed.
}

func NewColdStorageMemory(buffer Memory, db *sql.DB, flushPeriod time.Duration) *ColdStorageMemory {
	csm := &ColdStorageMemory{
		buffer:      buffer,
		db:          db,
		flushPeriod: flushPeriod,
		lastFlush:   time.Now(),
	}
	// Start background flush routine.
	go csm.backgroundFlush()
	return csm
}

// backgroundFlush periodically checks if older messages should be flushed to cold storage.
func (csm *ColdStorageMemory) backgroundFlush() {
	ticker := time.NewTicker(csm.flushPeriod)
	defer ticker.Stop()
	for {
		<-ticker.C
		csm.mu.Lock()
		if time.Since(csm.lastFlush) >= csm.flushPeriod {
			if err := csm.flushToDB(); err != nil {
				// Log error (use proper logging in production)
				fmt.Printf("Error flushing to DB: %v\n", err)
			} else {
				csm.lastFlush = time.Now()
			}
		}
		csm.mu.Unlock()
	}
}

// flushToDB persists messages older than a threshold from the buffer to the database.
func (csm *ColdStorageMemory) flushToDB() error {
	// Retrieve current context.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mem, err := csm.buffer.GetContext()
	if err != nil {
		return err
	}
	// Here, we assume a simple table "memories" with columns "id", "content", and "timestamp".
	// You would use your existing SQLite DB implementation.
	query := `INSERT INTO memories (content, timestamp) VALUES (?, ?)`
	_, err = csm.db.ExecContext(ctx, query, mem, time.Now())
	if err != nil {
		return fmt.Errorf("failed to flush memory to DB: %w", err)
	}
	// Optionally, clear the buffer after flush or remove only the flushed parts.
	// For demonstration, we do nothing.
	return nil
}

// Implement Memory interface by delegating to the underlying buffer.
func (csm *ColdStorageMemory) AddMessage(message string) error {
	return csm.buffer.AddMessage(message)
}

func (csm *ColdStorageMemory) GetContext() (string, error) {
	return csm.buffer.GetContext()
}

func (csm *ColdStorageMemory) RemoveMessage(index int) error {
	return csm.buffer.RemoveMessage(index)
}

func (csm *ColdStorageMemory) EditMessage(index int, newMessage string) error {
	return csm.buffer.EditMessage(index, newMessage)
}

func (csm *ColdStorageMemory) RegisterObserver(o Observer) {
	csm.buffer.RegisterObserver(o)
}

// Additional methods for user control:

// DeleteMemoryEntry removes a memory entry from cold storage.
func (csm *ColdStorageMemory) DeleteMemoryEntry(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := csm.db.ExecContext(ctx, "DELETE FROM memories WHERE id = ?", id)
	return err
}

// EditMemoryEntry updates a memory entry in cold storage.
func (csm *ColdStorageMemory) EditMemoryEntry(id int, newContent string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := csm.db.ExecContext(ctx, "UPDATE memories SET content = ? WHERE id = ?", newContent, id)
	return err
}

// GetColdMemories retrieves stored memories.
func (csm *ColdStorageMemory) GetColdMemories() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := csm.db.QueryContext(ctx, "SELECT content FROM memories ORDER BY timestamp ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []string
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			return nil, err
		}
		memories = append(memories, content)
	}
	return memories, nil
}
