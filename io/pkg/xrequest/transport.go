package xrequest

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	maxRetryAttempts int32 = 2
)

// Transport implements an HTTP transport that supports standard HTTP/HTTPS configurations
// and custom Alt-Svc tracking features using standard libraries.
type Transport struct {
	standardTransport *http.Transport
	altSvcCache       sync.Map
}

type altSvc struct {
	protocol string
	endpoint string
	expires  time.Time
	failures atomic.Int32
	persist  bool
}

// NewTransport creates a new Transport instance.
func NewTransport(tlsClientConfig *tls.Config) *Transport {
	return &Transport{
		standardTransport: &http.Transport{
			TLSClientConfig: tlsClientConfig,
		},
	}
}

// RoundTrip executes a single HTTP transaction.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if resp, err := t.tryAltServices(req); err == nil {
		return resp, nil
	}

	resp, err := t.standardTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	t.processAltSvc(resp.Header, req.URL)
	return resp, nil
}

func (t *Transport) tryAltServices(req *http.Request) (*http.Response, error) {
	val, ok := t.altSvcCache.Load(getOrigin(req.URL))
	if !ok {
		return nil, errors.New("no alt-svc cached")
	}

	services, ok := val.([]*altSvc)
	if !ok {
		return nil, errors.New("invalid cache entry")
	}

	for _, svc := range services {
		if !isServiceValid(svc) {
			continue
		}

		if resp, err := t.tryService(req, svc); err == nil {
			return resp, nil
		}
	}

	return nil, errors.New("alternative services failed")
}

func isServiceValid(svc *altSvc) bool {
	return !svc.expires.Before(time.Now()) && svc.failures.Load() < maxRetryAttempts
}

func (t *Transport) tryService(req *http.Request, svc *altSvc) (*http.Response, error) {
	altReq := req.Clone(req.Context())

	if endpoint := svc.endpoint; endpoint != "" {
		if after, ok := strings.CutPrefix(endpoint, ":"); ok {
			endpoint = net.JoinHostPort(req.URL.Hostname(), after)
		} else {
			host, _, err := net.SplitHostPort(endpoint)
			if err != nil {
				return nil, errors.New("invalid endpoint format")
			}
			if !strings.EqualFold(host, req.URL.Hostname()) {
				return nil, errors.New("origin mismatch")
			}
			if strings.ContainsAny(endpoint, "/?#") {
				return nil, errors.New("endpoint contains invalid characters")
			}
		}
		altReq.URL.Host = endpoint
	}

	resp, err := t.standardTransport.RoundTrip(altReq)
	if err != nil {
		svc.failures.Add(1)
	}
	return resp, err
}

func (t *Transport) processAltSvc(header http.Header, reqURL *url.URL) {
	altSvcHeader := header.Get("Alt-Svc")
	if altSvcHeader == "" {
		return
	}
	origin := getOrigin(reqURL)

	if altSvcHeader == "clear" {
		t.altSvcCache.Delete(origin)
		return
	}

	entries := parseAltSvc(altSvcHeader)
	if len(entries) > 0 {
		t.altSvcCache.Store(origin, entries)
	}
}

func parseAltSvc(value string) []*altSvc {
	var result []*altSvc
	now := time.Now()

	entries := strings.SplitSeq(value, ",")
	for entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}

		protocol := strings.TrimSpace(parts[0])
		params := strings.TrimSpace(parts[1])

		maxAge := int64(24 * 3600)
		persist := false

		paramParts := strings.Split(params, ";")
		endpoint := strings.Trim(strings.TrimSpace(paramParts[0]), `"`)

		for _, param := range paramParts[1:] {
			param = strings.TrimSpace(param)
			if param == "" {
				continue
			}

			kv := strings.SplitN(param, "=", 2)
			if len(kv) != 2 {
				continue
			}

			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])

			switch key {
			case "ma":
				if age, err := strconv.ParseInt(val, 10, 64); err == nil {
					const maxMaxAge int64 = 365 * 24 * 3600
					if age < 0 {
						age = 0
					} else if age > maxMaxAge {
						age = maxMaxAge
					}
					maxAge = age
				}
			case "persist":
				persist = val == "1"
			}
		}

		result = append(result, &altSvc{
			protocol: protocol,
			endpoint: endpoint,
			expires:  now.Add(time.Duration(maxAge) * time.Second),
			persist:  persist,
		})
	}

	return result
}

// Close closes idle connections on the standard transport.
func (t *Transport) Close() error {
	t.standardTransport.CloseIdleConnections()
	return nil
}

func getOrigin(u *url.URL) string {
	if u.Port() == "" {
		return net.JoinHostPort(u.Hostname(), defaultPort(u.Scheme))
	}
	return u.Host
}

func defaultPort(scheme string) string {
	if scheme == "https" {
		return "443"
	}
	return "80"
}
