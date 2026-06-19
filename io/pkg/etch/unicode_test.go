package etch

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestUtf16Len(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want int
	}{
		{"ascii", 'a', 1},
		{"bmp", '€', 1},
		{"just below surrogate self", surrSelf - 1, 1},
		{"supplementary plane", '😀', 2},
		{"max rune", maxRune, 2},
		{"out of range", maxRune + 1, 1},
		{"negative", -1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Utf16Len(tt.r))
		})
	}
}

func TestUtf16Count(t *testing.T) {
	assert.Equal(t, 0, Utf16Count(nil))
	assert.Equal(t, 5, Utf16Count([]byte("hello")))
	assert.Equal(t, 2, Utf16Count([]byte("😀")))            // one supplementary rune -> 2
	assert.Equal(t, 3, Utf16Count([]byte("a😀")))           // 1 + 2
	assert.Equal(t, 1, Utf16Count([]byte{0xff}))           // invalid byte -> RuneError -> 1
}

func TestUtf16CountString(t *testing.T) {
	assert.Equal(t, 0, Utf16CountString(""))
	assert.Equal(t, 5, Utf16CountString("hello"))
	assert.Equal(t, 2, Utf16CountString("😀"))
	assert.Equal(t, 3, Utf16CountString("a😀"))
}

func TestUtf8encodeString(t *testing.T) {
	assert.Equal(t, "", Utf8encodeString(""))
	assert.Equal(t, "abc", Utf8encodeString("abc"))
	// byte 0xe9 (é in latin1) -> encoded as the rune U+00E9
	assert.Equal(t, "é", Utf8encodeString("\xe9"))
}

func TestUtf8encodeBytes(t *testing.T) {
	assert.True(t, Utf8encodeBytes(nil) == nil)
	assert.Equal(t, "abc", string(Utf8encodeBytes([]byte("abc"))))
	assert.Equal(t, "é", string(Utf8encodeBytes([]byte{0xe9})))
}

func TestUtf8decodeString(t *testing.T) {
	assert.Equal(t, "", Utf8decodeString(""))
	assert.Equal(t, "abc", Utf8decodeString("abc"))
	// round-trip: encode then decode returns original bytes
	encoded := Utf8encodeString("\xe9")
	assert.Equal(t, "\xe9", Utf8decodeString(encoded))
}

func TestUtf8decodeBytes(t *testing.T) {
	assert.True(t, Utf8decodeBytes(nil) == nil)
	assert.Equal(t, "abc", string(Utf8decodeBytes([]byte("abc"))))
	encoded := Utf8encodeBytes([]byte{0xe9})
	assert.Equal(t, string([]byte{0xe9}), string(Utf8decodeBytes(encoded)))
}

func TestUtf8EncoderRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	enc := NewUtf8Encoder(&buf)
	assert.NotNil(t, enc)

	input := []byte("hello world")
	n, err := enc.Write(input)
	assert.NoError(t, err)
	assert.Equal(t, len(input), n)

	// decoding the encoded stream returns the original bytes
	dec := NewUtf8Decoder(&buf)
	out, err := io.ReadAll(dec)
	assert.NoError(t, err)
	assert.Equal(t, string(input), string(out))
}

func TestUtf8EncoderLargeInput(t *testing.T) {
	// Larger than bufferSize/2 to exercise the chunking loop.
	input := bytes.Repeat([]byte("a"), bufferSize*3)
	var buf bytes.Buffer
	enc := NewUtf8Encoder(&buf)
	n, err := enc.Write(input)
	assert.NoError(t, err)
	assert.Equal(t, len(input), n)

	dec := NewUtf8Decoder(&buf)
	out, err := io.ReadAll(dec)
	assert.NoError(t, err)
	assert.Equal(t, len(input), len(out))
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestUtf8EncoderWriteError(t *testing.T) {
	enc := NewUtf8Encoder(errWriter{})
	_, err := enc.Write([]byte("data"))
	assert.Error(t, err)
}

func TestUtf8DecoderEmptyRead(t *testing.T) {
	dec := NewUtf8Decoder(strings.NewReader("data"))
	n, err := dec.Read(nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

func TestUtf8DecoderReadError(t *testing.T) {
	dec := NewUtf8Decoder(errReader{})
	_, err := dec.Read(make([]byte, 4))
	assert.Error(t, err)
}
