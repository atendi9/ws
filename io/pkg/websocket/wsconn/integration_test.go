package wsconn

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// echoServer upgrades the request and echoes every data message back to the
// client until the connection closes.
func echoServer(t *testing.T) http.HandlerFunc {
	t.Helper()
	up := &Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := up.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		for {
			mt, rd, err := c.NextReader()
			if err != nil {
				return
			}
			data, _ := io.ReadAll(rd)
			if err := c.WriteMessage(mt, data); err != nil {
				return
			}
		}
	}
}

func dialTest(t *testing.T, srv *httptest.Server) *Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	d := &Dialer{HandshakeTimeout: 5 * time.Second}
	c, resp, err := d.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	return c
}

func readString(t *testing.T, c *Conn) (int, string) {
	t.Helper()
	mt, rd, err := c.NextReader()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	data, err := io.ReadAll(rd)
	if err != nil {
		t.Fatalf("readall: %v", err)
	}
	return mt, string(data)
}

func TestHandshakeAndEcho(t *testing.T) {
	srv := httptest.NewServer(echoServer(t))
	defer srv.Close()

	c := dialTest(t, srv)
	defer c.Close()

	cases := []struct {
		mt   int
		text string
	}{
		{TextMessage, "hello"},
		{BinaryMessage, "world"},
		{TextMessage, ""},
		{TextMessage, strings.Repeat("x", 200)},    // 16-bit length path
		{BinaryMessage, strings.Repeat("y", 70000)}, // 64-bit length path
	}
	for _, tc := range cases {
		if err := c.WriteMessage(tc.mt, []byte(tc.text)); err != nil {
			t.Fatalf("write: %v", err)
		}
		mt, got := readString(t, c)
		if mt != tc.mt || got != tc.text {
			t.Fatalf("echo mismatch: mt=%d len=%d want mt=%d len=%d", mt, len(got), tc.mt, len(tc.text))
		}
	}
}

func TestFragmentedMessage(t *testing.T) {
	srv := httptest.NewServer(echoServer(t))
	defer srv.Close()

	c := dialTest(t, srv)
	defer c.Close()

	// NextWriter buffers into a single frame; force fragmentation manually via
	// writeFrameDeadline to exercise the continuation path in messageReader.
	if err := c.writeFrameDeadline(TextMessage, false, []byte("frag-"), time.Time{}); err != nil {
		t.Fatalf("write frame 1: %v", err)
	}
	if err := c.writeFrameDeadline(continuationFrame, true, []byte("mented"), time.Time{}); err != nil {
		t.Fatalf("write frame 2: %v", err)
	}
	mt, got := readString(t, c)
	if mt != TextMessage || got != "frag-mented" {
		t.Fatalf("fragmented echo: mt=%d got=%q", mt, got)
	}
}

func TestPingHandledDuringRead(t *testing.T) {
	srv := httptest.NewServer(echoServer(t))
	defer srv.Close()

	c := dialTest(t, srv)
	defer c.Close()

	// A ping should be answered by the server with a pong and not surface to the
	// data reader. The subsequent text message must still arrive intact.
	if err := c.WriteControl(PingMessage, []byte("pingdata"), time.Now().Add(time.Second)); err != nil {
		t.Fatalf("ping: %v", err)
	}
	if err := c.WriteMessage(TextMessage, []byte("after-ping")); err != nil {
		t.Fatalf("write: %v", err)
	}
	// The client read loop also transparently swallows the server's pong.
	mt, got := readString(t, c)
	if mt != TextMessage || got != "after-ping" {
		t.Fatalf("post-ping echo: mt=%d got=%q", mt, got)
	}
}

func TestCloseFrameSurfacesAsCloseError(t *testing.T) {
	up := &Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := up.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		_, _, err = c.NextReader()
		if !IsUnexpectedCloseError(err) {
			t.Errorf("expected CloseError, got %v", err)
		}
		ce, ok := err.(*CloseError)
		if !ok || ce.Code != CloseNormalClosure {
			t.Errorf("expected normal closure, got %+v", err)
		}
	}))
	defer srv.Close()

	c := dialTest(t, srv)
	if err := c.WriteMessage(CloseMessage, FormatCloseMessage(CloseNormalClosure, "bye")); err != nil {
		t.Fatalf("write close: %v", err)
	}
	// Give the server goroutine time to observe the close.
	time.Sleep(100 * time.Millisecond)
	c.Close()
}

func TestReadLimitExceeded(t *testing.T) {
	up := &Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := up.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		c.SetReadLimit(8)
		_, rd, err := c.NextReader()
		if err == nil {
			_, err = io.ReadAll(rd)
		}
		if !IsUnexpectedCloseError(err) {
			t.Errorf("expected CloseError for over-limit message, got %v", err)
		}
	}))
	defer srv.Close()

	c := dialTest(t, srv)
	defer c.Close()
	_ = c.WriteMessage(TextMessage, []byte("this is definitely longer than eight bytes"))
	time.Sleep(100 * time.Millisecond)
}
