package etch

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestHost(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no port returned as-is", "example.com", "example.com"},
		{"strips port", "example.com:8080", "example.com"},
		{"ipv4 with port", "127.0.0.1:80", "127.0.0.1"},
		{"ipv6 with port", "[::1]:8080", "::1"},
		{"empty string", "", ""},
		{"invalid host port returned as-is", "example.com:80:90", "example.com:80:90"},
		{"trailing colon no port", "example.com:", "example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Host(tt.in))
		})
	}
}
