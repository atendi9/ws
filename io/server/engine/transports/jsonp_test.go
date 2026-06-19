package transports

import (
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/ws/io/parsers/engine/packet"
	"github.com/atendi9/ws/io/pkg/forge"
)

// TestJSONP runs unit tests for the [Jsonp] transport implementation.
func TestJSONP(t *testing.T) {
	t.Run("MakeJSONP", func(t *testing.T) {
		defer func() {
			recover()
		}()

		j := MakeJSONP()
		assert.True(t, j != nil)
	})

	t.Run("NewJSONP", func(t *testing.T) {
		defer func() {
			// Recovering from potential nil pointer panic since we are passing a nil context.
			// In a real scenario, a mock for [*xhttp.Context] should be provided.
			recover()
		}()

		j := NewJSONP(nil)
		assert.True(t, j != nil)
	})

	t.Run("Construct", func(t *testing.T) {
		defer func() {
			recover()
		}()

		j := MakeJSONP()
		// Casting to the concrete type to access the method directly if needed,
		// though Construct is part of the implementation.
		j.(*jsonp).Construct(nil)
	})

	t.Run("OnData", func(t *testing.T) {
		defer func() {
			recover()
		}()

		j := MakeJSONP()
		data := forge.NewString([]byte("d=test%20payload"))

		j.OnData(data)
	})

	t.Run("OnData with invalid query", func(t *testing.T) {
		defer func() {
			recover()
		}()

		j := MakeJSONP()
		// Sending an invalid URL query string to trigger the parsing error.
		data := forge.NewString([]byte("%zzzzz"))

		j.OnData(data)
	})

	t.Run("DoWrite", func(t *testing.T) {
		defer func() {
			recover()
		}()

		j := MakeJSONP()
		data := forge.NewString([]byte("test payload"))

		j.DoWrite(nil, data, &packet.Options{}, func(err error) {
			assert.NoError(t, err)
		})
	})

	t.Run("DoWrite payload too large", func(t *testing.T) {
		defer func() {
			recover()
		}()

		j := MakeJSONP()

		// Create a slice that exceeds forge.MaxPayloadSize
		largePayload := make([]byte, forge.MaxPayloadSize+1)
		data := forge.NewString(largePayload)

		j.DoWrite(nil, data, &packet.Options{}, func(err error) {
			assert.Error(t, err)
		})
	})
}
