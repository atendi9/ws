package websocket

import (
	"github.com/atendi9/ws/io/pkg/events"
	"github.com/atendi9/ws/io/pkg/websocket/wsconn"
)

// Conn wraps a low-level [*wsconn.Conn] and embeds an [events.Emitter]
// to provide event-driven behavior on top of the WebSocket connection.
type Conn struct {
	events.Emitter
	*wsconn.Conn
}

// Close closes the underlying network connection ([*wsconn.Conn]) and triggers
// the "close" event via the embedded [events.Emitter].
func (t *Conn) Close() error {
	defer t.Emit("close")
	return t.Conn.Close()
}
