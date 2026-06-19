package socket

import (
	"testing"
	"time"

	"github.com/atendi9/capivara/assert"
)

func TestCSRecovery_Defaults(t *testing.T) {
	c := DefaultConnectionStateRecovery()
	assert.NotNil(t, c)

	// Unset values return zero and nil raw optionals.
	assert.Equal(t, int64(0), c.MaxDisconnectionDuration())
	assert.False(t, c.SkipMiddlewares())
	assert.Equal(t, time.Duration(0), c.SessionCleanupInterval())
	assert.True(t, c.GetRawMaxDisconnectionDuration() == nil)
	assert.True(t, c.GetRawSkipMiddlewares() == nil)
	assert.True(t, c.GetRawSessionCleanupInterval() == nil)
}

func TestCSRecovery_SettersGetters(t *testing.T) {
	c := DefaultConnectionStateRecovery()

	c.SetMaxDisconnectionDuration(5000)
	assert.Equal(t, int64(5000), c.MaxDisconnectionDuration())
	assert.NotNil(t, c.GetRawMaxDisconnectionDuration())

	c.SetSkipMiddlewares(true)
	assert.True(t, c.SkipMiddlewares())
	assert.NotNil(t, c.GetRawSkipMiddlewares())

	c.SetSessionCleanupInterval(30 * time.Second)
	assert.Equal(t, 30*time.Second, c.SessionCleanupInterval())
	assert.NotNil(t, c.GetRawSessionCleanupInterval())
}

func TestCSRecovery_Assign(t *testing.T) {
	t.Run("nil source returns receiver unchanged", func(t *testing.T) {
		c := DefaultConnectionStateRecovery()
		out := c.Assign(nil)
		assert.NotNil(t, out)
		assert.Equal(t, int64(0), out.MaxDisconnectionDuration())
	})

	t.Run("copies set fields", func(t *testing.T) {
		src := DefaultConnectionStateRecovery()
		src.SetMaxDisconnectionDuration(1000)
		src.SetSkipMiddlewares(true)
		src.SetSessionCleanupInterval(time.Minute)

		dst := DefaultConnectionStateRecovery()
		out := dst.Assign(src)
		assert.Equal(t, int64(1000), out.MaxDisconnectionDuration())
		assert.True(t, out.SkipMiddlewares())
		assert.Equal(t, time.Minute, out.SessionCleanupInterval())
	})
}

func TestServerOpts_Defaults(t *testing.T) {
	s := DefaultServerOptions()
	assert.NotNil(t, s)

	assert.False(t, s.ServeClient())
	assert.Equal(t, "", s.ClientVersion())
	assert.True(t, s.Adapter() == nil)
	assert.True(t, s.Parser() == nil)
	assert.Equal(t, time.Duration(0), s.ConnectTimeout())
	assert.True(t, s.ConnectionStateRecovery() == nil)
	assert.False(t, s.CleanupEmptyChildNamespaces())

	assert.True(t, s.GetRawServeClient() == nil)
	assert.True(t, s.GetRawClientVersion() == nil)
	assert.True(t, s.GetRawAdapter() == nil)
	assert.True(t, s.GetRawParser() == nil)
	assert.True(t, s.GetRawConnectTimeout() == nil)
	assert.True(t, s.GetRawConnectionStateRecovery() == nil)
	assert.True(t, s.GetRawCleanupEmptyChildNamespaces() == nil)
}

func TestServerOpts_SettersGetters(t *testing.T) {
	s := DefaultServerOptions()

	s.SetServeClient(true)
	assert.True(t, s.ServeClient())
	assert.NotNil(t, s.GetRawServeClient())

	s.SetClientVersion("4.7.2")
	assert.Equal(t, "4.7.2", s.ClientVersion())

	s.SetConnectTimeout(45 * time.Second)
	assert.Equal(t, 45*time.Second, s.ConnectTimeout())

	s.SetCleanupEmptyChildNamespaces(true)
	assert.True(t, s.CleanupEmptyChildNamespaces())

	csr := DefaultConnectionStateRecovery()
	csr.SetMaxDisconnectionDuration(100)
	s.SetConnectionStateRecovery(csr)
	assert.NotNil(t, s.ConnectionStateRecovery())
	assert.Equal(t, int64(100), s.ConnectionStateRecovery().MaxDisconnectionDuration())
}

func TestServerOpts_Assign(t *testing.T) {
	t.Run("nil source returns receiver", func(t *testing.T) {
		s := DefaultServerOptions()
		out := s.Assign(nil)
		assert.NotNil(t, out)
		assert.False(t, out.ServeClient())
	})

	t.Run("copies set fields from source", func(t *testing.T) {
		src := DefaultServerOptions()
		src.SetServeClient(true)
		src.SetClientVersion("9.9.9")
		src.SetConnectTimeout(time.Second)
		src.SetCleanupEmptyChildNamespaces(true)

		dst := DefaultServerOptions()
		out := dst.Assign(src)
		assert.True(t, out.ServeClient())
		assert.Equal(t, "9.9.9", out.ClientVersion())
		assert.Equal(t, time.Second, out.ConnectTimeout())
		assert.True(t, out.CleanupEmptyChildNamespaces())
	})
}
