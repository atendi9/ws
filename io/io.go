// Package io provides utility structures and functions to set up and manage
// socket.io servers and connections.
package io

import (
	"context"
	"net/http"

	"github.com/atendi9/ws/io/pkg/xhttp"
	"github.com/atendi9/ws/io/server/socket"
	"github.com/atendi9/ws/io/pkg/events"
)

// Socket represents a named socket event and its corresponding handler factory.
// The Handler function receives a [socket.Socket] client and returns the actual
// function to be executed when the event occurs.
type Socket struct {
	Name    string
	Handler func(client *SocketClient) func(args ...any)
}

// Server wraps a [types.HttpServer] to provide socket.io initialization capabilities.
type Server struct {
	*xhttp.Server
}

// [socket.Server] alias for easier reference in the context of this package.
type IO = socket.Server

// [socket.Socket] alias for easier reference in the context of this package.
type SocketClient = socket.Socket

// IO initializes and returns a new [socket.Server]. It configures the underlying
// [types.HttpServer] with the provided [http.Handler], binds it to the given addr,
// and sets up CORS using allowedOrigins.
func (s *Server) IO(
	httpHandlers http.Handler,
	addr string,
	allowedOrigins []string,
) *IO {
	config := socket.DefaultServerOptions()
	// Credentials may only be enabled for an explicit origin allowlist:
	// the browser rejects "*" + credentials, and combining them invites
	// Cross-Site WebSocket Hijacking.
	var (
		origin           any
		allowCredentials bool
	)
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		origin = "*"
	} else {
		allowList := make([]any, len(allowedOrigins))
		for i, o := range allowedOrigins {
			allowList[i] = o
		}
		origin = allowList
		allowCredentials = true
	}
	config.SetCors(&xhttp.Cors{
		Origin:      origin,
		Credentials: allowCredentials,
	})
	config.SetMaxHttpBufferSize(MaxHTTPBufferSize.Value())
	s.Server = xhttp.NewServer(httpHandlers)
	io := socket.NewServer(s.Server, config)

	s.Server.Listen(addr, nil)
	return io
}

// [types.EventListener] alias for easier reference in the context of this package.
type EventListener = events.Listener

// ConnectionHandler defines the interface for socket connection management,
// allowing the binding of events and gracefully closing the connection.
type ConnectionHandler interface {
	On(event string, handlers ...EventListener)
	Close(func(error))
}

// NewConnectionHandler creates and returns a new instance of [ConnectionHandler]
// using the provided [*IO] instance.
func NewConnectionHandler(io *IO) ConnectionHandler {
	return &conectionHandlerAdapter{io: io}
}

// conectionHandlerAdapter implements the [ConnectionHandler] interface
// by wrapping an internal [*IO] instance.
type conectionHandlerAdapter struct {
	io *IO
}

// On registers one or more [EventListener] handlers for a specific event on the underlying [*IO] instance.
func (s *conectionHandlerAdapter) On(event string, handlers ...EventListener) {
	s.io.On(event, handlers...)
}

// Close gracefully closes the underlying [*IO] connection and triggers the provided callback.
func (s *conectionHandlerAdapter) Close(callback func(error)) {
	s.io.Close(callback)
}

// SocketCloser manages the application lifecycle by listening for shutdown signals
// and executing termination steps.
type SocketCloser struct {
	// Notifier returns a context that is canceled when a shutdown signal is received,
	// along with its cancellation function.
	Notifier func() (context.Context, context.CancelFunc)
	// Exit terminates the application.
	Exit func(int)
}

// InitSockets listens for incoming connections on the provided [ConnectionHandler]
// and binds the given slice of [Socket] definitions to each new client.
// It also blocks execution and gracefully shuts down the server upon receiving
// standard termination OS signals.
func InitSockets(
	io ConnectionHandler,
	sockets []Socket,
	closer SocketCloser,
) {
	io.On("connection", connectionListener(sockets))

	ctx, stop := closer.Notifier()
	defer stop()
	<-ctx.Done()
	io.Close(nil)
	closer.Exit(0)
}

// connectionListener builds the "connection" callback that binds every
// [Socket] handler to a newly connected client. It is separated from
// InitSockets so the binding logic can be unit-tested without registering
// an OS signal handler.
func connectionListener(sockets []Socket) func(clients ...any) {
	return func(clients ...any) {
		if len(clients) == 0 {
			return
		}
		client, ok := clients[0].(*socket.Socket)
		if !ok {
			return
		}
		for _, s := range sockets {
			client.On(s.Name, s.Handler(client))
		}
	}
}
