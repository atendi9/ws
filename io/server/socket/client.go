package socket

import (
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/atendi9/ws/io/parsers/socket/parser"
	"github.com/atendi9/ws/io/pkg/anvil"
	"github.com/atendi9/ws/io/pkg/chrono"
	"github.com/atendi9/ws/io/pkg/forge"
	"github.com/atendi9/ws/io/pkg/xhttp"
	"github.com/atendi9/ws/io/server/engine"
)

// Client represents a Socket.IO client connection.
type Client struct {
	conn engine.Socket

	id             string
	server         *Server
	encoder        parser.Encoder
	decoder        parser.Decoder
	sockets        *anvil.Map[SocketId, *Socket]
	nsps           *anvil.Map[string, *Socket]
	connectTimeout atomic.Pointer[chrono.Timer]

	mu sync.Mutex
}

// MakeClient creates a new Client instance.
func MakeClient() *Client {
	c := &Client{
		sockets: &anvil.Map[SocketId, *Socket]{},
		nsps:    &anvil.Map[string, *Socket]{},
	}

	return c
}

// NewClient creates a new Client and initializes it with the given server and connection.
func NewClient(server *Server, conn engine.Socket) *Client {
	c := MakeClient()

	c.Construct(server, conn)

	return c
}

// Conn returns the underlying Engine.IO socket connection.
func (c *Client) Conn() engine.Socket {
	return c.conn
}

// Construct initializes the client with the given server and connection.
func (c *Client) Construct(server *Server, conn engine.Socket) {
	c.server = server
	c.conn = conn
	c.encoder = server.Encoder()
	c.decoder = server._parser.NewDecoder()
	c.id = conn.Id()
	c.setup()
}

// Request returns the reference to the request that originated the Engine.IO connection.
func (c *Client) Request() *xhttp.Context {
	return c.conn.Request()
}

// setup sets up event listeners for the client.
func (c *Client) setup() {
	_ = c.decoder.On("decoded", c.ondecoded)
	_ = c.conn.On("data", c.ondata)
	_ = c.conn.On("error", c.onerror)
	_ = c.conn.Once("close", c.onclose)

	c.connectTimeout.Store(chrono.SetTimeout(func() {
		if c.nsps.Len() == 0 {
			c.close()
		}
	}, c.server._connectTimeout))
}

// connect connects a client to a namespace with optional auth parameters.
func (c *Client) connect(name string, auth map[string]any) {
	if _, ok := c.server._nsps.Load(name); ok {

		c.doConnect(name, auth)
		return
	}
	c.server._checkNamespace(name, auth, func(dynamicNspName Namespace) {
		if dynamicNspName != nil {
			c.doConnect(name, auth)
		} else {

			c._packet(&parser.Packet{
				Type: parser.CONNECT_ERROR,
				Nsp:  name,
				Data: map[string]any{
					"message": "Invalid namespace",
				},
			}, nil)
		}
	})
}

// doConnect connects a client to a namespace and adds the socket to the client.
func (c *Client) doConnect(name string, auth map[string]any) {
	nsp := c.server.Of(name, nil)
	nsp.Add(c, auth, func(socket *Socket) {
		c.sockets.Store(socket.Id(), socket)
		c.nsps.Store(nsp.Name(), socket)
		if connectTimeout := c.connectTimeout.Load(); connectTimeout != nil {
			chrono.ClearTimeout(connectTimeout)
			c.connectTimeout.Store(nil)
		}
	})
}

// _disconnect disconnects from all namespaces and closes the transport.
func (c *Client) _disconnect() {
	c.sockets.Range(func(id SocketId, socket *Socket) bool {
		socket.Disconnect(false)
		return true
	})
	c.sockets.Clear()
	c.close()
}

// _remove removes a socket from the client. Called by each Socket.
func (c *Client) _remove(socket *Socket) {
	if nsp, ok := c.sockets.Load(socket.Id()); ok {
		c.sockets.Delete(socket.Id())
		c.nsps.Delete(nsp.Nsp().Name())
	} else {

	}
}

// close closes the underlying connection.
func (c *Client) close() {
	if c.conn.ReadyState() == "open" {

		c.conn.Close(false)
		c.onclose("forced server close")
	}
}

// _packet writes a packet to the transport.
func (c *Client) _packet(packet *parser.Packet, opts *WriteOptions) {
	if c.conn.ReadyState() != "open" {

		return
	}

	if opts == nil {
		opts = &WriteOptions{}
	}

	c.WriteToEngine(c.encoder.Encode(packet), opts)
}

// WriteToEngine writes encoded packets to the Engine.IO transport.
func (c *Client) WriteToEngine(encodedPackets []forge.Interface, opts *WriteOptions) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if opts.Volatile && !c.conn.Transport().Writable() {

		return
	}

	for _, encodedPacket := range encodedPackets {
		c.conn.Write(encodedPacket.Clone(), &opts.Options, nil)
	}
}

// ondata is called with incoming transport data.
func (c *Client) ondata(args ...any) {
	// error is needed for protocol violations (GH-1880)
	if err := c.decoder.Add(args[0]); err != nil {

		c.onerror(err)
	}
}

// ondecoded is called when the parser fully decodes a packet.
func (c *Client) ondecoded(args ...any) {
	packet, _ := args[0].(*parser.Packet)
	var namespace string
	var authPayload map[string]any
	if c.conn.Protocol() == 3 {
		if parsed, err := url.Parse(packet.Nsp); err == nil {
			namespace = parsed.Path
			authPayload = anvil.Values(parsed.Query(), func(value []string) any {
				return value
			})
		}
	} else {
		namespace = packet.Nsp
		authPayload, _ = packet.Data.(map[string]any)
	}
	socket, ok := c.nsps.Load(namespace)
	if !ok && packet.Type == parser.CONNECT {
		c.connect(namespace, authPayload)
	} else if ok && packet.Type != parser.CONNECT && packet.Type != parser.CONNECT_ERROR {
		socket.Enqueue(func() { socket._onpacket(packet) })
	} else {

		c.close()
	}
}

// onerror handles an error from the transport or parser.
func (c *Client) onerror(args ...any) {
	c.sockets.Range(func(_ SocketId, socket *Socket) bool {
		socket._onerror(args[0])
		return true
	})
	c.conn.Close(false)
}

// onclose is called upon transport close.
func (c *Client) onclose(args ...any) {

	// ignore a potential subsequent `close` event
	c.destroy()

	// `nsps` and `sockets` are cleaned up seamlessly
	c.sockets.Range(func(id SocketId, socket *Socket) bool {
		socket._onclose(args...)
		return true
	})
	c.sockets.Clear()

	c.decoder.Destroy() // clean up decoder
}

// destroy cleans up event listeners and timers for the client.
func (c *Client) destroy() {
	c.conn.RemoveListener("data", c.ondata)
	c.conn.RemoveListener("error", c.onerror)
	c.conn.RemoveListener("close", c.onclose)
	c.decoder.RemoveListener("decoded", c.ondecoded)

	if connectTimeout := c.connectTimeout.Load(); connectTimeout != nil {
		chrono.ClearTimeout(connectTimeout)
		c.connectTimeout.Store(nil)
	}
}
