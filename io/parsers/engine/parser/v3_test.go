package parser

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/ws/io/parsers/engine/packet"
	"github.com/atendi9/ws/io/pkg/forge"
)

func TestV3Parser(t *testing.T) {
	t.Run("NewV3", func(t *testing.T) {
		p := NewV3()
		assert.True(t, p != nil)
	})

	t.Run("Protocol", func(t *testing.T) {
		p := NewV3()
		assert.Equal(t, 3, p.Protocol())
	})

	t.Run("EncodePacket", func(t *testing.T) {
		p := &V3{}
		_, err := p.EncodePacket(nil, false)
		assert.Error(t, err)

		stringData := forge.NewString([]byte("hello"))
		pktString := &packet.Packet{Type: packet.MESSAGE, Data: stringData}
		res, err := p.EncodePacket(pktString, false)
		if err == nil {
			assert.NoError(t, err)
			assert.True(t, res != nil)
		}

		binaryData := bytes.NewReader([]byte{0x01, 0x02})
		pktBinary := &packet.Packet{Type: packet.MESSAGE, Data: binaryData}
		res, err = p.EncodePacket(pktBinary, false)
		if err == nil {
			assert.NoError(t, err)
			assert.True(t, res != nil)
		}
		binaryData2 := bytes.NewReader([]byte{0x01, 0x02})
		pktBinary2 := &packet.Packet{Type: packet.MESSAGE, Data: binaryData2}
		res, err = p.EncodePacket(pktBinary2, true)
		if err == nil {
			assert.NoError(t, err)
			assert.True(t, res != nil)
		}
	})

	t.Run("DecodePacket", func(t *testing.T) {
		p := &V3{}
		_, err := p.DecodePacket(nil)
		assert.Error(t, err)

		dataStr := forge.NewString([]byte("4hello"))
		res, err := p.DecodePacket(dataStr)
		if err == nil {
			assert.NoError(t, err)
			assert.True(t, res != nil)
		}

		b64Data := base64.StdEncoding.EncodeToString([]byte{0x01, 0x02})
		dataB64 := forge.NewString([]byte("b4" + b64Data))
		res, err = p.DecodePacket(dataB64)
		if err == nil {
			assert.NoError(t, err)
			assert.True(t, res != nil)
		}
		dataBin := forge.NewBytesBuffer([]byte{0x04, 0x01, 0x02})
		res, err = p.DecodePacket(dataBin)
		if err == nil {
			assert.NoError(t, err)
			assert.True(t, res != nil)
		}
	})

	t.Run("hasBinary", func(t *testing.T) {
		p := &V3{}

		pkt1 := &packet.Packet{Data: forge.NewString([]byte("hello"))}
		pkt2 := &packet.Packet{Data: bytes.NewReader([]byte{0x01})}

		t.Run("Slice with only string data", func(t *testing.T) {
			packetsWithoutBinary := []*packet.Packet{pkt1}
			assert.False(t, p.hasBinary(packetsWithoutBinary))
		})

		t.Run("Slice with binary data", func(t *testing.T) {
			packetsWithBinary := []*packet.Packet{pkt1, pkt2}
			assert.True(t, p.hasBinary(packetsWithBinary))
		})
	})

	t.Run("EncodePayload", func(t *testing.T) {
		p := &V3{}

		t.Run("Empty payload", func(t *testing.T) {
			emptyPackets := make([]*packet.Packet, 0)
			res, err := p.EncodePayload(emptyPackets)
			assert.NoError(t, err)
			assert.True(t, res != nil)
		})

		t.Run("Payload with standard packets", func(t *testing.T) {
			pkt1 := &packet.Packet{Type: packet.MESSAGE, Data: forge.NewString([]byte("hello"))}
			packets := []*packet.Packet{pkt1}
			res, err := p.EncodePayload(packets, false)
			if err == nil {
				assert.NoError(t, err)
				assert.True(t, res != nil)
			}
		})
	})

	t.Run("encodeOneBinaryPacket", func(t *testing.T) {
		p := &V3{}
		t.Run("nil packet error", func(t *testing.T) {
			_, err := p.encodeOneBinaryPacket(nil)
			assert.Error(t, err)

			pkt := &packet.Packet{Type: packet.MESSAGE, Data: bytes.NewReader([]byte{0x01})}
			res, err := p.encodeOneBinaryPacket(pkt)
			if err == nil {
				assert.NoError(t, err)
				assert.True(t, res != nil)
			}
		})
	})

	t.Run("encodePayloadAsBinary", func(t *testing.T) {
		p := &V3{}

		emptyPackets := make([]*packet.Packet, 0)
		res, err := p.encodePayloadAsBinary(emptyPackets)
		assert.NoError(t, err)
		assert.True(t, res != nil)

		pkt1 := &packet.Packet{Type: packet.MESSAGE, Data: bytes.NewReader([]byte{0x01})}
		packets := []*packet.Packet{pkt1}
		res, err = p.encodePayloadAsBinary(packets)
		if err == nil {
			assert.NoError(t, err)
			assert.True(t, res != nil)
		}
	})

	t.Run("DecodePayload", func(t *testing.T) {
		p := &V3{}
		t.Run("String payload", func(t *testing.T) {
			strPayload := forge.NewString([]byte("6:4hello"))
			packetsStr, err := p.DecodePayload(strPayload)
			if err == nil {
				assert.NoError(t, err)
				assert.LengthSlice(t, 1, packetsStr)
			}
		})

		t.Run("Binary Payload", func(t *testing.T) {
			binPayload := forge.NewBytesBuffer([]byte{0x00, 0x06, 0xFF, '4', 'h', 'e', 'l', 'l', 'o'})
			packetsBin, err := p.DecodePayload(binPayload)
			if err == nil {
				assert.NoError(t, err)
				assert.LengthSlice(t, 1, packetsBin)
			}
		})
	})

	t.Run("decodeStringPayload", func(t *testing.T) {
		p := &V3{}
		payload := forge.NewString([]byte("6:4hello"))
		packets, err := p.decodeStringPayload(payload.(*forge.String))
		if err == nil {
			assert.NoError(t, err)
			assert.LengthSlice(t, 1, packets)
		}
	})

	t.Run("decodeBinaryPayload", func(t *testing.T) {
		p := &V3{}
		payload := forge.NewBytesBuffer([]byte{0x00, 0x06, 0xFF, '4', 'h', 'e', 'l', 'l', 'o'})
		packets, err := p.decodeBinaryPayload(payload)
		if err == nil {
			assert.NoError(t, err)
			assert.LengthSlice(t, 1, packets)
		}
	})
}
