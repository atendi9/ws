package anvil

import (
	"sync"
	"sync/atomic"
)

// Map is a concurrent map with amortized constant-time lookups, stores, and deletes.
// It is safe for concurrent use by multiple goroutines without additional locking.
type Map[Key comparable, Value any] struct {
	_             noCopy
	mutex         sync.Mutex
	readOnlyMap   atomic.Pointer[readStore[Key, Value]]
	dirtyMap      map[Key]*mapEntry[Value]
	missesCount   int
	currentLength atomic.Int64
}

// noCopy may be embedded into structs which must not be copied
// after the first use.
type noCopy struct{}

// Lock is a no-op used by go vet to check for copying.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

// readStore is an immutable structure stored atomically in the [Map].
type readStore[Key comparable, Value any] struct {
	readOnlyCache map[Key]*mapEntry[Value]
	hasNewKeys    bool
}

// mapEntry represents a single slot in the [Map].
type mapEntry[Value any] struct {
	valuePointer   atomic.Pointer[Value]
	tombstoneValue *Value
}

// newMapEntry creates a new [mapEntry] initializing it with the given value.
func newMapEntry[Value any](value Value) *mapEntry[Value] {
	entry := &mapEntry[Value]{tombstoneValue: new(Value)}
	entry.valuePointer.Store(&value)
	return entry
}

// loadReadOnlyCache retrieves the current atomic read-only snapshot from [Map].
func (m *Map[Key, Value]) loadReadOnlyCache() readStore[Key, Value] {
	pointer := m.readOnlyMap.Load()
	if pointer != nil {
		return *pointer
	}
	return readStore[Key, Value]{}
}

// Load returns the value stored in the [Map] for a key, or nil if no value is present.
func (m *Map[Key, Value]) Load(key Key) (value Value, ok bool) {
	read := m.loadReadOnlyCache()
	entry, ok := read.readOnlyCache[key]
	if !ok && read.hasNewKeys {
		m.mutex.Lock()
		read = m.loadReadOnlyCache()
		entry, ok = read.readOnlyCache[key]
		if !ok && read.hasNewKeys {
			entry, ok = m.dirtyMap[key]
			m.incrementMissesLocked()
		}
		m.mutex.Unlock()
	}
	if !ok {
		return value, false
	}
	return entry.loadValue()
}

// loadValue atomically loads the value inside a [mapEntry].
func (e *mapEntry[Value]) loadValue() (value Value, ok bool) {
	pointer := e.valuePointer.Load()
	if pointer == nil || pointer == e.tombstoneValue {
		return value, false
	}
	return *pointer, true
}

// Store sets the value for a key in the [Map].
func (m *Map[Key, Value]) Store(key Key, value Value) {
	_, _ = m.Swap(key, value)
}

