// Package queue provides a thread-safe, asynchronous task queue implementation.
// It allows tasks to be enqueued and executed sequentially in a background goroutine.
package queue

import (
	"runtime"
	"sync"
)

// Queue represents a thread-safe FIFO (First-In-First-Out) task queue.
// It manages sequential execution of tasks in a background goroutine.
type Queue struct {
	mu           sync.Mutex
	cond         *sync.Cond
	tasks        []func()
	shuttingDown bool
	done         chan struct{}
}

// New creates and initializes a new [Queue].
// It automatically starts the background execution loop and registers a finalizer
// to ensure resource cleanup when the [Queue] is garbage collected.
func New() *Queue {
	q := &Queue{
		tasks: make([]func(), 0, 1024),
		done:  make(chan struct{}),
	}
	q.cond = sync.NewCond(&q.mu)

	go q.loop()
	runtime.SetFinalizer(q, func(q *Queue) { q.TryClose() })
	return q
}

// Enqueue adds a new task to the [Queue].
// If the task is nil or the [Queue] is shutting down, the task is ignored.
func (q *Queue) Enqueue(task func()) {
	if task == nil {
		return
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.shuttingDown {
		return
	}

	q.tasks = append(q.tasks, task)
	q.cond.Signal()
}

// Size returns the current number of pending tasks in the [Queue].
func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.tasks)
}

// loop runs in a background goroutine, continuously fetching and executing tasks
// until the [Queue] is explicitly closed or shut down.
func (q *Queue) loop() {
	defer close(q.done)

	for {
		task, ok := q.get()
		if !ok {
			return
		}
		q.execute(task)
	}
}

// get retrieves the next task from the [Queue].
// It blocks if the queue is empty until a new task is enqueued or the [Queue] starts shutting down.
// Returns the task and true if successful, or nil and false if the [Queue] is shutting down.
func (q *Queue) get() (func(), bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.tasks) == 0 && !q.shuttingDown {
		q.cond.Wait()
	}

	if len(q.tasks) == 0 && q.shuttingDown {
		return nil, false
	}

	task := q.tasks[0]

	q.tasks[0] = nil
	q.tasks = q.tasks[1:]

	if len(q.tasks) == 0 {
		q.tasks = q.tasks[:0]
	}

	return task, true
}

// execute runs the given task safely, recovering from any potential panic
// to prevent the entire process from crashing.
func (q *Queue) execute(task func()) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	task()
}

// Close gracefully triggers the shutdown of the [Queue], signals all waiting goroutines,
// and blocks until the background execution loop completes.
func (q *Queue) Close() {
	q.mu.Lock()
	q.shuttingDown = true
	q.cond.Broadcast()
	q.mu.Unlock()

	<-q.done
}

// IsShuttingDown checks whether the [Queue] is currently in the process of shutting down or is already closed.
func (q *Queue) IsShuttingDown() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.shuttingDown
}

// TryClose triggers the shutdown sequence of the [Queue] without blocking or waiting
// for the background loop to finish. It is safely used by the garbage collection finalizer.
func (q *Queue) TryClose() {
	q.mu.Lock()
	q.shuttingDown = true
	q.cond.Broadcast()
	q.mu.Unlock()
}
