package anvil

import (
	"encoding/json"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestNewSet(t *testing.T) {
	s := NewSet("apple", "banana")

	assert.True(t, s.Has("apple"))
	assert.True(t, s.Has("banana"))
	assert.Equal(t, 2, s.Len())
}

func TestSet_Add(t *testing.T) {
	s := NewSet[string]()

	// Adding elements should return true
	ok := s.Add("golang", "rust")
	assert.True(t, ok)
	assert.True(t, s.Has("golang"))
	assert.True(t, s.Has("rust"))

	// Adding empty arguments should return false
	okEmpty := s.Add()
	assert.False(t, okEmpty)
}

func TestSet_Delete(t *testing.T) {
	s := NewSet("go", "js", "ts")

	// Deleting elements should return true
	ok := s.Delete("js", "ts")
	assert.True(t, ok)
	assert.False(t, s.Has("js"))
	assert.False(t, s.Has("ts"))
	assert.True(t, s.Has("go"))

	// Deleting empty arguments should return false
	okEmpty := s.Delete()
	assert.False(t, okEmpty)
}

func TestSet_Clear(t *testing.T) {
	s := NewSet(1, 2, 3)

	ok := s.Clear()
	assert.True(t, ok)
	assert.Equal(t, 0, s.Len())
	assert.False(t, s.Has(1))
}

func TestSet_Has(t *testing.T) {
	s := NewSet("find-me")

	assert.True(t, s.Has("find-me"))
	assert.False(t, s.Has("missing"))
}

func TestSet_Len(t *testing.T) {
	s := NewSet[int]()
	assert.Equal(t, 0, s.Len())

	s.Add(10, 20)
	assert.Equal(t, 2, s.Len())
}

func TestSet_All(t *testing.T) {
	s := NewSet("A", "B")

	cacheMap := s.All()
	assert.LengthMap(t, 2, cacheMap)

	// Validate content
	_, hasA := cacheMap["A"]
	_, hasB := cacheMap["B"]
	assert.True(t, hasA)
	assert.True(t, hasB)
}

func TestSet_Keys(t *testing.T) {
	s := NewSet("unique-key")

	keys := s.Keys()
	assert.LengthSlice(t, 1, keys)
	assert.Equal(t, "unique-key", keys[0])
}

func TestSet_MarshalJSON(t *testing.T) {
	s := NewSet("json-test")

	data, err := json.Marshal(s)
	assert.NoError(t, err)

	// A single item set marshals into an array containing that element
	var result []string
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)
	assert.LengthSlice(t, 1, result)
	assert.Equal(t, "json-test", result[0])
}

func TestSet_UnmarshalJSON(t *testing.T) {
	jsonData := []byte(`["unmarshal-1", "unmarshal-2"]`)
	s := NewSet[string]()

	err := json.Unmarshal(jsonData, s)
	assert.NoError(t, err)

	assert.Equal(t, 2, s.Len())
	assert.True(t, s.Has("unmarshal-1"))
	assert.True(t, s.Has("unmarshal-2"))

	// Test invalid JSON error handling
	invalidData := []byte(`{invalid-json}`)
	errInvalid := json.Unmarshal(invalidData, s)
	assert.Error(t, errInvalid)
}
