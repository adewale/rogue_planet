package timeprovider

import (
	"sync"
	"testing"
	"time"
)

func TestWallClock_Now(t *testing.T) {
	clock := WallClock{}

	before := time.Now()
	actual := clock.Now()
	after := time.Now()

	// WallClock.Now() should return a time between before and after
	if actual.Before(before) || actual.After(after) {
		t.Errorf("WallClock.Now() returned %v, expected between %v and %v", actual, before, after)
	}
}

func TestWallClock_Since(t *testing.T) {
	clock := WallClock{}

	start := time.Now()
	time.Sleep(10 * time.Millisecond)
	elapsed := clock.Since(start)

	// Should have elapsed at least 10ms
	if elapsed < 10*time.Millisecond {
		t.Errorf("WallClock.Since() = %v, expected >= 10ms", elapsed)
	}

	// Should not have elapsed more than 100ms (generous margin)
	if elapsed > 100*time.Millisecond {
		t.Errorf("WallClock.Since() = %v, expected < 100ms", elapsed)
	}
}

func TestFakeClock_Now(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	clock := NewFakeClock(fixedTime)

	// Should return the exact time we set
	if got := clock.Now(); !got.Equal(fixedTime) {
		t.Errorf("FakeClock.Now() = %v, want %v", got, fixedTime)
	}

	// Should return same time on repeated calls (frozen time)
	time.Sleep(10 * time.Millisecond)
	if got := clock.Now(); !got.Equal(fixedTime) {
		t.Errorf("FakeClock.Now() = %v, want %v (time should be frozen)", got, fixedTime)
	}
}

func TestFakeClock_Since(t *testing.T) {
	tests := []struct {
		name     string
		current  time.Time
		since    time.Time
		expected time.Duration
	}{
		{
			name:     "same time",
			current:  time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			since:    time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: 0,
		},
		{
			name:     "1 hour difference",
			current:  time.Date(2025, 1, 1, 13, 0, 0, 0, time.UTC),
			since:    time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: 1 * time.Hour,
		},
		{
			name:     "1 day difference",
			current:  time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC),
			since:    time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: 24 * time.Hour,
		},
		{
			name:     "negative duration (since is in future)",
			current:  time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			since:    time.Date(2025, 1, 1, 13, 0, 0, 0, time.UTC),
			expected: -1 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock := NewFakeClock(tt.current)

			got := clock.Since(tt.since)
			if got != tt.expected {
				t.Errorf("FakeClock.Since() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFakeClock_SetTime(t *testing.T) {
	initialTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	clock := NewFakeClock(initialTime)

	// Verify initial time
	if got := clock.Now(); !got.Equal(initialTime) {
		t.Fatalf("Initial time = %v, want %v", got, initialTime)
	}

	// Set to new time
	newTime := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	clock.SetTime(newTime)

	// Verify time changed
	if got := clock.Now(); !got.Equal(newTime) {
		t.Errorf("After SetTime(), Now() = %v, want %v", got, newTime)
	}
}

func TestFakeClock_Advance(t *testing.T) {
	tests := []struct {
		name     string
		initial  time.Time
		advance  time.Duration
		expected time.Time
	}{
		{
			name:     "advance 1 second",
			initial:  time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			advance:  1 * time.Second,
			expected: time.Date(2025, 1, 1, 12, 0, 1, 0, time.UTC),
		},
		{
			name:     "advance 1 hour",
			initial:  time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			advance:  1 * time.Hour,
			expected: time.Date(2025, 1, 1, 13, 0, 0, 0, time.UTC),
		},
		{
			name:     "advance 1 day",
			initial:  time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			advance:  24 * time.Hour,
			expected: time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "advance negative (go backwards)",
			initial:  time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			advance:  -1 * time.Hour,
			expected: time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC),
		},
		{
			name:     "advance 0 (no change)",
			initial:  time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			advance:  0,
			expected: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock := NewFakeClock(tt.initial)

			clock.Advance(tt.advance)

			if got := clock.Now(); !got.Equal(tt.expected) {
				t.Errorf("After Advance(%v), Now() = %v, want %v", tt.advance, got, tt.expected)
			}
		})
	}
}

func TestFakeClock_MultipleAdvances(t *testing.T) {
	initialTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	clock := NewFakeClock(initialTime)

	// Advance in steps
	clock.Advance(1 * time.Hour)
	clock.Advance(30 * time.Minute)
	clock.Advance(15 * time.Second)

	expected := time.Date(2025, 1, 1, 13, 30, 15, 0, time.UTC)
	if got := clock.Now(); !got.Equal(expected) {
		t.Errorf("After multiple advances, Now() = %v, want %v", got, expected)
	}
}

func TestFakeClock_ConcurrentAccess(t *testing.T) {
	clock := NewFakeClock(time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC))

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = clock.Now()
				_ = clock.Since(time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC))
			}
		}()
	}

	// Concurrent writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				clock.Advance(1 * time.Second)
			}
		}()
	}

	// Concurrent SetTime calls
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				newTime := time.Date(2025, 1, 1, 12, offset, j, 0, time.UTC)
				clock.SetTime(newTime)
			}
		}(i)
	}

	wg.Wait()

	// If we got here without a data race, the test passes
	// Run with: go test -race
}

func TestFakeClock_UsageExample(t *testing.T) {
	// Example: Test a function that checks if an event is "recent" (within last hour)
	isRecent := func(eventTime time.Time, tp TimeProvider) bool {
		return tp.Since(eventTime) < 1*time.Hour
	}

	// Event happened at noon
	eventTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	// Test at 12:30 (30 minutes later) - should be recent
	clock := NewFakeClock(time.Date(2025, 1, 1, 12, 30, 0, 0, time.UTC))
	if !isRecent(eventTime, clock) {
		t.Error("Event should be recent at 12:30")
	}

	// Advance to 13:30 (1.5 hours later) - should not be recent
	clock.Advance(1 * time.Hour)
	if isRecent(eventTime, clock) {
		t.Error("Event should not be recent at 13:30")
	}

	// Test exact boundary (exactly 1 hour)
	clock.SetTime(time.Date(2025, 1, 1, 13, 0, 0, 0, time.UTC))
	if isRecent(eventTime, clock) {
		t.Error("Event should not be recent at exactly 1 hour")
	}
}

// Verify both implementations satisfy the interface
var _ TimeProvider = WallClock{}
var _ TimeProvider = (*FakeClock)(nil)
