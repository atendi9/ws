package parser

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestDecoderOptions(t *testing.T) {
	t.Run("DefaultDecoderOptions", func(t *testing.T) {
		opts := DefaultDecoderOptions()

		assert.Equal(t, uint64(0), opts.MaxAttachments())
		assert.Equal(t, 0, opts.MaxNamespaceLength())
		assert.Equal(t, 0, opts.MaxPacketIDLength())
	})

	t.Run("Assign", func(t *testing.T) {
		opts := DefaultDecoderOptions()

		opts = opts.Assign(nil)
		assert.Equal(t, uint64(0), opts.MaxAttachments())

		sourceOpts := DefaultDecoderOptions()
		sourceOpts.SetMaxAttachments(15)
		sourceOpts.SetMaxNamespaceLength(256)
		sourceOpts.SetMaxPacketIDLength(30)

		opts.Assign(sourceOpts)

		assert.Equal(t, uint64(15), opts.MaxAttachments())
		assert.Equal(t, 256, opts.MaxNamespaceLength())
		assert.Equal(t, 30, opts.MaxPacketIDLength())
	})

	t.Run("SetMaxAttachments", func(t *testing.T) {
		opts := DefaultDecoderOptions()
		opts.SetMaxAttachments(50)

		assert.Equal(t, uint64(50), opts.MaxAttachments())
	})

	t.Run("GetRawMaxAttachments", func(t *testing.T) {
		opts := DefaultDecoderOptions()
		rawNil := opts.GetRawMaxAttachments()
		assert.True(t, rawNil == nil)

		opts.SetMaxAttachments(10)
		rawSome := opts.GetRawMaxAttachments()
		assert.False(t, rawSome == nil)
		assert.Equal(t, uint64(10), rawSome.Get())
	})

	t.Run("MaxAttachments", func(t *testing.T) {
		opts := DefaultDecoderOptions()
		assert.Equal(t, uint64(0), opts.MaxAttachments())

		opts.SetMaxAttachments(100)
		assert.Equal(t, uint64(100), opts.MaxAttachments())
	})

	t.Run("SetMaxNamespaceLength", func(t *testing.T) {
		opts := DefaultDecoderOptions()
		opts.SetMaxNamespaceLength(128)

		assert.Equal(t, 128, opts.MaxNamespaceLength())
	})

	t.Run("GetRawMaxNamespaceLength", func(t *testing.T) {
		opts := DefaultDecoderOptions()
		rawNil := opts.GetRawMaxNamespaceLength()
		assert.True(t, rawNil == nil)

		opts.SetMaxNamespaceLength(64)
		rawSome := opts.GetRawMaxNamespaceLength()
		assert.False(t, rawSome == nil)
		assert.Equal(t, 64, rawSome.Get())
	})

	t.Run("MaxNamespaceLength", func(t *testing.T) {
		opts := DefaultDecoderOptions()
		assert.Equal(t, 0, opts.MaxNamespaceLength())

		opts.SetMaxNamespaceLength(512)
		assert.Equal(t, 512, opts.MaxNamespaceLength())
	})

	t.Run("SetMaxPacketIDLength", func(t *testing.T) {
		opts := DefaultDecoderOptions()
		opts.SetMaxPacketIDLength(40)

		assert.Equal(t, 40, opts.MaxPacketIDLength())
	})

	t.Run("GetRawMaxPacketIDLength", func(t *testing.T) {
		opts := DefaultDecoderOptions()
		rawNil := opts.GetRawMaxPacketIDLength()
		assert.True(t, rawNil == nil)

		opts.SetMaxPacketIDLength(20)
		rawSome := opts.GetRawMaxPacketIDLength()
		assert.False(t, rawSome == nil)
		assert.Equal(t, 20, rawSome.Get())
	})

	t.Run("MaxPacketIDLength", func(t *testing.T) {
		opts := DefaultDecoderOptions()
		assert.Equal(t, 0, opts.MaxPacketIDLength())

		opts.SetMaxPacketIDLength(15)
		assert.Equal(t, 15, opts.MaxPacketIDLength())
	})
}
