package fetcher

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/adewale/rogue_planet/pkg/crawler"
	"github.com/adewale/rogue_planet/pkg/normalizer"
	"github.com/adewale/rogue_planet/pkg/repository"
)

// Mock implementations

type mockCrawler struct {
	resp *crawler.FeedResponse
	err  error
}

func (m *mockCrawler) FetchWithRetry(ctx context.Context, feedURL string, cache crawler.FeedCache, maxRetries int) (*crawler.FeedResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.resp, nil
}

type mockNormalizer struct {
	metadata *normalizer.FeedMetadata
	entries  []normalizer.Entry
	err      error
}

func (m *mockNormalizer) Parse(feedData []byte, feedURL string, fetchTime time.Time) (*normalizer.FeedMetadata, []normalizer.Entry, error) {
	if m.err != nil {
		return nil, nil, m.err
	}
	return m.metadata, m.entries, nil
}

type mockRepository struct {
	updateFeedErrorCalled   bool
	updateFeedURLCalled     bool
	updateFeedCacheCalled   bool
	updateFeedCalled        bool
	upsertEntryCalled       bool
	upsertEntryCount        int
	updateFeedURLNewURL     string
	updateFeedErrorMsg      string
	upsertEntryError        error
	updateFeedError         error
	updateFeedCacheError    error
	updateFeedURLError      error
	updateFeedErrorError    error
}

func (m *mockRepository) UpdateFeedError(id int64, errorMsg string) error {
	m.updateFeedErrorCalled = true
	m.updateFeedErrorMsg = errorMsg
	return m.updateFeedErrorError
}

func (m *mockRepository) UpdateFeedURL(id int64, newURL string) error {
	m.updateFeedURLCalled = true
	m.updateFeedURLNewURL = newURL
	return m.updateFeedURLError
}

func (m *mockRepository) UpdateFeedCache(id int64, etag, lastModified string, lastFetched time.Time) error {
	m.updateFeedCacheCalled = true
	return m.updateFeedCacheError
}

func (m *mockRepository) UpdateFeed(id int64, title, link string, updated time.Time) error {
	m.updateFeedCalled = true
	return m.updateFeedError
}

func (m *mockRepository) UpsertEntry(entry *repository.Entry) error {
	m.upsertEntryCalled = true
	m.upsertEntryCount++
	return m.upsertEntryError
}

// Implement remaining interface methods (not used in tests)
func (m *mockRepository) GetFeeds(activeOnly bool) ([]repository.Feed, error) {
	return nil, nil
}

func (m *mockRepository) AddFeed(url, title string) (int64, error) {
	return 0, nil
}

func (m *mockRepository) GetFeedByURL(url string) (*repository.Feed, error) {
	return nil, nil
}

func (m *mockRepository) RemoveFeed(id int64) error {
	return nil
}

func (m *mockRepository) GetRecentEntries(days int) ([]repository.Entry, error) {
	return nil, nil
}

func (m *mockRepository) GetRecentEntriesWithOptions(days int, filterByFirstSeen bool, sortBy string) ([]repository.Entry, error) {
	return nil, nil
}

func (m *mockRepository) CountEntries() (int64, error) {
	return 0, nil
}

func (m *mockRepository) CountRecentEntries(days int) (int64, error) {
	return 0, nil
}

func (m *mockRepository) PruneOldEntries(days int) (int64, error) {
	return 0, nil
}

func (m *mockRepository) Close() error {
	return nil
}

type mockLogger struct {
	debugCalls []string
	infoCalls  []string
	warnCalls  []string
	errorCalls []string
}

func (m *mockLogger) Debug(format string, args ...interface{}) {
	m.debugCalls = append(m.debugCalls, format)
}

func (m *mockLogger) Info(format string, args ...interface{}) {
	m.infoCalls = append(m.infoCalls, format)
}

func (m *mockLogger) Warn(format string, args ...interface{}) {
	m.warnCalls = append(m.warnCalls, format)
}

func (m *mockLogger) Error(format string, args ...interface{}) {
	m.errorCalls = append(m.errorCalls, format)
}

// Tests

