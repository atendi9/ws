package events

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestNewEmitter(t *testing.T) {
	emitter := NewEmitter()
	assert.Equal(t, 0, emitter.Len())
}

func TestAddListenerAndEmit(t *testing.T) {
	emitter := NewEmitter()
	var called bool
	var receivedVal string

	fn := func(args ...any) {
		called = true
		receivedVal = args[0].(string)
	}

	err := emitter.AddListener("test-event", fn)
	assert.NoError(t, err)
	assert.Equal(t, 1, emitter.ListenerCount("test-event"))

	emitter.Emit("test-event", "hello")
	assert.True(t, called)
	assert.Equal(t, "hello", receivedVal)
}

func TestOnAlias(t *testing.T) {
	emitter := NewEmitter()
	var called bool

	fn := func(args ...any) {
		called = true
	}

	err := emitter.On("test-on", fn)
	assert.NoError(t, err)
	assert.Equal(t, 1, emitter.ListenerCount("test-on"))

	emitter.Emit("test-on")
	assert.True(t, called)
}

func TestOnceListener(t *testing.T) {
	emitter := NewEmitter()
	count := 0

	fn := func(args ...any) {
		count++
	}

	err := emitter.Once("test-once", fn)
	assert.NoError(t, err)
	assert.Equal(t, 1, emitter.ListenerCount("test-once"))

	emitter.Emit("test-once")
	assert.Equal(t, 1, count)
	assert.Equal(t, 0, emitter.ListenerCount("test-once"))

	emitter.Emit("test-once")
	assert.Equal(t, 1, count)
}

func TestRemoveListener(t *testing.T) {
	emitter := NewEmitter()
	var called bool

	fn := func(args ...any) {
		called = true
	}

	_ = emitter.AddListener("test-remove", fn)
	assert.Equal(t, 1, emitter.ListenerCount("test-remove"))

	removed := emitter.RemoveListener("test-remove", fn)
	assert.True(t, removed)
	assert.Equal(t, 0, emitter.ListenerCount("test-remove"))

	emitter.Emit("test-remove")
	assert.False(t, called)
}

func TestRemoveAllListeners(t *testing.T) {
	emitter := NewEmitter()
	fn1 := func(args ...any) {}
	fn2 := func(args ...any) {}

	_ = emitter.AddListener("event-clear", fn1, fn2)
	assert.Equal(t, 2, emitter.ListenerCount("event-clear"))

	removed := emitter.RemoveAllListeners("event-clear")
	assert.True(t, removed)
	assert.Equal(t, 0, emitter.ListenerCount("event-clear"))
}

func TestClearAndLen(t *testing.T) {
	emitter := NewEmitter()
	fn := func(args ...any) {}

	_ = emitter.AddListener("evt-1", fn)
	_ = emitter.AddListener("evt-2", fn)
	assert.Equal(t, 2, emitter.Len())

	emitter.Clear()
	assert.Equal(t, 0, emitter.Len())
}

func TestNamesAndListeners(t *testing.T) {
	emitter := NewEmitter()
	fn := func(args ...any) {}

	_ = emitter.AddListener("evt-1", fn)

	names := emitter.Names()
	assert.LengthSlice(t, 1, names)
	assert.Equal(t, Name("evt-1"), names[0])

	listeners := emitter.Listeners("evt-1")
	assert.LengthSlice(t, 1, listeners)
}

func TestCopyTo(t *testing.T) {
	emitter := NewEmitter()
	var called bool
	fn := func(args ...any) {
		called = true
	}

	eventsMap := Events{
		"copy-evt": []Listener{fn},
	}

	eventsMap.CopyTo(emitter)
	assert.Equal(t, 1, emitter.ListenerCount("copy-evt"))

	emitter.Emit("copy-evt")
	assert.True(t, called)
}

func TestEmitPanicRecovery(t *testing.T) {
	emitter := NewEmitter()
	fnPanic := func(args ...any) {
		panic("something went wrong inside listener")
	}
	var calledNext bool
	fnNext := func(args ...any) {
		calledNext = true
	}

	_ = emitter.AddListener("panic-event", fnPanic, fnNext)

	// Should not crash the program due to internal recovery mechanism inside Emit()
	emitter.Emit("panic-event")
	assert.True(t, calledNext)
}
