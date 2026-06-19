package wsconn

import "encoding/binary"

// The message types are defined in RFC 6455, section 11.8. They mirror the
// opcodes used on the wire for data and control frames.
const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1
	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2
	// CloseMessage denotes a close control message. The optional message payload
	// contains a numeric code and text.
	CloseMessage = 8
	// PingMessage denotes a ping control message. The optional message payload is
	// UTF-8 encoded text.
	PingMessage = 9
	// PongMessage denotes a pong control message. The optional message payload is
	// UTF-8 encoded text.
	PongMessage = 10
)

// continuationFrame is the opcode used for the second and subsequent frames of
// a fragmented message. It is not a valid argument for the public API.
const continuationFrame = 0

// FormatCloseMessage formats closeCode and text as a WebSocket close message
// payload. An empty payload is returned for CloseNoStatusReceived.
func FormatCloseMessage(closeCode int, text string) []byte {
	if closeCode == CloseNoStatusReceived {
		return []byte{}
	}
	buf := make([]byte, 2+len(text))
	binary.BigEndian.PutUint16(buf, uint16(closeCode))
	copy(buf[2:], text)
	return buf
}

// PreparedMessage caches a message type and payload so the same logical message
// can be written to a connection without re-validating its arguments.
//
// The on-the-wire compression optimization performed by github.com/gorilla/websocket
// is intentionally omitted: this connection implementation does not negotiate
// permessage-deflate, so the payload is written as a single uncompressed frame.
type PreparedMessage struct {
	messageType int
	data        []byte
}

// NewPreparedMessage returns a prepared message bound to the provided message
// type and payload.
func NewPreparedMessage(messageType int, data []byte) (*PreparedMessage, error) {
	return &PreparedMessage{messageType: messageType, data: data}, nil
}
