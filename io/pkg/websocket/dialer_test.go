package websocket

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"testing"

	"github.com/atendi9/ws/io/pkg/websocket/wsconn"

	"github.com/atendi9/capivara/assert"
)

func TestNewDialer(t *testing.T) {
	tlsCfg := &tls.Config{InsecureSkipVerify: true}
	proxy := func(*http.Request) (*url.URL, error) { return nil, nil }
	config := DialerConfig{
		TLS:               tlsCfg,
		Subprotocols:      []string{"chat", "v2"},
		EnableCompression: true,
		Jar:               nil,
	}

	d := NewDialer(proxy, config)
	assert.NotNil(t, d)
	assert.NotNil(t, d.TLSClientConfig)
	assert.True(t, d.EnableCompression)
	assert.Equal(t, 2, len(d.Subprotocols))
	assert.NotNil(t, d.Proxy)
}

func TestNewDialerZeroConfig(t *testing.T) {
	d := NewDialer(nil, DialerConfig{})
	assert.NotNil(t, d)
	assert.False(t, d.EnableCompression)
	assert.True(t, d.TLSClientConfig == nil)
	assert.True(t, d.Proxy == nil)
}

func TestDialerTypeAlias(t *testing.T) {
	// Dialer is an alias of wsconn.Dialer; values are interchangeable.
	var d Dialer = wsconn.Dialer{}
	assert.NotNil(t, &d)
}
