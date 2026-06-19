package socket

import (
	"sync/atomic"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestStrictEmitter_OnEmit(t *testing.T) {
	e := NewStrictEmitter()
	assert.NotNil(t, e)

	var got atomic.Int32
	assert.NoError(t, e.On("tick", func(args ...any) {
		got.Add(1)
	}))

	e.Emit("tick")
	e.Emit("tick")
	assert.Equal(t, int32(2), got.Load())
}

func TestStrictEmitter_Once(t *testing.T) {
	e := NewStrictEmitter()

	var got atomic.Int32
	assert.NoError(t, e.Once("boot", func(args ...any) {
		got.Add(1)
	}))

	e.Emit("boot")
	e.Emit("boot")
	assert.Equal(t, int32(1), got.Load()) // only fires once
}

func TestStrictEmitter_EmitReservedAndUntyped(t *testing.T) {
	e := NewStrictEmitter()

	var reserved, untyped atomic.Bool
	assert.NoError(t, e.On("evt", func(args ...any) {
		if len(args) > 0 && args[0] == "reserved" {
			reserved.Store(true)
		}
		if len(args) > 0 && args[0] == "untyped" {
			untyped.Store(true)
		}
	}))

	e.EmitReserved("evt", "reserved")
	e.EmitUntyped("evt", "untyped")
	assert.True(t, reserved.Load())
	assert.True(t, untyped.Load())
}

func TestStrictEmitter_Listeners(t *testing.T) {
	e := NewStrictEmitter()
	assert.Equal(t, 0, len(e.Listeners("x")))

	assert.NoError(t, e.On("x", func(args ...any) {}))
	assert.NoError(t, e.On("x", func(args ...any) {}))
	assert.Equal(t, 2, len(e.Listeners("x")))
}

func TestStrictEmitter_EmitArguments(t *testing.T) {
	e := NewStrictEmitter()

	var received []any
	assert.NoError(t, e.On("data", func(args ...any) {
		received = args
	}))

	e.Emit("data", 1, "two", true)
	assert.Equal(t, 3, len(received))
	assert.Equal(t, 1, received[0])
	assert.Equal(t, "two", received[1])
	assert.Equal(t, true, received[2])
}
