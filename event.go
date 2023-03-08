package theta

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/go-playground/validator.v9"
)

//go:generate counterfeiter -generate

var validation = validator.New()

//counterfeiter:generate -o ./fake/event_handler.go . EventHandler

// EventHandler represents the event's handler
type EventHandler interface {
	// HandleContext handles event
	HandleContext(ctx context.Context, args *EventArgs) error
}

var _ EventHandler = CompositeEventHandler{}

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

var _ EventHandler = DictionaryEventHandler{}

// DictionaryEventHandler represents a dictionary of event handler
type DictionaryEventHandler map[string]EventHandler

// HandleContext handles event
func (h DictionaryEventHandler) HandleContext(ctx context.Context, args *EventArgs) error {
	handler, ok := h[args.Event.Name]

	if !ok {
		return fmt.Errorf("event handler %v not found", args.Event.Name)
	}

	return handler.HandleContext(ctx, args)
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

//counterfeiter:generate -o ./fake/event_decoder.go . EventDecoder

// EventDecoder represents an event decoder
type EventDecoder interface {
	Decode([]byte, interface{}) error
}

//counterfeiter:generate -o ./fake/event_encoder.go . EventEncoder

// EventEncoder represents an event decoder
type EventEncoder interface {
	Encode(interface{}) ([]byte, error)
}
