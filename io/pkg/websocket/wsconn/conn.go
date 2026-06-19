
package wsconn

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

const (
	finBit  = 1 << 7
	rsvBits = 0x70
	maskBit = 1 << 7

	// maxControlFramePayload is the largest payload allowed on a control frame
	// per RFC 6455, section 5.5.
	maxControlFramePayload = 125
)

// errWriteClosed is returned when writing to a message writer that has already
// been closed.
var errWriteClosed = errors.New("websocket: write to closed writer")

// Conn is a low-level WebSocket connection implementing the subset of the
// RFC 6455 framing protocol used by the project. It speaks both the server and
// client roles; client frames are masked, server frames are not.
type Conn struct {
	conn     net.Conn
	br       *bufio.Reader
	isServer bool

	subprotocol string

	// read state, mutated only by the active message reader.
	readErr         error
	readRemaining   int64
	readFinal       bool
	readMasked      bool
	readMaskKey     [4]byte
	readMaskPos     int
	readLimit       int64
	readLength      int64
	readMessageType int

	// writeMu serializes frame writes so concurrent control and data frames are
	// not interleaved on the wire.
	writeMu       sync.Mutex
	writeDeadline time.Time
}

// newConn builds a Conn around an already-hijacked or dialed network connection.
// When br is nil a fresh buffered reader is created with the requested size.
func newConn(netConn net.Conn, isServer bool, br *bufio.Reader, readBufferSize int, subprotocol string) *Conn {
	if br == nil {
		if readBufferSize <= 0 {
			br = bufio.NewReader(netConn)
		} else {
			br = bufio.NewReaderSize(netConn, readBufferSize)
		}
	}
	return &Conn{
		conn:        netConn,
		br:          br,
		isServer:    isServer,
		subprotocol: subprotocol,
	}
}

// Subprotocol returns the negotiated subprotocol, or the empty string when none
// was negotiated.
func (c *Conn) Subprotocol() string { return c.subprotocol }

// UnderlyingConn returns the network connection backing the WebSocket connection.
func (c *Conn) UnderlyingConn() net.Conn { return c.conn }

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr { return c.conn.LocalAddr() }

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr { return c.conn.RemoteAddr() }

// SetReadLimit sets the maximum size in bytes for a single read message. A
// value of zero or less disables the limit.
func (c *Conn) SetReadLimit(limit int64) { c.readLimit = limit }

// SetReadDeadline sets the read deadline on the underlying network connection.
func (c *Conn) SetReadDeadline(t time.Time) error { return c.conn.SetReadDeadline(t) }

// SetWriteDeadline sets the deadline applied to subsequent frame writes.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	c.writeDeadline = t
	return nil
}

// EnableWriteCompression is a no-op. permessage-deflate is intentionally not
// negotiated by this implementation, so per-message compression toggles have no
// effect. It is retained for API compatibility.
func (c *Conn) EnableWriteCompression(bool) {}

// Close closes the underlying network connection without sending a close frame.
func (c *Conn) Close() error { return c.conn.Close() }

// NextReader returns the message type and a reader for the next data message
// received on the connection. Control frames (ping, pong, close) are handled
// transparently; a received close frame surfaces as a *CloseError.
func (c *Conn) NextReader() (messageType int, r io.Reader, err error) {
	c.readErr = nil
	c.readLength = 0
	opcode, err := c.advanceFrame()
	if err != nil {
		return 0, nil, err
	}
	c.readMessageType = opcode
	return opcode, &messageReader{c: c}, nil
}

