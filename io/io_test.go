package io

import (
	"context"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/zishang520/socket.io/servers/socket/v3"
)

// mockConnectionHandler implements the ConnectionHandler interface for testing purposes.
type mockConnectionHandler struct {
	onFunc    func(event string, handlers ...EventListener)
	closeFunc func(func(error)) error
}

// On records or handles event registration.
func (m *mockConnectionHandler) On(event string, handlers ...EventListener) {
	if m.onFunc != nil {
		m.onFunc(event, handlers...)
	}
}

// Close simulates closing the connection handler.
func (m *mockConnectionHandler) Close(callback func(error)) {
	if m.closeFunc != nil {
		m.closeFunc(callback)
	}
}

func TestServer_IO(t *testing.T) {
	t.Run("wildcard origin", func(t *testing.T) {
		server := &Server{}
		mux := http.NewServeMux()

		ioServer := server.IO(mux, "127.0.0.1:0", []string{"*"})
		defer server.HttpServer.Close(nil)

		assert.True(t, ioServer != nil)
		// Internal check of the HttpServer assignment
		assert.True(t, server.HttpServer != nil)
	})

	t.Run("explicit origins", func(t *testing.T) {
		server := &Server{}
		mux := http.NewServeMux()
		allowedOrigins := []string{"http://localhost:3000", "https://example.com"}

		ioServer := server.IO(mux, "127.0.0.1:0", allowedOrigins)
		defer server.HttpServer.Close(nil)

		assert.True(t, ioServer != nil)
		assert.True(t, server.HttpServer != nil)
	})
}

func TestInitSockets(t *testing.T) {
	var onEventCalled string
	var connectionCallback EventListener
	var closeCalled bool

	mockIO := &mockConnectionHandler{
		onFunc: func(event string, handlers ...EventListener) {
			onEventCalled = event
			if len(handlers) > 0 {
				connectionCallback = handlers[0]
			}
		},
		closeFunc: func(callback func(error)) error {
			closeCalled = true
			return nil
		},
	}

	sockets := []Socket{
		{
			Name: "test-event",
			Handler: func(client *socket.Socket) func(args ...any) {
				return func(args ...any) {}
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Instantly trigger cancellation to avoid blocking during the test execution
	cancel()

	var exitCode int
	var exitCalled bool

	closer := SocketCloser{
		Notifier: func() (context.Context, context.CancelFunc) {
			return ctx, func() {}
		},
		Exit: func(code int) {
			exitCode = code
			exitCalled = true
		},
	}

	InitSockets(mockIO, sockets, closer)

	assert.Equal(t, "connection", onEventCalled)
	assert.True(t, connectionCallback != nil)
	assert.True(t, closeCalled)
	assert.True(t, exitCalled)
	assert.Equal(t, 0, exitCode)
}

func TestConnectionListener(t *testing.T) {
	t.Run("empty clients payload", func(t *testing.T) {
		sockets := []Socket{
			{
				Name: "chat",
				Handler: func(client *socket.Socket) func(args ...any) {
					return func(args ...any) {}
				},
			},
		}

		listener := connectionListener(sockets)
		// Should return gracefully without panicking
		listener()
	})

	t.Run("invalid client type", func(t *testing.T) {
		sockets := []Socket{
			{
				Name: "chat",
				Handler: func(client *socket.Socket) func(args ...any) {
					return func(args ...any) {}
				},
			},
		}

		listener := connectionListener(sockets)
		// Passing an invalid type (string) instead of *socket.Socket
		listener("invalid-client-type")
	})

}
