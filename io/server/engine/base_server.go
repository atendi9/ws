package engine

import (
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/atendi9/ws/io/pkg/anvil"
	"github.com/atendi9/ws/io/pkg/chrono"
	"github.com/atendi9/ws/io/pkg/errors"
	"github.com/atendi9/ws/io/pkg/etch"
	"github.com/atendi9/ws/io/pkg/events"
	"github.com/atendi9/ws/io/pkg/xhttp"
	"github.com/atendi9/ws/io/server/engine/config"
	"github.com/atendi9/ws/io/server/engine/transports"
)

// Protocol errors mappings.
var (
	// UnknownTransport indicates that the requested transport is not supported by the server. It uses [xhttp.CodeMessage].
	UnknownTransport = &xhttp.CodeMessage{Code: 0, Message: `Transport unknown`}
	// UnknownSID indicates that the provided session ID was not found. It uses [xhttp.CodeMessage].
	UnknownSID = &xhttp.CodeMessage{Code: 1, Message: `Session ID unknown`}
	// BadHandshakeMethod indicates that the HTTP method used for the handshake is invalid. It uses [xhttp.CodeMessage].
	BadHandshakeMethod = &xhttp.CodeMessage{Code: 2, Message: `Bad handshake method`}
	// BadRequest indicates a malformed or invalid request. It uses [xhttp.CodeMessage].
	BadRequest = &xhttp.CodeMessage{Code: 3, Message: `Bad request`}
	// Forbidden indicates that the request is not allowed by the server configuration. It uses [xhttp.CodeMessage].
	Forbidden = &xhttp.CodeMessage{Code: 4, Message: `Forbidden`}
	// UnsupportedProtocolVersion indicates that the requested protocol version is not supported. It uses [xhttp.CodeMessage].
	UnsupportedProtocolVersion = &xhttp.CodeMessage{Code: 4, Message: `Unsupported protocol version`}
)

// Middleware defines a function signature for server middlewares that intercept [xhttp.Context].
type Middleware func(*xhttp.Context, func(error))

// BaseServer defines the interface for the core engine server operations.
// It embeds [events.Emitter] to handle event dispatching.
type BaseServer interface {
	events.Emitter
	Prototype(BaseServer)
	Proto() BaseServer
	Opts() config.ServerOptions
	Clients() *anvil.Map[string, Socket]
	ClientsCount() uint64
	Middlewares() []Middleware
	Transports() *anvil.Set[string]
	TransportsByName() map[string]transports.TConstructor
	Construct(any)
	Init()
	ComputePath(config.AttachOptions) string
	Upgrades(string) []string
	Verify(*xhttp.Context, bool) (*xhttp.CodeMessage, map[string]any)
	Use(Middleware)
	ApplyMiddlewares(*xhttp.Context, func(error))
	Close() BaseServer
	Cleanup()
	GenerateId(*xhttp.Context) string
	Handshake(*xhttp.Context, string) (*xhttp.CodeMessage, transports.Transport)
	CreateTransport(*xhttp.Context, string) (transports.Transport, error)
}

// baseServer represents the underlying implementation of the [BaseServer] interface.
// It embeds [events.Emitter] to facilitate event-driven architecture.
type baseServer struct {
	events.Emitter

	// Prototype interface, used to implement interface method rewriting pointing to [BaseServer].
	_proto_ BaseServer

	opts config.ServerOptions

	transports        *anvil.Set[string]      // Available transport types
	_transportsByName map[string]TConstructor // Transport constructors by name

	clients      *anvil.Map[string, Socket]
	clientsCount atomic.Uint64
	middlewares  []Middleware
	middlewareMu sync.RWMutex
}

// MakeBaseServer creates and returns a new instance of [BaseServer].
func MakeBaseServer() BaseServer {
	baseServer := &baseServer{
		Emitter: events.NewEmitter(),
		clients: &anvil.Map[string, Socket]{},
	}

	baseServer.Prototype(baseServer)

	return baseServer
}

// Prototype sets the prototype [BaseServer] for interface method rewriting.
func (bs *baseServer) Prototype(server BaseServer) {
	bs._proto_ = server
}

// Proto returns the current prototype [BaseServer].
func (bs *baseServer) Proto() BaseServer {
	return bs._proto_
}

// Opts returns the currently configured [config.ServerOptions].
func (bs *baseServer) Opts() config.ServerOptions {
	return bs.opts
}

// Clients returns a thread-safe map containing the connected [Socket] clients. It returns a [anvil.Map].
func (bs *baseServer) Clients() *anvil.Map[string, Socket] {
	return bs.clients
}

// ClientsCount returns the atomic count of currently connected clients.
func (bs *baseServer) ClientsCount() uint64 {
	return bs.clientsCount.Load()
}

// Middlewares returns the list of registered [Middleware] functions.
func (bs *baseServer) Middlewares() []Middleware {
	return bs.middlewares
}

