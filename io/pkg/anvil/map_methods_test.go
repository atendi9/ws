package anvil

import (
	"sort"
	"sync"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestMap_LoadAndDelete(t *testing.T) {
	var m Map[string, int]
	m.Store("a", 1)

	v, loaded := m.LoadAndDelete("a")
	assert.True(t, loaded)
	assert.Equal(t, 1, v)

	_, ok := m.Load("a")
	assert.False(t, ok)

	_, loaded = m.LoadAndDelete("missing")
	assert.False(t, loaded)
}

func TestMap_SwapReturnsPrevious(t *testing.T) {
	var m Map[string, int]

	prev, loaded := m.Swap("a", 1)
	assert.False(t, loaded)
	assert.Equal(t, 0, prev)

	prev, loaded = m.Swap("a", 2)
	assert.True(t, loaded)
	assert.Equal(t, 1, prev)

	v, _ := m.Load("a")
	assert.Equal(t, 2, v)
}

func TestMap_CompareAndDelete(t *testing.T) {
	var m Map[string, int]
	m.Store("a", 1)

	assert.False(t, m.CompareAndDelete("a", 99)) // value mismatch
	assert.True(t, m.CompareAndDelete("a", 1))

	_, ok := m.Load("a")
	assert.False(t, ok)

	assert.False(t, m.CompareAndDelete("missing", 0))
}

func TestMap_KeysAndValues(t *testing.T) {
	var m Map[string, int]
	m.Store("a", 1)
	m.Store("b", 2)
	m.Store("c", 3)

	keys := m.Keys()
	sort.Strings(keys)
	assert.Equal(t, 3, len(keys))
	assert.Equal(t, "a", keys[0])

	values := m.Values()
	sort.Ints(values)
	assert.Equal(t, 3, len(values))
	assert.Equal(t, 1, values[0])
}

func TestMap_Len(t *testing.T) {
	var m Map[int, int]
	for i := 0; i < 5; i++ {
		m.Store(i, i*i)
	}
	assert.Equal(t, 5, m.Len())

	// Early termination of Range.
	seen := 0
	m.Range(func(k, v int) bool {
		seen++
		return seen < 2
	})
	assert.Equal(t, 2, seen)
}

// TestMap_PromotionViaMisses exercises the read/dirty promotion path by
// repeatedly missing on the read map, which triggers promoteDirtyToReadLocked.
func TestMap_PromotionViaMisses(t *testing.T) {
	var m Map[int, int]
	for i := 0; i < 100; i++ {
		m.Store(i, i)
	}
	for round := 0; round < 5; round++ {
		for i := 0; i < 200; i++ {
			m.Load(i)
		}
	}
	v, ok := m.Load(50)
	assert.True(t, ok)
	assert.Equal(t, 50, v)
}

func TestMap_ConcurrentMixed(t *testing.T) {
	var m Map[int, int]
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			m.Store(n, n)
			m.Load(n)
			m.LoadOrStore(n, n)
			m.Swap(n, n+1)
			m.CompareAndSwap(n, n+1, n)
			m.CompareAndDelete(n, 99999)
			_ = m.Keys()
			_ = m.Values()
			m.Range(func(int, int) bool { return true })
		}(i)
	}
	wg.Wait()
	assert.Equal(t, 50, m.Len())
}
