package llm

import (
	"fmt"
	"strings"
	"sync"
)

// ErrInvalidIndex is returned when an invalid index is provided.
var ErrInvalidIndex = fmt.Errorf("invalid index")

type Memory interface {
	GetContext() (string, error)
	AddMessage(message string)
	RemoveMessage(index int) error
	EditMessage(index int, newMessage string) error
	RegisterObserver(observer Observer)
}

type Observer interface {
	OnMemoryUpdated(newCotext string)
}

// BufferMemory is a simple Memory that stores the last N messages verbatim
type BufferMemory struct {
	messages      []string
	maxTokenLimit int
	observers     []Observer
	mu            sync.Mutex
}

func NewBufferMemory(maxTokenLimit int) *BufferMemory {
	return &BufferMemory{
		messages:      make([]string, 0, maxTokenLimit),
		maxTokenLimit: maxTokenLimit,
		observers:     make([]Observer, 0),
	}
}

// AddMessage adds a message to the memory and summarizes the messages if the memory is full
// Note: Should probably use a ring buffer instead of a slice
func (bm *BufferMemory) AddMessage(message string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.messages = append(bm.messages, message)

	if len(bm.messages) < bm.maxTokenLimit {
		return
	}

	summary := summarizeMessages(bm.messages[:len(bm.messages)-bm.maxTokenLimit])

	// Replace old messages with summary
	bm.messages = append([]string{summary}, bm.messages[len(bm.messages)-bm.maxTokenLimit+1:]...)

	bm.notifyObservers(summary)
}

func (bm *BufferMemory) RemoveMessage(index int) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	if index < 0 || index >= len(bm.messages) {
		return ErrInvalidIndex
	}
	bm.messages = append(bm.messages[:index], bm.messages[index+1:]...)
	return nil
}

func (bm *BufferMemory) EditMessage(index int, newMessage string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	if index < 0 || index >= len(bm.messages) {
		return ErrInvalidIndex
	}
	bm.messages[index] = newMessage
	return nil
}

// summarizeMessages returns a summary of the messages
// Note: This is a very naive implementation
func (bm *BufferMemory) GetContext() (string, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	return strings.Join(bm.messages, "\n"), nil
}

func (bm *BufferMemory) RegisterObserver(observer Observer) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.observers = append(bm.observers, observer)
}

// notifyObservers notifies all observers of the new context
// Note: This is a naive implementation that spawns a new goroutine for each observer
// This could be optimized, but it's fine for now
func (bm *BufferMemory) notifyObservers(newContext string) {
	for _, observer := range bm.observers {
		go observer.OnMemoryUpdated(newContext)
	}
}

// summarizeMessages is a stub that concatenates messages with a simple prefix.
func summarizeMessages(messages []string) string {
	return "Summary: " + strings.Join(messages, " | ")
}
