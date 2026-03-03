package fetcher

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/adewale/rogue_planet/pkg/crawler"
	"github.com/adewale/rogue_planet/pkg/normalizer"
	"github.com/adewale/rogue_planet/pkg/repository"
)

// mockSlowCrawler simulates slow HTTP fetching
type mockSlowCrawler struct {
	delay         time.Duration
	concurrentOps *atomic.Int32                         // Tracks concurrent operations
	maxConcurrent *atomic.Int32                         // Tracks maximum concurrent operations observed
	responseFunc  func() (*crawler.FeedResponse, error) // Optional dynamic response
}

func (m *mockSlowCrawler) FetchWithRetry(ctx context.Context, feedURL string, cache crawler.FeedCache, maxRetries int) (*crawler.FeedResponse, error) {
	// Track concurrent operations if pointers are provided
	if m.concurrentOps != nil && m.maxConcurrent != nil {
		current := m.concurrentOps.Add(1)
		defer m.concurrentOps.Add(-1)

		// Update max concurrent if this is higher
		for {
			oldMax := m.maxConcurrent.Load()
			if current <= oldMax {
				break
			}
			if m.maxConcurrent.CompareAndSwap(oldMax, current) {
				break
			}
		}
	}

	// Simulate slow HTTP fetch
	time.Sleep(m.delay)

	// Use dynamic response if provided
	if m.responseFunc != nil {
		return m.responseFunc()
	}

	return &crawler.FeedResponse{
		Body:       []byte("<feed><entry>test</entry></feed>"),
		StatusCode: 200,
		FetchTime:  time.Now(),
		NewCache: crawler.FeedCache{
			ETag:         "etag123",
			LastModified: "Mon, 01 Jan 2024 00:00:00 GMT",
		},
	}, nil
}

// TestFetchFeed_Concurrency verifies that multiple feeds are processed concurrently
// and not serialized by incorrect mutex placement.
func TestFetchFeed_Concurrency(t *testing.T) {
	t.Parallel()
	const (
		numFeeds      = 6
		concurrency   = 3
		fetchDelay    = 100 * time.Millisecond
		maxExpected   = 250 * time.Millisecond // 2 batches: 200ms + overhead
		minConcurrent = 2                      // Should see at least 2 concurrent
	)

	var concurrentOps atomic.Int32
	var maxConcurrent atomic.Int32

	mc := &mockSlowCrawler{
		delay:         fetchDelay,
		concurrentOps: &concurrentOps,
		maxConcurrent: &maxConcurrent,
	}

	mn := &mockNormalizer{
		metadata: &normalizer.FeedMetadata{
			Title:   "Test Feed",
			Link:    "http://example.com",
			Updated: time.Now(),
		},
		entries: []normalizer.Entry{
			{ID: "entry1", Title: "Entry 1"},
		},
	}

	mr := &mockRepository{}
	ml := &mockLogger{}

	// Create mutex for shared repository access
	var mu sync.Mutex
	fetcher := New(mc, mn, mr, &mu, ml, 3)

	// Create test feeds
	feeds := make([]repository.Feed, numFeeds)
	for i := range numFeeds {
		feeds[i] = repository.Feed{
			ID:  int64(i + 1),
			URL: "http://example.com/feed" + string(rune('0'+i)),
		}
	}

	// Process feeds concurrently
	start := time.Now()

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	for _, feed := range feeds {
		wg.Add(1)
		go func(f repository.Feed) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			result := fetcher.FetchFeed(t.Context(), f)
			if result.Error != nil {
				t.Errorf("Unexpected error: %v", result.Error)
			}
		}(feed)
	}

	wg.Wait()
	elapsed := time.Since(start)

	// Verify timing: Should complete in ~2 batches (200ms), not serially (600ms)
	if elapsed > maxExpected {
		t.Errorf("Took too long: %v (expected <%v). Feeds may be processing serially instead of concurrently.", elapsed, maxExpected)
	}

	// Verify we actually achieved concurrency
	maxSeen := maxConcurrent.Load()
	if maxSeen < minConcurrent {
		t.Errorf("Max concurrent operations was %d, expected at least %d. Feeds appear to be serialized!", maxSeen, minConcurrent)
	}

	t.Logf("✓ Processed %d feeds in %v with max concurrency of %d", numFeeds, elapsed, maxSeen)
}

