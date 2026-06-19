package engine

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	pkgerrors "github.com/atendi9/ws/io/pkg/errors"
	"github.com/atendi9/ws/io/pkg/xhttp"

	"github.com/atendi9/capivara/assert"
)

func newBaseServer(t *testing.T) BaseServer {
	t.Helper()
	bs := MakeBaseServer()
	bs.Construct(nil)
	return bs
}

func ctxFor(t *testing.T, method, target string) *xhttp.Context {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	return xhttp.NewContext(httptest.NewRecorder(), req)
}

func TestMakeBaseServerAndConstruct(t *testing.T) {
	bs := newBaseServer(t)
	assert.NotNil(t, bs)

	// Prototype wiring.
	assert.NotNil(t, bs.Proto())

	// Default options applied.
	opts := bs.Opts()
	assert.NotNil(t, opts)
	assert.True(t, opts.AllowUpgrades())

	// Transports registered from default options.
	assert.True(t, bs.Transports().Has(Polling.Name()))
	assert.True(t, bs.Transports().Has(WebSocket.Name()))

	byName := bs.TransportsByName()
	assert.NotNil(t, byName[Polling.Name()])
	assert.NotNil(t, byName[WebSocket.Name()])
}

func TestBaseServerClients(t *testing.T) {
	bs := newBaseServer(t)
	assert.NotNil(t, bs.Clients())
	assert.Equal(t, uint64(0), bs.ClientsCount())
}

func TestBaseServerMiddlewares(t *testing.T) {
	bs := newBaseServer(t)
	assert.Equal(t, 0, len(bs.Middlewares()))

	called := false
	bs.Use(func(ctx *xhttp.Context, next func(error)) {
		called = true
		next(nil)
	})
	assert.Equal(t, 1, len(bs.Middlewares()))

	done := false
	bs.ApplyMiddlewares(ctxFor(t, http.MethodGet, "/"), func(err error) {
		assert.NoError(t, err)
		done = true
	})
	assert.True(t, called)
	assert.True(t, done)
}

func TestBaseServerApplyMiddlewaresEmpty(t *testing.T) {
	bs := newBaseServer(t)
	done := false
	bs.ApplyMiddlewares(ctxFor(t, http.MethodGet, "/"), func(err error) {
		assert.NoError(t, err)
		done = true
	})
	assert.True(t, done)
}

func TestBaseServerApplyMiddlewaresError(t *testing.T) {
	bs := newBaseServer(t)
	bs.Use(func(ctx *xhttp.Context, next func(error)) {
		next(errors.New("boom"))
	})
	secondReached := false
	bs.Use(func(ctx *xhttp.Context, next func(error)) {
		secondReached = true
		next(nil)
	})

	var gotErr error
	bs.ApplyMiddlewares(ctxFor(t, http.MethodGet, "/"), func(err error) {
		gotErr = err
	})
	assert.Error(t, gotErr)
	assert.False(t, secondReached) // chain short-circuits on error
}

func TestBaseServerComputePath(t *testing.T) {
	bs := newBaseServer(t)
	// nil options -> default base path, trailing slash only applied when options present
	assert.Equal(t, "/engine.io", bs.ComputePath(nil))
}

func TestBaseServerUpgrades(t *testing.T) {
	bs := newBaseServer(t)
	// Polling upgrades to websocket.
	ups := bs.Upgrades(Polling.Name())
	assert.True(t, len(ups) > 0)

	// Unknown transport yields no upgrades.
	assert.True(t, len(bs.Upgrades("bogus")) == 0)
}

func TestBaseServerVerify(t *testing.T) {
	bs := newBaseServer(t)

	t.Run("unknown transport", func(t *testing.T) {
		code, payload := bs.Verify(ctxFor(t, http.MethodGet, "/?transport=bogus"), false)
		assert.Equal(t, UnknownTransport, code)
		assert.NotNil(t, payload)
	})

	t.Run("bad handshake method", func(t *testing.T) {
		code, _ := bs.Verify(ctxFor(t, http.MethodPost, "/?transport=polling"), false)
		assert.Equal(t, BadHandshakeMethod, code)
	})

	t.Run("websocket handshake without upgrade", func(t *testing.T) {
		code, _ := bs.Verify(ctxFor(t, http.MethodGet, "/?transport=websocket"), false)
		assert.Equal(t, BadRequest, code)
	})

	t.Run("invalid sid", func(t *testing.T) {
		code, _ := bs.Verify(ctxFor(t, http.MethodGet, "/?transport=polling&sid=%21%21bad"), false)
		assert.Equal(t, BadRequest, code)
	})

	t.Run("valid polling handshake passes", func(t *testing.T) {
		code, _ := bs.Verify(ctxFor(t, http.MethodGet, "/?transport=polling"), false)
		assert.True(t, code == nil)
	})
}

func TestBaseServerGenerateId(t *testing.T) {
	bs := newBaseServer(t)
	id := bs.GenerateId(ctxFor(t, http.MethodGet, "/"))
	assert.True(t, len(id) > 0)
	// A freshly generated id from two calls must differ.
	assert.True(t, id != bs.GenerateId(ctxFor(t, http.MethodGet, "/")))
}

func TestBaseServerCreateTransportStub(t *testing.T) {
	bs := newBaseServer(t)
	_, err := bs.CreateTransport(ctxFor(t, http.MethodGet, "/"), "polling")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, pkgerrors.ErrTransportNotImplemented))
}

func TestBaseServerCloseNoClients(t *testing.T) {
	bs := newBaseServer(t)
	assert.NotNil(t, bs.Close())
}

func TestBaseServerInitCleanupNoop(t *testing.T) {
	bs := newBaseServer(t)
	bs.Init()
	bs.Cleanup()
}
