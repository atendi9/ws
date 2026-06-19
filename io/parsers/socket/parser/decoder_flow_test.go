package parser

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

// collect subscribes to the "decoded" event and records emitted packets.
func collect(d Decoder) *[]*Packet {
	var packets []*Packet
	d.On("decoded", func(args ...any) {
		if len(args) > 0 {
			if p, ok := args[0].(*Packet); ok {
				packets = append(packets, p)
			}
		}
	})
	return &packets
}

func TestDecoderConnect(t *testing.T) {
	d := NewDecoder()
	got := collect(d)

	// CONNECT (type 0) to the default namespace.
	assert.NoError(t, d.Add("0"))
	assert.Equal(t, 1, len(*got))
	assert.Equal(t, CONNECT, (*got)[0].Type)
}

func TestDecoderEventWithNamespaceAndID(t *testing.T) {
	d := NewDecoder()
	got := collect(d)

	// EVENT (type 2), namespace /admin, ack id 12, payload ["hello","world"].
	assert.NoError(t, d.Add(`2/admin,12["hello","world"]`))
	assert.Equal(t, 1, len(*got))

	pkt := (*got)[0]
	assert.Equal(t, EVENT, pkt.Type)
	assert.Equal(t, "/admin", pkt.Nsp)
	assert.NotNil(t, pkt.Id)
	assert.Equal(t, uint64(12), *pkt.Id)

	data, ok := pkt.Data.([]any)
	assert.True(t, ok)
	assert.Equal(t, "hello", data[0])
}

func TestDecoderAckPayload(t *testing.T) {
	d := NewDecoder()
	got := collect(d)

	// ACK (type 3) with id 7 and a slice payload.
	assert.NoError(t, d.Add(`37[42]`))
	assert.Equal(t, 1, len(*got))
	assert.Equal(t, ACK, (*got)[0].Type)
}

func TestDecoderBinaryEventReconstruction(t *testing.T) {
	d := NewDecoder()
	got := collect(d)

	// BINARY_EVENT (type 5) with 1 attachment and a placeholder payload.
	header := `51-["file",{"_placeholder":true,"num":0}]`
	assert.NoError(t, d.Add(header))
	// Not yet complete: still awaiting the binary attachment.
	assert.Equal(t, 0, len(*got))

	// Supplying the binary attachment completes reconstruction.
	assert.NoError(t, d.Add([]byte("binary-bytes")))
	assert.Equal(t, 1, len(*got))
	assert.Equal(t, BINARY_EVENT, (*got)[0].Type)
}

func TestDecoderBinaryEventZeroAttachments(t *testing.T) {
	d := NewDecoder()
	got := collect(d)

	// BINARY_EVENT with 0 attachments emits immediately.
	assert.NoError(t, d.Add(`50-["evt"]`))
	assert.Equal(t, 1, len(*got))
}

func TestDecoderErrors(t *testing.T) {
	t.Run("unknown type byte", func(t *testing.T) {
		d := NewDecoder()
		assert.Error(t, d.Add("9"))
	})

	t.Run("empty buffer", func(t *testing.T) {
		d := NewDecoder()
		assert.Error(t, d.Add(""))
	})

	t.Run("binary without reconstruction", func(t *testing.T) {
		d := NewDecoder()
		assert.Error(t, d.Add([]byte("orphan")))
	})

	t.Run("plaintext during reconstruction", func(t *testing.T) {
		d := NewDecoder()
		// Start a binary packet expecting an attachment...
		assert.NoError(t, d.Add(`51-["file",{"_placeholder":true,"num":0}]`))
		// ...then send plaintext before the attachment: must error.
		assert.Error(t, d.Add("0"))
	})

	t.Run("invalid payload json", func(t *testing.T) {
		d := NewDecoder()
		assert.Error(t, d.Add(`2["unterminated`))
	})

	t.Run("invalid event payload type", func(t *testing.T) {
		d := NewDecoder()
		// EVENT payload must be a non-empty array; an object is invalid.
		assert.Error(t, d.Add(`2{"a":1}`))
	})
}

func TestDecoderDestroyDuringReconstruction(t *testing.T) {
	d := NewDecoder()
	assert.NoError(t, d.Add(`51-["file",{"_placeholder":true,"num":0}]`))
	// Destroy should finish the in-flight reconstructor without panicking.
	d.Destroy()
}
