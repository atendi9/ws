package etch

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestCheckInvalidHeaderChar(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want bool
	}{
		{"plain ascii", "application/json", false},
		{"with space", "text/html; charset=utf-8", false},
		{"with tab", "value\twith\ttab", false},
		{"empty string", "", false},
		{"contains newline", "line\nbreak", true},
		{"contains carriage return", "line\rbreak", true},
		{"contains null byte", "null\x00byte", true},
		{"contains del char", "del\x7fchar", true},
		{"high byte allowed", "caf\xe9", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CheckInvalidHeaderChar(tt.val))
		})
	}
}

func TestIsLWS(t *testing.T) {
	assert.True(t, isLWS(' '))
	assert.True(t, isLWS('\t'))
	assert.False(t, isLWS('\n'))
	assert.False(t, isLWS('a'))
}

func TestIsCTL(t *testing.T) {
	assert.True(t, isCTL('\n'))
	assert.True(t, isCTL('\x00'))
	assert.True(t, isCTL(0x7f))
	assert.False(t, isCTL('a'))
	assert.False(t, isCTL(' '))
}
