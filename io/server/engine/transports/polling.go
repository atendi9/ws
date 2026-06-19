package transports

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/atendi9/ws/io/parsers/engine/packet"
	"github.com/atendi9/ws/io/pkg/anvil"
	"github.com/atendi9/ws/io/pkg/chrono"
	"github.com/atendi9/ws/io/pkg/etch"
	"github.com/atendi9/ws/io/pkg/events"
	"github.com/atendi9/ws/io/pkg/forge"
	"github.com/atendi9/ws/io/pkg/queue"
	"github.com/atendi9/ws/io/pkg/xhttp"
)

const (
	// DefaultPollingCloseTimeout is the default time to wait for pending writes before closing a polling transport.
	DefaultPollingCloseTimeout = 30000 * time.Millisecond
)

// Polling defines the interface for the HTTP long-polling transport.
// It extends the base [Transport] interface.
type Polling interface {
	Transport

	// DoWrite executes the actual HTTP response write using the provided [xhttp.Context],
	// [forge.Interface] data, and [packet.Options].
	DoWrite(*xhttp.Context, forge.Interface, *packet.Options, func(error))
}

// polling implements the [Polling] interface to handle HTTP long-polling connections.
type polling struct {
	Transport

	closeTimeout time.Duration

	req     atomic.Pointer[xhttp.Context]
	dataCtx atomic.Pointer[xhttp.Context]

	shouldClose atomic.Pointer[xhttp.Callable]
	mu          sync.Mutex
	writeQueue  *queue.Queue
}

// MakePolling creates and returns a new uninitialized [Polling]  instance.
func MakePolling() Polling {
	p := &polling{Transport: MakeTransport()}

	p.Prototype(p)

	return p
}

// NewPolling creates and initializes a new [Polling]  instance with the given  [xhttp.Context].
func NewPolling(ctx *xhttp.Context) Polling {
	p := MakePolling()

	p.Construct(ctx)

	return p
}

// Construct initializes the [polling] instance with the given  [xhttp.Context] and sets up the [queue.Queue].
func (p *polling) Construct(ctx *xhttp.Context) {
	p.Transport.Construct(ctx)

	p.closeTimeout = DefaultPollingCloseTimeout
	p.writeQueue = queue.New()
}

// Name returns the transport name, which is always POLLING.
func (p *polling) Name() string {
	return POLLING
}

// OnRequest handles incoming HTTP requests from the given [xhttp.Context] and delegates
// to the appropriate handler based on the HTTP method.
func (p *polling) OnRequest(ctx *xhttp.Context) {
	method := ctx.Method()

	switch method {
	case http.MethodGet:
		p.onPollRequest(ctx)
	case http.MethodPost:
		p.onDataRequest(ctx)
	default:
		_ = ctx.SetStatusCode(http.StatusInternalServerError)
		_, _ = ctx.Write(nil)
	}
}

// onPollRequest handles HTTP GET requests for the given [xhttp.Context], which are used by the client to poll for data.
func (p *polling) onPollRequest(ctx *xhttp.Context) {
	if p.req.Load() != nil {
		p.OnError("overlap from client", nil)
		_ = ctx.SetStatusCode(http.StatusBadRequest)
		_, _ = ctx.Write(nil)
		return
	}

	p.req.Store(ctx)

	onClose := events.Listener(func(...any) {
		p.SetWritable(false)
		p.OnError("poll connection closed prematurely", nil)
	})

	ctx.Cleanup = func() {
		ctx.RemoveListener("close", onClose)
		p.req.Store(nil)
	}

	_ = ctx.Once("close", onClose)

	p.SetWritable(true)
	p.Emit("ready")

	if p.Writable() && p.shouldClose.Load() != nil {
		p.Send([]*packet.Packet{
			{
				Type: packet.NOOP,
			},
		})
	}
}

