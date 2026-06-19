package etch

import (
	"net"
	"strings"
)

func Host(h string) string {
	if !strings.Contains(h, ":") {
		return h
	}
	host, _, err := net.SplitHostPort(h)
	if err != nil {
		return h
	}
	return host
}
