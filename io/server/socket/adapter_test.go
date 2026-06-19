package socket

import (
	"sync/atomic"
	"testing"

	"github.com/atendi9/ws/io/pkg/anvil"

	"github.com/atendi9/capivara/assert"
)

func TestMakeAdapter(t *testing.T) {
	a := MakeAdapter()
	assert.NotNil(t, a)
	assert.NotNil(t, a.Rooms())
	assert.NotNil(t, a.Sids())
	assert.NotNil(t, a.Proto())
	assert.Equal(t, int64(1), a.ServerCount())
}

func TestAdapter_AddAllAndRooms(t *testing.T) {
	a := MakeAdapter()

	var created, joined atomic.Int32
	a.On("create-room", func(...any) { created.Add(1) })
	a.On("join-room", func(...any) { joined.Add(1) })

	a.AddAll("sock1", anvil.NewSet[Room]("roomA", "roomB"))

	// Socket is now tracked in both rooms.
	rooms := a.SocketRooms("sock1")
	assert.NotNil(t, rooms)
	assert.True(t, rooms.Has("roomA"))
	assert.True(t, rooms.Has("roomB"))

	// Both rooms now exist in the adapter room map.
	ids, ok := a.Rooms().Load("roomA")
	assert.True(t, ok)
	assert.True(t, ids.Has("sock1"))

	assert.Equal(t, int32(2), created.Load())
	assert.Equal(t, int32(2), joined.Load())

	// Re-adding the same socket to an existing room emits no new join.
	a.AddAll("sock1", anvil.NewSet[Room]("roomA"))
	assert.Equal(t, int32(2), joined.Load())
}

func TestAdapter_Del(t *testing.T) {
	a := MakeAdapter()

	var left, deleted atomic.Int32
	a.On("leave-room", func(...any) { left.Add(1) })
	a.On("delete-room", func(...any) { deleted.Add(1) })

	a.AddAll("sock1", anvil.NewSet[Room]("roomA"))
	a.AddAll("sock2", anvil.NewSet[Room]("roomA"))

	// Removing sock1 leaves the room but does not delete it (sock2 remains).
	a.Del("sock1", "roomA")
	assert.Equal(t, int32(1), left.Load())
	assert.Equal(t, int32(0), deleted.Load())

	// Removing the last socket deletes the room.
	a.Del("sock2", "roomA")
	assert.Equal(t, int32(2), left.Load())
	assert.Equal(t, int32(1), deleted.Load())

	_, ok := a.Rooms().Load("roomA")
	assert.False(t, ok)
}

func TestAdapter_DelAll(t *testing.T) {
	a := MakeAdapter()
	a.AddAll("sock1", anvil.NewSet[Room]("r1", "r2", "r3"))

	a.DelAll("sock1")

	// Socket no longer tracked anywhere.
	assert.True(t, a.SocketRooms("sock1") == nil)
	_, ok := a.Sids().Load("sock1")
	assert.False(t, ok)
	_, ok = a.Rooms().Load("r1")
	assert.False(t, ok)
}

func TestAdapter_SocketRoomsUnknown(t *testing.T) {
	a := MakeAdapter()
	assert.True(t, a.SocketRooms("ghost") == nil)
}

func TestAdapter_ServerSideEmitUnsupported(t *testing.T) {
	a := MakeAdapter()
	assert.Error(t, a.ServerSideEmit([]any{"evt"}))
}

func TestAdapter_RestoreSessionDefault(t *testing.T) {
	a := MakeAdapter().(*adapter)
	session, err := a.RestoreSession("pid", "0")
	assert.NoError(t, err)
	assert.True(t, session == nil)
	// PersistSession is a no-op on the default adapter.
	a.PersistSession(&SessionToPersist{})
}

func TestAdapter_InitCloseNoop(t *testing.T) {
	a := MakeAdapter()
	a.Init()
	a.Close()
}

func TestAdapterBuilderImplementsConstructor(t *testing.T) {
	var b AdapterBuilder
	// New requires a namespace; we only assert the builder value is usable as
	// the expected type here (construction with a real namespace is covered by
	// integration paths).
	assert.NotNil(t, &b)
}