// onDataRequest handles HTTP POST requests for the given [xhttp.Context], which are used to receive data from the client.
func (p *polling) onDataRequest(ctx *xhttp.Context) {
	if p.dataCtx.Load() != nil {
		p.OnError("data request overlap from client", nil)
		_ = ctx.SetStatusCode(http.StatusBadRequest)
		_, _ = ctx.Write(nil)
		return
	}

	isBinary := ctx.Headers().Peek("Content-Type") == "application/octet-stream"

	if isBinary && p.Protocol() == 4 {
		p.OnError("invalid content", nil)
		_ = ctx.SetStatusCode(http.StatusBadRequest)
		_, _ = ctx.Write(nil)
		return
	}

	p.dataCtx.Store(ctx)

	var cleanup xhttp.Callable

	onClose := func(...any) {
		if cleanup != nil {
			cleanup()
		}
		p.OnError("data request connection closed prematurely", nil)
	}

	cleanup = func() {
		ctx.RemoveListener("close", onClose)
		p.dataCtx.Store(nil)
	}

	_ = ctx.Once("close", onClose)

	if ctx.Request().ContentLength > p.MaxHttpBufferSize() {
		cleanup()

		_ = ctx.SetStatusCode(http.StatusRequestEntityTooLarge)
		_, _ = ctx.Write(nil)
		return
	}

	var packetData forge.Interface
	if isBinary {
		packetData = forge.NewBytesBuffer(nil)
	} else {
		packetData = forge.NewString(nil)
	}
	if body := ctx.Request().Body; body != nil {
		_, _ = packetData.ReadFrom(io.LimitReader(body, p.MaxHttpBufferSize()))
		_ = body.Close()
	}
	p.Proto().OnData(packetData)

	cleanup()

	headers := xhttp.NewParameterBag(map[string][]string{
		"Content-Type":   {"text/html"},
		"Content-Length": {"2"},
	})

	ctx.ResponseHeaders().With(p.headers(ctx, headers).All())
	_ = ctx.SetStatusCode(http.StatusOK)
	_, _ = io.WriteString(ctx, "ok")
}

// OnData decodes the incoming payload and processes each [packet.Packet] from the given [forge.Interface].
func (p *polling) OnData(data forge.Interface) {
	packets, _ := p.Parser().DecodePayload(data)
	for _, packetData := range packets {
		if packet.CLOSE == packetData.Type {
			p.OnClose()
			return
		}

		p.OnPacket(packetData)
	}
}

// OnClose handles the transport closing process and flushes pending writes to the [queue.Queue].
func (p *polling) OnClose() {
	if p.Writable() {
		p.Send([]*packet.Packet{
			{
				Type: packet.NOOP,
			},
		})
	}
	p.writeQueue.TryClose()
	p.Transport.OnClose()
}

// Send enqueues a slice of  [packet.Packet] to be sent to the client via the internal [queue.Queue].
func (p *polling) Send(packets []*packet.Packet) {
	p.SetWritable(false)
	p.writeQueue.Enqueue(func() { p.send(packets) })
}

// send processes and encodes the slice of  [packet.Packet], then writes them to the underlying connection.
func (p *polling) send(packets []*packet.Packet) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if shouldClose := p.shouldClose.Load(); shouldClose != nil {
		packets = append(packets, &packet.Packet{
			Type: packet.CLOSE,
		})
		(*shouldClose)()
		p.shouldClose.Store(nil)
	}

	compress := false
	for _, packetData := range packets {
		if packetData.Options != nil && packetData.Options.Compress != nil && *packetData.Options.Compress {
			compress = true
			break
		}
	}
	option := &packet.Options{Compress: anvil.New(compress)}
	if p.Protocol() == 3 {
		data, _ := p.Parser().EncodePayload(packets, p.SupportsBinary())
		p.write(data, option)
	} else {
		data, _ := p.Parser().EncodePayload(packets)
		p.write(data, option)
	}
}

