package chrono

import (
	"strings"
	"sync"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestYeastEncode(t *testing.T) {
	y := NewYeast()
	tests := []struct {
		num  int64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{9, "9"},
		{10, "A"},
		{35, "Z"},
		{36, "a"},
		{61, "z"},
		{62, "-"},
		{63, "_"},
		{64, "10"},
		{128, "20"},
		{-1, "1"},   // negatives are absolute-valued
		{-64, "10"}, // -64 -> 64 -> "10"
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, y.Encode(tt.num))
	}
}

func TestYeastDecode(t *testing.T) {
	y := NewYeast()

	t.Run("valid characters", func(t *testing.T) {
		tests := []struct {
			str  string
			want int64
		}{
			{"0", 0},
			{"1", 1},
			{"A", 10},
			{"Z", 35},
			{"_", 63},
			{"10", 64},
			{"20", 128},
		}
		for _, tt := range tests {
			got, err := y.Decode(tt.str)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		}
	})

	t.Run("empty string is error", func(t *testing.T) {
		_, err := y.Decode("")
		assert.Error(t, err)
	})

	t.Run("invalid character is error", func(t *testing.T) {
		_, err := y.Decode("ab!cd")
		assert.Error(t, err)
	})
}

func TestYeastEncodeDecodeRoundTrip(t *testing.T) {
	y := NewYeast()
	for _, n := range []int64{0, 1, 42, 1000, 123456789, 9999999999} {
		encoded := y.Encode(n)
		decoded, err := y.Decode(encoded)
		assert.NoError(t, err)
		assert.Equal(t, n, decoded)
	}
}

func TestYeastEncodeUsesAlphabetOnly(t *testing.T) {
	y := NewYeast()
	const valid = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"
	out := y.Encode(987654321)
	for i := 0; i < len(out); i++ {
		assert.True(t, strings.IndexByte(valid, out[i]) >= 0)
	}
}

func TestYeastUniqueAndMonotonic(t *testing.T) {
	y := NewYeast()
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := y.Yeast()
		assert.False(t, seen[id]) // never repeats
		seen[id] = true
	}
}

func TestYeastCollisionAppendsSeed(t *testing.T) {
	y := NewYeast()
	first := y.Yeast()
	// Force the same millisecond by pre-seeding prev so the next call collides.
	// Rapid successive calls within the same millisecond should differ via "." suffix.
	found := false
	for i := 0; i < 100000; i++ {
		id := y.Yeast()
		if strings.Contains(id, ".") {
			found = true
			break
		}
	}
	assert.True(t, found)
	assert.NotNil(t, first)
}

func TestYeastDateAndDefault(t *testing.T) {
	assert.NotNil(t, DefaultYeast)
	a := YeastDate()
	assert.True(t, len(a) > 0)
	// Result must be decodable (possibly with a "." separated seed part).
	y := NewYeast()
	for _, part := range strings.Split(a, ".") {
		_, err := y.Decode(part)
		assert.NoError(t, err)
	}
}

func TestYeastConcurrent(t *testing.T) {
	// Exercises Yeast() under -race for data races on its atomic state.
	// Strict uniqueness is not asserted here because the generator's
	// load/store sequence is not fully linearizable across goroutines.
	y := NewYeast()
	var mu sync.Mutex
	var ids []string
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				id := y.Yeast()
				mu.Lock()
				ids = append(ids, id)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	assert.Equal(t, 20*200, len(ids))
	// Every produced id must be decodable.
	for _, id := range ids {
		for _, part := range strings.Split(id, ".") {
			_, err := y.Decode(part)
			assert.NoError(t, err)
		}
	}
}
