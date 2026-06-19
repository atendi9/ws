package parser

import (
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/atendi9/ws/io/parsers/engine/packet"
	"github.com/atendi9/ws/io/pkg/errors"
	"github.com/atendi9/ws/io/pkg/forge"
)

// V4 represents the Engine.IO version 4 parser implementation.
type V4 struct{}

// defaultV4 holds the default singleton instance of [V4].
var defaultV4 Parser = &V4{}

// NewV4 returns a new Engine.IO version 4 [Parser].
func NewV4() Parser {
	return defaultV4
}

// Protocol returns the protocol version number, corresponding to [Version4].
func (*V4) Protocol() int {
	return int(Version4)
}

// EncodePacket encodes a given [packet.Packet] into an [forge.Interface].
// It determines the encoding strategy based on the data type and the supportsBinary flag.
func (p *V4) EncodePacket(pkt *packet.Packet, supportsBinary bool, _ ...bool) (forge.Interface, error) {
	if pkt == nil {
		return nil, errors.ErrParserPacketNil
	}

	if c, ok := pkt.Data.(io.Closer); ok {
		defer func() { _ = c.Close() }()
	}

	typeByte, ok := lookupPacketByte(pkt.Type)
	if !ok {
		return nil, errors.ErrParserPacketType
	}

	switch v := pkt.Data.(type) {
	case *forge.String, *strings.Reader:
		return p.encodeStringData(typeByte, v)

	case io.Reader:
		return p.encodeBinaryData(v, supportsBinary)
	}

	return p.encodeEmptyPacket(typeByte)
}

// encodeStringData encodes string-based data into an [forge.Interface], prefixing it with the typeByte.
func (p *V4) encodeStringData(typeByte byte, data io.Reader) (forge.Interface, error) {
	encode := forge.NewString(nil)
	if err := encode.WriteByte(typeByte); err != nil {
		return nil, err
	}
	if _, err := io.Copy(encode, data); err != nil {
		return nil, err
	}
	return encode, nil
}

// encodeBinaryData encodes binary data into an [forge.Interface].
// If supportsBinary is false, it falls back to base64 encoding using [V4.encodeAsBase64].
func (p *V4) encodeBinaryData(data io.Reader, supportsBinary bool) (forge.Interface, error) {
	if !supportsBinary {
		return p.encodeAsBase64(data)
	}

	encode := forge.NewBytesBuffer(nil)
	if _, err := io.Copy(encode, data); err != nil {
		return nil, err
	}
	return encode, nil
}

// encodeAsBase64 encodes binary data as a base64 string within an [forge.Interface].
func (p *V4) encodeAsBase64(data io.Reader) (forge.Interface, error) {
	encode := forge.NewString(nil)
	if err := encode.WriteByte('b'); err != nil {
		return nil, err
	}

	b64 := base64.NewEncoder(base64.StdEncoding, encode)
	if _, err := io.Copy(b64, data); err != nil {
		_ = b64.Close()
		return nil, err
	}
	if err := b64.Close(); err != nil {
		return nil, err
	}
	return encode, nil
}

// encodeEmptyPacket creates an empty [forge.Interface] containing only the typeByte.
func (p *V4) encodeEmptyPacket(typeByte byte) (forge.Interface, error) {
	encode := forge.NewString(nil)
	if err := encode.WriteByte(typeByte); err != nil {
		return nil, err
	}
	return encode, nil
}

// DecodePacket decodes the provided [forge.Interface] data into a [packet.Packet].
func (p *V4) DecodePacket(data forge.Interface, _ ...bool) (*packet.Packet, error) {
	if data == nil {
		return newErrorPacket(), errors.ErrParserDataNil
	}

	if sb, ok := data.(*forge.String); ok {
		return p.decodeStringPacket(sb)
	}

	return p.decodeBinaryPacket(data)
}

// decodeStringPacket decodes string data from an [forge.String] into a [packet.Packet].
func (p *V4) decodeStringPacket(sb *forge.String) (*packet.Packet, error) {
	msgType, err := sb.ReadByte()
	if err != nil {
		return newErrorPacket(), err
	}

	if msgType == 'b' {
		return p.decodeBase64Packet(sb)
	}

	packetType, ok := lookupPacketType(msgType)
	if !ok {
		return newErrorPacket(), fmt.Errorf("%w: [%c]", errors.ErrParserUnknownPacketType, msgType)
	}

	stringBuffer := forge.NewString(nil)
	if _, err := stringBuffer.ReadFrom(sb); err != nil {
		return newErrorPacket(), err
	}
	return &packet.Packet{Type: packetType, Data: stringBuffer}, nil
}

// decodeBase64Packet decodes base64-encoded string data from an [forge.String] into a [packet.Packet].
func (p *V4) decodeBase64Packet(sb *forge.String) (*packet.Packet, error) {
	decode := forge.NewBytesBuffer(nil)
	if _, err := decode.ReadFrom(base64.NewDecoder(base64.StdEncoding, sb)); err != nil {
		return newErrorPacket(), err
	}
	return &packet.Packet{Type: packet.MESSAGE, Data: decode}, nil
}

// decodeBinaryPacket decodes binary data from an [forge.Interface] into a [packet.Packet].
func (p *V4) decodeBinaryPacket(data forge.Interface) (*packet.Packet, error) {
	decode := forge.NewBytesBuffer(nil)
	if _, err := io.Copy(decode, data); err != nil {
		return newErrorPacket(), err
	}
	return &packet.Packet{Type: packet.MESSAGE, Data: decode}, nil
}

// EncodePayload encodes a slice of [packet.Packet] instances into a single [forge.Interface] payload.
func (p *V4) EncodePayload(packets []*packet.Packet, _ ...bool) (forge.Interface, error) {
	encodedPayload := forge.NewString(nil)

	if len(packets) == 0 {
		return encodedPayload, nil
	}

	for i, pkt := range packets {
		buf, err := p.EncodePacket(pkt, false)
		if err != nil {
			return nil, err
		}

		if i > 0 {
			if err := encodedPayload.WriteByte(Protocol.Separator); err != nil {
				return nil, err
			}
		}

		if _, err := io.Copy(encodedPayload, buf); err != nil {
			return nil, err
		}
	}

	return encodedPayload, nil
}

// DecodePayload decodes a single [forge.Interface] payload into a slice of [packet.Packet] instances.
func (p *V4) DecodePayload(data forge.Interface) ([]*packet.Packet, error) {
	packets := make([]*packet.Packet, 0, 4)

	for data.Len() > 0 {
		scanBytes, err := data.ReadBytes(Protocol.Separator)
		if err != nil && err != io.EOF {
			return packets, err
		}

		if len(scanBytes) > 0 && scanBytes[len(scanBytes)-1] == Protocol.Separator {
			scanBytes = scanBytes[:len(scanBytes)-1]
		}

		if len(scanBytes) == 0 {
			if err == io.EOF {
				break
			}
			continue
		}

		pkt, decodeErr := p.DecodePacket(forge.NewString(scanBytes))
		if decodeErr != nil {
			return packets, decodeErr
		}
		packets = append(packets, pkt)

		if err == io.EOF {
			break
		}
	}

	return packets, nil
}