// write attempts to write the encoded [forge.Interface] payload to the current active request with the specified [packet.Options].
func (p *polling) write(data forge.Interface, options *packet.Options) {
	ctx := p.req.Load()
	if ctx == nil {
		p.OnError("polling write error", nil)
		return
	}
	p.Proto().(Polling).DoWrite(ctx, data, options, func(err error) {
		if err != nil {
			p.OnError("polling write error", err)
			return
		}
		p.Emit("drain")
	})
}

// DoWrite executes the actual HTTP response write, handling content encoding and compression for the given [xhttp.Context] and [forge.Interface].
func (p *polling) DoWrite(ctx *xhttp.Context, data forge.Interface, options *packet.Options, callback func(error)) {
	contentType := "application/octet-stream"
	switch data.(type) {
	case *forge.String:
		contentType = "text/plain; charset=UTF-8"
	}

	headers := xhttp.NewParameterBag(map[string][]string{
		"Content-Type": {contentType},
	})

	respond := func(data forge.Interface, length string) {
		ctx.Cleanup()
		defer callback(nil)

		headers.Set("Content-Length", length)
		ctx.ResponseHeaders().With(p.headers(ctx, headers).All())
		_ = ctx.SetStatusCode(http.StatusOK)
		_, _ = io.Copy(ctx, data)
	}

	if p.HttpCompression() == nil || (options != nil && options.Compress != nil && !*options.Compress) {
		respond(data, strconv.Itoa(data.Len()))
		return
	}

	if data.Len() < p.HttpCompression().Threshold {
		respond(data, strconv.Itoa(data.Len()))
		return
	}

	encoding := etch.Contains(ctx.Headers().Peek("Accept-Encoding"), []string{"gzip", "deflate"})
	if encoding == "" {
		respond(data, strconv.Itoa(data.Len()))
		return
	}

	buf, err := p.compress(data, encoding)
	if err != nil {
		ctx.Cleanup()
		defer callback(err)

		_ = ctx.SetStatusCode(http.StatusInternalServerError)
		_, _ = ctx.Write(nil)
		return
	}

	headers.Set("Content-Encoding", encoding)
	respond(buf, strconv.Itoa(buf.Len()))
}

// compress applies the specified compression encoding to the [forge.Interface] returning the compressed [forge.Interface].
func (p *polling) compress(data forge.Interface, encoding string) (forge.Interface, error) {
	return p.Transport.HttpCompression().Compress(encoding, data)
}

// DoClose initiates the transport shutdown process and optionally executes a [xhttp.Callable].
func (p *polling) DoClose(fn xhttp.Callable) {
	p.writeQueue.TryClose()

	if dataCtx := p.dataCtx.Load(); dataCtx != nil && !dataCtx.IsDone() {
		dataCtx.ResponseHeaders().Set("Connection", "close")
		_ = dataCtx.SetStatusCode(http.StatusTooManyRequests)
		_, _ = dataCtx.Write(nil)
	}

	onClose := func() {
		if fn != nil {
			fn()
		}
		p.OnClose()
	}

	if p.Writable() {
		p.Send([]*packet.Packet{
			{
				Type: packet.CLOSE,
			},
		})
		onClose()
	} else if p.Discarded() {

		onClose()
	} else {

		closeTimeoutTimer := chrono.SetTimeout(onClose, p.closeTimeout)
		shouldClose := func() {
			chrono.ClearTimeout(closeTimeoutTimer)
			onClose()
		}
		p.shouldClose.Store(&shouldClose)
	}
}

// headers prepares and emits the HTTP headers using a [xhttp.ParameterBag] for the polling response in the given [xhttp.Context].
func (p *polling) headers(ctx *xhttp.Context, headers *xhttp.ParameterBag) *xhttp.ParameterBag {
	if ua := ctx.UserAgent(); (len(ua) > 0) && (strings.Contains(ua, ";MSIE") || strings.Contains(ua, "Trident/")) {
		headers.Set("X-XSS-Protection", "0")
	}
	headers.Set("Cache-Control", "no-store")
	p.Emit("headers", headers, ctx)
	return headers
}
