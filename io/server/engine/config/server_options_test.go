package config

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/ws/io/pkg/anvil"
	"github.com/atendi9/ws/io/pkg/xhttp"
	"github.com/atendi9/ws/io/server/engine/transports"
)

// TestDefaultServerOptions tests the initialization of default server options.
func TestDefaultServerOptions(t *testing.T) {
	opts := DefaultServerOptions()

	assert.Equal(t, time.Duration(0), opts.PingTimeout())
	assert.Equal(t, time.Duration(0), opts.PingInterval())
	assert.Equal(t, time.Duration(0), opts.UpgradeTimeout())
	assert.Equal(t, int64(0), opts.MaxHttpBufferSize())
	assert.False(t, opts.AllowUpgrades())
	assert.False(t, opts.AllowEIO3())
	assert.Equal(t, time.Duration(0), opts.IdleTimeout())
}

// TestServerOpts_Assign tests the Assign method for merging configurations.
func TestServerOpts_Assign(t *testing.T) {
	opts := DefaultServerOptions()
	override := DefaultServerOptions()

	expectedTimeout := 5 * time.Second
	override.SetPingTimeout(expectedTimeout)
	override.SetAllowUpgrades(true)

	opts.Assign(override)

	assert.Equal(t, expectedTimeout, opts.PingTimeout())
	assert.True(t, opts.AllowUpgrades())

	// Test assign with nil data
	opts.Assign(nil)
	assert.Equal(t, expectedTimeout, opts.PingTimeout())
	assert.True(t, opts.AllowUpgrades())
}

// TestServerOpts_PingTimeout tests the setter and getter for PingTimeout.
func TestServerOpts_PingTimeout(t *testing.T) {
	opts := DefaultServerOptions()
	expected := 10 * time.Second

	opts.SetPingTimeout(expected)

	assert.Equal(t, expected, opts.PingTimeout())
	assert.True(t, opts.GetRawPingTimeout() != nil)
}

// TestServerOpts_PingInterval tests the setter and getter for PingInterval.
func TestServerOpts_PingInterval(t *testing.T) {
	opts := DefaultServerOptions()
	expected := 25 * time.Millisecond

	opts.SetPingInterval(expected)

	assert.Equal(t, expected, opts.PingInterval())
	assert.True(t, opts.GetRawPingInterval() != nil)
}

// TestServerOpts_UpgradeTimeout tests the setter and getter for UpgradeTimeout.
func TestServerOpts_UpgradeTimeout(t *testing.T) {
	opts := DefaultServerOptions()
	expected := 10 * time.Second

	opts.SetUpgradeTimeout(expected)

	assert.Equal(t, expected, opts.UpgradeTimeout())
	assert.True(t, opts.GetRawUpgradeTimeout() != nil)
}

// TestServerOpts_MaxHttpBufferSize tests the setter and getter for MaxHttpBufferSize.
func TestServerOpts_MaxHttpBufferSize(t *testing.T) {
	opts := DefaultServerOptions()
	var expected int64 = 1048576

	opts.SetMaxHttpBufferSize(expected)

	assert.Equal(t, expected, opts.MaxHttpBufferSize())
	assert.True(t, opts.GetRawMaxHttpBufferSize() != nil)
}

// TestServerOpts_AllowRequest tests the setter and getter for AllowRequest.
func TestServerOpts_AllowRequest(t *testing.T) {
	opts := DefaultServerOptions()

	assert.True(t, opts.AllowRequest() == nil)

	expectedFunc := func(ctx *xhttp.Context) error {
		return nil
	}

	opts.SetAllowRequest(expectedFunc)

	assert.True(t, opts.AllowRequest() != nil)
	assert.True(t, opts.GetRawAllowRequest() != nil)
}

// TestServerOpts_Transports tests the setter and getter for Transports.
func TestServerOpts_Transports(t *testing.T) {
	opts := DefaultServerOptions()

	// Default state
	var emptyTransports *anvil.Set[transports.TConstructor]
	assert.Equal(t, emptyTransports, opts.Transports())

	expected := anvil.NewSet[transports.TConstructor]()
	opts.SetTransports(expected)

	assert.Equal(t, expected, opts.Transports())
	assert.True(t, opts.GetRawTransports() != nil)
}

