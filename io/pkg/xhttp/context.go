package xhttp

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/atendi9/box"
	"github.com/atendi9/ws/io/pkg/errors"
	"github.com/atendi9/ws/io/pkg/events"
	"github.com/atendi9/ws/io/pkg/websocket"
)

// HTTP status code boundaries as per RFC 9110.
const (
	minStatusCode = 100
	maxStatusCode = 599
)

type Callable = func()

// Context wraps an http.Request / http.ResponseWriter pair with extra
// features: event emission, lazy-computed request metadata, one-shot writing
// semantics, and a done-channel tied to the request context.
//
// Instances must be created with NewContext. A single Context is not
// meant to be written to more than once; subsequent writes return
// ErrResponseAlreadyWritten.
type Context struct {
	noCopy noCopy

	events.Emitter

	IdleTimeout time.Duration
	// Optional protocol upgrades. Set by the caller when applicable.
	Websocket *websocket.Conn

	// Cleanup is invoked exactly once when the context is closed.
	Cleanup Callable

	ctx      context.Context
	request  *http.Request
	response http.ResponseWriter

	statusCode box.Atomic[int]

	// written indicates that the response body has been (or is being) written.
	written box.Atomic[bool]
	done    chan struct{}

	closeOnce sync.Once
	writeOnce sync.Once

	// responseHeadersUsed tracks whether ResponseHeaders() was ever called.
	// When false, flushResponseHeaders can skip the redundant copy.
	responseHeadersUsed atomic.Bool

	// Lazily-computed, cached request/response metadata.
	onceHeaders         func() *ParameterBag
	onceQuery           func() *ParameterBag
	onceResponseHeaders func() *ParameterBag
	onceMethod          func() string
	onceHost            func() string
	oncePath            func() string
	onceUserAgent       func() string
}

// NewContext creates a fully-initialized Context. It panics if either
// w or r is nil since that indicates a programming error at the caller site.
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	if r == nil {
		panic(errors.ErrHTTPContextNilRequest)
	}
	if w == nil {
		panic(errors.ErrHTTPContextNilResponseWriter)
	}

	c := &Context{
		Emitter:    events.NewEmitter(),
		request:    r,
		response:   w,
		ctx:        r.Context(),
		done:       make(chan struct{}),
		written:    box.Atomic[bool]{},
		statusCode: box.Atomic[int]{},
	}

	c.initLazyAccessors()
	c.statusCode.Store(http.StatusOK)

	go c.contextWatcher()

	return c
}

// initLazyAccessors sets up OnceValue-backed accessors for commonly used
// request/response metadata. Centralizing this keeps the constructor short
// and easy to extend.
func (c *Context) initLazyAccessors() {
	r, w := c.request, c.response

	c.onceHeaders = sync.OnceValue(func() *ParameterBag {
		return NewParameterBag(r.Header)
	})
	c.onceQuery = sync.OnceValue(func() *ParameterBag {
		return NewParameterBag(r.URL.Query())
	})
	c.onceResponseHeaders = sync.OnceValue(func() *ParameterBag {
		return NewParameterBag(w.Header())
	})
	c.onceMethod = sync.OnceValue(func() string {
		return strings.ToUpper(r.Method)
	})
	c.onceHost = sync.OnceValue(func() string {
		host := strings.TrimSpace(r.Host)
		if h, _, err := net.SplitHostPort(host); err == nil {
			return h
		}
		return host
	})
	c.oncePath = sync.OnceValue(func() string {
		p := strings.Trim(r.URL.Path, "/")
		if p == "" {
			return "/"
		}
		return p
	})
	c.onceUserAgent = sync.OnceValue(func() string {
		return r.Header.Get("User-Agent")
	})
}

// IsDone reports whether the response body has been written or the context
// has been closed.
func (c *Context) IsDone() bool {
	return c.written.Load()
}

// Done returns a channel that is closed when the context is finalized.
func (c *Context) Done() <-chan struct{} {
	return c.done
}

// Flush finalizes the context without writing a response body.
func (c *Context) Flush() {
	c.closeWithError(nil)
}

// SetStatusCode sets the HTTP status code to be used by the next Write.
func (c *Context) SetStatusCode(code int) error {
	if code < minStatusCode || code > maxStatusCode {
		return errors.ErrHTTPContextInvalidStatusCode
	}
	if c.written.Load() {
		return errors.ErrHTTPContextResponseAlreadyWritten
	}
	c.statusCode.Store(code)
	return nil
}

// GetStatusCode returns the currently configured status code (default 200).
func (c *Context) GetStatusCode() int {
	return int(c.statusCode.Load())
}

// Write commits the response.
func (c *Context) Write(data []byte) (n int, err error) {
	if c.written.Load() {
		return 0, errors.ErrHTTPContextResponseAlreadyWritten
	}

	executed := false
	c.writeOnce.Do(func() {
		executed = true
		c.written.Store(true)

		n, err = c.performWrite(data)
		c.closeWithError(nil)
	})

	if !executed {
		return 0, errors.ErrHTTPContextResponseAlreadyWritten
	}
	return n, err
}

// performWrite flushes headers and body to the underlying ResponseWriter.
func (c *Context) performWrite(data []byte) (int, error) {
	c.flushResponseHeaders()
	c.response.WriteHeader(c.GetStatusCode())
	return c.response.Write(data)
}

// flushResponseHeaders copies any headers staged in the ParameterBag into the
// underlying http.Header. It is a no-op when the bag was never materialized.
func (c *Context) flushResponseHeaders() {
	if !c.responseHeadersUsed.Load() {
		return
	}
	dst := c.response.Header()
	for key, values := range c.onceResponseHeaders().All() {
		switch len(values) {
		case 0:
			continue
		case 1:
			dst.Set(key, values[0])
		default:
			dst.Del(key)
			for _, v := range values {
				dst.Add(key, v)
			}
		}
	}
}

func (c *Context) Request() *http.Request        { return c.request }
func (c *Context) Response() http.ResponseWriter { return c.response }
func (c *Context) Context() context.Context      { return c.ctx }
func (c *Context) Headers() *ParameterBag        { return c.onceHeaders() }
func (c *Context) Query() *ParameterBag          { return c.onceQuery() }
func (c *Context) ResponseHeaders() *ParameterBag {
	c.responseHeadersUsed.Store(true)
	return c.onceResponseHeaders()
}
func (c *Context) Method() string    { return c.onceMethod() }
func (c *Context) Host() string      { return c.onceHost() }
func (c *Context) Path() string      { return c.oncePath() }
func (c *Context) UserAgent() string { return c.onceUserAgent() }
func (c *Context) PathInfo() string  { return c.request.URL.Path }
func (c *Context) Secure() bool      { return c.request.TLS != nil }

// contextWatcher closes the context when the underlying request context is
// canceled. It exits once either signal is observed.
func (c *Context) contextWatcher() {
	select {
	case <-c.ctx.Done():
		c.closeWithError(c.ctx.Err())
	case <-c.done:
	}
}

// closeWithError finalizes the context exactly once, running Cleanup and
// emitting a "close" event asynchronously so it never blocks the caller.
func (c *Context) closeWithError(err error) {
	c.closeOnce.Do(func() {
		// Mark written to make IsDone reflect finalization as well,
		// even for paths that never wrote a response body.
		c.written.Store(true)

		close(c.done)
		if c.Cleanup != nil {
			c.Cleanup()
		}
		// Fire the event asynchronously so listeners can't deadlock us.
		// Clear the emitter afterwards to release listener references for GC.
		go func() {
			c.Emit("close", err)
			c.Clear()
		}()
	})
}
