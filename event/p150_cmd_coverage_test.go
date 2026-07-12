package event

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestP150_Batch(t *testing.T) {
	var count int32
	cmd1 := func() Event { atomic.AddInt32(&count, 1); return Event{Type: TypeCustom, Data: "1"} }
	cmd2 := func() Event { atomic.AddInt32(&count, 1); return Event{Type: TypeCustom, Data: "2"} }
	cmd3 := func() Event { atomic.AddInt32(&count, 1); return Event{Type: TypeCustom, Data: "3"} }

	batch := Batch(cmd1, cmd2, cmd3)
	ev := batch()
	if ev.Type != 0 {
		t.Errorf("expected no-op event, got %+v", ev)
	}
	if atomic.LoadInt32(&count) != 3 {
		t.Errorf("expected 3 cmds executed, got %d", count)
	}
}

func TestP150_BatchWithNil(t *testing.T) {
	var count int32
	cmd1 := func() Event { atomic.AddInt32(&count, 1); return Event{} }
	batch := Batch(cmd1, nil, nil)
	batch()
	if atomic.LoadInt32(&count) != 1 {
		t.Errorf("expected 1 cmd executed, got %d", count)
	}
}

func TestP150_BatchEmpty(t *testing.T) {
	batch := Batch()
	ev := batch()
	if ev.Type != 0 {
		t.Errorf("expected no-op event, got %+v", ev)
	}
}

func TestP150_Sequence(t *testing.T) {
	var order []int
	var mu sync.Mutex
	cmd1 := func() Event { mu.Lock(); order = append(order, 1); mu.Unlock(); return Event{Type: TypeCustom, Data: "1"} }
	cmd2 := func() Event { mu.Lock(); order = append(order, 2); mu.Unlock(); return Event{Type: TypeCustom, Data: "2"} }
	cmd3 := func() Event { mu.Lock(); order = append(order, 3); mu.Unlock(); return Event{Type: TypeCustom, Data: "3"} }

	seq := Sequence(cmd1, cmd2, cmd3)
	ev := seq()
	if ev.Data != "3" {
		t.Errorf("expected last event data '3', got %q", ev.Data)
	}
	if len(order) != 3 || order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Errorf("expected order [1,2,3], got %v", order)
	}
}

func TestP150_SequenceWithNil(t *testing.T) {
	var count int32
	cmd1 := func() Event { atomic.AddInt32(&count, 1); return Event{} }
	seq := Sequence(cmd1, nil, nil)
	ev := seq()
	if ev.Type != 0 {
		t.Errorf("expected empty event, got %+v", ev)
	}
	if atomic.LoadInt32(&count) != 1 {
		t.Errorf("expected 1 cmd, got %d", count)
	}
}

func TestP150_SequenceEmpty(t *testing.T) {
	seq := Sequence()
	ev := seq()
	if ev.Type != 0 {
		t.Errorf("expected empty event, got %+v", ev)
	}
}

func TestP150_Tick(t *testing.T) {
	tick := Tick(10*time.Millisecond, func(t time.Time) Event {
		return Event{Type: TypeCustom, Data: "ticked"}
	})
	ev := tick()
	if ev.Data != "ticked" {
		t.Errorf("expected 'ticked', got %q", ev.Data)
	}
}

func TestP150_BatchWithExecutor(t *testing.T) {
	loop := newBlockingLoop(NewDispatcher())
	defer loop.Quit()
	ce := NewCmdExecutor(loop)

	var count int32
	cmd1 := func() Event { atomic.AddInt32(&count, 1); return Event{Type: TypeCustom, Data: "1"} }
	cmd2 := func() Event { atomic.AddInt32(&count, 1); return Event{Type: TypeCustom, Data: "2"} }

	BatchWithExecutor(ce, cmd1, cmd2)
	time.Sleep(50 * time.Millisecond)
	ce.Stop()
	if atomic.LoadInt32(&count) != 2 {
		t.Errorf("expected 2 cmds, got %d", count)
	}
}