func TestFetchFeed_Success(t *testing.T) {
	// Setup
	mc := &mockCrawler{
		resp: &crawler.FeedResponse{
			Body:       []byte("<feed><entry>test</entry></feed>"),
			StatusCode: 200,
			FetchTime:  time.Now(),
			NewCache: crawler.FeedCache{
				ETag:         "etag123",
				LastModified: "Mon, 01 Jan 2024 00:00:00 GMT",
			},
		},
	}

	mn := &mockNormalizer{
		metadata: &normalizer.FeedMetadata{
			Title:   "Test Feed",
			Link:    "http://example.com",
			Updated: time.Now(),
		},
		entries: []normalizer.Entry{
			{
				ID:        "entry1",
				Title:     "Test Entry",
				Link:      "http://example.com/entry1",
				Published: time.Now(),
			},
			{
				ID:        "entry2",
				Title:     "Test Entry 2",
				Link:      "http://example.com/entry2",
				Published: time.Now(),
			},
		},
	}

	mr := &mockRepository{}
	ml := &mockLogger{}

	f := New(mc, mn, mr, ml, 3)

	feed := repository.Feed{
		ID:  1,
		URL: "http://example.com/feed",
	}

	// Execute
	result := f.FetchFeed(context.Background(), feed)

	// Verify
	if result.Error != nil {
		t.Errorf("Expected no error, got %v", result.Error)
	}

	if result.NotModified {
		t.Error("Expected NotModified to be false")
	}

	if result.StoredEntries != 2 {
		t.Errorf("Expected 2 stored entries, got %d", result.StoredEntries)
	}

	if !mr.updateFeedCalled {
		t.Error("Expected UpdateFeed to be called")
	}

	if !mr.updateFeedCacheCalled {
		t.Error("Expected UpdateFeedCache to be called")
	}

	if mr.upsertEntryCount != 2 {
		t.Errorf("Expected UpsertEntry to be called 2 times, got %d", mr.upsertEntryCount)
	}
}

func TestFetchFeed_FetchError(t *testing.T) {
	// Setup
	mc := &mockCrawler{
		err: errors.New("network error"),
	}

	mn := &mockNormalizer{}
	mr := &mockRepository{}
	ml := &mockLogger{}

	f := New(mc, mn, mr, ml, 3)

	feed := repository.Feed{
		ID:  1,
		URL: "http://example.com/feed",
	}

	// Execute
	result := f.FetchFeed(context.Background(), feed)

	// Verify
	if result.Error == nil {
		t.Error("Expected error, got nil")
	}

	if !mr.updateFeedErrorCalled {
		t.Error("Expected UpdateFeedError to be called")
	}

	if mr.updateFeedErrorMsg != "network error" {
		t.Errorf("Expected error message 'network error', got %q", mr.updateFeedErrorMsg)
	}

	if len(ml.errorCalls) == 0 {
		t.Error("Expected error to be logged")
	}
}

func TestFetchFeed_301Redirect(t *testing.T) {
	// Setup
	mc := &mockCrawler{
		resp: &crawler.FeedResponse{
			Body:              []byte("<feed><entry>test</entry></feed>"),
			StatusCode:        200,
			PermanentRedirect: true,
			FinalURL:          "http://new.example.com/feed",
			FetchTime:         time.Now(),
			NewCache: crawler.FeedCache{
				ETag:         "etag123",
				LastModified: "Mon, 01 Jan 2024 00:00:00 GMT",
			},
		},
	}

	mn := &mockNormalizer{
		metadata: &normalizer.FeedMetadata{
			Title:   "Test Feed",
			Link:    "http://example.com",
			Updated: time.Now(),
		},
		entries: []normalizer.Entry{},
	}

	mr := &mockRepository{}
	ml := &mockLogger{}

	f := New(mc, mn, mr, ml, 3)

	feed := repository.Feed{
		ID:  1,
		URL: "http://old.example.com/feed",
	}

	// Execute
	result := f.FetchFeed(context.Background(), feed)

	// Verify
	if result.Error != nil {
		t.Errorf("Expected no error, got %v", result.Error)
	}

	if !mr.updateFeedURLCalled {
		t.Error("Expected UpdateFeedURL to be called")
	}

	if mr.updateFeedURLNewURL != "http://new.example.com/feed" {
		t.Errorf("Expected new URL to be http://new.example.com/feed, got %q", mr.updateFeedURLNewURL)
	}

	if len(ml.infoCalls) < 2 {
		t.Errorf("Expected at least 2 info log calls, got %d", len(ml.infoCalls))
	}
}

