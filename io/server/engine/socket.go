package engine

import (
	"encoding/json"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/atendi9/box"
	"github.com/atendi9/ws/io/parsers/engine/packet"
	"github.com/atendi9/ws/io/pkg/anvil"
	"github.com/atendi9/ws/io/pkg/chrono"
	"github.com/atendi9/ws/io/pkg/errors"
	"github.com/atendi9/ws/io/pkg/events"
	"github.com/atendi9/ws/io/pkg/forge"
	"github.com/atendi9/ws/io/pkg/state"
	"github.com/atendi9/ws/io/pkg/xhttp"
	"github.com/atendi9/ws/io/server/engine/transports"
)

// SendCallback defines the function signature triggered after a packet transport operation.
type SendCallback func(transports.Transport)

// DefaultUpgradeCheckInterval is the duration between websocket upgrade status validations.
const DefaultUpgradeCheckInterval time.Duration = 100 * time.Millisecond

// Socket defines the contract for an Engine.IO socket connection.
type Socket interface {
	events.Emitter
	SetReadyState(state.State)
	Protocol() int
	Request() *xhttp.Context
	RemoteAddress() string
	Transport() transports.Transport
	Id() string
	ReadyState() string
	Upgraded() bool
	Upgrading() bool
	Construct(*xhttp.Context, string, BaseServer, transports.Transport, int)
	MaybeUpgrade(transports.Transport)
	Send(io.Reader, *packet.Options, SendCallback) Socket
	Write(io.Reader, *packet.Options, SendCallback) Socket
	Close(bool)
}

type socket struct {
	events.Emitter
	protocol          int
	request           *xhttp.Context
	remoteAddress     string
	readyState        box.Atomic[state.State]
	transport         atomic.Pointer[transports.Transport]
	id                string
	server            BaseServer
	upgrading         atomic.Bool
	upgraded          atomic.Bool
	writeBuffer       *anvil.Slice[*packet.Packet]
	pendingCallbacks  *anvil.Slice[SendCallback]
	bufferedCallbacks *anvil.Slice[[]SendCallback]
	cleanupListeners  *anvil.Slice[xhttp.Callable]
	pingTimeoutTimer  atomic.Pointer[chrono.Timer]
	pingIntervalTimer atomic.Pointer[chrono.Timer]
	mu                sync.Mutex
}

// Protocol returns the current Engine.IO protocol version.
func (s *socket) Protocol() int {
	return s.protocol
}

// Upgraded checks if the socket transport has successfully been upgraded.
func (s *socket) Upgraded() bool {
	return s.upgraded.Load()
}

// Upgrading checks if the socket is in the process of upgrading its transport.
func (s *socket) Upgrading() bool {
	return s.upgrading.Load()
}

// Id returns the unique session identifier for the socket.
func (s *socket) Id() string {
	return s.id
}

// RemoteAddress returns the network remote IP address of the connected client.
func (s *socket) RemoteAddress() string {
	return s.remoteAddress
}

// Request returns the initial HTTP context that established the socket.
func (s *socket) Request() *xhttp.Context {
	return s.request
}

// Transport returns the active underlying [transports.Transport] implementation.
func (s *socket) Transport() transports.Transport {
	if v := s.transport.Load(); v != nil {
		return *v
	}
	return nil
}

// Server returns the associated [BaseServer] instance.
func (s *socket) Server() BaseServer {
	return s.server
}

// ReadyState returns the current state connection string.
func (s *socket) ReadyState() string {
	return s.readyState.Load().String()
}

// SetReadyState updates the socket state and logs the transition.
func (s *socket) SetReadyState(state state.State) {

	s.readyState.Store(state)
}

// MakeSocket instantiates an uninitialized implementation of [Socket].
func MakeSocket() Socket {
	s := &socket{
		Emitter:           events.NewEmitter(),
		writeBuffer:       anvil.NewSlice[*packet.Packet](),
		pendingCallbacks:  anvil.NewSlice[SendCallback](),
		bufferedCallbacks: anvil.NewSlice[[]SendCallback](),
		cleanupListeners:  anvil.NewSlice[xhttp.Callable](),
	}
	s.readyState.Store(StateOpening)
	return s
}

