package xrequest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
)

// HTTPClient represents the HTTP client wrapper around the standard net/http client.
type HTTPClient struct {
	client  *http.Client
	options *clientOptions
	isDone  atomic.Bool
}

// NewHTTPClient instantiates a new HTTP client based strictly on standard library implementations.
func NewHTTPClient(options ...ClientOption) *HTTPClient {
	opts := applyOptions(options...)

	transport := opts.Transport
	if transport == nil {
		defaultTransport := http.DefaultTransport.(*http.Transport).Clone()
		if opts.TLSClientConfig != nil {
			defaultTransport.TLSClientConfig = opts.TLSClientConfig
		}
		if opts.Proxy != "" {
			if proxyURL, err := url.Parse(opts.Proxy); err == nil {
				defaultTransport.Proxy = http.ProxyURL(proxyURL)
			}
		}
		transport = defaultTransport
	}

	client := &http.Client{
		Timeout: opts.Timeout,
		Jar:     opts.Jar,
	}

	client.Transport = transport

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if !opts.FollowRedirects {
			return http.ErrUseLastResponse
		}
		if len(via) >= opts.MaxRedirects {
			return fmt.Errorf("maximum number of redirects (%d) followed", opts.MaxRedirects)
		}
		return nil
	}

	return &HTTPClient{
		client:  client,
		options: opts,
	}
}

// Request triggers an HTTP request based on method, target URL, and contextual options.
func (c *HTTPClient) Request(ctx context.Context, method, targetURL string, options *Options) (*Response, error) {
	if options == nil {
		options = &Options{}
	}

	fullURL := targetURL
	if c.options.BaseURL != "" && !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		fullURL = strings.TrimSuffix(c.options.BaseURL, "/") + "/" + strings.TrimPrefix(targetURL, "/")
	}

	if len(options.Query) > 0 {
		u, err := url.Parse(fullURL)
		if err != nil {
			return nil, err
		}
		q := u.Query()
		for k, v := range options.Query {
			for _, val := range v {
				q.Add(k, val)
			}
		}
		u.RawQuery = q.Encode()
		fullURL = u.String()
	}

	bodyReader, contentType, err := c.buildRequestBody(options)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "engine.io-go/1.0")
	req.Header.Set("Accept", "*/*")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	if len(options.Headers) > 0 {
		for name, values := range options.Headers {
			for _, value := range values {
				if err := checkInvalidHeaderChar(value); err != nil {
					return nil, fmt.Errorf("invalid character in header %q value", name)
				}
				req.Header.Add(name, value)
			}
		}
	}

	if options.BasicAuth != nil && options.BasicAuth.Username != "" {
		req.SetBasicAuth(options.BasicAuth.Username, options.BasicAuth.Password)
	}

	if options.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+options.BearerToken)
	}

	for _, cookie := range options.Cookies {
		req.AddCookie(cookie)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	return &Response{resp}, nil
}

func (c *HTTPClient) buildRequestBody(options *Options) (io.Reader, string, error) {
	if options.JSON != nil {
		data, err := json.Marshal(options.JSON)
		if err != nil {
			return nil, "", err
		}
		return bytes.NewReader(data), "application/json", nil
	}

	if options.Form != nil {
		formValues := url.Values{}
		for k, v := range options.Form {
			formValues.Set(k, v)
		}
		return strings.NewReader(formValues.Encode()), "application/x-www-form-urlencoded", nil
	}

	if options.Multipart != nil {
		bodyBuf := &bytes.Buffer{}
		mw := multipart.NewWriter(bodyBuf)
		for k, v := range options.Multipart {
			if v == nil {
				return nil, "", fmt.Errorf("multipart field %q has nil value", k)
			}
			if v.Reader == nil {
				return nil, "", fmt.Errorf("multipart field %q has nil Reader", k)
			}

			var pw io.Writer
			var err error
			if v.FileName != "" {
				pw, err = mw.CreateFormFile(k, v.FileName)
			} else {
				pw, err = mw.CreateFormField(k)
			}
			if err != nil {
				return nil, "", err
			}
			if _, err := io.Copy(pw, v.Reader); err != nil {
				return nil, "", err
			}
		}
		if err := mw.Close(); err != nil {
			return nil, "", err
		}
		return bodyBuf, mw.FormDataContentType(), nil
	}

	if options.Body != nil {
		switch v := options.Body.(type) {
		case string:
			return strings.NewReader(v), "", nil
		case []byte:
			return bytes.NewReader(v), "", nil
		case io.Reader:
			return v, "", nil
		default:
			return nil, "", fmt.Errorf("unsupported body type: %T", options.Body)
		}
	}

	return nil, "", nil
}

// Get sends a GET request.
func (c *HTTPClient) Get(url string, options *Options) (*Response, error) {
	return c.Request(context.Background(), http.MethodGet, url, options)
}

// Post sends a POST request.
func (c *HTTPClient) Post(url string, options *Options) (*Response, error) {
	return c.Request(context.Background(), http.MethodPost, url, options)
}

// Put sends a PUT request.
func (c *HTTPClient) Put(url string, options *Options) (*Response, error) {
	return c.Request(context.Background(), http.MethodPut, url, options)
}

// Delete sends a DELETE request.
func (c *HTTPClient) Delete(url string, options *Options) (*Response, error) {
	return c.Request(context.Background(), http.MethodDelete, url, options)
}

// Patch sends a PATCH request.
func (c *HTTPClient) Patch(url string, options *Options) (*Response, error) {
	return c.Request(context.Background(), http.MethodPatch, url, options)
}

// Head sends a HEAD request.
func (c *HTTPClient) Head(url string, options *Options) (*Response, error) {
	return c.Request(context.Background(), http.MethodHead, url, options)
}

// Options sends an OPTIONS request.
func (c *HTTPClient) Options(url string, options *Options) (*Response, error) {
	return c.Request(context.Background(), http.MethodOptions, url, options)
}

// Close closes idle connections on the underlying transport.
func (c *HTTPClient) Close() error {
	if c.isDone.CompareAndSwap(false, true) {
		if transport, ok := c.client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
	return nil
}

func checkInvalidHeaderChar(v string) error {
	for i := 0; i < len(v); i++ {
		b := v[i]
		if b == '\r' || b == '\n' || b == 0 {
			return fmt.Errorf("invalid character in header")
		}
	}
	return nil
}
