package parser

import (
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/atendi9/ws/io/parsers/engine/packet"
	"github.com/atendi9/ws/io/pkg/errors"
	"github.com/atendi9/ws/io/pkg/etch"
	"github.com/atendi9/ws/io/pkg/forge"
)

// V3 represents the Engine.IO version 3 parser implementation.
type V3 struct{}

// defaultV3 is the default instance of the [V3] parser.
var defaultV3 Parser = &V3{}

// NewV3 creates and returns a new instance of the Engine.IO version 3 [Parser].
func NewV3() Parser {
	return defaultV3
}

// Protocol returns the protocol version number, which is 3 for the [V3] parser.
func (*V3) Protocol() int {
	return int(Version3)
}

// EncodePacket encodes a given [packet.Packet] into an [forge.Interface].
// It returns an error if the provided [packet.Packet] is nil or if the packet type is invalid.
func (p *V3) EncodePacket(data *packet.Packet, supportsBinary bool, utf8encode ...bool) (forge.Interface, error) {
	if data == nil {
		return nil, errors.ErrParserPacketNil
	}

	if c, ok := data.Data.(io.Closer); ok {
		defer func() { _ = c.Close() }()
	}

	utf8en := len(utf8encode) > 0 && utf8encode[0]

	switch v := data.Data.(type) {
	case *forge.String, *strings.Reader:
		encode := forge.NewString(nil)
		typeByte, ok := lookupPacketByte(data.Type)
		if !ok {
			return nil, errors.ErrParserPacketType
		}
		if err := encode.WriteByte(typeByte); err != nil {
			return nil, err
		}
		if utf8en {
			if _, err := io.Copy(etch.NewUtf8Encoder(encode), v); err != nil {
				return nil, err
			}
		} else {
			if _, err := io.Copy(encode, v); err != nil {
				return nil, err
			}
		}
		return encode, nil

	case io.Reader:
		typeByte, ok := lookupPacketByte(data.Type)
		if !ok {
			return nil, errors.ErrParserPacketType
		}
		if !supportsBinary {
			encode := forge.NewString(nil)
			if _, err := encode.Write([]byte{'b', typeByte}); err != nil {
				return nil, err
			}
			b64 := base64.NewEncoder(base64.StdEncoding, encode)
			if _, err := io.Copy(b64, v); err != nil {
				_ = b64.Close()
				return nil, err
			}
			if err := b64.Close(); err != nil {
				return nil, err
			}
			return encode, nil
		}
		encode := forge.NewBytesBuffer(nil)
		if err := encode.WriteByte(typeByte - '0'); err != nil {
			return nil, err
		}
		if _, err := io.Copy(encode, v); err != nil {
			return nil, err
		}
		return encode, nil
	}

	encode := forge.NewString(nil)
	typeByte, ok := lookupPacketByte(data.Type)
	if !ok {
		return nil, errors.ErrParserPacketType
	}
	if err := encode.WriteByte(typeByte); err != nil {
		return nil, err
	}
	return encode, nil
}

// DecodePacket decodes an [forge.Interface] back into a [packet.Packet].
// It reads the packet type and data, handling both base64 strings and binary buffers.
func (p *V3) DecodePacket(data forge.Interface, utf8decode ...bool) (*packet.Packet, error) {
	if data == nil {
		return newErrorPacket(), errors.ErrParserDataNil
	}

	utf8de := len(utf8decode) > 0 && utf8decode[0]

	msgType, err := data.ReadByte()
	if err != nil {
		return newErrorPacket(), err
	}

	switch v := data.(type) {
	case *forge.String:
		if msgType == 'b' {
			msgType, err = data.ReadByte()
			if err != nil {
				return newErrorPacket(), err
			}
			packetType, ok := lookupPacketType(msgType)
			if !ok {
				return newErrorPacket(), fmt.Errorf("%w: [%c]", errors.ErrParserUnknownPacketType, msgType)
			}
			decode := forge.NewBytesBuffer(nil)
			if _, err := decode.ReadFrom(base64.NewDecoder(base64.StdEncoding, v)); err != nil {
				return newErrorPacket(), err
			}
			return &packet.Packet{Type: packetType, Data: decode}, nil
		}
		packetType, ok := lookupPacketType(msgType)
		if !ok {
			return newErrorPacket(), fmt.Errorf("%w: [%c]", errors.ErrParserUnknownPacketType, msgType)
		}
		decode := forge.NewString(nil)
		if utf8de {
			if _, err := decode.ReadFrom(etch.NewUtf8Decoder(v)); err != nil {
				return newErrorPacket(), err
			}
		} else {
			if _, err := decode.ReadFrom(v); err != nil {
				return newErrorPacket(), err
			}
		}
		return &packet.Packet{Type: packetType, Data: decode}, nil
	}

	packetType, ok := lookupPacketType(msgType + '0')
	if !ok {
		return newErrorPacket(), fmt.Errorf("%w: [%c]", errors.ErrParserUnknownPacketType, msgType+'0')
	}
	decode := forge.NewBytesBuffer(nil)
	if _, err := io.Copy(decode, data); err != nil {
		return newErrorPacket(), err
	}
	return &packet.Packet{Type: packetType, Data: decode}, nil
}