// NewSocket creates and fully initializes a [Socket] instance.
func NewSocket(ctx *xhttp.Context, id string, server BaseServer, transport transports.Transport, protocol int) Socket {
	s := MakeSocket()
	s.Construct(ctx, id, server, transport, protocol)
	return s
}

// Construct configures internal properties and structures the socket state during initialization.
func (s *socket) Construct(ctx *xhttp.Context, id string, server BaseServer, transport transports.Transport, protocol int) {
	s.id = id
	s.server = server
	s.request = ctx
	s.protocol = protocol

	if ctx.Websocket != nil && ctx.Websocket.Conn != nil {
		s.remoteAddress = ctx.Websocket.RemoteAddr().String()
	} else {
		s.remoteAddress = ctx.Request().RemoteAddr
	}

	s.setTransport(transport)
	s.onOpen()
}

func (s *socket) onOpen() {
	s.SetReadyState(StateOpen)
	s.Transport().SetSid(s.id)

	data, err := json.Marshal(map[string]any{
		"sid":          s.id,
		"upgrades":     s.getAvailableUpgrades(),
		"pingInterval": int64(s.server.Opts().PingInterval() / time.Millisecond),
		"pingTimeout":  int64(s.server.Opts().PingTimeout() / time.Millisecond),
		"maxPayload":   s.server.Opts().MaxHttpBufferSize(),
	})
	if err != nil {

		s.OnClose("encode error")
		return
	}
	s.sendPacket(
		packet.OPEN,
		forge.NewString(data),
		nil, nil,
	)

	if i := s.server.Opts().InitialPacket(); i != nil {
		s.sendPacket(packet.MESSAGE, i, nil, nil)
	}

	s.Emit("open")

	if s.protocol == 3 {
		s.resetPingTimeout()
	} else {
		s.schedulePing()
	}
}

func (s *socket) onPacket(data *packet.Packet) {
	if data == nil {

		return
	}

	if s.ReadyState() != "open" {

		return
	}

	s.Emit("packet", data)

	switch data.Type {
	case packet.PING:
		if s.Transport().Protocol() != 3 {
			s.onError(errors.ErrInvalidHeartbeat)
			return
		}

		if timer := s.pingTimeoutTimer.Load(); timer != nil {
			timer.Refresh()
		}
		s.sendPacket(packet.PONG, nil, nil, nil)
		s.Emit("heartbeat")
	case packet.PONG:
		if s.Transport().Protocol() == 3 {
			s.onError(errors.ErrInvalidHeartbeat)
			return
		}

		chrono.ClearTimeout(s.pingTimeoutTimer.Load())
		if timer := s.pingIntervalTimer.Load(); timer != nil {
			timer.Refresh()
		}
		s.Emit("heartbeat")
	case packet.ERROR:
		s.OnClose("parse error")
	case packet.MESSAGE:
		s.Emit("data", data.Data)
		s.Emit("message", data.Data)
	}
}

func (s *socket) onError(err error) {

	s.OnClose("transport error", err)
}

func (s *socket) schedulePing() {
	s.pingIntervalTimer.Store(chrono.SetTimeout(func() {

		s.sendPacket(packet.PING, nil, nil, nil)
		s.resetPingTimeout()
	}, s.server.Opts().PingInterval()))
}

func (s *socket) resetPingTimeout() {
	chrono.ClearTimeout(s.pingTimeoutTimer.Load())
	s.pingTimeoutTimer.Store(chrono.SetTimeout(func() {
		if s.ReadyState() == "closed" {
			return
		}
		s.OnClose("ping timeout")
	}, s.resetPingTimeoutDuration()))
}

func (s *socket) resetPingTimeoutDuration() time.Duration {
	if s.protocol == 3 {
		return s.server.Opts().PingInterval() + s.server.Opts().PingTimeout()
	}
	return s.server.Opts().PingTimeout()
}

