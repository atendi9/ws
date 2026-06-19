package config

import (
	"io"
	"net/http"
	"time"

	"github.com/atendi9/box"
	"github.com/atendi9/ws/io/pkg/anvil"
	"github.com/atendi9/ws/io/pkg/xhttp"
	"github.com/atendi9/ws/io/server/engine/transports"
)

// AllowRequest is a function type that receives an [xhttp.Context] and decides whether to continue processing.
// Returning an [error] indicates that the request was rejected.
type AllowRequest func(*xhttp.Context) error

// ServerOptions defines the interface for configuring Engine.IO server settings.
type ServerOptions interface {
	SetPingTimeout(time.Duration)
	GetRawPingTimeout() box.Optional[time.Duration]
	PingTimeout() time.Duration

	SetPingInterval(time.Duration)
	GetRawPingInterval() box.Optional[time.Duration]
	PingInterval() time.Duration

	SetUpgradeTimeout(time.Duration)
	GetRawUpgradeTimeout() box.Optional[time.Duration]
	UpgradeTimeout() time.Duration

	SetMaxHttpBufferSize(int64)
	GetRawMaxHttpBufferSize() box.Optional[int64]
	MaxHttpBufferSize() int64

	SetAllowRequest(AllowRequest)
	GetRawAllowRequest() box.Optional[AllowRequest]
	AllowRequest() AllowRequest

	SetTransports(*anvil.Set[transports.TConstructor])
	GetRawTransports() box.Optional[*anvil.Set[transports.TConstructor]]
	Transports() *anvil.Set[transports.TConstructor]

	SetAllowUpgrades(bool)
	GetRawAllowUpgrades() box.Optional[bool]
	AllowUpgrades() bool

	SetPerMessageDeflate(*xhttp.PerMessageDeflate)
	GetRawPerMessageDeflate() box.Optional[*xhttp.PerMessageDeflate]
	PerMessageDeflate() *xhttp.PerMessageDeflate

	SetHttpCompression(*xhttp.Compression)
	GetRawHttpCompression() box.Optional[*xhttp.Compression]
	HttpCompression() *xhttp.Compression

	SetInitialPacket(io.Reader)
	GetRawInitialPacket() box.Optional[io.Reader]
	InitialPacket() io.Reader

	SetCookie(*http.Cookie)
	GetRawCookie() box.Optional[*http.Cookie]
	Cookie() *http.Cookie

	SetCors(*xhttp.Cors)
	GetRawCors() box.Optional[*xhttp.Cors]
	Cors() *xhttp.Cors

	SetAllowEIO3(bool)
	GetRawAllowEIO3() box.Optional[bool]
	AllowEIO3() bool

	SetIdleTimeout(time.Duration)
	GetRawIdleTimeout() box.Optional[time.Duration]
	IdleTimeout() time.Duration
}

// ServerOpts contains the configuration fields for the server options.
type ServerOpts struct {
	// pingTimeout defines how many ms without a pong packet to consider the connection closed, stored as a [box.Optional].
	pingTimeout box.Optional[time.Duration]

	// pingInterval defines how many ms before sending a new ping packet, stored as a [box.Optional].
	pingInterval box.Optional[time.Duration]

	// upgradeTimeout defines how many ms before an uncompleted transport upgrade is canceled, stored as a [box.Optional].
	upgradeTimeout box.Optional[time.Duration]

	// maxHttpBufferSize defines how many bytes or characters a message can be, before closing the session, stored as a [box.Optional].
	maxHttpBufferSize box.Optional[int64]

	// allowRequest holds the function that decides whether a handshake or upgrade request should continue, stored as a [box.Optional].
	allowRequest box.Optional[AllowRequest]

	// transports defines the low-level transports that are enabled, stored as a [box.Optional].
	transports box.Optional[*anvil.Set[transports.TConstructor]]

	// allowUpgrades defines whether to allow transport upgrades, stored as a [box.Optional].
	allowUpgrades box.Optional[bool]

	// perMessageDeflate holds the parameters of the WebSocket permessage-deflate extension, stored as a [box.Optional].
	perMessageDeflate box.Optional[*xhttp.PerMessageDeflate]

	// httpCompression holds the parameters of the http compression for the polling transports, stored as a [box.Optional].
	httpCompression box.Optional[*xhttp.Compression]

	// TODO: Implement pluggable WebSocket engine support.
	// The default engine will be gorilla/websocket. Future engines to support include
	// coder/websocket, gobwas/ws, coder/websocket, etc.
	// Field type: box.Optional[WsEngine]
	// wsEngine box.Optional[WsEngine]

	// initialPacket holds an optional packet to be concatenated to the handshake packet, stored as a [box.Optional].
	initialPacket box.Optional[io.Reader]

	// cookie holds the configuration of the cookie that contains the client sid, stored as a [box.Optional].
	cookie box.Optional[*http.Cookie]

	// cors holds the options that will be forwarded to the cors module, stored as a [box.Optional].
	cors box.Optional[*xhttp.Cors]

	// allowEIO3 defines whether to enable compatibility with Socket.IO v2 clients, stored as a [box.Optional].
	allowEIO3 box.Optional[bool]

	// idleTimeout defines the maximum amount of seconds that may pass without sending or getting a message, stored as a [box.Optional].
	idleTimeout box.Optional[time.Duration]
}

