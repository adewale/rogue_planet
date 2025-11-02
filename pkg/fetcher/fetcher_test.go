package fetcher

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adewale/rogue_planet/pkg/crawler"
	"github.com/adewale/rogue_planet/pkg/normalizer"
	"github.com/adewale/rogue_planet/pkg/repository"
)

// Mock implementations

type mockCrawler struct {
	resp         *crawler.FeedResponse
	err          error
	responseFunc func() (*crawler.FeedResponse, error) // Optional dynamic response
}

func (m *mockCrawler) FetchWithRetry(ctx context.Context, feedURL string, cache crawler.FeedCache, maxRetries int) (*crawler.FeedResponse, error) {
	// Use dynamic response if provided
	if m.responseFunc != nil {
		return m.responseFunc()
	}
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

func (m *mockNormalizer) Parse(ctx context.Context, feedData []byte, feedURL string, fetchTime time.Time) (*normalizer.FeedMetadata, []normalizer.Entry, error) {
	if m.err != nil {
		return nil, nil, m.err
	}
	return m.metadata, m.entries, nil
}

type mockRepository struct {
	updateFeedErrorCalled bool
	updateFeedURLCalled   bool
	updateFeedCacheCalled bool
	updateFeedCalled      bool
	upsertEntryCalled     bool
	upsertEntryCount      int
	updateFeedURLNewURL   string
	updateFeedErrorMsg    string
	upsertEntryError      error
	updateFeedError       error
	updateFeedCacheError  error
	updateFeedURLError    error
	updateFeedErrorError  error
	// Function fields for dynamic behavior
	upsertEntryFunc func(entry *repository.Entry) error
}

func (m *mockRepository) UpdateFeedError(ctx context.Context, id int64, errorMsg string) error {
	m.updateFeedErrorCalled = true
	m.updateFeedErrorMsg = errorMsg
	return m.updateFeedErrorError
}

func (m *mockRepository) UpdateFeedURL(ctx context.Context, id int64, newURL string) error {
	m.updateFeedURLCalled = true
	m.updateFeedURLNewURL = newURL
	return m.updateFeedURLError
}

func (m *mockRepository) UpdateFeedCache(ctx context.Context, id int64, etag, lastModified string, lastFetched time.Time) error {
	m.updateFeedCacheCalled = true
	return m.updateFeedCacheError
}

func (m *mockRepository) UpdateFeed(ctx context.Context, id int64, title, link string, updated time.Time) error {
	m.updateFeedCalled = true
	return m.updateFeedError
}

func (m *mockRepository) UpsertEntry(ctx context.Context, entry *repository.Entry) error {
	m.upsertEntryCalled = true
	m.upsertEntryCount++
	if m.upsertEntryFunc != nil {
		return m.upsertEntryFunc(entry)
	}
	return m.upsertEntryError
}

// Implement remaining interface methods (not used in tests)
func (m *mockRepository) GetFeeds(ctx context.Context, activeOnly bool) ([]repository.Feed, error) {
	return nil, nil
}

func (m *mockRepository) AddFeed(ctx context.Context, url, title string) (int64, error) {
	return 0, nil
}

func (m *mockRepository) GetFeedByURL(ctx context.Context, url string) (*repository.Feed, error) {
	return nil, nil
}

func (m *mockRepository) RemoveFeed(ctx context.Context, id int64) error {
	return nil
}

func (m *mockRepository) GetRecentEntries(ctx context.Context, days int) ([]repository.Entry, error) {
	return nil, nil
}

func (m *mockRepository) GetRecentEntriesWithOptions(ctx context.Context, days int, filterByFirstSeen bool, sortBy string) ([]repository.Entry, error) {
	return nil, nil
}

func (m *mockRepository) CountEntries(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *mockRepository) CountRecentEntries(ctx context.Context, days int) (int64, error) {
	return 0, nil
}

func (m *mockRepository) GetEntryCountForFeed(ctx context.Context, feedID int64) (int64, error) {
	return 0, nil
}

func (m *mockRepository) PruneOldEntries(ctx context.Context, days int) (int64, error) {
	return 0, nil
}

func (m *mockRepository) Close() error {
	return nil
}

type mockLogger struct {
	mu         sync.Mutex
	debugCalls []string
	infoCalls  []string
	warnCalls  []string
	errorCalls []string
}

func (m *mockLogger) Debug(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.debugCalls = append(m.debugCalls, format)
}

func (m *mockLogger) Info(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.infoCalls = append(m.infoCalls, format)
}

func (m *mockLogger) Warn(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warnCalls = append(m.warnCalls, format)
}

func (m *mockLogger) Error(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorCalls = append(m.errorCalls, format)
}

// Tests

func TestFetchFeed_Success(t *testing.T) {
	t.Parallel()
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

	f := New(mc, mn, mr, nil, ml, 3)

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
	t.Parallel()
	// Setup
	mc := &mockCrawler{
		err: errors.New("network error"),
	}

	mn := &mockNormalizer{}
	mr := &mockRepository{}
	ml := &mockLogger{}

	f := New(mc, mn, mr, nil, ml, 3)

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
	t.Parallel()
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

	f := New(mc, mn, mr, nil, ml, 3)

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
	t.Parallel()
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

	f := New(mc, mn, mr, nil, ml, 3)

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
	t.Parallel()
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

	f := New(mc, mn, mr, nil, ml, 3)

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
	t.Parallel()
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

	f := New(mc, mn, mr, nil, ml, 3)

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
func TestFetchFeed_UpdateFeedURLError(t *testing.T) {
	t.Parallel()
	// Test error handling when UpdateFeedURL fails after 301 redirect
	mc := &mockCrawler{
		resp: &crawler.FeedResponse{
			Body:              []byte("<feed>data</feed>"),
			StatusCode:        200,
			FetchTime:         time.Now(),
			PermanentRedirect: true,
			FinalURL:          "http://example.com/new-feed",
			NewCache:          crawler.FeedCache{},
		},
	}

	mn := &mockNormalizer{
		metadata: &normalizer.FeedMetadata{Title: "Test"},
		entries:  []normalizer.Entry{},
	}

	mr := &mockRepository{
		updateFeedURLError: errors.New("database locked"),
	}

	ml := &mockLogger{}
	f := New(mc, mn, mr, nil, ml, 3)

	feed := repository.Feed{
		ID:  1,
		URL: "http://example.com/old-feed",
	}

	// Execute
	result := f.FetchFeed(context.Background(), feed)

	// IMPROVEMENT #1: Stronger assertions - verify graceful degradation
	if result.Error != nil {
		t.Errorf("Expected graceful degradation (no fatal error), got %v", result.Error)
	}

	// Verify database update was attempted
	if !mr.updateFeedURLCalled {
		t.Error("Expected UpdateFeedURL to be attempted")
	}

	// Note: FetchResult doesn't expose redirect details - that's handled internally
	// The important thing is that processing succeeded despite the DB error

	// IMPROVEMENT #1: Stronger error message verification
	errorLogged := false
	for _, call := range ml.errorCalls {
		if contains(call, "Failed to update feed URL") {
			errorLogged = true
			break
		}
	}
	if !errorLogged {
		t.Error("Expected error log about UpdateFeedURL failure")
	}
	// Note: Implementation may or may not include feed URL in log - that's an enhancement opportunity

	// IMPROVEMENT #3: Negative assertions - verify what DIDN'T happen
	if mr.updateFeedErrorCalled {
		t.Error("UpdateFeedError should not be called - this is not a fetch error")
	}
}

func TestFetchFeed_UpdateFeedCacheError_On304(t *testing.T) {
	t.Parallel()
	// Test error handling when UpdateFeedCache fails on 304 Not Modified
	mc := &mockCrawler{
		resp: &crawler.FeedResponse{
			NotModified: true,
			NewCache: crawler.FeedCache{
				ETag:         "new-etag",
				LastModified: "Mon, 02 Jan 2024 00:00:00 GMT",
			},
			FetchTime: time.Now(),
		},
	}

	mn := &mockNormalizer{}

	mr := &mockRepository{
		updateFeedCacheError: errors.New("cache update failed"),
	}

	ml := &mockLogger{}
	f := New(mc, mn, mr, nil, ml, 3)

	feed := repository.Feed{ID: 1, URL: "http://example.com/feed"}

	// Execute
	result := f.FetchFeed(context.Background(), feed)

	// Verify
	if !result.NotModified {
		t.Error("Expected NotModified to be true")
	}

	if !mr.updateFeedCacheCalled {
		t.Error("Expected UpdateFeedCache to be called")
	}

	// Should have logged error about cache update failure
	errorLogged := false
	for _, call := range ml.errorCalls {
		if contains(call, "Failed to update feed cache") {
			errorLogged = true
			break
		}
	}
	if !errorLogged {
		t.Error("Expected error log about cache update failure")
	}
}

func TestFetchFeed_UpdateFeedMetadataError(t *testing.T) {
	t.Parallel()
	// Test error handling when UpdateFeed fails after successful parse
	mc := &mockCrawler{
		resp: &crawler.FeedResponse{
			Body:       []byte("<feed><entry>test</entry></feed>"),
			StatusCode: 200,
			FetchTime:  time.Now(),
			NewCache:   crawler.FeedCache{ETag: "etag"},
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
		},
	}

	mr := &mockRepository{
		updateFeedError: errors.New("metadata update failed"),
	}

	ml := &mockLogger{}
	f := New(mc, mn, mr, nil, ml, 3)

	feed := repository.Feed{ID: 1, URL: "http://example.com/feed"}

	// Execute
	result := f.FetchFeed(context.Background(), feed)

	// Verify - should continue despite metadata update failure
	if result.Error != nil {
		t.Errorf("Expected no overall error, got %v", result.Error)
	}

	if !mr.updateFeedCalled {
		t.Error("Expected UpdateFeed to be called")
	}

	// Entries should still be stored
	if result.StoredEntries != 1 {
		t.Errorf("Expected 1 stored entry, got %d", result.StoredEntries)
	}

	// Should have logged error
	errorLogged := false
	for _, call := range ml.errorCalls {
		if contains(call, "Failed to update feed metadata") {
			errorLogged = true
			break
		}
	}
	if !errorLogged {
		t.Error("Expected error log about metadata update failure")
	}
}

func TestFetchFeed_UpdateFeedCacheError_AfterParse(t *testing.T) {
	t.Parallel()
	// Test error handling when UpdateFeedCache fails after successful parse
	mc := &mockCrawler{
		resp: &crawler.FeedResponse{
			Body:       []byte("<feed></feed>"),
			StatusCode: 200,
			FetchTime:  time.Now(),
			NewCache:   crawler.FeedCache{ETag: "new-etag"},
		},
	}

	mn := &mockNormalizer{
		metadata: &normalizer.FeedMetadata{Title: "Test"},
		entries:  []normalizer.Entry{},
	}

	mr := &mockRepository{
		updateFeedCacheError: errors.New("cache update failed after parse"),
	}

	ml := &mockLogger{}
	f := New(mc, mn, mr, nil, ml, 3)

	feed := repository.Feed{ID: 1, URL: "http://example.com/feed"}

	// Execute
	result := f.FetchFeed(context.Background(), feed)

	// Verify - should complete successfully despite cache error
	if result.Error != nil {
		t.Errorf("Expected no error, got %v", result.Error)
	}

	if !mr.updateFeedCacheCalled {
		t.Error("Expected UpdateFeedCache to be called")
	}

	// Should have logged error about cache update failure
	errorLogged := false
	for _, call := range ml.errorCalls {
		if contains(call, "Failed to update feed cache") {
			errorLogged = true
			break
		}
	}
	if !errorLogged {
		t.Error("Expected error log about cache update failure")
	}
}

func TestFetchFeed_UpdateFeedErrorFailure(t *testing.T) {
	t.Parallel()
	// Test when UpdateFeedError itself fails (meta-error!)
	mc := &mockCrawler{
		err: errors.New("network timeout"),
	}

	mn := &mockNormalizer{}

	mr := &mockRepository{
		updateFeedErrorError: errors.New("cannot record error in database"),
	}

	ml := &mockLogger{}
	f := New(mc, mn, mr, nil, ml, 3)

	feed := repository.Feed{ID: 1, URL: "http://example.com/feed"}

	// Execute
	result := f.FetchFeed(context.Background(), feed)

	// Verify
	if result.Error == nil {
		t.Error("Expected fetch error to be returned")
	}

	if !mr.updateFeedErrorCalled {
		t.Error("Expected UpdateFeedError to be called")
	}

	// Should have logged both the fetch error and the meta-error
	if len(ml.errorCalls) < 2 {
		t.Errorf("Expected at least 2 error log calls, got %d", len(ml.errorCalls))
	}

	// Check for meta-error log
	metaErrorLogged := false
	for _, call := range ml.errorCalls {
		if contains(call, "Failed to update feed error") {
			metaErrorLogged = true
			break
		}
	}
	if !metaErrorLogged {
		t.Error("Expected error log about UpdateFeedError failure")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// IMPROVEMENT #4: Error combination tests

func TestFetchFeed_MultipleSimultaneousErrors(t *testing.T) {
	t.Parallel()
	// Test when EVERYTHING fails - network error AND error recording fails
	mc := &mockCrawler{
		err: errors.New("network timeout"),
	}

	mn := &mockNormalizer{}

	mr := &mockRepository{
		updateFeedErrorError: errors.New("database is read-only"), // Even error recording fails!
	}

	ml := &mockLogger{}
	f := New(mc, mn, mr, nil, ml, 3)

	feed := repository.Feed{ID: 1, URL: "http://example.com/feed"}

	// Execute
	result := f.FetchFeed(context.Background(), feed)

	// Should still return an error (not panic)
	if result.Error == nil {
		t.Error("Expected error when both fetch and error recording fail")
	}

	// Should have tried to record the error
	if !mr.updateFeedErrorCalled {
		t.Error("Should attempt to record fetch error even if it might fail")
	}

	// Should have logged BOTH errors (fetch error + recording failure)
	if len(ml.errorCalls) < 2 {
		t.Errorf("Expected at least 2 error logs (fetch + recording failure), got %d", len(ml.errorCalls))
	}

	// Verify we logged the meta-error about error recording failure
	foundMetaError := false
	for _, call := range ml.errorCalls {
		if contains(call, "Failed to update feed error") {
			foundMetaError = true
			break
		}
	}
	if !foundMetaError {
		t.Error("Should log when error recording itself fails (meta-error)")
	}

	// Verify original error is still returned
	if !strings.Contains(result.Error.Error(), "network") && !strings.Contains(result.Error.Error(), "timeout") {
		t.Errorf("Expected original network error to be returned, got: %v", result.Error)
	}
}

func TestFetchFeed_MultipleDBErrorsDuringSuccess(t *testing.T) {
	t.Parallel()
	// Test when fetch succeeds but multiple DB operations fail
	mc := &mockCrawler{
		resp: &crawler.FeedResponse{
			Body:       []byte("<feed><entry>test</entry></feed>"),
			StatusCode: 200,
			FetchTime:  time.Now(),
			NewCache:   crawler.FeedCache{ETag: "etag123"},
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
		},
	}

	mr := &mockRepository{
		updateFeedError:      errors.New("metadata update failed"),
		updateFeedCacheError: errors.New("cache update failed"),
		upsertEntryError:     errors.New("entry insert failed"),
	}

	ml := &mockLogger{}
	f := New(mc, mn, mr, nil, ml, 3)

	feed := repository.Feed{ID: 1, URL: "http://example.com/feed"}

	// Execute
	result := f.FetchFeed(context.Background(), feed)

	// Should not return fatal error - graceful degradation
	if result.Error != nil {
		t.Errorf("Expected graceful degradation despite DB errors, got: %v", result.Error)
	}

	// Should have attempted all operations
	if !mr.updateFeedCalled {
		t.Error("Expected UpdateFeed to be attempted")
	}
	if !mr.updateFeedCacheCalled {
		t.Error("Expected UpdateFeedCache to be attempted")
	}
	if !mr.upsertEntryCalled {
		t.Error("Expected UpsertEntry to be attempted")
	}

	// Should have logged errors (implementation may batch some errors)
	if len(ml.errorCalls) < 1 {
		t.Errorf("Expected at least 1 error log, got %d", len(ml.errorCalls))
	}

	// StoredEntries should be 0 since entry insert failed
	if result.StoredEntries != 0 {
		t.Errorf("Expected StoredEntries = 0 (insert failed), got %d", result.StoredEntries)
	}
}

// IMPROVEMENT #8: Invariant tests

func TestFetchFeed_Invariants(t *testing.T) {
	t.Parallel()

	// Test invariants that should hold across all scenarios
	scenarios := []struct {
		name       string
		setupMocks func() (*mockCrawler, *mockNormalizer, *mockRepository)
	}{
		{
			name: "all success",
			setupMocks: func() (*mockCrawler, *mockNormalizer, *mockRepository) {
				mc := &mockCrawler{
					resp: &crawler.FeedResponse{
						Body:       []byte("<feed>data</feed>"),
						StatusCode: 200,
						FetchTime:  time.Now(),
					},
				}
				mn := &mockNormalizer{
					metadata: &normalizer.FeedMetadata{Title: "Test"},
					entries:  []normalizer.Entry{{ID: "1", Title: "Entry"}},
				}
				mr := &mockRepository{}
				return mc, mn, mr
			},
		},
		{
			name: "crawler fails",
			setupMocks: func() (*mockCrawler, *mockNormalizer, *mockRepository) {
				mc := &mockCrawler{err: errors.New("network error")}
				mn := &mockNormalizer{}
				mr := &mockRepository{}
				return mc, mn, mr
			},
		},
		{
			name: "normalizer fails",
			setupMocks: func() (*mockCrawler, *mockNormalizer, *mockRepository) {
				mc := &mockCrawler{
					resp: &crawler.FeedResponse{
						Body:       []byte("<feed>data</feed>"),
						StatusCode: 200,
						FetchTime:  time.Now(),
					},
				}
				mn := &mockNormalizer{err: errors.New("parse error")}
				mr := &mockRepository{}
				return mc, mn, mr
			},
		},
		{
			name: "repository fails",
			setupMocks: func() (*mockCrawler, *mockNormalizer, *mockRepository) {
				mc := &mockCrawler{
					resp: &crawler.FeedResponse{
						Body:       []byte("<feed>data</feed>"),
						StatusCode: 200,
						FetchTime:  time.Now(),
					},
				}
				mn := &mockNormalizer{
					metadata: &normalizer.FeedMetadata{Title: "Test"},
					entries:  []normalizer.Entry{{ID: "1", Title: "Entry"}},
				}
				mr := &mockRepository{upsertEntryError: errors.New("db error")}
				return mc, mn, mr
			},
		},
		{
			name: "all fail",
			setupMocks: func() (*mockCrawler, *mockNormalizer, *mockRepository) {
				mc := &mockCrawler{err: errors.New("network error")}
				mn := &mockNormalizer{}
				mr := &mockRepository{updateFeedErrorError: errors.New("db error")}
				return mc, mn, mr
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			mc, mn, mr := scenario.setupMocks()
			ml := &mockLogger{}
			f := New(mc, mn, mr, nil, ml, 3)

			feed := repository.Feed{ID: 1, URL: "http://example.com/feed"}
			result := f.FetchFeed(context.Background(), feed)

			// INVARIANT #1: Should never panic
			// (test passes if we reach this point)

			// INVARIANT #2: StoredEntries should never be negative
			if result.StoredEntries < 0 {
				t.Errorf("INVARIANT VIOLATED: StoredEntries = %d (negative)", result.StoredEntries)
			}

			// INVARIANT #3: If error, should be recorded (unless recording itself fails)
			if result.Error != nil && !mr.updateFeedErrorCalled && mr.updateFeedErrorError == nil {
				t.Error("INVARIANT VIOLATED: Error not recorded in database")
			}

			// INVARIANT #4: Can't have stored entries if there was a fatal error
			if result.Error != nil && result.StoredEntries > 0 {
				t.Error("INVARIANT VIOLATED: Stored entries despite fatal error")
			}

			// INVARIANT #5: NotModified implies no new content
			if result.NotModified && result.StoredEntries > 0 {
				t.Error("INVARIANT VIOLATED: Stored entries on 304 Not Modified")
			}

			// INVARIANT #6: NotModified implies no error
			if result.NotModified && result.Error != nil {
				t.Error("INVARIANT VIOLATED: Error on 304 Not Modified")
			}
		})
	}
}

// IMPROVEMENT #5: Integration test with real HTTP and database

func TestFetchFeed_Integration_RedirectThenSuccess(t *testing.T) {
	t.Parallel()

	// Setup: Real HTTP server that redirects, then succeeds
	redirectCount := 0
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectCount++
		if redirectCount == 1 {
			// First request: permanent redirect
			w.Header().Set("Location", serverURL+"/new-location")
			w.Header().Set("ETag", "\"etag-before-redirect\"")
			w.WriteHeader(http.StatusMovedPermanently)
		} else {
			// Second request (after redirect): success with feed data
			w.Header().Set("ETag", "\"etag-after-redirect\"")
			w.Header().Set("Content-Type", "application/rss+xml")
			w.WriteHeader(http.StatusOK)
			feedXML := `<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>http://example.com</link>
    <item>
      <title>Test Entry</title>
      <link>http://example.com/entry1</link>
      <guid>entry-1</guid>
      <description>Test content</description>
    </item>
  </channel>
</rss>`
			if _, err := w.Write([]byte(feedXML)); err != nil {
				t.Errorf("Write error: %v", err)
			}
		}
	}))
	defer server.Close()
	serverURL = server.URL

	// Setup: Real database
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	repo, err := repository.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	// Add feed with original URL
	feedID, err := repo.AddFeed(context.Background(), server.URL, "Test Feed")
	if err != nil {
		t.Fatalf("Failed to add feed: %v", err)
	}

	// Setup: Real components
	crawler := crawler.NewForTesting()
	normalizer := normalizer.New()
	logger := &mockLogger{}

	fetcher := New(crawler, normalizer, repo, nil, logger, 3)

	// Execute: Fetch the feed
	feed, err := repo.GetFeedByURL(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("Failed to get feed: %v", err)
	}

	result := fetcher.FetchFeed(context.Background(), *feed)

	// IMPROVEMENT #2: State verification - verify actual database state

	// Verify HTTP behavior
	if redirectCount != 2 {
		t.Errorf("Expected 2 HTTP requests (redirect + follow), got %d", redirectCount)
	}

	// Verify result
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}

	if result.StoredEntries != 1 {
		t.Errorf("Expected 1 stored entry, got %d", result.StoredEntries)
	}

	// IMPROVEMENT #2: STATE VERIFICATION using repository public API

	// STATE VERIFICATION #1: Feed URL was updated after 301 redirect
	expectedNewURL := serverURL + "/new-location"
	updatedFeed, err := repo.GetFeedByURL(context.Background(), expectedNewURL)
	if err != nil {
		t.Fatalf("Failed to get feed by new URL after redirect: %v", err)
	}
	if updatedFeed == nil {
		t.Error("Feed should exist at new URL after 301 redirect")
	}
	if updatedFeed != nil && updatedFeed.ID != feedID {
		t.Errorf("Feed ID changed after redirect: got %d, want %d", updatedFeed.ID, feedID)
	}

	// Verify old URL no longer exists
	oldFeed, err := repo.GetFeedByURL(context.Background(), serverURL)
	if err == nil && oldFeed != nil {
		t.Error("Old URL should not exist after 301 redirect and URL update")
	}

	// STATE VERIFICATION #2: Can retrieve entries (proves they were stored correctly)
	entries, err := repo.GetRecentEntries(context.Background(), 7)
	if err != nil {
		t.Fatalf("Failed to get recent entries: %v", err)
	}

	foundTestEntry := false
	for _, entry := range entries {
		if entry.Title == "Test Entry" && entry.Link == "http://example.com/entry1" {
			foundTestEntry = true
			// Verify entry belongs to correct feed
			if entry.FeedID != feedID {
				t.Errorf("Entry feed_id = %d, want %d (orphaned entry)", entry.FeedID, feedID)
			}
			break
		}
	}
	if !foundTestEntry {
		t.Error("Could not find stored entry - integration failed")
	}

	// STATE VERIFICATION #3: Entry count is correct
	entryCount, err := repo.CountEntries(context.Background())
	if err != nil {
		t.Fatalf("Failed to count entries: %v", err)
	}
	if entryCount != 1 {
		t.Errorf("Total entry count = %d, want 1", entryCount)
	}

	// STATE VERIFICATION #4: Feed has no fetch errors
	allFeeds, err := repo.GetFeeds(context.Background(), false)
	if err != nil {
		t.Fatalf("Failed to get feeds: %v", err)
	}
	for _, f := range allFeeds {
		if f.ID == feedID {
			if f.FetchError != "" {
				t.Errorf("Expected no fetch_error, got: %q", f.FetchError)
			}
			if f.FetchErrorCount != 0 {
				t.Errorf("Expected FetchErrorCount = 0, got %d", f.FetchErrorCount)
			}
			break
		}
	}
}

