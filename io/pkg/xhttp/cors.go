package xhttp

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/atendi9/ws/io/pkg/anvil"
)

// Cors defines the configuration options for Cross-Origin Resource Sharing.
// It allows customization of origin validation, allowed methods, headers, and security credentials.
type Cors struct {
	// Origin defines the allowed origins. Supported types: string, []string, []any, *regexp.Regexp, bool, func(string) bool.
	Origin any
	// Methods defines the allowed HTTP methods. Supported types: string, []string.
	Methods any
	// AllowedHeaders defines the headers allowed. Supported types: nil, string, []string.
	AllowedHeaders any
	// Headers defines additional headers. Supported types: nil, string, []string.
	Headers any
	// ExposedHeaders defines which headers are exposed to the browser. Supported types: string, []string.
	ExposedHeaders any
	// MaxAge defines the duration the CORS preflight response can be cached.
	MaxAge string
	// Credentials indicates whether the request can include user credentials.
	Credentials bool
	// PreflightContinue indicates whether to pass the preflight request to the next middleware.
	PreflightContinue bool
	// OptionsSuccessStatus defines the HTTP status code to return for successful preflight requests.
	OptionsSuccessStatus int
}

// Kv represents a key-value pair used for HTTP headers.
type Kv struct {
	Key   string
	Value string
}

// cors is an internal structure used to manage the state of the CORS response during middleware execution.
type cors struct {
	options *Cors
	ctx     *Context
	headers []*Kv
	varys   []string
}

// IsOriginAllowed evaluates if the provided origin string is permitted based on the [Cors] configuration.
func (c *Cors) IsOriginAllowed(origin string, allowedOrigin any) bool {
	switch v := allowedOrigin.(type) {
	case []any:
		for _, value := range v {
			if c.IsOriginAllowed(origin, value) {
				return true
			}
		}
	case []string:
		for _, value := range v {
			if strings.EqualFold(origin, value) {
				return true
			}
		}
	case func(string) bool:
		return v(origin)
	case string:
		if v == "*" {
			return true
		}
		return strings.EqualFold(origin, v)
	case *regexp.Regexp:
		return v.MatchString(origin)
	case bool:
		return v
	}
	return false
}

// configureOrigin sets the Access-Control-Allow-Origin header based on the [Cors] configuration.
func (c *cors) configureOrigin() *cors {
	requestOrigin := c.ctx.Headers().Peek("Origin")
	// Requests without an Origin header (same-origin or non-browser clients)
	// are not subject to CORS; skip origin configuration entirely.
	if requestOrigin == "" {
		c.varys = append(c.varys, "Origin")
		return c
	}
	if o, ok := c.options.Origin.(string); ok {
		if o == "*" && c.options.Credentials {
			// Credentials + wildcard origin: must reflect the specific origin
			// per the CORS specification (Access-Control-Allow-Origin: * is
			// incompatible with Access-Control-Allow-Credentials: true).
			c.headers = append(c.headers, &Kv{
				Key:   "Access-Control-Allow-Origin",
				Value: requestOrigin,
			})
			c.varys = append(c.varys, "Origin")
		} else if o == "*" {
			// allow any origin (no credentials)
			c.headers = append(c.headers, &Kv{
				Key:   "Access-Control-Allow-Origin",
				Value: "*",
			})
		} else {
			// fixed origin — only reflect if request origin matches;
			// if not, omit Access-Control-Allow-Origin entirely so the
			// browser blocks the request without leaking the allowed origin.
			if strings.EqualFold(requestOrigin, o) {
				c.headers = append(c.headers, &Kv{
					Key:   "Access-Control-Allow-Origin",
					Value: o,
				})
			}
			c.varys = append(c.varys, "Origin")
		}
	} else {
		// reflect origin — only set the header when the origin is allowed;
		// omitting it causes the browser to block the cross-origin request.
		if c.options.IsOriginAllowed(requestOrigin, c.options.Origin) {
			c.headers = append(c.headers, &Kv{
				Key:   "Access-Control-Allow-Origin",
				Value: requestOrigin,
			})
		}
		c.varys = append(c.varys, "Origin")
	}
	return c
}

// configureMethods sets the Access-Control-Allow-Methods header.
func (c *cors) configureMethods() *cors {
	switch methods := c.options.Methods.(type) {
	case string:
		c.headers = append(c.headers, &Kv{
			Key:   "Access-Control-Allow-Methods",
			Value: methods,
		})
	case []string:
		c.headers = append(c.headers, &Kv{
			Key:   "Access-Control-Allow-Methods",
			Value: strings.Join(methods, ","),
		})
	}
	return c
}

// configureCredentials sets the Access-Control-Allow-Credentials header if required by [Cors].
func (c *cors) configureCredentials() *cors {
	if c.options.Credentials {
		c.headers = append(c.headers, &Kv{
			Key:   "Access-Control-Allow-Credentials",
			Value: "true",
		})
	}
	return c
}