// Transports returns a [anvil.Set] of available transport names.
func (bs *baseServer) Transports() *anvil.Set[string] {
	return bs.transports
}

// TransportsByName returns a map linking transport names to their constructors via [transports.TConstructor].
func (bs *baseServer) TransportsByName() map[string]transports.TConstructor {
	return bs._transportsByName
}

// Construct initializes the [baseServer] with the provided options.
// The option parameter should implement [config.ServerOptions].
func (bs *baseServer) Construct(opt any) {
	opts, _ := opt.(config.ServerOptions)

	options := config.DefaultServerOptions()
	options.SetPingTimeout(20_000 * time.Millisecond)
	options.SetPingInterval(25_000 * time.Millisecond)
	options.SetUpgradeTimeout(10_000 * time.Millisecond)
	options.SetMaxHttpBufferSize(1e6)
	options.SetIdleTimeout(120 * time.Second)
	options.SetTransports(anvil.NewSet(Polling, WebSocket))
	options.SetAllowUpgrades(true)
	options.SetHttpCompression(&xhttp.Compression{Threshold: 1024})
	options.SetCors(nil)
	options.SetAllowEIO3(false)

	bs.opts = options.Assign(opts)

	bs.transports = anvil.NewSet[string]()
	bs._transportsByName = map[string]TConstructor{}
	if transports := bs.opts.Transports(); transports != nil {
		for _, transport := range transports.Keys() {
			transportName := transport.Name()
			bs.transports.Add(transportName)
			bs._transportsByName[transportName] = transport
		}
	}

	if opts != nil {
		if cookie := opts.Cookie(); cookie != nil {
			if len(cookie.Name) == 0 {
				cookie.Name = "io"
			}
			if len(cookie.Path) > 0 {
				cookie.HttpOnly = true
			}
			if len(cookie.Path) == 0 {
				cookie.Path = "/"
			}
			if cookie.SameSite == http.SameSiteDefaultMode {
				cookie.SameSite = http.SameSiteLaxMode
			}
			bs.opts.SetCookie(cookie)
		}
	}

	if cors := bs.opts.Cors(); cors != nil {
		bs.Use(xhttp.MiddlewareWrapper(cors))
	}

	bs._proto_.Init()
}

// Init handles further initialization logic. It is meant to be overridden by the embedding struct.
func (bs *baseServer) Init() {}

// ComputePath computes and normalizes the pathname of the requests that are handled by the server using [config.AttachOptions].
func (bs *baseServer) ComputePath(options config.AttachOptions) string {
	path := "/engine.io"

	if options != nil {
		if options.GetRawPath() != nil {
			path = strings.TrimRight(options.Path(), "/")
		}
		if options.GetRawAddTrailingSlash() == nil || options.AddTrailingSlash() {
			// normalize path
			path += "/"
		}
	}

	return path
}

// Upgrades returns a list of available transports for upgrade given a specific transport name.
func (bs *baseServer) Upgrades(transport string) []string {
	if !bs.opts.AllowUpgrades() {
		return nil
	}
	ctor, ok := bs._transportsByName[transport]
	if !ok {
		return nil
	}
	return ctor.UpgradesTo()
}

// Verify checks the validity of an incoming request based on its [xhttp.Context] and whether it is an upgrade request.
// It returns an [xhttp.CodeMessage] with a corresponding payload on failure, or nil on success.
func (bs *baseServer) Verify(ctx *xhttp.Context, upgrade bool) (*xhttp.CodeMessage, map[string]any) {
	// transport check
	transport := ctx.Query().Peek("transport")
	if !bs.transports.Has(transport) {

		return UnknownTransport, map[string]any{"transport": transport}
	}

	// 'Origin' header check
	if origin := ctx.Headers().Peek("Origin"); etch.CheckInvalidHeaderChar(origin) {
		ctx.Headers().Remove("Origin")

		return BadRequest, map[string]any{"name": "INVALID_ORIGIN", "origin": origin}
	}

	// sid check
	sessionId := ctx.Query().Peek("sid")
	if len(sessionId) > 0 {
		// Validate SID format to prevent abuse (e.g. excessively long values)
		if !chrono.IsValid(sessionId) {

			return BadRequest, map[string]any{"name": "INVALID_SID", "sid": sessionId}
		}
		socket, ok := bs.clients.Load(sessionId)
		if !ok {

			return UnknownSID, map[string]any{"sid": sessionId}
		}
		if previousTransport := socket.Transport().Name(); !upgrade && previousTransport != transport {

			return BadRequest, map[string]any{"name": "TRANSPORT_MISMATCH", "transport": transport, "previousTransport": previousTransport}
		}
	} else {
		// handshake is GET only
		if method := ctx.Method(); method != http.MethodGet {
			return BadHandshakeMethod, map[string]any{"method": method}
		}

		if transport == transports.WEBSOCKET && !upgrade {

			return BadRequest, map[string]any{"name": "TRANSPORT_HANDSHAKE_ERROR"}
		}

		if allowRequest := bs.opts.AllowRequest(); allowRequest != nil {
			if err := allowRequest(ctx); err != nil {
				return Forbidden, map[string]any{"message": err.Error()}
			}
		}
	}

	return nil, nil
}

