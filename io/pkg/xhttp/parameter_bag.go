package xhttp

import (
	"slices"
	"sync"
)

// ParameterBag provides a thread-safe container for HTTP parameters where keys map to multiple values.
// It utilizes a [sync.RWMutex] to handle concurrent access.
type ParameterBag struct {
	mu         sync.RWMutex
	parameters map[string][]string
}

// NewParameterBag creates a new [ParameterBag] instance.
// If the provided parameters map is nil, it initializes an empty one.
func NewParameterBag(parameters map[string][]string) *ParameterBag {
	if parameters == nil {
		parameters = make(map[string][]string)
	}
	return &ParameterBag{parameters: parameters}
}

// All returns a copy of all parameters currently stored in the [ParameterBag].
// The returned map is a clone to ensure thread safety and avoid external mutation.
func (p *ParameterBag) All() map[string][]string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	_tmp := make(map[string][]string, len(p.parameters))
	for k, v := range p.parameters {
		_tmp[k] = slices.Clone(v)
	}

	return _tmp
}

// Keys returns a slice containing all parameter keys present in the [ParameterBag].
func (p *ParameterBag) Keys() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	keys := make([]string, 0, len(p.parameters))
	for k := range p.parameters {
		keys = append(keys, k)
	}
	return keys
}

// Replace replaces the current collection of parameters with a new map.
// Note: This operation is thread-safe.
func (p *ParameterBag) Replace(parameters map[string][]string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.parameters = parameters
}

// With adds or updates the existing parameters with the provided map.
// Each value in the provided map is cloned to ensure internal consistency.
func (p *ParameterBag) With(parameters map[string][]string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for k, v := range parameters {
		p.parameters[k] = slices.Clone(v)
	}
}

// Add appends a value to the list of values associated with the given key.
// If the key does not exist, it creates a new slice.
func (p *ParameterBag) Add(key string, value string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.parameters[key] = append(p.parameters[key], value)
}

// Get returns the last value associated with the specified key.
// If the key is not found or empty, it returns the first element of _default if provided.
// It returns the found string and a boolean indicating success.
func (p *ParameterBag) Get(key string, _default ...string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if value, ok := p.parameters[key]; ok && len(value) > 0 {
		return value[len(value)-1], true
	}
	if len(_default) > 0 {
		return _default[0], false
	}
	return "", false
}

// Peek returns the last value associated with the specified key,
// defaulting to an empty string if the key does not exist.
func (p *ParameterBag) Peek(key string, _default ...string) string {
	v, _ := p.Get(key, _default...)
	return v
}

// GetFirst returns the first value associated with the specified key.
// If the key is not found, it returns the first element of _default if provided.
func (p *ParameterBag) GetFirst(key string, _default ...string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if value, ok := p.parameters[key]; ok && len(value) > 0 {
		return value[0], true
	}
	if len(_default) > 0 {
		return _default[0], false
	}
	return "", false
}

// GetLast returns the last value associated with the specified key.
// It behaves the same as [ParameterBag.Get].
func (p *ParameterBag) GetLast(key string, _default ...string) (string, bool) {
	return p.Get(key, _default...)
}

// Gets returns a clone of the slice of values associated with the specified key.
// If the key is not found, it returns the provided _default value (if any).
func (p *ParameterBag) Gets(key string, _default ...[]string) ([]string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if v, ok := p.parameters[key]; ok {
		return slices.Clone(v), true
	}
	if len(_default) > 0 {
		return _default[0], false
	}
	return []string{}, false
}

// Set assigns a single value to the specified key, replacing any existing values.
func (p *ParameterBag) Set(key string, value string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.parameters[key] = []string{value}
}

// Has checks if the specified key exists in the [ParameterBag].
func (p *ParameterBag) Has(key string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	_, ok := p.parameters[key]
	return ok
}

// Remove deletes the specified key and all its associated values from the [ParameterBag].
func (p *ParameterBag) Remove(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.parameters, key)
}

// Count returns the total number of unique keys stored in the [ParameterBag].
func (p *ParameterBag) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.parameters)
}
