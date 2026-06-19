package xhttp

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/atendi9/capivara/assert"
)

// TestIncomingHeaders_Header tests the conversion of IncomingHeaders to http.Header.
func TestIncomingHeaders_Header(t *testing.T) {
	t.Run("nil map returns nil", func(t *testing.T) {
		var h IncomingHeaders
		assert.Equal(t, fmt.Sprint(http.Header(nil)), fmt.Sprint(h.Header()))
	})

	t.Run("valid map converts types correctly", func(t *testing.T) {
		h := IncomingHeaders{
			"Content-Type": "application/json",
			"X-Rate-Limit": 100,
			"X-Active":     true,
			"Tags":         []any{"api", "v1"},
		}
		result := h.Header()

		assert.Equal(t, fmt.Sprint([]string{"application/json"}), fmt.Sprint(result["Content-Type"]))
		assert.Equal(t, fmt.Sprint([]string{"100"}), fmt.Sprint(result["X-Rate-Limit"]))
		assert.Equal(t, fmt.Sprint([]string{"true"}), fmt.Sprint(result["X-Active"]))
		assert.Equal(t, fmt.Sprint([]string{"api", "v1"}), fmt.Sprint(result["Tags"]))
	})
}

// TestParsedUrlQuery_Query tests the conversion of ParsedUrlQuery to url.Values.
func TestParsedUrlQuery_Query(t *testing.T) {
	t.Run("nil map returns nil", func(t *testing.T) {
		var q ParsedUrlQuery
		assert.Equal(t, fmt.Sprint(url.Values(nil)), fmt.Sprint(q.Query()))
	})

	t.Run("valid map converts types correctly", func(t *testing.T) {
		q := ParsedUrlQuery{
			"page":  1,
			"order": "desc",
			"items": []string{"a", "b"},
		}
		result := q.Query()

		assert.Equal(t, fmt.Sprint([]string{"1"}), fmt.Sprint(result["page"]))
		assert.Equal(t, fmt.Sprint([]string{"desc"}), fmt.Sprint(result["order"]))
		assert.Equal(t, fmt.Sprint([]string{"a", "b"}), fmt.Sprint(result["items"]))
	})
}

// TestConvertToStringSlice tests the utility function convertToStringSlice.
func TestConvertToStringSlice(t *testing.T) {
	t.Run("converts single string", func(t *testing.T) {
		result := convertToStringSlice("test")
		assert.Equal(t, fmt.Sprint([]string{"test"}), fmt.Sprint(result))
	})

	t.Run("converts string slice", func(t *testing.T) {
		input := []string{"a", "b"}
		result := convertToStringSlice(input)
		assert.Equal(t, fmt.Sprint([]string{"a", "b"}), fmt.Sprint(result))
	})

	t.Run("converts any slice", func(t *testing.T) {
		input := []any{1, "string", true}
		result := convertToStringSlice(input)
		assert.Equal(t, fmt.Sprint([]string{"1", "string", "true"}), fmt.Sprint(result))
	})

	t.Run("converts integer", func(t *testing.T) {
		result := convertToStringSlice(42)
		assert.Equal(t, fmt.Sprint([]string{"42"}), fmt.Sprint(result))
	})
}

// TestConvertAnyToString tests the utility function convertAnyToString.
func TestConvertAnyToString(t *testing.T) {
	t.Run("converts various types", func(t *testing.T) {
		assert.Equal(t, "hello", convertAnyToString("hello"))
		assert.Equal(t, "100", convertAnyToString(100))
		assert.Equal(t, "3.14", convertAnyToString(3.14))
		assert.Equal(t, "true", convertAnyToString(true))
	})

	t.Run("handles nil", func(t *testing.T) {
		assert.Equal(t, "", convertAnyToString(nil))
	})
}
