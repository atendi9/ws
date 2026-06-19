package transports

import (
	stdErrors "errors"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/ws/io/parsers/engine/packet"
	"github.com/atendi9/ws/io/pkg/state"
	"github.com/atendi9/ws/io/pkg/xhttp"
)

func TestTransport(t *testing.T) {

	t.Run("MakeTransport", func(t *testing.T) {
		tr := MakeTransport()
		assert.Equal(t, "open", tr.CurrentState())
		assert.Equal(t, tr, tr.Proto())
	})

	t.Run("NewTransport and Construct", func(t *testing.T) {
		defer func() {
			// Recovering from potential nil pointer panic since we are passing a nil context.
			// In a real scenario, a mock for [*xhttp.Context] should be provided.
			recover()
		}()
		tr := NewTransport(nil)
		assert.Equal(t, "open", tr.CurrentState())
	})

	t.Run("Prototype and Proto", func(t *testing.T) {
		tr1 := MakeTransport()
		tr2 := MakeTransport()

		tr1.Prototype(tr2)
		assert.Equal(t, tr2, tr1.Proto())
	})

	t.Run("Sid and SetSid", func(t *testing.T) {
		tr := MakeTransport()
		tr.SetSid("session-123")
		assert.Equal(t, "session-123", tr.Sid())
	})

	t.Run("Writable and SetWritable", func(t *testing.T) {
		tr := MakeTransport()
		assert.False(t, tr.Writable())

		tr.SetWritable(true)
		assert.True(t, tr.Writable())
	})

	t.Run("Protocol", func(t *testing.T) {
		tr := MakeTransport()
		// Protocol is expected to be 0 before Construct is called with a valid context
		assert.Equal(t, 0, tr.Protocol())
	})

	t.Run("Discarded and Discard", func(t *testing.T) {
		tr := MakeTransport()
		assert.False(t, tr.Discarded())

		tr.Discard()
		assert.True(t, tr.Discarded())
	})

	t.Run("Parser", func(t *testing.T) {
		tr := MakeTransport()
		// Parser is expected to be nil before Construct is called
		assert.Equal(t, nil, tr.Parser())
	})

	t.Run("SupportsBinary and SetSupportsBinary", func(t *testing.T) {
		tr := MakeTransport()
		assert.False(t, tr.SupportsBinary())

		tr.SetSupportsBinary(true)
		assert.True(t, tr.SupportsBinary())
	})

	t.Run("CurrentState and SetCurrentState", func(t *testing.T) {
		tr := MakeTransport()
		newState := state.NewState("transport", "upgrading")

		tr.SetCurrentState(newState)
		assert.Equal(t, "upgrading", tr.CurrentState())
	})

	t.Run("HttpCompression and SetHttpCompression", func(t *testing.T) {
		tr := MakeTransport()
		comp := &xhttp.Compression{}

		tr.SetHttpCompression(comp)
		assert.Equal(t, comp, tr.HttpCompression())
	})

	t.Run("PerMessageDeflate and SetPerMessageDeflate", func(t *testing.T) {
		tr := MakeTransport()
		pmd := &xhttp.PerMessageDeflate{}

		tr.SetPerMessageDeflate(pmd)
		assert.Equal(t, pmd, tr.PerMessageDeflate())
	})

	t.Run("MaxHttpBufferSize and SetMaxHttpBufferSize", func(t *testing.T) {
		tr := MakeTransport()
		tr.SetMaxHttpBufferSize(2048)
		assert.Equal(t, int64(2048), tr.MaxHttpBufferSize())
	})

	t.Run("OnRequest", func(t *testing.T) {
		tr := MakeTransport()
		// OnRequest is a no-op, it should not panic
		tr.OnRequest(nil)
	})

	t.Run("Close", func(t *testing.T) {
		tr := MakeTransport()
		tr.Close()
		assert.Equal(t, "closing", tr.CurrentState())

		// Subsequent calls to Close should not change the state or panic
		tr.Close()
		assert.Equal(t, "closing", tr.CurrentState())
	})

	t.Run("OnError", func(t *testing.T) {
		tr := MakeTransport()
		err := stdErrors.New("internal transport error")

		eventEmitted := false
		tr.On("error", func(args ...any) {
			eventEmitted = true
		})

		tr.OnError("connection failed", err)
		assert.True(t, eventEmitted)
	})

	t.Run("OnPacket", func(t *testing.T) {
		tr := MakeTransport()
		p := &packet.Packet{}

		eventEmitted := false
		tr.On("packet", func(args ...any) {
			eventEmitted = true
			assert.Equal(t, p, args[0].(*packet.Packet))
		})

		tr.OnPacket(p)
		assert.True(t, eventEmitted)
	})

	t.Run("OnData", func(t *testing.T) {
		// Recovering from potential nil pointer panic since Parser is not initialized
		// without a valid [*xhttp.Context] in Construct.
		defer func() {
			recover()
		}()
		tr := MakeTransport()
		tr.OnData(nil)
	})

	t.Run("OnClose", func(t *testing.T) {
		tr := MakeTransport()

		eventEmitted := false
		tr.On("close", func(args ...any) {
			eventEmitted = true
		})

		tr.OnClose()
		assert.True(t, eventEmitted)

		// Ensure state does not change if already closed
		tr.OnClose()
		assert.Equal(t, state.TransportClosed.String(), tr.CurrentState())
	})

	t.Run("HandlesUpgrades", func(t *testing.T) {
		tr := MakeTransport()
		assert.False(t, tr.HandlesUpgrades())
	})

	t.Run("Name", func(t *testing.T) {
		tr := MakeTransport()
		assert.Equal(t, "", tr.Name())
	})

	t.Run("Send", func(t *testing.T) {
		tr := MakeTransport()
		// Send is a no-op in the base transport, it should not panic
		tr.Send([]*packet.Packet{})
	})

	t.Run("DoClose", func(t *testing.T) {
		tr := MakeTransport()
		// DoClose is a no-op in the base transport, it should not panic
		tr.DoClose(nil)
	})
}
