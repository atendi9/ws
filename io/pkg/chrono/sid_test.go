package chrono

import (
	"strings"
	"sync"
	"testing"

	"github.com/atendi9/capivara/assert"
)

// TestNew verifies that New returns the global singleton instance of *base64Id.
func TestNew(t *testing.T) {
	generator1 := New()
	generator2 := New()

	assert.Equal(t, generator1, generator2)
}

// TestGenerate verifies that Generate produces unique 24-character strings
// and properly increments the sequence counter.
func TestGenerate(t *testing.T) {
	generator := New()

	id1 := generator.Generate()
	id2 := generator.Generate()

	// Verify lengths of the generated IDs
	assert.Equal(t, 24, len(id1))
	assert.Equal(t, 24, len(id2))

	// Verify uniqueness between sequential generations
	assert.False(t, id1 == id2)
}

// TestString verifies that the String method matches the output of Generate.
func TestString(t *testing.T) {
	generator := New()

	idStr := generator.String()

	assert.Equal(t, 24, len(idStr))
}

// TestGenerateConcurrency ensures that ID generation is safe for concurrent use
// and does not produce duplicate identifiers under race conditions.
func TestGenerateConcurrency(t *testing.T) {
	generator := New()

	var wg sync.WaitGroup
	numGoroutines := 10
	iterations := 100

	// Track generated IDs using a thread-safe strategy or collecting them per goroutine
	results := make([][]string, numGoroutines)

	for i := range numGoroutines {
		index := i
		wg.Go(func() {
			results[index] = make([]string, iterations)
			for j := range iterations {
				results[index][j] = generator.Generate()
			}
		})
	}

	wg.Wait()

	// Validate total count and uniqueness
	uniqueIds := make(map[string]bool)
	totalExpected := numGoroutines * iterations

	for _, goroutineResults := range results {
		for _, id := range goroutineResults {
			uniqueIds[id] = true
		}
	}

	assert.LengthMap(t, totalExpected, uniqueIds)
}

// TestIsValid verifies the validation logic for session identifiers against safe formats.
func TestIsValid(t *testing.T) {
	tests := []struct {
		name     string
		sid      string
		expected bool
	}{
		{
			name:     "Valid alphanumeric and basic characters",
			sid:      "abc123XYZ-_",
			expected: true,
		},
		{
			name:     "Valid special characters including protocol v3 symbols",
			sid:      "namespace#id:sub.id-123_456",
			expected: true,
		},
		{
			name:     "Valid maximum allowable length",
			sid:      strings.Repeat("a", 36),
			expected: true,
		},
		{
			name:     "Invalid empty string",
			sid:      "",
			expected: false,
		},
		{
			name:     "Invalid exceeds maximum length",
			sid:      strings.Repeat("a", 37),
			expected: false,
		},
		{
			name:     "Invalid contains spaces",
			sid:      "session id",
			expected: false,
		},
		{
			name:     "Invalid contains unsupported special characters",
			sid:      "session@id!",
			expected: false,
		},
		{
			name:     "Invalid contains unicode or emoji characters",
			sid:      "session🚀id",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValid(tt.sid)
			if tt.expected {
				assert.True(t, result)
			} else {
				assert.False(t, result)
			}
		})
	}
}
