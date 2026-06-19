package events

import (
	"reflect"
	"sync"

	"github.com/atendi9/ws/io/pkg/anvil"
)

// Name represents a strongly-typed string identifier used for tracking specific events.
type Name string

// Listener defines the functional signature used for generic dynamic callbacks.
type Listener func(...any)

// Events maps an [Name] identifier to a collection of [Listener] callbacks.
type Events map[Name][]Listener

// Emitter defines the interface for subscribing to, removing, and broadcasting events.
type Emitter interface {
	// AddListener appends one or more [Listener] callbacks to the specified [Name].
	AddListener(Name, ...Listener) error
	// Emit triggers all callbacks associated with the given [Name] , supplying optional dynamic payloads.
	Emit(Name, ...any)
	// Names fetches all unique keys registered inside the event tracker.
	Names() []Name
	// ListenerCount evaluates how many handlers are associated with a single [Name].
	ListenerCount(Name) int
	// Listeners returns an array of all active [Listener] callbacks under a single [Name].
	Listeners(Name) []Listener
	// On functions as an alternate alias for the [AddListener] capability.
	On(Name, ...Listener) error
	// Once registers an [Listener] configured to execute explicitly one single time before removal.
	Once(Name, ...Listener) error
	// RemoveAllListeners drops all registrations allocated for a unique [Name].
	RemoveAllListeners(Name) bool
	// RemoveListener clears down an exact [Listener] handler bound to an [Name].
	RemoveListener(Name, Listener) bool
	// Clear flushes out all configurations and event names present within the emitter storage.
	Clear()
	// Len returns the raw total of distinct event domains actively handled.
	Len() int
}

// eventEntry holds the reference wrapper to an [Listener] along with its verified memory pointer.
type eventEntry struct {
	fn  Listener
	ptr uintptr
}

// emitter manages independent thread-safe events using [anvil.Map] and concurrent slice helpers.
type emitter struct {
	listeners anvil.Map[Name, *anvil.Slice[*eventEntry]]
}

// CopyTo copies and injects the current layout directly into a targets [Emitter] framework.
func (e Events) CopyTo(target Emitter) {
	if len(e) > 0 {
		for evt, listeners := range e {
			if len(listeners) > 0 {
				_ = target.AddListener(evt, listeners...)
			}
		}
	}
}

// NewEmitter instantiates a concrete thread-safe [Emitter] construct.
func NewEmitter() Emitter {
	emitter := &emitter{
		listeners: anvil.Map[Name, *anvil.Slice[*eventEntry]]{},
	}

	return emitter
}

// addListeners aggregates structured slices of event entries safely into the concurrent data structures.
func (e *emitter) addListeners(evt Name, listeners []*eventEntry) error {
	if len(listeners) == 0 {
		return nil
	}

	evtEntry, _ := e.listeners.LoadOrStore(evt, anvil.NewSlice[*eventEntry]())
	evtEntry.Push(listeners...)
	return nil
}

// AddListener parses individual [Listener] components and maps them efficiently.
func (e *emitter) AddListener(evt Name, listeners ...Listener) error {
	if len(listeners) == 0 {
		return nil
	}

	var events []*eventEntry
	for _, event := range listeners {
		if event != nil {
			events = append(events, &eventEntry{fn: event, ptr: reflect.ValueOf(event).Pointer()})
		}
	}

	return e.addListeners(evt, events)
}

// On implements an effortless gateway targeting the standard [AddListener] mechanism.
func (e *emitter) On(evt Name, listeners ...Listener) error {
	return e.AddListener(evt, listeners...)
}

// Emit executes sequential triggers against registered hooks while capturing externalized run panics safely.
func (e *emitter) Emit(evt Name, data ...any) {
	evtEntry, ok := e.listeners.Load(evt)
	if !ok {
		return
	}

	if evtEntry.Len() == 0 {
		return
	}

	for _, event := range evtEntry.All() {
		if event != nil {
			func() {
				defer func() {
					if r := recover(); r != nil {
					}
				}()
				event.fn(data...)
			}()
		}
	}
}

// Names reads and summarizes all present dynamic keys inside the map structures.
func (e *emitter) Names() []Name {
	return e.listeners.Keys()
}

// ListenerCount exposes how many elements are bound underneath an [Name].
func (e *emitter) ListenerCount(evt Name) int {
	evtEntry, ok := e.listeners.Load(evt)
	if !ok {
		return 0
	}

	return evtEntry.Len()
}

// Listeners formats an accessible extraction list containing current tracking [Listener] targets.
func (e *emitter) Listeners(evt Name) []Listener {
	evtEntry, ok := e.listeners.Load(evt)
	if !ok {
		return nil
	}

	entries := evtEntry.All()
	listeners := make([]Listener, len(entries))
	for i, l := range entries {
		listeners[i] = l.fn
	}

	return listeners
}

// oneTimeListener handles single execution semantics safely using [sync.Once] controls.
type oneTimeListener struct {
	fired *sync.Once

	evt     Name
	emitter *emitter
	fn      Listener
}

// execute executes the bound callback one single time and manages structural detach cleanups.
func (l *oneTimeListener) execute(vals ...any) {
	l.fired.Do(func() {
		defer l.emitter.RemoveListener(l.evt, l.fn)
		l.fn(vals...)
	})
}

// Once registers an [Listener] to fire exactly once, guaranteeing isolated disposal afterwards.
func (e *emitter) Once(evt Name, listeners ...Listener) error {
	if len(listeners) == 0 {
		return nil
	}

	var events []*eventEntry
	for _, event := range listeners {
		if event != nil {
			oneTime := &oneTimeListener{fired: &sync.Once{}, evt: evt, emitter: e, fn: event}
			events = append(events, &eventEntry{fn: oneTime.execute, ptr: reflect.ValueOf(event).Pointer()})
		}
	}
	return e.addListeners(evt, events)
}

// RemoveListener evaluates matching function contexts via pointers and excludes them seamlessly.
func (e *emitter) RemoveListener(evt Name, listener Listener) bool {
	if listener == nil {
		return false
	}

	evtEntry, ok := e.listeners.Load(evt)

	if !ok {
		return false
	}

	if evtEntry.Len() == 0 {
		return false
	}

	targetPtr := reflect.ValueOf(listener).Pointer()

	remove, _ := evtEntry.RangeAndSplice(func(listener *eventEntry, i int) (bool, int, int, []*eventEntry) {
		return listener.ptr == targetPtr, i, 1, nil
	})
	return len(remove) > 0
}

// RemoveAllListeners sweeps away all subscription nodes linked to a distinct [Name].
func (e *emitter) RemoveAllListeners(evt Name) bool {
	_, loaded := e.listeners.LoadAndDelete(evt)
	return loaded
}

// Clear drops everything inside the internal concurrent registry mapping arrays.
func (e *emitter) Clear() {
	e.listeners.Clear()
}

// Len gives out the overall structural count of uniquely mapped event groups.
func (e *emitter) Len() int {
	return e.listeners.Len()
}
