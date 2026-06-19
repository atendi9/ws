package wsconn

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestConnAccessors(t *testing.T) {
	srv := httptest.NewServer(echoServer(t))
	defer srv.Close()

	c := dialTest(t, srv)
	defer c.Close()

	if c.UnderlyingConn() == nil {
		t.Fatal("UnderlyingConn returned nil")
	}
	if c.LocalAddr() == nil {
		t.Fatal("LocalAddr returned nil")
	}
	if c.RemoteAddr() == nil {
		t.Fatal("RemoteAddr returned nil")
	}
	// Default (no subprotocol negotiated) is empty.
	if c.Subprotocol() != "" {
		t.Fatalf("expected empty subprotocol, got %q", c.Subprotocol())
	}

	if err := c.SetReadDeadline(time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("SetReadDeadline: %v", err)
	}
	if err := c.SetWriteDeadline(time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("SetWriteDeadline: %v", err)
	}
	// No-op for API compatibility; must not panic.
	c.EnableWriteCompression(true)
	c.SetReadLimit(1 << 20)
}

func TestConnNextWriterAndWrite(t *testing.T) {
	srv := httptest.NewServer(echoServer(t))
	defer srv.Close()

	c := dialTest(t, srv)
	defer c.Close()

	w, err := c.NextWriter(TextMessage)
	if err != nil {
		t.Fatalf("NextWriter: %v", err)
	}
	if _, err := w.Write([]byte("chunk-one ")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := io.WriteString(w, "chunk-two"); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("writer Close: %v", err)
	}

	mt, got := readString(t, c)
	if mt != TextMessage || got != "chunk-one chunk-two" {
		t.Fatalf("echo mismatch: mt=%d got=%q", mt, got)
	}
}

func TestConnWritePreparedMessage(t *testing.T) {
	srv := httptest.NewServer(echoServer(t))
	defer srv.Close()

	c := dialTest(t, srv)
	defer c.Close()

	pm, err := NewPreparedMessage(BinaryMessage, []byte("prep"))
	if err != nil {
		t.Fatalf("NewPreparedMessage: %v", err)
	}
	if err := c.WritePreparedMessage(pm); err != nil {
		t.Fatalf("WritePreparedMessage: %v", err)
	}
	mt, got := readString(t, c)
	if mt != BinaryMessage || got != "prep" {
		t.Fatalf("prepared echo mismatch: mt=%d got=%q", mt, got)
	}
}

func TestConnSubprotocolNegotiation(t *testing.T) {
	up := &Upgrader{
		CheckOrigin:  func(*http.Request) bool { return true },
		Subprotocols: []string{"chat.v2", "chat.v1"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := up.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		if c.Subprotocol() != "chat.v1" {
			t.Errorf("server subprotocol = %q, want chat.v1", c.Subprotocol())
		}
		c.Close()
	}))
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	d := &Dialer{HandshakeTimeout: 5 * time.Second, Subprotocols: []string{"chat.v1"}}
	c, _, err := d.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close()
	if c.Subprotocol() != "chat.v1" {
		t.Fatalf("client subprotocol = %q, want chat.v1", c.Subprotocol())
	}
}

func TestCloseError(t *testing.T) {
	cases := []struct {
		code int
		text string
		want string
	}{
		{CloseNormalClosure, "", "websocket: close 1000<normal>"},
		{CloseGoingAway, "leaving", "websocket: close 1001<going away>: leaving"},
		{CloseProtocolError, "", "websocket: close 1002<protocol error>"},
		{CloseMessageTooBig, "big", "websocket: close 1009<message too big>: big"},
		{9999, "", "websocket: close 9999"},
	}
	for _, tc := range cases {
		ce := &CloseError{Code: tc.code, Text: tc.text}
		if got := ce.Error(); got != tc.want {
			t.Errorf("CloseError(%d,%q).Error() = %q, want %q", tc.code, tc.text, got, tc.want)
		}
	}
}

func TestIsUnexpectedCloseError(t *testing.T) {
	if !IsUnexpectedCloseError(&CloseError{Code: CloseNormalClosure}) {
		t.Error("expected true for *CloseError")
	}
	if IsUnexpectedCloseError(errors.New("plain")) {
		t.Error("expected false for non-CloseError")
	}
	if IsUnexpectedCloseError(nil) {
		t.Error("expected false for nil")
	}
}

func TestHandshakeErrorError(t *testing.T) {
	err := HandshakeError{message: "bad handshake"}
	if err.Error() != "bad handshake" {
		t.Fatalf("HandshakeError.Error() = %q", err.Error())
	}
}

func TestIsWebSocketUpgrade(t *testing.T) {
	t.Run("valid upgrade request", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Connection", "Upgrade")
		r.Header.Set("Upgrade", "websocket")
		if !IsWebSocketUpgrade(r) {
			t.Error("expected true for valid upgrade request")
		}
	})
	t.Run("plain request", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		if IsWebSocketUpgrade(r) {
			t.Error("expected false for plain request")
		}
	})
}

func TestNewPreparedMessageContents(t *testing.T) {
	pm, err := NewPreparedMessage(TextMessage, []byte("hi"))
	if err != nil {
		t.Fatalf("NewPreparedMessage: %v", err)
	}
	if pm == nil {
		t.Fatal("expected non-nil prepared message")
	}
}

func TestUpgradeRejectsNonUpgrade(t *testing.T) {
	up := &Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := up.Upgrade(w, r, nil)
		if err == nil {
			t.Error("expected error upgrading a plain GET request")
		}
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusSwitchingProtocols {
		t.Fatal("plain request must not be upgraded")
	}
}
