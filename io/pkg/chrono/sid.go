// Package chrono provides time-related primitives — timers ([SetTimeout],
// [SetInterval]) and the yeast monotonic clock — together with utilities for
// generating and validating secure, unique session identifiers.
package chrono

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"sync"
	"sync/atomic"
)

// base64Id is a generator that produces unique, thread-safe identifiers
// utilizing random bytes and a monotonic sequence counter.
type base64Id struct {
	sequenceNumber atomic.Uint64
	pool           sync.Pool
}

var bid = &base64Id{
	pool: sync.Pool{
		New: func() any {
			// 18 bytes for the raw ID + 24 bytes for the base64 output buffer
			b := make([]byte, 42)
			return &b
		},
	},
}

// New returns the global singleton instance of [*base64Id] used for
// generating identifiers.
func New() *base64Id {
	return bid
}

// String implements the [fmt.Stringer] interface.
// It calls [*base64Id.Generate] to return the unique string representation of the ID.
func (b *base64Id) String() string {
	return b.Generate()
}

// Generate creates a unique 24-character string representation of an ID.
// It combines 10 random bytes with an 8-byte monotonically increasing sequence
// number, encoded using [base64.RawURLEncoding]. This method is concurrent-safe.
func (b *base64Id) Generate() string {
	bufPtr := b.pool.Get().(*[]byte)
	buf := *bufPtr

	// buf[0:18] is used for raw bytes, buf[18:42] is used for base64 output
	raw := buf[:18]
	_, _ = rand.Read(raw[:10])
	binary.BigEndian.PutUint64(raw[10:], b.sequenceNumber.Add(1)-1)

	base64.RawURLEncoding.Encode(buf[18:42], raw)
	id := string(buf[18:42])

	b.pool.Put(bufPtr)
	return id
}

// IsValid checks whether the given session ID has a safe format.
//   - Valid characters: alphanumeric, '-', '_', '.', '#', ':' (for protocol v3 namespace#id format).
//   - Maximum length: 36 characters.
func IsValid(sid string) bool {
	if len(sid) == 0 || len(sid) > 36 {
		return false
	}
	// Loop byte by byte avoiding UTF-8 decoding overhead
	for _, c := range sid {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' || c == '#' || c == ':' {
			continue
		}
		return false
	}
	return true
}
