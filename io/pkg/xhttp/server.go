package xhttp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/atendi9/ws/io/pkg/anvil"
	"github.com/atendi9/ws/io/pkg/errors"
	"github.com/atendi9/ws/io/pkg/events"
)

// Mux defines the interface for an HTTP request multiplexer.
type Mux interface {
	http.Handler
	HandleFunc(path string, handler http.HandlerFunc)
}

// Server manages a standard [net/http.Server] instance, incorporating [events.Emitter]
// capabilities and tracking active server instances using [anvil.Slice].
type Server struct {
	events.Emitter
	Mux

	// activeServers holds the registered internal server instances.
	activeServers *anvil.Slice[any]
}

// ServeHTTP dispatches the incoming request to the underlying [Mux].
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Mux.ServeHTTP(w, r)
}

// NewServer creates and initializes a new [Server] instance with a default handler.
func NewServer(defaultHandler http.Handler) *Server {
	s := &Server{
		Emitter:       events.NewEmitter(),
		Mux:           NewServeMux(defaultHandler),
		activeServers: anvil.NewSlice[any](),
	}
	return s
}

// createServer instantiates and registers a new standard [net/http.Server].
func (s *Server) createServer(addr string, handler http.Handler) *http.Server {
	server := &http.Server{Addr: addr, Handler: handler}

	s.activeServers.Push(server)

	return server
}

// Close gracefully shuts down or closes all underlying registered servers.
// It emits a "close" event and executes the provided 'fn' callback if provided,
// passing any error encountered during the shutdown process.
func (s *Server) Close(fn func(error)) (err error) {
	s.Emit("close")

	var closingErr, serverErr error
	s.activeServers.Range(func(server any, _ int) bool {
		switch srv := server.(type) {
		case *http.Server:
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			serverErr = srv.Shutdown(shutdownCtx)
			cancel()
		default:
			serverErr = errors.ErrUnknownServerType
		}
		if serverErr != nil && closingErr == nil {
			closingErr = serverErr
		}
		return true
	})

	if closingErr != nil {
		err = fmt.Errorf("error occurred while closing servers: %v", closingErr)
	}

	if fn != nil {
		defer fn(err)
	}

	return err
}

// Listen starts a standard HTTP server on the provided address and returns the
// [net/http.Server] instance. It executes the provided 'fn' callback once the server
// is listening and emits a "listening" event.
func (s *Server) Listen(addr string, fn Callable) *http.Server {
	server := s.createServer(addr, s)

	// Start the listener asynchronously
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	if fn != nil {
		defer fn()
	}
	s.Emit("listening")

	return server
}

// ListenTLS starts a standard HTTPS server with TLS configuration on the provided
// address and returns the [net/http.Server] instance. It executes the provided 'fn'
// callback once the server is listening and emits a "listening" event.
func (s *Server) ListenTLS(addr string, certFile string, keyFile string, fn Callable) *http.Server {
	server := s.createServer(addr, s)

	// Start the TLS listener asynchronously
	go func() {
		if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	if fn != nil {
		defer fn()
	}
	s.Emit("listening")

	return server
}
