package backoff

import (
	"math"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestNewBackoff_Defaults(t *testing.T) {
	b := NewBackoff()

	assert.Equal(t, defaultMin, b.GetMin())
	assert.Equal(t, defaultMax, b.GetMax())
	assert.Equal(t, defaultFactor, b.GetFactor())
	assert.Equal(t, 0.0, b.GetJitter())
	assert.Equal(t, uint64(0), b.Attempts())
}

func TestNewBackoff_WithOptions(t *testing.T) {
	b := NewBackoff(
		WithMin(200.0),
		WithMax(5000.0),
		WithFactor(3.0),
		WithJitter(0.5),
	)

	assert.Equal(t, 200.0, b.GetMin())
	assert.Equal(t, 5000.0, b.GetMax())
	assert.Equal(t, 3.0, b.GetFactor())
	assert.Equal(t, 0.5, b.GetJitter())
}

func TestNewBackoff_MinGreaterThanMax(t *testing.T) {
	b := NewBackoff(WithMin(20000.0), WithMax(10000.0))
	assert.Equal(t, 10000.0, b.GetMin())
}

func TestBackoff_SettersAndGetters(t *testing.T) {
	b := NewBackoff()

	b.SetMin(500.0)
	assert.Equal(t, 500.0, b.GetMin())

	b.SetMax(2000.0)
	assert.Equal(t, 2000.0, b.GetMax())

	b.SetFactor(2.5)
	assert.Equal(t, 2.5, b.GetFactor())

	b.SetJitter(0.2)
	assert.Equal(t, 0.2, b.GetJitter())
}

func TestBackoff_SettersValidation(t *testing.T) {
	b := NewBackoff(WithMin(100.0), WithMax(1000.0), WithFactor(2.0), WithJitter(0.5))

	// Invalid SetMin
	b.SetMin(-10.0)
	assert.Equal(t, 100.0, b.GetMin())
	b.SetMin(math.NaN())
	assert.Equal(t, 100.0, b.GetMin())
	b.SetMin(2000.0) // clamped to Max
	assert.Equal(t, 1000.0, b.GetMin())

	// Invalid SetMax
	b.SetMax(-50.0)
	assert.Equal(t, 1000.0, b.GetMax())
	b.SetMax(math.Inf(1))
	assert.Equal(t, 1000.0, b.GetMax())
	b.SetMax(50.0) // clamped to Min
	assert.Equal(t, 1000.0, b.GetMax())

	// Invalid SetFactor
	b.SetFactor(0.5)
	assert.Equal(t, 2.0, b.GetFactor())
	b.SetFactor(math.NaN())
	assert.Equal(t, 2.0, b.GetFactor())

	// Invalid SetJitter
	b.SetJitter(1.5)
	assert.Equal(t, 0.5, b.GetJitter())
	b.SetJitter(-0.1)
	assert.Equal(t, 0.5, b.GetJitter())
}

func TestBackoff_DurationAndReset(t *testing.T) {
	b := NewBackoff(WithMin(100.0), WithMax(1000.0), WithFactor(2.0), WithJitter(0.0))

	// First attempt: factor^0 * min = 100
	assert.Equal(t, int64(100), b.Duration())
	assert.Equal(t, uint64(1), b.Attempts())

	// Second attempt: factor^1 * min = 200
	assert.Equal(t, int64(200), b.Duration())
	assert.Equal(t, uint64(2), b.Attempts())

	// Third attempt: factor^2 * min = 400
	assert.Equal(t, int64(400), b.Duration())

	// Fourth attempt: factor^3 * min = 800
	assert.Equal(t, int64(800), b.Duration())

	// Fifth attempt: factor^4 * min = 1600 (Clamped to Max: 1000)
	assert.Equal(t, int64(1000), b.Duration())

	b.Reset()
	assert.Equal(t, uint64(0), b.Attempts())
	assert.Equal(t, int64(100), b.Duration())
}

func TestBackoff_DurationWithJitter(t *testing.T) {
	b := NewBackoff(WithMin(100.0), WithMax(1000.0), WithFactor(2.0), WithJitter(0.1))

	duration := b.Duration()
	// Without jitter, it's 100. Max variance is 10% (90 to 110)
	isWithinBounds := duration >= 90 && duration <= 110
	assert.True(t, isWithinBounds)
}

func TestBackoff_ClampInternalEdgeCases(t *testing.T) {
	// Tests internal clamp logic with invalid floats directly via public methods boundaries
	b := NewBackoff(WithMin(100.0), WithMax(500.0))

	// Forcefully inverted boundaries validation inside Duration
	storeFloat(&b.min, 500.0)
	storeFloat(&b.max, 100.0)

	duration := b.Duration()
	isClamped := duration >= 100 && duration <= 500
	assert.True(t, isClamped)
}
