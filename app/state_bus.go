package app

import (
	"sync"
	"sync/atomic"
)

// ─── StateBus: Type-Safe Pub/Sub ───
//
// StateBus replaces bubbletea's Msg/Cmd pattern. Instead of:
//
//	type agentDoneMsg struct{ ... }
//	// ... 136 more msg types
//	func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	    switch msg := msg.(type) {
//	    case agentDoneMsg: m.handleAgentDone(msg) // case 1
//	    case streamMsg:    m.handleStream(msg)    // case 2
//	    // ... 134 more cases
//	    }
//	}
//
// Components subscribe in their constructor and publish on state change:
//
//	bus.Subscribe[AgentDoneEvent]("agent.done", func(e AgentDoneEvent) {
//	    panel.SetStatus("idle")
//	})
//	bus.Publish("agent.done", AgentDoneEvent{Duration: 2*time.Second})
//
// Key differences from tea.Msg:
//   - No Msg type definitions needed (payload is a generic type param)
//   - Synchronous delivery (no goroutines, no channels, no Cmd pipeline)
//   - Components only receive topics they care about (no 90-case switch)
//   - Thread-safe; handlers run on the publisher's goroutine
//   - Unsubscribe via returned subscription handle

// SubscriberID is a unique identifier for a subscription.
type SubscriberID uint64

// subscription holds a single topic subscriber.
type subscription struct {
	id   SubscriberID
	fn   func(any)
}

// StateBus is a type-safe publish/subscribe event bus.
// It replaces the Msg/Cmd pattern from bubbletea with direct,
// synchronous event delivery between components.
//
// Thread-safe. Handlers run on the publisher's goroutine — no goroutines
// are spawned. This means handlers should not block (keep them fast).
// If you need async work, start a goroutine inside your handler.
type StateBus struct {
	mu     sync.RWMutex
	nextID  atomic.Uint64
	topics map[string][]*subscription
}

// NewStateBus creates a new StateBus.
func NewStateBus() *StateBus {
	return &StateBus{
		topics: make(map[string][]*subscription),
	}
}

// Subscribe registers a handler for a topic and returns a subscription
// that can be used to Unsubscribe later.
//
// Usage:
//
//	sub := bus.Subscribe("agent.done", func(e AgentDoneEvent) {
//	    fmt.Println("agent finished:", e.Duration)
//	})
//	defer sub.Unsubscribe()
func Subscribe[T any](bus *StateBus, topic string, handler func(T)) *Subscription {
	id := SubscriberID(bus.nextID.Add(1))
	wrapped := func(payload any) {
		// For nil payloads (PublishRaw), pass zero value if T is any, else assert
		if payload == nil {
			var zero T
			handler(zero)
			return
		}
		handler(payload.(T))
	}
	sub := &subscription{id: id, fn: wrapped}

	bus.mu.Lock()
	bus.topics[topic] = append(bus.topics[topic], sub)
	bus.mu.Unlock()

	return &Subscription{bus: bus, topic: topic, id: id}
}

// Publish delivers a payload to all subscribers of a topic.
// Delivery is synchronous — handlers run on the caller's goroutine.
// If a handler panics, it is recovered and the next handler still runs.
func Publish[T any](bus *StateBus, topic string, payload T) {
	bus.mu.RLock()
	subs := bus.topics[topic]
	// Copy the slice so handlers can safely unsubscribe during dispatch
	snapshot := make([]*subscription, len(subs))
	copy(snapshot, subs)
	bus.mu.RUnlock()

	for _, sub := range snapshot {
		func() {
			defer func() { _ = recover() }()
			sub.fn(payload)
		}()
	}
}

// PublishRaw publishes without a typed payload (for events with no data).
func (bus *StateBus) PublishRaw(topic string) {
	bus.mu.RLock()
	subs := bus.topics[topic]
	snapshot := make([]*subscription, len(subs))
	copy(snapshot, subs)
	bus.mu.RUnlock()

	for _, sub := range snapshot {
		func() {
			defer func() { _ = recover() }()
			sub.fn(nil)
		}()
	}
}

// Subscription represents an active subscription that can be cancelled.
type Subscription struct {
	bus   *StateBus
	topic string
	id    SubscriberID
}

// Unsubscribe removes this subscription from the bus.
// Safe to call multiple times — subsequent calls are no-ops.
func (s *Subscription) Unsubscribe() {
	if s == nil || s.bus == nil {
		return
	}
	s.bus.mu.Lock()
	defer s.bus.mu.Unlock()
	subs := s.bus.topics[s.topic]
	for i, sub := range subs {
		if sub.id == s.id {
			s.bus.topics[s.topic] = append(subs[:i], subs[i+1:]...)
			return
		}
	}
}

// SubscriberCount returns the number of subscribers for a topic.
func (bus *StateBus) SubscriberCount(topic string) int {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	return len(bus.topics[topic])
}

// TopicCount returns the number of active topics.
func (bus *StateBus) TopicCount() int {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	return len(bus.topics)
}

// Clear removes all subscriptions for all topics.
func (bus *StateBus) Clear() {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.topics = make(map[string][]*subscription)
}