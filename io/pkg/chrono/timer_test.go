package chrono

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/atendi9/capivara/assert"
)

func TestSetTimeoutFires(t *testing.T) {
	done := make(chan struct{})
	SetTimeout(func() { close(done) }, 10*time.Millisecond)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout callback never fired")
	}
}

func TestClearTimeoutPreventsFire(t *testing.T) {
	var fired atomic.Bool
	timer := SetTimeout(func() { fired.Store(true) }, 50*time.Millisecond)
	ClearTimeout(timer)

	time.Sleep(120 * time.Millisecond)
	assert.False(t, fired.Load())
}

func TestClearTimeoutNilSafe(t *testing.T) {
	// Should not panic on nil.
	ClearTimeout(nil)
}

func TestTimerStop(t *testing.T) {
	var fired atomic.Bool
	timer := SetTimeout(func() { fired.Store(true) }, 50*time.Millisecond)
	timer.Stop()

	time.Sleep(120 * time.Millisecond)
	assert.False(t, fired.Load())
}

func TestSetIntervalFiresRepeatedly(t *testing.T) {
	var count atomic.Int32
	timer := SetInterval(func() { count.Add(1) }, 15*time.Millisecond)
	defer ClearInterval(timer)

	time.Sleep(100 * time.Millisecond)
	ClearInterval(timer)

	assert.True(t, count.Load() >= 2)
}

func TestClearIntervalStops(t *testing.T) {
	var count atomic.Int32
	timer := SetInterval(func() { count.Add(1) }, 15*time.Millisecond)
	ClearInterval(timer)

	time.Sleep(80 * time.Millisecond)
	stopped := count.Load()
	time.Sleep(80 * time.Millisecond)
	// No further increments after clearing.
	assert.Equal(t, stopped, count.Load())
}

func TestTimerRefresh(t *testing.T) {
	fired := make(chan struct{}, 1)
	timer := SetTimeout(func() { fired <- struct{}{} }, 40*time.Millisecond)

	// Refresh resets the countdown and returns the same timer.
	refreshed := timer.Refresh()
	assert.NotNil(t, refreshed)

	select {
	case <-fired:
	case <-time.After(time.Second):
		t.Fatal("refreshed timer never fired")
	}
}

func TestTimerUnref(t *testing.T) {
	// Unref should not panic and returns nothing; just ensure it is callable.
	timer := SetTimeout(func() {}, time.Hour)
	timer.Unref()
	timer.Stop()
}
