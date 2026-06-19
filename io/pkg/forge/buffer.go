package forge

import (
	"io"
	"unicode/utf8"

	"github.com/atendi9/ws/io/pkg/errors"
)

// smallBufferSize is an initial allocation minimal capacity.
const smallBufferSize = 64

// Buffer is a variable-sized buffer of bytes.
type Buffer struct {
	lastRead readOperation
	buf      []byte
	off      int
}

// NewBuffer creates and initializes a new [Buffer] using buf as its
// initial contents.
func NewBuffer(buf []byte) *Buffer {
	return &Buffer{buf: buf}
}

// NewBufferString creates and initializes a new [Buffer] using string s as its
// initial contents.
func NewBufferString(s string) *Buffer {
	return NewBuffer([]byte(s))
}

// readOperation represents the type of the last read operation for unread functionalities.
type readOperation int

// Constants defining different types of read operations.
const (
	ReadOperation     readOperation = -1
	InvalidRead       readOperation = 0
	RuneReadOperation readOperation = 1
)

// maxInt represents the maximum integer value.
const maxInt = int(^uint(0) >> 1)

// Bytes returns a slice of length b.Len() holding the unread portion of the buffer.
func (b *Buffer) Bytes() []byte { return b.buf[b.off:] }

// AvailableBuffer returns an empty buffer with b.Available() capacity.
func (b *Buffer) AvailableBuffer() []byte { return b.buf[len(b.buf):] }

// String returns the contents of the unread portion of the buffer as a string.
func (b *Buffer) String() string {
	if b == nil {
		return "<nil>"
	}
	return string(b.buf[b.off:])
}

// Peek returns the next n bytes without advancing the reader.
// It returns [EOF] [io.EOF] if the buffer has less than n bytes.
func (b *Buffer) Peek(n int) ([]byte, error) {
	if b.Len() < n {
		return b.buf[b.off:], io.EOF
	}
	return b.buf[b.off : b.off+n], nil
}

// empty reports whether the unread portion of the buffer is empty.
func (b *Buffer) empty() bool { return len(b.buf) <= b.off }

// Len returns the number of bytes of the unread portion of the buffer.
func (b *Buffer) Len() int { return len(b.buf) - b.off }

// Cap returns the capacity of the buffer's underlying byte slice.
func (b *Buffer) Cap() int { return cap(b.buf) }

// Available returns how many bytes are unused in the buffer.
func (b *Buffer) Available() int { return cap(b.buf) - len(b.buf) }

// Truncate discards all but the first n unread bytes from the buffer.
func (b *Buffer) Truncate(n int) {
	if n == 0 {
		b.Reset()
		return
	}
	b.lastRead = InvalidRead
	if n < 0 || n > b.Len() {
		panic("bytes.Buffer: truncation out of range")
	}
	b.buf = b.buf[:b.off+n]
}

// Reset resets the buffer to be empty.
func (b *Buffer) Reset() {
	b.buf = b.buf[:0]
	b.off = 0
	b.lastRead = InvalidRead
}

// tryGrowByReslice attempts to grow the buffer by reslicing.
func (b *Buffer) tryGrowByReslice(n int) (int, bool) {
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return l, true
	}
	return 0, false
}

// grow grows the buffer to guarantee space for n more bytes.
func (b *Buffer) grow(n int) int {
	m := b.Len()
	if m == 0 && b.off != 0 {
		b.Reset()
	}
	if i, ok := b.tryGrowByReslice(n); ok {
		return i
	}
	if b.buf == nil && n <= smallBufferSize {
		b.buf = make([]byte, n, smallBufferSize)
		return 0
	}
	c := cap(b.buf)
	if n <= c/2-m {
		copy(b.buf, b.buf[b.off:])
	} else if c > maxInt-c-n {
		panic(errors.ErrTooLarge)
	} else {
		b.buf = growSlice(b.buf[b.off:], b.off+n)
	}
	b.off = 0
	b.buf = b.buf[:m+n]
	return m
}

