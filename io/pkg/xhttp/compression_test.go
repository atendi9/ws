package xhttp

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"testing"

	"github.com/atendi9/capivara/assert"
)

// TestCompress validates the Compress method for gzip, deflate, and fallback scenarios.
func TestCompress(t *testing.T) {
	c := &Compression{}
	data := []byte("hello world")

	t.Run("gzip compression", func(t *testing.T) {
		reader := bytes.NewBuffer(data)
		compressed, err := c.Compress("gzip", reader)
		assert.NoError(t, err)

		// Decompress to verify
		gzReader, err := gzip.NewReader(compressed)
		assert.NoError(t, err)
		defer gzReader.Close()

		result, err := io.ReadAll(gzReader)
		assert.NoError(t, err)
		assert.Equal(t, string(data), string(result))
	})

	t.Run("deflate compression", func(t *testing.T) {
		reader := bytes.NewBuffer(data)
		compressed, err := c.Compress("deflate", reader)
		assert.NoError(t, err)

		// Decompress to verify
		flReader := flate.NewReader(compressed)
		defer flReader.Close()

		result, err := io.ReadAll(flReader)
		assert.NoError(t, err)
		assert.Equal(t, string(data), string(result))
	})

	t.Run("no compression (fallback)", func(t *testing.T) {
		reader := bytes.NewBuffer(data)
		compressed, err := c.Compress("unknown", reader)
		assert.NoError(t, err)

		result, err := io.ReadAll(compressed)
		assert.NoError(t, err)
		assert.Equal(t, string(data), string(result))
	})
}

// TestNewWriter validates the NewWriter method for correct compression types and error handling.
func TestNewWriter(t *testing.T) {
	c := &Compression{}
	buf := new(bytes.Buffer)

	t.Run("new gzip writer", func(t *testing.T) {
		writer, err := c.NewWriter(buf, "gzip")
		assert.NoError(t, err)
		assert.True(t, writer != nil)
		writer.Close()
	})

	t.Run("new deflate writer", func(t *testing.T) {
		writer, err := c.NewWriter(buf, "deflate")
		assert.NoError(t, err)
		assert.True(t, writer != nil)
		writer.Close()
	})

	t.Run("unsupported compression", func(t *testing.T) {
		writer, err := c.NewWriter(buf, "invalid")
		assert.Error(t, err)
		assert.True(t, writer == nil)
	})
}
