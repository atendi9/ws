// Package transports provides the definitions and registry for Socket.IO transport mechanisms.
package transports

import "github.com/atendi9/ws/io/pkg/xhttp"

const (
	// POLLING represents the HTTP long-polling transport mechanism.
	POLLING string = "polling"
	// WEBSOCKET represents the WebSocket transport mechanism.
	WEBSOCKET string = "websocket"
)

// TConstructor defines the interface for building and managing different transport types.
type TConstructor interface {
	// HandlesUpgrades indicates whether this transport mechanism supports upgrading to another.
	HandlesUpgrades() bool

	// Name returns the identifier of the transport.
	Name() string

	// New creates a new [Transport] instance using the provided [xhttp.Context].
	New(*xhttp.Context) Transport

	// UpgradesTo returns a list of transport names that this transport can be upgraded to.
	UpgradesTo() []string
}

// transports holds the registry of available transport builders.
var transports map[string]TConstructor

// init initializes the default transport mechanisms.
func init() {
	transports = map[string]TConstructor{
		POLLING:   &PollingBuilder{},
		WEBSOCKET: &WebSocketBuilder{},
	}
}

// Transports returns the map of all registered [TConstructor] instances.
func Transports() map[string]TConstructor {
	return transports
}
