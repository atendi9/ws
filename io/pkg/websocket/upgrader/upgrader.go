package upgrader

import (
	"net/http"

	"github.com/atendi9/ws/io/pkg/websocket/wsconn"
)

// BufferSize defines the read and write buffer sizes for the WebSocket connection.
type BufferSize struct {
	// Reader is the size of the read buffer.
	Reader int
	// Writer is the size of the write buffer.
	Writer int
}

// Definition holds the configuration settings used to initialize a new WebSocket upgrader.
type Definition struct {
	// Size specifies the read and write buffer sizes.
	Size BufferSize
	// EnableCompression determines whether compression should be enabled for the connection.
	EnableCompression bool
	// ErrorHandler is an optional function to handle upgrade errors.
	ErrorHandler func(w http.ResponseWriter, r *http.Request, status int, reason error)
}

// New creates and returns a configured pointer to [wsconn.Upgrader] based on the provided [Definition].
func New(def Definition) *wsconn.Upgrader {
	return &wsconn.Upgrader{
		ReadBufferSize:    def.Size.Reader,
		WriteBufferSize:   def.Size.Writer,
		EnableCompression: def.EnableCompression,
		Error:             def.ErrorHandler,
	}
}

// Is checks whether the incoming HTTP request is a WebSocket upgrade request.
func Is(r *http.Request) bool {
	return wsconn.IsWebSocketUpgrade(r)
}
