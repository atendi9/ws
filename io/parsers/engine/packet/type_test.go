package packet

import (
	"io"
	"strings"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestType(t *testing.T) {
	t.Run("String representation", func(t *testing.T) {
		assert.Equal(t, "open", OPEN.String())
		assert.Equal(t, "message", MESSAGE.String())
	})

	t.Run("IsValid checks", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			assert.True(t, OPEN.IsValid())
			assert.True(t, CLOSE.IsValid())
			assert.True(t, PING.IsValid())
			assert.True(t, PONG.IsValid())
			assert.True(t, MESSAGE.IsValid())
			assert.True(t, UPGRADE.IsValid())
			assert.True(t, NOOP.IsValid())
		})
		t.Run("Invalid", func(t *testing.T) {
			assert.False(t, ERROR.IsValid())
			assert.False(t, Type("unknown").IsValid())
		})
	})
}

func TestOptions(t *testing.T) {
	t.Run("NewOptions initialization", func(t *testing.T) {
		opts := NewOptions(true)

		assert.True(t, *opts.Compress)
	})
}

func TestPacket(t *testing.T) {
	packetMsg := "hello world"
	t.Run("New packet creation", func(t *testing.T) {
		data := strings.NewReader(packetMsg)
		p := New(MESSAGE, data)

		assert.Equal(t, MESSAGE, p.Type)
		b, _ := io.ReadAll(p.Data)
		assert.Equal(t, packetMsg, string(b))
	})

	t.Run("NewWithOptions packet creation", func(t *testing.T) {
		data := strings.NewReader(packetMsg)
		opts := NewOptions(false)

		p := NewWithOptions(MESSAGE, data, opts)

		assert.Equal(t, MESSAGE, p.Type)
		b, _ := io.ReadAll(p.Data)
		assert.Equal(t, packetMsg, string(b))
		assert.Equal(t, opts, p.Options)
	})
}
