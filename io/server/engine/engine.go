package engine

import (
	"net/http"

	"github.com/atendi9/ws/io/pkg/xhttp"
	"github.com/atendi9/ws/io/server/engine/transports"
)

type (
	TConstructor = transports.TConstructor

	WebSocketBuilder = transports.WebSocketBuilder
	PollingBuilder   = transports.PollingBuilder
)

var (
	Polling   TConstructor = &PollingBuilder{}
	WebSocket TConstructor = &WebSocketBuilder{}
)

func New(server any, args ...any) Server {
	switch s := server.(type) {
	case *xhttp.Server:
		return Attach(s, append(args, nil)[0])
	case any:
		return NewServer(s)
	}
	return NewServer(nil)
}

// Creates an http.Server exclusively used for WS upgrades.
func Listen(addr string, options any, fn xhttp.Callable) Server {
	server := xhttp.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}))

	// create engine server
	engine := Attach(server, options)
	engine.SetHttpServer(server)

	server.Listen(addr, fn)

	return engine
}

// Captures upgrade requests for a anvil.HttpServer.
func Attach(server *xhttp.Server, options any) Server {
	engine := NewServer(options)
	engine.Attach(server, options)
	return engine
}
