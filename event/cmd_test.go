package event

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCmd_Exec_Basic(t *testing.T) {
	loop := NewLoop(nil, NewDispatcher())
	exec := NewCmdExecutor(loop)

	var got atomic.Bool
	exec.Exec(func() Event {
		got.Store(true)
		return Event{}
	})

	// Wait for goroutine
	time.Sleep(50 * time.Millisecond)
	if !got.Load() {
		t.Fatal("Cmd should have executed")
	}
	exec.Stop()
}

func TestCmd_Exec_SendsResult(t *testing.T) {
	loop := NewLoop(nil, NewDispatcher())
	exec := NewCmdExecutor(loop)

	exec.Exec(func() Event {
		return Event{Type: TypeCustom, Data: "done"}
	})

	// Event should arrive in loop's customCh
	select {
	case ev := <-loop.customCh:
		if ev.Data != "done" {
			t.Fatalf("expected Data='done', got %q", ev.Data)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for Cmd result")
	}
	exec.Stop()
}

func TestCmd_NilCmd(t *testing.T) {
	loop := NewLoop(nil, NewDispatcher())
	exec := NewCmdExecutor(loop)
	exec.Exec(nil) // should not panic
	exec.Stop()
}

func TestCmd_Stop_Cancels(t *testing.T) {
	loop := NewLoop(nil, NewDispatcher())
	exec := NewCmdExecutor(loop)

	var ran atomic.Bool
	exec.Exec(func() Event {
		time.Sleep(200 * time.Millisecond)
		ran.Store(true)
		return Event{}
	})

	exec.Stop() // cancels immediately
	if ran.Load() {
		t.Fatal("Cmd should have been cancelled")
	}
}

func TestCmd_PanicRecovery(t *testing.T) {
	loop := NewLoop(nil, NewDispatcher())
	exec := NewCmdExecutor(loop)

	exec.Exec(func() Event {
		panic("test panic")
	})

	// Should get a panic error event
	select {
	case ev := <-loop.customCh:
		if ev.Data == "" {
			t.Fatal("expected non-empty panic error data")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for panic recovery event")
	}
	exec.Stop()
}

func TestCmd_Tick(t *testing.T) {
	loop := NewLoop(nil, NewDispatcher())
	exec := NewCmdExecutor(loop)

	TickWithExecutor(exec, 20*time.Millisecond, func(t time.Time) Event {
		return Event{Type: TypeCustom, Data: "tick"}
	})

	select {
	case ev := <-loop.customCh:
		if ev.Data != "tick" {
			t.Fatalf("expected 'tick', got %q", ev.Data)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for tick")
	}
	exec.Stop()
}

func TestCmd_Every(t *testing.T) {
	loop := NewLoop(nil, NewDispatcher())
	exec := NewCmdExecutor(loop)

	var count atomic.Int32
	Every(exec, 10*time.Millisecond, func(t time.Time) Event {
		count.Add(1)
		return Event{Type: TypeCustom, Data: "tock"}
	})

	time.Sleep(55 * time.Millisecond) // should get ~5 ticks
	exec.Stop()

	if count.Load() < 3 {
		t.Fatalf("expected at least 3 ticks, got %d", count.Load())
	}
}

func TestCmd_BatchWithExecutor(t *testing.T) {
	loop := NewLoop(nil, NewDispatcher())
	exec := NewCmdExecutor(loop)

	var wg sync.WaitGroup
	cmds := make([]Cmd, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		idx := i
		cmds[i] = func() Event {
			defer wg.Done()
			return Event{Type: TypeCustom, Data: string(rune('a' + idx))}
		}
	}

	BatchWithExecutor(exec, cmds...)

	// Should receive 5 events
	results := make(map[string]bool)
	for i := 0; i < 5; i++ {
		select {
		case ev := <-loop.customCh:
			results[ev.Data] = true
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("timed out waiting for result %d, got %v", i, results)
		}
	}

	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}
	exec.Stop()
}

func TestCmd_SequenceWithExecutor(t *testing.T) {
	loop := NewLoop(nil, NewDispatcher())
	exec := NewCmdExecutor(loop)

	var order []int
	var mu sync.Mutex

	cmds := make([]Cmd, 3)
	for i := 0; i < 3; i++ {
		idx := i
		cmds[i] = func() Event {
			mu.Lock()
			order = append(order, idx)
			mu.Unlock()
			return Event{Type: TypeCustom, Data: string(rune('A' + idx))}
		}
	}

	SequenceWithExecutor(exec, cmds...)

	time.Sleep(100 * time.Millisecond)
	exec.Stop()

	mu.Lock()
	defer mu.Unlock()
	if len(order) != 3 {
		t.Fatalf("expected 3 sequential results, got %d", len(order))
	}
	for i, v := range order {
		if v != i {
			t.Fatalf("expected order[%d]=%d, got %d", i, i, v)
		}
	}
}

func TestCmd_LoopExec(t *testing.T) {
	loop := NewLoop(nil, NewDispatcher())

	loop.Exec(func() Event {
		return Event{Type: TypeCustom, Data: "from-loop"}
	})

	select {
	case ev := <-loop.customCh:
		if ev.Data != "from-loop" {
			t.Fatalf("expected 'from-loop', got %q", ev.Data)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out")
	}

	loop.cmdExec.Stop()
}

func TestCmd_Quit(t *testing.T) {
	loop := NewLoop(nil, NewDispatcher())
	exec := NewCmdExecutor(loop)

	exec.Exec(Quit())

	select {
	case ev := <-loop.customCh:
		if ev.Type != TypeQuit {
			t.Fatalf("expected TypeQuit, got %d", ev.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out")
	}
	exec.Stop()
}

func TestDispatcher_CustomHandler(t *testing.T) {
	d := NewDispatcher()
	var got atomic.Bool
	d.OnCustom(func(e Event) bool {
		if e.Data == "test" {
			got.Store(true)
		}
		return true
	})

	d.Dispatch(Event{Type: TypeCustom, Data: "test"})
	if !got.Load() {
		t.Fatal("custom handler should have been called")
	}
}
