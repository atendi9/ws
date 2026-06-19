package transports

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/atendi9/ws/io/parsers/engine/packet"
	"github.com/atendi9/ws/io/pkg/anvil"
	"github.com/atendi9/ws/io/pkg/forge"
	"github.com/atendi9/ws/io/pkg/queue"
	ws "github.com/atendi9/ws/io/pkg/websocket"
	"github.com/atendi9/ws/io/pkg/xhttp"
)

// Websocket is an alias for the [Transport] interface, representing a WebSocket transport connection.
type Websocket = Transport

// websocket is the internal implementation of the [Websocket] transport.
// It embeds the base [Transport] and manages the underlying [ws.Conn].
type websocket struct {
	Transport

	idleTimeout time.Duration

	socket     *ws.Conn
	mu         sync.Mutex
	writeQueue *queue.Queue
}

// MakeWebSocket creates a new, uninitialized [Websocket] instance.
func MakeWebSocket() Websocket {
	w := &websocket{Transport: MakeTransport()}

	w.Prototype(w)

	return w
}

// NewWebSocket creates and constructs a new [Websocket] using the provided [xhttp.Context].
func NewWebSocket(ctx *xhttp.Context) Websocket {
	w := MakeWebSocket()

	w.Construct(ctx)

	return w
}

// Construct initializes the [websocket] transport using the provided [xhttp.Context].
// It sets up the internal write queue, assigns the socket, and binds error and close event listeners.
func (w *websocket) Construct(ctx *xhttp.Context) {
	w.Transport.Construct(ctx)

	w.idleTimeout = ctx.IdleTimeout
	w.socket = ctx.Websocket
	w.writeQueue = queue.New()

	_ = w.socket.Emitter.On("error", func(errs ...any) {
		w.OnError("websocket error", anvil.TryGetAny[error](errs, 0))
	})
	_ = w.socket.Emitter.Once("close", func(...any) {
		w.OnClose()
	})

	go w.message()

	w.SetWritable(true)
	w.SetPerMessageDeflate(nil)
}

// Name returns the identifier name of the [websocket] transport.
func (w *websocket) Name() string {
	return WEBSOCKET
}

// HandlesUpgrades returns true, as the [websocket] transport supports protocol upgrades.
func (w *websocket) HandlesUpgrades() bool {
	return true
}

// _error handles underlying connection errors and emits the appropriate close or error events to the [ws.Conn].
func (w *websocket) _error(err error) {
	if ws.IsUnexpectedCloseError(err) || errors.Is(err, net.ErrClosed) {
		w.socket.Emit("close")
	} else {
		w.socket.Emit("error", err)
	}
}

// message acts as a continuous read loop, processing incoming messages from the underlying [ws.Conn].
func (w *websocket) message() {
	defer func() {
		if !w.writeQueue.IsShuttingDown() {
			w.socket.Emit("close")
		}
	}()

	for {
		if w.idleTimeout > 0 {
			_ = w.socket.SetReadDeadline(time.Now().Add(w.idleTimeout))
		}
		mt, message, err := w.socket.NextReader()
		if err != nil {
			w._error(err)
			return
		}

		switch mt {
		case ws.BinaryMessage:
			read := forge.NewBytesBuffer(nil)
			if _, err := read.ReadFrom(message); err != nil {
				w._error(err)
			} else {
				w.onMessage(read)
			}
		case ws.TextMessage:
			read := forge.NewString(nil)
			if _, err := read.ReadFrom(message); err != nil {
				w._error(err)
			} else {
				w.onMessage(read)
			}
		case ws.CloseMessage:
			w.socket.Emit("close")
			if c, ok := message.(io.Closer); ok {
				_ = c.Close()
			}
			return
		case ws.PingMessage:
		case ws.PongMessage:
		}
		if c, ok := message.(io.Closer); ok {
			_ = c.Close()
		}
	}
}

// onMessage delegates the incoming [forge.Interface] payload to the underlying [Transport] OnData method.
func (w *websocket) onMessage(data forge.Interface) {
	w.OnData(data)
}

// Send queues a slice of [packet.Packet] instances to be transmitted over the [websocket].
func (w *websocket) Send(packets []*packet.Packet) {
	w.SetWritable(false)
	w.writeQueue.Enqueue(func() { w.send(packets) })
}

// send processes, encodes, and writes the queued [packet.Packet] instances to the underlying socket.
func (w *websocket) send(packets []*packet.Packet) {
	defer func() {
		w.Emit("drain")
		w.SetWritable(true)
		w.Emit("ready")
	}()

	w.mu.Lock()
	defer w.mu.Unlock()

	for _, packet := range packets {
		compress := true
		if packet.Options != nil {
			if packet.Options.Compress != nil && !*packet.Options.Compress {
				compress = false
			}

			if w.PerMessageDeflate() == nil && packet.Options.WsPreEncodedFrame != nil {
				mt := ws.BinaryMessage
				if _, ok := packet.Options.WsPreEncodedFrame.(*forge.String); ok {
					mt = ws.TextMessage
				}
				pm, err := ws.NewPreparedMessage(mt, packet.Options.WsPreEncodedFrame.Bytes())
				if err != nil {

					w._error(err)
					return
				}
				if err := w.socket.WritePreparedMessage(pm); err != nil {

					w._error(err)
					return
				}
				continue

			}
		}

		data, err := w.Parser().EncodePacket(packet, w.SupportsBinary())
		if err != nil {

			w._error(err)
			return
		}
		w.write(data, compress)
	}
}

// write handles the compression and transmission of the encoded [forge.Interface] payload through the [ws.Conn].
func (w *websocket) write(data forge.Interface, compress bool) {
	if w.PerMessageDeflate() != nil {
		if data.Len() < w.PerMessageDeflate().Threshold {
			compress = false
		}
	}

	w.socket.EnableWriteCompression(compress)
	mt := ws.BinaryMessage
	if _, ok := data.(*forge.String); ok {
		mt = ws.TextMessage
	}
	write, err := w.socket.NextWriter(mt)
	if err != nil {
		w._error(err)
		return
	}
	defer func() {
		if err := write.Close(); err != nil {
			w._error(err)
			return
		}
	}()
	if _, err := io.Copy(write, data); err != nil {
		w._error(err)
		return
	}
}

// OnClose handles the closure of the [websocket], shutting down the internal write queue and triggering the base [Transport] OnClose.
func (w *websocket) OnClose() {
	w.writeQueue.TryClose()
	w.Transport.OnClose()
}

// DoClose executes the underlying connection shutdown sequence and invokes the provided [xhttp.Callable] if present.
func (w *websocket) DoClose(fn xhttp.Callable) {
	w.writeQueue.TryClose()
	defer func() { _ = w.socket.Close() }()
	if fn != nil {
		fn()
	}
}
