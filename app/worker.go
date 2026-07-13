package app

import (
	"context"
	"sync"
	"sync/atomic"
)

// ─── Worker: Structured Background Task Management ───
//
// Worker provides structured background task management with state tracking,
// cancellation, and safe communication back to the UI thread.
// Inspired by Textual's Worker system.
//
// Usage:
//	w := NewWorker("load-config", func(ctx context.Context, progress func(int)) error {
//	    // do work, check ctx.Done() for cancellation
//	    progress(50)
//	    // ...
//	    progress(100)
//	    return nil
//	})
//	w.OnComplete(func(result error) { ... })
//	w.OnProgress(func(pct int) { ... })
//	w.Start()
//	// Later: w.Cancel()
//	w.Wait() // blocks until done

// WorkerState tracks the lifecycle of a worker.
type WorkerState int32

const (
	WorkerPending WorkerState = iota
	WorkerRunning
	WorkerSuccess
	WorkerError
	WorkerCancelled
)

func (s WorkerState) String() string {
	switch s {
	case WorkerPending:
		return "pending"
	case WorkerRunning:
		return "running"
	case WorkerSuccess:
		return "success"
	case WorkerError:
		return "error"
	case WorkerCancelled:
		return "cancelled"
	}
	return "unknown"
}

// Worker manages a single background task.
type Worker struct {
	mu         sync.Mutex
	name       string
	state      atomic.Int32 // WorkerState
	task       func(ctx context.Context, progress func(int)) error
	ctx        context.Context
	cancel     context.CancelFunc
	result     error
	progress   atomic.Int64
	done       chan struct{}

	onComplete func(error)
	onProgress func(int)
	onStateChange func(WorkerState)
}

// NewWorker creates a worker with the given name and task function.
// The task receives a context (for cancellation) and a progress callback.
func NewWorker(name string, task func(ctx context.Context, progress func(int)) error) *Worker {
	w := &Worker{
		name: name,
		task: task,
		done: make(chan struct{}),
	}
	w.state.Store(int32(WorkerPending))
	return w
}

// Name returns the worker's name.
func (w *Worker) Name() string {
	return w.name
}

// State returns the current worker state.
func (w *Worker) State() WorkerState {
	return WorkerState(w.state.Load())
}

// Progress returns the current progress (0-100).
func (w *Worker) Progress() int {
	return int(w.progress.Load())
}

// Result returns the error result (nil on success).
func (w *Worker) Result() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.result
}

// OnComplete sets a callback fired when the task finishes (success, error, or cancel).
func (w *Worker) OnComplete(fn func(error)) {
	w.mu.Lock()
	w.onComplete = fn
	w.mu.Unlock()
}

// OnProgress sets a callback fired on progress updates.
func (w *Worker) OnProgress(fn func(int)) {
	w.mu.Lock()
	w.onProgress = fn
	w.mu.Unlock()
}

// OnStateChange sets a callback fired on state transitions.
func (w *Worker) OnStateChange(fn func(WorkerState)) {
	w.mu.Lock()
	w.onStateChange = fn
	w.mu.Unlock()
}

func (w *Worker) setState(s WorkerState) {
	w.state.Store(int32(s))
	w.mu.Lock()
	cb := w.onStateChange
	w.mu.Unlock()
	if cb != nil {
		cb(s)
	}
}

// Start begins the background task in a goroutine.
// Returns false if already started.
func (w *Worker) Start() bool {
	w.mu.Lock()
	if WorkerState(w.state.Load()) != WorkerPending {
		w.mu.Unlock()
		return false
	}
	w.ctx, w.cancel = context.WithCancel(context.Background())
	w.mu.Unlock()

	w.setState(WorkerRunning)

	go func() {
		ctx := w.ctx
		task := w.task

		err := task(ctx, func(pct int) {
			w.progress.Store(int64(pct))
			w.mu.Lock()
			cb := w.onProgress
			w.mu.Unlock()
			if cb != nil {
				cb(pct)
			}
		})

		w.mu.Lock()
		w.result = err
		completeCB := w.onComplete
		w.mu.Unlock()

		if ctx.Err() != nil {
			w.setState(WorkerCancelled)
		} else if err != nil {
			w.setState(WorkerError)
		} else {
			w.setState(WorkerSuccess)
		}

		if completeCB != nil {
			completeCB(err)
		}
		close(w.done)
	}()

	return true
}

// Cancel requests cancellation of the background task.
func (w *Worker) Cancel() {
	w.mu.Lock()
	if w.cancel != nil {
		w.cancel()
	}
	w.mu.Unlock()
}

// Wait blocks until the task completes. Returns the result error.
func (w *Worker) Wait() error {
	<-w.done
	return w.Result()
}

// IsDone returns true if the worker has finished (any terminal state).
func (w *Worker) IsDone() bool {
	s := w.State()
	return s == WorkerSuccess || s == WorkerError || s == WorkerCancelled
}

// ─── WorkerManager: Multiple Worker Tracking ───

// WorkerManager tracks multiple workers.
type WorkerManager struct {
	mu      sync.RWMutex
	workers map[string]*Worker
}

// NewWorkerManager creates a worker manager.
func NewWorkerManager() *WorkerManager {
	return &WorkerManager{
		workers: make(map[string]*Worker),
	}
}

// Add registers a worker.
func (wm *WorkerManager) Add(w *Worker) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.workers[w.Name()] = w
}

// Get returns a worker by name.
func (wm *WorkerManager) Get(name string) *Worker {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.workers[name]
}

// CancelAll cancels all running workers.
func (wm *WorkerManager) CancelAll() {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	for _, w := range wm.workers {
		if !w.IsDone() {
			w.Cancel()
		}
	}
}

// ActiveCount returns the number of running workers.
func (wm *WorkerManager) ActiveCount() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	count := 0
	for _, w := range wm.workers {
		if w.State() == WorkerRunning {
			count++
		}
	}
	return count
}

// All returns all workers.
func (wm *WorkerManager) All() []*Worker {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	result := make([]*Worker, 0, len(wm.workers))
	for _, w := range wm.workers {
		result = append(result, w)
	}
	return result
}

// Remove deletes a worker from tracking.
func (wm *WorkerManager) Remove(name string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	delete(wm.workers, name)
}

// Count returns total number of tracked workers.
func (wm *WorkerManager) Count() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return len(wm.workers)
}