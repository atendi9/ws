package transports

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

// TestWebSocketBuilder runs unit tests for the [WebSocketBuilder] implementation.
func TestWebSocketBuilder(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		defer func() {
			recover()
		}()

		builder := &WebSocketBuilder{}
		transport := builder.New(nil)
		assert.True(t, transport != nil)
	})

	t.Run("Name", func(t *testing.T) {
		builder := &WebSocketBuilder{}
		name := builder.Name()
		assert.Equal(t, WEBSOCKET, name)
	})

	t.Run("HandlesUpgrades", func(t *testing.T) {
		builder := &WebSocketBuilder{}
		handles := builder.HandlesUpgrades()
		assert.True(t, handles)
	})

	t.Run("UpgradesTo", func(t *testing.T) {
		builder := &WebSocketBuilder{}
		upgrades := builder.UpgradesTo()
		assert.True(t, upgrades == nil)
	})
}

// TestPollingBuilder runs unit tests for the [PollingBuilder] implementation.
func TestPollingBuilder(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		defer func() {
			recover()
		}()

		builder := &PollingBuilder{}
		transport := builder.New(nil)
		assert.True(t, transport != nil)
	})

	t.Run("Name", func(t *testing.T) {
		builder := &PollingBuilder{}
		name := builder.Name()
		assert.Equal(t, POLLING, name)
	})

	t.Run("HandlesUpgrades", func(t *testing.T) {
		builder := &PollingBuilder{}
		handles := builder.HandlesUpgrades()
		assert.False(t, handles)
	})

	t.Run("UpgradesTo", func(t *testing.T) {
		builder := &PollingBuilder{}
		upgrades := builder.UpgradesTo()
		assert.LengthSlice(t, 1, upgrades)
		assert.Equal(t, WEBSOCKET, upgrades[0])
	})
}