func TestP150_SequenceWithExecutor(t *testing.T) {
	loop := newBlockingLoop(NewDispatcher())
	defer loop.Quit()
	ce := NewCmdExecutor(loop)

	var count int32
	cmd1 := func() Event { atomic.AddInt32(&count, 1); return Event{Type: TypeCustom, Data: "1"} }
	cmd2 := func() Event { atomic.AddInt32(&count, 1); return Event{Type: TypeCustom, Data: "2"} }

	SequenceWithExecutor(ce, cmd1, cmd2)
	ce.Stop()
	// Give a moment for goroutine to finish
	time.Sleep(10 * time.Millisecond)
	if atomic.LoadInt32(&count) != 2 {
		t.Errorf("expected 2 cmds, got %d", count)
	}
}

func TestP150_TickWithExecutor_Fires(t *testing.T) {
	loop := newBlockingLoop(NewDispatcher())
	defer loop.Quit()
	ce := NewCmdExecutor(loop)

	var fired int32
	TickWithExecutor(ce, 10*time.Millisecond, func(t time.Time) Event {
		atomic.StoreInt32(&fired, 1)
		return Event{Type: TypeCustom, Data: "tick"}
	})
	time.Sleep(50 * time.Millisecond)
	ce.Stop()
	if atomic.LoadInt32(&fired) != 1 {
		t.Errorf("expected tick to fire, got %d", fired)
	}
}

func TestP150_TickWithExecutor_Cancelled(t *testing.T) {
	loop := newBlockingLoop(NewDispatcher())
	defer loop.Quit()
	ce := NewCmdExecutor(loop)

	var fired int32
	TickWithExecutor(ce, 1*time.Second, func(t time.Time) Event {
		atomic.StoreInt32(&fired, 1)
		return Event{Type: TypeCustom, Data: "tick"}
	})
	// Stop before timer fires
	ce.Stop()
	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt32(&fired) != 0 {
		t.Errorf("expected tick NOT to fire after cancel, got %d", fired)
	}
}

func TestP150_LoopCmdExecutor(t *testing.T) {
	loop := newBlockingLoop(NewDispatcher())
	defer loop.Quit()
	ce := loop.CmdExecutor()
	if ce == nil {
		t.Fatal("expected non-nil CmdExecutor")
	}
}

func TestP150_LoopExec(t *testing.T) {
	loop := newBlockingLoop(NewDispatcher())
	defer loop.Quit()

	var count int32
	loop.Exec(func() Event {
		atomic.AddInt32(&count, 1)
		return Event{Type: TypeCustom, Data: "exec"}
	})
	// Give goroutine time to run
	time.Sleep(20 * time.Millisecond)
	if atomic.LoadInt32(&count) != 1 {
		t.Errorf("expected 1 exec, got %d", count)
	}
}

func TestP150_CmdExecutor_StopIdempotent(t *testing.T) {
	loop := newBlockingLoop(NewDispatcher())
	defer loop.Quit()
	ce := NewCmdExecutor(loop)
	ce.Stop()
	ce.Stop() // should not panic
}

func TestP150_CmdExec_PanicRecovery(t *testing.T) {
	loop := newBlockingLoop(NewDispatcher())
	defer loop.Quit()
	ce := NewCmdExecutor(loop)

	ce.Exec(func() Event {
		panic("test panic")
	})
	// Give goroutine time to run
	time.Sleep(20 * time.Millisecond)
	// Should not hang — panic recovered
	ce.Stop()
}

func TestP150_CmdExec_AfterStop(t *testing.T) {
	loop := newBlockingLoop(NewDispatcher())
	defer loop.Quit()
	ce := NewCmdExecutor(loop)
	ce.Stop()

	var count int32
	ce.Exec(func() Event {
		atomic.AddInt32(&count, 1)
		return Event{}
	})
	// Should not execute since executor is stopped
	time.Sleep(20 * time.Millisecond)
	if atomic.LoadInt32(&count) != 0 {
		t.Errorf("expected 0 execs after stop, got %d", count)
	}
}