// DefaultServerOptions returns a new [ServerOpts] with default values.
func DefaultServerOptions() *ServerOpts {
	return &ServerOpts{}
}

// Assign merges data from another [ServerOptions] into the current [ServerOpts].
func (s *ServerOpts) Assign(data ServerOptions) ServerOptions {
	if data == nil {
		return s
	}

	if data.GetRawPingTimeout() != nil {
		s.SetPingTimeout(data.PingTimeout())
	}
	if data.GetRawPingInterval() != nil {
		s.SetPingInterval(data.PingInterval())
	}
	if data.GetRawUpgradeTimeout() != nil {
		s.SetUpgradeTimeout(data.UpgradeTimeout())
	}
	if data.GetRawMaxHttpBufferSize() != nil {
		s.SetMaxHttpBufferSize(data.MaxHttpBufferSize())
	}
	if data.GetRawAllowRequest() != nil {
		s.SetAllowRequest(data.AllowRequest())
	}
	if data.GetRawTransports() != nil {
		s.SetTransports(data.Transports())
	}
	if data.GetRawAllowUpgrades() != nil {
		s.SetAllowUpgrades(data.AllowUpgrades())
	}
	if data.GetRawPerMessageDeflate() != nil {
		s.SetPerMessageDeflate(data.PerMessageDeflate())
	}
	if data.GetRawHttpCompression() != nil {
		s.SetHttpCompression(data.HttpCompression())
	}
	if data.GetRawInitialPacket() != nil {
		s.SetInitialPacket(data.InitialPacket())
	}
	if data.GetRawCookie() != nil {
		s.SetCookie(data.Cookie())
	}
	if data.GetRawCors() != nil {
		s.SetCors(data.Cors())
	}
	if data.GetRawAllowEIO3() != nil {
		s.SetAllowEIO3(data.AllowEIO3())
	}
	if data.GetRawIdleTimeout() != nil {
		s.SetIdleTimeout(data.IdleTimeout())
	}

	return s
}

// SetPingTimeout sets how many ms without a pong packet to consider the connection closed.
func (s *ServerOpts) SetPingTimeout(pingTimeout time.Duration) {
	s.pingTimeout = box.NewSome(pingTimeout)
}

// GetRawPingTimeout returns the ping timeout setting as a [box.Optional].
func (s *ServerOpts) GetRawPingTimeout() box.Optional[time.Duration] {
	return s.pingTimeout
}

// PingTimeout returns the ping timeout setting as a [time.Duration].
func (s *ServerOpts) PingTimeout() time.Duration {
	if s.pingTimeout == nil {
		return 0
	}

	return s.pingTimeout.Get()
}

// SetPingInterval sets how many ms before sending a new ping packet.
func (s *ServerOpts) SetPingInterval(pingInterval time.Duration) {
	s.pingInterval = box.NewSome(pingInterval)
}

// GetRawPingInterval returns the ping interval setting as a [box.Optional].
func (s *ServerOpts) GetRawPingInterval() box.Optional[time.Duration] {
	return s.pingInterval
}

// PingInterval returns the ping interval setting as a [time.Duration].
func (s *ServerOpts) PingInterval() time.Duration {
	if s.pingInterval == nil {
		return 0
	}

	return s.pingInterval.Get()
}

