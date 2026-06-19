package socket

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/atendi9/ws/io/parsers/socket/parser"
	"github.com/atendi9/ws/io/pkg/anvil"
	"github.com/atendi9/ws/io/pkg/etch"
	"github.com/atendi9/ws/io/pkg/events"
	"github.com/atendi9/ws/io/pkg/xhttp"
	"github.com/atendi9/ws/io/server/engine"
)

const (
	// DefaultConnectTimeout is the default time a client has to send its first namespace connection request.
	DefaultConnectTimeout = 45_000 * time.Millisecond

	// DefaultMaxDisconnectionDuration is the default maximum time a session can be disconnected before being discarded.
	DefaultMaxDisconnectionDuration int64 = 2 * 60 * 1000

	// DefaultSessionCleanupInterval is the default interval between two session cleanup sweeps.
	DefaultSessionCleanupInterval = 60_000 * time.Millisecond
)

var dotMapRegex = regexp.MustCompile(`\.map`)

type (
	ParentNspNameMatchFn *func(string, map[string]any, func(error, bool))

	Compress interface {
		NewWriter(w io.Writer, encoding string) (io.WriteCloser, error)
	}
	Server struct {
		*StrictEmitter

		compressWriter Compress

		// #readonly
		sockets Namespace
		// A reference to the underlying Engine.IO server.
		//
		//	clientsCount := io.Engine().ClientsCount()
		engine     engine.BaseServer
		_parser    parser.Parser
		encoder    parser.Encoder
		_nsps      *anvil.Map[string, Namespace]
		parentNsps *anvil.Map[ParentNspNameMatchFn, ParentNamespace]
		//
		// A subset of the {parentNsps} map, only containing {ParentNamespace} which are based on a regular
		// expression.
		parentNamespacesFromRegExp *anvil.Map[*regexp.Regexp, ParentNamespace]
		_adapter                   AdapterConstructor
		_serveClient               bool
		// #readonly
		opts            ServerOptions
		eio             engine.Server
		_path           string
		clientPathRegex *regexp.Regexp
		_connectTimeout time.Duration
		httpServer      *xhttp.Server
		_corsMiddleware engine.Middleware
	}
)

func MakeServer() *Server {
	s := &Server{
		_nsps:                      &anvil.Map[string, Namespace]{},
		parentNsps:                 &anvil.Map[ParentNspNameMatchFn, ParentNamespace]{},
		parentNamespacesFromRegExp: &anvil.Map[*regexp.Regexp, ParentNamespace]{},
		compressWriter:             new(xhttp.Compression{}),
	}
	return s
}

func NewServer(srv any, opts ServerOptions) *Server {
	s := MakeServer()

	s.Construct(srv, opts)

	return s
}

func (s *Server) Sockets() Namespace {
	return s.sockets
}

func (s *Server) Engine() engine.BaseServer {
	return s.engine
}

func (s *Server) Encoder() parser.Encoder {
	return s.encoder
}

func (s *Server) Construct(srv any, opts ServerOptions) {
	if opts == nil {
		opts = DefaultServerOptions()
	}

	if opts.GetRawPath() != nil {
		s.SetPath(opts.Path())
	} else {
		s.SetPath("/socket.io")
	}
	if opts.GetRawConnectTimeout() != nil {
		s.SetConnectTimeout(opts.ConnectTimeout())
	} else {
		s.SetConnectTimeout(DefaultConnectTimeout)
	}
	s.SetServeClient(opts.ServeClient())
	if _parser := opts.Parser(); _parser != nil {
		s._parser = _parser
	} else {
		s._parser = parser.NewParser()
	}
	s.encoder = s._parser.NewEncoder()
	s.opts = opts
	if adapter := opts.Adapter(); adapter != nil {
		s.SetAdapter(adapter)
	} else {
		if connectionStateRecovery := opts.ConnectionStateRecovery(); connectionStateRecovery != nil {
			if connectionStateRecovery.GetRawMaxDisconnectionDuration() == nil {
				connectionStateRecovery.SetMaxDisconnectionDuration(DefaultMaxDisconnectionDuration)
			}
			if connectionStateRecovery.GetRawSkipMiddlewares() == nil {
				connectionStateRecovery.SetSkipMiddlewares(true)
			}
			s.SetAdapter(&SessionAwareAdapterBuilder{})
		} else {
			s.SetAdapter(&AdapterBuilder{})
		}
	}
	s.sockets = s.Of("/", nil)

	s.StrictEmitter = s.sockets.Emitter()

	if srv != nil {
		s.Attach(srv, nil)
	}

	if cors := s.opts.Cors(); cors != nil {
		s._corsMiddleware = xhttp.MiddlewareWrapper(cors)
	}
}

func (s *Server) Opts() ServerOptions {
	return s.opts
}

