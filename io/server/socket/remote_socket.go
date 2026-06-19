package socket

import (
	"time"

	"github.com/atendi9/ws/io/pkg/anvil"
)

type (
	SocketDetails interface {
		Id() SocketId
		Handshake() *Handshake
		Rooms() *anvil.Set[Room]
		Data() any
	}

	// Expose of subset of the attributes and methods of the Socket struct
	RemoteSocket struct {
		id        SocketId
		handshake *Handshake
		rooms     *anvil.Set[Room]
		data      any

		operator *BroadcastOperator
	}
)

func MakeRemoteSocket() *RemoteSocket {
	r := &RemoteSocket{}
	return r
}

func NewRemoteSocket(adapter Adapter, details SocketDetails) *RemoteSocket {
	r := MakeRemoteSocket()

	r.Construct(adapter, details)

	return r
}

func (r *RemoteSocket) Id() SocketId {
	return r.id
}

func (r *RemoteSocket) Handshake() *Handshake {
	return r.handshake
}

func (r *RemoteSocket) Rooms() *anvil.Set[Room] {
	return r.rooms
}

func (r *RemoteSocket) Data() any {
	return r.data
}

func (r *RemoteSocket) Construct(adapter Adapter, details SocketDetails) {
	r.id = details.Id()
	r.handshake = details.Handshake()
	r.rooms = anvil.NewSet(details.Rooms().Keys()...)
	r.data = details.Data()
	r.operator = NewBroadcastOperator(adapter, anvil.NewSet(Room(r.id)), anvil.NewSet[Room](), &BroadcastFlags{
		ExpectSingleResponse: true, // so that remoteSocket.Emit() with acknowledgement behaves like socket.Emit()
	})
}

// Timeout sets a timeout for the next operation on this remote socket.
// The timeout parameter specifies the duration to wait for an acknowledgement.
func (r *RemoteSocket) Timeout(timeout time.Duration) *BroadcastOperator {
	return r.operator.Timeout(timeout)
}

func (r *RemoteSocket) Emit(ev string, args ...any) error {
	return r.operator.Emit(ev, args...)
}

// Join adds this remote socket to one or more rooms.
// The room parameter accepts one or more Room values.
func (r *RemoteSocket) Join(room ...Room) {
	r.operator.SocketsJoin(room...)
}

// Leave removes this remote socket from one or more rooms.
// The room parameter accepts one or more Room values.
func (r *RemoteSocket) Leave(room ...Room) {
	r.operator.SocketsLeave(room...)
}

// Disconnect disconnects this client. If status is true, the underlying connection will be closed.
func (r *RemoteSocket) Disconnect(status bool) *RemoteSocket {
	r.operator.DisconnectSockets(status)
	return r
}