func (s *socket) setTransport(transport transports.Transport) {
	onError := func(err ...any) {
		s.onError(anvil.TryGetAny[error](err, 0))
	}
	onReady := func(...any) { s.flush() }
	onPacket := func(packets ...any) {
		s.onPacket(anvil.TryGetAny[*packet.Packet](packets, 0))
	}
	onDrain := func(...any) { s.onDrain() }
	onClose := func(...any) { s.OnClose("transport close") }

	s.transport.Store(&transport)

	_ = transport.Once("error", onError)
	_ = transport.On("ready", onReady)
	_ = transport.On("packet", onPacket)
	_ = transport.On("drain", onDrain)
	_ = transport.Once("close", onClose)

	s.cleanupListeners.Push(func() {
		transport.RemoveListener("error", onError)
		transport.RemoveListener("ready", onReady)
		transport.RemoveListener("packet", onPacket)
		transport.RemoveListener("drain", onDrain)
		transport.RemoveListener("close", onClose)
	})
}

func (s *socket) onDrain() {
	if seqFn, err := s.bufferedCallbacks.Shift(); err == nil {

		for _, fn := range seqFn {
			fn(s.Transport())
		}
	}
}

// MaybeUpgrade attempts to switch the current transport layer to an optimized one.
func (s *socket) MaybeUpgrade(transport transports.Transport) {

	s.upgrading.Store(true)

	var check, cleanup func()
	var onPacket, onError, onTransportClose, onClose events.Listener
	var upgradeTimeoutTimer, checkIntervalTimer atomic.Pointer[chrono.Timer]

	onPacket = func(datas ...any) {
		data, ok := datas[0].(*packet.Packet)
		if !ok {
			return
		}
		sb := new(strings.Builder)
		_, _ = io.Copy(sb, data.Data)
		if data.Type == packet.PING && sb.String() == "probe" {

			transport.Send([]*packet.Packet{{Type: packet.PONG, Data: strings.NewReader("probe")}})
			s.Emit("upgrading", transport)

			chrono.ClearInterval(checkIntervalTimer.Load())
			checkIntervalTimer.Store(chrono.SetInterval(check, DefaultUpgradeCheckInterval))

		} else if packet.UPGRADE == data.Type && s.ReadyState() != "closed" {

			cleanup()
			s.Transport().Discard()

			s.upgraded.Store(true)

			s.clearTransport()
			s.setTransport(transport)
			s.Emit("upgrade", transport)
			s.flush()
			if s.ReadyState() == "closing" {
				transport.Close(func() {
					s.OnClose("forced close")
				})
			}
		} else {
			cleanup()
			transport.Close()
		}
	}

	check = func() {
		if transports.POLLING == s.Transport().Name() && s.Transport().Writable() {

			s.Transport().Send([]*packet.Packet{{Type: packet.NOOP}})
		}
	}

	cleanup = func() {
		s.upgrading.Store(false)

		chrono.ClearInterval(checkIntervalTimer.Load())
		chrono.ClearTimeout(upgradeTimeoutTimer.Load())

		if transport != nil {
			transport.RemoveListener("packet", onPacket)
			transport.RemoveListener("close", onTransportClose)
			transport.RemoveListener("error", onError)
		}
		s.RemoveListener("close", onClose)
	}

	onError = func(errs ...any) {

		cleanup()
		if transport != nil {
			transport.Close()
		}
	}

	onTransportClose = func(...any) {
		onError("transport closed")
	}

	onClose = func(...any) {
		onError("socket closed")
	}

	upgradeTimeoutTimer.Store(chrono.SetTimeout(func() {

		cleanup()
		if transport != nil {
			if transport.CurrentState() == "open" {
				transport.Close()
			}
		}
	}, s.server.Opts().UpgradeTimeout()))

	_ = transport.On("packet", onPacket)
	_ = transport.Once("close", onTransportClose)
	_ = transport.Once("error", onError)

	_ = s.Once("close", onClose)
}

