package event

import (
	"context"
	"time"
)

// EventMetadata contains metadata for domain events
type EventMetadata struct {
	Version     int
	Timestamp   time.Time
	UserID      string
	AggregateID string
}

// EventEnvelope wraps an event with its metadata
type EventEnvelope struct {
	Event    Event
	Metadata EventMetadata
}

// EventStore defines the interface for storing and retrieving domain events
type EventStore interface {
	// Save stores events for an aggregate
	Save(ctx context.Context, aggregateID string, events ...EventEnvelope) error

	// Load retrieves all events for an aggregate
	Load(ctx context.Context, aggregateID string) ([]EventEnvelope, error)

	// LoadByType retrieves events of a specific type
	LoadByType(ctx context.Context, eventType string) ([]EventEnvelope, error)

	// LoadByTimeRange retrieves events within a time range
	LoadByTimeRange(ctx context.Context, start, end time.Time) ([]EventEnvelope, error)
}

// EventManager combines Publisher, Subscriber and EventStore interfaces
type EventManager interface {
	Publisher
	Subscriber
	EventStore
}

// NewEventEnvelope creates a new event envelope with metadata
func NewEventEnvelope(event Event, version int, userID string, aggregateID string) EventEnvelope {
	return EventEnvelope{
		Event: event,
		Metadata: EventMetadata{
			Version:     version,
			Timestamp:   time.Now(),
			UserID:      userID,
			AggregateID: aggregateID,
		},
	}
}
