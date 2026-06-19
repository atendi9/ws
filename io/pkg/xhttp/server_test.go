package xhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/atendi9/capivara/assert"
)

// TestNewServer verifies if the [xhttp.Server] is correctly initialized.
func TestNewServer(t *testing.T) {
	// Assuming NewServeMux returns a Mux
	s := NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	assert.True(t, s != nil)
	assert.True(t, s.Emitter != nil)
	assert.True(t, s.Mux != nil)
}

// TestServeHTTP verifies that the [xhttp.Server] delegates requests to the [xhttp.Mux].
func TestServeHTTP(t *testing.T) {
	called := false
	mockMux := &MockMux{
		ServeHTTPFunc: func(w http.ResponseWriter, r *http.Request) {
			called = true
		},
	}

	s := NewServer(nil)
	s.Mux = mockMux

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	assert.True(t, called)
}

// TestListenAndClose verifies that the server starts, registers, and shuts down correctly.
func TestListenAndClose(t *testing.T) {
	s := NewServer(http.DefaultServeMux)

	// Use port 0 to let the OS choose a random available port
	addr := "127.0.0.1:0"
	listeningCalled := false

	// Test Listen
	srv := s.Listen(addr, func() {
		listeningCalled = true
	})

	// Wait briefly for the server to spin up
	time.Sleep(100 * time.Millisecond)

	assert.True(t, listeningCalled)
	assert.True(t, srv != nil)

	// Test Close
	err := s.Close(func(err error) {
		assert.NoError(t, err)
	})

	assert.NoError(t, err)
}

// TestCloseWithUnknownServerType verifies error handling when closing invalid servers.
func TestCloseWithUnknownServerType(t *testing.T) {
	s := NewServer(http.DefaultServeMux)

	// Manually inject an invalid type to trigger the error path
	s.activeServers.Push("invalid-server-type")

	err := s.Close(nil)

	assert.Error(t, err)
}

// MockMux is a mock implementation of the [xhttp.Mux] interface for testing purposes.
type MockMux struct {
	ServeHTTPFunc  func(w http.ResponseWriter, r *http.Request)
	HandleFuncFunc func(path string, handler http.HandlerFunc)
}

// ServeHTTP delegates to the [MockMux.ServeHTTPFunc] if defined.
func (m *MockMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.ServeHTTPFunc != nil {
		m.ServeHTTPFunc(w, r)
	}
}

// HandleFunc delegates to the [MockMux.HandleFuncFunc] if defined.
func (m *MockMux) HandleFunc(path string, handler http.HandlerFunc) {
	if m.HandleFuncFunc != nil {
		m.HandleFuncFunc(path, handler)
	}
}