// Clear removes all elements from the [Map].
func (m *Map[Key, Value]) Clear() {
	read := m.loadReadOnlyCache()
	if len(read.readOnlyCache) == 0 && !read.hasNewKeys {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	read = m.loadReadOnlyCache()
	if len(read.readOnlyCache) > 0 || read.hasNewKeys {
		m.readOnlyMap.Store(&readStore[Key, Value]{})
	}

	clear(m.dirtyMap)
	m.missesCount = 0
	m.currentLength.Store(0)
}

// tryCompareAndSwap updates the entry value if it matches the expected old value.
func (e *mapEntry[Value]) tryCompareAndSwap(oldValue, newValue Value) bool {
	pointer := e.valuePointer.Load()
	if pointer == nil || pointer == e.tombstoneValue || any(*pointer) != any(oldValue) {
		return false
	}
	newCopy := newValue
	for {
		if e.valuePointer.CompareAndSwap(pointer, &newCopy) {
			return true
		}
		pointer = e.valuePointer.Load()
		if pointer == nil || pointer == e.tombstoneValue || any(*pointer) != any(oldValue) {
			return false
		}
	}
}

// unexpungeLocked unmarks a tombstoned entry, preparing it to receive values again.
func (e *mapEntry[Value]) unexpungeLocked() (wasExpunged bool) {
	return e.valuePointer.CompareAndSwap(e.tombstoneValue, nil)
}

// swapValueLocked unconditionally exchanges the value pointer inside the [mapEntry].
func (e *mapEntry[Value]) swapValueLocked(newValue *Value) *Value {
	return e.valuePointer.Swap(newValue)
}

// LoadOrStore returns the existing value for the key if present. Otherwise, it stores and returns the given value.
func (m *Map[Key, Value]) LoadOrStore(key Key, value Value) (actual Value, loaded bool) {
	read := m.loadReadOnlyCache()
	if entry, ok := read.readOnlyCache[key]; ok {
		var stored bool
		actual, loaded, stored = entry.tryLoadOrStore(value)
		if stored {
			if !loaded {
				m.currentLength.Add(1)
			}
			return actual, loaded
		}
	}

	m.mutex.Lock()
	read = m.loadReadOnlyCache()
	if entry, ok := read.readOnlyCache[key]; ok {
		if entry.unexpungeLocked() {
			m.dirtyMap[key] = entry
		}
		actual, loaded, _ = entry.tryLoadOrStore(value)
	} else if entry, ok := m.dirtyMap[key]; ok {
		actual, loaded, _ = entry.tryLoadOrStore(value)
		m.incrementMissesLocked()
	} else {
		if !read.hasNewKeys {
			m.promoteDirtyToReadLocked()
			m.readOnlyMap.Store(&readStore[Key, Value]{readOnlyCache: read.readOnlyCache, hasNewKeys: true})
		}
		m.dirtyMap[key] = newMapEntry(value)
		actual, loaded = value, false
	}
	m.mutex.Unlock()

	if !loaded {
		m.currentLength.Add(1)
	}
	return actual, loaded
}

// tryLoadOrStore attempts to safely load or store a value inside a [mapEntry].
func (e *mapEntry[Value]) tryLoadOrStore(newValue Value) (actual Value, loaded, ok bool) {
	pointer := e.valuePointer.Load()
	if pointer == e.tombstoneValue {
		return actual, false, false
	}
	if pointer != nil {
		return *pointer, true, true
	}

	newValueCopy := newValue
	for {
		if e.valuePointer.CompareAndSwap(nil, &newValueCopy) {
			return newValue, false, true
		}
		pointer = e.valuePointer.Load()
		if pointer == e.tombstoneValue {
			return actual, false, false
		}
		if pointer != nil {
			return *pointer, true, true
		}
	}
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
func (m *Map[Key, Value]) LoadAndDelete(key Key) (value Value, loaded bool) {
	read := m.loadReadOnlyCache()
	entry, ok := read.readOnlyCache[key]
	if !ok && read.hasNewKeys {
		m.mutex.Lock()
		read = m.loadReadOnlyCache()
		entry, ok = read.readOnlyCache[key]
		if !ok && read.hasNewKeys {
			entry, ok = m.dirtyMap[key]
			delete(m.dirtyMap, key)
			m.incrementMissesLocked()
		}
		m.mutex.Unlock()
	}
	if ok {
		value, ok = entry.deleteValue()
		if ok {
			m.currentLength.Add(-1)
		}
		return value, ok
	}
	return value, false
}

// Delete deletes the value for a key from the [Map].
func (m *Map[Key, Value]) Delete(key Key) {
	m.LoadAndDelete(key)
}

// deleteValue soft-deletes the value inside a [mapEntry] by replacing it with nil.
func (e *mapEntry[Value]) deleteValue() (value Value, ok bool) {
	for {
		pointer := e.valuePointer.Load()
		if pointer == nil || pointer == e.tombstoneValue {
			return value, false
		}
		if e.valuePointer.CompareAndSwap(pointer, nil) {
			return *pointer, true
		}
	}
}

// trySwap attempts to exchange the value inside a [mapEntry] if it hasn't been expunged.
func (e *mapEntry[Value]) trySwap(newValue *Value) (*Value, bool) {
	for {
		pointer := e.valuePointer.Load()
		if pointer == e.tombstoneValue {
			return nil, false
		}
		if e.valuePointer.CompareAndSwap(pointer, newValue) {
			return pointer, true
		}
	}
}

// Swap swaps the value for a key and returns the previous value if any.
func (m *Map[Key, Value]) Swap(key Key, value Value) (previous Value, loaded bool) {
	read := m.loadReadOnlyCache()
	if entry, ok := read.readOnlyCache[key]; ok {
		if previousPointer, ok := entry.trySwap(&value); ok {
			if previousPointer == nil {
				m.currentLength.Add(1)
				return previous, false
			}
			return *previousPointer, true
		}
	}

	m.mutex.Lock()
	read = m.loadReadOnlyCache()
	if entry, ok := read.readOnlyCache[key]; ok {
		if entry.unexpungeLocked() {
			m.dirtyMap[key] = entry
		}
		if previousPointer := entry.swapValueLocked(&value); previousPointer != nil {
			loaded = true
			previous = *previousPointer
		}
	} else if entry, ok := m.dirtyMap[key]; ok {
		if previousPointer := entry.swapValueLocked(&value); previousPointer != nil {
			loaded = true
			previous = *previousPointer
		}
	} else {
		if !read.hasNewKeys {
			m.promoteDirtyToReadLocked()
			m.readOnlyMap.Store(&readStore[Key, Value]{readOnlyCache: read.readOnlyCache, hasNewKeys: true})
		}
		m.dirtyMap[key] = newMapEntry(value)
	}
	m.mutex.Unlock()

	if !loaded {
		m.currentLength.Add(1)
	}

	return previous, loaded
}

// CompareAndSwap executes an atomic compare-and-swap operation for a key in the [Map].
func (m *Map[Key, Value]) CompareAndSwap(key Key, old, new Value) (swapped bool) {
	read := m.loadReadOnlyCache()
	if entry, ok := read.readOnlyCache[key]; ok {
		return entry.tryCompareAndSwap(old, new)
	} else if !read.hasNewKeys {
		return false
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	read = m.loadReadOnlyCache()
	swapped = false
	if entry, ok := read.readOnlyCache[key]; ok {
		swapped = entry.tryCompareAndSwap(old, new)
	} else if entry, ok := m.dirtyMap[key]; ok {
		swapped = entry.tryCompareAndSwap(old, new)
		m.incrementMissesLocked()
	}
	return swapped
}

// CompareAndDelete deletes the entry for a key if its current value equals the old value.
func (m *Map[Key, Value]) CompareAndDelete(key Key, old Value) (deleted bool) {
	read := m.loadReadOnlyCache()
	entry, ok := read.readOnlyCache[key]
	if !ok && read.hasNewKeys {
		m.mutex.Lock()
		read = m.loadReadOnlyCache()
		entry, ok = read.readOnlyCache[key]
		if !ok && read.hasNewKeys {
			entry, ok = m.dirtyMap[key]
			m.incrementMissesLocked()
		}
		m.mutex.Unlock()
	}
	for ok {
		pointer := entry.valuePointer.Load()
		if pointer == nil || pointer == entry.tombstoneValue || any(*pointer) != any(old) {
			return false
		}
		if entry.valuePointer.CompareAndSwap(pointer, nil) {
			m.currentLength.Add(-1)
			return true
		}
	}
	return false
}

// Range calls f sequentially for each key and value present in the [Map].
func (m *Map[Key, Value]) Range(f func(key Key, value Value) bool) {
	read := m.loadReadOnlyCache()
	if read.hasNewKeys {
		m.mutex.Lock()
		read = m.loadReadOnlyCache()
		if read.hasNewKeys {
			read = readStore[Key, Value]{readOnlyCache: m.dirtyMap}
			copyRead := read
			m.readOnlyMap.Store(&copyRead)
			m.dirtyMap = nil
			m.missesCount = 0
		}
		m.mutex.Unlock()
	}

	for k, entry := range read.readOnlyCache {
		v, ok := entry.loadValue()
		if !ok {
			continue
		}
		if !f(k, v) {
			break
		}
	}
}

// Len returns the number of active elements inside the [Map].
func (m *Map[Key, Value]) Len() (n int) {
	return int(m.currentLength.Load())
}

// Keys extracts all valid keys currently stored in the [Map].
func (m *Map[Key, Value]) Keys() (keys []Key) {
	m.Range(func(k Key, _ Value) bool {
		keys = append(keys, k)
		return true
	})
	return keys
}

// Values extracts all valid values currently stored in the [Map].
func (m *Map[Key, Value]) Values() (values []Value) {
	m.Range(func(_ Key, v Value) bool {
		values = append(values, v)
		return true
	})
	return values
}

// incrementMissesLocked notes a miss and promotes the dirty map if thresholds are reached.
func (m *Map[Key, Value]) incrementMissesLocked() {
	m.missesCount++
	if m.missesCount < len(m.dirtyMap) {
		return
	}
	m.readOnlyMap.Store(&readStore[Key, Value]{readOnlyCache: m.dirtyMap})
	m.dirtyMap = nil
	m.missesCount = 0
}

// promoteDirtyToReadLocked initializes the dirty map from the current clean read data.
func (m *Map[Key, Value]) promoteDirtyToReadLocked() {
	if m.dirtyMap != nil {
		return
	}

	read := m.loadReadOnlyCache()
	m.dirtyMap = make(map[Key]*mapEntry[Value], len(read.readOnlyCache))
	for k, entry := range read.readOnlyCache {
		if !entry.tryTombstoneLocked() {
			m.dirtyMap[k] = entry
		}
	}
}

// tryTombstoneLocked tries to mark a nil entry as clean expunged (tombstoned).
func (e *mapEntry[Value]) tryTombstoneLocked() (isExpunged bool) {
	pointer := e.valuePointer.Load()
	for pointer == nil {
		if e.valuePointer.CompareAndSwap(nil, e.tombstoneValue) {
			return true
		}
		pointer = e.valuePointer.Load()
	}
	return pointer == e.tombstoneValue
}
