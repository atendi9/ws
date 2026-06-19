package socket

import (
	"time"

	"github.com/atendi9/ws/io/parsers/engine/packet"
	"github.com/atendi9/ws/io/parsers/socket/parser"
	"github.com/atendi9/ws/io/pkg/anvil"
	"github.com/atendi9/ws/io/pkg/events"
)

type (
	// A public ID, sent by the server at the beginning of the Socket.IO session and which can be used for private messaging
	SocketId string

	// A private ID, sent by the server at the beginning of the Socket.IO session and used for connection state recovery
	// upon reconnection
	PrivateSessionId string

	// we could extend the Room type to "string", but that would be a breaking change
	// Related: https://github.com/socketio/socket.io-redis-adapter/issues/418
	Room string

	WriteOptions struct {
		packet.Options

		Volatile   bool `json:"volatile"`
		PreEncoded bool `json:"preEncoded"`
	}

	BroadcastFlags struct {
		WriteOptions

		Local     bool           `json:"local"`
		Broadcast bool           `json:"broadcast"`
		Binary    bool           `json:"binary"`
		Timeout   *time.Duration `json:"timeout,omitempty"`

		ExpectSingleResponse bool `json:"expectSingleResponse"`
	}

	BroadcastOptions struct {
		Rooms  *anvil.Set[Room] `json:"rooms,omitempty"`
		Except *anvil.Set[Room] `json:"except,omitempty"`
		Flags  *BroadcastFlags  `json:"flags,omitempty"`
	}

	SessionToPersist struct {
		Sid   SocketId         `json:"sid"`
		Pid   PrivateSessionId `json:"pid"`
		Rooms *anvil.Set[Room] `json:"rooms,omitempty"`
		Data  any              `json:"data"`
	}

	Session struct {
		*SessionToPersist

		MissedPackets []any `json:"missedPackets"`
	}

	PersistedPacket struct {
		Id        string            `json:"id"`
		EmittedAt int64             `json:"emittedAt"`
		Data      any               `json:"data"`
		Opts      *BroadcastOptions `json:"opts,omitempty"`
	}

	SessionWithTimestamp struct {
		*SessionToPersist

		DisconnectedAt int64 `json:"disconnectedAt"`
	}

	Adapter interface {
		events.Emitter

		// #prototype

		Prototype(Adapter)
		Proto() Adapter

		Rooms() *anvil.Map[Room, *anvil.Set[SocketId]]
		Sids() *anvil.Map[SocketId, *anvil.Set[Room]]
		Nsp() Namespace

		// Construct() should be called after calling Prototype()
		Construct(Namespace)

		// To be overridden
		Init()

		// To be overridden
		Close()

		// Returns the number of Socket.IO servers in the cluster
		ServerCount() int64

		// Adds a socket to a list of room.
		AddAll(SocketId, *anvil.Set[Room])

		// Removes a socket from a room.
		Del(SocketId, Room)

		// Removes a socket from all rooms it's joined.
		DelAll(SocketId)

		// Broadcasts a packet.
		//
		// Options:
		//  - `Flags` {*BroadcastFlags} flags for this packet
		//  - `Except` {*anvil.Set[Room]} sids that should be excluded
		//  - `Rooms` {*anvil.Set[Room]} list of rooms to broadcast to
		Broadcast(*parser.Packet, *BroadcastOptions)

		// Broadcasts a packet and expects multiple acknowledgements.
		//
		// Options:
		//  - `Flags` {*BroadcastFlags} flags for this packet
		//  - `Except` {*anvil.Set[Room]} sids that should be excluded
		//  - `Rooms` {*anvil.Set[Room]} list of rooms to broadcast to
		BroadcastWithAck(*parser.Packet, *BroadcastOptions, func(uint64), Ack)

		// Gets a list of sockets by sid.
		Sockets(*anvil.Set[Room]) *anvil.Set[SocketId]

		// Gets the list of rooms a given socket has joined.
		SocketRooms(SocketId) *anvil.Set[Room]

		// Returns the matching socket instances
		FetchSockets(*BroadcastOptions) func(func([]SocketDetails, error))

		// Makes the matching socket instances join the specified rooms
		AddSockets(*BroadcastOptions, []Room)

		// Makes the matching socket instances leave the specified rooms
		DelSockets(*BroadcastOptions, []Room)

		// Makes the matching socket instances disconnect
		DisconnectSockets(*BroadcastOptions, bool)

		// Send a packet to the other Socket.IO servers in the cluster
		ServerSideEmit([]any) error

		// Save the client session in order to restore it upon reconnection.
		PersistSession(*SessionToPersist)

		// Restore the session and find the packets that were missed by the client.
		RestoreSession(PrivateSessionId, string) (*Session, error)
	}

	SessionAwareAdapter interface {
		Adapter
	}

	ParentBroadcastAdapter interface {
		Adapter
	}

	AdapterConstructor interface {
		New(Namespace) Adapter
	}
)
