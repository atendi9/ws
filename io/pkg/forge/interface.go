package forge

import (
	"fmt"
	"io"
)

// Interface defines a comprehensive buffer interface that combines standard
// [io.ReadWriteSeeker], [io.ReaderFrom], [io.WriterTo], and other byte and rune
// manipulation interfaces, alongside typical buffer operations.
type Interface interface {
	// [io.ReadWriteSeeker] is embedded to provide basic Read, Write, and Seek capabilities.
	io.ReadWriteSeeker

	// [io.ReaderFrom] is embedded to allow reading data directly from an [io.Reader].
	io.ReaderFrom

	// [io.WriterTo] is embedded to allow writing data directly to an [io.Writer].
	io.WriterTo

	// [io.ByteScanner] is embedded to provide ReadByte and UnreadByte capabilities.
	io.ByteScanner

	// [io.ByteWriter] is embedded to provide WriteByte capabilities.
	io.ByteWriter

	// [io.RuneScanner] is embedded to provide ReadRune and UnreadRune capabilities.
	io.RuneScanner

	// [io.StringWriter] is embedded to provide WriteString capabilities.
	io.StringWriter

	// WriteRune appends the UTF-8 encoding of the rune to the buffer,
	// returning its length and an error, if any.
	WriteRune(rune) (int, error)

	// Bytes returns a slice of length Len holding the unread portion of the buffer.
	Bytes() []byte

	// AvailableBuffer returns an empty buffer with a capacity equal to the available space
	// in the underlying byte slice.
	AvailableBuffer() []byte

	// [fmt.Stringer] is embedded to allow the buffer to format itself as a string.
	fmt.Stringer

	// Peek returns the next n bytes without advancing the reader.
	Peek(int) ([]byte, error)

	// [fmt.GoStringer] is embedded to allow the buffer to format itself as a Go value.
	fmt.GoStringer

	// Len returns the number of bytes of the unread portion of the buffer.
	Len() int

	// Size returns the total size of the underlying buffer.
	Size() int64

	// Cap returns the capacity of the buffer's underlying byte slice.
	Cap() int

	// Available returns how many bytes are unused in the buffer.
	Available() int

	// Truncate discards all but the first n unread bytes from the buffer.
	Truncate(int)

	// Reset resets the buffer to be empty.
	Reset()

	// Grow grows the buffer's capacity, if necessary, to guarantee space for another n bytes.
	Grow(int)

	// Next returns a slice containing the next n bytes from the buffer,
	// advancing the buffer as if the bytes had been returned by a read operation.
	Next(int) []byte

	// ReadBytes reads until the first occurrence of the delimiter in the input,
	// returning a slice containing the data up to and including the delimiter.
	ReadBytes(byte) ([]byte, error)

	// ReadString reads until the first occurrence of the delimiter in the input,
	// returning a string containing the data up to and including the delimiter.
	ReadString(byte) (string, error)

	// Clone returns a copy of the current [Interface] state.
	Clone() Interface
}