// SetUpgradeTimeout sets how many ms before an uncompleted transport upgrade is canceled.
func (s *ServerOpts) SetUpgradeTimeout(upgradeTimeout time.Duration) {
	s.upgradeTimeout = box.NewSome(upgradeTimeout)
}

// GetRawUpgradeTimeout returns the upgrade timeout setting as a [box.Optional].
func (s *ServerOpts) GetRawUpgradeTimeout() box.Optional[time.Duration] {
	return s.upgradeTimeout
}

// UpgradeTimeout returns the upgrade timeout setting as a [time.Duration].
func (s *ServerOpts) UpgradeTimeout() time.Duration {
	if s.upgradeTimeout == nil {
		return 0
	}

	return s.upgradeTimeout.Get()
}

// SetMaxHttpBufferSize sets how many bytes or characters a message can be before closing the session.
func (s *ServerOpts) SetMaxHttpBufferSize(maxHttpBufferSize int64) {
	s.maxHttpBufferSize = box.NewSome(maxHttpBufferSize)
}

// GetRawMaxHttpBufferSize returns the max HTTP buffer size setting as a [box.Optional].
func (s *ServerOpts) GetRawMaxHttpBufferSize() box.Optional[int64] {
	return s.maxHttpBufferSize
}

// MaxHttpBufferSize returns the max HTTP buffer size setting as an [int64].
func (s *ServerOpts) MaxHttpBufferSize() int64 {
	if s.maxHttpBufferSize == nil {
		return 0
	}

	return s.maxHttpBufferSize.Get()
}

// SetAllowRequest sets a function that receives a given handshake or upgrade request as its first parameter,
// and can decide whether to continue or not.
func (s *ServerOpts) SetAllowRequest(allowRequest AllowRequest) {
	s.allowRequest = box.NewSome(allowRequest)
}

// GetRawAllowRequest returns the allow request function setting as a [box.Optional].
func (s *ServerOpts) GetRawAllowRequest() box.Optional[AllowRequest] {
	return s.allowRequest
}

// AllowRequest returns the allow request function setting as an [AllowRequest].
func (s *ServerOpts) AllowRequest() AllowRequest {
	if s.allowRequest == nil {
		return nil
	}

	return s.allowRequest.Get()
}

// SetTransports sets the low-level transports that are enabled. WebTransport is disabled by default and must be manually enabled.
func (s *ServerOpts) SetTransports(transports *anvil.Set[transports.TConstructor]) {
	s.transports = box.NewSome(transports)
}

// GetRawTransports returns the enabled transports setting as a [box.Optional].
func (s *ServerOpts) GetRawTransports() box.Optional[*anvil.Set[transports.TConstructor]] {
	return s.transports
}

// Transports returns the enabled transports setting as a [*anvil.Set].
func (s *ServerOpts) Transports() *anvil.Set[transports.TConstructor] {
	if s.transports == nil {
		return nil
	}

	return s.transports.Get()
}

// SetAllowUpgrades sets whether to allow transport upgrades.
func (s *ServerOpts) SetAllowUpgrades(allowUpgrades bool) {
	s.allowUpgrades = box.NewSome(allowUpgrades)
}

// GetRawAllowUpgrades returns the allow upgrades setting as a [box.Optional].
func (s *ServerOpts) GetRawAllowUpgrades() box.Optional[bool] {
	return s.allowUpgrades
}

// AllowUpgrades returns the allow upgrades setting as a [bool].
func (s *ServerOpts) AllowUpgrades() bool {
	if s.allowUpgrades == nil {
		return false
	}

	return s.allowUpgrades.Get()
}

// SetPerMessageDeflate sets the parameters of the WebSocket permessage-deflate extension. Set to false to disable.
func (s *ServerOpts) SetPerMessageDeflate(perMessageDeflate *xhttp.PerMessageDeflate) {
	s.perMessageDeflate = box.NewSome(perMessageDeflate)
}

// GetRawPerMessageDeflate returns the permessage-deflate setting as a [box.Optional].
func (s *ServerOpts) GetRawPerMessageDeflate() box.Optional[*xhttp.PerMessageDeflate] {
	return s.perMessageDeflate
}