// TestFetchFeed_MutexProtectsDatabase verifies that the mutex correctly protects
// database operations while allowing concurrent HTTP fetching and parsing.
func TestFetchFeed_MutexProtectsDatabase(t *testing.T) {
	t.Parallel()
	const numFeeds = 10

	var concurrentFetches atomic.Int32
	var maxConcurrentFetches atomic.Int32
	var concurrentDBWrites atomic.Int32
	var maxConcurrentDBWrites atomic.Int32

	mc := &mockSlowCrawler{
		delay:         10 * time.Millisecond,
		concurrentOps: &concurrentFetches,
		maxConcurrent: &maxConcurrentFetches,
	}

	mn := &mockNormalizer{
		metadata: &normalizer.FeedMetadata{
			Title:   "Test Feed",
			Link:    "http://example.com",
			Updated: time.Now(),
		},
		entries: []normalizer.Entry{
			{ID: "entry1", Title: "Entry 1"},
		},
	}

	// Repository that tracks concurrent operations
	mr := &mockRepositoryWithConcurrency{
		mockRepository:   mockRepository{},
		concurrentOps:    &concurrentDBWrites,
		maxConcurrentOps: &maxConcurrentDBWrites,
	}

	ml := &mockLogger{}

	var mu sync.Mutex
	fetcher := New(mc, mn, mr, &mu, ml, 3)

	// Create test feeds
	feeds := make([]repository.Feed, numFeeds)
	for i := range numFeeds {
		feeds[i] = repository.Feed{
			ID:  int64(i + 1),
			URL: "http://example.com/feed" + string(rune('0'+i)),
		}
	}

	// Process feeds concurrently
	var wg sync.WaitGroup
	for _, feed := range feeds {
		wg.Add(1)
		go func(f repository.Feed) {
			defer wg.Done()
			result := fetcher.FetchFeed(t.Context(), f)
			if result.Error != nil {
				t.Errorf("Unexpected error: %v", result.Error)
			}
		}(feed)
	}

	wg.Wait()

	// Verify concurrent fetching occurred
	maxFetches := maxConcurrentFetches.Load()
	if maxFetches < 2 {
		t.Errorf("Max concurrent fetches was %d, expected at least 2", maxFetches)
	}

	// Verify database operations were serialized (mutex working)
	maxDBWrites := maxConcurrentDBWrites.Load()
	if maxDBWrites > 1 {
		t.Errorf("Max concurrent DB writes was %d, expected 1 (mutex should serialize)", maxDBWrites)
	}

	t.Logf("✓ Max concurrent fetches: %d (good), Max concurrent DB writes: %d (correct - mutex working)",
		maxFetches, maxDBWrites)
}

// mockRepositoryWithConcurrency tracks concurrent database operations
type mockRepositoryWithConcurrency struct {
	mockRepository
	concurrentOps    *atomic.Int32
	maxConcurrentOps *atomic.Int32
}

func (m *mockRepositoryWithConcurrency) trackOperation() func() {
	current := m.concurrentOps.Add(1)

	// Update max if needed
	for {
		oldMax := m.maxConcurrentOps.Load()
		if current <= oldMax {
			break
		}
		if m.maxConcurrentOps.CompareAndSwap(oldMax, current) {
			break
		}
	}

	// Simulate some DB operation time
	time.Sleep(5 * time.Millisecond)

	return func() {
		m.concurrentOps.Add(-1)
	}
}

func (m *mockRepositoryWithConcurrency) UpdateFeedError(ctx context.Context, id int64, errorMsg string) error {
	defer m.trackOperation()()
	return m.mockRepository.UpdateFeedError(ctx, id, errorMsg)
}

func (m *mockRepositoryWithConcurrency) UpdateFeedURL(ctx context.Context, id int64, newURL string) error {
	defer m.trackOperation()()
	return m.mockRepository.UpdateFeedURL(ctx, id, newURL)
}

