// Package backoff provides a thread-safe, configurable exponential backoff implementation
// using atomic operations.
package backoff

import (
	"math"
	"math/rand/v2"
	"sync/atomic"
)

// Backoff manages exponential backoff state and parameters safely across multiple goroutines.
// To instantiate a new instance, use the [NewBackoff] function.
type Backoff struct {
	min      atomic.Uint64
	max      atomic.Uint64
	factor   atomic.Uint64
	jitter   atomic.Uint64
	attempts atomic.Uint64
}

// BackoffOption defines a configuration function used to customize a [*Backoff] instance.
type BackoffOption func(*Backoff)

const (
	defaultMin    = 100.0
	defaultMax    = 10_000.0
	defaultFactor = 2.0
	// maxAttempts prevent math.Pow overflow
	maxAttempts = 63
)

// WithMin returns a [BackoffOption] that configures the minimum duration.
// The value must be greater than 0 and a valid finite number.
func WithMin(min float64) BackoffOption {
	return func(b *Backoff) {
		if isValid(min) && min > 0 {
			storeFloat(&b.min, min)
		}
	}
}

// WithMax returns a [BackoffOption] that configures the maximum duration.
// The value must be greater than 0 and a valid finite number.
func WithMax(max float64) BackoffOption {
	return func(b *Backoff) {
		if isValid(max) && max > 0 {
			storeFloat(&b.max, max)
		}
	}
}

// WithFactor returns a [BackoffOption] that configures the multiplication factor.
// The value must be greater than 1 and a valid finite number.
func WithFactor(factor float64) BackoffOption {
	return func(b *Backoff) {
		if isValid(factor) && factor > 1 {
			storeFloat(&b.factor, factor)
		}
	}
}

// WithJitter returns a [BackoffOption] that configures the jitter ratio.
// The value must be between 0 and 1 inclusive.
func WithJitter(jitter float64) BackoffOption {
	return func(b *Backoff) {
		if isValid(jitter) && jitter >= 0 && jitter <= 1 {
			storeFloat(&b.jitter, jitter)
		}
	}
}

// NewBackoff initializes and returns a new [*Backoff] instance with default values
// and applies any provided [BackoffOption]. If the configured minimum value exceeds
// the maximum value, the minimum will be clamped to the maximum value.
func NewBackoff(opts ...BackoffOption) *Backoff {
	b := &Backoff{}
	storeFloat(&b.min, defaultMin)
	storeFloat(&b.max, defaultMax)
	storeFloat(&b.factor, defaultFactor)

	for _, opt := range opts {
		opt(b)
	}

	if b.GetMin() > b.GetMax() {
		storeFloat(&b.min, b.GetMax())
	}

	return b
}

// Attempts returns the current number of backoff attempts made.
func (b *Backoff) Attempts() uint64 {
	return b.attempts.Load()
}

// Duration calculates and returns the next backoff duration in nanoseconds.
// It increments the attempt counter atomically and applies exponential growth and jitter if configured.
func (b *Backoff) Duration() int64 {
	attempt := min(b.attempts.Add(1)-1, maxAttempts)

	minVal := loadFloat(&b.min)
	maxVal := loadFloat(&b.max)
	factor := loadFloat(&b.factor)
	jitter := loadFloat(&b.jitter)

	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}

	duration := minVal * math.Pow(factor, float64(attempt))
	duration = clamp(duration, minVal, maxVal)

	if jitter > 0 {
		offset := jitter * duration * (rand.Float64()*2 - 1)
		duration = clamp(duration+offset, minVal, maxVal)
	}

	return int64(duration)
}

// Reset resets the internal attempt counter back to 0.
func (b *Backoff) Reset() {
	b.attempts.Store(0)
}

// SetMin updates the minimum duration value dynamically.
// The value is automatically clamped to the current maximum value.
func (b *Backoff) SetMin(val float64) {
	if !isValid(val) || val <= 0 {
		return
	}
	storeFloat(&b.min, min(val, b.GetMax()))
}

// SetMax updates the maximum duration value dynamically.
// The value is automatically clamped to the current minimum value.
func (b *Backoff) SetMax(val float64) {
	if !isValid(val) || val <= 0 {
		return
	}
	storeFloat(&b.max, max(val, b.GetMin()))
}

// SetFactor updates the multiplication factor dynamically.
// The value must be greater than 1 and a valid finite number.
func (b *Backoff) SetFactor(val float64) {
	if isValid(val) && val > 1 {
		storeFloat(&b.factor, val)
	}
}

// SetJitter updates the jitter ratio dynamically.
// The value must be between 0 and 1 inclusive.
func (b *Backoff) SetJitter(val float64) {
	if isValid(val) && val >= 0 && val <= 1 {
		storeFloat(&b.jitter, val)
	}
}

// GetMin returns the current minimum duration value.
func (b *Backoff) GetMin() float64 {
	return loadFloat(&b.min)
}

// GetMax returns the current maximum duration value.
func (b *Backoff) GetMax() float64 {
	return loadFloat(&b.max)
}

// GetFactor returns the current multiplication factor.
func (b *Backoff) GetFactor() float64 {
	return loadFloat(&b.factor)
}

// GetJitter returns the current jitter ratio.
func (b *Backoff) GetJitter() float64 {
	return loadFloat(&b.jitter)
}

func storeFloat(target *atomic.Uint64, val float64) {
	target.Store(math.Float64bits(val))
}

func loadFloat(source *atomic.Uint64) float64 {
	return math.Float64frombits(source.Load())
}

func isValid(val float64) bool {
	return !math.IsNaN(val) && !math.IsInf(val, 0)
}

func clamp(val, minVal, maxVal float64) float64 {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return maxVal
	}
	return max(minVal, min(val, maxVal))
}
