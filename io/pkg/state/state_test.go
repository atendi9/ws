package state

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestState_Id(t *testing.T) {
	t.Run("should successfully return Id for valid states", func(t *testing.T) {
		assert.Equal(t, "socket", NewState("socket", "init").Id())
		assert.Equal(t, "transport", NewState("transport", "init").Id())
	})

	t.Run("should return empty string for invalid states formatting", func(t *testing.T) {
		invalidStateNoSeparator := State("invalidstate")
		invalidStateTooManySeparators := State("invalid::state::format")

		assert.Equal(t, "", invalidStateNoSeparator.Id())
		assert.Equal(t, "", invalidStateTooManySeparators.Id())
	})
}

func TestState_Value(t *testing.T) {
	t.Run("should successfully split and return value for valid states", func(t *testing.T) {
		val, ok := SocketOpen.Value()
		assert.True(t, ok)
		assert.Equal(t, "open", val.Get())

		val, ok = TransportPaused.Value()
		assert.True(t, ok)
		assert.Equal(t, "paused", val.Get())
	})

	t.Run("should return false when state structure is invalid", func(t *testing.T) {
		invalidState := State("malformedState")
		_, ok := invalidState.Value()
		assert.False(t, ok)
	})
}

func TestState_String(t *testing.T) {
	t.Run("should return extracted value as string", func(t *testing.T) {
		assert.Equal(t, "opening", SocketOpening.String())
		assert.Equal(t, "closing", SocketClosing.String())
	})

	t.Run("should return empty string representation for malformed state", func(t *testing.T) {
		invalidState := State("bad-format")
		assert.Equal(t, "", invalidState.String())
	})
}
