package transports

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestInit(t *testing.T) {
	ts := Transports()
	assert.Equal(t, transports[POLLING], ts[POLLING])
	assert.Equal(t, transports[WEBSOCKET], ts[WEBSOCKET])
}
