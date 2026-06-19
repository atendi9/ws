package parser

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestEncoder(t *testing.T) {
	t.Run("NewEncoder", func(t *testing.T) {
		enc := NewEncoder()

		isNotNil := enc != nil
		assert.True(t, isNotNil)
	})

	t.Run("Encode", func(t *testing.T) {
		enc := NewEncoder()

		t.Run("StandardEventPacket", func(t *testing.T) {
			packet := &Packet{
				Type: EVENT,
				Nsp:  "/",
				Data: []any{"message", "hello"},
			}

			result := enc.Encode(packet)

			assert.LengthSlice(t, 1, result)
		})

		t.Run("PacketWithNamespaceAndId", func(t *testing.T) {
			packet := &Packet{
				Type: ACK,
				Nsp:  "/admin",
				Id:   new(uint64(42)),
				Data: []any{"ok"},
			}

			result := enc.Encode(packet)

			assert.LengthSlice(t, 1, result)
		})

		t.Run("PacketWithoutData", func(t *testing.T) {
			packet := &Packet{
				Type: EVENT,
				Nsp:  "/",
				Id:   new(uint64(1)),
			}

			result := enc.Encode(packet)

			assert.LengthSlice(t, 1, result)
		})

		t.Run("PacketWithAttachments", func(t *testing.T) {
			packet := &Packet{
				Type:        BINARY_EVENT,
				Nsp:         "/",
				Attachments: new(uint64(3)),
				Data:        []any{"file_upload"},
			}

			result := enc.Encode(packet)

			isGreaterThanZero := len(result) > 0
			assert.True(t, isGreaterThanZero)
		})
	})
}