// advanceFrame reads frame headers until it reaches a data or continuation
// frame, processing any intervening control frames. It returns the opcode of the
// data frame and leaves the connection positioned at the start of its payload.
func (c *Conn) advanceFrame() (int, error) {
	for {
		var hdr [2]byte
		if _, err := io.ReadFull(c.br, hdr[:]); err != nil {
			return 0, err
		}

		final := hdr[0]&finBit != 0
		if hdr[0]&rsvBits != 0 {
			return 0, c.protocolError("unexpected reserved bits set")
		}
		opcode := int(hdr[0] & 0x0f)
		masked := hdr[1]&maskBit != 0
		length := int64(hdr[1] & 0x7f)

		switch length {
		case 126:
			var b [2]byte
			if _, err := io.ReadFull(c.br, b[:]); err != nil {
				return 0, err
			}
			length = int64(binary.BigEndian.Uint16(b[:]))
		case 127:
			var b [8]byte
			if _, err := io.ReadFull(c.br, b[:]); err != nil {
				return 0, err
			}
			length = int64(binary.BigEndian.Uint64(b[:]))
		}

		var maskKey [4]byte
		if masked {
			if _, err := io.ReadFull(c.br, maskKey[:]); err != nil {
				return 0, err
			}
		}

		switch opcode {
		case continuationFrame, TextMessage, BinaryMessage:
			c.readFinal = final
			c.readRemaining = length
			c.readMasked = masked
			c.readMaskKey = maskKey
			c.readMaskPos = 0
			return opcode, nil
		case CloseMessage, PingMessage, PongMessage:
			if length > maxControlFramePayload || !final {
				return 0, c.protocolError("invalid control frame")
			}
			payload := make([]byte, length)
			if _, err := io.ReadFull(c.br, payload); err != nil {
				return 0, err
			}
			if masked {
				pos := 0
				maskBytes(maskKey, &pos, payload)
			}
			if err := c.handleControl(opcode, payload); err != nil {
				return 0, err
			}
		default:
			return 0, c.protocolError("unknown opcode")
		}
	}
}

// handleControl processes a fully read control frame payload. Pings are answered
// with a pong, pongs are ignored, and a close frame is echoed and reported as a
// *CloseError so callers can detect a clean shutdown.
func (c *Conn) handleControl(opcode int, payload []byte) error {
	switch opcode {
	case PingMessage:
		_ = c.WriteControl(PongMessage, payload, time.Now().Add(time.Second))
		return nil
	case PongMessage:
		return nil
	case CloseMessage:
		closeCode := CloseNoStatusReceived
		closeText := ""
		if len(payload) >= 2 {
			closeCode = int(binary.BigEndian.Uint16(payload[:2]))
			closeText = string(payload[2:])
		}
		_ = c.WriteControl(CloseMessage, FormatCloseMessage(closeCode, ""), time.Now().Add(time.Second))
		return &CloseError{Code: closeCode, Text: closeText}
	default:
		return c.protocolError("unknown control opcode")
	}
}

// protocolError closes the connection and returns a *CloseError describing a
// protocol violation.
func (c *Conn) protocolError(text string) error {
	_ = c.conn.Close()
	return &CloseError{Code: CloseProtocolError, Text: text}
}

// messageReader streams the payload of a single data message, transparently
// pulling continuation frames until the final frame is consumed.
type messageReader struct {
	c *Conn
}

func (r *messageReader) Read(p []byte) (int, error) {
	c := r.c
	if c.readErr != nil {
		return 0, c.readErr
	}
	for {
		if c.readRemaining > 0 {
			if int64(len(p)) > c.readRemaining {
				p = p[:c.readRemaining]
			}
			n, err := c.br.Read(p)
			c.readRemaining -= int64(n)
			if c.readMasked {
				maskBytes(c.readMaskKey, &c.readMaskPos, p[:n])
			}
			c.readLength += int64(n)
			if c.readLimit > 0 && c.readLength > c.readLimit {
				_ = c.WriteControl(CloseMessage, FormatCloseMessage(CloseMessageTooBig, ""), time.Now().Add(time.Second))
				c.readErr = &CloseError{Code: CloseMessageTooBig, Text: "read limit exceeded"}
				return n, c.readErr
			}
			if err != nil {
				c.readErr = err
			}
			return n, c.readErr
		}

		if c.readFinal {
			c.readErr = io.EOF
			return 0, io.EOF
		}

		opcode, err := c.advanceFrame()
		if err != nil {
			c.readErr = err
			return 0, err
		}
		if opcode != continuationFrame {
			c.readErr = c.protocolError("expected continuation frame")
			return 0, c.readErr
		}
	}
}

