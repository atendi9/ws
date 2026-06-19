package parser

import (
	"bytes"
	"strings"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/ws/io/pkg/forge"
)

func TestIsBinary(t *testing.T) {
	t.Run("should return false for *forge.String", func(t *testing.T) {
		data := &forge.String{}
		assert.False(t, IsBinary(data))
	})

	t.Run("should return false for *strings.Reader", func(t *testing.T) {
		data := strings.NewReader("text content")
		assert.False(t, IsBinary(data))
	})

	t.Run("should return true for byte slice", func(t *testing.T) {
		data := []byte("binary content")
		assert.True(t, IsBinary(data))
	})

	t.Run("should return true for io.Reader", func(t *testing.T) {
		data := bytes.NewReader([]byte("binary content via reader"))
		assert.True(t, IsBinary(data))
	})

	t.Run("should return false for string", func(t *testing.T) {
		data := "just a normal string"
		assert.False(t, IsBinary(data))
	})

	t.Run("should return false for integer", func(t *testing.T) {
		data := 42
		assert.False(t, IsBinary(data))
	})
}

// TestHasBinary tests the HasBinary function, including recursive nested structures.
func TestHasBinary(t *testing.T) {
	t.Run("should return false for nil", func(t *testing.T) {
		assert.False(t, HasBinary(nil))
	})

	t.Run("should return false for simple non-binary type", func(t *testing.T) {
		assert.False(t, HasBinary("simple text"))
	})

	t.Run("should return true for simple binary type", func(t *testing.T) {
		assert.True(t, HasBinary([]byte("binary")))
	})

	t.Run("should return false for slice without binary", func(t *testing.T) {
		data := []any{"text", 123, false}
		assert.False(t, HasBinary(data))
	})

	t.Run("should return true for slice with binary", func(t *testing.T) {
		data := []any{"text", []byte("binary data"), 123}
		assert.True(t, HasBinary(data))
	})

	t.Run("should return false for map without binary", func(t *testing.T) {
		data := map[string]any{
			"name": "John",
			"age":  30,
		}
		assert.False(t, HasBinary(data))
	})

	t.Run("should return true for map with binary", func(t *testing.T) {
		data := map[string]any{
			"name": "John",
			"file": []byte("file content"),
		}
		assert.True(t, HasBinary(data))
	})

	t.Run("should return true for deeply nested structures with binary", func(t *testing.T) {
		data := map[string]any{
			"metadata": []any{
				"info",
				map[string]any{
					"payload": []byte("deep binary payload"),
				},
			},
		}
		assert.True(t, HasBinary(data))
	})

	t.Run("should return false for deeply nested structures without binary", func(t *testing.T) {
		data := map[string]any{
			"metadata": []any{
				"info",
				map[string]any{
					"payload": "deep text payload",
				},
			},
		}
		assert.False(t, HasBinary(data))
	})
}
