// Package engine provides the core functionality for managing socket.io engine servers.
package engine

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/atendi9/ws/io/pkg/anvil"
	"github.com/atendi9/ws/io/pkg/errors"
	"github.com/atendi9/ws/io/pkg/events"
	"github.com/atendi9/ws/io/pkg/websocket"
	"github.com/atendi9/ws/io/pkg/websocket/upgrader"
	"github.com/atendi9/ws/io/pkg/xhttp"
	"github.com/atendi9/ws/io/server/engine/config"
	"github.com/atendi9/ws/io/server/engine/transports"
)

const (
	// DefaultBufferSize is the default buffer size used for operations.
	DefaultBufferSize = 1024

	// DefaultWSReadBufferSize is the default read buffer size for WebSocket connections.
	DefaultWSReadBufferSize = DefaultBufferSize

	// DefaultWSWriteBufferSize is the default write buffer size for WebSocket connections.
	DefaultWSWriteBufferSize = DefaultBufferSize
)

// Server defines the interface for the engine server.
// It embeds [BaseServer] and [http.Handler].
type Server interface {
	BaseServer
	http.Handler
	SetHttpServer(*xhttp.Server)
	HttpServer() *xhttp.Server
	CreateTransport(*xhttp.Context, string) (transports.Transport, error)
	HandleRequest(*xhttp.Context)
	HandleUpgrade(*xhttp.Context)
	Attach(*xhttp.Server, any)
}

// server implements the [Server] interface.
type server struct {
	BaseServer

	httpServer *xhttp.Server
}

// MakeServer creates and returns a new instance of [Server] with default base settings.
func MakeServer() Server {
	s := &server{BaseServer: MakeBaseServer()}

	s.Prototype(s)

	return s
}

// NewServer creates a new [Server] configured with the provided options.
func NewServer(opt any) Server {
	s := MakeServer()

	s.Construct(opt)

	return s
}

// SetHttpServer sets the underlying [xhttp.Server] for the engine server.
func (s *server) SetHttpServer(httpServer *xhttp.Server) {
	s.httpServer = httpServer
}

// HttpServer returns the current [xhttp.Server] associated with the engine server.
func (s *server) HttpServer() *xhttp.Server {
	return s.httpServer
}

// Init initializes the server.
func (s *server) Init() {}

// Cleanup performs necessary teardown operations for the server.
func (s *server) Cleanup() {}

// CreateTransport instantiates a new [transports.Transport] based on the given transportName.
// It uses the provided [xhttp.Context] for initialization.
func (s *server) CreateTransport(ctx *xhttp.Context, transportName string) (transports.Transport, error) {
	if transport, ok := s.TransportsByName()[transportName]; ok {
		return transport.New(ctx), nil
	}
	return nil, errors.ErrUnsupportedTransport
}

// HandleRequest processes standard HTTP long-polling requests using the given [xhttp.Context].
func (s *server) HandleRequest(ctx *xhttp.Context) {
	callback := func(codeMessage *xhttp.CodeMessage, errorContext map[string]any) {
		if codeMessage != nil {
			s.emitAbortRequest(ctx, codeMessage, errorContext)
			return
		}

		if sid := ctx.Query().Peek("sid"); sid != "" {

			if socket, ok := s.Clients().Load(sid); ok {
				socket.Transport().OnRequest(ctx)
			} else {
				abortRequest(ctx, UnknownSID, map[string]any{"sid": sid})
			}
		} else {
			if codeMessage, t := s.Handshake(ctx, ctx.Query().Peek("transport")); t == nil {
				abortRequest(ctx, codeMessage, nil)
			}
		}
	}

	s.ApplyMiddlewares(ctx, func(err error) {
		if err != nil {
			callback(BadRequest, map[string]any{"name": "MIDDLEWARE_FAILURE"})
		} else {
			callback(s.Verify(ctx, false))
		}
	})

	<-ctx.Done()
}

// HandleUpgrade handles the protocol upgrade process for WebSocket connections using the given [xhttp.Context].
func (s *server) HandleUpgrade(ctx *xhttp.Context) {
	callback := func(codeMessage *xhttp.CodeMessage, errorContext map[string]any) {
		if codeMessage != nil {
			s.emitAbortUpgrade(ctx, codeMessage, errorContext)
			return
		}

		wsc := &websocket.Conn{Emitter: events.NewEmitter()}
		bufferSize := upgrader.BufferSize{
			Reader: DefaultWSReadBufferSize,
			Writer: DefaultWSWriteBufferSize,
		}
		enableCompression := s.Opts().PerMessageDeflate() != nil
		errorHandler := func(w http.ResponseWriter, r *http.Request, status int, reason error) {
			http.Error(w, reason.Error(), status)
			wsc.Emit("error", reason)
		}
		ws := upgrader.New(
			upgrader.Definition{
				Size:              bufferSize,
				EnableCompression: enableCompression,
				ErrorHandler:      errorHandler,
			},
		)

		if cors := s.Opts().Cors(); cors != nil && cors.Origin != nil {
			ws.CheckOrigin = func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				return cors.IsOriginAllowed(origin, cors.Origin)
			}
		}

		if conn, err := ws.Upgrade(ctx.Response(), ctx.Request(), ctx.ResponseHeaders().All()); err != nil {
			s.emitAbortRequest(ctx, BadRequest, map[string]any{"name": "UPGRADE_FAILURE"})

		} else {
			conn.SetReadLimit(s.Opts().MaxHttpBufferSize())
			wsc.Conn = conn
			s.onWebSocket(ctx, wsc)
		}
	}

	s.ApplyMiddlewares(ctx, func(err error) {
		if err != nil {
			callback(BadRequest, map[string]any{"name": "MIDDLEWARE_FAILURE"})
		} else {
			callback(s.Verify(ctx, true))
		}
	})
}

