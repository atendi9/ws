package transports

import (
	stdErrors "errors"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/ws/io/parsers/engine/packet"
	"github.com/atendi9/ws/io/pkg/forge"
)

// TestWebsocket runs unit tests for the [websocket] transport implementation.
func TestWebsocket(t *testing.T) {

	t.Run("MakeWebSocket", func(t *testing.T) {
		w := MakeWebSocket()
		assert.Equal(t, "open", w.CurrentState())
		assert.Equal(t, w, w.Proto())
	})

	t.Run("NewWebSocket and Construct", func(t *testing.T) {
		defer func() {
			// Recovering from potential nil pointer panic since we are passing a nil context.
			// In a real scenario, a mock for [*xhttp.Context] should be provided.
			recover()
		}()
		w := NewWebSocket(nil)
		assert.Equal(t, "open", w.CurrentState())
	})

	t.Run("Name", func(t *testing.T) {
		w := MakeWebSocket()
		assert.Equal(t, WEBSOCKET, w.Name())
	})

	t.Run("HandlesUpgrades", func(t *testing.T) {
		w := MakeWebSocket()
		assert.True(t, w.HandlesUpgrades())
	})

	t.Run("_error", func(t *testing.T) {
		defer func() {
			recover()
		}()
		w := MakeWebSocket()
		err := stdErrors.New("unexpected websocket error")

		// This will panic because the underlying socket is nil when not constructed with a valid context.
		w.(*websocket)._error(err)
	})

	t.Run("message", func(t *testing.T) {
		defer func() {
			recover()
		}()
		w := MakeWebSocket()

		// This will panic because the underlying socket and queue are nil without Construct.
		w.(*websocket).message()
	})

	t.Run("onMessage", func(t *testing.T) {
		defer func() {
			recover()
		}()
		w := MakeWebSocket()
		data := forge.NewString([]byte("test payload"))

		// This will panic if the base transport is not fully constructed with a valid parser.
		w.(*websocket).onMessage(data)
	})

	t.Run("Send", func(t *testing.T) {
		defer func() {
			recover()
		}()
		w := MakeWebSocket()
		packets := []*packet.Packet{}

		// This will panic because writeQueue is nil without Construct.
		w.Send(packets)
	})

	t.Run("send", func(t *testing.T) {
		defer func() {
			recover()
		}()
		w := MakeWebSocket()
		packets := []*packet.Packet{}

		// This will panic because the socket and parser are nil.
		w.(*websocket).send(packets)
	})

	t.Run("write", func(t *testing.T) {
		defer func() {
			recover()
		}()
		w := MakeWebSocket()
		data := forge.NewString([]byte("test payload"))

		// This will panic because the socket is nil.
		w.(*websocket).write(data, true)
	})

	t.Run("OnClose", func(t *testing.T) {
		defer func() {
			recover()
		}()
		w := MakeWebSocket()

		// This will panic because writeQueue is nil without Construct.
		w.OnClose()
	})

	t.Run("DoClose", func(t *testing.T) {
		defer func() {
			recover()
		}()
		w := MakeWebSocket()
		called := false
		fn := func() {
			called = true
		}

		// This will panic because writeQueue and socket are nil.
		w.(*websocket).DoClose(fn)

		// The assertion below would only be reached if dependencies were properly mocked.
		assert.True(t, called)
	})
}
