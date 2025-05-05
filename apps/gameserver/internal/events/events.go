package events

import (
	"sync"
)

// EventType represents the type of event
type EventType string

const (
	RoomCreated EventType = "room.created"
	RoomRemoved EventType = "room.removed"
	RoomUpdated EventType = "room.updated"
)

// Event contains information about an event
type Event struct {
	Type     EventType
	RoomID   string
	GameType string // maybe not needed
	Data     interface{}
}

// HandlerFunc defines an event handler function
type HandlerFunc func(event Event)

// EventBus implements a simple pub-sub event system
type EventBus struct {
	subscribers map[EventType][]HandlerFunc
	mu          sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[EventType][]HandlerFunc),
	}
}

// Subscribe registers a handler for a specific event type
func (b *EventBus) Subscribe(eventType EventType, handler HandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[eventType] = append(b.subscribers[eventType], handler)
}

// Publish sends an event to all subscribers of that event type
func (b *EventBus) Publish(event Event) {
	b.mu.RLock()
	handlers, exists := b.subscribers[event.Type]
	b.mu.RUnlock()

	if !exists {
		return
	}

	for _, handler := range handlers {
		go handler(event)
	}
}
