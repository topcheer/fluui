package app

import (
	"sync"
	"testing"
)

// ─── StateBus Tests ───

func TestStateBus_BasicPubSub(t *testing.T) {
	bus := NewStateBus()
	defer bus.Clear()

	type TestEvent struct{ Value int }

	var received TestEvent
	var mu sync.Mutex

	sub := Subscribe[TestEvent](bus, "test.event", func(e TestEvent) {
		mu.Lock()
		received = e
		mu.Unlock()
	})
	defer sub.Unsubscribe()

	Publish(bus, "test.event", TestEvent{Value: 42})

	if received.Value != 42 {
		t.Errorf("expected 42, got %d", received.Value)
	}
}

func TestStateBus_MultipleSubscribers(t *testing.T) {
	bus := NewStateBus()
	defer bus.Clear()

	var count1, count2 int

	s1 := Subscribe[int](bus, "counter", func(v int) { count1 = v })
	defer s1.Unsubscribe()
	s2 := Subscribe[int](bus, "counter", func(v int) { count2 = v })
	defer s2.Unsubscribe()

	Publish(bus, "counter", 99)

	if count1 != 99 || count2 != 99 {
		t.Errorf("subscribers didn't both receive: %d, %d", count1, count2)
	}
}

func TestStateBus_Unsubscribe(t *testing.T) {
	bus := NewStateBus()
	defer bus.Clear()

	var count int
	sub := Subscribe[int](bus, "inc", func(v int) { count += v })
	sub.Unsubscribe()

	Publish(bus, "inc", 10)

	if count != 0 {
		t.Errorf("unsubscribed handler still received event: %d", count)
	}
}

func TestStateBus_DifferentTopics(t *testing.T) {
	bus := NewStateBus()
	defer bus.Clear()

	var gotA, gotB bool

	Subscribe[struct{}](bus, "a", func(_ struct{}) { gotA = true })
	Subscribe[struct{}](bus, "b", func(_ struct{}) { gotB = true })

	Publish(bus, "a", struct{}{})

	if !gotA {
		t.Error("topic A not received")
	}
	if gotB {
		t.Error("topic B should not receive topic A's event")
	}
}

func TestStateBus_SubscriberCount(t *testing.T) {
	bus := NewStateBus()
	defer bus.Clear()

	if bus.SubscriberCount("x") != 0 {
		t.Error("expected 0 subscribers")
	}

	s1 := Subscribe[int](bus, "x", func(_ int) {})
	defer s1.Unsubscribe()

	if bus.SubscriberCount("x") != 1 {
		t.Errorf("expected 1 subscriber, got %d", bus.SubscriberCount("x"))
	}

	s2 := Subscribe[int](bus, "x", func(_ int) {})
	defer s2.Unsubscribe()

	if bus.SubscriberCount("x") != 2 {
		t.Errorf("expected 2 subscribers, got %d", bus.SubscriberCount("x"))
	}

	s1.Unsubscribe()

	if bus.SubscriberCount("x") != 1 {
		t.Errorf("expected 1 after unsubscribe, got %d", bus.SubscriberCount("x"))
	}
}

func TestStateBus_PublishRaw(t *testing.T) {
	bus := NewStateBus()
	defer bus.Clear()

	var received bool

	Subscribe[any](bus, "ping", func(_ any) { received = true })

	bus.PublishRaw("ping")

	if !received {
		t.Error("PublishRaw didn't deliver")
	}
}

func TestStateBus_PanicRecovery(t *testing.T) {
	bus := NewStateBus()
	defer bus.Clear()

	var secondCalled bool

	Subscribe[int](bus, "danger", func(_ int) { panic("boom") })
	Subscribe[int](bus, "danger", func(_ int) { secondCalled = true })

	// Should not crash
	Publish(bus, "danger", 1)

	if !secondCalled {
		t.Error("second handler should still run after first panics")
	}
}

func TestStateBus_TopicCount(t *testing.T) {
	bus := NewStateBus()
	defer bus.Clear()

	Subscribe[int](bus, "a", func(_ int) {})
	Subscribe[int](bus, "b", func(_ int) {})
	Subscribe[int](bus, "c", func(_ int) {})

	if bus.TopicCount() != 3 {
		t.Errorf("expected 3 topics, got %d", bus.TopicCount())
	}
}

func TestStateBus_UnsubscribeTwice(t *testing.T) {
	bus := NewStateBus()

	sub := Subscribe[int](bus, "x", func(_ int) {})
	sub.Unsubscribe()
	sub.Unsubscribe() // should not panic

	if bus.SubscriberCount("x") != 0 {
		t.Error("should be 0 after unsubscribe")
	}
}

func TestStateBus_ConcurrentAccess(t *testing.T) {
	bus := NewStateBus()
	defer bus.Clear()

	var wg sync.WaitGroup
	var counter int
	var mu sync.Mutex

	// Concurrent subscribers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sub := Subscribe[int](bus, "concurrent", func(v int) {
				mu.Lock()
				counter += v
				mu.Unlock()
			})
			_ = sub
		}()
	}
	wg.Wait()

	// Concurrent publish
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Publish(bus, "concurrent", 1)
		}()
	}
	wg.Wait()

	if counter == 0 {
		t.Error("concurrent access test: counter should be > 0")
	}
}