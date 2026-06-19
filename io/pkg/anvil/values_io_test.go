package anvil

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestNewInput(t *testing.T) {
	assert.Equal(t, 42, NewInput(42))
	assert.Equal(t, "str", NewInput("str"))
	assert.True(t, NewInput(nil) == nil)
}

func TestNewOutput(t *testing.T) {
	assert.Equal(t, 7, NewOutput(7))
	assert.Equal(t, "x", NewOutput("x"))
	assert.True(t, NewOutput(nil) == nil)
}
