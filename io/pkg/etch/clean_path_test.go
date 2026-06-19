package etch

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestCleanPath(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty becomes root", "", "/"},
		{"adds leading slash", "foo/bar", "/foo/bar"},
		{"already absolute", "/foo/bar", "/foo/bar"},
		{"collapses dot", "/foo/./bar", "/foo/bar"},
		{"collapses dotdot", "/foo/../bar", "/bar"},
		{"collapses double slash", "/foo//bar", "/foo/bar"},
		{"root stays root", "/", "/"},
		{"keeps trailing slash", "/foo/bar/", "/foo/bar/"},
		{"trailing slash after clean", "/foo/./bar/", "/foo/bar/"},
		{"relative with trailing slash", "foo/", "/foo/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CleanPath(tt.in))
		})
	}
}
