package socket

import (
	"github.com/atendi9/ws/io/pkg/events"
)

// StrictEmitter strictly typed version of an [events.Emitter]. A `TypedEmitter` takes type
// parameters for mappings of event names to event data types, and strictly
// types method calls to the [events.Emitter] according to these event maps.
type StrictEmitter struct {
	events.Emitter
}

func NewStrictEmitter() *StrictEmitter {
	return &StrictEmitter{Emitter: events.NewEmitter()}
}

// On adds the listener function as an event listener for the given event name.
func (s *StrictEmitter) On(ev string, listeners ...events.Listener) error {
	return s.Emitter.On(events.Name(ev), listeners...)
}

// Once adds a one-time listener function for the given event name.
func (s *StrictEmitter) Once(ev string, listeners ...events.Listener) error {
	return s.Emitter.Once(events.Name(ev), listeners...)
}

// Emit emits an event with the specified name and arguments to all listeners.
func (s *StrictEmitter) Emit(ev string, args ...any) {
	s.Emitter.Emit(events.Name(ev), args...)
}

// EmitReserved emits a reserved event. Only subclasses should use this method.
func (s *StrictEmitter) EmitReserved(ev string, args ...any) {
	s.Emitter.Emit(events.Name(ev), args...)
}

// EmitUntyped emits an event without strict type checking. Only subclasses should use this method.
func (s *StrictEmitter) EmitUntyped(ev string, args ...any) {
	s.Emitter.Emit(events.Name(ev), args...)
}

// Listeners returns a slice of listeners subscribed to the given event name.
func (s *StrictEmitter) Listeners(ev string) []events.Listener {
	return s.Emitter.Listeners(events.Name(ev))
}
