package xhttp

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// IncomingHeaders maps raw header keys to their corresponding values.
type IncomingHeaders map[string]any

// Header converts the [IncomingHeaders] into standard [net/http.Header].
func (h IncomingHeaders) Header() http.Header {
	if h == nil {
		return nil
	}
	header := make(http.Header, len(h))
	for key, value := range h {
		canonicalKey := http.CanonicalHeaderKey(key)
		if values := convertToStringSlice(value); len(values) > 0 {
			header[canonicalKey] = values
		}
	}

	return header
}

// ParsedUrlQuery maps raw query parameter keys to their corresponding values.
type ParsedUrlQuery map[string]any

// Query converts the [ParsedUrlQuery] into standard [net/url.Values].
func (q ParsedUrlQuery) Query() url.Values {
	if q == nil {
		return nil
	}
	values := make(url.Values, len(q))
	for key, value := range q {
		if vals := convertToStringSlice(value); len(vals) > 0 {
			values[key] = vals
		}
	}

	return values
}

// convertToStringSlice converts an underlying interface value into a slice of strings.
func convertToStringSlice(value any) []string {
	switch v := value.(type) {
	case string:
		return []string{v}
	case []string:
		return v
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str := convertAnyToString(item); str != "" {
				result = append(result, str)
			}
		}
		return result
	default:
		if str := convertAnyToString(v); str != "" {
			return []string{str}
		}
		return nil
	}
}

// convertAnyToString formats a generic value into its standard string representation.
func convertAnyToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		return strconv.FormatBool(v)
	case fmt.Stringer:
		return v.String()
	default:
		if v == nil {
			return ""
		}
		return fmt.Sprintf("%v", v)
	}
}