// hasBinary checks if the provided slice of [packet.Packet] contains any binary data.
func (p *V3) hasBinary(packets []*packet.Packet) bool {
	for _, pkt := range packets {
		if pkt == nil {
			continue
		}
		switch pkt.Data.(type) {
		case *forge.String, *strings.Reader, nil:
		default:
			return true
		}
	}
	return false
}

// EncodePayload encodes a slice of [packet.Packet] into a single [forge.Interface] payload.
// It delegates to binary encoding if binary is supported and present in the packets.
func (p *V3) EncodePayload(packets []*packet.Packet, supportsBinary ...bool) (forge.Interface, error) {
	supportsBin := len(supportsBinary) > 0 && supportsBinary[0]

	if supportsBin && p.hasBinary(packets) {
		return p.encodePayloadAsBinary(packets)
	}

	enPayload := forge.NewString(nil)

	if len(packets) == 0 {
		if _, err := enPayload.WriteString("0:"); err != nil {
			return nil, err
		}
		return enPayload, nil
	}

	for _, pkt := range packets {
		buf, err := p.EncodePacket(pkt, supportsBin, false)
		if err != nil {
			return nil, err
		}
		if _, err := enPayload.WriteString(strconv.FormatInt(int64(etch.Utf16Count(buf.Bytes())), 10)); err != nil {
			return nil, err
		}
		if err := enPayload.WriteByte(':'); err != nil {
			return nil, err
		}
		if _, err := enPayload.Write(buf.Bytes()); err != nil {
			return nil, err
		}
	}

	return enPayload, nil
}

// encodeOneBinaryPacket encodes a single [packet.Packet] into a binary [forge.Interface].
func (p *V3) encodeOneBinaryPacket(pkt *packet.Packet) (forge.Interface, error) {
	if pkt == nil {
		return nil, errors.ErrParserPacketNil
	}

	buf, err := p.EncodePacket(pkt, true, true)
	if err != nil {
		return nil, err
	}

	binaryPacket := forge.NewBytesBuffer(nil)

	if _, ok := buf.(*forge.String); ok {
		encodingLength := strconv.FormatInt(int64(etch.Utf16Count(buf.Bytes())), 10) // JS length
		if err := binaryPacket.WriteByte(0x00); err != nil {
			return nil, err
		}
		for i := range len(encodingLength) {
			if err := binaryPacket.WriteByte(encodingLength[i] - '0'); err != nil {
				return nil, err
			}
		}
		if err := binaryPacket.WriteByte(0xFF); err != nil {
			return nil, err
		}
		if _, err := buf.WriteTo(etch.NewUtf8Encoder(binaryPacket)); err != nil {
			return nil, err
		}
		return binaryPacket, nil
	}

	encodingLength := strconv.FormatInt(int64(buf.Len()), 10)
	if err := binaryPacket.WriteByte(0x01); err != nil {
		return nil, err
	}
	for i := range len(encodingLength) {
		if err := binaryPacket.WriteByte(encodingLength[i] - '0'); err != nil {
			return nil, err
		}
	}
	if err := binaryPacket.WriteByte(0xFF); err != nil {
		return nil, err
	}
	if _, err := binaryPacket.ReadFrom(buf); err != nil {
		return nil, err
	}
	return binaryPacket, nil
}