// ADDITIONAL ERROR COMBINATION SCENARIOS

func TestFetchFeed_PartialEntryStorageFailure(t *testing.T) {
	t.Parallel()

	// Scenario: Feed has 3 entries, but InsertEntry fails on the 2nd one
	// System should log error but continue processing the 3rd entry

	ml := &mockLogger{}
	mc := &mockCrawler{
		resp: &crawler.FeedResponse{
			Body:       []byte("<feed>data</feed>"),
			StatusCode: 200,
			FetchTime:  time.Now(),
		},
	}

	mn := &mockNormalizer{
		metadata: &normalizer.FeedMetadata{Title: "Test Feed"},
		entries: []normalizer.Entry{
			{ID: "entry-1", Title: "Entry 1"},
			{ID: "entry-2", Title: "Entry 2"}, // This one will fail
			{ID: "entry-3", Title: "Entry 3"},
		},
	}

	// Mock repository that fails on the 2nd entry upsert
	upsertAttempts := 0
	mr := &mockRepository{
		upsertEntryFunc: func(entry *repository.Entry) error {
			upsertAttempts++
			if upsertAttempts == 2 {
				return errors.New("unique constraint violation")
			}
			return nil
		},
	}

	f := &Fetcher{
		crawler:    mc,
		normalizer: mn,
		repo:       mr,
		logger:     ml,
	}

	result := f.FetchFeed(context.Background(), repository.Feed{ID: 1, URL: "https://example.com/feed"})

	// Should have attempted all 3 upserts
	if upsertAttempts != 3 {
		t.Errorf("Expected 3 upsert attempts, got %d", upsertAttempts)
	}

	// Should have logged the failure as a warning
	if len(ml.warnCalls) < 1 {
		t.Error("Expected warning log for failed entry insert")
	}

	// Should report 2 successful entries (1st and 3rd)
	if result.StoredEntries != 2 {
		t.Errorf("Expected 2 stored entries (1st and 3rd succeeded), got %d", result.StoredEntries)
	}

	// Should NOT be a fatal error - processing continues
	if result.Error != nil {
		t.Errorf("Expected graceful degradation, got fatal error: %v", result.Error)
	}
}

