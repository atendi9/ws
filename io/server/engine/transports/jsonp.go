package transports

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/atendi9/ws/io/parsers/engine/packet"
	"github.com/atendi9/ws/io/pkg/errors"
	"github.com/atendi9/ws/io/pkg/forge"
	"github.com/atendi9/ws/io/pkg/xhttp"
)

var (
	// rNumber is a [regexp.Regexp] used to sanitize the JSONP callback index.
	rNumber = regexp.MustCompile(`[^0-9]`)
	// jsonpReplacer is a [strings.Replacer] used to escape newlines in JSONP payloads.
	jsonpReplacer = strings.NewReplacer(`\\\\n`, `\n`, `\\n`, "\n")
)

// Jsonp is an alias for [Polling] that defines the JSONP transport interface.
type Jsonp = Polling

// jsonp implements the [Jsonp] interface to handle JSONP long-polling connections.
// It extends the base [Polling] transport.
type jsonp struct {
	Polling

	head string
	foot string
}

// MakeJSONP creates and returns a new uninitialized [Jsonp] instance.
func MakeJSONP() Jsonp {
	j := &jsonp{Polling: MakePolling()}

	j.Prototype(j)

	return j
}

// NewJSONP creates and initializes a new [Jsonp] instance with the given [xhttp.Context].
func NewJSONP(ctx *xhttp.Context) Jsonp {
	j := MakeJSONP()

	j.Construct(ctx)

	return j
}

// Construct initializes the [jsonp] instance with the given [xhttp.Context] and sets up the JSONP response wrapper.
func (j *jsonp) Construct(ctx *xhttp.Context) {
	j.Polling.Construct(ctx)

	j.head = "___eio[" + rNumber.ReplaceAllString(ctx.Query().Peek("j"), "") + "]("
	j.foot = ");"
}

// OnData decodes the incoming JSONP payload from the given [forge.Interface] and delegates processing to the underlying [Polling] transport.
func (j *jsonp) OnData(data forge.Interface) {
	payload, err := url.ParseQuery(data.String())
	if err != nil {
		j.OnError("jsonp read error", err)
		return
	}

	if payload.Has("d") {
		j.Polling.OnData(forge.NewFromString(jsonpReplacer.Replace(payload.Get("d"))))
	}
}

// DoWrite executes the actual HTTP response write, wrapping the [forge.Interface] data in the JSONP callback format for the given [xhttp.Context] and [packet.Options].
func (j *jsonp) DoWrite(ctx *xhttp.Context, data forge.Interface, options *packet.Options, callback func(error)) {
	payload, err := json.Marshal(data.String())
	if err != nil {
		ctx.Cleanup()
		defer callback(err)

		_ = ctx.SetStatusCode(http.StatusInternalServerError)
		_, _ = ctx.Write(nil)
		return
	}
	if len(payload) > forge.MaxPayloadSize {
		ctx.Cleanup()
		defer callback(errors.ErrPayloadTooLarge)

		_ = ctx.SetStatusCode(http.StatusInternalServerError)
		_, _ = ctx.Write(nil)
		return
	}

	res := forge.NewFromString(j.head)
	_, _ = res.Write(payload)
	_, _ = res.WriteString(j.foot)
	j.Polling.DoWrite(ctx, res, options, callback)
}
