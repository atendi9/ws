package state

import (
	"strings"

	"github.com/atendi9/box"
)

// SocketState represents the state of a socket connection.
// It aliases the [State] type.
type SocketState = State

const (
	// SocketOpening indicates the socket is in the process of opening.
	SocketOpening SocketState = "socket::opening"
	// SocketOpen indicates the socket connection is fully established.
	SocketOpen SocketState = "socket::open"
	// SocketClosing indicates the socket is in the process of closing.
	SocketClosing SocketState = "socket::closing"
	// SocketClosed indicates the socket connection is completely closed.
	SocketClosed SocketState = "socket::closed"
)

// TransportState represents the state of an underlying transport layer.
// It aliases the [State] type.
type TransportState = State

const (
	// TransportOpening indicates the transport is preparing to connect.
	TransportOpening TransportState = "transport::opening"
	// TransportOpen indicates the transport connection is active.
	TransportOpen TransportState = "transport::open"
	// TransportClosed indicates the transport connection is closed.
	TransportClosed TransportState = "transport::closed"
	// TransportPausing indicates the transport is transitioning to a paused status.
	TransportPausing TransportState = "transport::pausing"
	// TransportPaused indicates the transport is temporarily suspended.
	TransportPaused TransportState = "transport::paused"
)

// Separator is the delimiter used to split the prefix identifier from the state value.
const Separator = "::"

// State defines a specialized string format composed of an identifier and a value separated by [Separator].
type State string

// String extracts and returns the underlying value part of the [State] formatted string.
func (s State) String() string {
	value, _ := s.Value()
	return value.Get()
}

// Id extracts and returns the prefix identifier of the [State] if valid, otherwise returns an empty string from [box.Optional].
func (s State) Id() string {
	parts := strings.Split(string(s), Separator)
	if len(parts) == 0 || len(parts) != 2 {
		return box.NewNone[string]().Get()
	}
	return box.NewSome(parts[0]).Get()
}

// Value splits the [State] and returns a [box.Optional] containing the suffix value along with a success boolean flag.
func (s State) Value() (value box.Optional[string], ok bool) {
	parts := strings.Split(string(s), Separator)
	if len(parts) == 0 || len(parts) != 2 {
		return box.NewNone[string](), false
	}
	return box.NewSome(parts[len(parts)-1]), true
}

// NewState creates and returns a new [State] by joining the provided identifier
// and state value using the [Separator].
func NewState(
	id string,
	state string,
) State {
	return State(id + Separator + state)
}
