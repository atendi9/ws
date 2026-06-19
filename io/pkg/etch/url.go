package etch

import (
	"fmt"
	"net"
	"net/url"

	"github.com/atendi9/ws/io/pkg/errors"
)

type ParsedUrl struct {
	*url.URL

	Hostname string
	Port     string
	Id       string
}

func Url(uri string, path string) (parsedUrl *ParsedUrl, err error) {
	if uri == "" {
		return nil, errors.ErrEmptyURI
	}

	url, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	switch url.Scheme {
	case "http", "https", "ws", "wss":
		// valid schemes
	case "":
		return nil, fmt.Errorf("%w: %q", errors.ErrUnsupportedScheme, url.Scheme)
	default:
		return nil, fmt.Errorf("%w: %q", errors.ErrUnsupportedScheme, url.Scheme)
	}

	parsedUrl = &ParsedUrl{URL: url}
	parsedUrl.Hostname = parsedUrl.URL.Hostname()
	parsedUrl.Port = parsedUrl.URL.Port()

	if parsedUrl.Port == "" {
		switch parsedUrl.Scheme {
		case "http", "ws":
			parsedUrl.Port = "80"
		case "https", "wss":
			parsedUrl.Port = "443"
		}
	}

	if parsedUrl.Path == "" {
		parsedUrl.Path = "/"
	}

	parsedUrl.Id = fmt.Sprintf("%s://%s%s", parsedUrl.Scheme, net.JoinHostPort(parsedUrl.Hostname, parsedUrl.Port), path)

	return parsedUrl, nil
}
