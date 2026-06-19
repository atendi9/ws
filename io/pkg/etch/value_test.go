package etch

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		_default string
		want     string
	}{
		{"returns value when set", "real", "fallback", "real"},
		{"returns default when empty", "", "fallback", "fallback"},
		{"both empty", "", "", ""},
		{"whitespace is a value", " ", "fallback", " "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Value(tt.value, tt._default))
		})
	}
}