// TestServerOpts_AllowUpgrades tests the setter and getter for AllowUpgrades.
func TestServerOpts_AllowUpgrades(t *testing.T) {
	opts := DefaultServerOptions()
	expected := true

	opts.SetAllowUpgrades(expected)

	assert.True(t, opts.AllowUpgrades())
	assert.True(t, opts.GetRawAllowUpgrades() != nil)
}

// TestServerOpts_PerMessageDeflate tests the setter and getter for PerMessageDeflate.
func TestServerOpts_PerMessageDeflate(t *testing.T) {
	opts := DefaultServerOptions()

	var emptyDeflate *xhttp.PerMessageDeflate
	assert.Equal(t, emptyDeflate, opts.PerMessageDeflate())

	expected := &xhttp.PerMessageDeflate{Threshold: 1024}
	opts.SetPerMessageDeflate(expected)

	assert.Equal(t, expected, opts.PerMessageDeflate())
	assert.True(t, opts.GetRawPerMessageDeflate() != nil)
}

// TestServerOpts_HttpCompression tests the setter and getter for HttpCompression.
func TestServerOpts_HttpCompression(t *testing.T) {
	opts := DefaultServerOptions()

	var emptyComp *xhttp.Compression
	assert.Equal(t, emptyComp, opts.HttpCompression())

	expected := &xhttp.Compression{Threshold: 512}
	opts.SetHttpCompression(expected)

	assert.Equal(t, expected, opts.HttpCompression())
	assert.True(t, opts.GetRawHttpCompression() != nil)
}

// TestServerOpts_InitialPacket tests the setter and getter for InitialPacket.
func TestServerOpts_InitialPacket(t *testing.T) {
	opts := DefaultServerOptions()

	var emptyPacket io.Reader
	assert.Equal(t, emptyPacket, opts.InitialPacket())
	var hello = "hello"
	expected := strings.NewReader(hello)
	opts.SetInitialPacket(expected)
	b, _ := io.ReadAll(opts.InitialPacket())
	assert.Equal(t, hello, string(b))
	assert.True(t, opts.GetRawInitialPacket() != nil)
}

// TestServerOpts_Cookie tests the setter and getter for Cookie.
func TestServerOpts_Cookie(t *testing.T) {
	opts := DefaultServerOptions()

	var emptyCookie *http.Cookie
	assert.Equal(t, emptyCookie, opts.Cookie())

	expected := &http.Cookie{Name: "sid", Value: "123"}
	opts.SetCookie(expected)

	assert.Equal(t, expected, opts.Cookie())
	assert.True(t, opts.GetRawCookie() != nil)
}

// TestServerOpts_Cors tests the setter and getter for Cors.
func TestServerOpts_Cors(t *testing.T) {
	opts := DefaultServerOptions()

	var emptyCors *xhttp.Cors
	assert.Equal(t, emptyCors, opts.Cors())

	expected := &xhttp.Cors{Origin: "*"}
	opts.SetCors(expected)

	assert.Equal(t, expected, opts.Cors())
	assert.True(t, opts.GetRawCors() != nil)
}

// TestServerOpts_AllowEIO3 tests the setter and getter for AllowEIO3.
func TestServerOpts_AllowEIO3(t *testing.T) {
	opts := DefaultServerOptions()
	expected := true

	opts.SetAllowEIO3(expected)

	assert.True(t, opts.AllowEIO3())
	assert.True(t, opts.GetRawAllowEIO3() != nil)
}

// TestServerOpts_IdleTimeout tests the setter and getter for IdleTimeout.
func TestServerOpts_IdleTimeout(t *testing.T) {
	opts := DefaultServerOptions()
	expected := 30 * time.Second

	opts.SetIdleTimeout(expected)

	assert.Equal(t, expected, opts.IdleTimeout())
	assert.True(t, opts.GetRawIdleTimeout() != nil)
}