// Grow grows the buffer's capacity, if necessary, to guarantee space for another n bytes.
func (b *Buffer) Grow(n int) {
	if n < 0 {
		panic("bytes.Buffer.Grow: negative count")
	}
	m := b.grow(n)
	b.buf = b.buf[:m]
}

// Write appends the contents of p to the buffer.
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.lastRead = InvalidRead
	m, ok := b.tryGrowByReslice(len(p))
	if !ok {
		m = b.grow(len(p))
	}
	return copy(b.buf[m:], p), nil
}

// WriteString appends the contents of s to the buffer.
func (b *Buffer) WriteString(s string) (n int, err error) {
	b.lastRead = InvalidRead
	m, ok := b.tryGrowByReslice(len(s))
	if !ok {
		m = b.grow(len(s))
	}
	return copy(b.buf[m:], s), nil
}

// MinRead is the minimum slice size passed to a Read call by [Buffer.ReadFrom].
const MinRead = 512

// ReadFrom reads data from r until [EOF] [io.EOF] and appends it to the buffer.
// It takes an [Reader] [io.Reader] as argument.
func (b *Buffer) ReadFrom(r io.Reader) (n int64, err error) {
	b.lastRead = InvalidRead
	for {
		i := b.grow(MinRead)
		b.buf = b.buf[:i]
		m, e := r.Read(b.buf[i:cap(b.buf)])
		if m < 0 {
			panic(errors.ErrNegativeRead)
		}

		b.buf = b.buf[:i+m]
		n += int64(m)
		if e == io.EOF {
			return n, nil
		}
		if e != nil {
			return n, e
		}
	}
}

// growSlice creates a new byte slice with increased capacity.
func growSlice(b []byte, n int) []byte {
	defer func() {
		if recover() != nil {
			panic(errors.ErrTooLarge)
		}
	}()
	if n < 0 || len(b) > maxInt-n {
		panic(errors.ErrTooLarge)
	}
	c := len(b) + n
	if c < 0 || c < len(b) {
		panic(errors.ErrTooLarge)
	}
	if capB := cap(b); capB <= maxInt/2 {
		if doubled := 2 * capB; doubled > c {
			c = doubled
		}
	}

	b2 := append([]byte(nil), make([]byte, c)...)
	i := copy(b2, b)
	return b2[:i]
}

// WriteTo writes data to w until the buffer is drained or an error occurs.
// It takes a [Writer] [io.Writer] as argument.
func (b *Buffer) WriteTo(w io.Writer) (n int64, err error) {
	b.lastRead = InvalidRead
	if nBytes := b.Len(); nBytes > 0 {
		m, e := w.Write(b.buf[b.off:])
		if m > nBytes {
			panic("bytes.Buffer.WriteTo: invalid Write count")
		}
		b.off += m
		n = int64(m)
		if e != nil {
			return n, e
		}
		if m != nBytes {
			return n, io.ErrShortWrite
		}
	}
	b.Reset()
	return n, nil
}

