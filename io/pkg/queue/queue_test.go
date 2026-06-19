package queue

import (
	"sync"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestQueue_EnqueueAndExecute(t *testing.T) {
	q := New()
	defer q.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	var executed bool
	q.Enqueue(func() {
		executed = true
		wg.Done()
	})

	wg.Wait()
	assert.True(t, executed)
}

func TestQueue_Size(t *testing.T) {
	q := New()
	defer q.Close()

	// Pause queue processing by locking the internal execution (simulated via blocking task)
	var block wgBlock
	block.Add(1)
	var started sync.WaitGroup
	started.Add(1)

	q.Enqueue(func() {
		started.Done()
		block.Wait()
	})

	started.Wait() // Ensure the blocking task started running

	q.Enqueue(func() {})
	q.Enqueue(func() {})

	assert.Equal(t, 2, q.Size())

	block.Done()
}

func TestQueue_Close(t *testing.T) {
	q := New()

	var executed bool
	q.Enqueue(func() {
		executed = true
	})

	q.Close()

	assert.True(t, q.IsShuttingDown())
	assert.True(t, executed)

	// Enqueueing after close should be ignored
	q.Enqueue(func() {
		t.Error("Should not execute task after close")
	})
	assert.Equal(t, 0, q.Size())
}

func TestQueue_PanicRecovery(t *testing.T) {
	q := New()
	defer q.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	var secondTaskExecuted bool

	q.Enqueue(func() {
		defer wg.Done()
		panic("something went wrong")
	})

	q.Enqueue(func() {
		secondTaskExecuted = true
		wg.Done()
	})

	wg.Wait()
	assert.True(t, secondTaskExecuted)
}

func TestQueue_EnqueueNilTask(t *testing.T) {
	q := New()
	defer q.Close()

	q.Enqueue(nil)
	assert.Equal(t, 0, q.Size())
}

func TestQueue_TryClose(t *testing.T) {
	q := New()
	q.TryClose()

	assert.True(t, q.IsShuttingDown())
}

// Helper type to mimic waitgroups without violating modern Go rules in production
type wgBlock struct {
	sync.WaitGroup
}
