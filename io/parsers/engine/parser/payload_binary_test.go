package parser

import (
	"testing"

	"github.com/atendi9/ws/io/parsers/engine/packet"
	"github.com/atendi9/ws/io/pkg/forge"

	"github.com/atendi9/capivara/assert"
)

func TestV3_EncodePayloadBinary(t *testing.T) {
	p := NewV3().(*V3)

	t.Run("empty payload", func(t *testing.T) {
		buf, err := p.EncodePayload(nil, true)
		assert.NoError(t, err)
		assert.NotNil(t, buf)
	})

	t.Run("binary packets round-trip", func(t *testing.T) {
		packets := []*packet.Packet{
			{Type: packet.MESSAGE, Data: forge.NewBytesBuffer([]byte{1, 2, 3})},
			{Type: packet.MESSAGE, Data: forge.NewString([]byte("text"))},
		}
		encoded, err := p.EncodePayload(packets, true)
		assert.NoError(t, err)
		assert.NotNil(t, encoded)

		decoded, err := p.DecodePayload(encoded)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(decoded))
	})
}

func TestV4_EncodePayloadMixed(t *testing.T) {
	p := NewV4().(*V4)

	packets := []*packet.Packet{
		{Type: packet.MESSAGE, Data: forge.NewString([]byte("hello"))},
		{Type: packet.MESSAGE, Data: forge.NewBytesBuffer([]byte{9, 8, 7})},
		{Type: packet.PING},
	}
	encoded, err := p.EncodePayload(packets)
	assert.NoError(t, err)
	assert.NotNil(t, encoded)

	decoded, err := p.DecodePayload(encoded)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(decoded))
}

func TestV4_EncodeBase64ForBinaryWithoutSupport(t *testing.T) {
	p := NewV4().(*V4)
	pkt := &packet.Packet{Type: packet.MESSAGE, Data: forge.NewBytesBuffer([]byte{10, 20, 30})}

	// supportsBinary=false forces the base64 ("b" prefixed) encoding path.
	buf, err := p.EncodePacket(pkt, false)
	assert.NoError(t, err)
	assert.Equal(t, byte('b'), buf.String()[0])

	// And it decodes back to the original bytes.
	decoded, err := p.DecodePacket(buf)
	assert.NoError(t, err)
	assert.NotNil(t, decoded)
}
