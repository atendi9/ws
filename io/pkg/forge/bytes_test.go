package forge

import (
	"bytes"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestBytesBuffer_Clone(t *testing.T) {
	var b *BytesBuffer
	clonedNil := b.Clone()
	assert.Equal(t, nil, clonedNil)

	validBuf := NewBytesBuffer([]byte("capivara payload"))
	clonedValid := validBuf.(*BytesBuffer).Clone()

	ok := clonedValid != nil
	assert.True(t, ok)
}

func TestBytesBuffer_GoString(t *testing.T) {
	var b *BytesBuffer
	strNil := b.GoString()
	assert.Equal(t, "<nil>", strNil)

	validBuf := NewBytesBuffer([]byte{1, 2, 3})
	strValid := validBuf.(*BytesBuffer).GoString()
	assert.Equal(t, "[1 2 3]", strValid)
}

func TestNewBytesBuffer(t *testing.T) {
	buf := []byte("hello")
	b := NewBytesBuffer(buf)

	ok := b != nil
	assert.True(t, ok)
}

func TestNewBytesBufferString(t *testing.T) {
	b := NewBytesBufferString("hello")

	ok := b != nil
	assert.True(t, ok)
}

func TestNewBytesBufferReader(t *testing.T) {
	r := bytes.NewReader([]byte("reader payload data"))
	b, err := NewBytesBufferReader(r)

	assert.NoError(t, err)

	ok := b != nil
	assert.True(t, ok)
}
