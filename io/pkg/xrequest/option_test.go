package xrequest

import (
	"crypto/tls"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/atendi9/capivara/assert"
)

func TestApplyOptionsDefaults(t *testing.T) {
	opts := applyOptions()
	assert.True(t, opts.FollowRedirects)
	assert.Equal(t, 10, opts.MaxRedirects)
	assert.Equal(t, time.Duration(0), opts.Timeout)
	assert.Equal(t, "", opts.BaseURL)
}

func TestWithTransport(t *testing.T) {
	rt := http.DefaultTransport
	opts := applyOptions(WithTransport(rt))
	assert.NotNil(t, opts.Transport)
}

func TestWithFollowRedirects(t *testing.T) {
	opts := applyOptions(WithFollowRedirects(false, 3))
	assert.False(t, opts.FollowRedirects)
	assert.Equal(t, 3, opts.MaxRedirects)
}

func TestWithBaseURL(t *testing.T) {
	opts := applyOptions(WithBaseURL("http://example.com"))
	assert.Equal(t, "http://example.com", opts.BaseURL)
}

func TestWithTimeout(t *testing.T) {
	opts := applyOptions(WithTimeout(5 * time.Second))
	assert.Equal(t, 5*time.Second, opts.Timeout)
}

func TestWithCookieJar(t *testing.T) {
	jar, err := cookiejar.New(nil)
	assert.NoError(t, err)
	opts := applyOptions(WithCookieJar(jar))
	assert.NotNil(t, opts.Jar)
}

func TestWithTLSClientConfig(t *testing.T) {
	cfg := &tls.Config{InsecureSkipVerify: true}
	opts := applyOptions(WithTLSClientConfig(cfg))
	assert.NotNil(t, opts.TLSClientConfig)
}

func TestWithProxy(t *testing.T) {
	opts := applyOptions(WithProxy("http://proxy.local:8080"))
	assert.Equal(t, "http://proxy.local:8080", opts.Proxy)
}

func TestApplyOptionsCombined(t *testing.T) {
	opts := applyOptions(
		WithBaseURL("http://a"),
		WithTimeout(time.Second),
		WithProxy("http://p"),
	)
	assert.Equal(t, "http://a", opts.BaseURL)
	assert.Equal(t, time.Second, opts.Timeout)
	assert.Equal(t, "http://p", opts.Proxy)
}
