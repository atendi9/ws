package events

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestGlobalAddListenerAndEmit(t *testing.T) {
	// Ensure a pristine global state before running tests
	Clear()

	var called bool
	var receivedVal int

	fn := func(args ...any) {
		called = true
		receivedVal = args[0].(int)
	}

	err := AddListener("global-evt", fn)
	assert.NoError(t, err)
	assert.Equal(t, 1, ListenerCount("global-evt"))

	Emit("global-evt", 42)
	assert.True(t, called)
	assert.Equal(t, 42, receivedVal)
}

func TestGlobalOnAlias(t *testing.T) {
	Clear()

	var called bool
	fn := func(args ...any) {
		called = true
	}

	err := On("global-on", fn)
	assert.NoError(t, err)
	assert.Equal(t, 1, ListenerCount("global-on"))

	Emit("global-on")
	assert.True(t, called)
}

func TestGlobalOnce(t *testing.T) {
	Clear()

	count := 0
	fn := func(args ...any) {
		count++
	}

	err := Once("global-once", fn)
	assert.NoError(t, err)
	assert.Equal(t, 1, ListenerCount("global-once"))

	Emit("global-once")
	assert.Equal(t, 1, count)
	assert.Equal(t, 0, ListenerCount("global-once"))

	Emit("global-once")
	assert.Equal(t, 1, count)
}

func TestGlobalRemoveListener(t *testing.T) {
	Clear()

	var called bool
	fn := func(args ...any) {
		called = true
	}

	_ = AddListener("global-remove", fn)
	assert.Equal(t, 1, ListenerCount("global-remove"))

	removed := RemoveListener("global-remove", fn)
	assert.True(t, removed)
	assert.Equal(t, 0, ListenerCount("global-remove"))

	Emit("global-remove")
	assert.False(t, called)
}

func TestGlobalRemoveAllListeners(t *testing.T) {
	Clear()

	fn := func(args ...any) {}
	_ = AddListener("global-clear-all", fn, fn)
	assert.Equal(t, 2, ListenerCount("global-clear-all"))

	removed := RemoveAllListeners("global-clear-all")
	assert.True(t, removed)
	assert.Equal(t, 0, ListenerCount("global-clear-all"))
}

func TestGlobalClearAndLen(t *testing.T) {
	Clear()

	fn := func(args ...any) {}
	_ = AddListener("evt-a", fn)
	_ = AddListener("evt-b", fn)
	assert.Equal(t, 2, Len())

	Clear()
	assert.Equal(t, 0, Len())
}

func TestGlobalNamesAndListeners(t *testing.T) {
	Clear()

	fn := func(args ...any) {}
	_ = AddListener("evt-names-test", fn)

	names := Names()
	assert.LengthSlice(t, 1, names)
	assert.Equal(t, Name("evt-names-test"), names[0])

	listeners := Listeners("evt-names-test")
	assert.LengthSlice(t, 1, listeners)
}