func TestFetchFeed_304NotModified(t *testing.T) {
	// Setup
	mc := &mockCrawler{
		resp: &crawler.FeedResponse{
			StatusCode:  304,
			NotModified: true,
			FetchTime:   time.Now(),
			NewCache: crawler.FeedCache{
				ETag:         "etag123",
				LastModified: "Mon, 01 Jan 2024 00:00:00 GMT",
			},
		},
	}

	mn := &mockNormalizer{}
	mr := &mockRepository{}
	ml := &mockLogger{}

	f := New(mc, mn, mr, ml, 3)

	feed := repository.Feed{
		ID:  1,
		URL: "http://example.com/feed",
	}

	// Execute
	result := f.FetchFeed(context.Background(), feed)

	// Verify
	if result.Error != nil {
		t.Errorf("Expected no error, got %v", result.Error)
	}

	if !result.NotModified {
		t.Error("Expected NotModified to be true")
	}

	if result.StoredEntries != 0 {
		t.Errorf("Expected 0 stored entries, got %d", result.StoredEntries)
	}

	if !mr.updateFeedCacheCalled {
		t.Error("Expected UpdateFeedCache to be called")
	}

	if mr.updateFeedCalled {
		t.Error("Expected UpdateFeed NOT to be called for 304")
	}

	if mr.upsertEntryCalled {
		t.Error("Expected UpsertEntry NOT to be called for 304")
	}
}

func TestFetchFeed_ParseError(t *testing.T) {
	// Setup
	mc := &mockCrawler{
		resp: &crawler.FeedResponse{
			Body:       []byte("invalid xml"),
			StatusCode: 200,
			FetchTime:  time.Now(),
		},
	}

	mn := &mockNormalizer{
		err: errors.New("parse error"),
	}

	mr := &mockRepository{}
	ml := &mockLogger{}

	f := New(mc, mn, mr, ml, 3)

	feed := repository.Feed{
		ID:  1,
		URL: "http://example.com/feed",
	}

	// Execute
	result := f.FetchFeed(context.Background(), feed)

	// Verify
	if result.Error == nil {
		t.Error("Expected error, got nil")
	}

	if !mr.updateFeedErrorCalled {
		t.Error("Expected UpdateFeedError to be called")
	}

	if mr.updateFeedCalled {
		t.Error("Expected UpdateFeed NOT to be called on parse error")
	}

	if mr.upsertEntryCalled {
		t.Error("Expected UpsertEntry NOT to be called on parse error")
	}

	if len(ml.errorCalls) == 0 {
		t.Error("Expected error to be logged")
	}
}

func TestFetchFeed_EntryStorageError(t *testing.T) {
	// Setup - some entries fail to store
	mc := &mockCrawler{
		resp: &crawler.FeedResponse{
			Body:       []byte("<feed><entry>test</entry></feed>"),
			StatusCode: 200,
			FetchTime:  time.Now(),
			NewCache: crawler.FeedCache{
				ETag:         "etag123",
				LastModified: "Mon, 01 Jan 2024 00:00:00 GMT",
			},
		},
	}

	mn := &mockNormalizer{
		metadata: &normalizer.FeedMetadata{
			Title:   "Test Feed",
			Link:    "http://example.com",
			Updated: time.Now(),
		},
		entries: []normalizer.Entry{
			{ID: "entry1", Title: "Entry 1"},
			{ID: "entry2", Title: "Entry 2"},
			{ID: "entry3", Title: "Entry 3"},
		},
	}

	mr := &mockRepository{
		upsertEntryError: errors.New("duplicate entry"),
	}

	ml := &mockLogger{}

	f := New(mc, mn, mr, ml, 3)

	feed := repository.Feed{
		ID:  1,
		URL: "http://example.com/feed",
	}

	// Execute
	result := f.FetchFeed(context.Background(), feed)

	// Verify - should continue despite entry storage errors
	if result.Error != nil {
		t.Errorf("Expected no overall error, got %v", result.Error)
	}

	if result.StoredEntries != 0 {
		t.Errorf("Expected 0 stored entries (all failed), got %d", result.StoredEntries)
	}

	if mr.upsertEntryCount != 3 {
		t.Errorf("Expected UpsertEntry to be called 3 times, got %d", mr.upsertEntryCount)
	}

	if len(ml.warnCalls) != 3 {
		t.Errorf("Expected 3 warning log calls, got %d", len(ml.warnCalls))
	}
}
