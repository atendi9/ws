package transports

import (
	"context"
	"sync/atomic"

	"github.com/atendi9/box"
	"github.com/atendi9/ws/io/parsers/engine/packet"
	"github.com/atendi9/ws/io/parsers/engine/parser"
	"github.com/atendi9/ws/io/pkg/errors"
	"github.com/atendi9/ws/io/pkg/events"
	"github.com/atendi9/ws/io/pkg/forge"
	"github.com/atendi9/ws/io/pkg/state"
	"github.com/atendi9/ws/io/pkg/xhttp"
)

// Transport represents the base interface for all transport mechanisms.
// It embeds [events.Emitter] to handle event dispatching.
type Transport interface {
	events.Emitter

	// Prototype sets the prototype [Transport] used to implement interface method rewriting.
	Prototype(Transport)
	// Proto returns the underlying prototype [Transport].
	Proto() Transport

	// SetSid assigns the session ID to the [Transport].
	SetSid(string)
	// SetWritable sets the writable state of the [Transport].
	SetWritable(bool)

	// SetSupportsBinary specifies whether the [Transport] supports binary payloads.
	SetSupportsBinary(bool)
	// SetHttpCompression assigns the [xhttp.Compression] configuration.
	SetHttpCompression(*xhttp.Compression)
	// SetPerMessageDeflate assigns the [xhttp.PerMessageDeflate] configuration.
	SetPerMessageDeflate(*xhttp.PerMessageDeflate)
	// SetMaxHttpBufferSize sets the maximum allowed HTTP buffer size.
	SetMaxHttpBufferSize(int64)

	// Sid returns the session ID of the [Transport].
	Sid() string

	// Writable returns true if the [Transport] is currently writable.
	Writable() bool

	// Protocol returns the Engine.IO protocol version used by the [Transport].
	Protocol() int

	// Discarded returns true if the [Transport] has been discarded.
	Discarded() bool

	// Parser returns the [parser.Parser] instance used by the [Transport].
	Parser() parser.Parser

	// SupportsBinary returns true if the [Transport] supports binary data.
	SupportsBinary() bool

	// SetCurrentState updates the current [state.State] of the [Transport].
	SetCurrentState(state.State)
	// CurrentState returns the string representation of the current [state.State].
	CurrentState() string

	// HttpCompression returns the current [xhttp.Compression] configuration.
	HttpCompression() *xhttp.Compression
	// PerMessageDeflate returns the current [xhttp.PerMessageDeflate] configuration.
	PerMessageDeflate() *xhttp.PerMessageDeflate
	// MaxHttpBufferSize returns the current maximum HTTP buffer size.
	MaxHttpBufferSize() int64

	// HandlesUpgrades returns true if the [Transport] supports protocol upgrades.
	HandlesUpgrades() bool

	// Name returns the identifier name of the [Transport].
	Name() string

	// Construct initializes the [Transport] based on the provided [xhttp.Context].
	Construct(*xhttp.Context)

	// Discard marks the [Transport] as discarded and stops further operations.
	Discard()

	// OnRequest handles incoming requests using the provided [xhttp.Context].
	OnRequest(*xhttp.Context)

	// Close initiates the teardown sequence for the [Transport], optionally executing a [xhttp.Callable].
	Close(...xhttp.Callable)

	// OnError emits an error event with a custom message and an underlying [error].
	OnError(string, error)

	// OnPacket emits a packet event containing the provided [packet.Packet].
	OnPacket(*packet.Packet)

	// OnData processes the incoming [forge.Interface] payload and decodes it into a [packet.Packet].
	OnData(forge.Interface)

	// OnClose handles the closure of the [Transport] and updates its internal state.
	OnClose()

	// Send transmits a slice of [packet.Packet] instances over the [Transport].
	Send([]*packet.Packet)

	// DoClose executes the underlying close operation, invoking the provided [xhttp.Callable].
	DoClose(xhttp.Callable)
}

// transport is the default implementation of the [Transport] interface.
type transport struct {
	events.Emitter

	// Prototype interface, used to implement interface method rewriting
	_proto_ Transport

	maxHttpBufferSize atomic.Int64
	httpCompression   *xhttp.Compression
	perMessageDeflate *xhttp.PerMessageDeflate

	sid      string // The session ID.
	protocol int

	currentState box.Atomic[state.State]

	_discarded atomic.Bool

	parser parser.Parser

	supportsBinary bool

	_writable atomic.Bool
}

// MakeTransport creates a new base [Transport] with initialized events and an "open" state.
func MakeTransport() Transport {
	t := &transport{
		Emitter:         events.NewEmitter(),
		httpCompression: new(xhttp.Compression),
	}
	t.currentState.Store(state.NewState("transport", "open"))

	t.Prototype(t)

	return t
}

// NewTransport creates and constructs a new [Transport] using the provided [xhttp.Context].
func NewTransport(ctx *xhttp.Context) Transport {
	t := MakeTransport()

	t.Construct(ctx)

	return t
}

// Prototype sets the prototype [Transport] used to implement interface method rewriting.
func (t *transport) Prototype(_t Transport) {
	t._proto_ = _t
}

// Proto returns the underlying prototype [Transport].
func (t *transport) Proto() Transport {
	return t._proto_
}

// Sid returns the session ID of the [Transport].
func (t *transport) Sid() string {
	return t.sid
}

