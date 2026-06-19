package transports

import (
	"github.com/atendi9/ws/io/pkg/xhttp"
)

// WebSocketBuilder represents a builder for creating new [Transport] instances
// specifically for WebSocket connections.
type WebSocketBuilder struct{}

// New creates and returns a new WebSocket [Transport] using the provided [xhttp.Context].
func (*WebSocketBuilder) New(ctx *xhttp.Context) Transport {
	return NewWebSocket(ctx)
}

// Name returns the transport identifier name, which is WEBSOCKET.
func (*WebSocketBuilder) Name() string {
	return WEBSOCKET
}

// HandlesUpgrades returns true, indicating that the WebSocket transport supports protocol upgrades.
func (*WebSocketBuilder) HandlesUpgrades() bool {
	return true
}

// UpgradesTo returns nil, as WebSocket is the final upgraded transport and does not upgrade further.
func (*WebSocketBuilder) UpgradesTo() []string {
	return nil
}

// PollingBuilder represents a builder for creating new [Transport] instances
// for HTTP long-polling or JSONP connections.
type PollingBuilder struct{}

// New creates and returns a new [Transport] using the provided [xhttp.Context].
// It returns a JSONP transport if the "j" query parameter is present, otherwise a standard polling transport.
func (*PollingBuilder) New(ctx *xhttp.Context) Transport {
	if ctx.Query().Has("j") {
		return NewJSONP(ctx)
	}
	return NewPolling(ctx)
}

// Name returns the transport identifier name, which is POLLING.
func (*PollingBuilder) Name() string {
	return POLLING
}

// HandlesUpgrades returns false, indicating that the polling transport itself does not handle protocol upgrades.
func (*PollingBuilder) HandlesUpgrades() bool {
	return false
}

// UpgradesTo returns a slice containing the WEBSOCKET string, indicating that polling can be upgraded to WebSocket.
func (*PollingBuilder) UpgradesTo() []string {
	return []string{WEBSOCKET}
}
