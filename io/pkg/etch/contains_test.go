package etch

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		haystack string
		needles  []string
		want     string
	}{
		{"returns matching needle", "hello world", []string{"world"}, "world"},
		{"returns first match", "hello world", []string{"foo", "lo", "world"}, "lo"},
		{"no match returns empty", "hello", []string{"foo", "bar"}, ""},
		{"empty needle is skipped", "hello", []string{"", "ell"}, "ell"},
		{"only empty needles", "hello", []string{"", ""}, ""},
		{"nil needles", "hello", nil, ""},
		{"empty haystack", "", []string{"x"}, ""},
		{"empty haystack empty needle", "", []string{""}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Contains(tt.haystack, tt.needles))
		})
	}
}