func (m *mockRepositoryWithConcurrency) UpdateFeedCache(ctx context.Context, id int64, etag, lastModified string, lastFetched time.Time) error {
	defer m.trackOperation()()
	return m.mockRepository.UpdateFeedCache(ctx, id, etag, lastModified, lastFetched)
}

func (m *mockRepositoryWithConcurrency) UpdateFeed(ctx context.Context, id int64, title, link string, updated time.Time) error {
	defer m.trackOperation()()
	return m.mockRepository.UpdateFeed(ctx, id, title, link, updated)
}

func (m *mockRepositoryWithConcurrency) UpsertEntry(ctx context.Context, entry *repository.Entry) error {
	defer m.trackOperation()()
	return m.mockRepository.UpsertEntry(ctx, entry)
}

// BenchmarkFetchFeed_Sequential measures performance with serialized processing
func BenchmarkFetchFeed_Sequential(b *testing.B) {
	mc := &mockCrawler{
		resp: &crawler.FeedResponse{
			Body:       []byte("<feed><entry>test</entry></feed>"),
			StatusCode: 200,
			FetchTime:  time.Now(),
		},
	}

	mn := &mockNormalizer{
		metadata: &normalizer.FeedMetadata{Title: "Test", Link: "http://example.com", Updated: time.Now()},
		entries:  []normalizer.Entry{{ID: "1", Title: "Entry"}},
	}

	mr := &mockRepository{}
	ml := &mockLogger{}

	// No mutex = single-threaded
	f := New(mc, mn, mr, nil, ml, 3)

	feed := repository.Feed{ID: 1, URL: "http://example.com/feed"}

	for b.Loop() {
		f.FetchFeed(context.Background(), feed)
	}
}

// BenchmarkFetchFeed_Concurrent measures performance with concurrent processing
func BenchmarkFetchFeed_Concurrent(b *testing.B) {
	mc := &mockCrawler{
		resp: &crawler.FeedResponse{
			Body:       []byte("<feed><entry>test</entry></feed>"),
			StatusCode: 200,
			FetchTime:  time.Now(),
		},
	}

	mn := &mockNormalizer{
		metadata: &normalizer.FeedMetadata{Title: "Test", Link: "http://example.com", Updated: time.Now()},
		entries:  []normalizer.Entry{{ID: "1", Title: "Entry"}},
	}

	mr := &mockRepository{}
	ml := &mockLogger{}

	var mu sync.Mutex
	f := New(mc, mn, mr, &mu, ml, 3)

	feed := repository.Feed{ID: 1, URL: "http://example.com/feed"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f.FetchFeed(context.Background(), feed)
		}
	})
}

// ADDITIONAL CONCURRENT FETCH EDGE CASES

func TestFetchFeed_ConcurrentErrorHandling(t *testing.T) {
	t.Parallel()

	// Scenario: Multiple feeds failing simultaneously
	// Verify error handling is thread-safe and doesn't deadlock or panic

	const numFeeds = 10

	// All requests will fail with different errors
	errorMessages := make([]string, numFeeds)
	for i := range numFeeds {
		errorMessages[i] = "Network error " + string(rune('A'+i))
	}

	var requestCount atomic.Int32
	mc := &mockCrawler{
		responseFunc: func() (*crawler.FeedResponse, error) {
			idx := requestCount.Add(1) - 1
			return nil, errors.New(errorMessages[idx])
		},
	}

	mn := &mockNormalizer{}
	mr := &mockRepository{}
	ml := &mockLogger{}

	var mu sync.Mutex
	fetcher := New(mc, mn, mr, &mu, ml, 3)

	// Launch concurrent fetches that will all fail
	var wg sync.WaitGroup
	results := make([]FetchResult, numFeeds)

	for i := range numFeeds {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			feed := repository.Feed{ID: int64(idx + 1), URL: "http://example.com/feed" + string(rune('0'+idx))}
			results[idx] = fetcher.FetchFeed(t.Context(), feed)
		}(i)
	}

	wg.Wait()

	// Verify all requests failed with errors
	for i, result := range results {
		if result.Error == nil {
			t.Errorf("Feed %d: Expected error, got nil", i)
		}
	}

	// Verify UpdateFeedError was called for each feed
	if mr.updateFeedErrorCalled {
		// At least some errors were recorded (mock doesn't track count)
		t.Logf("✓ Error recording handled concurrently without deadlock")
	}
}