// configureAllowedHeaders sets the Access-Control-Allow-Headers header.
func (c *cors) configureAllowedHeaders() *cors {
	allowedHeaders := c.options.AllowedHeaders
	if allowedHeaders == nil {
		allowedHeaders = c.options.Headers
	}

	switch h := allowedHeaders.(type) {
	case nil:
		// .c.headers wasn't specified, so reflect the request c.headers
		if head := c.ctx.Headers().Peek("Access-Control-Request-Headers"); head != "" {
			c.headers = append(c.headers, &Kv{
				Key:   "Access-Control-Allow-Headers",
				Value: head,
			})
			c.varys = append(c.varys, "Access-Control-Request-Headers")
		}
	case string:
		if len(h) > 0 {
			c.headers = append(c.headers, &Kv{
				Key:   "Access-Control-Allow-Headers",
				Value: h,
			})
		}
	case []string:
		if len(h) > 0 {
			c.headers = append(c.headers, &Kv{
				Key:   "Access-Control-Allow-Headers",
				Value: strings.Join(h, ","),
			})
		}
	}
	return c
}

// configureExposedHeaders sets the Access-Control-Expose-Headers header.
func (c *cors) configureExposedHeaders() *cors {
	switch headers := c.options.ExposedHeaders.(type) {
	case string:
		if len(headers) > 0 {
			c.headers = append(c.headers, &Kv{
				Key:   "Access-Control-Expose-Headers",
				Value: headers,
			})
		}
	case []string:
		if len(headers) > 0 {
			c.headers = append(c.headers, &Kv{
				Key:   "Access-Control-Expose-Headers",
				Value: strings.Join(headers, ","),
			})
		}
	}
	return c
}

// configureMaxAge sets the Access-Control-Max-Age header.
func (c *cors) configureMaxAge() *cors {
	if c.options.MaxAge != "" {
		c.headers = append(c.headers, &Kv{
			Key:   "Access-Control-Max-Age",
			Value: c.options.MaxAge,
		})
	}
	return c
}

// parseVary parses a Vary header string and returns a [anvil.Set] containing the vary tokens.
func parseVary(vary string) *anvil.Set[string] {
	end := 0
	start := 0
	list := anvil.NewSet[string]()

	// gather tokens
	for i, l := 0, len(vary); i < l; i++ {
		switch vary[i] {
		case ' ': /*   */
			if start == end {
				end = i + 1
				start = end
			}
		case ',': /* , */
			list.Add(vary[start:end])
			end = i + 1
			start = end
		default:
			end = i + 1
		}
	}

	if end := vary[start:end]; len(end) > 0 {
		// final token
		list.Add(end)
	}

	return list
}

// applyHeaders applies all configured CORS headers to the [Context] response.
func (c *cors) applyHeaders() {
	for _, header := range c.headers {
		c.ctx.ResponseHeaders().Set(header.Key, header.Value)
	}
	if vary := c.ctx.ResponseHeaders().Peek("Vary"); vary == "*" {
		c.ctx.ResponseHeaders().Set("Vary", "*")
	} else {
		if len(c.varys) > 0 {
			varys := parseVary(vary)
			varys.Add(c.varys...)
			c.ctx.ResponseHeaders().Set("Vary", strings.Join(varys.Keys(), ", "))
		}
	}
}

// CorsMiddleware implements the CORS middleware logic for the given [Context].
// It handles both preflight requests and actual responses.
func CorsMiddleware(ctx *Context, options *Cors, next func(error)) {
	if options == nil {
		options = defaultCors
	}
	c := &cors{
		options: options,
		ctx:     ctx,
		headers: []*Kv{},
	}
	method := c.ctx.Method()

	if http.MethodOptions == method {
		// preflight — Expose-Headers is only meaningful for actual responses,
		// so it is intentionally omitted here.
		c.configureOrigin().configureCredentials().configureMethods().configureAllowedHeaders().configureMaxAge().applyHeaders()
		if options.PreflightContinue {
			next(nil)
		} else {
			// Safari (and potentially other browsers) need content-length 0,
			// for 204 or they just hang waiting for a body
			ctx.ResponseHeaders().Set("Content-Length", "0")
			_ = ctx.SetStatusCode(options.OptionsSuccessStatus)
			_, _ = ctx.Write(nil)
		}
	} else {
		// actual response
		c.configureOrigin().configureCredentials().configureExposedHeaders().applyHeaders()
		next(nil)
	}
}

// defaultCors holds the default configuration for CORS.
var defaultCors = &Cors{
	Origin:               `*`,
	Methods:              `GET,HEAD,PUT,PATCH,POST,DELETE`,
	PreflightContinue:    false,
	OptionsSuccessStatus: 204,
}

// MiddlewareWrapper returns a functional middleware for the [Cors] configuration.
// It initializes defaults if the provided options are nil.
func MiddlewareWrapper(options *Cors) func(*Context, func(error)) {
	if options == nil {
		options = defaultCors
	} else {
		if options.Origin == nil {
			options.Origin = "*"
		}

		if options.Methods == nil {
			options.Methods = `GET,HEAD,PUT,PATCH,POST,DELETE`
		}

		if options.OptionsSuccessStatus == 0 {
			options.OptionsSuccessStatus = http.StatusNoContent
		}
	}

	return func(ctx *Context, next func(error)) {
		CorsMiddleware(ctx, options, next)
	}
}
