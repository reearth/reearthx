package event

import "context"

// Publisher defines the interface for publishing domain events
type Publisher interface {
	Publish(ctx context.Context, events ...Event) error
}

// Handler defines the interface for handling domain events
type Handler interface {
	Handle(ctx context.Context, event Event) error
}

// Subscriber defines the interface for subscribing to domain events
type Subscriber interface {
	Subscribe(eventType string, handler Handler) error
	Unsubscribe(eventType string, handler Handler) error
}

// EventBus combines Publisher and Subscriber interfaces
type EventBus interface {
	Publisher
	Subscriber
}