// NextWriter returns a writer for the next message of the given type. The
// message is sent as a single frame when the returned writer is closed.
func (c *Conn) NextWriter(messageType int) (io.WriteCloser, error) {
	return &messageWriter{c: c, messageType: messageType}, nil
}

// messageWriter buffers a message and flushes it as one frame on Close.
type messageWriter struct {
	c           *Conn
	messageType int
	buf         []byte
	closed      bool
}

func (w *messageWriter) Write(p []byte) (int, error) {
	if w.closed {
		return 0, errWriteClosed
	}
	w.buf = append(w.buf, p...)
	return len(p), nil
}

func (w *messageWriter) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true
	return w.c.WriteMessage(w.messageType, w.buf)
}

// WriteMessage writes a single data or control message as one frame.
func (c *Conn) WriteMessage(messageType int, data []byte) error {
	return c.writeFrame(messageType, true, data)
}

// WriteControl writes a control message. The deadline bounds the write; a zero
// deadline applies no additional bound beyond any configured write deadline.
func (c *Conn) WriteControl(messageType int, data []byte, deadline time.Time) error {
	if len(data) > maxControlFramePayload {
		return errors.New("websocket: invalid control frame")
	}
	return c.writeFrameDeadline(messageType, true, data, deadline)
}

// WritePreparedMessage writes a previously prepared message.
func (c *Conn) WritePreparedMessage(pm *PreparedMessage) error {
	return c.WriteMessage(pm.messageType, pm.data)
}

func (c *Conn) writeFrame(opcode int, final bool, data []byte) error {
	return c.writeFrameDeadline(opcode, final, data, time.Time{})
}

// writeFrameDeadline encodes and writes a single frame. Client frames are masked
// with a fresh key as required by RFC 6455, section 5.3.
func (c *Conn) writeFrameDeadline(opcode int, final bool, data []byte, deadline time.Time) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	var header [14]byte
	header[0] = byte(opcode)
	if final {
		header[0] |= finBit
	}

	length := len(data)
	n := 2
	var mask byte
	if !c.isServer {
		mask = maskBit
	}
	switch {
	case length < 126:
		header[1] = mask | byte(length)
	case length < 1<<16:
		header[1] = mask | 126
		binary.BigEndian.PutUint16(header[2:4], uint16(length))
		n = 4
	default:
		header[1] = mask | 127
		binary.BigEndian.PutUint64(header[2:10], uint64(length))
		n = 10
	}

	var payload []byte
	if c.isServer {
		payload = data
	} else {
		var key [4]byte
		if _, err := rand.Read(key[:]); err != nil {
			return err
		}
		copy(header[n:n+4], key[:])
		n += 4
		payload = make([]byte, length)
		copy(payload, data)
		pos := 0
		maskBytes(key, &pos, payload)
	}

	frame := make([]byte, 0, n+length)
	frame = append(frame, header[:n]...)
	frame = append(frame, payload...)

	wd := c.writeDeadline
	if !deadline.IsZero() {
		wd = deadline
	}
	if err := c.conn.SetWriteDeadline(wd); err != nil {
		return err
	}

	_, err := c.conn.Write(frame)
	return err
}

// maskBytes applies the WebSocket masking transform to b in place, continuing
// from the mask offset pointed to by pos so a frame can be masked across calls.
func maskBytes(key [4]byte, pos *int, b []byte) {
	p := *pos
	for i := range b {
		b[i] ^= key[p&3]
		p++
	}
	*pos = p
}
