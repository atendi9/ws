package xrequest

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ClientOption defines the configuration function type for request options.
type ClientOption func(*clientOptions)

// clientOptions contains all available client configurations.
type clientOptions struct {
	Timeout         time.Duration
	FollowRedirects bool
	MaxRedirects    int
	Proxy           string
	TLSClientConfig *tls.Config
	Transport       http.RoundTripper
	BaseURL         string
	Jar             http.CookieJar
}

// WithTransport sets a custom http.RoundTripper.
func WithTransport(transport http.RoundTripper) ClientOption {
	return func(o *clientOptions) {
		o.Transport = transport
	}
}

// WithFollowRedirects configures redirect behavior.
func WithFollowRedirects(followRedirects bool, maxRedirects int) ClientOption {
	return func(o *clientOptions) {
		o.FollowRedirects = followRedirects
		o.MaxRedirects = maxRedirects
	}
}

// WithBaseURL sets the default base URL for the client.
func WithBaseURL(baseURL string) ClientOption {
	return func(o *clientOptions) {
		o.BaseURL = baseURL
	}
}

// WithTimeout sets the timeout duration for requests.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.Timeout = timeout
	}
}

// WithCookieJar sets the cookie jar implementation.
func WithCookieJar(jar http.CookieJar) ClientOption {
	return func(o *clientOptions) {
		o.Jar = jar
	}
}

// WithTLSClientConfig sets SSL/TLS client options.
func WithTLSClientConfig(config *tls.Config) ClientOption {
	return func(o *clientOptions) {
		o.TLSClientConfig = config
	}
}

// WithProxy sets the proxy URL string.
func WithProxy(proxy string) ClientOption {
	return func(o *clientOptions) {
		o.Proxy = proxy
	}
}

// applyOptions applies request options.
func applyOptions(opts ...ClientOption) *clientOptions {
	options := &clientOptions{
		FollowRedirects: true,
		MaxRedirects:    10,
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// BasicAuth holds credentials for HTTP Basic Authentication.
type BasicAuth struct {
	Username string
	Password string
}

// Multipart represents a multipart form field metadata and stream.
type Multipart struct {
	FileName    string
	ContentType string
	io.Reader
}

// Options contains parameters applied to an individual request.
type Options struct {
	Headers     http.Header
	BasicAuth   *BasicAuth
	BearerToken string
	Cookies     []*http.Cookie
	Query       url.Values
	Body        any
	JSON        any
	Form        map[string]string
	Multipart   map[string]*Multipart
}
