package wsconn

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Dialer contains options for connecting to a WebSocket server.
type Dialer struct {
	// NetDial specifies the dial function for creating TCP connections. When nil,
	// net.Dial is used.
	NetDial func(network, addr string) (net.Conn, error)

	// Proxy specifies a function to return a proxy for a given request.
	Proxy func(*http.Request) (*url.URL, error)

	// TLSClientConfig specifies the TLS configuration to use with tls.Client for
	// wss connections.
	TLSClientConfig *tls.Config

	// HandshakeTimeout bounds the duration of the WebSocket handshake.
	HandshakeTimeout time.Duration

	// ReadBufferSize and WriteBufferSize specify I/O buffer sizes in bytes.
	ReadBufferSize  int
	WriteBufferSize int

	// Subprotocols lists the client's supported protocols in order of preference.
	Subprotocols []string

	// EnableCompression is accepted for API compatibility. permessage-deflate is
	// not negotiated, so this field has no effect on the connection.
	EnableCompression bool

	// Jar specifies the cookie jar used to populate request cookies.
	Jar http.CookieJar
}

// Dial connects to the WebSocket server at urlStr and performs the client
// handshake, returning the connection and the HTTP response from the server.
func (d *Dialer) Dial(urlStr string, requestHeader http.Header) (*Conn, *http.Response, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, nil, err
	}

	var secure bool
	switch u.Scheme {
	case "ws":
		secure = false
	case "wss":
		secure = true
	default:
		return nil, nil, ErrBadHandshake
	}

	hostPort := u.Host
	if u.Port() == "" {
		if secure {
			hostPort = net.JoinHostPort(u.Host, "443")
		} else {
			hostPort = net.JoinHostPort(u.Host, "80")
		}
	}

	netDial := d.NetDial
	if netDial == nil {
		netDial = net.Dial
	}

	netConn, err := netDial("tcp", hostPort)
	if err != nil {
		return nil, nil, err
	}

	if secure {
		cfg := d.TLSClientConfig
		if cfg == nil {
			cfg = &tls.Config{ServerName: u.Hostname()}
		} else if cfg.ServerName == "" {
			cfg = cfg.Clone()
			cfg.ServerName = u.Hostname()
		}
		tlsConn := tls.Client(netConn, cfg)
		if err := tlsConn.Handshake(); err != nil {
			_ = netConn.Close()
			return nil, nil, err
		}
		netConn = tlsConn
	}

	challengeKey, err := generateChallengeKey()
	if err != nil {
		_ = netConn.Close()
		return nil, nil, err
	}

	req := &http.Request{
		Method:     http.MethodGet,
		URL:        u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Host:       u.Host,
	}
	for k, vs := range requestHeader {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	if d.Jar != nil {
		for _, cookie := range d.Jar.Cookies(u) {
			req.AddCookie(cookie)
		}
	}
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", challengeKey)
	req.Header.Set("Sec-WebSocket-Version", "13")
	if len(d.Subprotocols) > 0 {
		req.Header.Set("Sec-WebSocket-Protocol", strings.Join(d.Subprotocols, ", "))
	}

	if d.HandshakeTimeout != 0 {
		if err := netConn.SetDeadline(time.Now().Add(d.HandshakeTimeout)); err != nil {
			_ = netConn.Close()
			return nil, nil, err
		}
	}

	if err := req.Write(netConn); err != nil {
		_ = netConn.Close()
		return nil, nil, err
	}

	br := bufio.NewReader(netConn)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		_ = netConn.Close()
		return nil, nil, err
	}
	if d.Jar != nil {
		if rc := resp.Cookies(); len(rc) > 0 {
			d.Jar.SetCookies(u, rc)
		}
	}

	if resp.StatusCode != http.StatusSwitchingProtocols ||
		!tokenListContainsValue(resp.Header, "Upgrade", "websocket") ||
		!tokenListContainsValue(resp.Header, "Connection", "upgrade") ||
		resp.Header.Get("Sec-Websocket-Accept") != computeAcceptKey(challengeKey) {
		_ = netConn.Close()
		return nil, resp, ErrBadHandshake
	}

	if d.HandshakeTimeout != 0 {
		if err := netConn.SetDeadline(time.Time{}); err != nil {
			_ = netConn.Close()
			return nil, resp, err
		}
	}

	conn := newConn(netConn, false, br, d.ReadBufferSize, resp.Header.Get("Sec-Websocket-Protocol"))
	resp.Body = noBody{}
	return conn, resp, nil
}

// noBody is an empty response body for the handshake response.
type noBody struct{}

func (noBody) Read([]byte) (int, error) { return 0, io.EOF }
func (noBody) Close() error             { return nil }

// generateChallengeKey returns a base64-encoded 16-byte nonce for the handshake.
func generateChallengeKey() (string, error) {
	p := make([]byte, 16)
	if _, err := rand.Read(p); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(p), nil
}
