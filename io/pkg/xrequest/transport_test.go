package xrequest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/atendi9/capivara/assert"
)

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	assert.NoError(t, err)
	return u
}

func TestNewTransport(t *testing.T) {
	tr := NewTransport(nil)
	assert.NotNil(t, tr)
	assert.NoError(t, tr.Close())
}

func TestTransportRoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	tr := NewTransport(nil)
	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	resp, err := tr.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "ok", string(body))
}

func TestTransportRoundTripError(t *testing.T) {
	tr := NewTransport(nil)
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:1/", nil)
	_, err := tr.RoundTrip(req)
	assert.Error(t, err)
}

func TestTransportProcessesAndClearsAltSvc(t *testing.T) {
	tr := NewTransport(nil)
	reqURL := mustParseURL(t, "http://example.com/path")

	// Store via header
	tr.processAltSvc(http.Header{"Alt-Svc": {`h2=":443"; ma=3600`}}, reqURL)
	_, ok := tr.altSvcCache.Load(getOrigin(reqURL))
	assert.True(t, ok)

	// "clear" removes it
	tr.processAltSvc(http.Header{"Alt-Svc": {"clear"}}, reqURL)
	_, ok = tr.altSvcCache.Load(getOrigin(reqURL))
	assert.False(t, ok)
}

func TestProcessAltSvcEmptyHeader(t *testing.T) {
	tr := NewTransport(nil)
	reqURL := mustParseURL(t, "http://example.com")
	tr.processAltSvc(http.Header{}, reqURL)
	_, ok := tr.altSvcCache.Load(getOrigin(reqURL))
	assert.False(t, ok)
}

func TestParseAltSvc(t *testing.T) {
	t.Run("single entry with port endpoint", func(t *testing.T) {
		entries := parseAltSvc(`h2=":8443"; ma=7200; persist=1`)
		assert.Equal(t, 1, len(entries))
		assert.Equal(t, "h2", entries[0].protocol)
		assert.Equal(t, ":8443", entries[0].endpoint)
		assert.True(t, entries[0].persist)
		assert.True(t, entries[0].expires.After(time.Now()))
	})

	t.Run("multiple entries", func(t *testing.T) {
		entries := parseAltSvc(`h2=":443", h3=":443"; ma=100`)
		assert.Equal(t, 2, len(entries))
	})

	t.Run("skips malformed entries", func(t *testing.T) {
		entries := parseAltSvc(`garbage, h2=":443"`)
		assert.Equal(t, 1, len(entries))
	})

	t.Run("empty value", func(t *testing.T) {
		assert.Equal(t, 0, len(parseAltSvc("")))
	})

	t.Run("clamps negative ma to zero", func(t *testing.T) {
		entries := parseAltSvc(`h2=":443"; ma=-5`)
		assert.Equal(t, 1, len(entries))
		assert.False(t, entries[0].expires.After(time.Now().Add(time.Second)))
	})

	t.Run("clamps huge ma", func(t *testing.T) {
		entries := parseAltSvc(`h2=":443"; ma=99999999999`)
		assert.Equal(t, 1, len(entries))
	})

	t.Run("ignores unknown and malformed params", func(t *testing.T) {
		entries := parseAltSvc(`h2=":443"; foo; bar=baz; ma=abc`)
		assert.Equal(t, 1, len(entries))
	})
}

func TestIsServiceValid(t *testing.T) {
	valid := &altSvc{expires: time.Now().Add(time.Hour)}
	assert.True(t, isServiceValid(valid))

	expired := &altSvc{expires: time.Now().Add(-time.Hour)}
	assert.False(t, isServiceValid(expired))

	failed := &altSvc{expires: time.Now().Add(time.Hour)}
	failed.failures.Store(maxRetryAttempts)
	assert.False(t, isServiceValid(failed))
}

func TestTryAltServices(t *testing.T) {
	tr := NewTransport(nil)
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	t.Run("no cache", func(t *testing.T) {
		_, err := tr.tryAltServices(req)
		assert.Error(t, err)
	})

	t.Run("invalid cache entry type", func(t *testing.T) {
		tr.altSvcCache.Store(getOrigin(req.URL), "not-a-slice")
		_, err := tr.tryAltServices(req)
		assert.Error(t, err)
		tr.altSvcCache.Delete(getOrigin(req.URL))
	})

	t.Run("all services invalid", func(t *testing.T) {
		tr.altSvcCache.Store(getOrigin(req.URL), []*altSvc{
			{expires: time.Now().Add(-time.Hour)},
		})
		_, err := tr.tryAltServices(req)
		assert.Error(t, err)
		tr.altSvcCache.Delete(getOrigin(req.URL))
	})
}

func TestTransportUsesAltService(t *testing.T) {
	// A working alt-service endpoint should be used and returned.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	defer srv.Close()

	altURL := mustParseURL(t, srv.URL)
	tr := NewTransport(nil)
	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	// Cache an alt-svc pointing at the same host:port via ":port" form.
	tr.altSvcCache.Store(getOrigin(req.URL), []*altSvc{
		{protocol: "h2", endpoint: ":" + altURL.Port(), expires: time.Now().Add(time.Hour)},
	})

	resp, err := tr.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestTryServiceEndpointValidation(t *testing.T) {
	tr := NewTransport(nil)
	req, _ := http.NewRequest(http.MethodGet, "http://example.com:80/path", nil)

	t.Run("invalid endpoint format", func(t *testing.T) {
		_, err := tr.tryService(req, &altSvc{endpoint: "noportcolon-but-not-prefixed"})
		assert.Error(t, err)
	})

	t.Run("origin mismatch", func(t *testing.T) {
		_, err := tr.tryService(req, &altSvc{endpoint: "other.com:443"})
		assert.Error(t, err)
	})
}

func TestGetOrigin(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{"http://example.com", "example.com:80"},
		{"https://example.com", "example.com:443"},
		{"http://example.com:8080", "example.com:8080"},
		{"https://example.com:9443", "example.com:9443"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, getOrigin(mustParseURL(t, tt.raw)))
	}
}

func TestDefaultPort(t *testing.T) {
	assert.Equal(t, "443", defaultPort("https"))
	assert.Equal(t, "80", defaultPort("http"))
	assert.Equal(t, "80", defaultPort("ws"))
}
