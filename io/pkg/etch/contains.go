package etch

import "strings"

func Contains(
	haystack string,
	needles []string,
) string {
	for _, needle := range needles {
		if needle != "" && strings.Contains(haystack, needle) {
			return needle
		}
	}

	return ""
}
