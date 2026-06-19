// Package config provides tests for the configuration types and options of the Engine.IO server.
package config

import (
	"testing"
	"time"

	"github.com/atendi9/capivara/assert"
)

// TestDefaultAttachOptions tests the [DefaultAttachOptions] function to ensure it creates an empty [AttachOpts].
func TestDefaultAttachOptions(t *testing.T) {
	opts := DefaultAttachOptions()

	assert.Equal(t, "", opts.Path())
	assert.False(t, opts.DestroyUpgrade())
	assert.Equal(t, time.Duration(0), opts.DestroyUpgradeTimeout())
	assert.False(t, opts.AddTrailingSlash())
}

// TestAttachOpts_Path tests the [AttachOpts.SetPath], [AttachOpts.GetRawPath], and [AttachOpts.Path] methods.
func TestAttachOpts_Path(t *testing.T) {
	opts := DefaultAttachOptions()

	// Ensure the default state is empty and nil.
	assert.Equal(t, "", opts.Path())
	assert.True(t, opts.GetRawPath() == nil)

	expectedPath := "/socket.io/"
	opts.SetPath(expectedPath)

	assert.Equal(t, expectedPath, opts.Path())
	assert.True(t, opts.GetRawPath() != nil)
}

// TestAttachOpts_DestroyUpgrade tests the [AttachOpts.SetDestroyUpgrade], [AttachOpts.GetRawDestroyUpgrade], and [AttachOpts.DestroyUpgrade] methods.
func TestAttachOpts_DestroyUpgrade(t *testing.T) {
	opts := DefaultAttachOptions()

	// Ensure the default state is false and nil.
	assert.False(t, opts.DestroyUpgrade())
	assert.True(t, opts.GetRawDestroyUpgrade() == nil)

	opts.SetDestroyUpgrade(true)

	assert.True(t, opts.DestroyUpgrade())
	assert.True(t, opts.GetRawDestroyUpgrade() != nil)
}

// TestAttachOpts_DestroyUpgradeTimeout tests the [AttachOpts.SetDestroyUpgradeTimeout], [AttachOpts.GetRawDestroyUpgradeTimeout], and [AttachOpts.DestroyUpgradeTimeout] methods.
func TestAttachOpts_DestroyUpgradeTimeout(t *testing.T) {
	opts := DefaultAttachOptions()

	// Ensure the default state is zero and nil.
	assert.Equal(t, time.Duration(0), opts.DestroyUpgradeTimeout())
	assert.True(t, opts.GetRawDestroyUpgradeTimeout() == nil)

	expectedTimeout := 5 * time.Second
	opts.SetDestroyUpgradeTimeout(expectedTimeout)

	assert.Equal(t, expectedTimeout, opts.DestroyUpgradeTimeout())
	assert.True(t, opts.GetRawDestroyUpgradeTimeout() != nil)
}

// TestAttachOpts_AddTrailingSlash tests the [AttachOpts.SetAddTrailingSlash], [AttachOpts.GetRawAddTrailingSlash], and [AttachOpts.AddTrailingSlash] methods.
func TestAttachOpts_AddTrailingSlash(t *testing.T) {
	opts := DefaultAttachOptions()

	// Ensure the default state is false and nil.
	assert.False(t, opts.AddTrailingSlash())
	assert.True(t, opts.GetRawAddTrailingSlash() == nil)

	opts.SetAddTrailingSlash(true)

	assert.True(t, opts.AddTrailingSlash())
	assert.True(t, opts.GetRawAddTrailingSlash() != nil)
}

// TestAttachOpts_Assign tests the [AttachOpts.Assign] method to ensure it merges [AttachOptions] correctly.
func TestAttachOpts_Assign(t *testing.T) {
	t.Run("Assign with nil data", func(t *testing.T) {
		opts := DefaultAttachOptions()
		opts.SetPath("/default")

		// Assigning nil should not mutate the existing options.
		opts.Assign(nil)

		assert.Equal(t, "/default", opts.Path())
	})

	t.Run("Assign with valid fully populated data", func(t *testing.T) {
		opts := DefaultAttachOptions()

		source := DefaultAttachOptions()
		source.SetPath("/new-path")
		source.SetDestroyUpgrade(true)
		source.SetDestroyUpgradeTimeout(10 * time.Second)
		source.SetAddTrailingSlash(true)

		opts.Assign(source)

		assert.Equal(t, "/new-path", opts.Path())
		assert.True(t, opts.DestroyUpgrade())
		assert.Equal(t, 10*time.Second, opts.DestroyUpgradeTimeout())
		assert.True(t, opts.AddTrailingSlash())
	})

	t.Run("Assign with partial data", func(t *testing.T) {
		opts := DefaultAttachOptions()
		opts.SetPath("/old-path")
		opts.SetAddTrailingSlash(true)

		source := DefaultAttachOptions()
		source.SetDestroyUpgrade(true)
		// Path and AddTrailingSlash are explicitly not set on the source.

		opts.Assign(source)

		// Values defined on opts should be preserved if not present on the source.
		assert.Equal(t, "/old-path", opts.Path())
		assert.True(t, opts.AddTrailingSlash())

		// Value from the source should be successfully merged.
		assert.True(t, opts.DestroyUpgrade())
	})
}