// SetServeClient sets whether to serve the client code to browsers.
func (s *Server) SetServeClient(v bool) *Server {
	s._serveClient = v
	return s
}

// ServeClient returns whether the server is serving client code.
func (s *Server) ServeClient() bool {
	return s._serveClient
}

// _checkNamespace executes the middleware for an incoming namespace not already created on the server.
// name is the name of the incoming namespace, auth is the auth parameters, fn is the callback.
func (s *Server) _checkNamespace(name string, auth map[string]any, fn func(nsp Namespace)) {
	end := true
	s.parentNsps.Range(func(nextFn ParentNspNameMatchFn, pnsp ParentNamespace) bool {
		status := false
		(*nextFn)(name, auth, func(err error, allow bool) {
			if err != nil || !allow {
				status = true
				return
			}
			if nsp, ok := s._nsps.Load(name); ok {
				// the namespace was created in the meantime

				fn(nsp)
				end = false
				return
			}
			namespace := pnsp.CreateChild(name)

			fn(namespace)
			end = false
		})
		return status // whether to continue traversing.
	})
	if end {
		fn(nil)
	}
}

// SetPath sets the client serving path.
func (s *Server) SetPath(v string) *Server {
	s._path = strings.TrimRight(v, "/")
	s.clientPathRegex = regexp.MustCompile(`^` + regexp.QuoteMeta(s._path) + `/socket\.io(\.msgpack|\.esm)?(\.min)?\.js(\.map)?(?:\?|$)`)
	return s
}

// Path returns the current client serving path.
func (s *Server) Path() string {
	return s._path
}

// SetConnectTimeout sets the delay after which a client without namespace is closed.
func (s *Server) SetConnectTimeout(v time.Duration) *Server {
	s._connectTimeout = v
	return s
}

// ConnectTimeout returns the current connect timeout duration.
func (s *Server) ConnectTimeout() time.Duration {
	return s._connectTimeout
}

// SetAdapter sets the adapter for rooms.
func (s *Server) SetAdapter(v AdapterConstructor) *Server {
	s._adapter = v
	s._nsps.Range(func(_ string, nsp Namespace) bool {
		nsp.InitAdapter()
		return true
	})
	return s
}

func (s *Server) Adapter() AdapterConstructor {
	return s._adapter
}

// Listen attaches socket.io to a server or port.
// srv is the server or port, opts are options passed to engine.io.
func (s *Server) Listen(srv any, opts *ServerOpts) *Server {
	return s.Attach(srv, opts)
}

// Attach attaches socket.io to a server or port.
// srv is the server or port, opts are options passed to engine.io.
func (s *Server) Attach(srv any, opts *ServerOpts) *Server {
	var server *xhttp.Server
	switch address := srv.(type) {
	case int:
		_address := fmt.Sprintf(":%d", address)
		// handle a port as a int

		server = xhttp.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "404 page not found", http.StatusNotFound)
		}))
		server.Listen(_address, nil)
	case string:
		// handle a port as a string

		server = xhttp.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "404 page not found", http.StatusNotFound)
		}))
		server.Listen(address, nil)
	case *xhttp.Server:
		server = address
	default:
		panic(fmt.Errorf("trying to attach socket.io to express request handler %T, please pass a *xhttp.Server instance", address))
	}
	if opts == nil {
		opts = DefaultServerOptions()
	}

	// merge the options passed to the Socket.IO server
	opts.Assign(s.opts)
	// set engine.io path to `/socket.io`
	if opts.GetRawPath() == nil {
		opts.SetPath(s._path)
	}
	s.initEngine(server, opts)

	return s
}

// ServeHandler returns an http.Handler for the server.
func (s *Server) ServeHandler(opts *ServerOpts) http.Handler {
	// If an instance already exists, reuse it.
	if s.eio != nil {
		return s.eio
	}

	if opts == nil {
		opts = DefaultServerOptions()
	}

	// merge the options passed to the Socket.IO server
	opts.Assign(s.opts)
	// set engine.io path to `/socket.io`
	if opts.GetRawPath() == nil {
		opts.SetPath(s._path)
	}

	// initialize engine

	s.eio = engine.NewServer(opts)
	// bind to engine events
	s.Bind(s.eio)

	return s.eio
}

// initEngine initializes the engine.io server and attaches it to the HTTP server.
func (s *Server) initEngine(srv *xhttp.Server, opts ServerOptions) {
	// initialize engine

	s.eio = engine.Attach(srv, opts)

	// attach static file serving
	if s._serveClient {
		s.attachServe(srv, s.eio, opts)
	}

	// Export http server
	s.httpServer = srv

	// bind to engine events
	s.Bind(s.eio)
}

