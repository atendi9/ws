package xrequest

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestSanitizeCookieName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "session", "session"},
		{"newline replaced", "ses\nsion", "ses-sion"},
		{"carriage return replaced", "ses\rsion", "ses-sion"},
		{"both replaced", "a\r\nb", "a--b"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SanitizeCookieName(tt.in))
		})
	}
}

func TestSanitizeCookieValue(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		quoted bool
		want   string
	}{
		{"simple", "abc123", false, "abc123"},
		{"empty stays empty", "", false, ""},
		{"forced quotes", "abc", true, `"abc"`},
		{"space triggers quotes", "a b", false, `"a b"`},
		{"comma triggers quotes", "a,b", false, `"a,b"`},
		{"strips invalid bytes", "a\x00b\x7f", false, "ab"},
		{"strips quote semicolon backslash", `a"b;c\d`, false, "abcd"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SanitizeCookieValue(tt.in, tt.quoted))
		})
	}
}

func TestValidCookieValueByte(t *testing.T) {
	assert.True(t, validCookieValueByte('a'))
	assert.True(t, validCookieValueByte(' '))
	assert.False(t, validCookieValueByte('"'))
	assert.False(t, validCookieValueByte(';'))
	assert.False(t, validCookieValueByte('\\'))
	assert.False(t, validCookieValueByte(0x7f))
	assert.False(t, validCookieValueByte(0x1f))
}

func TestSanitize(t *testing.T) {
	// all valid -> returns input unchanged
	assert.Equal(t, "abc", sanitize(validCookieValueByte, "abc"))
	// invalid bytes filtered out
	assert.Equal(t, "abc", sanitize(validCookieValueByte, "a;b\\c"))
}

func TestRandomString(t *testing.T) {
	s := RandomString()
	// timestamp (<=6) + random (exactly 6) -> always 12 chars in practice
	assert.Equal(t, 12, len(s))

	// two consecutive calls should not collide
	assert.True(t, s != RandomString())
}