// PerMessageDeflate returns the permessage-deflate setting as an [*xhttp.PerMessageDeflate].
func (s *ServerOpts) PerMessageDeflate() *xhttp.PerMessageDeflate {
	if s.perMessageDeflate == nil {
		return nil
	}

	return s.perMessageDeflate.Get()
}

// SetHttpCompression sets the parameters of the http compression for the polling transports. Set to false to disable.
func (s *ServerOpts) SetHttpCompression(httpCompression *xhttp.Compression) {
	s.httpCompression = box.NewSome(httpCompression)
}

// GetRawHttpCompression returns the HTTP compression setting as a [box.Optional].
func (s *ServerOpts) GetRawHttpCompression() box.Optional[*xhttp.Compression] {
	return s.httpCompression
}

// HttpCompression returns the HTTP compression setting as an [*xhttp.Compression].
func (s *ServerOpts) HttpCompression() *xhttp.Compression {
	if s.httpCompression == nil {
		return nil
	}

	return s.httpCompression.Get()
}

// SetInitialPacket sets an optional packet which will be concatenated to the handshake packet emitted by Engine.IO.
func (s *ServerOpts) SetInitialPacket(initialPacket io.Reader) {
	s.initialPacket = box.NewSome(initialPacket)
}

// GetRawInitialPacket returns the initial packet setting as a [box.Optional].
func (s *ServerOpts) GetRawInitialPacket() box.Optional[io.Reader] {
	return s.initialPacket
}

// InitialPacket returns the initial packet setting as an [io.Reader].
func (s *ServerOpts) InitialPacket() io.Reader {
	if s.initialPacket == nil {
		return nil
	}

	return s.initialPacket.Get()
}

// SetCookie sets the configuration of the cookie that contains the client sid to send as part of handshake response headers.
func (s *ServerOpts) SetCookie(cookie *http.Cookie) {
	s.cookie = box.NewSome(cookie)
}

// GetRawCookie returns the cookie setting as a [box.Optional].
func (s *ServerOpts) GetRawCookie() box.Optional[*http.Cookie] {
	return s.cookie
}

// Cookie returns the cookie setting as an [*http.Cookie].
func (s *ServerOpts) Cookie() *http.Cookie {
	if s.cookie == nil {
		return nil
	}

	return s.cookie.Get()
}

// SetCors sets the options that will be forwarded to the cors module.
func (s *ServerOpts) SetCors(cors *xhttp.Cors) {
	s.cors = box.NewSome(cors)
}

// GetRawCors returns the CORS setting as a [box.Optional].
func (s *ServerOpts) GetRawCors() box.Optional[*xhttp.Cors] {
	return s.cors
}

// Cors returns the CORS setting as an [*xhttp.Cors].
func (s *ServerOpts) Cors() *xhttp.Cors {
	if s.cors == nil {
		return nil
	}

	return s.cors.Get()
}

// SetAllowEIO3 sets whether to enable compatibility with Socket.IO v2 clients.
func (s *ServerOpts) SetAllowEIO3(allowEIO3 bool) {
	s.allowEIO3 = box.NewSome(allowEIO3)
}

// GetRawAllowEIO3 returns the AllowEIO3 setting as a [box.Optional].
func (s *ServerOpts) GetRawAllowEIO3() box.Optional[bool] {
	return s.allowEIO3
}

// AllowEIO3 returns the AllowEIO3 setting as a [bool].
func (s *ServerOpts) AllowEIO3() bool {
	if s.allowEIO3 == nil {
		return false
	}

	return s.allowEIO3.Get()
}

// SetIdleTimeout sets the maximum amount of seconds that may pass without sending or getting a message.
func (s *ServerOpts) SetIdleTimeout(idleTimeout time.Duration) {
	s.idleTimeout = box.NewSome(idleTimeout)
}

// GetRawIdleTimeout returns the idle timeout setting as a [box.Optional].
func (s *ServerOpts) GetRawIdleTimeout() box.Optional[time.Duration] {
	return s.idleTimeout
}

// IdleTimeout returns the idle timeout setting as a [time.Duration].
func (s *ServerOpts) IdleTimeout() time.Duration {
	if s.idleTimeout == nil {
		return 0
	}

	return s.idleTimeout.Get()
}
