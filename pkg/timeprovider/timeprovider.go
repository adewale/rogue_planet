// Package timeprovider provides an abstraction over time operations for testability.
//
// Production code should use WallClock which returns the actual system time.
// Test code should use FakeClock which allows controlling time for deterministic tests.
package timeprovider

import (
	"sync"
	"time"
)

// TimeProvider abstracts time operations to enable deterministic testing.
//
// In production, use WallClock to get actual system time.
// In tests, use FakeClock to control time and avoid race conditions.
type TimeProvider interface {
	// Now returns the current time according to this provider.
	Now() time.Time

	// Since returns the time elapsed since t according to this provider.
	// It is shorthand for Now().Sub(t).
	Since(t time.Time) time.Duration
}

// WallClock provides actual system time using the standard time package.
//
// This is the production implementation that should be used in real applications.
type WallClock struct{}

// Now returns the current system time.
func (w WallClock) Now() time.Time {
	return time.Now()
}

// Since returns the time elapsed since t using the current system time.
func (w WallClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}

// FakeClock provides controllable time for testing.
//
// FakeClock allows tests to:
//   - Set a specific time
//   - Advance time by arbitrary durations
//   - Test time-dependent logic without race conditions
//   - Run tests instantly without waiting for real time
//
// FakeClock is safe for concurrent use.
type FakeClock struct {
	mu      sync.RWMutex
	current time.Time
}

// NewFakeClock creates a FakeClock initialized to the given time.
func NewFakeClock(t time.Time) *FakeClock {
	return &FakeClock{
		current: t,
	}
}

// Now returns the current fake time.
func (f *FakeClock) Now() time.Time {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.current
}

// Since returns the time elapsed since t according to the fake clock.
func (f *FakeClock) Since(t time.Time) time.Duration {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.current.Sub(t)
}

// SetTime sets the fake clock to a specific time.
//
// This is useful for testing specific moments in time.
func (f *FakeClock) SetTime(t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.current = t
}

// Advance moves the fake clock forward by the given duration.
//
// This is useful for testing timeouts, expirations, and time-based logic.
func (f *FakeClock) Advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.current = f.current.Add(d)
}
