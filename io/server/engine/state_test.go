package engine

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestState(t *testing.T) {
	stateMsg := "error"
	state := NewServerState(stateMsg)
	assert.Equal(t, "server-engine", state.Id())
	assert.Equal(t, state.String(), stateMsg)
}
