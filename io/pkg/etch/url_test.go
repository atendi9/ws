package etch

import (
	"testing"

	pkgerrors "github.com/atendi9/ws/io/pkg/errors"

	"github.com/atendi9/capivara/assert"
)

func TestUrl(t *testing.T) {
	t.Run("empty uri returns ErrEmptyURI", func(t *testing.T) {
		parsed, err := Url("", "/")
		assert.Error(t, err)
		assert.True(t, parsed == nil)
		assert.True(t, errorsIs(err, pkgerrors.ErrEmptyURI))
	})

	t.Run("unparseable uri returns error", func(t *testing.T) {
		parsed, err := Url("://bad uri", "/")
		assert.Error(t, err)
		assert.True(t, parsed == nil)
	})

	t.Run("missing scheme returns ErrUnsupportedScheme", func(t *testing.T) {
		parsed, err := Url("example.com/path", "/")
		assert.Error(t, err)
		assert.True(t, parsed == nil)
		assert.True(t, errorsIs(err, pkgerrors.ErrUnsupportedScheme))
	})

	t.Run("unsupported scheme returns ErrUnsupportedScheme", func(t *testing.T) {
		parsed, err := Url("ftp://example.com", "/")
		assert.Error(t, err)
		assert.True(t, parsed == nil)
		assert.True(t, errorsIs(err, pkgerrors.ErrUnsupportedScheme))
	})

	t.Run("http defaults to port 80 and root path", func(t *testing.T) {
		parsed, err := Url("http://example.com", "/socket.io")
		assert.NoError(t, err)
		assert.NotNil(t, parsed)
		assert.Equal(t, "example.com", parsed.Hostname)
		assert.Equal(t, "80", parsed.Port)
		assert.Equal(t, "/", parsed.Path)
		assert.Equal(t, "http://example.com:80/socket.io", parsed.Id)
	})

	t.Run("https defaults to port 443", func(t *testing.T) {
		parsed, err := Url("https://example.com", "")
		assert.NoError(t, err)
		assert.Equal(t, "443", parsed.Port)
	})

	t.Run("ws defaults to port 80", func(t *testing.T) {
		parsed, err := Url("ws://example.com", "")
		assert.NoError(t, err)
		assert.Equal(t, "80", parsed.Port)
	})

	t.Run("wss defaults to port 443", func(t *testing.T) {
		parsed, err := Url("wss://example.com", "")
		assert.NoError(t, err)
		assert.Equal(t, "443", parsed.Port)
	})

	t.Run("explicit port is preserved", func(t *testing.T) {
		parsed, err := Url("http://example.com:3000/chat", "/path")
		assert.NoError(t, err)
		assert.Equal(t, "3000", parsed.Port)
		assert.Equal(t, "/chat", parsed.Path)
		assert.Equal(t, "http://example.com:3000/path", parsed.Id)
	})
}

// errorsIs is a tiny local wrapper to keep the import list intentional.
func errorsIs(err, target error) bool {
	for e := err; e != nil; {
		if e == target {
			return true
		}
		u, ok := e.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		e = u.Unwrap()
	}
	return false
}
