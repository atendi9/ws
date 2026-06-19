package xrequest

import (
	"math/rand/v2"
	"strconv"
	"strings"
	"time"
)

var (
	cookieNameSanitizer = strings.NewReplacer("\n", "-", "\r", "-")
)

// SanitizeCookieName cleans up a cookie name by replacing newline (\n)
// and carriage return (\r) characters with hyphens.
// It returns the sanitized [string].
func SanitizeCookieName(n string) string {
	return cookieNameSanitizer.Replace(n)
}

// SanitizeCookieValue validates and cleans a cookie value based on allowed bytes.
// If the value contains spaces or commas, or if the quoted [bool] parameter
// is true, the resulting [string] will be wrapped in double quotes.
func SanitizeCookieValue(v string, quoted bool) string {
	v = sanitize(validCookieValueByte, v)
	if len(v) == 0 {
		return v
	}
	if strings.ContainsAny(v, " ,") || quoted {
		return `"` + v + `"`
	}
	return v
}

// sanitize filters the input [string] v, keeping only the bytes that
// satisfy the provided valid predicate function.
func sanitize(valid func(byte) bool, v string) string {
	ok := true
	for i := 0; i < len(v); i++ {
		if valid(v[i]) {
			continue
		}
		ok = false
		break
	}
	if ok {
		return v
	}
	buf := make([]byte, 0, len(v))
	for i := 0; i < len(v); i++ {
		if b := v[i]; valid(b) {
			buf = append(buf, b)
		}
	}
	return string(buf)
}

// validCookieValueByte checks if a given [byte] is valid according to
// RFC 6265 cookie value rules (excluding quotes, semicolons, and backslashes).
func validCookieValueByte(b byte) bool {
	return 0x20 <= b && b < 0x7f && b != '"' && b != ';' && b != '\\'
}

// RandomString generates a pseudo-random alphanumeric string based on the
// current Unix timestamp in nanoseconds and a random uint64 value, both encoded in base36.
// It guarantees a safe [string] construction without slice panic risks.
func RandomString() string {
	// Convert current nanoseconds to base36 string.
	// In 2026, this typically results in an 11-character string.
	timestampStr := strconv.FormatInt(time.Now().UnixNano(), 36)
	if len(timestampStr) > 6 {
		// Safely take the last 6 characters.
		timestampStr = timestampStr[len(timestampStr)-6:]
	}

	// Convert a random uint64 to base36 string.
	randomBase36 := strconv.FormatUint(rand.Uint64(), 36)
	if len(randomBase36) > 6 {
		// Safely take the first 6 characters.
		randomBase36 = randomBase36[:6]
	} else {
		// Pad with trailing zeros if the random string is shorter than 6 characters.
		for len(randomBase36) < 6 {
			randomBase36 += "0"
		}
	}

	return timestampStr + randomBase36
}
