package websocket

import (
	"encoding/binary"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/atendi9/ws/io/pkg/events"
	"github.com/atendi9/ws/io/pkg/websocket/wsconn"

	"github.com/atendi9/capivara/assert"
)

func TestMessageTypeConstants(t *testing.T) {
	assert.Equal(t, wsconn.TextMessage, TextMessage)
	assert.Equal(t, wsconn.BinaryMessage, BinaryMessage)
	assert.Equal(t, wsconn.PingMessage, PingMessage)
	assert.Equal(t, wsconn.PongMessage, PongMessage)
	assert.Equal(t, wsconn.CloseMessage, CloseMessage)
}

func TestFormatCloseMessage(t *testing.T) {
	t.Run("normal code with text", func(t *testing.T) {
		msg := FormatCloseMessage(wsconn.CloseNormalClosure, "bye")
		// 2-byte big-endian code + text
		assert.Equal(t, 2+len("bye"), len(msg))
		assert.Equal(t, uint16(wsconn.CloseNormalClosure), binary.BigEndian.Uint16(msg[:2]))
		assert.Equal(t, "bye", string(msg[2:]))
	})

	t.Run("no status received yields empty payload", func(t *testing.T) {
		msg := FormatCloseMessage(wsconn.CloseNoStatusReceived, "ignored")
		assert.Equal(t, 0, len(msg))
	})
}

func TestNewPreparedMessage(t *testing.T) {
	pm, err := NewPreparedMessage(TextMessage, []byte("hello"))
	assert.NoError(t, err)
	assert.NotNil(t, pm)
}

// wsTestServer stands up an echo websocket server and returns its ws:// URL.
func wsTestServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	up := &wsconn.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := up.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		for {
			mt, rd, err := c.NextReader()
			if err != nil {
				return
			}
			data, _ := io.ReadAll(rd)
			if err := c.WriteMessage(mt, data); err != nil {
				return
			}
		}
	}))
	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	return srv, url
}

func dialWrapped(t *testing.T, url string) *Conn {
	t.Helper()
	d := &wsconn.Dialer{HandshakeTimeout: 5 * time.Second}
	raw, _, err := d.Dial(url, nil)
	assert.NoError(t, err)
	return &Conn{Emitter: events.NewEmitter(), Conn: raw}
}

func TestConnEchoAndClose(t *testing.T) {
	srv, url := wsTestServer(t)
	defer srv.Close()

	conn := dialWrapped(t, url)

	// Round-trip a message through the wrapped connection.
	err := conn.WriteMessage(TextMessage, []byte("ping"))
	assert.NoError(t, err)

	mt, rd, err := conn.NextReader()
	assert.NoError(t, err)
	data, _ := io.ReadAll(rd)
	assert.Equal(t, TextMessage, mt)
	assert.Equal(t, "ping", string(data))

	// Close fires the "close" event via the embedded emitter.
	var closed atomic.Bool
	conn.On("close", func(...any) { closed.Store(true) })
	assert.NoError(t, conn.Close())
	assert.True(t, closed.Load())
}

func TestConnWritePreparedMessage(t *testing.T) {
	srv, url := wsTestServer(t)
	defer srv.Close()

	conn := dialWrapped(t, url)
	defer conn.Close()

	pm, err := NewPreparedMessage(TextMessage, []byte("prepared"))
	assert.NoError(t, err)
	assert.NoError(t, conn.WritePreparedMessage(pm))

	mt, rd, err := conn.NextReader()
	assert.NoError(t, err)
	data, _ := io.ReadAll(rd)
	assert.Equal(t, TextMessage, mt)
	assert.Equal(t, "prepared", string(data))
}
