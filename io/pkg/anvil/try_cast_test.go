package anvil

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestTryCast(t *testing.T) {
	t.Run("should correctly cast primitive types and handle mismatches", func(t *testing.T) {
		expectedStr := "hello"
		expectedNum := 0
		val := any(expectedStr)

		str := TryCast[string](val)
		assert.Equal(t, expectedStr, str)

		num := TryCast[int](val)
		assert.Equal(t, expectedNum, num)
	})

	t.Run("should correctly cast struct types", func(t *testing.T) {
		type User struct {
			Name string
			Age  int
		}

		expectedUser := User{Name: "Alice", Age: 30}
		val := any(expectedUser)

		resultUser := TryCast[User](val)
		assert.Equal(t, expectedUser, resultUser)

		type Product struct {
			ID string
		}
		failedResult := TryCast[Product](val)
		assert.Equal(t, Product{}, failedResult)
	})
}
