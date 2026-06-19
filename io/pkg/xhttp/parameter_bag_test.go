package xhttp

import (
	"sync"
	"testing"

	"github.com/atendi9/capivara/assert"
)

// TestNewParameterBag verifies that [ParameterBag] initializes correctly
// both with and without initial parameters.
func TestNewParameterBag(t *testing.T) {
	t.Run("Initialize with nil", func(t *testing.T) {
		pb := NewParameterBag(nil)
		assert.NotNil(t, pb)
		assert.Equal(t, 0, pb.Count())
	})

	t.Run("Initialize with values", func(t *testing.T) {
		params := map[string][]string{"key": {"value"}}
		pb := NewParameterBag(params)
		assert.Equal(t, 1, pb.Count())
		val, ok := pb.Get("key")
		assert.True(t, ok)
		assert.Equal(t, "value", val)
	})
}

// TestAll verifies that [ParameterBag.All] returns a valid clone
// and ensures thread safety for the returned map.
func TestAll(t *testing.T) {
	pb := NewParameterBag(map[string][]string{"a": {"b"}})
	all := pb.All()

	assert.Equal(t, 1, len(all))
	assert.Equal(t, "b", all["a"][0])

	// Verify modification of the returned map does not affect the bag
	all["a"][0] = "changed"
	val, _ := pb.Get("a")
	assert.Equal(t, "b", val)
}

// TestKeys verifies that [ParameterBag.Keys] returns all stored keys.
func TestKeys(t *testing.T) {
	pb := NewParameterBag(map[string][]string{"k1": {"v1"}, "k2": {"v2"}})
	keys := pb.Keys()
	assert.LengthSlice(t, 2, keys)
}

// TestReplace verifies that [ParameterBag.Replace] overrides existing parameters.
func TestReplace(t *testing.T) {
	pb := NewParameterBag(map[string][]string{"k1": {"v1"}})
	pb.Replace(map[string][]string{"k2": {"v2"}})

	assert.Equal(t, 1, pb.Count())
	assert.False(t, pb.Has("k1"))
	assert.True(t, pb.Has("k2"))
}

// TestWith verifies that [ParameterBag.With] merges new parameters.
func TestWith(t *testing.T) {
	pb := NewParameterBag(map[string][]string{"k1": {"v1"}})
	pb.With(map[string][]string{"k2": {"v2"}})

	assert.Equal(t, 2, pb.Count())
	assert.True(t, pb.Has("k1"))
	assert.True(t, pb.Has("k2"))
}

// TestAdd verifies that [ParameterBag.Add] appends values to the slice.
func TestAdd(t *testing.T) {
	pb := NewParameterBag(nil)
	pb.Add("k1", "v1")
	pb.Add("k1", "v2")

	values, ok := pb.Gets("k1")
	assert.True(t, ok)
	assert.LengthSlice(t, 2, values)
	assert.Equal(t, "v1", values[0])
	assert.Equal(t, "v2", values[1])
}

// TestGetAndVariations verifies [ParameterBag.Get], [ParameterBag.GetFirst],
// and [ParameterBag.GetLast] behaviors.
func TestGetAndVariations(t *testing.T) {
	pb := NewParameterBag(map[string][]string{"k1": {"v1", "v2"}})

	t.Run("Get / GetLast returns last value", func(t *testing.T) {
		val, ok := pb.Get("k1")
		assert.True(t, ok)
		assert.Equal(t, "v2", val)
	})

	t.Run("GetFirst returns first value", func(t *testing.T) {
		val, ok := pb.GetFirst("k1")
		assert.True(t, ok)
		assert.Equal(t, "v1", val)
	})

	t.Run("Get with default", func(t *testing.T) {
		val, ok := pb.Get("missing", "default")
		assert.False(t, ok)
		assert.Equal(t, "default", val)
	})
}

// TestGets verifies that [ParameterBag.Gets] returns a clone of the slice.
func TestGets(t *testing.T) {
	pb := NewParameterBag(map[string][]string{"k1": {"v1", "v2"}})
	vals, ok := pb.Gets("k1")
	assert.True(t, ok)
	assert.LengthSlice(t, 2, vals)

	// Verify immutability
	vals[0] = "changed"
	check, _ := pb.Gets("k1")
	assert.Equal(t, "v1", check[0])
}

// TestSet verifies that [ParameterBag.Set] replaces existing values.
func TestSet(t *testing.T) {
	pb := NewParameterBag(map[string][]string{"k1": {"old"}})
	pb.Set("k1", "new")

	val, _ := pb.Get("k1")
	assert.Equal(t, "new", val)
}

// TestHas verifies that [ParameterBag.Has] correctly identifies keys.
func TestHas(t *testing.T) {
	pb := NewParameterBag(map[string][]string{"k1": {"v1"}})
	assert.True(t, pb.Has("k1"))
	assert.False(t, pb.Has("k2"))
}

// TestRemove verifies that [ParameterBag.Remove] deletes keys.
func TestRemove(t *testing.T) {
	pb := NewParameterBag(map[string][]string{"k1": {"v1"}})
	pb.Remove("k1")
	assert.False(t, pb.Has("k1"))
	assert.Equal(t, 0, pb.Count())
}

// TestConcurrency verifies that [ParameterBag] is thread-safe using concurrent operations.
func TestConcurrency(t *testing.T) {
	pb := NewParameterBag(nil)
	var wg sync.WaitGroup

	// Concurrent writes
	for range 100 {
		wg.Go(func() {
			pb.Add("key", "val")
		})
	}
	wg.Wait()

	// Concurrent reads
	for range 100 {
		wg.Go(func() {
			_ = pb.Has("key")
			_, _ = pb.Get("key")
		})
	}
	wg.Wait()

	vals, _ := pb.Gets("key")
	assert.LengthSlice(t, 100, vals)
}