// Use adds a new [Middleware] function to the request handling chain.
func (bs *baseServer) Use(fn Middleware) {
	bs.middlewareMu.Lock()
	defer bs.middlewareMu.Unlock()
	bs.middlewares = append(bs.middlewares, fn)
}

// ApplyMiddlewares applies all registered [Middleware] functions sequentially to the [xhttp.Context].
// The callback is invoked with an error if one occurs, or nil upon successful completion.
func (bs *baseServer) ApplyMiddlewares(ctx *xhttp.Context, callback func(error)) {
	bs.middlewareMu.RLock()
	middlewares := make([]Middleware, len(bs.middlewares))
	copy(middlewares, bs.middlewares)
	bs.middlewareMu.RUnlock()

	if len(middlewares) == 0 {

		callback(nil)
		return
	}
	var apply func(int)
	apply = func(i int) {

		middlewares[i](ctx, func(err error) {
			if err != nil {
				callback(err)
				return
			}
			if i+1 < len(middlewares) {
				apply(i + 1)
			} else {
				callback(nil)
			}
		})
	}

	apply(0)
}

// Close gracefully shuts down the [baseServer], forcing all active clients to close their connections.
func (bs *baseServer) Close() BaseServer {

	bs.clients.Range(func(_ string, client Socket) bool {
		client.Close(true)
		return true
	})

	bs._proto_.Cleanup()

	return bs
}

// Cleanup performs background teardown operations for the [baseServer]. Meant to be overridden.
func (bs *baseServer) Cleanup() {}

// GenerateId generates a unique socket ID for an incoming connection based on its [xhttp.Context].
// Overwrite this method to generate your custom socket id.
func (bs *baseServer) GenerateId(*xhttp.Context) string {
	return chrono.New().Generate()
}

// Handshake performs the initial connection handshake process using the [xhttp.Context] and transport name.
// It sets up the [transports.Transport] and initializes the client [Socket]. It returns an [xhttp.CodeMessage] if an error occurs.
func (bs *baseServer) Handshake(ctx *xhttp.Context, transportName string) (*xhttp.CodeMessage, transports.Transport) {
	protocol := 3 // 3rd revision by default
	if ctx.Query().Peek("EIO") == "4" {
		protocol = 4
	}

	if protocol == 3 && !bs.opts.AllowEIO3() {

		bs.Emit("connection_error", &xhttp.ErrorMessage{
			CodeMessage: UnsupportedProtocolVersion,
			Req:         ctx,
			Context: map[string]any{
				"protocol": protocol,
			},
		})
		return UnsupportedProtocolVersion, nil
	}

	id := bs.GenerateId(ctx)

	ctx.IdleTimeout = bs.opts.IdleTimeout()
	transport, err := bs._proto_.CreateTransport(ctx, transportName)
	if err != nil {

		bs.Emit("connection_error", &xhttp.ErrorMessage{
			CodeMessage: BadRequest,
			Req:         ctx,
			Context: map[string]any{
				"name":  "TRANSPORT_HANDSHAKE_ERROR",
				"error": err,
			},
		})
		return BadRequest, nil
	}

	if transports.POLLING == transportName {
		transport.SetMaxHttpBufferSize(bs.opts.MaxHttpBufferSize())
		transport.SetHttpCompression(bs.opts.HttpCompression())
	} else if transports.WEBSOCKET == transportName {
		transport.SetPerMessageDeflate(bs.opts.PerMessageDeflate())
	}

	_ = transport.On("headers", func(args ...any) {
		headers, req := anvil.TryGetAny[*xhttp.ParameterBag](args, 0), anvil.TryGetAny[*xhttp.Context](args, 1)
		if !ctx.Query().Has("sid") {
			if cookie := bs.opts.Cookie(); cookie != nil {
				headers.Set("Set-Cookie", cookie.String())
			}
			bs.Emit("initial_headers", headers, req)
		}
		bs.Emit("headers", headers, req)
	})

	transport.OnRequest(ctx)

	socket := NewSocket(ctx, id, bs, transport, protocol)

	bs.clients.Store(id, socket)
	bs.clientsCount.Add(1)

	_ = socket.Once("close", func(...any) {
		bs.clients.Delete(id)
		bs.clientsCount.Add(^uint64(0))
	})

	bs.Emit("connection", socket)

	return nil, transport
}

// CreateTransport constructs a new [transports.Transport] based on the provided [xhttp.Context] and transport name.
// This is a stub implementation that defaults to returning an unimplemented transport error.
func (*baseServer) CreateTransport(*xhttp.Context, string) (transports.Transport, error) {
	return nil, errors.ErrTransportNotImplemented
}
