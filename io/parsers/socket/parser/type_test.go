package parser

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestPacketType(t *testing.T) {
	tests := []struct {
		name     string
		pType    PacketType
		expected bool
	}{
		{"CONNECT is valid", CONNECT, true},
		{"DISCONNECT is valid", DISCONNECT, true},
		{"EVENT is valid", EVENT, true},
		{"ACK is valid", ACK, true},
		{"CONNECT_ERROR is valid", CONNECT_ERROR, true},
		{"BINARY_EVENT is valid", BINARY_EVENT, true},
		{"BINARY_ACK is valid", BINARY_ACK, true},
		{"Negative value is invalid", PacketType(-1), false},
		{"Out of range value is invalid", PacketType(7), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expected {
				assert.True(t, tt.pType.Valid())
			} else {
				assert.False(t, tt.pType.Valid())
			}
		})
	}
	stringTests := []struct {
		name     string
		pType    PacketType
		expected string
	}{
		{"CONNECT string", CONNECT, "CONNECT"},
		{"DISCONNECT string", DISCONNECT, "DISCONNECT"},
		{"EVENT string", EVENT, "EVENT"},
		{"ACK string", ACK, "ACK"},
		{"CONNECT_ERROR string", CONNECT_ERROR, "CONNECT_ERROR"},
		{"BINARY_EVENT string", BINARY_EVENT, "BINARY_EVENT"},
		{"BINARY_ACK string", BINARY_ACK, "BINARY_ACK"},
		{"Unknown packet type", PacketType(99), "UNKNOWN"},
	}

	for _, tt := range stringTests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.pType.String())
		})
	}
}
