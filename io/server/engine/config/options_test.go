package config

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

// TestDefaultOptions tests the [DefaultOptions] function to ensure it creates an [Opts]
// with both [AttachOpts] and [ServerOpts] correctly initialized.
func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	assert.True(t, opts.AttachOpts != nil)
	assert.True(t, opts.ServerOpts != nil)
}

// TestOpts_Assign tests the [Opts.Assign] method to ensure it merges data from another [Options] correctly.
func TestOpts_Assign(t *testing.T) {
	t.Run("Assign with nil data", func(t *testing.T) {
		opts := DefaultOptions()

		// Assigning nil should not mutate the existing options and should return the original pointer.
		result := opts.Assign(nil)

		assert.Equal(t, opts, result.(*Opts))
	})

	t.Run("Assign with valid data", func(t *testing.T) {
		opts := DefaultOptions()

		source := DefaultOptions()

		// Set some values that modify the inner [AttachOpts] and [ServerOpts]
		source.SetPath("/new-global-path")
		source.SetAllowUpgrades(true)

		opts.Assign(source)

		// Assert that the properties were properly delegated to both internal structures
		assert.Equal(t, "/new-global-path", opts.Path())
		assert.True(t, opts.AllowUpgrades())
	})
}
