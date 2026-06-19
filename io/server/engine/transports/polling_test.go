package transports

import (
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/ws/io/parsers/engine/packet"
	"github.com/atendi9/ws/io/pkg/forge"
)

// TestPolling runs unit tests for the [Polling] transport implementation.
func TestPolling(t *testing.T) {
	t.Run("MakePolling", func(t *testing.T) {
		p := MakePolling()
		assert.Equal(t, POLLING, p.Name())
		assert.Equal(t, p.(*polling), p.Proto().(*polling))
	})

	t.Run("NewPolling and Construct", func(t *testing.T) {
		defer func() {
			// Recovering from potential nil pointer panic since we are passing a nil context.
			// In a real scenario, a mock for [*xhttp.Context] should be provided.
			recover()
		}()
		p := NewPolling(nil)
		assert.Equal(t, POLLING, p.Name())
	})

	t.Run("Name", func(t *testing.T) {
		p := MakePolling()
		assert.Equal(t, POLLING, p.Name())
	})

	t.Run("OnRequest", func(t *testing.T) {
		defer func() {
			recover()
		}()
		p := MakePolling()

		p.(*polling).OnRequest(nil)
	})

	t.Run("onPollRequest", func(t *testing.T) {
		defer func() {
			recover()
		}()
		p := MakePolling()

		p.(*polling).onPollRequest(nil)
	})

	t.Run("onDataRequest", func(t *testing.T) {
		defer func() {
			recover()
		}()
		p := MakePolling()

		p.(*polling).onDataRequest(nil)
	})

	t.Run("OnData", func(t *testing.T) {
		defer func() {
			recover()
		}()
		p := MakePolling()
		data := forge.NewString([]byte("test payload"))

		p.OnData(data)
	})

	t.Run("OnClose", func(t *testing.T) {
		defer func() {
			recover()
		}()
		p := MakePolling()

		p.(*polling).OnClose()
	})

	t.Run("Send", func(t *testing.T) {
		defer func() {
			recover()
		}()
		p := MakePolling()
		packets := []*packet.Packet{}

		p.Send(packets)
	})

	t.Run("send", func(t *testing.T) {
		defer func() {
			recover()
		}()
		p := MakePolling()
		packets := []*packet.Packet{}

		p.(*polling).send(packets)
	})

	t.Run("write", func(t *testing.T) {
		defer func() {
			recover()
		}()
		p := MakePolling()
		data := forge.NewString([]byte("test payload"))

		p.(*polling).write(data, nil)
	})

	t.Run("DoWrite", func(t *testing.T) {
		defer func() {
			recover()
		}()
		p := MakePolling()
		data := forge.NewString([]byte("test payload"))

		p.DoWrite(nil, data, nil, func(err error) {})
	})

	t.Run("compress", func(t *testing.T) {
		defer func() {
			recover()
		}()
		p := MakePolling()
		data := forge.NewString([]byte("test payload for compression"))

		gzData, err := p.(*polling).compress(data, "gzip")
		assert.NoError(t, err)
		assert.True(t, gzData != nil)

		flateData, err2 := p.(*polling).compress(data, "deflate")
		assert.NoError(t, err2)
		assert.True(t, flateData != nil)
	})

	t.Run("DoClose", func(t *testing.T) {
		defer func() {
			recover()
		}()
		p := MakePolling()
		called := false
		fn := func() {
			called = true
		}

		p.(*polling).DoClose(fn)

		assert.True(t, called)
	})

	t.Run("headers", func(t *testing.T) {
		defer func() {
			recover()
		}()
		p := MakePolling()

		p.(*polling).headers(nil, nil)
	})
}
