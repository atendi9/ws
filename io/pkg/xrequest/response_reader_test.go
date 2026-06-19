package xrequest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func newResponse(t *testing.T, status int, body string) *Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	return &Response{&http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}}
}

func TestResponseOk(t *testing.T) {
	tests := []struct {
		status int
		want   bool
	}{
		{200, true},
		{204, true},
		{299, true},
		{300, false},
		{404, false},
		{500, false},
		{199, false},
	}
	for _, tt := range tests {
		r := newResponse(t, tt.status, "")
		assert.Equal(t, tt.want, r.Ok())
	}
}

func TestResponseErr(t *testing.T) {
	t.Run("ok response has no error", func(t *testing.T) {
		r := newResponse(t, 200, "fine")
		assert.NoError(t, r.Err())
	})

	t.Run("non-ok response returns error with body", func(t *testing.T) {
		r := newResponse(t, 500, "boom")
		err := r.Err()
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "boom"))
	})
}

func TestDummyReader(t *testing.T) {
	d := &dummyReader{body: io.NopCloser(strings.NewReader("payload"))}

	out, err := io.ReadAll(d)
	assert.NoError(t, err)
	assert.Equal(t, "payload", string(out))

	assert.NoError(t, d.Close())
}
