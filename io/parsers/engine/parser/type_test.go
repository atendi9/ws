package parser

import (
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/ws/io/parsers/engine/packet"
)

func TestPacketByte(t *testing.T) {
	tests := []struct {
		name     string
		pktType  packet.Type
		expected byte
		ok       bool
	}{
		{"Open", packet.OPEN, '0', true},
		{"Close", packet.CLOSE, '1', true},
		{"Ping", packet.PING, '2', true},
		{"Pong", packet.PONG, '3', true},
		{"Message", packet.MESSAGE, '4', true},
		{"Upgrade", packet.UPGRADE, '5', true},
		{"Noop", packet.NOOP, '6', true},
		{"Unknown", packet.Type(""), 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := PacketByte(tt.pktType)
			if tt.ok {
				assert.True(t, ok)
				assert.Equal(t, tt.expected, result)
			} else {
				assert.False(t, ok)
			}
		})
	}
}

func TestConvertByteToPacketType(t *testing.T) {
	tests := []struct {
		name     string
		input    byte
		expected packet.Type
		ok       bool
	}{
		{"Open", '0', packet.OPEN, true},
		{"Close", '1', packet.CLOSE, true},
		{"Ping", '2', packet.PING, true},
		{"Pong", '3', packet.PONG, true},
		{"Message", '4', packet.MESSAGE, true},
		{"Upgrade", '5', packet.UPGRADE, true},
		{"Noop", '6', packet.NOOP, true},
		{"Unknown", '9', packet.Type(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := ConvertByteToPacketType(tt.input)
			if tt.ok {
				assert.True(t, ok)
				assert.Equal(t, tt.expected, result)
			} else {
				assert.False(t, ok)
			}
		})
	}
}

func TestNewErrorPacket(t *testing.T) {
	pkt := newErrorPacket()

	isNotNil := pkt != nil
	assert.True(t, isNotNil)
	assert.Equal(t, packet.ERROR, pkt.Type)
}

func TestLookupPacketType(t *testing.T) {
	t.Run("Valid Byte", func(t *testing.T) {
		result, ok := lookupPacketType('4')
		assert.True(t, ok)
		assert.Equal(t, packet.MESSAGE, result)
	})

	t.Run("Invalid Byte", func(t *testing.T) {
		_, ok := lookupPacketType('9')
		assert.False(t, ok)
	})
}

func TestLookupPacketByte(t *testing.T) {
	t.Run("Valid Packet Type", func(t *testing.T) {
		result, ok := lookupPacketByte(packet.PING)
		assert.True(t, ok)
		assert.Equal(t, byte('2'), result)
	})

	t.Run("Invalid Packet Type", func(t *testing.T) {
		_, ok := lookupPacketByte(packet.Type(""))
		assert.False(t, ok)
	})
}
