package socket

import (
	"time"

	"github.com/atendi9/box"
	"github.com/atendi9/ws/io/parsers/socket/parser"
	"github.com/atendi9/ws/io/server/engine/config"
)

type (
	ConnectionStateRecovery interface {
		SetMaxDisconnectionDuration(int64)
		GetRawMaxDisconnectionDuration() box.Optional[int64]
		MaxDisconnectionDuration() int64

		SetSkipMiddlewares(bool)
		GetRawSkipMiddlewares() box.Optional[bool]
		SkipMiddlewares() bool

		SetSessionCleanupInterval(time.Duration)
		GetRawSessionCleanupInterval() box.Optional[time.Duration]
		SessionCleanupInterval() time.Duration
	}

	CSRecovery struct {
		// The backup duration of the sessions and the packets.
		maxDisconnectionDuration box.Optional[int64]

		// Whether to skip middlewares upon successful connection state recovery.
		skipMiddlewares box.Optional[bool]

		// The interval between two session cleanup sweeps.
		sessionCleanupInterval box.Optional[time.Duration]
	}

	ServerOptions interface {
		config.Options

		SetServeClient(bool)
		GetRawServeClient() box.Optional[bool]
		ServeClient() bool

		SetClientVersion(string)
		GetRawClientVersion() box.Optional[string]
		ClientVersion() string

		SetAdapter(AdapterConstructor)
		GetRawAdapter() box.Optional[AdapterConstructor]
		Adapter() AdapterConstructor

		SetParser(parser.Parser)
		GetRawParser() box.Optional[parser.Parser]
		Parser() parser.Parser

		SetConnectTimeout(time.Duration)
		GetRawConnectTimeout() box.Optional[time.Duration]
		ConnectTimeout() time.Duration

		SetConnectionStateRecovery(ConnectionStateRecovery)
		GetRawConnectionStateRecovery() box.Optional[ConnectionStateRecovery]
		ConnectionStateRecovery() ConnectionStateRecovery

		SetCleanupEmptyChildNamespaces(bool)
		GetRawCleanupEmptyChildNamespaces() box.Optional[bool]
		CleanupEmptyChildNamespaces() bool
	}

	ServerOpts struct {
		*config.Opts

		// whether to serve the client files
		serveClient box.Optional[bool]

		// Client file version
		clientVersion box.Optional[string]

		// the adapter to use
		adapter box.Optional[AdapterConstructor]

		// the parser to use
		parser box.Optional[parser.Parser]

		// how many ms before a client without namespace is closed
		connectTimeout box.Optional[time.Duration]

		// Whether to enable the recovery of connection state when a client temporarily disconnects.
		//
		// The connection state includes the missed packets, the rooms the socket was in and the `data` attribute.
		connectionStateRecovery box.Optional[ConnectionStateRecovery]

		// Whether to remove child namespaces that have no sockets connected to them
		cleanupEmptyChildNamespaces box.Optional[bool]
	}
)

func DefaultConnectionStateRecovery() *CSRecovery {
	return &CSRecovery{}
}

func (c *CSRecovery) Assign(data ConnectionStateRecovery) ConnectionStateRecovery {
	if data == nil {
		return c
	}

	if data.GetRawMaxDisconnectionDuration() != nil {
		c.SetMaxDisconnectionDuration(data.MaxDisconnectionDuration())
	}
	if data.GetRawSkipMiddlewares() != nil {
		c.SetSkipMiddlewares(data.SkipMiddlewares())
	}
	if data.GetRawSessionCleanupInterval() != nil {
		c.SetSessionCleanupInterval(data.SessionCleanupInterval())
	}

	return c
}

func DefaultServerOptions() *ServerOpts {
	return &ServerOpts{
		Opts: config.DefaultOptions(),
	}
}

func (s *ServerOpts) Assign(data ServerOptions) ServerOptions {
	if data == nil {
		return s
	}

	s.Opts.Assign(data)

	if data.GetRawServeClient() != nil {
		s.SetServeClient(data.ServeClient())
	}
	if data.GetRawClientVersion() != nil {
		s.SetClientVersion(data.ClientVersion())
	}
	if data.GetRawAdapter() != nil {
		s.SetAdapter(data.Adapter())
	}
	if data.GetRawParser() != nil {
		s.SetParser(data.Parser())
	}
	if data.GetRawConnectTimeout() != nil {
		s.SetConnectTimeout(data.ConnectTimeout())
	}
	if data.GetRawConnectionStateRecovery() != nil {
		s.SetConnectionStateRecovery(data.ConnectionStateRecovery())
	}
	if data.GetRawCleanupEmptyChildNamespaces() != nil {
		s.SetCleanupEmptyChildNamespaces(data.CleanupEmptyChildNamespaces())
	}

	return s
}