func (s *socket) clearTransport() {
	s.cleanupListeners.DoWrite(func(cleanups []xhttp.Callable) []xhttp.Callable {
		for _, cleanup := range cleanups {
			cleanup()
		}
		return cleanups[:0]
	})

	_ = s.Transport().On("error", func(...any) {

	})

	s.Transport().Close()

	chrono.ClearTimeout(s.pingTimeoutTimer.Load())
}

// OnClose triggers the destruction lifecycle of the socket interface context.
func (s *socket) OnClose(reason string, description ...error) {
	if s.ReadyState() != "closed" {
		description = append(description, nil)

		s.SetReadyState(StateClosed)

		chrono.ClearTimeout(s.pingIntervalTimer.Load())
		chrono.ClearTimeout(s.pingTimeoutTimer.Load())

		defer s.writeBuffer.Clear()

		s.pendingCallbacks.Clear()
		s.bufferedCallbacks.Clear()

		s.clearTransport()
		s.Emit("close", reason, description[0])
	}
}

// Send schedules an message data stream payload via a packet option modifier.
func (s *socket) Send(data io.Reader, options *packet.Options, callback SendCallback) Socket {
	s.sendPacket(packet.MESSAGE, data, options, callback)
	return s
}

// Write wraps the Send functionality ensuring compliance with traditional writer interfaces.
func (s *socket) Write(data io.Reader, options *packet.Options, callback SendCallback) Socket {
	s.sendPacket(packet.MESSAGE, data, options, callback)
	return s
}

func (s *socket) sendPacket(packetType packet.Type, data io.Reader, options *packet.Options, callback SendCallback) {
	if readystate := s.ReadyState(); readystate != "closing" && readystate != "closed" {

		opts := &packet.Options{}
		if options != nil {
			opts.WsPreEncodedFrame = options.WsPreEncodedFrame
			if options.Compress != nil {
				compress := *options.Compress
				opts.Compress = &compress
			}
		}

		if opts.Compress == nil || *opts.Compress {
			opts.Compress = anvil.New(true)
		} else {
			opts.Compress = anvil.New(false)
		}

		pkt := &packet.Packet{
			Type:    packetType,
			Data:    data,
			Options: opts,
		}

		s.Emit("packetCreate", pkt)

		s.writeBuffer.Push(pkt)

		if callback != nil {
			s.pendingCallbacks.Push(callback)
		}

		s.flush()
	}
}

func (s *socket) flush() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ReadyState() != "closed" && s.Transport().Writable() {
		if wbuf := s.writeBuffer.AllAndClear(); len(wbuf) > 0 {

			s.Emit("flush", wbuf)
			s.server.Emit("flush", s, wbuf)
			if pendingCallbacks := s.pendingCallbacks.AllAndClear(); len(pendingCallbacks) > 0 {
				s.bufferedCallbacks.Push(pendingCallbacks)
			} else {
				s.bufferedCallbacks.Push(nil)
			}
			s.Transport().Send(wbuf)
			s.Emit("drain")
			s.server.Emit("drain", s)
		}
	}
}

func (s *socket) getAvailableUpgrades() []string {
	availableUpgrades := []string{}
	for _, upg := range s.server.Upgrades(s.Transport().Name()) {
		if s.server.Transports().Has(upg) {
			availableUpgrades = append(availableUpgrades, upg)
		}
	}
	return availableUpgrades
}

// Close initiates orderly disconnection sequences or forces a rigid connection dropping behavior.
func (s *socket) Close(discard bool) {
	if discard && (s.ReadyState() == "open" || s.ReadyState() == "closing") {
		s.closeTransport(discard)
		return
	}

	if s.ReadyState() != "open" {
		return
	}

	s.SetReadyState(StateClosing)

	if length := s.writeBuffer.Len(); length > 0 {

		_ = s.Once("drain", func(...any) {

			s.closeTransport(discard)
		})
		return
	}

	s.closeTransport(discard)
}

func (s *socket) closeTransport(discard bool) {

	if discard {
		s.Transport().Discard()
	}
	s.Transport().Close(func() { s.OnClose("forced close") })
}
