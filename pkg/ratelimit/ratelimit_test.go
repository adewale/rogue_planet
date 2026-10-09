package ratelimit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/adewale/rogue_planet/pkg/timeprovider"
)

// newFakeClock returns a fake clock at a fixed instant. Waits on it advance
// the fake time and are recorded, so tests assert exact delays instantly.
func newFakeClock() *timeprovider.FakeClock {
	return timeprovider.NewFakeClock(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
}

func TestNew(t *testing.T) {
	t.Parallel()
	m := New(60, 10) // 60 requests/minute, burst of 10

	if m == nil {
		t.Fatal("New() returned nil")
	}

	if len(m.limiters) != 0 {
		t.Errorf("New manager should start with no limiters, got %d", len(m.limiters))
	}

	// Check that limit is correctly converted from requests/minute to requests/second
	requestsPerMinute := 60.0
	expectedLimit := requestsPerMinute / 60.0 // 1 request/second
	if float64(m.limit) != expectedLimit {
		t.Errorf("limit = %f, want %f", m.limit, expectedLimit)
	}

	if m.burst != 10 {
		t.Errorf("burst = %d, want 10", m.burst)
	}
}

func TestExtractDomain(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name: "http URL",
			url:  "http://example.com/feed.xml",
			want: "example.com",
		},
		{
			name: "https URL",
			url:  "https://blog.example.com/rss",
			want: "blog.example.com",
		},
		{
			name: "URL with port",
			url:  "http://example.com:8080/feed",
			want: "example.com",
		},
		{
			name: "URL with path and query",
			url:  "https://example.com/feeds?format=xml",
			want: "example.com",
		},
		{
			name: "URL without scheme gets parsed",
			url:  "not-a-url",
			want: "", // url.Parse is permissive, returns empty hostname
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractDomain(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractDomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("extractDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLimiter(t *testing.T) {
	t.Parallel()
	m := New(60, 10)

	// First call should create a new limiter
	limiter1 := m.getLimiter("example.com")
	if limiter1 == nil {
		t.Fatal("getLimiter() returned nil")
	}

	if len(m.limiters) != 1 {
		t.Errorf("After first getLimiter, expected 1 limiter, got %d", len(m.limiters))
	}

	// Second call for same domain should return the same limiter
	limiter2 := m.getLimiter("example.com")
	if limiter1 != limiter2 {
		t.Error("getLimiter() should return same limiter for same domain")
	}

	if len(m.limiters) != 1 {
		t.Errorf("After second getLimiter for same domain, expected 1 limiter, got %d", len(m.limiters))
	}

	// Different domain should create a different limiter
	limiter3 := m.getLimiter("another.com")
	if limiter3 == nil {
		t.Fatal("getLimiter() for different domain returned nil")
	}

	if limiter1 == limiter3 {
		t.Error("getLimiter() should return different limiters for different domains")
	}

	if len(m.limiters) != 2 {
		t.Errorf("After getLimiter for different domain, expected 2 limiters, got %d", len(m.limiters))
	}
}

func TestAllow(t *testing.T) {
	t.Parallel()
	// Create limiter with 60 req/min (1 req/sec), burst of 5
	m := New(60, 5)

	url := "https://example.com/feed.xml"

	// First 5 requests should be allowed immediately (burst)
	for i := range 5 {
		if !m.Allow(url) {
			t.Errorf("Request %d should be allowed (within burst)", i+1)
		}
	}

	// 6th request should be rate-limited
	if m.Allow(url) {
		t.Error("Request 6 should be rate-limited (burst exhausted)")
	}

	// Different domain should have its own limiter
	if !m.Allow("https://other.com/feed.xml") {
		t.Error("Request to different domain should be allowed")
	}
}

func TestWait(t *testing.T) {
	t.Parallel()
	clock := newFakeClock()
	m := NewWithClock(600, 5, clock) // 10 req/sec, burst of 5

	url := "https://example.com/feed.xml"
	ctx := t.Context()

	// First 5 are covered by the burst and must not wait at all
	for i := range 5 {
		if err := m.Wait(ctx, url); err != nil {
			t.Fatalf("Wait() error on request %d: %v", i+1, err)
		}
	}
	if sleeps := clock.Sleeps(); len(sleeps) != 0 {
		t.Fatalf("burst requests waited %v, want no waits", sleeps)
	}

	// 6th request must wait exactly one token interval (10 req/sec = 100ms)
	if err := m.Wait(ctx, url); err != nil {
		t.Fatalf("Wait() error on delayed request: %v", err)
	}
	sleeps := clock.Sleeps()
	if len(sleeps) != 1 || sleeps[0] != 100*time.Millisecond {
		t.Errorf("6th request waited %v, want exactly [100ms]", sleeps)
	}
}

func TestWait_WallClockDefault(t *testing.T) {
	t.Parallel()
	// New() must use the real clock: once the burst is spent, Wait blocks.
	m := New(600, 1) // 10 req/sec, burst of 1
	url := "https://example.com/feed.xml"

	if err := m.Wait(t.Context(), url); err != nil {
		t.Fatalf("first Wait() error: %v", err)
	}
	start := time.Now()
	if err := m.Wait(t.Context(), url); err != nil {
		t.Fatalf("second Wait() error: %v", err)
	}
	// Lower bound only: a real timer never fires early, but may fire late under load.
	if elapsed := time.Since(start); elapsed < 50*time.Millisecond {
		t.Errorf("second Wait() returned after %v, want it to block for ~100ms", elapsed)
	}
}

func TestWait_WouldExceedDeadline(t *testing.T) {
	t.Parallel()
	// Context deadlines are wall-clock times, so start the fake clock at the
	// real time so that both clocks agree on when the deadline falls.
	clock := timeprovider.NewFakeClock(time.Now())
	m := NewWithClock(6, 1, clock) // 6 req/min = one token every 10s, burst of 1
	url := "https://example.com/feed.xml"

	if err := m.Wait(t.Context(), url); err != nil {
		t.Fatalf("first Wait() error: %v", err)
	}

	// The next token is 10s away; a context that ends sooner can never succeed.
	ctx, cancel := context.WithDeadline(t.Context(), clock.Now().Add(5*time.Second))
	defer cancel()
	if err := m.Wait(ctx, url); !errors.Is(err, ErrWouldExceedDeadline) {
		t.Fatalf("Wait() error = %v, want ErrWouldExceedDeadline", err)
	}
	if sleeps := clock.Sleeps(); len(sleeps) != 0 {
		t.Errorf("Wait() slept %v before failing, want no wait", sleeps)
	}

	// The rejected request must not have consumed the next token.
	if err := m.Wait(t.Context(), url); err != nil {
		t.Fatalf("Wait() after rejection error: %v", err)
	}
	if sleeps := clock.Sleeps(); len(sleeps) != 1 || sleeps[0] != 10*time.Second {
		t.Errorf("Wait() after rejection waited %v, want exactly [10s]", sleeps)
	}
}

func TestWait_NeverAllowed(t *testing.T) {
	t.Parallel()
	m := NewWithClock(60, 0, newFakeClock()) // burst 0: no request can ever pass
	if err := m.Wait(t.Context(), "https://example.com/feed.xml"); !errors.Is(err, ErrNeverAllowed) {
		t.Errorf("Wait() error = %v, want ErrNeverAllowed", err)
	}
}

func TestWait_CancelledWhileWaitingReturnsToken(t *testing.T) {
	t.Parallel()
	fake := newFakeClock()
	ctx, cancel := context.WithCancel(t.Context())
	clock := &cancellingClock{FakeClock: fake, cancel: cancel}
	m := NewWithClock(60, 1, clock) // 1 req/sec, burst of 1
	url := "https://example.com/feed.xml"

	if err := m.Wait(t.Context(), url); err != nil {
		t.Fatalf("first Wait() error: %v", err)
	}
	// The second request has to wait; the caller gives up during that wait.
	if err := m.Wait(ctx, url); !errors.Is(err, context.Canceled) {
		t.Fatalf("Wait() error = %v, want context.Canceled", err)
	}
	// No fake time has passed, so no token has been refilled yet.
	if m.Allow(url) {
		t.Fatal("Allow() = true with no elapsed time, want false (limiter must read the injected clock)")
	}
	// Its reserved token was handed back, so after 1s the next caller proceeds
	// without waiting a further second.
	fake.Advance(time.Second)
	if !m.Allow(url) {
		t.Error("Allow() = false after cancelled wait, want the reservation to have been returned")
	}
}

// cancellingClock cancels the caller's context instead of sleeping, modelling
// a caller that gives up while Wait is blocked.
type cancellingClock struct {
	*timeprovider.FakeClock
	cancel context.CancelFunc
}

func (c *cancellingClock) Sleep(ctx context.Context, _ time.Duration) error {
	c.cancel()
	<-ctx.Done()
	return ctx.Err()
}

func TestWaitWithCancelledContext(t *testing.T) {
	t.Parallel()
	m := New(60, 10)

	// Create an already-cancelled context
	ctx, cancel := context.WithCancel(t.Context())
	cancel() // Cancel immediately

	// Wait should respect the cancelled context
	err := m.Wait(ctx, "https://example.com/feed.xml")
	if err == nil {
		t.Error("Wait() with cancelled context should return error")
	}

	if ctx.Err() != context.Canceled {
		t.Errorf("Context error = %v, want Canceled", ctx.Err())
	}
}

func TestStats(t *testing.T) {
	t.Parallel()
	m := New(60, 10)

	// Initially no limiters
	stats := m.Stats()
	if stats.TotalDomains != 0 {
		t.Errorf("TotalDomains = %d, want 0", stats.TotalDomains)
	}

	// Create limiters for two domains
	m.Allow("https://example.com/feed.xml")
	m.Allow("https://another.com/feed.xml")

	stats = m.Stats()
	if stats.TotalDomains != 2 {
		t.Errorf("TotalDomains = %d, want 2", stats.TotalDomains)
	}

	// Check stats for a specific domain
	exampleStats, exists := stats.Limiters["example.com"]
	if !exists {
		t.Error("Stats should include example.com")
	}

	if exampleStats.Domain != "example.com" {
		t.Errorf("Domain = %s, want example.com", exampleStats.Domain)
	}

	if exampleStats.Burst != 10 {
		t.Errorf("Burst = %d, want 10", exampleStats.Burst)
	}

	if exampleStats.RequestsPerMinute != 60.0 {
		t.Errorf("RequestsPerMinute = %f, want 60.0", exampleStats.RequestsPerMinute)
	}
}

func TestResetAll(t *testing.T) {
	t.Parallel()
	m := New(60, 10)

	// Create some limiters
	m.Allow("https://example.com/feed.xml")
	m.Allow("https://another.com/feed.xml")

	if len(m.limiters) != 2 {
		t.Fatalf("Expected 2 limiters before reset, got %d", len(m.limiters))
	}

	// Reset all
	m.ResetAll()

	if len(m.limiters) != 0 {
		t.Errorf("Expected 0 limiters after reset, got %d", len(m.limiters))
	}

	stats := m.Stats()
	if stats.TotalDomains != 0 {
		t.Errorf("TotalDomains = %d after reset, want 0", stats.TotalDomains)
	}
}

func TestSetLimit(t *testing.T) {
	t.Parallel()
	m := New(60, 10)

	// Create a limiter
	m.Allow("https://example.com/feed.xml")

	// Change the limit
	m.SetLimit(120, 20)

	if float64(m.limit) != 2.0 { // 120 req/min = 2 req/sec
		t.Errorf("limit = %f, want 2.0", m.limit)
	}

	if m.burst != 20 {
		t.Errorf("burst = %d, want 20", m.burst)
	}

	// Existing limiters should be updated
	limiter := m.getLimiter("example.com")
	if limiter.Burst() != 20 {
		t.Errorf("Existing limiter burst = %d, want 20", limiter.Burst())
	}
}

func TestInvalidURL(t *testing.T) {
	t.Parallel()
	m := New(60, 10)

	// Invalid URLs should fail gracefully (allow the request)
	if !m.Allow("not a url") {
		t.Error("Invalid URL should be allowed (fail open)")
	}

	err := m.Wait(t.Context(), "also not a url")
	if err != nil {
		t.Errorf("Wait() with invalid URL should not error, got: %v", err)
	}
}

func TestConcurrentAccess(t *testing.T) {
	t.Parallel()
	m := NewWithClock(600, 10, newFakeClock()) // Waits advance fake time, not the wall clock

	url := "https://example.com/feed.xml"
	concurrency := 20
	iterations := 10

	// Run concurrent goroutines that all access the same domain
	done := make(chan bool)
	errors := make(chan error, concurrency*iterations)
	for range concurrency {
		go func() {
			for range iterations {
				m.Allow(url)
				if err := m.Wait(t.Context(), url); err != nil {
					errors <- err
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for range concurrency {
		<-done
	}

	// Check if any errors occurred
	close(errors)
	for err := range errors {
		t.Errorf("Wait() error during concurrent access: %v", err)
	}

	// Should have created only one limiter for the domain
	if len(m.limiters) != 1 {
		t.Errorf("Concurrent access created %d limiters, want 1", len(m.limiters))
	}
}