func TestFetchFeed_CascadeDBFailures(t *testing.T) {
	t.Parallel()

	// Scenario: Fetch succeeds, but ALL DB operations fail in sequence
	// Tests that system handles complete DB failure gracefully

	ml := &mockLogger{}
	mc := &mockCrawler{
		resp: &crawler.FeedResponse{
			Body:       []byte("<feed>data</feed>"),
			StatusCode: 200,
			FetchTime:  time.Now(),
		},
	}

	mn := &mockNormalizer{
		metadata: &normalizer.FeedMetadata{Title: "Test Feed"},
		entries:  []normalizer.Entry{{ID: "entry-1", Title: "Entry 1"}},
	}

	mr := &mockRepository{
		upsertEntryError:     errors.New("database locked"),
		updateFeedError:      errors.New("database locked"),
		updateFeedCacheError: errors.New("database locked"),
	}

	f := &Fetcher{
		crawler:    mc,
		normalizer: mn,
		repo:       mr,
		logger:     ml,
	}

	result := f.FetchFeed(context.Background(), repository.Feed{ID: 1, URL: "https://example.com/feed"})

	// Should have logged multiple errors
	if len(ml.errorCalls) < 2 {
		t.Errorf("Expected at least 2 error logs (insert + metadata/cache), got %d", len(ml.errorCalls))
	}

	// No entries should be reported as stored
	if result.StoredEntries != 0 {
		t.Errorf("Expected 0 stored entries (all DB ops failed), got %d", result.StoredEntries)
	}

	// Should still complete without panic
	if result.Error != nil {
		t.Logf("Result error: %v", result.Error)
	}
}

