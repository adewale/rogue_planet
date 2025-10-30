package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	m := New(60, 10) // 60 requests/minute, burst of 10

	if m == nil {
		t.Fatal("New() returned nil")
	}

	if len(m.limiters) != 0 {
		t.Errorf("New manager should start with no limiters, got %d", len(m.limiters))
	}

	// Check that limit is correctly converted from requests/minute to requests/second
	expectedLimit := 60.0 / 60.0 // 1 request/second
	if float64(m.limit) != expectedLimit {
		t.Errorf("limit = %f, want %f", m.limit, expectedLimit)
	}

	if m.burst != 10 {
		t.Errorf("burst = %d, want 10", m.burst)
	}
}

func TestExtractDomain(t *testing.T) {
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
	// Create limiter with 60 req/min (1 req/sec), burst of 5
	m := New(60, 5)

	url := "https://example.com/feed.xml"

	// First 5 requests should be allowed immediately (burst)
	for i := 0; i < 5; i++ {
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
	// Create limiter with high rate for faster test
	m := New(600, 5) // 10 req/sec, burst of 5

	url := "https://example.com/feed.xml"
	ctx := context.Background()

	// First 5 should not block
	for i := 0; i < 5; i++ {
		start := time.Now()
		if err := m.Wait(ctx, url); err != nil {
			t.Fatalf("Wait() error on request %d: %v", i+1, err)
		}
		elapsed := time.Since(start)
		// Should be nearly instant
		if elapsed > 10*time.Millisecond {
			t.Errorf("Request %d took %v, expected < 10ms", i+1, elapsed)
		}
	}

	// 6th request should block briefly (rate is 10 req/sec = 100ms between requests)
	start := time.Now()
	if err := m.Wait(ctx, url); err != nil {
		t.Fatalf("Wait() error on delayed request: %v", err)
	}
	elapsed := time.Since(start)

	// Should have waited approximately 100ms (10 req/sec = 1 every 100ms)
	// Allow some variance: 50ms - 200ms
	if elapsed < 50*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Errorf("Delayed request took %v, expected ~100ms", elapsed)
	}
}

func TestWaitWithCancelledContext(t *testing.T) {
	m := New(60, 10)

	// Create an already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
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
	m := New(60, 10)

	// Invalid URLs should fail gracefully (allow the request)
	if !m.Allow("not a url") {
		t.Error("Invalid URL should be allowed (fail open)")
	}

	err := m.Wait(context.Background(), "also not a url")
	if err != nil {
		t.Errorf("Wait() with invalid URL should not error, got: %v", err)
	}
}

func TestConcurrentAccess(t *testing.T) {
	m := New(600, 10) // High rate for testing

	url := "https://example.com/feed.xml"
	concurrency := 20
	iterations := 10

	// Run concurrent goroutines that all access the same domain
	done := make(chan bool)
	for i := 0; i < concurrency; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				m.Allow(url)
				m.Wait(context.Background(), url)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}

	// Should have created only one limiter for the domain
	if len(m.limiters) != 1 {
		t.Errorf("Concurrent access created %d limiters, want 1", len(m.limiters))
	}
}