// SetSid assigns the session ID to the [Transport].
func (t *transport) SetSid(sid string) {
	t.sid = sid
}

// Writable returns true if the [Transport] is currently writable.
func (t *transport) Writable() bool {
	return t._writable.Load()
}

// SetWritable sets the writable state of the [Transport].
func (t *transport) SetWritable(writable bool) {
	t._writable.Store(writable)
}

// Protocol returns the Engine.IO protocol version used by the [Transport].
func (t *transport) Protocol() int {
	return t.protocol
}

// Discarded returns true if the [Transport] has been discarded.
func (t *transport) Discarded() bool {
	return t._discarded.Load()
}

// Parser returns the [parser.Parser] instance used by the [Transport].
func (t *transport) Parser() parser.Parser {
	return t.parser
}

// SupportsBinary returns true if the [Transport] supports binary data.
func (t *transport) SupportsBinary() bool {
	return t.supportsBinary
}

// SetSupportsBinary specifies whether the [Transport] supports binary payloads.
func (t *transport) SetSupportsBinary(supportsBinary bool) {
	t.supportsBinary = supportsBinary
}

// CurrentState returns the string representation of the current [state.State].
func (t *transport) CurrentState() string {
	return t.currentState.Load().String()
}

// SetCurrentState updates the current [state.State] of the [Transport].
func (t *transport) SetCurrentState(state state.State) {
	t.currentState.Store(state)
}

// HttpCompression returns the current [xhttp.Compression] configuration.
func (t *transport) HttpCompression() *xhttp.Compression {
	return t.httpCompression
}

// SetHttpCompression assigns the [xhttp.Compression] configuration.
func (t *transport) SetHttpCompression(httpCompression *xhttp.Compression) {
	t.httpCompression = httpCompression
}

// PerMessageDeflate returns the current [xhttp.PerMessageDeflate] configuration.
func (t *transport) PerMessageDeflate() *xhttp.PerMessageDeflate {
	return t.perMessageDeflate
}

// SetPerMessageDeflate assigns the [xhttp.PerMessageDeflate] configuration.
func (t *transport) SetPerMessageDeflate(perMessageDeflate *xhttp.PerMessageDeflate) {
	t.perMessageDeflate = perMessageDeflate
}

// MaxHttpBufferSize returns the current maximum HTTP buffer size.
func (t *transport) MaxHttpBufferSize() int64 {
	return t.maxHttpBufferSize.Load()
}

// SetMaxHttpBufferSize sets the maximum allowed HTTP buffer size.
func (t *transport) SetMaxHttpBufferSize(maxHttpBufferSize int64) {
	t.maxHttpBufferSize.Store(maxHttpBufferSize)
}

// Construct initializes the [Transport] based on the provided [xhttp.Context], determining the parser protocol and binary support.
func (t *transport) Construct(ctx *xhttp.Context) {
	if eio, ok := ctx.Query().Get("EIO"); ok && eio == "4" {
		t.parser = parser.NewV4()
	} else {
		t.parser = parser.NewV3()
	}

	t.protocol = t.parser.Protocol()
	t.supportsBinary = !ctx.Query().Has("b64")
}

// Discard marks the [Transport] as discarded and stops further operations.
func (t *transport) Discard() {
	t._discarded.Store(true)
}

// OnRequest handles incoming requests using the provided [xhttp.Context].
func (t *transport) OnRequest(req *xhttp.Context) {}

// Close initiates the teardown sequence for the [Transport], updating the state and delegating to the underlying [Transport] prototype.
func (t *transport) Close(fn ...xhttp.Callable) {
	if t.CurrentState() == "closed" || t.CurrentState() == "closing" {
		return
	}
	t.SetCurrentState(state.NewState("transport", "closing"))
	fn = append(fn, nil)
	t._proto_.DoClose(fn[0])
}

// OnError emits an error event with a custom message and an underlying [error], propagating a transport error.
func (t *transport) OnError(msg string, desc error) {
	if t.ListenerCount("error") > 0 {
		t.Emit("error", errors.NewTransportError(context.Background(), msg, desc))
	} else {

	}
}

// OnPacket emits a packet event containing the provided [packet.Packet].
func (t *transport) OnPacket(packet *packet.Packet) {
	t.Emit("packet", packet)
}

// OnData processes the incoming [forge.Interface] payload, decodes it using the [parser.Parser], and triggers OnPacket.
func (t *transport) OnData(data forge.Interface) {
	p, _ := t.parser.DecodePacket(data)
	t.OnPacket(p)
}

// OnClose handles the closure of the [Transport], setting the state to closed and emitting the close event.
func (t *transport) OnClose() {
	if t.CurrentState() == "closed" {
		return
	}
	t.SetCurrentState(state.TransportClosed)
	t.Emit("close")
}

// HandlesUpgrades returns true if the [Transport] supports protocol upgrades.
func (t *transport) HandlesUpgrades() bool {
	return false
}

// Name returns the identifier name of the [Transport].
func (t *transport) Name() string {
	return ""
}

// Send transmits a slice of [packet.Packet] instances over the [Transport].
func (t *transport) Send([]*packet.Packet) {}

// DoClose executes the underlying close operation, invoking the provided [xhttp.Callable].
func (t *transport) DoClose(xhttp.Callable) {}
