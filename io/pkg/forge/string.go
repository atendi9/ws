// Package forge provides byte buffer implementations for buffer management,
// reading, and JSON serialization utilizing an internal [Buffer].
package forge

import (
	"encoding/json"
	"io"
)

// NewFromString creates a new [Interface] utilizing a string as the initial
// buffer content.
func NewFromString(s string) Interface {
	return &String{NewBufferString(s)}
}

// NewStringReader reads all contents from an [io.Reader] and creates a
// new [Interface] with the read data.
func NewStringReader(r io.Reader) (Interface, error) {
	b := NewString(nil)
	_, err := b.(*String).ReadFrom(r)
	return b, err
}

// NewString creates a new [Interface] from a provided byte slice.
func NewString(buf []byte) Interface {
	return &String{NewBuffer(buf)}
}

// String represents a specialized buffer that implements [Interface]
// and provides JSON marshaling and unmarshaling capabilities wrapped around
// a base [*Buffer].
type String struct {
	*Buffer
}

// Clone creates a deep copy of the [String] and returns it as an [Interface].
// If the current buffer or its underlying [*Buffer] is nil, it returns nil.
func (sb *String) Clone() Interface {
	if sb == nil || sb.Buffer == nil {
		return nil
	}
	return &String{sb.Buffer.Clone()}
}

// GoString returns the string representation of the [String].
// It returns "<nil>" if the buffer is uninitialized.
func (sb *String) GoString() string {
	if sb == nil || sb.Buffer == nil {
		return "<nil>"
	}
	return sb.String()
}

// MarshalJSON encodes the [String] content into a valid JSON string.
// If the buffer is nil, it returns a JSON representation of an empty string.
func (sb *String) MarshalJSON() ([]byte, error) {
	if sb == nil || sb.Buffer == nil {
		return []byte(`""`), nil
	}

	return json.Marshal(sb.String())
}

// UnmarshalJSON decodes a JSON payload into the [String].
// It initializes the underlying [*Buffer] with the parsed string.
func (sb *String) UnmarshalJSON(data []byte) error {
	if sb == nil {
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	sb.Buffer = NewBufferString(str)
	return nil
}
