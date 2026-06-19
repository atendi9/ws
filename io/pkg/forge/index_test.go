package forge

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestIndex(t *testing.T) {
	t.Run("Byte", func(t *testing.T) {
		input1 := []byte("hello world")
		result1 := IndexByte(input1, 'o')
		assert.Equal(t, 4, result1)

		input2 := []byte("golang")
		result2 := IndexByte(input2, 'g')
		assert.Equal(t, 0, result2)

		input3 := []byte("capivara")
		result3 := IndexByte(input3, 'z')
		assert.Equal(t, -1, result3)

		input4 := []byte("")
		result4 := IndexByte(input4, 'a')
		assert.Equal(t, -1, result4)
	})
}
