package anvil

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestGet(t *testing.T) {
	s := []int{10, 20, 30}

	val, ok := Get(s, 1)
	assert.Equal(t, 20, val)
	assert.True(t, ok)

	val, ok = Get(s, 5)
	assert.Equal(t, 0, val)
	assert.False(t, ok)
}

func TestGetAny(t *testing.T) {
	s := []any{10, "hello", 3.14}

	val, ok := GetAny[int](s, 0)
	assert.Equal(t, 10, val)
	assert.True(t, ok)

	_, ok = GetAny[int](s, 1) // Type mismatch
	assert.False(t, ok)

	_, ok = GetAny[int](s, 10) // Out of bounds
	assert.False(t, ok)
}

func TestTryGet(t *testing.T) {
	s := []string{"a", "b"}

	val := TryGet(s, 0)
	assert.Equal(t, "a", val)

	val = TryGet(s, 5)
	assert.Equal(t, "", val)
}

func TestTryGetAny(t *testing.T) {
	s := []any{100, "data"}

	val := TryGetAny[int](s, 0)
	assert.Equal(t, 100, val)

	val = TryGetAny[int](s, 1) // Type mismatch
	assert.Equal(t, 0, val)
}

func TestGetWithDefault(t *testing.T) {
	s := []int{1, 2}

	val := GetWithDefault(s, 0, 99)
	assert.Equal(t, 1, val)

	val = GetWithDefault(s, 10, 99)
	assert.Equal(t, 99, val)
}

func TestGetPtr(t *testing.T) {
	s := []int{10, 20}

	ptr := GetPtr(s, 0)
	assert.Equal(t, 10, *ptr)

	ptr = GetPtr(s, 10)
	assert.Equal(t, (*int)(nil), ptr)
}

func TestSubSlice(t *testing.T) {
	s := []int{0, 1, 2, 3, 4}

	// Positive index
	res := SubSlice(s, 2)
	assert.LengthSlice(t, 3, res)
	assert.Equal(t, 2, res[0])

	// Negative index (last 2 elements: 3, 4)
	res = SubSlice(s, -2)
	assert.LengthSlice(t, 2, res)
	assert.Equal(t, 3, res[0])

	// Out of bounds
	res = SubSlice(s, 10)
	assert.LengthSlice(t, 0, res)
}

func TestFirst(t *testing.T) {
	s := []int{10, 20}
	val, ok := First(s)
	assert.Equal(t, 10, val)
	assert.True(t, ok)

	val, ok = First([]int{})
	assert.False(t, ok)
}

func TestLast(t *testing.T) {
	s := []int{10, 20}
	val, ok := Last(s)
	assert.Equal(t, 20, val)
	assert.True(t, ok)

	val, ok = Last([]int{})
	assert.False(t, ok)
}

func TestFilter(t *testing.T) {
	s := []int{1, 2, 3, 4}
	res := Filter(s, func(n int) bool { return n%2 == 0 })
	assert.LengthSlice(t, 2, res)
	assert.Equal(t, 2, res[0])
	assert.Equal(t, 4, res[1])
}

func TestMap(t *testing.T) {
	s := []int{1, 2, 3}
	res := Transform(s, func(n int) string { return "num" })
	assert.LengthSlice(t, 3, res)
	assert.Equal(t, "num", res[0])
}

func TestReduce(t *testing.T) {
	s := []int{1, 2, 3}
	sum := Reduce(s, 0, func(acc, val int) int { return acc + val })
	assert.Equal(t, 6, sum)
}

func TestContains(t *testing.T) {
	s := []int{1, 2, 3}
	assert.True(t, Contains(s, 2))
	assert.False(t, Contains(s, 5))
}

func TestFindIndex(t *testing.T) {
	s := []int{10, 20, 30}
	idx := FindIndex(s, func(n int) bool { return n == 20 })
	assert.Equal(t, 1, idx)

	idx = FindIndex(s, func(n int) bool { return n == 99 })
	assert.Equal(t, -1, idx)
}

func TestFlatten(t *testing.T) {
	s := [][]int{{1, 2}, {3, 4}}
	res := Flatten(s)
	assert.LengthSlice(t, 4, res)
	assert.Equal(t, 1, res[0])
	assert.Equal(t, 4, res[3])
}

func TestUnique(t *testing.T) {
	s := []int{1, 2, 2, 3, 1}
	res := Unique(s)
	assert.LengthSlice(t, 3, res)
	assert.Equal(t, 1, res[0])
	assert.Equal(t, 2, res[1])
	assert.Equal(t, 3, res[2])
}

func TestIsEmpty(t *testing.T) {
	assert.True(t, IsEmpty([]int{}))
	assert.True(t, IsEmpty([]int(nil)))
	assert.False(t, IsEmpty([]int{1}))
}

func TestIsValidIndex(t *testing.T) {
	s := []int{10}
	assert.True(t, IsValidIndex(s, 0))
	assert.False(t, IsValidIndex(s, 1))
	assert.False(t, IsValidIndex(s, -1))
}
