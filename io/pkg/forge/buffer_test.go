package forge

import (
	"io"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestBuffer_WriteAndRead(t *testing.T) {
	b := &Buffer{}

	t.Run("Write", func(t *testing.T) {
		n, err := b.WriteString("capivara")
		assert.NoError(t, err)
		assert.Equal(t, 8, n)
		assert.False(t, b.empty())
	})
	t.Run("Read", func(t *testing.T) {
		p := make([]byte, 4)
		n, err := b.Read(p)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, "capi", string(p))

		// Verify remaining string
		assert.Equal(t, "vara", b.String())
	})
}

func TestBuffer_ReadEOF(t *testing.T) {
	b := &Buffer{}
	p := make([]byte, 5)

	n, err := b.Read(p)
	assert.Error(t, err)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 0, n)
}

func TestBuffer_BytesAndLength(t *testing.T) {
	b := &Buffer{}
	b.Write([]byte("gopher"))

	bytesRes := b.Bytes()
	assert.LengthSlice(t, 6, bytesRes)
	assert.Equal(t, 6, b.Len())

	// Ensure cap is properly allocated
	assert.True(t, b.Cap() >= 6)
}

func TestBuffer_ResetAndEmpty(t *testing.T) {
	b := &Buffer{}
	b.WriteString("test data")

	assert.False(t, b.empty())
	b.Reset()
	assert.True(t, b.empty())
	assert.Equal(t, 0, b.Len())
}

func TestBuffer_Clone(t *testing.T) {
	b1 := NewBuffer(nil)
	b1.WriteString("original")

	b2 := b1.Clone()
	assert.Equal(t, b1.String(), b2.String())

	// Modifying b1 should not affect b2
	b1.ReadByte()
	assert.Equal(t, "riginal", b1.String())
	assert.Equal(t, "original", b2.String())
}

func TestBuffer_Peek(t *testing.T) {
	b := NewBufferString("golang")

	peeked, err := b.Peek(2)
	assert.NoError(t, err)
	assert.LengthSlice(t, 2, peeked)
	assert.Equal(t, "go", string(peeked))

	// The buffer should not have advanced
	assert.Equal(t, "golang", b.String())

	// Peek more than available
	_, err = b.Peek(10)
	assert.Error(t, err)
	assert.Equal(t, io.EOF, err)
}
