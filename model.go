package theta

import (
	"context"
	"encoding/json"
	"time"

	"gopkg.in/go-playground/validator.v9"
)

var (
	validation = validator.New()
)

//go:generate counterfeiter -fake-name EventHandler -o ./fake/event_handler.go . EventHandler

// EventHandler represents the event's handler
type EventHandler interface {
	// HandleContext handles event
	HandleContext(ctx context.Context, args *EventArgs) error
}

// CompositeEventHandler composes the event handler
type CompositeEventHandler []EventHandler

// HandleContext handles event
func (h CompositeEventHandler) HandleContext(ctx context.Context, args *EventArgs) error {
	for _, handler := range h {
		if err := handler.HandleContext(ctx, args); err != nil {
			return err
		}
	}
	return nil
}

// EventArgs is the actual event
type EventArgs struct {
	Event *Event          `json:"event" validate:"required"`
	Meta  Metadata        `json:"meta" validate:"-"`
	Body  json.RawMessage `json:"body" validate:"required"`
}

// Event represents a event's information
type Event struct {
	ID        string    `json:"id" validate:"required"`
	Name      string    `json:"name" validate:"required"`
	Source    string    `json:"source" validate:"required"`
	Sender    string    `json:"sender" validate:"required"`
	Timestamp time.Time `json:"timestamp" validate:"required"`
}

// Metadata information
type Metadata map[string]string

// Get returns a value for given key
func (m Metadata) Get(key string) string {
	if value, ok := m[key]; ok {
		return value
	}

	return ""
}
