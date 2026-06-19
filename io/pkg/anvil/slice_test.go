package anvil

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestSlice_NewAndPushLen(t *testing.T) {
	s := NewSlice(1, 2, 3)
	assert.Equal(t, 3, s.Len())
	assert.Equal(t, 5, s.Push(4, 5))
	assert.Equal(t, 5, s.Len())
}

func TestSlice_Unshift(t *testing.T) {
	s := NewSlice(3, 4)
	assert.Equal(t, 4, s.Unshift(1, 2))
	got := s.All()
	assert.Equal(t, 4, len(got))
	assert.Equal(t, 1, got[0])
	assert.Equal(t, 2, got[1])

	// Unshift with no elements is a no-op returning current length.
	assert.Equal(t, 4, s.Unshift())
}

func TestSlice_PopShift(t *testing.T) {
	s := NewSlice("a", "b", "c")

	last, err := s.Pop()
	assert.NoError(t, err)
	assert.Equal(t, "c", last)

	first, err := s.Shift()
	assert.NoError(t, err)
	assert.Equal(t, "a", first)

	assert.Equal(t, 1, s.Len())

	// Drain and verify empty-errors.
	_, err = s.Pop()
	assert.NoError(t, err)
	_, err = s.Pop()
	assert.Error(t, err)
	_, err = s.Shift()
	assert.Error(t, err)
}

func TestSlice_GetSet(t *testing.T) {
	s := NewSlice(10, 20, 30)

	v, err := s.Get(1)
	assert.NoError(t, err)
	assert.Equal(t, 20, v)

	_, err = s.Get(-1)
	assert.Error(t, err)
	_, err = s.Get(99)
	assert.Error(t, err)

	assert.NoError(t, s.Set(0, 99))
	v, _ = s.Get(0)
	assert.Equal(t, 99, v)
	assert.Error(t, s.Set(99, 0))
}

func TestSlice_Slice(t *testing.T) {
	s := NewSlice(1, 2, 3, 4, 5)
	sub, err := s.Slice(1, 4)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(sub))
	assert.Equal(t, 2, sub[0])

	_, err = s.Slice(-1, 2)
	assert.Error(t, err)
	_, err = s.Slice(0, 99)
	assert.Error(t, err)
	_, err = s.Slice(3, 1)
	assert.Error(t, err)
}

func TestSlice_Filter(t *testing.T) {
	s := NewSlice(1, 2, 3, 4, 5, 6)
	even := s.Filter(func(n int) bool { return n%2 == 0 })
	assert.Equal(t, 3, len(even))
}

func TestSlice_Splice(t *testing.T) {
	t.Run("exact replacement", func(t *testing.T) {
		s := NewSlice(1, 2, 3)
		removed, err := s.Splice(1, 1, 20)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(removed))
		assert.Equal(t, 2, removed[0])
		assert.Equal(t, 20, mustGet(t, s, 1))
	})

	t.Run("growing", func(t *testing.T) {
		s := NewSlice(1, 2, 3)
		_, err := s.Splice(1, 1, 20, 21, 22)
		assert.NoError(t, err)
		assert.Equal(t, 5, s.Len())
	})

	t.Run("shrinking", func(t *testing.T) {
		s := NewSlice(1, 2, 3, 4, 5)
		_, err := s.Splice(1, 3, 99)
		assert.NoError(t, err)
		assert.Equal(t, 3, s.Len())
	})

	t.Run("negative deleteCount treated as zero", func(t *testing.T) {
		s := NewSlice(1, 2, 3)
		removed, err := s.Splice(0, -5)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(removed))
	})

	t.Run("out of range start errors", func(t *testing.T) {
		s := NewSlice(1, 2, 3)
		_, err := s.Splice(99, 1)
		assert.Error(t, err)
	})
}

func TestSlice_RemoveAndRemoveAll(t *testing.T) {
	s := NewSlice(1, 2, 3, 2, 1)
	s.Remove(func(n int) bool { return n == 2 })
	assert.Equal(t, 4, s.Len()) // only first match removed

	s.RemoveAll(func(n int) bool { return n == 1 })
	assert.Equal(t, 2, s.Len()) // both 1s removed
}

