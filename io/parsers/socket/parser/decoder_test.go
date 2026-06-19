package parser

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestDecoder(t *testing.T) {
	t.Run("NewDecoder", func(t *testing.T) {
		d := NewDecoder()

		isCreated := d != nil
		assert.True(t, isCreated)
	})

	t.Run("Add", func(t *testing.T) {
		d := NewDecoder()
		err := d.Add(12345)
		assert.Error(t, err)

		err = d.Add("0")
		assert.NoError(t, err)
	})

	t.Run("Destroy", func(t *testing.T) {
		d := NewDecoder()
		d.Destroy()

		assert.True(t, true)
	})
}