// onWebSocket handles the lifecycle and setup of an upgraded [websocket.Conn].
func (s *server) onWebSocket(ctx *xhttp.Context, wsc *websocket.Conn) {
	onUpgradeError := func(...any) {}

	_ = wsc.Emitter.On("error", onUpgradeError)

	transportName := ctx.Query().Peek("transport")
	if transport, ok := s.TransportsByName()[transportName]; ok && !transport.HandlesUpgrades() {
		_ = wsc.Close()
		return
	}

	id := ctx.Query().Peek("sid")
	ctx.Websocket = wsc

	if len(id) == 0 {
		if codeMessage, t := s.Handshake(ctx, transportName); t == nil {
			abortUpgrade(ctx, codeMessage, nil)
		} else {
			wsc.Emitter.RemoveListener("error", onUpgradeError)
		}
		return
	}

	client, ok := s.Clients().Load(id)

	if !ok {
		_ = wsc.Close()
	} else if client.Upgrading() {
		_ = wsc.Close()
	} else if client.Upgraded() {
		_ = wsc.Close()
	} else {
		wsc.Emitter.RemoveListener("error", onUpgradeError)

		ctx.IdleTimeout = s.Opts().IdleTimeout()
		transport, err := s.CreateTransport(ctx, transportName)
		if err != nil {
			_ = wsc.Close()
		} else {
			transport.SetPerMessageDeflate(s.Opts().PerMessageDeflate())
			client.MaybeUpgrade(transport)
		}
	}
}

// Attach binds the engine server to an existing [xhttp.Server] using the provided options.
func (s *server) Attach(server *xhttp.Server, opts any) {
	options, _ := opts.(config.AttachOptions)
	path := s.ComputePath(options)

	_ = server.Once("close", func(...any) {
		s.Close()
	})

	_ = server.Once("listening", func(...any) {
		s.Proto().Init()
	})

	server.HandleFunc(path, s.ServeHTTP)
}

// ServeHTTP handles HTTP requests, determining whether to process them as standard requests
// or upgrade them to WebSockets, implementing the [http.Handler] interface.
func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !upgrader.Is(r) {

		s.HandleRequest(xhttp.NewContext(w, r))
	} else if s.Transports().Has(transports.WEBSOCKET) {
		s.HandleUpgrade(xhttp.NewContext(w, r))
	} else {
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}

// abortRequest closes the HTTP long-polling request with an error code and message.
// It utilizes [xhttp.Context] and [xhttp.CodeMessage].
func abortRequest(ctx *xhttp.Context, codeMessage *xhttp.CodeMessage, errorContext map[string]any) {
	statusCode := http.StatusBadRequest
	if codeMessage == Forbidden {
		statusCode = http.StatusForbidden
	}
	message := codeMessage.Message
	if errorContext != nil {
		if m, ok := errorContext["message"]; ok {
			message = anvil.TryCast[string](m)
		}
	}
	ctx.ResponseHeaders().Set("Content-Type", "application/json")
	_ = ctx.SetStatusCode(statusCode)
	if b, err := json.Marshal(xhttp.CodeMessage{Code: codeMessage.Code, Message: message}); err == nil {
		_, _ = ctx.Write(b)
		return
	}
	_, _ = io.WriteString(ctx, `{"code":400,"message":"Bad request"}`)
}

// emitAbortRequest emits a connection error event before aborting the HTTP request.
func (s *server) emitAbortRequest(ctx *xhttp.Context, codeMessage *xhttp.CodeMessage, errorContext map[string]any) {
	s.Emit("connection_error", &xhttp.ErrorMessage{
		CodeMessage: codeMessage,
		Req:         ctx,
		Context:     errorContext,
	})
	abortRequest(ctx, codeMessage, errorContext)
}

// abortUpgrade closes the WebSocket connection gracefully or terminates the HTTP response with an error.
// It utilizes [xhttp.Context] and [xhttp.CodeMessage].
func abortUpgrade(ctx *xhttp.Context, codeMessage *xhttp.CodeMessage, errorContext map[string]any) {
	_ = ctx.On("error", func(...any) {})

	message := codeMessage.Message
	if errorContext != nil {
		if m, ok := errorContext["message"]; ok {
			message = anvil.TryCast[string](m)
		}
	}

	if ctx.Websocket != nil {
		defer func() { _ = ctx.Websocket.Close() }()
		_ = ctx.Websocket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, message))
	} else {
		_ = ctx.SetStatusCode(http.StatusBadRequest)
		_, _ = io.WriteString(ctx, message)
	}
}

// emitAbortUpgrade emits a connection error event before aborting the WebSocket upgrade process.
func (s *server) emitAbortUpgrade(ctx *xhttp.Context, codeMessage *xhttp.CodeMessage, errorContext map[string]any) {
	s.Emit("connection_error", &xhttp.ErrorMessage{
		CodeMessage: codeMessage,
		Req:         ctx,
		Context:     errorContext,
	})
	abortUpgrade(ctx, codeMessage, errorContext)
}
