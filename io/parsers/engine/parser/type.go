// Package parser provides Engine.IO protocol packet encoding and decoding.
package parser

import (
	"github.com/atendi9/ws/io/parsers/engine/packet"
	"github.com/atendi9/ws/io/pkg/forge"
)

// Parser defines the interface for Engine.IO packet encoding and decoding.
type Parser interface {
	// Protocol returns the Engine.IO protocol version.
	Protocol() int
	// EncodePacket encodes a packet to a buffer.
	// supportsBinary indicates whether the transport supports binary data.
	// utf8encode (v3 only) indicates whether to perform UTF-8 encoding.
	EncodePacket(pkt *packet.Packet, supportsBinary bool, utf8encode ...bool) (forge.Interface, error)
	// DecodePacket decodes a buffer to a packet.
	// utf8decode (v3 only) indicates whether to perform UTF-8 decoding.
	DecodePacket(data forge.Interface, utf8decode ...bool) (*packet.Packet, error)
	// EncodePayload encodes multiple packets into a single payload.
	// supportsBinary (v3 only) indicates whether to use binary encoding.
	EncodePayload(packets []*packet.Packet, supportsBinary ...bool) (forge.Interface, error)
	// DecodePayload decodes a payload buffer into multiple packets.
	DecodePayload(data forge.Interface) ([]*packet.Packet, error)
}

// Version represents the Engine.IO protocol version.
type Version int

const (
	// Version4 represents the Engine.IO protocol version 4.
	Version4 Version = 4
	// Version3 represents the Engine.IO protocol version 3.
	Version3 Version = 3
)

// protocol holds the configuration for a specific Engine.IO protocol version,
// including the target protocol [Version], the [Parser] implementation,
// and the separator byte used for encoding and decoding payloads.
type protocol struct {
	Version   Version
	Parser    Parser
	Separator byte
}

// Protocol is the default configured [protocol] instance,
// currently set to use [Version4], the [V4] [Parser], and the standard separator byte.
var Protocol protocol = protocol{
	Version:   Version4,
	Parser:    NewV4(),
	Separator: 0x1E,
}

// PacketByte returns the corresponding byte representation for a given [packet.Type].
// It returns the mapped byte and a boolean indicating whether the given type is valid.
func PacketByte(t packet.Type) (byte, bool) {
	m := map[packet.Type]byte{
		packet.OPEN:    '0',
		packet.CLOSE:   '1',
		packet.PING:    '2',
		packet.PONG:    '3',
		packet.MESSAGE: '4',
		packet.UPGRADE: '5',
		packet.NOOP:    '6',
	}
	c, ok := m[t]
	return c, ok
}

// ConvertByteToPacketType parses a byte and returns its corresponding [packet.Type].
// It returns the mapped packet type and a boolean indicating whether the given byte is a valid type identifier.
func ConvertByteToPacketType(c byte) (packet.Type, bool) {
	m := map[byte]packet.Type{
		'0': packet.OPEN,
		'1': packet.CLOSE,
		'2': packet.PING,
		'3': packet.PONG,
		'4': packet.MESSAGE,
		'5': packet.UPGRADE,
		'6': packet.NOOP,
	}
	t, ok := m[c]
	return t, ok
}

// newErrorPacket creates a fresh error packet for parser errors.
// A new instance is returned each time to avoid sharing mutable state across goroutines.
func newErrorPacket() *packet.Packet {
	return &packet.Packet{
		Type: packet.ERROR,
		Data: forge.NewFromString("parser error"),
	}
}

// lookupPacketType returns the packet type for the given byte.
// Returns the packet type and true if found, otherwise empty type and false.
func lookupPacketType(b byte) (packet.Type, bool) {
	pt, ok := ConvertByteToPacketType(b)
	return pt, ok
}

// lookupPacketByte returns the wire format byte for the given packet type.
// Returns the byte and true if found, otherwise 0 and false.
func lookupPacketByte(t packet.Type) (byte, bool) {
	b, ok := PacketByte(t)
	return b, ok
}
