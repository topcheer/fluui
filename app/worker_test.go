package app

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// ─── Worker Tests ───

func TestWorker_BasicLifecycle(t *testing.T) {
	w := NewWorker("test", func(ctx context.Context, progress func(int)) error {
		progress(50)
		time.Sleep(10 * time.Millisecond)
		progress(100)
		return nil
	})

	if w.State() != WorkerPending {
		t.Error("should start pending")
	}

	w.Start()

	// Wait for completion
	err := w.Wait()
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	if w.State() != WorkerSuccess {
		t.Errorf("expected success, got %s", w.State())
	}
}

func TestWorker_Error(t *testing.T) {
	w := NewWorker("error-task", func(ctx context.Context, progress func(int)) error {
		return context.DeadlineExceeded
	})

	w.Start()
	err := w.Wait()

	if err == nil {
		t.Error("expected error")
	}
	if w.State() != WorkerError {
		t.Errorf("expected error state, got %s", w.State())
	}
}

func TestWorker_Cancel(t *testing.T) {
	w := NewWorker("cancel-test", func(ctx context.Context, progress func(int)) error {
		<-ctx.Done()
		return ctx.Err()
	})

	w.Start()
	time.Sleep(5 * time.Millisecond)
	w.Cancel()

	w.Wait()

	if w.State() != WorkerCancelled {
		t.Errorf("expected cancelled, got %s", w.State())
	}
}

func TestWorker_Progress(t *testing.T) {
	var lastProgress int
	w := NewWorker("progress-test", func(ctx context.Context, progress func(int)) error {
		progress(42)
		progress(84)
		progress(100)
		return nil
	})

	w.OnProgress(func(pct int) {
		lastProgress = pct
	})

	w.Start()
	w.Wait()

	if lastProgress != 100 {
		t.Errorf("expected 100, got %d", lastProgress)
	}
	if w.Progress() != 100 {
		t.Errorf("expected 100, got %d", w.Progress())
	}
}

func TestWorker_OnComplete(t *testing.T) {
	var completed bool
	w := NewWorker("complete-test", func(ctx context.Context, progress func(int)) error {
		return nil
	})

	w.OnComplete(func(err error) {
		completed = true
	})

	w.Start()
	w.Wait()

	if !completed {
		t.Error("OnComplete should fire")
	}
}

func TestWorker_OnStateChange(t *testing.T) {
	var states []WorkerState
	w := NewWorker("state-test", func(ctx context.Context, progress func(int)) error {
		return nil
	})

	w.OnStateChange(func(s WorkerState) {
		states = append(states, s)
	})

	w.Start()
	w.Wait()

	// Should have at least Running and Success
	if len(states) < 2 {
		t.Errorf("expected at least 2 state changes, got %d", len(states))
	}
	if states[0] != WorkerRunning {
		t.Error("first state should be running")
	}
}

func TestWorker_DoubleStart(t *testing.T) {
	w := NewWorker("double", func(ctx context.Context, progress func(int)) error {
		return nil
	})

	if !w.Start() {
		t.Error("first start should succeed")
	}
	if w.Start() {
		t.Error("second start should fail")
	}
}

func TestWorker_IsDone(t *testing.T) {
	w := NewWorker("done-test", func(ctx context.Context, progress func(int)) error {
		return nil
	})

	if w.IsDone() {
		t.Error("should not be done before start")
	}

	w.Start()
	w.Wait()

	if !w.IsDone() {
		t.Error("should be done after wait")
	}
}

func TestWorker_Name(t *testing.T) {
	w := NewWorker("my-worker", func(ctx context.Context, progress func(int)) error {
		return nil
	})
	if w.Name() != "my-worker" {
		t.Error("name mismatch")
	}
}

func TestWorker_ConcurrentAccess(t *testing.T) {
	w := NewWorker("concurrent", func(ctx context.Context, progress func(int)) error {
		progress(50)
		time.Sleep(5 * time.Millisecond)
		progress(100)
		return nil
	})

	// Read state while running
	w.Start()

	// Concurrent reads should not race
	done := make(chan struct{})
	go func() {
		for !w.IsDone() {
			_ = w.State()
			_ = w.Progress()
		}
		close(done)
	}()

	w.Wait()
	<-done
}

// ─── WorkerManager Tests ───

func TestWorkerManager_Basic(t *testing.T) {
	wm := NewWorkerManager()
	w := NewWorker("test", func(ctx context.Context, progress func(int)) error {
		return nil
	})

	wm.Add(w)

	if wm.Count() != 1 {
		t.Error("should have 1 worker")
	}
	if wm.Get("test") != w {
		t.Error("Get should return the worker")
	}
}

func TestWorkerManager_CancelAll(t *testing.T) {
	wm := NewWorkerManager()

	var cancelled atomic.Int32
	for i := 0; i < 3; i++ {
		w := NewWorker("worker-"+string(rune('A'+i)), func(ctx context.Context, progress func(int)) error {
			<-ctx.Done()
			cancelled.Add(1)
			return ctx.Err()
		})
		wm.Add(w)
		w.Start()
	}

	time.Sleep(5 * time.Millisecond)
	wm.CancelAll()

	// Wait for all to finish
	for _, w := range wm.All() {
		w.Wait()
	}

	if cancelled.Load() != 3 {
		t.Errorf("expected 3 cancelled, got %d", cancelled.Load())
	}
}

func TestWorkerManager_ActiveCount(t *testing.T) {
	wm := NewWorkerManager()

	w1 := NewWorker("w1", func(ctx context.Context, progress func(int)) error {
		time.Sleep(20 * time.Millisecond)
		return nil
	})
	w2 := NewWorker("w2", func(ctx context.Context, progress func(int)) error {
		time.Sleep(20 * time.Millisecond)
		return nil
	})

	wm.Add(w1)
	wm.Add(w2)
	w1.Start()
	w2.Start()

	time.Sleep(5 * time.Millisecond)
	if wm.ActiveCount() != 2 {
		t.Errorf("expected 2 active, got %d", wm.ActiveCount())
	}

	w1.Wait()
	w2.Wait()

	if wm.ActiveCount() != 0 {
		t.Errorf("expected 0 active, got %d", wm.ActiveCount())
	}
}

func TestWorkerManager_Remove(t *testing.T) {
	wm := NewWorkerManager()
	w := NewWorker("test", func(ctx context.Context, progress func(int)) error {
		return nil
	})
	wm.Add(w)

	wm.Remove("test")
	if wm.Count() != 0 {
		t.Error("should have 0 after remove")
	}
}

func TestWorkerManager_All(t *testing.T) {
	wm := NewWorkerManager()
	wm.Add(NewWorker("a", func(ctx context.Context, progress func(int)) error { return nil }))
	wm.Add(NewWorker("b", func(ctx context.Context, progress func(int)) error { return nil }))

	all := wm.All()
	if len(all) != 2 {
		t.Errorf("expected 2 workers, got %d", len(all))
	}
}

func TestWorkerState_String(t *testing.T) {
	tests := []struct {
		state WorkerState
		want  string
	}{
		{WorkerPending, "pending"},
		{WorkerRunning, "running"},
		{WorkerSuccess, "success"},
		{WorkerError, "error"},
		{WorkerCancelled, "cancelled"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("state %d: expected %s, got %s", tt.state, tt.want, got)
		}
	}
}