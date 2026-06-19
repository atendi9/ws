package socket

import (
	"testing"
	"time"

	"github.com/atendi9/ws/io/pkg/anvil"

	"github.com/atendi9/capivara/assert"
)

func TestMakeBroadcastOperator(t *testing.T) {
	b := MakeBroadcastOperator()
	assert.NotNil(t, b)
	assert.NotNil(t, b.rooms)
	assert.NotNil(t, b.exceptRooms)
	assert.NotNil(t, b.flags)
	assert.Equal(t, 0, b.rooms.Len())
}

func TestBroadcastOperator_ToIn(t *testing.T) {
	b := MakeBroadcastOperator()

	to := b.To("room1", "room2")
	assert.True(t, to.rooms.Has("room1"))
	assert.True(t, to.rooms.Has("room2"))
	// Original operator is untouched (immutability of chaining).
	assert.Equal(t, 0, b.rooms.Len())

	in := b.In("room3")
	assert.True(t, in.rooms.Has("room3"))

	// Chaining accumulates rooms.
	chained := to.To("room4")
	assert.True(t, chained.rooms.Has("room1"))
	assert.True(t, chained.rooms.Has("room4"))
}

func TestBroadcastOperator_Except(t *testing.T) {
	b := MakeBroadcastOperator()
	ex := b.Except("blocked")
	assert.True(t, ex.exceptRooms.Has("blocked"))
	assert.Equal(t, 0, b.exceptRooms.Len())
}

func TestBroadcastOperator_Flags(t *testing.T) {
	b := MakeBroadcastOperator()

	comp := b.Compress(true)
	assert.NotNil(t, comp.flags.Compress)
	assert.True(t, *comp.flags.Compress)

	vol := b.Volatile()
	assert.True(t, vol.flags.Volatile)

	local := b.Local()
	assert.True(t, local.flags.Local)

	to := b.Timeout(3 * time.Second)
	assert.NotNil(t, to.flags.Timeout)
	assert.Equal(t, 3*time.Second, *to.flags.Timeout)

	// Original flags remain default.
	assert.False(t, b.flags.Volatile)
	assert.False(t, b.flags.Local)
	assert.True(t, b.flags.Timeout == nil)
}

func TestBroadcastOperator_Construct(t *testing.T) {
	b := MakeBroadcastOperator()
	rooms := anvil.NewSet[Room]("a")
	except := anvil.NewSet[Room]("b")
	flags := &BroadcastFlags{Local: true}

	b.Construct(nil, rooms, except, flags)
	assert.True(t, b.rooms.Has("a"))
	assert.True(t, b.exceptRooms.Has("b"))
	assert.True(t, b.flags.Local)

	// nil rooms/flags keep existing values.
	b.Construct(nil, nil, nil, nil)
	assert.True(t, b.rooms.Has("a"))
	assert.True(t, b.flags.Local)
}

func TestBroadcastOperator_EmitReservedEvent(t *testing.T) {
	b := MakeBroadcastOperator()
	// Reserved event names are rejected before the adapter is touched.
	for _, ev := range []string{"connect", "disconnect", "disconnecting"} {
		assert.Error(t, b.Emit(ev))
	}
}
