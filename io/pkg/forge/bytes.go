package forge

import (
	"fmt"
	"io"
)

// MaxPayloadSize is the upper bound (128 MiB) for a single encoded payload.
// It prevents unbounded allocations from untrusted input.
const MaxPayloadSize = 128 * 1024 * 1024

// BytesBuffer represents a wrapper around [Buffer] that implements the [Interface].
type BytesBuffer struct {
	*Buffer
}

// Clone creates and returns a deep copy of the [BytesBuffer].
// It returns an [Interface]. If the [BytesBuffer] or its underlying [Buffer] is nil, it returns nil.
func (b *BytesBuffer) Clone() Interface {
	if b == nil || b.Buffer == nil {
		return nil
	}
	return &BytesBuffer{b.Buffer.Clone()}
}

// GoString implements the [fmt.GoStringer] interface for [BytesBuffer].
// It returns the string representation of the underlying byte slice.
// If the [BytesBuffer] or its underlying [Buffer] is nil, it returns "<nil>".
func (b *BytesBuffer) GoString() string {
	if b == nil || b.Buffer == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", b.Bytes())
}

// NewBytesBuffer creates and returns a new [BytesBuffer] initialized with the provided byte slice.
// It returns an [Interface].
func NewBytesBuffer(buf []byte) Interface {
	return &BytesBuffer{NewBuffer(buf)}
}

// NewBytesBufferString creates and returns a new [BytesBuffer] initialized with the provided string.
// It returns an [Interface].
func NewBytesBufferString(s string) Interface {
	return &BytesBuffer{NewBufferString(s)}
}

// NewBytesBufferReader reads from the provided [io.Reader] and creates a new [BytesBuffer] containing the read data.
// It returns an [Interface] and any error encountered during reading.
func NewBytesBufferReader(r io.Reader) (Interface, error) {
	b := NewBytesBuffer(nil)
	_, err := b.ReadFrom(r)
	return b, err
}
