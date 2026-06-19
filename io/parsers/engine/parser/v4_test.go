package parser

import (
	"encoding/base64"
	"io"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/ws/io/parsers/engine/packet"
	"github.com/atendi9/ws/io/pkg/errors"
	"github.com/atendi9/ws/io/pkg/forge"
)

func TestNewV4(t *testing.T) {
	p := NewV4()
	assert.True(t, p != nil)
}

func TestV4_Protocol(t *testing.T) {
	p := NewV4().(*V4)
	expected := int(Version4)
	assert.Equal(t, expected, p.Protocol())
}

func TestV4_EncodePacket(t *testing.T) {
	p := NewV4().(*V4)

	t.Run("nil packet", func(t *testing.T) {
		_, err := p.EncodePacket(nil, true)
		assert.Error(t, err)
		assert.Equal(t, errors.ErrParserPacketNil.Error(), err.Error())
	})

	t.Run("string data", func(t *testing.T) {
		pkt := &packet.Packet{
			Type: packet.MESSAGE,
			Data: forge.NewString([]byte("hello")),
		}
		buf, err := p.EncodePacket(pkt, true)
		assert.NoError(t, err)

		typeByte, _ := lookupPacketByte(packet.MESSAGE)
		expectedStr := string(typeByte) + "hello"
		assert.Equal(t, expectedStr, buf.String())
	})

	t.Run("binary data with binary support", func(t *testing.T) {
		pkt := &packet.Packet{
			Type: packet.MESSAGE,
			Data: forge.NewBytesBuffer([]byte{1, 2, 3}),
		}
		buf, err := p.EncodePacket(pkt, true)
		assert.NoError(t, err)

		expectedBytes := []byte{1, 2, 3}
		assert.Equal(t, string(expectedBytes), string(buf.Bytes()))
	})

	t.Run("binary data without binary support", func(t *testing.T) {
		pkt := &packet.Packet{
			Type: packet.MESSAGE,
			Data: forge.NewBytesBuffer([]byte{1, 2, 3}),
		}
		buf, err := p.EncodePacket(pkt, false)
		assert.NoError(t, err)

		expectedB64 := "b" + base64.StdEncoding.EncodeToString([]byte{1, 2, 3})
		assert.Equal(t, expectedB64, buf.String())
	})

	t.Run("empty packet", func(t *testing.T) {
		pkt := &packet.Packet{
			Type: packet.PING,
		}
		buf, err := p.EncodePacket(pkt, true)
		assert.NoError(t, err)

		typeByte, _ := lookupPacketByte(packet.PING)
		assert.Equal(t, string(typeByte), buf.String())
	})
}

func TestV4_DecodePacket(t *testing.T) {
	p := NewV4().(*V4)

	t.Run("nil data", func(t *testing.T) {
		_, err := p.DecodePacket(nil)
		assert.Error(t, err)
		assert.Equal(t, errors.ErrParserDataNil.Error(), err.Error())
	})

	t.Run("string packet", func(t *testing.T) {
		typeByte, _ := lookupPacketByte(packet.MESSAGE)
		data := forge.NewString([]byte{typeByte, 'w', 'o', 'r', 'l', 'd'})

		pkt, err := p.DecodePacket(data)
		assert.NoError(t, err)
		assert.Equal(t, packet.MESSAGE, pkt.Type)

		dataBuf, _ := io.ReadAll(pkt.Data)
		assert.Equal(t, "world", string(dataBuf))
	})

	t.Run("base64 packet", func(t *testing.T) {
		b64Data := base64.StdEncoding.EncodeToString([]byte{1, 2, 3})
		data := forge.NewString([]byte("b" + b64Data))

		pkt, err := p.DecodePacket(data)
		assert.NoError(t, err)
		assert.Equal(t, packet.MESSAGE, pkt.Type)

		dataBuf, _ := io.ReadAll(pkt.Data)
		assert.Equal(t, string([]byte{1, 2, 3}), string(dataBuf))
	})

	t.Run("binary packet", func(t *testing.T) {
		data := forge.NewBytesBuffer([]byte{4, 5, 6})

		pkt, err := p.DecodePacket(data)
		assert.NoError(t, err)
		assert.Equal(t, packet.MESSAGE, pkt.Type)

		dataBuf, _ := io.ReadAll(pkt.Data)
		assert.Equal(t, string([]byte{4, 5, 6}), string(dataBuf))
	})
}

func TestV4_EncodePayload(t *testing.T) {
	p := NewV4().(*V4)

	t.Run("empty slice", func(t *testing.T) {
		buf, err := p.EncodePayload(nil)
		assert.NoError(t, err)
		assert.Equal(t, 0, buf.Len())
	})

	t.Run("multiple packets", func(t *testing.T) {
		pkt1 := &packet.Packet{
			Type: packet.MESSAGE,
			Data: forge.NewString([]byte("one")),
		}
		pkt2 := &packet.Packet{
			Type: packet.MESSAGE,
			Data: forge.NewString([]byte("two")),
		}

		buf, err := p.EncodePayload([]*packet.Packet{pkt1, pkt2})
		assert.NoError(t, err)

		typeByte, _ := lookupPacketByte(packet.MESSAGE)
		expected := string(typeByte) + "one" + string(Protocol.Separator) + string(typeByte) + "two"
		assert.Equal(t, expected, buf.String())
	})
}

func TestV4_DecodePayload(t *testing.T) {
	p := NewV4().(*V4)

	t.Run("multiple packets", func(t *testing.T) {
		typeByte, _ := lookupPacketByte(packet.MESSAGE)
		payloadStr := string(typeByte) + "foo" + string(Protocol.Separator) + string(typeByte) + "bar"
		payload := forge.NewString([]byte(payloadStr))

		pkts, err := p.DecodePayload(payload)
		assert.NoError(t, err)
		assert.LengthSlice(t, 2, pkts)

		assert.Equal(t, packet.MESSAGE, pkts[0].Type)
		data1, _ := io.ReadAll(pkts[0].Data)
		assert.Equal(t, "foo", string(data1))

		assert.Equal(t, packet.MESSAGE, pkts[1].Type)
		data2, _ := io.ReadAll(pkts[1].Data)
		assert.Equal(t, "bar", string(data2))
	})
}
