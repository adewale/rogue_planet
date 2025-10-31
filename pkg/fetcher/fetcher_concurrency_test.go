package fetcher

import (
	"context"
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
	concurrentOps *int32 // Tracks concurrent operations
	maxConcurrent *int32 // Tracks maximum concurrent operations observed
}

func (m *mockSlowCrawler) FetchWithRetry(ctx context.Context, feedURL string, cache crawler.FeedCache, maxRetries int) (*crawler.FeedResponse, error) {
	// Track concurrent operations
	current := atomic.AddInt32(m.concurrentOps, 1)
	defer atomic.AddInt32(m.concurrentOps, -1)

	// Update max concurrent if this is higher
	for {
		oldMax := atomic.LoadInt32(m.maxConcurrent)
		if current <= oldMax {
			break
		}
		if atomic.CompareAndSwapInt32(m.maxConcurrent, oldMax, current) {
			break
		}
	}

	// Simulate slow HTTP fetch
	time.Sleep(m.delay)

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
	const (
		numFeeds      = 6
		concurrency   = 3
		fetchDelay    = 100 * time.Millisecond
		maxExpected   = 250 * time.Millisecond // 2 batches: 200ms + overhead
		minConcurrent = 2                      // Should see at least 2 concurrent
	)

	var concurrentOps int32
	var maxConcurrent int32

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
	for i := 0; i < numFeeds; i++ {
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

			result := fetcher.FetchFeed(context.Background(), f)
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
	maxSeen := atomic.LoadInt32(&maxConcurrent)
	if maxSeen < minConcurrent {
		t.Errorf("Max concurrent operations was %d, expected at least %d. Feeds appear to be serialized!", maxSeen, minConcurrent)
	}

	t.Logf("✓ Processed %d feeds in %v with max concurrency of %d", numFeeds, elapsed, maxSeen)
}

// TestFetchFeed_MutexProtectsDatabase verifies that the mutex correctly protects
// database operations while allowing concurrent HTTP fetching and parsing.
func TestFetchFeed_MutexProtectsDatabase(t *testing.T) {
	const numFeeds = 10

	var concurrentFetches int32
	var concurrentDBWrites int32
	var maxConcurrentFetches int32
	var maxConcurrentDBWrites int32

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
	for i := 0; i < numFeeds; i++ {
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
			result := fetcher.FetchFeed(context.Background(), f)
			if result.Error != nil {
				t.Errorf("Unexpected error: %v", result.Error)
			}
		}(feed)
	}

	wg.Wait()

	// Verify concurrent fetching occurred
	maxFetches := atomic.LoadInt32(&maxConcurrentFetches)
	if maxFetches < 2 {
		t.Errorf("Max concurrent fetches was %d, expected at least 2", maxFetches)
	}

	// Verify database operations were serialized (mutex working)
	maxDBWrites := atomic.LoadInt32(&maxConcurrentDBWrites)
	if maxDBWrites > 1 {
		t.Errorf("Max concurrent DB writes was %d, expected 1 (mutex should serialize)", maxDBWrites)
	}

	t.Logf("✓ Max concurrent fetches: %d (good), Max concurrent DB writes: %d (correct - mutex working)",
		maxFetches, maxDBWrites)
}

// mockRepositoryWithConcurrency tracks concurrent database operations
type mockRepositoryWithConcurrency struct {
	mockRepository
	concurrentOps    *int32
	maxConcurrentOps *int32
}

func (m *mockRepositoryWithConcurrency) trackOperation() func() {
	current := atomic.AddInt32(m.concurrentOps, 1)

	// Update max if needed
	for {
		oldMax := atomic.LoadInt32(m.maxConcurrentOps)
		if current <= oldMax {
			break
		}
		if atomic.CompareAndSwapInt32(m.maxConcurrentOps, oldMax, current) {
			break
		}
	}

	// Simulate some DB operation time
	time.Sleep(5 * time.Millisecond)

	return func() {
		atomic.AddInt32(m.concurrentOps, -1)
	}
}

func (m *mockRepositoryWithConcurrency) UpdateFeedError(id int64, errorMsg string) error {
	defer m.trackOperation()()
	return m.mockRepository.UpdateFeedError(id, errorMsg)
}

func (m *mockRepositoryWithConcurrency) UpdateFeedURL(id int64, newURL string) error {
	defer m.trackOperation()()
	return m.mockRepository.UpdateFeedURL(id, newURL)
}

func (m *mockRepositoryWithConcurrency) UpdateFeedCache(id int64, etag, lastModified string, lastFetched time.Time) error {
	defer m.trackOperation()()
	return m.mockRepository.UpdateFeedCache(id, etag, lastModified, lastFetched)
}

func (m *mockRepositoryWithConcurrency) UpdateFeed(id int64, title, link string, updated time.Time) error {
	defer m.trackOperation()()
	return m.mockRepository.UpdateFeed(id, title, link, updated)
}

func (m *mockRepositoryWithConcurrency) UpsertEntry(entry *repository.Entry) error {
	defer m.trackOperation()()
	return m.mockRepository.UpsertEntry(entry)
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
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
