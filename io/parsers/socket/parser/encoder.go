package parser

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/atendi9/ws/io/pkg/forge"
)

// encoder implements the Encoder interface for Socket.IO packet encoder.
type encoder struct{}

// NewEncoder creates a new Encoder instance.
func NewEncoder() Encoder {
	return &encoder{}
}

// Encode encodes a Socket.IO packet into a sequence of buffers.
// For non-binary packets, it returns a single string buffer.
// For binary packets, it returns the encoded packet header followed by binary buffers.
func (e *encoder) Encode(packet *Packet) []forge.Interface {
	if packet.Type == EVENT || packet.Type == ACK {
		if HasBinary(packet.Data) {
			data := *packet
			if data.Type == EVENT {
				data.Type = BINARY_EVENT
			} else {
				data.Type = BINARY_ACK
			}
			return e.encodeAsBinary(&data)
		}
	}

	return []forge.Interface{e.encodeAsString(packet)}
}

// encodeAsString encodes a packet as a string buffer.
// The format is: <type>[<attachments>-][/<namespace>,][<id>][<data>]
func (e *encoder) encodeAsString(packet *Packet) forge.Interface {
	buffer := forge.NewString([]byte{byte(packet.Type) + '0'})

	if (packet.Type == BINARY_EVENT || packet.Type == BINARY_ACK) && packet.Attachments != nil {
		_, _ = buffer.WriteString(strconv.FormatUint(*packet.Attachments, 10))
		_ = buffer.WriteByte('-')
	}

	if len(packet.Nsp) > 0 && packet.Nsp != "/" {
		_, _ = buffer.WriteString(packet.Nsp)
		_ = buffer.WriteByte(',')
	}

	if packet.Id != nil {
		_, _ = buffer.WriteString(strconv.FormatUint(*packet.Id, 10))
	}

	if packet.Data != nil {
		processedData := preprocessData(packet.Data)
		if jsonBytes, err := json.Marshal(processedData); err == nil {
			if len(jsonBytes) <= forge.MaxPayloadSize {
				_, _ = buffer.Write(jsonBytes)
			}
		}
	}

	return buffer
}

// encodeAsBinary encodes a packet that contains binary data.
// It deconstructs the packet to extract binary data, then encodes the packet header
// followed by all binary buffers.
func (e *encoder) encodeAsBinary(packet *Packet) []forge.Interface {
	deconstructedPacket, buffers := DeconstructPacket(packet)
	header := e.encodeAsString(deconstructedPacket)
	return append([]forge.Interface{header}, buffers...)
}

// preprocessData recursively processes data to convert special types
// that need transformation before JSON encoding.
func preprocessData(data any) any {
	switch typedData := data.(type) {
	case nil:
		return nil
	case *strings.Reader:
		buffer, _ := forge.NewStringReader(typedData)
		return buffer
	case []any:
		result := make([]any, 0, len(typedData))
		for _, item := range typedData {
			result = append(result, preprocessData(item))
		}
		return result
	case map[string]any:
		result := make(map[string]any, len(typedData))
		for key, value := range typedData {
			result[key] = preprocessData(value)
		}
		return result
	default:
		return data
	}
}