// attachServe attaches the static file serving handler.
func (s *Server) attachServe(srv *xhttp.Server, egs engine.Server, opts ServerOptions) {

	srv.HandleFunc(s._path+"/", func(w http.ResponseWriter, r *http.Request) {
		if s.clientPathRegex.MatchString(r.URL.Path) {
			if s._corsMiddleware != nil {
				s._corsMiddleware(xhttp.NewContext(w, r), func(error) {
					s.serve(w, r)
				})
			} else {
				s.serve(w, r)
			}
		} else {
			if opts.GetRawAddTrailingSlash() == nil || opts.AddTrailingSlash() {
				egs.ServeHTTP(w, r)
			} else {
				srv.ServeHTTP(w, r)
			}
		}
	})
}

// serve handles a request for serving client source and map files.
func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	filename := filepath.Base(r.URL.Path)
	isMap := dotMapRegex.MatchString(filename)
	expectedEtag := `"` + s.opts.ClientVersion() + `"`
	if s.opts.GetRawClientVersion() != nil {
		expectedEtag = `"` + s.opts.ClientVersion() + `"`
	}
	weakEtag := "W/" + expectedEtag

	if etag := r.Header.Get("If-None-Match"); etag != "" {
		if expectedEtag == etag || weakEtag == etag {

			w.WriteHeader(http.StatusNotModified)
			_, _ = w.Write(nil)
			return
		}
	}

	w.Header().Set("Cache-Control", "public, max-age=0")
	if isMap {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	} else {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	}
	w.Header().Set("ETag", expectedEtag)
	s.sendFile(filename, w, r)
}