// WriteByte appends the byte c to the buffer.
func (b *Buffer) WriteByte(c byte) error {
	b.lastRead = InvalidRead
	m, ok := b.tryGrowByReslice(1)
	if !ok {
		m = b.grow(1)
	}
	b.buf[m] = c
	return nil
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to the buffer.
func (b *Buffer) WriteRune(r rune) (n int, err error) {
	if uint32(r) < utf8.RuneSelf {
		_ = b.WriteByte(byte(r))
		return 1, nil
	}
	b.lastRead = InvalidRead
	m, ok := b.tryGrowByReslice(utf8.UTFMax)
	if !ok {
		m = b.grow(utf8.UTFMax)
	}
	b.buf = utf8.AppendRune(b.buf[:m], r)
	return len(b.buf) - m, nil
}

// Read reads the next len(p) bytes from the buffer or until the buffer is drained.
func (b *Buffer) Read(p []byte) (n int, err error) {
	b.lastRead = InvalidRead
	if b.empty() {
		b.Reset()
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n = copy(p, b.buf[b.off:])
	b.off += n
	if n > 0 {
		b.lastRead = ReadOperation
	}
	return n, nil
}

// Next returns a slice containing the next n bytes from the buffer, advancing the buffer as if the bytes had been returned by [Buffer.Read].
func (b *Buffer) Next(n int) []byte {
	b.lastRead = InvalidRead
	m := b.Len()
	if n > m {
		n = m
	}
	data := b.buf[b.off : b.off+n]
	b.off += n
	if n > 0 {
		b.lastRead = ReadOperation
	}
	return data
}

// ReadByte reads and returns the next byte from the buffer.
func (b *Buffer) ReadByte() (byte, error) {
	if b.empty() {
		b.Reset()
		return 0, io.EOF
	}
	c := b.buf[b.off]
	b.off++
	b.lastRead = ReadOperation
	return c, nil
}

// ReadRune reads and returns the next UTF-8-encoded Unicode character from the buffer.
func (b *Buffer) ReadRune() (r rune, size int, err error) {
	if b.empty() {
		b.Reset()
		return 0, 0, io.EOF
	}
	c := b.buf[b.off]
	if c < utf8.RuneSelf {
		b.off++
		b.lastRead = RuneReadOperation
		return rune(c), 1, nil
	}
	r, n := utf8.DecodeRune(b.buf[b.off:])
	b.off += n
	b.lastRead = readOperation(n)
	return r, n, nil
}

// UnreadRune unreads the last rune returned by [Buffer.ReadRune].
func (b *Buffer) UnreadRune() error {
	if b.lastRead <= InvalidRead {
		return errors.ErrUnreadRune
	}
	if b.off >= int(b.lastRead) {
		b.off -= int(b.lastRead)
	}
	b.lastRead = InvalidRead
	return nil
}

// UnreadByte unreads the last byte returned by the most recent successful read operation.
func (b *Buffer) UnreadByte() error {
	if b.lastRead == InvalidRead {
		return errors.ErrUnreadByte
	}
	b.lastRead = InvalidRead
	if b.off > 0 {
		b.off--
	}
	return nil
}

// ReadBytes reads until the first occurrence of delim in the input.
func (b *Buffer) ReadBytes(delim byte) (line []byte, err error) {
	slice, err := b.readSlice(delim)
	line = append(line, slice...)
	return line, err
}

// readSlice reads until the first occurrence of delim in the input and returns a slice pointing at the bytes in the buffer.
func (b *Buffer) readSlice(delim byte) (line []byte, err error) {
	i := IndexByte(b.buf[b.off:], delim)
	end := b.off + i + 1
	if i < 0 {
		end = len(b.buf)
		err = io.EOF
	}
	line = b.buf[b.off:end]
	b.off = end
	b.lastRead = ReadOperation
	return line, err
}

// ReadString reads until the first occurrence of delim in the input and returns a string.
func (b *Buffer) ReadString(delim byte) (line string, err error) {
	slice, err := b.readSlice(delim)
	return string(slice), err
}

// Clone returns a deep copy of the [Buffer].
func (b *Buffer) Clone() *Buffer {
	if b == nil {
		return nil
	}

	clone := &Buffer{
		buf:      make([]byte, len(b.buf), cap(b.buf)),
		off:      b.off,
		lastRead: b.lastRead,
	}

	copy(clone.buf, b.buf)

	return clone
}

// Size returns the total length of the buffer's underlying byte slice.
func (b *Buffer) Size() int64 { return int64(len(b.buf)) }

// Seek sets the offset for the next Read or Write on the buffer.
func (b *Buffer) Seek(offset int64, whence int) (int64, error) {
	b.lastRead = InvalidRead
	var abs int
	switch whence {
	case io.SeekStart:
		abs = int(offset)
	case io.SeekCurrent:
		abs = b.off + int(offset)
	case io.SeekEnd:
		abs = len(b.buf) + int(offset)
	default:
		return 0, errors.ErrTypesBufferSeekInvalidWhence
	}
	if abs < 0 || abs > len(b.buf) {
		return 0, errors.ErrTypesBufferSeekNegativePosition
	}
	b.off = abs
	return int64(abs), nil
}
