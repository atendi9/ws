package anvil

import (
	"encoding/json"
	"maps"
	"slices"
	"sync"

	"github.com/atendi9/box"
)

// Set represents a thread-safe collection of unique elements.
// It uses a sync.RWMutex to ensure safe concurrent access and a map with [box.Void] values.
type Set[T comparable] struct {
	mu    sync.RWMutex
	cache map[T]box.Void
}

// NewSet creates and returns a new initialization pointer of [Set] containing the provided keys.
func NewSet[T comparable](keys ...T) *Set[T] {
	s := &Set[T]{cache: make(map[T]box.Void, len(keys))}
	for _, key := range keys {
		s.cache[key] = box.NULL
	}
	return s
}

// Add inserts one or multiple keys into the [Set].
// It returns true if elements were provided and added, and false otherwise.
func (s *Set[T]) Add(keys ...T) bool {
	if len(keys) == 0 {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		s.cache[key] = box.NULL
	}
	return true
}

// Delete removes one or multiple keys from the [Set].
// It returns true if elements were provided and targeted for removal, and false otherwise.
func (s *Set[T]) Delete(keys ...T) bool {
	if len(keys) == 0 {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		delete(s.cache, key)
	}
	return true
}

// Clear removes all elements from the [Set] and returns true.
func (s *Set[T]) Clear() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	clear(s.cache)
	return true
}

// Has checks if a specific key exists within the [Set].
// It returns true if the key is present, and false otherwise.
func (s *Set[T]) Has(key T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.cache[key]
	return exists
}

// Len returns the current number of elements inside the [Set].
// It fulfills the standard length reporting functionality.
func (s *Set[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.cache)
}

// All returns a shallow clone of the internal map holding the [Set] data.
// It ensures that the internal cache map cannot be mutated directly from the outside.
func (s *Set[T]) All() map[T]box.Void {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return maps.Clone(s.cache)
}

// Keys extracts and returns all elements of the [Set] as a slice of type T.
func (s *Set[T]) Keys() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.keys()
}

// keys is an internal unexported helper that fetches all map keys into a slice.
// It must be called within an active mutex lock state.
func (s *Set[T]) keys() []T {
	return slices.Collect(maps.Keys(s.cache))
}

// populate is an internal unexported helper that reinitializes the cache map
// with the provided slice of keys.
func (s *Set[T]) populate(keys []T) {
	s.cache = make(map[T]box.Void, len(keys))
	for _, key := range keys {
		s.cache[key] = box.NULL
	}
}

// MarshalJSON encodes the [Set] into a JSON array containing all its keys.
// It implements the [json.Marshaler] interface.
func (s *Set[T]) MarshalJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return json.Marshal(s.keys())
}

// UnmarshalJSON decodes a JSON array of keys into the [Set], overriding any existing state.
// It implements the [json.Unmarshaler] interface.
func (s *Set[T]) UnmarshalJSON(data []byte) error {
	var keys []T
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.populate(keys)
	return nil
}
