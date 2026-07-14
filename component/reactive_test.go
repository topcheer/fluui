package component

import (
	"sync/atomic"
	"testing"
)

func TestReactive_BasicGetSet(t *testing.T) {
	r := NewReactiveInt(10)
	if r.Get() != 10 {
		t.Fatal("initial value should be 10")
	}
	r.Set(20)
	if r.Get() != 20 {
		t.Fatal("value should be 20 after Set")
	}
}

func TestReactive_NoChangeNoWatch(t *testing.T) {
	r := NewReactiveInt(5)
	called := false
	r.Watch(func(old, new int) {
		called = true
	})
	r.Set(5) // same value
	if called {
		t.Fatal("watcher should not fire when value unchanged")
	}
}

func TestReactive_WatchFires(t *testing.T) {
	r := NewReactiveInt(1)
	events := []int{}
	r.Watch(func(old, new int) {
		events = append(events, new)
	})
	r.Set(2)
	r.Set(3)
	r.Set(4)
	if len(events) != 3 || events[0] != 2 || events[1] != 3 || events[2] != 4 {
		t.Fatalf("expected [2,3,4], got %v", events)
	}
}

func TestReactic_Unwatch(t *testing.T) {
	r := NewReactiveInt(0)
	called := 0
	unwatch := r.Watch(func(old, new int) {
		called++
	})
	r.Set(1) // fires
	unwatch()
	r.Set(2) // should not fire
	if called != 1 {
		t.Fatalf("expected 1 call, got %d", called)
	}
}

func TestReactive_DirtyFlag(t *testing.T) {
	var dirty atomic.Bool
	r := NewReactiveWithDirty(0, &dirty)
	if dirty.Load() {
		t.Fatal("should start clean")
	}
	r.Set(1)
	if !dirty.Load() {
		t.Fatal("should be dirty after Set")
	}
	dirty.Store(false)
	r.Set(1) // no change
	if dirty.Load() {
		t.Fatal("should not be dirty when value unchanged")
	}
}

func TestReactive_MultipleWatchers(t *testing.T) {
	r := NewReactiveInt(0)
	c1, c2 := 0, 0
	r.Watch(func(old, new int) { c1++ })
	r.Watch(func(old, new int) { c2++ })
	r.Set(1)
	if c1 != 1 || c2 != 1 {
		t.Fatalf("expected both watchers called once, got %d/%d", c1, c2)
	}
}

func TestReactive_SetDirtyAfterCreation(t *testing.T) {
	r := NewReactiveInt(0)
	var dirty atomic.Bool
	r.SetDirty(&dirty)
	r.Set(1)
	if !dirty.Load() {
		t.Fatal("should be dirty after Set with late-bound dirty flag")
	}
}

func TestReactiveString(t *testing.T) {
	r := NewReactiveString("hello")
	if r.Get() != "hello" {
		t.Fatal("initial should be hello")
	}
	r.Set("world")
	if r.Get() != "world" {
		t.Fatal("should be world after Set")
	}
}

func TestReactiveBool(t *testing.T) {
	r := NewReactiveBool(false)
	if r.Get() {
		t.Fatal("initial should be false")
	}
	r.Set(true)
	if !r.Get() {
		t.Fatal("should be true after Set")
	}
}

func TestReactive_EqualFunc(t *testing.T) {
	type person struct{ name string; age int }
	r := NewReactive(person{name: "A", age: 30})
	r.SetEqualFunc(func(a, b person) bool {
		return a.name == b.name // only compare name
	})

	called := false
	r.Watch(func(old, new person) {
		called = true
	})

	r.Set(person{name: "A", age: 31}) // same name, different age
	if called {
		t.Fatal("should not fire when only non-compared field changes")
	}

	r.Set(person{name: "B", age: 31}) // different name
	if !called {
		t.Fatal("should fire when compared field changes")
	}
}

func TestReactiveField(t *testing.T) {
	var dirty atomic.Bool
	rf := NewReactiveField(42, &dirty)
	if rf.Get() != 42 {
		t.Fatal("initial should be 42")
	}
	rf.Set(100)
	if rf.Get() != 100 {
		t.Fatal("should be 100 after Set")
	}
	if !dirty.Load() {
		t.Fatal("dirty flag should be set")
	}
}

func TestReactiveField_Bind(t *testing.T) {
	rf := NewReactiveField(0, nil) // no dirty initially
	var dirty atomic.Bool
	rf.Bind(&dirty)
	rf.Set(1)
	if !dirty.Load() {
		t.Fatal("should be dirty after Bind + Set")
	}
}

func TestReactive_Concurrent(t *testing.T) {
	r := NewReactiveInt(0)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			r.Set(i)
		}
		close(done)
	}()
	for i := 0; i < 100; i++ {
		_ = r.Get()
	}
	<-done
}

func TestComputed_Basic(t *testing.T) {
	a := NewReactiveInt(2)
	b := NewReactiveInt(3)

	sum := ComputedFrom(func() int { return a.Get() + b.Get() }, a, b)
	if sum.Get() != 5 {
		t.Fatalf("expected 5, got %d", sum.Get())
	}

	a.Set(10)
	if sum.Get() != 13 {
		t.Fatalf("expected 13, got %d", sum.Get())
	}
}

func TestComputed_Watch(t *testing.T) {
	a := NewReactiveInt(1)
	b := NewReactiveInt(1)

	sum := ComputedFrom(func() int { return a.Get() + b.Get() }, a, b)

	notifications := []int{}
	sum.Watch(func(newVal int) {
		notifications = append(notifications, newVal)
	})

	a.Set(5) // sum: 6
	b.Set(10) // sum: 15

	if len(notifications) != 2 || notifications[0] != 6 || notifications[1] != 15 {
		t.Fatalf("expected [6,15], got %v", notifications)
	}
}

func TestComputed_NoChangeNoNotify(t *testing.T) {
	a := NewReactiveInt(2)
	b := NewReactiveInt(3)

	sum := ComputedFrom(func() int { return a.Get() + b.Get() }, a, b)

	called := 0
	sum.Watch(func(newVal int) { called++ })

	a.Set(1) // sum: 4 → notify
	b.Set(3) // sum: 4 → no change, no notify

	if called != 1 {
		t.Fatalf("expected 1 notification, got %d", called)
	}
}

func TestComputed_Dirty(t *testing.T) {
	a := NewReactiveInt(1)
	b := NewReactiveInt(2)

	sum := ComputedFrom(func() int { return a.Get() + b.Get() }, a, b)

	var dirty atomic.Bool
	sum.SetDirty(&dirty)

	a.Set(10)
	if !dirty.Load() {
		t.Fatal("should be dirty after dependency change")
	}
}

func TestComputed_Unwatch(t *testing.T) {
	a := NewReactiveInt(1)
	sum := ComputedFrom(func() int { return a.Get() * 2 }, a)

	called := 0
	unwatch := sum.Watch(func(newVal int) { called++ })

	a.Set(2) // notify
	unwatch()
	a.Set(3) // should not notify

	if called != 1 {
		t.Fatalf("expected 1 call, got %d", called)
	}
}