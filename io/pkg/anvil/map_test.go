package anvil

import (
	"sync"
	"testing"

	"github.com/atendi9/capivara/assert"
)

// TestMap_StoreAndLoad verifies basic storage and retrieval functionality.
func TestMap_StoreAndLoad(t *testing.T) {
	m := &Map[string, int]{}

	m.Store("key1", 100)
	val, ok := m.Load("key1")

	assert.True(t, ok)
	assert.Equal(t, 100, val)
	assert.Equal(t, 1, m.Len())
}

// TestMap_LoadOrStore verifies the behavior of LoadOrStore.
func TestMap_LoadOrStore(t *testing.T) {
	m := &Map[string, int]{}

	// First store
	actual, loaded := m.LoadOrStore("key1", 100)
	assert.Equal(t, 100, actual)
	assert.False(t, loaded)

	// Second store (should return existing)
	actual, loaded = m.LoadOrStore("key1", 200)
	assert.Equal(t, 100, actual)
	assert.True(t, loaded)
}

// TestMap_Swap verifies the swap functionality.
func TestMap_Swap(t *testing.T) {
	m := &Map[string, int]{}
	m.Store("key1", 100)

	prev, loaded := m.Swap("key1", 200)
	assert.Equal(t, 100, prev)
	assert.True(t, loaded)

	val, _ := m.Load("key1")
	assert.Equal(t, 200, val)
}

// TestMap_Delete verifies deletion logic.
func TestMap_Delete(t *testing.T) {
	m := &Map[string, int]{}
	m.Store("key1", 100)
	m.Delete("key1")

	_, ok := m.Load("key1")
	assert.False(t, ok)
	assert.Equal(t, 0, m.Len())
}

// TestMap_CompareAndSwap verifies the atomic compare and swap logic.
func TestMap_CompareAndSwap(t *testing.T) {
	m := &Map[string, int]{}
	m.Store("key1", 100)

	// Should fail
	swapped := m.CompareAndSwap("key1", 999, 200)
	assert.False(t, swapped)

	// Should succeed
	swapped = m.CompareAndSwap("key1", 100, 200)
	assert.True(t, swapped)

	val, _ := m.Load("key1")
	assert.Equal(t, 200, val)
}

// TestMap_Range verifies the iteration functionality.
func TestMap_Range(t *testing.T) {
	m := &Map[string, int]{}
	m.Store("a", 1)
	m.Store("b", 2)

	count := 0
	m.Range(func(key string, value int) bool {
		count++
		return true
	})

	assert.Equal(t, 2, count)
}

// TestMap_Concurrency verifies concurrent access using the requested WaitGroup pattern.
func TestMap_Concurrency(t *testing.T) {
	m := &Map[int, int]{}
	var wg sync.WaitGroup

	for i := range 100 {
		key := i
		wg.Go(func() {
			m.Store(key, key)
		})
	}
	wg.Wait()

	assert.Equal(t, 100, m.Len())
}

// TestMap_Clear verifies clearing the map.
func TestMap_Clear(t *testing.T) {
	m := &Map[string, int]{}
	m.Store("a", 1)
	m.Store("b", 2)
	m.Clear()

	assert.Equal(t, 0, m.Len())
	_, ok := m.Load("a")
	assert.False(t, ok)
}