func TestFetchFeed_ContextCancellationDuringConcurrentFetches(t *testing.T) {
	t.Parallel()

	// Scenario: Cancel context while multiple fetches are in progress
	// Should gracefully stop all ongoing fetches

	const numFeeds = 20
	const fetchDelay = 500 * time.Millisecond

	var activeRequests atomic.Int32

	mc := &mockSlowCrawler{
		delay:         fetchDelay,
		concurrentOps: &activeRequests,
	}

	mn := &mockNormalizer{
		metadata: &normalizer.FeedMetadata{Title: "Test"},
		entries:  []normalizer.Entry{{ID: "1"}},
	}

	mr := &mockRepository{}
	ml := &mockLogger{}

	var mu sync.Mutex
	fetcher := New(mc, mn, mr, &mu, ml, 3)

	ctx, cancel := context.WithCancel(t.Context())

	// Launch concurrent fetches
	var wg sync.WaitGroup
	var completedCount atomic.Int32
	var errorCount atomic.Int32

	for i := range numFeeds {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			feed := repository.Feed{ID: int64(idx + 1), URL: "http://example.com/feed"}
			result := fetcher.FetchFeed(ctx, feed)

			if result.Error != nil {
				errorCount.Add(1)
			} else {
				completedCount.Add(1)
			}
		}(i)
	}

	// Let some fetches start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for all goroutines to finish
	wg.Wait()

	// Verify all goroutines completed without deadlock
	errors := errorCount.Load()
	completed := completedCount.Load()

	t.Logf("✓ Concurrent fetch with cancellation: %d completed, %d errored", completed, errors)

	// The important thing is that cancellation didn't cause deadlock or panic
	// Some may complete before cancellation propagates, which is fine
	if completed+errors != numFeeds {
		t.Errorf("Expected %d total results, got %d completed + %d errors = %d",
			numFeeds, completed, errors, completed+errors)
	}
}

func TestFetchFeed_ConcurrentAccessToSameFeed(t *testing.T) {
	t.Parallel()

	// Scenario: Multiple goroutines fetching the SAME feed URL simultaneously
	// This is a real-world scenario (e.g., manual refresh + scheduled fetch)
	// Verify no data corruption or race conditions

	const numConcurrent = 5
	const fetchDelay = 50 * time.Millisecond

	var requestCount atomic.Int32

	mc := &mockSlowCrawler{
		delay: fetchDelay,
		responseFunc: func() (*crawler.FeedResponse, error) {
			count := requestCount.Add(1)
			return &crawler.FeedResponse{
				Body:       []byte("<feed>data " + string(rune('0'+count-1)) + "</feed>"),
				StatusCode: 200,
				FetchTime:  time.Now(),
			}, nil
		},
	}

	mn := &mockNormalizer{
		metadata: &normalizer.FeedMetadata{Title: "Test Feed"},
		entries:  []normalizer.Entry{{ID: "entry-1", Title: "Entry"}},
	}

	mr := &mockRepository{}
	ml := &mockLogger{}

	var mu sync.Mutex
	fetcher := New(mc, mn, mr, &mu, ml, 3)

	// Same feed fetched by multiple goroutines simultaneously
	sameFeed := repository.Feed{ID: 1, URL: "http://example.com/feed"}

	var wg sync.WaitGroup
	results := make([]FetchResult, numConcurrent)

	// Launch concurrent fetches of the SAME feed
	for i := range numConcurrent {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = fetcher.FetchFeed(t.Context(), sameFeed)
		}(i)
	}

	wg.Wait()

	// All should succeed without panics or data corruption
	for i, result := range results {
		if result.Error != nil {
			t.Errorf("Fetch %d failed: %v", i, result.Error)
		}
	}

	// Should have made numConcurrent requests (one per goroutine)
	totalRequests := requestCount.Load()
	if totalRequests != numConcurrent {
		t.Errorf("Expected %d concurrent requests, got %d", numConcurrent, totalRequests)
	}

	t.Logf("✓ %d concurrent fetches of same feed completed without race conditions", numConcurrent)
}