func (c *CSRecovery) SetMaxDisconnectionDuration(maxDisconnectionDuration int64) {
	c.maxDisconnectionDuration = box.NewSome(maxDisconnectionDuration)
}
func (c *CSRecovery) GetRawMaxDisconnectionDuration() box.Optional[int64] {
	return c.maxDisconnectionDuration
}
func (c *CSRecovery) MaxDisconnectionDuration() int64 {
	if c.maxDisconnectionDuration == nil {
		return 0
	}

	return c.maxDisconnectionDuration.Get()
}

func (c *CSRecovery) SetSkipMiddlewares(skipMiddlewares bool) {
	c.skipMiddlewares = box.NewSome(skipMiddlewares)
}
func (c *CSRecovery) GetRawSkipMiddlewares() box.Optional[bool] {
	return c.skipMiddlewares
}
func (c *CSRecovery) SkipMiddlewares() bool {
	if c.skipMiddlewares == nil {
		return false
	}

	return c.skipMiddlewares.Get()
}

// SetSessionCleanupInterval sets the interval between two session cleanup sweeps.
func (c *CSRecovery) SetSessionCleanupInterval(sessionCleanupInterval time.Duration) {
	c.sessionCleanupInterval = box.NewSome(sessionCleanupInterval)
}

// GetRawSessionCleanupInterval returns the raw optional value of the session cleanup interval.
func (c *CSRecovery) GetRawSessionCleanupInterval() box.Optional[time.Duration] {
	return c.sessionCleanupInterval
}

// SessionCleanupInterval returns the configured session cleanup interval, or 0 if not set.
func (c *CSRecovery) SessionCleanupInterval() time.Duration {
	if c.sessionCleanupInterval == nil {
		return 0
	}

	return c.sessionCleanupInterval.Get()
}

func (s *ServerOpts) SetServeClient(serveClient bool) {
	s.serveClient = box.NewSome(serveClient)
}
func (s *ServerOpts) GetRawServeClient() box.Optional[bool] {
	return s.serveClient
}
func (s *ServerOpts) ServeClient() bool {
	if s.serveClient == nil {
		return false
	}

	return s.serveClient.Get()
}

func (s *ServerOpts) SetClientVersion(clientVersion string) {
	s.clientVersion = box.NewSome(clientVersion)
}
func (s *ServerOpts) GetRawClientVersion() box.Optional[string] {
	return s.clientVersion
}
func (s *ServerOpts) ClientVersion() string {
	if s.clientVersion == nil {
		return ""
	}

	return s.clientVersion.Get()
}

func (s *ServerOpts) SetAdapter(adapter AdapterConstructor) {
	s.adapter = box.NewSome(adapter)
}
func (s *ServerOpts) GetRawAdapter() box.Optional[AdapterConstructor] {
	return s.adapter
}
func (s *ServerOpts) Adapter() AdapterConstructor {
	if s.adapter == nil {
		return nil
	}

	return s.adapter.Get()
}

func (s *ServerOpts) SetParser(parser parser.Parser) {
	s.parser = box.NewSome(parser)
}
func (s *ServerOpts) GetRawParser() box.Optional[parser.Parser] {
	return s.parser
}
func (s *ServerOpts) Parser() parser.Parser {
	if s.parser == nil {
		return nil
	}

	return s.parser.Get()
}

func (s *ServerOpts) SetConnectTimeout(connectTimeout time.Duration) {
	s.connectTimeout = box.NewSome(connectTimeout)
}
func (s *ServerOpts) GetRawConnectTimeout() box.Optional[time.Duration] {
	return s.connectTimeout
}
func (s *ServerOpts) ConnectTimeout() time.Duration {
	if s.connectTimeout == nil {
		return 0
	}

	return s.connectTimeout.Get()
}

func (s *ServerOpts) SetConnectionStateRecovery(connectionStateRecovery ConnectionStateRecovery) {
	s.connectionStateRecovery = box.NewSome(connectionStateRecovery)
}
func (s *ServerOpts) GetRawConnectionStateRecovery() box.Optional[ConnectionStateRecovery] {
	return s.connectionStateRecovery
}
func (s *ServerOpts) ConnectionStateRecovery() ConnectionStateRecovery {
	if s.connectionStateRecovery == nil {
		return nil
	}

	return s.connectionStateRecovery.Get()
}

func (s *ServerOpts) SetCleanupEmptyChildNamespaces(cleanupEmptyChildNamespaces bool) {
	s.cleanupEmptyChildNamespaces = box.NewSome(cleanupEmptyChildNamespaces)
}
func (s *ServerOpts) GetRawCleanupEmptyChildNamespaces() box.Optional[bool] {
	return s.cleanupEmptyChildNamespaces
}
func (s *ServerOpts) CleanupEmptyChildNamespaces() bool {
	if s.cleanupEmptyChildNamespaces == nil {
		return false
	}

	return s.cleanupEmptyChildNamespaces.Get()
}