// sendFile sends a static file to the client.
func (s Server) sendFile(filename string, w http.ResponseWriter, r *http.Request) {
	_file, err := os.Executable()
	if err != nil {

		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	// Construct the full, intended destination path
	basePath := filepath.Dir(filepath.Dir(_file))
	targetPath := filepath.Clean(filepath.Join(basePath, "client-dist", filename))

	// Verify the target path is still within the intended directory boundary
	if !strings.HasPrefix(targetPath, basePath) {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	file, err := os.Open(targetPath)
	if err != nil {

		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	defer func() { _ = file.Close() }()

	// Get file size for Content-Length in uncompressed responses
	fi, statErr := file.Stat()
	if statErr != nil {

		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	encoding := etch.Contains(r.Header.Get("Accept-Encoding"), []string{"gzip", "deflate"})
	writer, err := s.compressWriter.NewWriter(w, encoding)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	switch encoding {
	case "gzip":
		gz := writer
		defer func() { _ = gz.Close() }()
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(gz, file)
	case "deflate":
		fl := writer
		defer func() { _ = fl.Close() }()
		w.Header().Set("Content-Encoding", "deflate")
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(fl, file)
	default:
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fi.Size()))
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, file)
	}
}

// Bind binds socket.io to an engine.io instance.
// egs is the engine.io (or compatible) server.
func (s *Server) Bind(egs engine.BaseServer) *Server {
	s.engine = egs
	_ = s.engine.On("connection", s.onconnection)
	return s
}

// onconnection is called with each incoming transport connection.
func (s *Server) onconnection(conns ...any) {
	conn := anvil.TryGetAny[engine.Socket](conns, 0)

	client := NewClient(s, conn)
	if conn.Protocol() == 3 {
		client.connect("/", nil)
	}
}

// Of looks up a namespace by name or pattern and optionally registers a connection event handler.
// name can be a string, regexp, or ParentNspNameMatchFn; fn is the connection event handler.
func (s *Server) Of(name any, fn events.Listener) Namespace {
	switch n := name.(type) {
	case ParentNspNameMatchFn:
		parentNsp := NewParentNamespace(s)

		s.parentNsps.Store(n, parentNsp)

		if fn != nil {
			_ = parentNsp.On("connect", fn)
		}
		return parentNsp
	case *regexp.Regexp:
		parentNsp := NewParentNamespace(s)

		s.parentNsps.Store(ParentNspNameMatchFn(anvil.New(func(nsp string, _ map[string]any, next func(error, bool)) {
			next(nil, n.MatchString(nsp))
		})), parentNsp)
		s.parentNamespacesFromRegExp.Store(n, parentNsp)

		if fn != nil {
			_ = parentNsp.On("connect", fn)
		}
		return parentNsp
	}

	n, ok := name.(string)
	if ok {
		if len(n) > 0 {
			if n[0] != '/' {
				n = "/" + n
			}
		} else {
			n = "/"
		}
	} else {
		n = "/"
	}

	var namespace Namespace

	if nsp, ok := s._nsps.Load(n); ok {
		namespace = nsp
	} else {
		s.parentNamespacesFromRegExp.Range(func(regex *regexp.Regexp, parentNamespace ParentNamespace) bool {
			if regex.MatchString(n) {

				namespace = parentNamespace.CreateChild(n)
				return false
			}
			return true
		})

		if namespace != nil {
			return namespace
		}

		namespace = NewNamespace(s, n)
		s._nsps.Store(n, namespace)
		if n != "/" {
			s.sockets.EmitReserved("new_namespace", namespace)
		}
	}

	if fn != nil {
		_ = namespace.On("connect", fn)
	}
	return namespace
}

// Close closes the server and all client connections. If fn is provided, it is called on error or when all connections are closed.
func (s *Server) Close(fn func(error)) {
	s._nsps.Range(func(_ string, nsp Namespace) bool {
		nsp.Sockets().Range(func(_ SocketId, socket *Socket) bool {
			socket._onclose("server shutting down")
			return true
		})
		nsp.Adapter().Close()
		return true
	})

	if s.httpServer != nil {
		_ = s.httpServer.Close(fn)
		// The engine has been closed through the close event processing, and the subsequent process is exited here.
		return
	}

	if s.engine != nil {
		s.engine.Close()
	}

	if fn != nil {
		fn(nil)
	}
}

// Use registers a middleware function that is executed for every incoming Socket.
func (s *Server) Use(fn NamespaceMiddleware) *Server {
	s.sockets.Use(fn)
	return s
}

// To targets a room when emitting events. Returns a new BroadcastOperator for chaining.
func (s *Server) To(room ...Room) *BroadcastOperator {
	return s.sockets.To(room...)
}

// In targets a room when emitting events. Returns a new BroadcastOperator for chaining.
func (s *Server) In(room ...Room) *BroadcastOperator {
	return s.sockets.In(room...)
}

// Except excludes a room when emitting events. Returns a new BroadcastOperator for chaining.
func (s *Server) Except(room ...Room) *BroadcastOperator {
	return s.sockets.Except(room...)
}

// Emit broadcasts an event to all connected clients.
func (s *Server) Emit(ev string, args ...any) *Server {
	_ = s.sockets.Emit(ev, args...)
	return s
}

// Send sends a "message" event to all clients. This mimics the WebSocket.send() method.
func (s *Server) Send(args ...any) *Server {
	// This type-cast is needed because EmitEvents likely doesn't have `message` as a key.
	// if you specify the EmitEvents, the type of args will be never.
	_ = s.sockets.Emit("message", args...)
	return s
}

// Write sends a "message" event to all clients. Alias of Send.
func (s *Server) Write(args ...any) *Server {
	// This type-cast is needed because EmitEvents likely doesn't have `message` as a key.
	// if you specify the EmitEvents, the type of args will be never.
	_ = s.sockets.Emit("message", args...)
	return s
}

// ServerSideEmit sends a message to other Socket.IO servers in the cluster.
// ev is the event name, args are the arguments (may include an acknowledgement callback).
func (s *Server) ServerSideEmit(ev string, args ...any) error {
	return s.sockets.ServerSideEmit(ev, args...)
}

// ServerSideEmitWithAck sends a message and expects an acknowledgement from other Socket.IO servers in the cluster.
// Returns a function that will be fulfilled when all servers have acknowledged the event.
func (s *Server) ServerSideEmitWithAck(ev string, args ...any) func(Ack) error {
	return s.sockets.ServerSideEmitWithAck(ev, args...)
}

// Compress sets the compress flag for subsequent event emissions.
func (s *Server) Compress(compress bool) *BroadcastOperator {
	return s.sockets.Compress(compress)
}

// Volatile sets a modifier for a subsequent event emission that the event data may be lost if the client is not ready to receive messages.
func (s *Server) Volatile() *BroadcastOperator {
	return s.sockets.Volatile()
}

// Local sets a modifier for a subsequent event emission that the event data will only be broadcast to the current node.
func (s *Server) Local() *BroadcastOperator {
	return s.sockets.Local()
}

// Timeout adds a timeout for the next operation.
func (s *Server) Timeout(timeout time.Duration) *BroadcastOperator {
	return s.sockets.Timeout(timeout)
}

// FetchSockets returns a function to fetch the matching socket instances.
func (s *Server) FetchSockets() func(func([]*RemoteSocket, error)) {
	return s.sockets.FetchSockets()
}

// SocketsJoin makes the matching socket instances join the specified rooms.
func (s *Server) SocketsJoin(room ...Room) {
	s.sockets.SocketsJoin(room...)
}

// SocketsLeave makes the matching socket instances leave the specified rooms.
func (s *Server) SocketsLeave(room ...Room) {
	s.sockets.SocketsLeave(room...)
}

// DisconnectSockets makes the matching socket instances disconnect. If status is true, closes the underlying connection.
func (s *Server) DisconnectSockets(status bool) {
	s.sockets.DisconnectSockets(status)
}