// encodePayloadAsBinary encodes a slice of [packet.Packet] into a binary [forge.Interface] payload.
func (p *V3) encodePayloadAsBinary(packets []*packet.Packet) (forge.Interface, error) {
	enPayload := forge.NewBytesBuffer(nil)

	if len(packets) == 0 {
		return enPayload, nil
	}

	for _, pkt := range packets {
		buf, err := p.encodeOneBinaryPacket(pkt)
		if err != nil {
			return nil, err
		}
		if _, err := enPayload.ReadFrom(buf); err != nil {
			return nil, err
		}
	}

	return enPayload, nil
}

// DecodePayload decodes an [forge.Interface] payload into a slice of [packet.Packet].
// It determines whether to decode as a string or binary payload based on the underlying type.
func (p *V3) DecodePayload(data forge.Interface) ([]*packet.Packet, error) {
	if v, ok := data.(*forge.String); ok {
		return p.decodeStringPayload(v)
	}
	return p.decodeBinaryPayload(data)
}

// decodeStringPayload decodes a string-based [forge.String] payload into a slice of [packet.Packet].
func (p *V3) decodeStringPayload(v *forge.String) ([]*packet.Packet, error) {
	packets := make([]*packet.Packet, 0, 8)

	for v.Len() > 0 {
		length, err := v.ReadString(':')
		if err != nil {
			return packets, err
		}
		l := len(length)
		if l < 2 {
			return packets, errors.ErrParserInvalidDataLength
		}
		packetLen, err := strconv.Atoi(length[:l-1])
		if err != nil {
			return packets, err
		}
		if packetLen < 0 {
			return packets, errors.ErrParserInvalidDataLength
		}
		msg := forge.NewString(nil)
		for i := 0; i < packetLen; {
			r, _, e := v.ReadRune()
			if e != nil {
				return packets, e
			}
			i += etch.Utf16Len(r)
			if _, err := msg.WriteRune(r); err != nil {
				return packets, err
			}
		}

		if msg.Len() > 0 {
			pkt, err := p.DecodePacket(msg, false)
			if err != nil {
				return packets, err
			}
			packets = append(packets, pkt)
		}
	}
	return packets, nil
}

// decodeBinaryPayload decodes a binary-based [forge.Interface] payload into a slice of [packet.Packet].
func (p *V3) decodeBinaryPayload(bufferTail forge.Interface) ([]*packet.Packet, error) {
	packets := make([]*packet.Packet, 0, 8)

	for bufferTail.Len() > 0 {
		startByte, err := bufferTail.ReadByte()
		if err != nil {
			return packets, err
		}
		isString := startByte == 0x00

		lengthBytes, err := bufferTail.ReadBytes(0xFF)
		if err != nil {
			return packets, err
		}
		l := len(lengthBytes)
		if l < 1 {
			return packets, errors.ErrParserInvalidDataLength
		}
		lenByte := lengthBytes[:l-1]
		for k := range lenByte {
			lenByte[k] += '0'
		}
		packetLen, err := strconv.Atoi(string(lenByte))
		if err != nil {
			return packets, err
		}
		if packetLen < 0 {
			return packets, errors.ErrParserInvalidDataLength
		}

		if isString {
			data := forge.NewString(nil)
			runeBuf := make([]byte, 0, 4)

			for k := 0; k < packetLen; {
				runeBuf = runeBuf[:0]
				for len(runeBuf) < 4 {
					r, _, err := bufferTail.ReadRune()
					if err != nil {
						if err == io.EOF && len(runeBuf) > 0 {
							break
						}
						return packets, err
					}
					runeBuf = append(runeBuf, byte(r))
					if utf8.FullRune(runeBuf) {
						break
					}
				}
				r, runeLen := utf8.DecodeRune(runeBuf)
				k += etch.Utf16Len(r)
				if _, err := data.Write(etch.Utf8decodeBytes(runeBuf[:runeLen])); err != nil {
					return packets, err
				}
			}

			if data.Len() > 0 {
				pkt, err := p.DecodePacket(data, false)
				if err != nil {
					return packets, err
				}
				packets = append(packets, pkt)
			}
		} else {
			if rawData := bufferTail.Next(packetLen); len(rawData) > 0 {
				pkt, err := p.DecodePacket(forge.NewBytesBuffer(rawData), false)
				if err != nil {
					return packets, err
				}
				packets = append(packets, pkt)
			}
		}
	}
	return packets, nil
}
