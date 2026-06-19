package anvil

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestTap(t *testing.T) {
	t.Run("should execute callback and return the same value", func(t *testing.T) {
		type User struct {
			Name string
			Age  int
		}

		user := User{Name: "Gabriel Luiz", Age: 25}
		var called bool
		var receivedUser User

		result := Tap(user, func(u User) {
			called = true
			receivedUser = u
		})

		assert.True(t, called)
		assert.Equal(t, user.Name, receivedUser.Name)
		assert.Equal(t, user.Age, receivedUser.Age)
		assert.Equal(t, user, result)
	})

	t.Run("should safely ignore execution when callback is nil", func(t *testing.T) {
		value := "capivara"

		result := Tap(value, nil)

		assert.Equal(t, value, result)
	})

	t.Run("should work seamlessly with integers", func(t *testing.T) {
		number := 42
		var capturedNumber int

		result := Tap(number, func(n int) {
			capturedNumber = n
		})

		assert.Equal(t, number, capturedNumber)
		assert.Equal(t, number, result)
	})
}