func TestSlice_Range(t *testing.T) {
	s := NewSlice(1, 2, 3, 4)

	var forward []int
	s.Range(func(v, _ int) bool { forward = append(forward, v); return true })
	assert.Equal(t, 4, len(forward))
	assert.Equal(t, 1, forward[0])

	var reverse []int
	s.Range(func(v, _ int) bool { reverse = append(reverse, v); return true }, true)
	assert.Equal(t, 4, reverse[0])

	// Early termination.
	count := 0
	s.Range(func(_, _ int) bool { count++; return count < 2 })
	assert.Equal(t, 2, count)
}

func TestSlice_RangeAndSplice(t *testing.T) {
	t.Run("forward match splices", func(t *testing.T) {
		s := NewSlice(1, 2, 3)
		removed, err := s.RangeAndSplice(func(v, i int) (bool, int, int, []int) {
			if v == 2 {
				return true, i, 1, []int{99}
			}
			return false, 0, 0, nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(removed))
		assert.Equal(t, 99, mustGet(t, s, 1))
	})

	t.Run("no match returns nil", func(t *testing.T) {
		s := NewSlice(1, 2, 3)
		removed, err := s.RangeAndSplice(func(int, int) (bool, int, int, []int) {
			return false, 0, 0, nil
		})
		assert.NoError(t, err)
		assert.True(t, removed == nil)
	})

	t.Run("reverse match", func(t *testing.T) {
		s := NewSlice(1, 2, 1)
		_, err := s.RangeAndSplice(func(v, i int) (bool, int, int, []int) {
			if v == 1 {
				return true, i, 1, nil
			}
			return false, 0, 0, nil
		}, true)
		assert.NoError(t, err)
		assert.Equal(t, 2, s.Len())
	})
}

func TestSlice_FindIndex(t *testing.T) {
	s := NewSlice(5, 6, 7)
	assert.Equal(t, 1, s.FindIndex(func(n int) bool { return n == 6 }))
	assert.Equal(t, -1, s.FindIndex(func(n int) bool { return n == 100 }))
}

func TestSlice_DoReadDoWrite(t *testing.T) {
	s := NewSlice(1, 2, 3)

	sum := 0
	s.DoRead(func(els []int) {
		for _, v := range els {
			sum += v
		}
	})
	assert.Equal(t, 6, sum)

	s.DoWrite(func(els []int) []int { return append(els, 4) })
	assert.Equal(t, 4, s.Len())
}

func TestSlice_ReplaceAllClear(t *testing.T) {
	s := NewSlice(1, 2, 3)
	s.Replace([]int{9, 8})
	assert.Equal(t, 2, s.Len())

	all := s.All()
	assert.Equal(t, 2, len(all))

	cleared := NewSlice(1, 2, 3)
	got := cleared.AllAndClear()
	assert.Equal(t, 3, len(got))
	assert.Equal(t, 0, cleared.Len())

	s.Clear()
	assert.Equal(t, 0, s.Len())
}

func TestSlice_JSON(t *testing.T) {
	s := NewSlice(1, 2, 3)
	data, err := json.Marshal(s)
	assert.NoError(t, err)
	assert.Equal(t, "[1,2,3]", string(data))

	var decoded Slice[int]
	assert.NoError(t, json.Unmarshal([]byte("[4,5,6]"), &decoded))
	assert.Equal(t, 3, decoded.Len())

	assert.Error(t, json.Unmarshal([]byte("not-json"), &decoded))
}

func TestSlice_Concurrent(t *testing.T) {
	s := NewSlice[int]()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			s.Push(n)
			_ = s.Len()
			s.Range(func(int, int) bool { return true })
		}(i)
	}
	wg.Wait()
	assert.Equal(t, 50, s.Len())
}

func mustGet(t *testing.T, s *Slice[int], i int) int {
	t.Helper()
	v, err := s.Get(i)
	assert.NoError(t, err)
	return v
}