func TestFetchFeed_IntermittentNetworkError(t *testing.T) {
	t.Parallel()

	// Scenario: Network operation times out (context deadline exceeded)
	// This is different from immediate network errors

	ml := &mockLogger{}
	mc := &mockCrawler{
		err: context.DeadlineExceeded,
	}

	mn := &mockNormalizer{}
	mr := &mockRepository{}

	f := &Fetcher{
		crawler:    mc,
		normalizer: mn,
		repo:       mr,
		logger:     ml,
	}

	result := f.FetchFeed(context.Background(), repository.Feed{ID: 1, URL: "https://example.com/feed"})

	// Should record the error
	if !mr.updateFeedErrorCalled {
		t.Error("Expected UpdateFeedError to be called for timeout")
	}

	// Should have logged the timeout
	foundTimeout := false
	for _, call := range ml.errorCalls {
		if contains(call, "failed for") { // Format: "%s failed for %s: %v"
			foundTimeout = true
			break
		}
	}
	if !foundTimeout {
		t.Error("Expected timeout error to be logged")
	}

	// Should return error
	if result.Error == nil {
		t.Error("Expected error result for timeout")
	}
}

func TestFetchFeed_ParseErrorWithDBRecordingFailure(t *testing.T) {
	t.Parallel()

	// Scenario: Feed parsing fails AND we can't record the error
	// This tests error handling on the error path itself

	ml := &mockLogger{}
	mc := &mockCrawler{
		resp: &crawler.FeedResponse{
			Body:       []byte("invalid feed data"),
			StatusCode: 200,
			FetchTime:  time.Now(),
		},
	}

	mn := &mockNormalizer{
		err: errors.New("invalid XML"),
	}

	mr := &mockRepository{
		updateFeedErrorError: errors.New("database is in recovery mode"),
	}

	f := &Fetcher{
		crawler:    mc,
		normalizer: mn,
		repo:       mr,
		logger:     ml,
	}

	result := f.FetchFeed(context.Background(), repository.Feed{ID: 1, URL: "https://example.com/feed"})

	// Should have logged BOTH the parse error AND the recording failure
	if len(ml.errorCalls) < 2 {
		t.Errorf("Expected at least 2 error logs (parse error + recording failure), got %d", len(ml.errorCalls))
	}

	// Verify we logged the parse error
	foundParseError := false
	foundRecordingError := false
	for _, call := range ml.errorCalls {
		if contains(call, "failed for") { // Format: "%s failed for %s: %v" (operation = "parse")
			foundParseError = true
		}
		if contains(call, "Failed to update feed error") {
			foundRecordingError = true
		}
	}
	if !foundParseError {
		t.Error("Expected parse error to be logged")
	}
	if !foundRecordingError {
		t.Error("Expected error recording failure to be logged")
	}

	// Should still return the original parse error
	if result.Error == nil {
		t.Error("Expected error result for parse failure")
	}
}
