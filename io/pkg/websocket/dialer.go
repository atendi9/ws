package websocket

import (
	"crypto/tls"
	"net/http"
	"net/url"

	"github.com/atendi9/ws/io/pkg/websocket/wsconn"
)

type Dialer = wsconn.Dialer

type DialerProxy = func(*http.Request) (*url.URL, error)

type DialerConfig struct {
	TLS               *tls.Config
	Subprotocols      []string
	EnableCompression bool
	Jar               http.CookieJar
}

func NewDialer(
	proxy DialerProxy,
	config DialerConfig,
) *Dialer {
	return &Dialer{
		Proxy:             proxy,
		TLSClientConfig:   config.TLS,
		Subprotocols:      config.Subprotocols,
		EnableCompression: config.EnableCompression,
		Jar:               config.Jar,
	}
}
