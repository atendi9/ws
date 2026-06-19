package anvil

import (
	"strconv"
	"testing"

	"github.com/atendi9/capivara/assert"
)

// TestValues tests the [anvil.Values] function covering different scenarios
// such as nil maps, empty maps, and populated maps.
func TestValues(t *testing.T) {
	t.Run("should return nil when input is nil", func(t *testing.T) {
		var input map[string]int = nil
		result := Values(input, func(v int) string {
			return strconv.Itoa(v)
		})

		assert.True(t, nil == result)
	})

	t.Run("should return empty map when input is empty", func(t *testing.T) {
		input := map[string]int{}
		result := Values(input, func(v int) string {
			return strconv.Itoa(v)
		})

		assert.LengthMap(t, 0, result)
	})

	t.Run("should transform values correctly", func(t *testing.T) {
		input := map[string]int{
			"one":   1,
			"two":   2,
			"three": 3,
		}

		// Transform int to string
		transform := func(v int) string {
			return strconv.Itoa(v * 10)
		}

		result := Values(input, transform)

		assert.LengthMap(t, 3, result)
		assert.Equal(t, "10", result["one"])
		assert.Equal(t, "20", result["two"])
		assert.Equal(t, "30", result["three"])
	})

	t.Run("should handle custom types", func(t *testing.T) {
		type User struct {
			Name string
		}

		input := map[int]User{
			1: {Name: "Alice"},
			2: {Name: "Bob"},
		}

		// Transform User to string
		transform := func(u User) string {
			return u.Name + "!"
		}

		result := Values(input, transform)

		assert.LengthMap(t, 2, result)
		assert.Equal(t, "Alice!", result[1])
		assert.Equal(t, "Bob!", result[2])
	})
}
