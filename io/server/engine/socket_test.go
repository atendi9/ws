package engine

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestMakeSocketDefaults(t *testing.T) {
	s := MakeSocket()
	assert.NotNil(t, s)

	assert.Equal(t, 0, s.Protocol())
	assert.False(t, s.Upgraded())
	assert.False(t, s.Upgrading())
	assert.Equal(t, "", s.Id())
	assert.Equal(t, "", s.RemoteAddress())
	assert.True(t, s.Request() == nil)
	assert.True(t, s.Transport() == nil)

	// Fresh socket starts in the opening state.
	assert.Equal(t, StateOpening.String(), s.ReadyState())
}

func TestSocketSetReadyState(t *testing.T) {
	s := MakeSocket()

	s.SetReadyState(StateOpen)
	assert.Equal(t, StateOpen.String(), s.ReadyState())

	s.SetReadyState(StateClosing)
	assert.Equal(t, StateClosing.String(), s.ReadyState())

	s.SetReadyState(StateClosed)
	assert.Equal(t, StateClosed.String(), s.ReadyState())
}
