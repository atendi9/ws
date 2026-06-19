package etch

import (
	"path"
	"strings"
)

func CleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)

	if pl := len(p); p[pl-1] == '/' && np != "/" {
		if pl == len(np)+1 && strings.HasPrefix(p, np) {
			np = p
		} else {
			np += "/"
		}
	}
	return np
}
