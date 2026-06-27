package animation

import (
	"sync"
	"time"
)

// Animation is the interface for all animations.
type Animation interface {
	// Update advances the animation by delta and returns true when the
	// animation has finished.
	Update(delta time.Duration) bool
	// Done returns a channel that is closed when the animation finishes.
	Done() <-chan struct{}
}

// Manager manages all active animations.
type Manager struct {
	mu         sync.Mutex
	animations []Animation
	ticker     *time.Ticker
	onDirty    func()
	quitCh     chan struct{}
	quitOnce   sync.Once
	interval   time.Duration
}

// NewManager creates a new animation manager that ticks at the given fps.
// onDirty is called whenever an animation produces a visible change and
// the screen should be redrawn.
func NewManager(fps int, onDirty func()) *Manager {
	if fps <= 0 {
		fps = 60
	}
	interval := time.Duration(float64(time.Second) / float64(fps))
	return &Manager{
		onDirty:  onDirty,
		quitCh:   make(chan struct{}),
		interval: interval,
	}
}

// Start launches the background ticker goroutine.
func (m *Manager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ticker != nil {
		return
	}
	m.ticker = time.NewTicker(m.interval)
	tickerC := m.ticker.C // capture channel so run goroutine never races with Stop
	go m.run(tickerC)
}

// Stop halts the background ticker and removes finished animations.
// It is safe to call multiple times.
func (m *Manager) Stop() {
	m.quitOnce.Do(func() {
		close(m.quitCh)
	})
	m.mu.Lock()
	if m.ticker != nil {
		m.ticker.Stop()
	}
	m.mu.Unlock()
}

// Add registers a new animation.
func (m *Manager) Add(a Animation) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.animations = append(m.animations, a)
}

// Tick manually advances every animation by one frame interval. This is
// primarily used for testing.
func (m *Manager) Tick() {
	m.advance(m.interval)
}

// run is the background loop driven by the ticker. It receives the ticker
// channel directly so it never races with Stop modifying the Manager state.
func (m *Manager) run(tickerC <-chan time.Time) {
	for {
		select {
		case <-m.quitCh:
			return
		case <-tickerC:
			m.advance(m.interval)
		}
	}
}

// advance updates all animations, removes finished ones, and notifies onDirty
// if any animation is still active.
func (m *Manager) advance(delta time.Duration) {
	m.mu.Lock()
	// Snapshot slice to avoid holding the lock during Update calls.
	snapshot := make([]Animation, len(m.animations))
	copy(snapshot, m.animations)
	m.mu.Unlock()

	var stillActive bool
	for _, a := range snapshot {
		if !a.Update(delta) {
			stillActive = true
		}
	}

	// Remove finished animations.
	m.mu.Lock()
	alive := m.animations[:0]
	for _, a := range m.animations {
		select {
		case <-a.Done():
			continue
		default:
			alive = append(alive, a)
		}
	}
	m.animations = alive
	m.mu.Unlock()

	if stillActive && m.onDirty != nil {
		m.onDirty()
	}
}
