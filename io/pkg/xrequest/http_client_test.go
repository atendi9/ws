package xrequest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/atendi9/capivara/assert"
)

func TestNewHTTPClientDefaults(t *testing.T) {
	c := NewHTTPClient()
	assert.NotNil(t, c)
	assert.NoError(t, c.Close())
}

func TestNewHTTPClientWithProxyAndTLS(t *testing.T) {
	// Exercises the default-transport configuration branches.
	c := NewHTTPClient(
		WithProxy("http://proxy.local:8080"),
		WithTLSClientConfig(nil),
	)
	assert.NotNil(t, c)
}

func TestRequestGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "engine.io-go/1.0", r.Header.Get("User-Agent"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	}))
	defer srv.Close()

	c := NewHTTPClient()
	resp, err := c.Get(srv.URL, nil)
	assert.NoError(t, err)
	assert.True(t, resp.Ok())
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "hello", string(body))
}

func TestRequestAllVerbs(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewHTTPClient()
	cases := []struct {
		method string
		call   func(string, *Options) (*Response, error)
	}{
		{http.MethodPost, c.Post},
		{http.MethodPut, c.Put},
		{http.MethodDelete, c.Delete},
		{http.MethodPatch, c.Patch},
		{http.MethodHead, c.Head},
		{http.MethodOptions, c.Options},
	}
	for _, tc := range cases {
		t.Run(tc.method, func(t *testing.T) {
			resp, err := tc.call(srv.URL, nil)
			assert.NoError(t, err)
			assert.True(t, resp.Ok())
			assert.Equal(t, tc.method, gotMethod)
		})
	}
}

func TestRequestJSONBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		var payload map[string]string
		json.NewDecoder(r.Body).Decode(&payload)
		assert.Equal(t, "bar", payload["foo"])
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewHTTPClient()
	resp, err := c.Post(srv.URL, &Options{JSON: map[string]string{"foo": "bar"}})
	assert.NoError(t, err)
	assert.True(t, resp.Ok())
}

func TestRequestFormBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		r.ParseForm()
		assert.Equal(t, "v", r.FormValue("k"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewHTTPClient()
	resp, err := c.Post(srv.URL, &Options{Form: map[string]string{"k": "v"}})
	assert.NoError(t, err)
	assert.True(t, resp.Ok())
}

func TestRequestMultipartBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data"))
		err := r.ParseMultipartForm(1 << 20)
		assert.NoError(t, err)
		assert.Equal(t, "field-value", r.FormValue("field"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewHTTPClient()
	resp, err := c.Post(srv.URL, &Options{
		Multipart: map[string]*Multipart{
			"field": {Reader: strings.NewReader("field-value")},
			"file":  {FileName: "f.txt", Reader: strings.NewReader("contents")},
		},
	})
	assert.NoError(t, err)
	assert.True(t, resp.Ok())
}

func TestRequestMultipartNilValue(t *testing.T) {
	c := NewHTTPClient()
	_, err := c.Post("http://example.com", &Options{
		Multipart: map[string]*Multipart{"x": nil},
	})
	assert.Error(t, err)
}

func TestRequestMultipartNilReader(t *testing.T) {
	c := NewHTTPClient()
	_, err := c.Post("http://example.com", &Options{
		Multipart: map[string]*Multipart{"x": {FileName: "f"}},
	})
	assert.Error(t, err)
}

func TestRequestRawBodyVariants(t *testing.T) {
	var received string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		received = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewHTTPClient()

	t.Run("string body", func(t *testing.T) {
		_, err := c.Post(srv.URL, &Options{Body: "str-body"})
		assert.NoError(t, err)
		assert.Equal(t, "str-body", received)
	})
	t.Run("bytes body", func(t *testing.T) {
		_, err := c.Post(srv.URL, &Options{Body: []byte("byte-body")})
		assert.NoError(t, err)
		assert.Equal(t, "byte-body", received)
	})
	t.Run("reader body", func(t *testing.T) {
		_, err := c.Post(srv.URL, &Options{Body: bytes.NewReader([]byte("reader-body"))})
		assert.NoError(t, err)
		assert.Equal(t, "reader-body", received)
	})
	t.Run("unsupported body type", func(t *testing.T) {
		_, err := c.Post(srv.URL, &Options{Body: 12345})
		assert.Error(t, err)
	})
}

func TestRequestHeadersAuthCookies(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "custom", r.Header.Get("X-Custom"))
		user, pass, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "u", user)
		assert.Equal(t, "p", pass)
		ck, err := r.Cookie("sid")
		assert.NoError(t, err)
		assert.Equal(t, "123", ck.Value)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewHTTPClient()
	resp, err := c.Get(srv.URL, &Options{
		Headers:   http.Header{"X-Custom": {"custom"}},
		BasicAuth: &BasicAuth{Username: "u", Password: "p"},
		Cookies:   []*http.Cookie{{Name: "sid", Value: "123"}},
	})
	assert.NoError(t, err)
	assert.True(t, resp.Ok())
}

func TestRequestBearerToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewHTTPClient()
	_, err := c.Get(srv.URL, &Options{BearerToken: "tok"})
	assert.NoError(t, err)
}

func TestRequestInvalidHeaderChar(t *testing.T) {
	c := NewHTTPClient()
	_, err := c.Get("http://example.com", &Options{
		Headers: http.Header{"X-Bad": {"bad\nvalue"}},
	})
	assert.Error(t, err)
}

func TestRequestQueryParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "v1", r.URL.Query().Get("a"))
		assert.Equal(t, "v2", r.URL.Query().Get("b"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewHTTPClient()
	_, err := c.Get(srv.URL, &Options{Query: map[string][]string{"a": {"v1"}, "b": {"v2"}}})
	assert.NoError(t, err)
}

func TestRequestBaseURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/users", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewHTTPClient(WithBaseURL(srv.URL + "/api"))
	_, err := c.Get("/users", nil)
	assert.NoError(t, err)
}

func TestRequestQueryParseError(t *testing.T) {
	c := NewHTTPClient()
	_, err := c.Get("://bad", &Options{Query: map[string][]string{"a": {"1"}}})
	assert.Error(t, err)
}

func TestRequestConnectionError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close() // server no longer accepting

	c := NewHTTPClient()
	_, err := c.Get(url, nil)
	assert.Error(t, err)
}

func TestRequestNilOptionsContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewHTTPClient()
	resp, err := c.Request(context.Background(), http.MethodGet, srv.URL, nil)
	assert.NoError(t, err)
	assert.True(t, resp.Ok())
}

func TestFollowRedirectsDisabled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/end", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewHTTPClient(WithFollowRedirects(false, 0))
	resp, err := c.Get(srv.URL+"/start", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
}

func TestMaxRedirectsExceeded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/again", http.StatusFound)
	}))
	defer srv.Close()

	c := NewHTTPClient(WithFollowRedirects(true, 1))
	_, err := c.Get(srv.URL, nil)
	assert.Error(t, err)
}

func TestCloseIsIdempotent(t *testing.T) {
	c := NewHTTPClient()
	assert.NoError(t, c.Close())
	assert.NoError(t, c.Close()) // second call: CompareAndSwap returns false
}

func TestCloseWithCustomTransport(t *testing.T) {
	// Custom (non-*http.Transport) transport: Close should still be a no-op success.
	c := NewHTTPClient(WithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(""))}, nil
	})))
	assert.NoError(t, c.Close())
}

func TestCheckInvalidHeaderChar(t *testing.T) {
	assert.NoError(t, checkInvalidHeaderChar("normal-value"))
	assert.Error(t, checkInvalidHeaderChar("with\rcr"))
	assert.Error(t, checkInvalidHeaderChar("with\nlf"))
	assert.Error(t, checkInvalidHeaderChar("with\x00null"))
}

func TestRequestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewHTTPClient(WithTimeout(5 * time.Millisecond))
	_, err := c.Get(srv.URL, nil)
	assert.Error(t, err)
}

// roundTripFunc adapts a function to http.RoundTripper for tests.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
