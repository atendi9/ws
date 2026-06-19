package forge

import (
	"io"
	"strings"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestBuffer_AvailableAndGrow(t *testing.T) {
	b := NewBuffer(nil)
	b.Grow(64)
	assert.True(t, b.Available() >= 64)
	assert.True(t, b.Cap() >= 64)

	// AvailableBuffer returns an empty slice with spare capacity.
	avail := b.AvailableBuffer()
	assert.Equal(t, 0, len(avail))
}

func TestBuffer_StringNil(t *testing.T) {
	var b *Buffer
	assert.Equal(t, "<nil>", b.String())
}

func TestBuffer_Truncate(t *testing.T) {
	b := NewBufferString("hello world")
	b.Truncate(5)
	assert.Equal(t, "hello", b.String())

	b.Truncate(0)
	assert.Equal(t, "", b.String())
	assert.True(t, b.Len() == 0)
}

func TestBuffer_WriteByteAndRune(t *testing.T) {
	b := NewBuffer(nil)
	assert.NoError(t, b.WriteByte('h'))
	assert.NoError(t, b.WriteByte('i'))

	n, err := b.WriteRune('世')
	assert.NoError(t, err)
	assert.Equal(t, 3, n) // 世 is 3 bytes in UTF-8

	n, err = b.WriteRune('A')
	assert.NoError(t, err)
	assert.Equal(t, 1, n)

	assert.Equal(t, "hi世A", b.String())
}

func TestBuffer_Next(t *testing.T) {
	b := NewBufferString("abcdef")
	got := b.Next(3)
	assert.Equal(t, "abc", string(got))
	assert.Equal(t, "def", b.String())

	// Next more than available returns the remainder.
	got = b.Next(100)
	assert.Equal(t, "def", string(got))
}

func TestBuffer_ReadByte(t *testing.T) {
	b := NewBufferString("xy")
	c, err := b.ReadByte()
	assert.NoError(t, err)
	assert.Equal(t, byte('x'), c)
	c, err = b.ReadByte()
	assert.NoError(t, err)
	assert.Equal(t, byte('y'), c)

	_, err = b.ReadByte()
	assert.Error(t, err) // EOF
}

func TestBuffer_ReadRune(t *testing.T) {
	b := NewBufferString("世界")
	r, size, err := b.ReadRune()
	assert.NoError(t, err)
	assert.Equal(t, '世', r)
	assert.Equal(t, 3, size)

	// UnreadRune restores it.
	assert.NoError(t, b.UnreadRune())
	r, _, err = b.ReadRune()
	assert.NoError(t, err)
	assert.Equal(t, '世', r)

	_, _, _ = b.ReadRune() // consume 界
	_, _, err = b.ReadRune()
	assert.Error(t, err) // EOF
}

func TestBuffer_UnreadByte(t *testing.T) {
	b := NewBufferString("ab")
	_, err := b.ReadByte()
	assert.NoError(t, err)
	assert.NoError(t, b.UnreadByte())
	c, err := b.ReadByte()
	assert.NoError(t, err)
	assert.Equal(t, byte('a'), c)

	// UnreadByte without a preceding successful read should error.
	fresh := NewBufferString("z")
	assert.Error(t, fresh.UnreadByte())
}

func TestBuffer_ReadBytesAndString(t *testing.T) {
	b := NewBufferString("line1\nline2\n")
	got, err := b.ReadBytes('\n')
	assert.NoError(t, err)
	assert.Equal(t, "line1\n", string(got))

	s, err := b.ReadString('\n')
	assert.NoError(t, err)
	assert.Equal(t, "line2\n", s)

	// Delimiter not found returns remaining data with an error.
	b2 := NewBufferString("nodelim")
	_, err = b2.ReadBytes('\n')
	assert.Error(t, err)
}

func TestBuffer_WriteTo(t *testing.T) {
	b := NewBufferString("payload")
	var sink strings.Builder
	n, err := b.WriteTo(&sink)
	assert.NoError(t, err)
	assert.Equal(t, int64(len("payload")), n)
	assert.Equal(t, "payload", sink.String())
	// Buffer is drained after WriteTo.
	assert.Equal(t, 0, b.Len())
}

func TestBuffer_SizeAndSeek(t *testing.T) {
	b := NewBufferString("0123456789")
	assert.Equal(t, int64(10), b.Size())

	pos, err := b.Seek(3, io.SeekStart)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), pos)
	assert.Equal(t, "3456789", b.String())

	// Reset offset, then seek from current.
	_, err = b.Seek(0, io.SeekStart)
	assert.NoError(t, err)
	pos, err = b.Seek(2, io.SeekCurrent)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), pos)

	pos, err = b.Seek(-1, io.SeekEnd)
	assert.NoError(t, err)
	assert.Equal(t, int64(9), pos)

	// Invalid whence and out-of-range offsets error.
	_, err = b.Seek(0, 99)
	assert.Error(t, err)
	_, err = b.Seek(-5, io.SeekStart)
	assert.Error(t, err)
	_, err = b.Seek(100, io.SeekStart)
	assert.Error(t, err)
}

func TestBuffer_ReadFromLargeInput(t *testing.T) {
	b := NewBuffer(nil)
	src := strings.NewReader(strings.Repeat("x", 5000))
	n, err := b.ReadFrom(src)
	assert.NoError(t, err)
	assert.Equal(t, int64(5000), n)
	assert.Equal(t, 5000, b.Len())
}
