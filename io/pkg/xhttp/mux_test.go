package xhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atendi9/capivara/assert"
)

// TestNewServeMux validates the initialization of the ServeMux.
func TestNewServeMux(t *testing.T) {
	mux := NewServeMux(nil)

	// Validate that the default handler was set to DefaultServeMux when nil is passed.
	assert.Equal(t, any(http.DefaultServeMux).(http.Handler), mux.DefaultHandler)
}

// TestServeMux_Handle validates the registration of handlers.
func TestServeMux_Handle(t *testing.T) {
	mux := NewServeMux(nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux.Handle("/test", handler)

	// Verify the map registration.
	assert.LengthMap(t, 1, mux.m)
	entry, ok := mux.m["/test"]
	assert.True(t, ok)
	assert.Equal(t, "/test", entry.pattern)

	// Validate panic on duplicate registration.
	defer func() {
		r := recover()
		assert.True(t, r != nil)
	}()
	mux.Handle("/test", handler)
}

// TestServeMux_Handler_Matching verifies the logic for matching paths.
func TestServeMux_Handler_Matching(t *testing.T) {
	mux := NewServeMux(nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r)
		ctx.Write([]byte(ctx.Path()))
	})

	mux.Handle("/api/", handler)
	mux.Handle("/login", handler)

	// Test exact match.
	reqExact, _ := http.NewRequest("GET", "/login", nil)
	h1, pattern1 := mux.Handler(reqExact)
	assert.Equal(t, "/login", pattern1)
	w := httptest.NewRecorder()
	h1.ServeHTTP(w, reqExact)
	assert.Equal(t, w.Body.String(), "login")

	// Test prefix match.
	reqPrefix, _ := http.NewRequest("GET", "/api/users", nil)
	h2, pattern2 := mux.Handler(reqPrefix)
	assert.Equal(t, "/api/", pattern2)
	w = httptest.NewRecorder()
	h2.ServeHTTP(w, reqPrefix)
	assert.Equal(t, w.Body.String(), "api/users")
}

// TestServeMux_ServeHTTP verifies the dispatching of requests.
func TestServeMux_ServeHTTP(t *testing.T) {
	// Create a recorder to capture the response.
	w := httptest.NewRecorder()

	// Create a custom handler to verify it was called.
	called := false
	mux := NewServeMux(nil)
	mux.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	// Create a valid request.
	r := httptest.NewRequest("GET", "/home", nil)

	mux.ServeHTTP(w, r)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestServeMux_DefaultHandler verifies the fallback mechanism.
func TestServeMux_DefaultHandler(t *testing.T) {
	defaultCalled := false
	defaultHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defaultCalled = true
	})

	mux := NewServeMux(defaultHandler)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/unknown-path", nil)

	mux.ServeHTTP(w, r)

	assert.True(t, defaultCalled)
}
