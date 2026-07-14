package component

import (
	"sync"
	"sync/atomic"
)

// Reactive is a thread-safe reactive value that auto-triggers re-render
// when changed. Inspired by Textual's reactive attributes.
//
// Usage:
//   count := NewReactive(0)
//   count.Set(42)         // triggers onChange + MarkDirty
//   v := count.Get()       // 42
//   count.Watch(func(old, new int) {
//       fmt.Println("changed from", old, "to", new)
//   })
//
// For components, embed ReactiveField and call Set() — the component's
// MarkDirty() is called automatically if the component implements DirtyMarker.
type Reactive[T any] struct {
	mu        sync.RWMutex
	value     T
	watchers  []func(old, new T)
	dirty     *atomic.Bool // optional: set to mark parent dirty
	equalFunc func(a, b T) bool
}

// NewReactive creates a reactive value with the given initial value.
func NewReactive[T any](initial T) *Reactive[T] {
	return &Reactive[T]{value: initial}
}

// NewReactiveWithDirty creates a reactive value that marks the given
// atomic.Bool dirty when the value changes.
func NewReactiveWithDirty[T any](initial T, dirty *atomic.Bool) *Reactive[T] {
	return &Reactive[T]{value: initial, dirty: dirty}
}

// Get returns the current value (thread-safe).
func (r *Reactive[T]) Get() T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.value
}

// Set updates the value and triggers watchers + dirty flag.
// If the value is unchanged (per Equal or == comparison), no watchers fire.
func (r *Reactive[T]) Set(newVal T) {
	r.mu.Lock()
	oldVal := r.value
	if r.equalFunc != nil {
		if r.equalFunc(oldVal, newVal) {
			r.mu.Unlock()
			return // no change
		}
	} else {
		// Use generic comparison — works for comparable types
		if any(oldVal) == any(newVal) {
			r.mu.Unlock()
			return // no change
		}
	}
	r.value = newVal
	watchers := make([]func(old, new T), len(r.watchers))
	copy(watchers, r.watchers)
	dirty := r.dirty
	r.mu.Unlock()

	// Fire watchers outside lock
	for _, w := range watchers {
		w(oldVal, newVal)
	}

	// Mark dirty
	if dirty != nil {
		dirty.Store(true)
	}
}

// Watch registers a callback that fires when the value changes.
// Returns an unwatch function to deregister.
func (r *Reactive[T]) Watch(fn func(old, new T)) func() {
	r.mu.Lock()
	r.watchers = append(r.watchers, fn)
	idx := len(r.watchers) - 1
	r.mu.Unlock()

	return func() {
		r.mu.Lock()
		if idx < len(r.watchers) {
			r.watchers = append(r.watchers[:idx], r.watchers[idx+1:]...)
		}
		r.mu.Unlock()
	}
}

// SetEqualFunc sets a custom equality function. If nil, == comparison is used.
func (r *Reactive[T]) SetEqualFunc(fn func(a, b T) bool) {
	r.mu.Lock()
	r.equalFunc = fn
	r.mu.Unlock()
}

// SetDirty connects this reactive to an atomic.Bool that gets set to true
// when the value changes.
func (r *Reactive[T]) SetDirty(dirty *atomic.Bool) {
	r.mu.Lock()
	r.dirty = dirty
	r.mu.Unlock()
}

// ReactiveInt is a convenience type for reactive integers.
type ReactiveInt = Reactive[int]

// ReactiveString is a convenience type for reactive strings.
type ReactiveString = Reactive[string]

// ReactiveBool is a convenience type for reactive booleans.
type ReactiveBool = Reactive[bool]

// NewReactiveInt creates a reactive integer.
func NewReactiveInt(initial int) *ReactiveInt {
	return NewReactive[int](initial)
}

// NewReactiveString creates a reactive string.
func NewReactiveString(initial string) *ReactiveString {
	return NewReactive[string](initial)
}

// NewReactiveBool creates a reactive boolean.
func NewReactiveBool(initial bool) *ReactiveBool {
	return NewReactive[bool](initial)
}

// ReactiveField embeds a Reactive value into a component, automatically
// connecting it to the component's dirty flag.
//
// Usage in a component:
//   type Counter struct {
//       BaseComponent
//       Count *ReactiveField[int]
//   }
//   func NewCounter() *Counter {
//       c := &Counter{}
//       c.Count = NewReactiveField[int](0, &c.dirty)
//       return c
//   }
type ReactiveField[T any] struct {
	*Reactive[T]
}

// NewReactiveField creates a reactive field connected to a dirty flag.
func NewReactiveField[T any](initial T, dirty *atomic.Bool) *ReactiveField[T] {
	return &ReactiveField[T]{
		Reactive: NewReactiveWithDirty(initial, dirty),
	}
}

// Bind connects a reactive field to a different dirty flag (e.g. after
// the component is added to an App).
func (rf *ReactiveField[T]) Bind(dirty *atomic.Bool) {
	rf.SetDirty(dirty)
}

// Computed is a read-only reactive value derived from other reactives.
// It recomputes when any dependency changes.
//
// Usage:
//   a := NewReactiveInt(2)
//   b := NewReactiveInt(3)
//   sum := Computed(func() int { return a.Get() + b.Get() })
//   sum.Watch(func(_, new int) { fmt.Println("sum:", new) })
//   a.Set(10) // triggers sum watcher: "sum: 13"
type ComputedValue[T any] struct {
	mu        sync.RWMutex
	value     T
	compute   func() T
	watchers  []func(newVal T)
	dirty     *atomic.Bool
}

// ComputedFrom creates a computed from typed dependencies.
// Usage: ComputedFrom(func() int { return a.Get() + b.Get() }, a, b)
func ComputedFrom[T any, D any](compute func() T, deps ...*Reactive[D]) *ComputedValue[T] {
	c := &ComputedValue[T]{
		compute: compute,
	}
	c.recompute()

	for _, dep := range deps {
		dep.Watch(func(_, _ D) {
			c.recomputeAndNotify()
		})
	}

	return c
}

func (c *ComputedValue[T]) recompute() {
	c.value = c.compute()
}

func (c *ComputedValue[T]) recomputeAndNotify() {
	c.mu.Lock()
	oldVal := c.value
	c.value = c.compute()
	newVal := c.value
	watchers := make([]func(newVal T), len(c.watchers))
	copy(watchers, c.watchers)
	dirty := c.dirty
	c.mu.Unlock()

	// Only notify if value changed
	if any(oldVal) == any(newVal) {
		return
	}

	for _, w := range watchers {
		w(newVal)
	}

	if dirty != nil {
		dirty.Store(true)
	}
}

// Get returns the current computed value.
func (c *ComputedValue[T]) Get() T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

// Watch registers a callback that fires when the computed value changes.
func (c *ComputedValue[T]) Watch(fn func(newVal T)) func() {
	c.mu.Lock()
	c.watchers = append(c.watchers, fn)
	idx := len(c.watchers) - 1
	c.mu.Unlock()

	return func() {
		c.mu.Lock()
		if idx < len(c.watchers) {
			c.watchers = append(c.watchers[:idx], c.watchers[idx+1:]...)
		}
		c.mu.Unlock()
	}
}

// SetDirty connects this computed to a dirty flag.
func (c *ComputedValue[T]) SetDirty(dirty *atomic.Bool) {
	c.mu.Lock()
	c.dirty = dirty
	c.mu.Unlock()
}