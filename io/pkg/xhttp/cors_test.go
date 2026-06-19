package xhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atendi9/capivara/assert"
)

// TestIsOriginAllowed verifies the origin validation logic for different input types.
func TestIsOriginAllowed(t *testing.T) {
	c := &Cors{}

	// Test case: Wildcard
	assert.True(t, c.IsOriginAllowed("http://example.com", "*"))

	// Test case: Exact string match
	assert.True(t, c.IsOriginAllowed("http://example.com", "http://example.com"))
	assert.False(t, c.IsOriginAllowed("http://example.com", "http://other.com"))

	// Test case: Slice of strings
	allowedSlice := []string{"http://a.com", "http://b.com"}
	assert.True(t, c.IsOriginAllowed("http://a.com", allowedSlice))
	assert.False(t, c.IsOriginAllowed("http://c.com", allowedSlice))

	// Test case: Slice of any
	allowedAny := []any{"http://a.com", "http://b.com"}
	assert.True(t, c.IsOriginAllowed("http://b.com", allowedAny))

	// Test case: Function
	fn := func(origin string) bool {
		return origin == "http://dynamic.com"
	}
	assert.True(t, c.IsOriginAllowed("http://dynamic.com", fn))
	assert.False(t, c.IsOriginAllowed("http://other.com", fn))

	// Test case: Boolean
	assert.True(t, c.IsOriginAllowed("anything", true))
	assert.False(t, c.IsOriginAllowed("anything", false))
}

// TestMiddlewareWrapper verifies the initialization of defaults and function returned by [MiddlewareWrapper].
func TestMiddlewareWrapper(t *testing.T) {
	// Test case: Nil options (should use defaults)
	middleware := MiddlewareWrapper(nil)
	assert.True(t, middleware != nil)

	// Test case: Custom options
	custom := &Cors{
		Origin: "http://custom.com",
	}
	middlewareCustom := MiddlewareWrapper(custom)
	assert.True(t, middlewareCustom != nil)
	assert.Equal(t, "http://custom.com", custom.Origin)
	assert.Equal(t, `GET,HEAD,PUT,PATCH,POST,DELETE`, custom.Methods)
	assert.Equal(t, 204, custom.OptionsSuccessStatus)
}

// TestCorsMiddleware_Preflight verifies the handling of OPTIONS preflight requests.
// This test relies on a mocked [xhttp.Context] environment.
func TestCorsMiddleware_Preflight(t *testing.T) {
	r, _ := http.NewRequest(http.MethodOptions, "/hello", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, r)
	options := &Cors{
		Origin:               "*",
		Methods:              "GET,POST",
		OptionsSuccessStatus: 200,
	}

	nextCalled := false
	next := func(err error) {
		nextCalled = true
	}

	CorsMiddleware(ctx, options, next)

	// Agora nextCalled deve ser false
	assert.False(t, nextCalled)
}

// TestCorsMiddleware_Actual verifies the handling of standard cross-origin requests.
func TestCorsMiddleware_Actual(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, r)
	options := &Cors{
		Origin: "http://example.com",
	}

	nextCalled := false
	next := func(err error) {
		nextCalled = true
		assert.NoError(t, err)
	}

	CorsMiddleware(ctx, options, next)

	// Verify next was called
	assert.True(t, nextCalled)
}
