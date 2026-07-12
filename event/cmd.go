package event

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Cmd is a declarative asynchronous operation. When executed by the Loop,
// it runs in a goroutine and the returned Event is injected back into the loop.
// A nil Cmd is a no-op.
type Cmd func() Event

// Msg is an alias for Event — the result of an async Cmd execution.
// Users create custom Event values with TypeCustom to carry arbitrary data.
type Msg = Event

// TypeCustom is the EventType for Cmd/Msg results. Users attach data via Data field.
// Extend EventType to include this value.
const TypeCustom EventType = 99

// CmdExecutor manages the lifecycle of async Cmd goroutines.
// It ensures all goroutines are cancelled when the Loop stops.
type CmdExecutor struct {
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	loop   *Loop
}

// NewCmdExecutor creates a new executor bound to the given loop.
func NewCmdExecutor(loop *Loop) *CmdExecutor {
	ctx, cancel := context.WithCancel(context.Background())
	return &CmdExecutor{
		ctx:    ctx,
		cancel: cancel,
		loop:   loop,
	}
}

// Exec runs a Cmd in a goroutine. The returned Event is sent back to the loop.
// If the executor has been cancelled (loop stopped), the Cmd is not executed.
// Panics in Cmd are recovered to prevent goroutine leaks.
func (ce *CmdExecutor) Exec(cmd Cmd) {
	if cmd == nil {
		return
	}

	ce.mu.Lock()
	ctx := ce.ctx
	ce.mu.Unlock()

	if ctx.Err() != nil {
		return // executor cancelled
	}

	ce.wg.Add(1)
	go func() {
		defer ce.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				// Send error event on panic
				ce.loop.Send(Event{
					Type: TypeCustom,
					Data: fmt.Sprintf("cmd panic: %v", r),
				})
			}
		}()

		// Check cancellation before executing
		select {
		case <-ctx.Done():
			return
		default:
		}

		ev := cmd()
		if ev.Type != 0 || ev.Data != "" {
			ce.loop.Send(ev)
		}
	}()
}

// Stop cancels all pending Cmd goroutines and waits for them to finish.
func (ce *CmdExecutor) Stop() {
	ce.mu.Lock()
	ce.cancel()
	ce.mu.Unlock()
	ce.wg.Wait()
}

// Batch creates a Cmd that runs multiple Cmds in parallel.
// The results are collected and sent as individual Events.
func Batch(cmds ...Cmd) Cmd {
	return func() Event {
		var wg sync.WaitGroup
		for _, cmd := range cmds {
			if cmd == nil {
				continue
			}
			wg.Add(1)
			go func(c Cmd) {
				defer wg.Done()
				_ = c() // result is sent by executor, not here
			}(cmd)
		}
		wg.Wait()
		return Event{} // no-op event; individual results already sent
	}
}

// BatchWithExecutor runs multiple Cmds in parallel using the given executor.
// Each Cmd's result Event is sent back to the loop independently.
func BatchWithExecutor(ce *CmdExecutor, cmds ...Cmd) {
	for _, cmd := range cmds {
		ce.Exec(cmd)
	}
}

// Sequence creates a Cmd that runs multiple Cmds serially.
// Each Cmd runs only after the previous one completes.
// Returns the Event from the last Cmd.
func Sequence(cmds ...Cmd) Cmd {
	return func() Event {
		var last Event
		for _, cmd := range cmds {
			if cmd == nil {
				continue
			}
			last = cmd()
		}
		return last
	}
}

// SequenceWithExecutor runs Cmds serially, each result sent to the loop.
func SequenceWithExecutor(ce *CmdExecutor, cmds ...Cmd) {
	go func() {
		for _, cmd := range cmds {
			if cmd == nil {
				continue
			}
			ev := cmd()
			if ev.Type != 0 || ev.Data != "" {
				ce.loop.Send(ev)
			}
		}
	}()
}

// Tick creates a Cmd that fires once after the given duration.
// The function f receives the current time and returns an Event to send.
func Tick(d time.Duration, f func(time.Time) Event) Cmd {
	return func() Event {
		time.Sleep(d)
		return f(time.Now())
	}
}

// TickWithExecutor schedules a one-shot timer using the executor's context.
// If the loop stops before the timer fires, the Cmd is cancelled.
func TickWithExecutor(ce *CmdExecutor, d time.Duration, f func(time.Time) Event) {
	ce.wg.Add(1)
	go func() {
		defer ce.wg.Done()
		timer := time.NewTimer(d)
		defer timer.Stop()
		select {
		case <-ce.ctx.Done():
			return
		case t := <-timer.C:
			ev := f(t)
			if ev.Type != 0 || ev.Data != "" {
				ce.loop.Send(ev)
			}
		}
	}()
}

// Every creates a Cmd that fires repeatedly at the given interval.
// The function f receives the tick time and returns an Event to send.
// The Cmd runs forever (until loop stops) — use with Exec or TickWithExecutor.
// Returns immediately after scheduling; each tick sends an Event.
func Every(ce *CmdExecutor, interval time.Duration, f func(time.Time) Event) {
	ce.wg.Add(1)
	go func() {
		defer ce.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ce.ctx.Done():
				return
			case t := <-ticker.C:
				ev := f(t)
				if ev.Type != 0 || ev.Data != "" {
					ce.loop.Send(ev)
				}
			}
		}
	}()
}

// Quit returns a Cmd that signals the loop to quit.
func Quit() Cmd {
	return func() Event {
		return Event{Type: TypeQuit}
	}
}
