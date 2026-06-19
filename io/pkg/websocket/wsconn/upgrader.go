package wsconn

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"
)

// keyGUID is the globally unique identifier appended to the client key when
// computing the handshake accept value (RFC 6455, section 1.3).
const keyGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// HandshakeError describes a failed WebSocket handshake.
type HandshakeError struct {
	message string
}

func (e HandshakeError) Error() string { return e.message }

// ErrBadHandshake is returned by the client dialer when the server response does
// not complete the WebSocket handshake.
var ErrBadHandshake = errors.New("websocket: bad handshake")

// Upgrader upgrades an HTTP server connection to the WebSocket protocol.
type Upgrader struct {
	// ReadBufferSize and WriteBufferSize specify I/O buffer sizes in bytes. A
	// value of zero selects a default size.
	ReadBufferSize  int
	WriteBufferSize int

	// EnableCompression is accepted for API compatibility. permessage-deflate is
	// not negotiated, so this field has no effect on the connection.
	EnableCompression bool

	// Subprotocols lists the server's supported protocols in order of preference.
	Subprotocols []string

	// Error specifies the function used to generate HTTP error responses. When
	// nil, http.Error is used.
	Error func(w http.ResponseWriter, r *http.Request, status int, reason error)

	// CheckOrigin returns true if the request Origin header is acceptable. When
	// nil, a same-origin policy is applied.
	CheckOrigin func(r *http.Request) bool
}

// returnError reports a handshake failure to the client and returns the error.
func (u *Upgrader) returnError(w http.ResponseWriter, r *http.Request, status int, reason string) (*Conn, error) {
	err := HandshakeError{reason}
	if u.Error != nil {
		u.Error(w, r, status, err)
	} else {
		w.Header().Set("Sec-Websocket-Version", "13")
		http.Error(w, http.StatusText(status), status)
	}
	return nil, err
}

// checkSameOrigin reports whether the Origin header matches the Host header.
func checkSameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	_, host, ok := strings.Cut(origin, "://")
	if !ok {
		return false
	}
	return strings.EqualFold(host, r.Host)
}

// selectSubprotocol negotiates a subprotocol from the client request against the
// upgrader's supported list.
func (u *Upgrader) selectSubprotocol(r *http.Request) string {
	if len(u.Subprotocols) == 0 {
		return ""
	}
	requested := parseTokenList(r.Header.Get("Sec-Websocket-Protocol"))
	for _, server := range u.Subprotocols {
		for _, client := range requested {
			if strings.EqualFold(server, client) {
				return server
			}
		}
	}
	return ""
}

// Upgrade performs the server handshake and returns a WebSocket connection.
func (u *Upgrader) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*Conn, error) {
	if r.Method != http.MethodGet {
		return u.returnError(w, r, http.StatusMethodNotAllowed, "websocket: not a GET request")
	}
	if !tokenListContainsValue(r.Header, "Connection", "upgrade") {
		return u.returnError(w, r, http.StatusBadRequest, "websocket: 'upgrade' token not found in 'Connection' header")
	}
	if !tokenListContainsValue(r.Header, "Upgrade", "websocket") {
		return u.returnError(w, r, http.StatusBadRequest, "websocket: 'websocket' token not found in 'Upgrade' header")
	}
	if !tokenListContainsValue(r.Header, "Sec-Websocket-Version", "13") {
		return u.returnError(w, r, http.StatusBadRequest, "websocket: unsupported version")
	}

	checkOrigin := u.CheckOrigin
	if checkOrigin == nil {
		checkOrigin = checkSameOrigin
	}
	if !checkOrigin(r) {
		return u.returnError(w, r, http.StatusForbidden, "websocket: request origin not allowed by Upgrader.CheckOrigin")
	}

	challengeKey := r.Header.Get("Sec-Websocket-Key")
	if challengeKey == "" {
		return u.returnError(w, r, http.StatusBadRequest, "websocket: not a websocket handshake: 'Sec-WebSocket-Key' header is missing or blank")
	}

	subprotocol := u.selectSubprotocol(r)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return u.returnError(w, r, http.StatusInternalServerError, "websocket: response does not implement http.Hijacker")
	}
	netConn, brw, err := hijacker.Hijack()
	if err != nil {
		return u.returnError(w, r, http.StatusInternalServerError, err.Error())
	}

	// Refuse connections with buffered request data we cannot replay.
	if brw.Reader.Buffered() > 0 {
		_ = netConn.Close()
		return nil, errors.New("websocket: client sent data before handshake is complete")
	}

	var b strings.Builder
	b.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
	b.WriteString("Upgrade: websocket\r\n")
	b.WriteString("Connection: Upgrade\r\n")
	b.WriteString("Sec-WebSocket-Accept: ")
	b.WriteString(computeAcceptKey(challengeKey))
	b.WriteString("\r\n")
	if subprotocol != "" {
		b.WriteString("Sec-WebSocket-Protocol: ")
		b.WriteString(subprotocol)
		b.WriteString("\r\n")
	}
	for k, vs := range responseHeader {
		if k == "Sec-Websocket-Protocol" {
			continue
		}
		for _, v := range vs {
			b.WriteString(k)
			b.WriteString(": ")
			b.WriteString(v)
			b.WriteString("\r\n")
		}
	}
	b.WriteString("\r\n")

	if err := netConn.SetWriteDeadline(time.Time{}); err != nil {
		_ = netConn.Close()
		return nil, err
	}
	if _, err := netConn.Write([]byte(b.String())); err != nil {
		_ = netConn.Close()
		return nil, err
	}

	br := brw.Reader
	if u.ReadBufferSize > 0 && br.Size() < u.ReadBufferSize {
		br = bufio.NewReaderSize(netConn, u.ReadBufferSize)
	}
	return newConn(netConn, true, br, u.ReadBufferSize, subprotocol), nil
}

// IsWebSocketUpgrade reports whether the request is requesting an upgrade to the
// WebSocket protocol.
func IsWebSocketUpgrade(r *http.Request) bool {
	return tokenListContainsValue(r.Header, "Connection", "upgrade") &&
		tokenListContainsValue(r.Header, "Upgrade", "websocket")
}

// computeAcceptKey returns the Sec-WebSocket-Accept value for a challenge key.
func computeAcceptKey(challengeKey string) string {
	h := sha1.New()
	h.Write([]byte(challengeKey))
	h.Write([]byte(keyGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// parseTokenList splits a comma-separated header value into trimmed tokens.
func parseTokenList(s string) []string {
	var tokens []string
	for part := range strings.SplitSeq(s, ",") {
		if t := strings.TrimSpace(part); t != "" {
			tokens = append(tokens, t)
		}
	}
	return tokens
}

// tokenListContainsValue reports whether the comma-separated list values of the
// named header contain the given value, matched case-insensitively.
func tokenListContainsValue(header http.Header, name, value string) bool {
	for _, v := range header[http.CanonicalHeaderKey(name)] {
		for _, token := range parseTokenList(v) {
			if strings.EqualFold(token, value) {
				return true
			}
		}
	}
	return false
}
