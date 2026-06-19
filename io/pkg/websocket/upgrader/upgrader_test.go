package upgrader

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestNew(t *testing.T) {
	called := false
	mockErrorHandler := func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		called = true
	}

	def := Definition{
		Size: BufferSize{
			Reader: 1024,
			Writer: 2048,
		},
		EnableCompression: true,
		ErrorHandler:      mockErrorHandler,
	}

	u := New(def)

	assert.Equal(t, def.Size.Reader, u.ReadBufferSize)
	assert.Equal(t, def.Size.Writer, u.WriteBufferSize)
	assert.True(t, u.EnableCompression)

	// Verify if the error handler was correctly assigned by executing it
	if u.Error != nil {
		u.Error(nil, nil, 0, nil)
	}
	assert.True(t, called)
}

func TestIs(t *testing.T) {
	t.Run("should return false when request is not a websocket upgrade", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://localhost/test", nil)

		result := Is(req)

		assert.False(t, result)
	})

	t.Run("should return true when request is a valid websocket upgrade", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://localhost/test", nil)
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Connection", "Upgrade")

		result := Is(req)

		assert.True(t, result)
	})
}
