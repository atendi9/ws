package xhttp_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/ws/io/pkg/xhttp"
)

func TestNewContext(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test/path?foo=bar", nil)
	r.Header.Set("User-Agent", "TestAgent")

	c := xhttp.NewContext(w, r)

	assert.Equal(t, http.StatusOK, c.GetStatusCode())
	assert.Equal(t, "GET", c.Method())
	assert.Equal(t, "test/path", c.Path())
	assert.Equal(t, "TestAgent", c.UserAgent())
	assert.False(t, c.IsDone())
}

func TestSetStatusCode(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c := xhttp.NewContext(w, r)

	// Valid status code
	err := c.SetStatusCode(http.StatusCreated)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, c.GetStatusCode())

	// Invalid status code
	err = c.SetStatusCode(999)
	assert.Error(t, err)
	assert.Equal(t, http.StatusCreated, c.GetStatusCode()) // Should remain unchanged
}

func TestWrite(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello"))
	c := xhttp.NewContext(w, r)

	c.SetStatusCode(http.StatusAccepted)
	n, err := c.Write([]byte("response body"))

	assert.NoError(t, err)
	assert.Equal(t, 13, n)
	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Equal(t, "response body", w.Body.String())
	assert.True(t, c.IsDone())
}

func TestWrite_DoubleWrite(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c := xhttp.NewContext(w, r)

	_, err := c.Write([]byte("first write"))
	assert.NoError(t, err)

	_, err = c.Write([]byte("second write"))
	assert.Error(t, err)
}

func TestFlush(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c := xhttp.NewContext(w, r)

	c.Flush()

	assert.True(t, c.IsDone())

	// Ensure that subsequent write attempts fail
	_, err := c.Write([]byte("too late"))
	assert.Error(t, err)
}

func TestCleanup(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c := xhttp.NewContext(w, r)

	cleanupCalled := 0
	c.Cleanup = func() {
		cleanupCalled++
	}

	c.Flush() // Triggers cleanup
	c.Flush() // Should not trigger cleanup again

	assert.Equal(t, 1, cleanupCalled)
}

func TestResponseHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c := xhttp.NewContext(w, r)

	headers := c.ResponseHeaders()
	headers.Set("X-Custom-Header", "value1")
	headers.Add("X-Multi-Header", "a")
	headers.Add("X-Multi-Header", "b")

	_, err := c.Write([]byte("body"))
	assert.NoError(t, err)

	assert.Equal(t, "value1", w.Header().Get("X-Custom-Header"))
	assert.Equal(t, fmt.Sprint([]string{"a", "b"}), fmt.Sprint(w.Header().Values("X-Multi-Header")))
}
