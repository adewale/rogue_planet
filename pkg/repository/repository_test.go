package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) (*Repository, string) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	return repo, dbPath
}

func TestNew(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	// Verify schema was created
	var count int
	err := repo.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='feeds'").Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if count != 1 {
		t.Errorf("feeds table not created")
	}
}

func TestAddFeed(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	id, err := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")
	if err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	if id == 0 {
		t.Error("Expected non-zero ID")
	}

	// Verify feed was added
	feed, err := repo.GetFeedByURL(t.Context(), "https://example.com/feed")
	if err != nil {
		t.Fatalf("GetFeedByURL() error = %v", err)
	}

	if feed.URL != "https://example.com/feed" {
		t.Errorf("URL = %q, want %q", feed.URL, "https://example.com/feed")
	}

	if feed.Title != "Test Feed" {
		t.Errorf("Title = %q, want %q", feed.Title, "Test Feed")
	}
}

func TestAddDuplicateFeed(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	_, err := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")
	if err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	// Try to add duplicate
	_, err = repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed 2")
	if err == nil {
		t.Error("Expected error for duplicate feed, got nil")
	}
}

func TestUpdateFeed(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	id, _ := repo.AddFeed(t.Context(), "https://example.com/feed", "Old Title")

	updated := time.Now()
	err := repo.UpdateFeed(t.Context(), id, "New Title", "https://example.com", updated)
	if err != nil {
		t.Fatalf("UpdateFeed() error = %v", err)
	}

	feed, _ := repo.GetFeedByURL(t.Context(), "https://example.com/feed")

	if feed.Title != "New Title" {
		t.Errorf("Title = %q, want %q", feed.Title, "New Title")
	}

	if feed.Link != "https://example.com" {
		t.Errorf("Link = %q, want %q", feed.Link, "https://example.com")
	}
}

func TestUpdateFeedCache(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	id, _ := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")

	fetchTime := time.Now()
	err := repo.UpdateFeedCache(t.Context(), id, `"etag123"`, "Mon, 02 Jan 2006 15:04:05 GMT", fetchTime)
	if err != nil {
		t.Fatalf("UpdateFeedCache() error = %v", err)
	}

	feed, _ := repo.GetFeedByURL(t.Context(), "https://example.com/feed")

	if feed.ETag != `"etag123"` {
		t.Errorf("ETag = %q, want %q", feed.ETag, `"etag123"`)
	}

	if feed.LastModified != "Mon, 02 Jan 2006 15:04:05 GMT" {
		t.Errorf("LastModified = %q, want %q", feed.LastModified, "Mon, 02 Jan 2006 15:04:05 GMT")
	}
}

func TestGetFeeds(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	if _, err := repo.AddFeed(t.Context(), "https://example.com/feed1", "Feed 1"); err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}
	if _, err := repo.AddFeed(t.Context(), "https://example.com/feed2", "Feed 2"); err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	feeds, err := repo.GetFeeds(t.Context(), false)
	if err != nil {
		t.Fatalf("GetFeeds() error = %v", err)
	}

	if len(feeds) != 2 {
		t.Errorf("len(feeds) = %d, want 2", len(feeds))
	}
}

func TestGetFeedByURL(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	if _, err := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed"); err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	feed, err := repo.GetFeedByURL(t.Context(), "https://example.com/feed")
	if err != nil {
		t.Fatalf("GetFeedByURL() error = %v", err)
	}

	if feed.URL != "https://example.com/feed" {
		t.Errorf("URL = %q, want %q", feed.URL, "https://example.com/feed")
	}

	// Test non-existent feed
	_, err = repo.GetFeedByURL(t.Context(), "https://example.com/nonexistent")
	if err != ErrFeedNotFound {
		t.Errorf("Expected ErrFeedNotFound, got %v", err)
	}
}

func TestRemoveFeed(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	id, _ := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")

	err := repo.RemoveFeed(t.Context(), id)
	if err != nil {
		t.Fatalf("RemoveFeed() error = %v", err)
	}

	// Verify feed was removed
	_, err = repo.GetFeedByURL(t.Context(), "https://example.com/feed")
	if err != ErrFeedNotFound {
		t.Error("Feed should have been removed")
	}
}

func TestUpsertEntry(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	feedID, _ := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")

	entry := &Entry{
		FeedID:      feedID,
		EntryID:     "entry-1",
		Title:       "Test Entry",
		Link:        "https://example.com/post1",
		Author:      "John Doe",
		Published:   time.Now(),
		Updated:     time.Now(),
		Content:     "<p>Test content</p>",
		ContentType: "html",
		Summary:     "Test summary",
		FirstSeen:   time.Now(),
	}

	err := repo.UpsertEntry(t.Context(), entry)
	if err != nil {
		t.Fatalf("UpsertEntry() error = %v", err)
	}

	// Test update
	entry.Title = "Updated Title"
	err = repo.UpsertEntry(t.Context(), entry)
	if err != nil {
		t.Fatalf("UpsertEntry() update error = %v", err)
	}

	// Verify entry was updated
	var count int
	var title string
	err = repo.db.QueryRow("SELECT COUNT(*), MAX(title) FROM entries WHERE feed_id = ?", feedID).Scan(&count, &title)
	if err != nil {
		t.Fatalf("Query error = %v", err)
	}
	if count != 1 {
		t.Fatalf("entry count = %d, want 1", count)
	}
	if title != "Updated Title" {
		t.Errorf("Title = %q, want %q", title, "Updated Title")
	}
}

func TestUniqueConstraintHandling(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	feedID, _ := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")

	// Create an entry
	entry1 := &Entry{
		FeedID:      feedID,
		EntryID:     "unique-entry-1",
		Title:       "Original Title",
		Link:        "https://example.com/entry/1",
		Content:     "Original content",
		ContentType: "html",
		Author:      "Author 1",
		Published:   time.Now().Add(-1 * time.Hour),
		Updated:     time.Now().Add(-1 * time.Hour),
		FirstSeen:   time.Now().Add(-1 * time.Hour),
	}

	err := repo.UpsertEntry(t.Context(), entry1)
	if err != nil {
		t.Fatalf("UpsertEntry() error = %v", err)
	}

	// Try to insert the same entry again with different data
	// This tests that the UNIQUE constraint on (feed_id, entry_id) triggers an UPDATE
	entry2 := &Entry{
		FeedID:      feedID,
		EntryID:     "unique-entry-1", // Same EntryID - violates unique constraint
		Title:       "Modified Title",
		Link:        "https://example.com/entry/1-modified",
		Content:     "Modified content",
		ContentType: "html",
		Author:      "Author 2",
		Published:   time.Now(),
		Updated:     time.Now(),
		FirstSeen:   time.Now(),
	}

	err = repo.UpsertEntry(t.Context(), entry2)
	if err != nil {
		t.Fatalf("UpsertEntry() should handle unique constraint gracefully, got error: %v", err)
	}

	// Verify that we still have exactly one entry (not two)
	var count int
	err = repo.db.QueryRow("SELECT COUNT(*) FROM entries WHERE feed_id = ? AND entry_id = ?",
		feedID, "unique-entry-1").Scan(&count)

	if err != nil {
		t.Fatalf("Query error: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 entry after unique constraint conflict, got %d", count)
	}

	// Verify the entry was updated (not inserted as duplicate)
	var title, author string
	err = repo.db.QueryRow("SELECT title, author FROM entries WHERE feed_id = ? AND entry_id = ?",
		feedID, "unique-entry-1").Scan(&title, &author)

	if err != nil {
		t.Fatalf("Query error: %v", err)
	}

	if title != "Modified Title" {
		t.Errorf("Title = %q, want %q (should be updated)", title, "Modified Title")
	}

	if author != "Author 2" {
		t.Errorf("Author = %q, want %q (should be updated)", author, "Author 2")
	}
}

func TestGetRecentEntries(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	feedID, _ := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")

	// Add recent entry
	recentEntry := &Entry{
		FeedID:    feedID,
		EntryID:   "recent",
		Title:     "Recent Entry",
		Published: time.Now(),
		Updated:   time.Now(),
		FirstSeen: time.Now(),
	}
	if err := repo.UpsertEntry(t.Context(), recentEntry); err != nil {
		t.Fatalf("UpsertEntry() error = %v", err)
	}

	// Add old entry
	oldEntry := &Entry{
		FeedID:    feedID,
		EntryID:   "old",
		Title:     "Old Entry",
		Published: time.Now().AddDate(0, 0, -10),
		Updated:   time.Now().AddDate(0, 0, -10),
		FirstSeen: time.Now().AddDate(0, 0, -10),
	}
	if err := repo.UpsertEntry(t.Context(), oldEntry); err != nil {
		t.Fatalf("UpsertEntry() error = %v", err)
	}

	// Get recent entries (last 7 days)
	entries, err := repo.GetRecentEntries(t.Context(), 7)
	if err != nil {
		t.Fatalf("GetRecentEntries() error = %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("len(entries) = %d, want 1", len(entries))
	}

	if entries[0].Title != "Recent Entry" {
		t.Errorf("Title = %q, want %q", entries[0].Title, "Recent Entry")
	}
}

func TestPruneOldEntries(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	feedID, _ := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")

	// Add recent entry
	recentEntry := &Entry{
		FeedID:    feedID,
		EntryID:   "recent",
		Title:     "Recent Entry",
		Published: time.Now(),
		Updated:   time.Now(),
		FirstSeen: time.Now(),
	}
	if err := repo.UpsertEntry(t.Context(), recentEntry); err != nil {
		t.Fatalf("UpsertEntry() error = %v", err)
	}

	// Add old entry
	oldEntry := &Entry{
		FeedID:    feedID,
		EntryID:   "old",
		Title:     "Old Entry",
		Published: time.Now().AddDate(0, 0, -100),
		Updated:   time.Now().AddDate(0, 0, -100),
		FirstSeen: time.Now().AddDate(0, 0, -100),
	}
	if err := repo.UpsertEntry(t.Context(), oldEntry); err != nil {
		t.Fatalf("UpsertEntry() error = %v", err)
	}

	// Prune entries older than 90 days
	deleted, err := repo.PruneOldEntries(t.Context(), 90)
	if err != nil {
		t.Fatalf("PruneOldEntries() error = %v", err)
	}

	if deleted != 1 {
		t.Errorf("deleted = %d, want 1", deleted)
	}

	// Verify only recent entry remains
	var count int
	var title string
	err = repo.db.QueryRow("SELECT COUNT(*), MAX(title) FROM entries WHERE feed_id = ?", feedID).Scan(&count, &title)
	if err != nil {
		t.Errorf("Query error = %v", err)
	}
	if count != 1 {
		t.Errorf("entry count = %d, want 1", count)
	}
	if title != "Recent Entry" {
		t.Errorf("Wrong entry remained: %q", title)
	}
}

func TestRemoveFeedCascade(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	feedID, _ := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")

	// Add entry
	entry := &Entry{
		FeedID:    feedID,
		EntryID:   "entry-1",
		Title:     "Test Entry",
		Published: time.Now(),
		Updated:   time.Now(),
		FirstSeen: time.Now(),
	}
	if err := repo.UpsertEntry(t.Context(), entry); err != nil {
		t.Fatalf("UpsertEntry() error = %v", err)
	}

	// Remove feed
	if err := repo.RemoveFeed(t.Context(), feedID); err != nil {
		t.Fatalf("RemoveFeed() error = %v", err)
	}

	// Verify entries were also removed
	var count int
	err := repo.db.QueryRow("SELECT COUNT(*) FROM entries WHERE feed_id = ?", feedID).Scan(&count)
	if err != nil {
		t.Errorf("Query error = %v", err)
	}
	if count != 0 {
		t.Error("Entries should have been cascade deleted")
	}
}

func TestDatabasePersistence(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "persist.db")

	// Create repository and add data
	repo1, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	if _, err := repo1.AddFeed(t.Context(), "https://example.com/feed", "Test Feed"); err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}
	_ = repo1.Close()

	// Reopen database
	repo2, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to reopen repository: %v", err)
	}
	defer func() { _ = repo2.Close() }()

	// Verify data persisted
	feeds, _ := repo2.GetFeeds(t.Context(), false)
	if len(feeds) != 1 {
		t.Errorf("Data did not persist: len(feeds) = %d, want 1", len(feeds))
	}

	// Verify database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file should exist")
	}
}

func TestGetRecentEntriesFallback(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	// Add a feed
	feedID, err := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")
	if err != nil {
		t.Fatalf("Failed to add feed: %v", err)
	}

	// Add entries with old publish dates (100 days ago)
	oldDate := time.Now().AddDate(0, 0, -100)
	for i := range 10 {
		entry := &Entry{
			FeedID:    feedID,
			EntryID:   fmt.Sprintf("entry-%d", i),
			Title:     fmt.Sprintf("Old Entry %d", i),
			Link:      fmt.Sprintf("https://example.com/entry%d", i),
			Published: oldDate.Add(time.Duration(i) * time.Hour),
			Updated:   oldDate.Add(time.Duration(i) * time.Hour),
			FirstSeen: time.Now(),
		}
		if err := repo.UpsertEntry(t.Context(), entry); err != nil {
			t.Fatalf("Failed to upsert entry: %v", err)
		}
	}

	// Test 1: Requesting entries from last 7 days should return 0 (within window)
	// But with fallback, should return the old entries
	entries, err := repo.GetRecentEntries(t.Context(), 7)
	if err != nil {
		t.Fatalf("GetRecentEntries failed: %v", err)
	}

	// Should fall back to most recent entries
	if len(entries) == 0 {
		t.Error("GetRecentEntries should fall back to old entries when no recent ones exist")
	}

	if len(entries) != 10 {
		t.Errorf("GetRecentEntries fallback returned %d entries, want 10", len(entries))
	}

	// Verify they're sorted by published date (most recent first)
	for i := 1; i < len(entries); i++ {
		if entries[i].Published.After(entries[i-1].Published) {
			t.Error("Entries should be sorted by published date DESC")
		}
	}
}

func TestGetRecentEntriesWithinWindow(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	// Add a feed
	feedID, err := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")
	if err != nil {
		t.Fatalf("Failed to add feed: %v", err)
	}

	// Add recent entries (within last 7 days)
	now := time.Now()
	for i := range 5 {
		entry := &Entry{
			FeedID:    feedID,
			EntryID:   fmt.Sprintf("recent-%d", i),
			Title:     fmt.Sprintf("Recent Entry %d", i),
			Link:      fmt.Sprintf("https://example.com/recent%d", i),
			Published: now.Add(time.Duration(-i) * 24 * time.Hour), // Last 5 days
			Updated:   now.Add(time.Duration(-i) * 24 * time.Hour),
			FirstSeen: now,
		}
		if err := repo.UpsertEntry(t.Context(), entry); err != nil {
			t.Fatalf("Failed to upsert entry: %v", err)
		}
	}

	// Add old entries (100 days ago)
	oldDate := time.Now().AddDate(0, 0, -100)
	for i := range 5 {
		entry := &Entry{
			FeedID:    feedID,
			EntryID:   fmt.Sprintf("old-%d", i),
			Title:     fmt.Sprintf("Old Entry %d", i),
			Link:      fmt.Sprintf("https://example.com/old%d", i),
			Published: oldDate.Add(time.Duration(i) * time.Hour),
			Updated:   oldDate.Add(time.Duration(i) * time.Hour),
			FirstSeen: now,
		}
		if err := repo.UpsertEntry(t.Context(), entry); err != nil {
			t.Fatalf("Failed to upsert entry: %v", err)
		}
	}

	// Request entries from last 7 days
	entries, err := repo.GetRecentEntries(t.Context(), 7)
	if err != nil {
		t.Fatalf("GetRecentEntries failed: %v", err)
	}

	// Should only get the recent entries, not old ones
	if len(entries) != 5 {
		t.Errorf("GetRecentEntries returned %d entries, want 5 (only recent ones)", len(entries))
	}

	// Verify we got recent entries, not old ones
	for _, entry := range entries {
		if !strings.HasPrefix(entry.EntryID, "recent-") {
			t.Errorf("Expected only recent entries, got %s", entry.EntryID)
		}
	}
}

func TestUpdateFeedError(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	id, _ := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")

	err := repo.UpdateFeedError(t.Context(), id, "Connection timeout")
	if err != nil {
		t.Fatalf("UpdateFeedError() error = %v", err)
	}

	feed, _ := repo.GetFeedByURL(t.Context(), "https://example.com/feed")

	if feed.FetchError != "Connection timeout" {
		t.Errorf("FetchError = %q, want %q", feed.FetchError, "Connection timeout")
	}

	if feed.FetchErrorCount != 1 {
		t.Errorf("FetchErrorCount = %d, want 1", feed.FetchErrorCount)
	}

	// Call again to increment error count
	if err := repo.UpdateFeedError(t.Context(), id, "Another error"); err != nil {
		t.Fatalf("UpdateFeedError() error = %v", err)
	}
	feed, _ = repo.GetFeedByURL(t.Context(), "https://example.com/feed")

	if feed.FetchErrorCount != 2 {
		t.Errorf("FetchErrorCount = %d, want 2", feed.FetchErrorCount)
	}
}

func TestCountEntries(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	// Initially should be 0
	count, err := repo.CountEntries(t.Context())
	if err != nil {
		t.Fatalf("CountEntries() error = %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	// Add some entries
	feedID, _ := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")
	for i := range 5 {
		entry := &Entry{
			FeedID:    feedID,
			EntryID:   fmt.Sprintf("entry-%d", i),
			Title:     fmt.Sprintf("Entry %d", i),
			Published: time.Now(),
			Updated:   time.Now(),
			FirstSeen: time.Now(),
		}
		if err := repo.UpsertEntry(t.Context(), entry); err != nil {
			t.Fatalf("UpsertEntry() error = %v", err)
		}
	}

	count, err = repo.CountEntries(t.Context())
	if err != nil {
		t.Fatalf("CountEntries() error = %v", err)
	}
	if count != 5 {
		t.Errorf("count = %d, want 5", count)
	}
}

func TestCountRecentEntries(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	feedID, _ := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")

	// Add recent entries (last 3 days)
	now := time.Now()
	for i := range 3 {
		entry := &Entry{
			FeedID:    feedID,
			EntryID:   fmt.Sprintf("recent-%d", i),
			Title:     fmt.Sprintf("Recent %d", i),
			Published: now.Add(time.Duration(-i) * 24 * time.Hour),
			Updated:   now,
			FirstSeen: now,
		}
		if err := repo.UpsertEntry(t.Context(), entry); err != nil {
			t.Fatalf("UpsertEntry() error = %v", err)
		}
	}

	// Add old entries (100 days ago)
	oldDate := now.AddDate(0, 0, -100)
	for i := range 2 {
		entry := &Entry{
			FeedID:    feedID,
			EntryID:   fmt.Sprintf("old-%d", i),
			Title:     fmt.Sprintf("Old %d", i),
			Published: oldDate,
			Updated:   oldDate,
			FirstSeen: now,
		}
		if err := repo.UpsertEntry(t.Context(), entry); err != nil {
			t.Fatalf("UpsertEntry() error = %v", err)
		}
	}

	// Count recent entries (last 7 days)
	count, err := repo.CountRecentEntries(t.Context(), 7)
	if err != nil {
		t.Fatalf("CountRecentEntries() error = %v", err)
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}

	// Count last 200 days (should include all)
	count, err = repo.CountRecentEntries(t.Context(), 200)
	if err != nil {
		t.Fatalf("CountRecentEntries() error = %v", err)
	}
	if count != 5 {
		t.Errorf("count = %d, want 5", count)
	}
}

func TestNewErrors(t *testing.T) {
	t.Parallel()
	// Test with invalid path
	_, err := New("/invalid/path/to/nonexistent/dir/test.db")
	if err == nil {
		t.Error("New() should fail with invalid path")
	}
}

func TestGetRecentEntriesFilterByFirstSeen(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	// Add a feed
	feedID, _ := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")

	// Create entries with different published and first_seen dates
	// Use current time as base for testing
	baseTime := time.Now()

	entries := []struct {
		published time.Time // Original published date
		firstSeen time.Time // When aggregator saw it
	}{
		{baseTime.AddDate(0, 0, -30), baseTime.AddDate(0, 0, -1)}, // Old entry, recently seen
		{baseTime.AddDate(0, 0, -2), baseTime.AddDate(0, 0, -2)},  // Recent entry, recently seen
		{baseTime.AddDate(0, 0, -3), baseTime.AddDate(0, 0, -10)}, // Recent entry, seen long ago
	}

	for i, e := range entries {
		err := repo.UpsertEntry(t.Context(), &Entry{
			FeedID:    feedID,
			EntryID:   fmt.Sprintf("entry-%d", i),
			Title:     fmt.Sprintf("Entry %d", i),
			Published: e.published,
			FirstSeen: e.firstSeen,
		})
		if err != nil {
			t.Fatalf("UpsertEntry() error = %v", err)
		}
	}

	// Test 1: Filter by published (default behavior)
	// Should return entries 1 and 2 (published within 7 days)
	publishedFiltered, err := repo.GetRecentEntriesWithOptions(t.Context(), 7, false, "published")
	if err != nil {
		t.Fatalf("GetRecentEntriesWithOptions() error = %v", err)
	}
	if len(publishedFiltered) != 2 {
		t.Errorf("Filter by published: got %d entries, want 2", len(publishedFiltered))
	}

	// Test 2: Filter by first_seen
	// Should return entries 0 and 1 (first_seen within 7 days)
	firstSeenFiltered, err := repo.GetRecentEntriesWithOptions(t.Context(), 7, true, "published")
	if err != nil {
		t.Fatalf("GetRecentEntriesWithOptions() error = %v", err)
	}
	if len(firstSeenFiltered) != 2 {
		t.Errorf("Filter by first_seen: got %d entries, want 2", len(firstSeenFiltered))
	}

	// Verify which entries were returned
	titles := make(map[string]bool)
	for _, e := range firstSeenFiltered {
		titles[e.Title] = true
	}

	if !titles["Entry 0"] || !titles["Entry 1"] {
		t.Errorf("Filter by first_seen returned wrong entries: %v", titles)
	}
	if titles["Entry 2"] {
		t.Errorf("Filter by first_seen should not include Entry 2 (first_seen too old)")
	}
}

func TestGetRecentEntriesSortByFirstSeen(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	feedID, _ := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")
	baseTime := time.Now()

	// Create entries where first_seen order differs from published order
	entries := []struct {
		title     string
		published time.Time
		firstSeen time.Time
	}{
		{"Entry A", baseTime.AddDate(0, 0, -1), baseTime.AddDate(0, 0, -3)}, // Published recently, seen first
		{"Entry B", baseTime.AddDate(0, 0, -2), baseTime.AddDate(0, 0, -2)}, // Published middle, seen second
		{"Entry C", baseTime.AddDate(0, 0, -3), baseTime.AddDate(0, 0, -1)}, // Published oldest, seen last
	}

	for i, e := range entries {
		err := repo.UpsertEntry(t.Context(), &Entry{
			FeedID:    feedID,
			EntryID:   fmt.Sprintf("entry-%d", i),
			Title:     e.title,
			Published: e.published,
			FirstSeen: e.firstSeen,
		})
		if err != nil {
			t.Fatalf("UpsertEntry() error = %v", err)
		}
	}

	// Sort by published (default)
	byPublished, _ := repo.GetRecentEntriesWithOptions(t.Context(), 7, false, "published")
	if byPublished[0].Title != "Entry A" {
		t.Errorf("Sort by published: first entry = %s, want Entry A", byPublished[0].Title)
	}

	// Sort by first_seen
	byFirstSeen, _ := repo.GetRecentEntriesWithOptions(t.Context(), 7, false, "first_seen")
	if byFirstSeen[0].Title != "Entry C" {
		t.Errorf("Sort by first_seen: first entry = %s, want Entry C", byFirstSeen[0].Title)
	}
	if byFirstSeen[1].Title != "Entry B" {
		t.Errorf("Sort by first_seen: second entry = %s, want Entry B", byFirstSeen[1].Title)
	}
	if byFirstSeen[2].Title != "Entry A" {
		t.Errorf("Sort by first_seen: third entry = %s, want Entry A", byFirstSeen[2].Title)
	}
}

func TestGetRecentEntriesFilterAndSortByFirstSeen(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	feedID, _ := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")
	baseTime := time.Now()

	entries := []struct {
		title     string
		published time.Time
		firstSeen time.Time
	}{
		{"Recent discovery", baseTime.AddDate(0, 0, -30), baseTime.AddDate(0, 0, -1)}, // Old content, just discovered
		{"Recent post", baseTime.AddDate(0, 0, -1), baseTime.AddDate(0, 0, -1)},       // Recent content, recently discovered
		{"Old discovery", baseTime.AddDate(0, 0, -2), baseTime.AddDate(0, 0, -10)},    // Should be filtered out
	}

	for i, e := range entries {
		if err := repo.UpsertEntry(t.Context(), &Entry{
			FeedID:    feedID,
			EntryID:   fmt.Sprintf("entry-%d", i),
			Title:     e.title,
			Published: e.published,
			FirstSeen: e.firstSeen,
		}); err != nil {
			t.Fatalf("UpsertEntry() error = %v", err)
		}
	}

	// Filter by first_seen AND sort by first_seen
	results, _ := repo.GetRecentEntriesWithOptions(t.Context(), 7, true, "first_seen")

	// Should have 2 entries (first_seen within 7 days)
	if len(results) != 2 {
		t.Fatalf("got %d entries, want 2", len(results))
	}

	// Should be sorted by first_seen DESC (both on day -1, order may vary by insert time)
	// Just verify the old discovery is not included
	for _, e := range results {
		if e.Title == "Old discovery" {
			t.Errorf("Old discovery should be filtered out (first_seen too old)")
		}
	}
}

func TestGetEntryCountForFeed(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	// Add a feed
	feedID, err := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")
	if err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	// Add another feed
	feed2ID, err := repo.AddFeed(t.Context(), "https://example.com/feed2", "Test Feed 2")
	if err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	// Initially should be 0 entries
	count, err := repo.GetEntryCountForFeed(t.Context(), feedID)
	if err != nil {
		t.Fatalf("GetEntryCountForFeed() error = %v", err)
	}
	if count != 0 {
		t.Errorf("GetEntryCountForFeed() = %d, want 0", count)
	}

	// Add 3 entries to first feed
	now := time.Now()
	for i := range 3 {
		entry := &Entry{
			FeedID:      feedID,
			EntryID:     fmt.Sprintf("entry%d", i),
			Title:       fmt.Sprintf("Entry %d", i),
			Link:        fmt.Sprintf("https://example.com/entry%d", i),
			Published:   now,
			Updated:     now,
			FirstSeen:   now,
			Content:     "Test content",
			ContentType: "html",
		}
		if err := repo.UpsertEntry(t.Context(), entry); err != nil {
			t.Fatalf("UpsertEntry() error = %v", err)
		}
	}

	// Add 2 entries to second feed
	for i := range 2 {
		entry := &Entry{
			FeedID:      feed2ID,
			EntryID:     fmt.Sprintf("entry%d", i),
			Title:       fmt.Sprintf("Entry %d", i),
			Link:        fmt.Sprintf("https://example.com/entry%d", i),
			Published:   now,
			Updated:     now,
			FirstSeen:   now,
			Content:     "Test content",
			ContentType: "html",
		}
		if err := repo.UpsertEntry(t.Context(), entry); err != nil {
			t.Fatalf("UpsertEntry() error = %v", err)
		}
	}

	// Should have 3 entries for first feed
	count, err = repo.GetEntryCountForFeed(t.Context(), feedID)
	if err != nil {
		t.Fatalf("GetEntryCountForFeed() error = %v", err)
	}
	if count != 3 {
		t.Errorf("GetEntryCountForFeed() = %d, want 3", count)
	}

	// Should have 2 entries for second feed
	count, err = repo.GetEntryCountForFeed(t.Context(), feed2ID)
	if err != nil {
		t.Fatalf("GetEntryCountForFeed() error = %v", err)
	}
	if count != 2 {
		t.Errorf("GetEntryCountForFeed() = %d, want 2", count)
	}

	// Non-existent feed should return 0
	count, err = repo.GetEntryCountForFeed(t.Context(), 999)
	if err != nil {
		t.Fatalf("GetEntryCountForFeed() error = %v", err)
	}
	if count != 0 {
		t.Errorf("GetEntryCountForFeed(999) = %d, want 0", count)
	}
}

func TestUpdateFeedURL(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	// Add a feed with ETag and Last-Modified
	feedID, err := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")
	if err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	// Set cache headers
	if err := repo.UpdateFeedCache(t.Context(), feedID, "\"abc123\"", "Mon, 01 Jan 2024 00:00:00 GMT", time.Now()); err != nil {
		t.Fatalf("UpdateFeedCache() error = %v", err)
	}

	// Verify cache headers are set
	feed, err := repo.GetFeedByURL(t.Context(), "https://example.com/feed")
	if err != nil {
		t.Fatalf("GetFeedByURL() error = %v", err)
	}
	if feed.ETag != "\"abc123\"" {
		t.Errorf("ETag = %q, want \"abc123\"", feed.ETag)
	}
	if feed.LastModified != "Mon, 01 Jan 2024 00:00:00 GMT" {
		t.Errorf("LastModified = %q, want \"Mon, 01 Jan 2024 00:00:00 GMT\"", feed.LastModified)
	}

	// Update URL (simulating 301 redirect)
	newURL := "https://example.com/new-feed"
	if err := repo.UpdateFeedURL(t.Context(), feedID, newURL); err != nil {
		t.Fatalf("UpdateFeedURL() error = %v", err)
	}

	// Old URL should not exist
	_, err = repo.GetFeedByURL(t.Context(), "https://example.com/feed")
	if err != ErrFeedNotFound {
		t.Errorf("GetFeedByURL(old URL) error = %v, want ErrFeedNotFound", err)
	}

	// New URL should exist
	updatedFeed, err := repo.GetFeedByURL(t.Context(), newURL)
	if err != nil {
		t.Fatalf("GetFeedByURL(new URL) error = %v", err)
	}
	if updatedFeed.URL != newURL {
		t.Errorf("Feed URL = %q, want %q", updatedFeed.URL, newURL)
	}

	// Cache headers should be cleared (as per spec)
	if updatedFeed.ETag != "" {
		t.Errorf("ETag = %q, want empty (should be cleared after URL update)", updatedFeed.ETag)
	}
	if updatedFeed.LastModified != "" {
		t.Errorf("LastModified = %q, want empty (should be cleared after URL update)", updatedFeed.LastModified)
	}

	// Feed ID should remain the same
	if updatedFeed.ID != feedID {
		t.Errorf("Feed ID = %d, want %d", updatedFeed.ID, feedID)
	}
}

func TestRemoveFeedCascadeDelete(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	// Add a feed
	feedID, err := repo.AddFeed(t.Context(), "https://example.com/feed", "Test Feed")
	if err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	// Add entries
	now := time.Now()
	for i := range 5 {
		entry := &Entry{
			FeedID:      feedID,
			EntryID:     fmt.Sprintf("entry%d", i),
			Title:       fmt.Sprintf("Entry %d", i),
			Link:        fmt.Sprintf("https://example.com/entry%d", i),
			Published:   now,
			Updated:     now,
			FirstSeen:   now,
			Content:     "Test content",
			ContentType: "html",
		}
		if err := repo.UpsertEntry(t.Context(), entry); err != nil {
			t.Fatalf("UpsertEntry() error = %v", err)
		}
	}

	// Verify entries exist
	count, err := repo.GetEntryCountForFeed(t.Context(), feedID)
	if err != nil {
		t.Fatalf("GetEntryCountForFeed() error = %v", err)
	}
	if count != 5 {
		t.Errorf("GetEntryCountForFeed() = %d, want 5", count)
	}

	// Remove feed
	if err := repo.RemoveFeed(t.Context(), feedID); err != nil {
		t.Fatalf("RemoveFeed() error = %v", err)
	}

	// Feed should be gone
	_, err = repo.GetFeedByURL(t.Context(), "https://example.com/feed")
	if err != ErrFeedNotFound {
		t.Errorf("GetFeedByURL() after delete: got error %v, want ErrFeedNotFound", err)
	}

	// Entries should be cascade deleted
	count, err = repo.GetEntryCountForFeed(t.Context(), feedID)
	if err != nil {
		t.Fatalf("GetEntryCountForFeed() after feed delete: error = %v", err)
	}
	if count != 0 {
		t.Errorf("GetEntryCountForFeed() after feed delete = %d, want 0 (cascade delete failed)", count)
	}
}

// Boolean flag tests for branch coverage

func TestGetFeeds_ActiveOnly(t *testing.T) {
	t.Parallel()
	// Test branch where activeOnly is true (line 394)
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() { _ = repo.Close() }()

	// Add some feeds with different active statuses
	id1, err := repo.AddFeed(t.Context(), "http://example.com/feed1", "Active Feed 1")
	if err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	id2, err := repo.AddFeed(t.Context(), "http://example.com/feed2", "Active Feed 2")
	if err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	id3, err := repo.AddFeed(t.Context(), "http://example.com/feed3", "Inactive Feed")
	if err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	// Mark feed 3 as inactive
	_, err = repo.db.Exec("UPDATE feeds SET active = 0 WHERE id = ?", id3)
	if err != nil {
		t.Fatalf("Failed to mark feed as inactive: %v", err)
	}

	// Test with activeOnly = false (should get all 3 feeds)
	allFeeds, err := repo.GetFeeds(t.Context(), false)
	if err != nil {
		t.Fatalf("GetFeeds(false) error = %v", err)
	}
	if len(allFeeds) != 3 {
		t.Errorf("GetFeeds(false) returned %d feeds, want 3", len(allFeeds))
	}

	// Test with activeOnly = true (should get only 2 active feeds)
	activeFeeds, err := repo.GetFeeds(t.Context(), true)
	if err != nil {
		t.Fatalf("GetFeeds(true) error = %v", err)
	}
	if len(activeFeeds) != 2 {
		t.Errorf("GetFeeds(true) returned %d feeds, want 2", len(activeFeeds))
	}

	// Verify we got the right feeds (id1 and id2, not id3)
	foundIds := make(map[int64]bool)
	for _, feed := range activeFeeds {
		foundIds[feed.ID] = true
	}

	if !foundIds[id1] {
		t.Error("Active feed 1 not returned by GetFeeds(true)")
	}
	if !foundIds[id2] {
		t.Error("Active feed 2 not returned by GetFeeds(true)")
	}
	if foundIds[id3] {
		t.Error("Inactive feed 3 should not be returned by GetFeeds(true)")
	}
}

func TestPruneOldEntries_InvalidDays(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	// Zero days should be rejected (would delete all entries)
	_, err := repo.PruneOldEntries(context.Background(), 0)
	if err == nil {
		t.Error("PruneOldEntries(0) should return error")
	}
	if err != nil && !strings.Contains(err.Error(), "days must be >= 1") {
		t.Errorf("PruneOldEntries(0) error = %v, want 'days must be >= 1'", err)
	}

	// Negative days should be rejected
	_, err = repo.PruneOldEntries(context.Background(), -1)
	if err == nil {
		t.Error("PruneOldEntries(-1) should return error")
	}
	if err != nil && !strings.Contains(err.Error(), "days must be >= 1") {
		t.Errorf("PruneOldEntries(-1) error = %v, want 'days must be >= 1'", err)
	}

	// Verify that valid days still works
	_, err = repo.PruneOldEntries(context.Background(), 1)
	if err != nil {
		t.Errorf("PruneOldEntries(1) should succeed, got error: %v", err)
	}
}

func TestGetSchemaVersion_EmptyTable(t *testing.T) {
	t.Parallel()
	// Test that getSchemaVersion returns 0 when the schema_version table
	// is empty, without relying on fragile error string matching.
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a database with schema_version table but no rows
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Open db: %v", err)
	}
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`
		CREATE TABLE schema_version (
			version INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Create table: %v", err)
	}

	repo := &Repository{db: db}
	defer func() { _ = repo.Close() }()

	version, err := repo.getSchemaVersion(context.Background())
	if err != nil {
		t.Fatalf("getSchemaVersion() error = %v", err)
	}
	if version != 0 {
		t.Errorf("getSchemaVersion() = %d, want 0 for empty table", version)
	}
}

func TestMaxOpenConnsForPragmaConsistency(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	// Verify that MaxOpenConns is set to 1 to ensure PRAGMA settings
	// (like foreign_keys=ON) apply consistently across all operations.
	// Without this, new connections from the pool would not have
	// foreign_keys enabled.
	stats := repo.db.Stats()
	// MaxOpenConnections of 1 means only one connection in the pool
	if stats.MaxOpenConnections != 1 {
		t.Errorf("MaxOpenConnections = %d, want 1 (for consistent PRAGMA settings)", stats.MaxOpenConnections)
	}

	// Verify foreign keys are actually enforced by trying to insert
	// an entry with a non-existent feed_id
	_, err := repo.db.Exec(`
		INSERT INTO entries (feed_id, entry_id, title, published, updated, first_seen)
		VALUES (99999, 'test', 'Test', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z')
	`)
	if err == nil {
		t.Error("Foreign key constraint should reject entry with non-existent feed_id")
	}
}

func TestMigrationTransactional(t *testing.T) {
	t.Parallel()
	// Verify that schema version and schema changes are in the same transaction
	// by checking that both exist after successful migration
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() { _ = repo.Close() }()

	// Verify schema version was set
	var version int
	err = repo.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		t.Fatalf("Query schema version: %v", err)
	}

	if version != currentSchemaVersion {
		t.Errorf("Schema version = %d, want %d", version, currentSchemaVersion)
	}

	// Verify first_seen column exists (added in v2 migration)
	var hasFirstSeen bool
	err = repo.db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('entries')
		WHERE name = 'first_seen'
	`).Scan(&hasFirstSeen)
	if err != nil {
		t.Fatalf("Check first_seen column: %v", err)
	}

	if !hasFirstSeen {
		t.Error("first_seen column should exist after migration")
	}
}

func TestMigrationFromV1(t *testing.T) {
	t.Parallel()
	// Create a v1-style database (without first_seen column) and verify migration
	// runs within the transaction (uses tx, not r.db)
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Manually create a v1-like database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Open db: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE schema_version (
			version INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO schema_version (version) VALUES (1);

		CREATE TABLE feeds (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT NOT NULL UNIQUE,
			title TEXT,
			link TEXT,
			updated TEXT,
			last_fetched TEXT,
			etag TEXT,
			last_modified TEXT,
			fetch_error TEXT,
			fetch_error_count INTEGER DEFAULT 0,
			next_fetch TEXT,
			active INTEGER DEFAULT 1,
			fetch_interval INTEGER DEFAULT 3600
		);

		CREATE TABLE entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			feed_id INTEGER NOT NULL,
			entry_id TEXT NOT NULL,
			title TEXT,
			link TEXT,
			author TEXT,
			published TEXT,
			updated TEXT,
			content TEXT,
			content_type TEXT DEFAULT 'html',
			summary TEXT,
			FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
			UNIQUE(feed_id, entry_id)
		);

		CREATE INDEX idx_entries_published ON entries(published DESC);
		CREATE INDEX idx_entries_updated ON entries(updated DESC);
		CREATE INDEX idx_entries_feed_id ON entries(feed_id);
	`)
	if err != nil {
		t.Fatalf("Create v1 schema: %v", err)
	}

	// Add a test entry so migration backfill has data to work on
	_, err = db.Exec(`
		INSERT INTO feeds (url, title) VALUES ('https://example.com/feed', 'Test');
		INSERT INTO entries (feed_id, entry_id, title, published, updated)
		VALUES (1, 'e1', 'Test Entry', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z');
	`)
	if err != nil {
		t.Fatalf("Insert test data: %v", err)
	}
	_ = db.Close()

	// Now open with Repository which should run the v1->v2 migration
	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() on v1 database error = %v", err)
	}
	defer func() { _ = repo.Close() }()

	// Verify migration completed: first_seen column should exist
	var hasFirstSeen bool
	err = repo.db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('entries')
		WHERE name = 'first_seen'
	`).Scan(&hasFirstSeen)
	if err != nil {
		t.Fatalf("Check first_seen column: %v", err)
	}
	if !hasFirstSeen {
		t.Error("first_seen column should exist after v1->v2 migration")
	}

	// Verify schema version was updated to 2
	var version int
	err = repo.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		t.Fatalf("Query schema version: %v", err)
	}
	if version != 2 {
		t.Errorf("Schema version = %d, want 2", version)
	}

	// Verify backfill happened (first_seen should be populated)
	var firstSeen string
	err = repo.db.QueryRow("SELECT first_seen FROM entries WHERE entry_id = 'e1'").Scan(&firstSeen)
	if err != nil {
		t.Fatalf("Query first_seen: %v", err)
	}
	if firstSeen == "" {
		t.Error("first_seen should be backfilled after migration")
	}
}

// =============================================================================
// Tests for L7: Security-Hardening SQLite PRAGMAs
// =============================================================================

func TestSecurityPragmas(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	// Verify trusted_schema = OFF
	var trustedSchema int
	err := repo.db.QueryRow("PRAGMA trusted_schema").Scan(&trustedSchema)
	if err != nil {
		t.Fatalf("PRAGMA trusted_schema query error: %v", err)
	}
	if trustedSchema != 0 {
		t.Errorf("trusted_schema = %d, want 0 (OFF)", trustedSchema)
	}

	// Verify cell_size_check = ON
	var cellSizeCheck int
	err = repo.db.QueryRow("PRAGMA cell_size_check").Scan(&cellSizeCheck)
	if err != nil {
		t.Fatalf("PRAGMA cell_size_check query error: %v", err)
	}
	if cellSizeCheck != 1 {
		t.Errorf("cell_size_check = %d, want 1 (ON)", cellSizeCheck)
	}

	// Verify busy_timeout = 5000
	var busyTimeout int
	err = repo.db.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout)
	if err != nil {
		t.Fatalf("PRAGMA busy_timeout query error: %v", err)
	}
	if busyTimeout != 5000 {
		t.Errorf("busy_timeout = %d, want 5000", busyTimeout)
	}
}

// =============================================================================
// Tests for L10: Database File Permissions
// =============================================================================

func TestDatabaseFilePermissions(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "permissions_test.db")

	// Database file should not exist yet
	_, err := os.Stat(dbPath)
	if !os.IsNotExist(err) {
		t.Fatal("Database file should not exist before New()")
	}

	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() { _ = repo.Close() }()

	// Verify file permissions are 0600 (owner read/write only)
	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("Database file permissions = %o, want 0600", perm)
	}
}

func TestDatabaseFilePermissions_ExistingFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "existing.db")

	// Create a database file first
	repo1, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_ = repo1.Close()

	// Change permissions to something different
	if err := os.Chmod(dbPath, 0644); err != nil {
		t.Fatalf("Chmod() error = %v", err)
	}

	// Reopen - should NOT change permissions of existing file
	repo2, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() reopen error = %v", err)
	}
	defer func() { _ = repo2.Close() }()

	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0644 {
		t.Errorf("Existing file permissions changed to %o, should remain 0644", perm)
	}
}

// =============================================================================
// Tests for L11: Context/Timeout on Schema Initialization
// =============================================================================

func TestSchemaInitializationWithContext(t *testing.T) {
	t.Parallel()
	// Verify that schema initialization completes successfully
	// (the context/timeout is internal, we just verify it works)
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "context_test.db")

	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() { _ = repo.Close() }()

	// Verify schema was created successfully
	var tableCount int
	err = repo.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('feeds', 'entries', 'schema_version')").Scan(&tableCount)
	if err != nil {
		t.Fatalf("Query error: %v", err)
	}
	if tableCount != 3 {
		t.Errorf("Expected 3 tables (feeds, entries, schema_version), got %d", tableCount)
	}
}

// =============================================================================
// Tests for M9: Input Validation in Repository Layer
// =============================================================================

func TestAddFeed_Validation(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	tests := []struct {
		name    string
		url     string
		title   string
		wantErr string
	}{
		{
			name:    "empty URL",
			url:     "",
			title:   "Test Feed",
			wantErr: "empty",
		},
		{
			name:    "whitespace only URL",
			url:     "   ",
			title:   "Test Feed",
			wantErr: "empty",
		},
		{
			name:    "ftp scheme",
			url:     "ftp://example.com/feed",
			title:   "Test Feed",
			wantErr: "scheme",
		},
		{
			name:    "javascript scheme",
			url:     "javascript:alert(1)",
			title:   "Test Feed",
			wantErr: "scheme",
		},
		{
			name:    "no scheme",
			url:     "example.com/feed",
			title:   "Test Feed",
			wantErr: "scheme",
		},
		{
			name:    "valid http URL",
			url:     "http://example.com/feed",
			title:   "Test Feed",
			wantErr: "",
		},
		{
			name:    "valid https URL",
			url:     "https://example.com/feed",
			title:   "Test Feed",
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := repo.AddFeed(context.Background(), tt.url, tt.title)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("AddFeed() unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("AddFeed() expected error containing %q, got nil", tt.wantErr)
				} else if !strings.Contains(strings.ToLower(err.Error()), tt.wantErr) {
					t.Errorf("AddFeed() error = %v, want error containing %q", err, tt.wantErr)
				}
			}
		})
	}
}

func TestUpdateFeedURL_Validation(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	feedID, err := repo.AddFeed(context.Background(), "https://example.com/feed", "Test Feed")
	if err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	tests := []struct {
		name    string
		newURL  string
		wantErr string
	}{
		{
			name:    "empty URL",
			newURL:  "",
			wantErr: "empty",
		},
		{
			name:    "ftp scheme",
			newURL:  "ftp://example.com/feed2",
			wantErr: "scheme",
		},
		{
			name:    "data scheme",
			newURL:  "data:text/html,<h1>test</h1>",
			wantErr: "scheme",
		},
		{
			name:    "valid https URL",
			newURL:  "https://example.com/new-feed",
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateFeedURL(context.Background(), feedID, tt.newURL)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("UpdateFeedURL() unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("UpdateFeedURL() expected error containing %q, got nil", tt.wantErr)
				} else if !strings.Contains(strings.ToLower(err.Error()), tt.wantErr) {
					t.Errorf("UpdateFeedURL() error = %v, want error containing %q", err, tt.wantErr)
				}
			}
		})
	}
}

func TestUpsertEntry_Validation(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	feedID, err := repo.AddFeed(context.Background(), "https://example.com/feed", "Test Feed")
	if err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	now := time.Now()

	tests := []struct {
		name    string
		entry   *Entry
		wantErr string
	}{
		{
			name: "empty entry_id",
			entry: &Entry{
				FeedID:    feedID,
				EntryID:   "",
				Title:     "Test",
				Published: now,
				Updated:   now,
				FirstSeen: now,
			},
			wantErr: "entry_id",
		},
		{
			name: "whitespace only entry_id",
			entry: &Entry{
				FeedID:    feedID,
				EntryID:   "   ",
				Title:     "Test",
				Published: now,
				Updated:   now,
				FirstSeen: now,
			},
			wantErr: "entry_id",
		},
		{
			name: "zero feed_id",
			entry: &Entry{
				FeedID:    0,
				EntryID:   "entry-1",
				Title:     "Test",
				Published: now,
				Updated:   now,
				FirstSeen: now,
			},
			wantErr: "feed_id",
		},
		{
			name: "valid entry",
			entry: &Entry{
				FeedID:    feedID,
				EntryID:   "entry-valid",
				Title:     "Test",
				Published: now,
				Updated:   now,
				FirstSeen: now,
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpsertEntry(context.Background(), tt.entry)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("UpsertEntry() unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("UpsertEntry() expected error containing %q, got nil", tt.wantErr)
				} else if !strings.Contains(strings.ToLower(err.Error()), tt.wantErr) {
					t.Errorf("UpsertEntry() error = %v, want error containing %q", err, tt.wantErr)
				}
			}
		})
	}
}

// =============================================================================
// Tests for M8: Batch Transactions for Entry Operations
// =============================================================================

func TestUpsertEntriesBatch_EmptySlice(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	count, err := repo.UpsertEntriesBatch(context.Background(), nil)
	if err != nil {
		t.Fatalf("UpsertEntriesBatch(nil) error = %v", err)
	}
	if count != 0 {
		t.Errorf("UpsertEntriesBatch(nil) count = %d, want 0", count)
	}

	count, err = repo.UpsertEntriesBatch(context.Background(), []*Entry{})
	if err != nil {
		t.Fatalf("UpsertEntriesBatch([]) error = %v", err)
	}
	if count != 0 {
		t.Errorf("UpsertEntriesBatch([]) count = %d, want 0", count)
	}
}

func TestUpsertEntriesBatch_SingleEntry(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	feedID, err := repo.AddFeed(context.Background(), "https://example.com/feed", "Test Feed")
	if err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	now := time.Now()
	entries := []*Entry{
		{
			FeedID:      feedID,
			EntryID:     "batch-single-1",
			Title:       "Single Batch Entry",
			Link:        "https://example.com/single",
			Published:   now,
			Updated:     now,
			Content:     "<p>Content</p>",
			ContentType: "html",
			FirstSeen:   now,
		},
	}

	count, err := repo.UpsertEntriesBatch(context.Background(), entries)
	if err != nil {
		t.Fatalf("UpsertEntriesBatch() error = %v", err)
	}
	if count != 1 {
		t.Errorf("UpsertEntriesBatch() count = %d, want 1", count)
	}

	// Verify entry exists
	total, err := repo.CountEntries(context.Background())
	if err != nil {
		t.Fatalf("CountEntries() error = %v", err)
	}
	if total != 1 {
		t.Errorf("CountEntries() = %d, want 1", total)
	}
}

func TestUpsertEntriesBatch_MultipleEntries(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	feedID, err := repo.AddFeed(context.Background(), "https://example.com/feed", "Test Feed")
	if err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	now := time.Now()
	entries := make([]*Entry, 10)
	for i := 0; i < 10; i++ {
		entries[i] = &Entry{
			FeedID:      feedID,
			EntryID:     fmt.Sprintf("batch-multi-%d", i),
			Title:       fmt.Sprintf("Batch Entry %d", i),
			Link:        fmt.Sprintf("https://example.com/batch/%d", i),
			Published:   now.Add(time.Duration(-i) * time.Hour),
			Updated:     now,
			Content:     fmt.Sprintf("<p>Content %d</p>", i),
			ContentType: "html",
			FirstSeen:   now,
		}
	}

	count, err := repo.UpsertEntriesBatch(context.Background(), entries)
	if err != nil {
		t.Fatalf("UpsertEntriesBatch() error = %v", err)
	}
	if count != 10 {
		t.Errorf("UpsertEntriesBatch() count = %d, want 10", count)
	}

	// Verify all entries exist
	total, err := repo.CountEntries(context.Background())
	if err != nil {
		t.Fatalf("CountEntries() error = %v", err)
	}
	if total != 10 {
		t.Errorf("CountEntries() = %d, want 10", total)
	}
}

func TestUpsertEntriesBatch_Atomicity(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	feedID, err := repo.AddFeed(context.Background(), "https://example.com/feed", "Test Feed")
	if err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	now := time.Now()

	// First, insert some valid entries
	validEntries := []*Entry{
		{
			FeedID:    feedID,
			EntryID:   "existing-1",
			Title:     "Existing Entry",
			Published: now,
			Updated:   now,
			FirstSeen: now,
		},
	}
	_, err = repo.UpsertEntriesBatch(context.Background(), validEntries)
	if err != nil {
		t.Fatalf("UpsertEntriesBatch() setup error = %v", err)
	}

	// Now try a batch with a validation error (empty entry_id should fail)
	badBatch := []*Entry{
		{
			FeedID:    feedID,
			EntryID:   "good-entry",
			Title:     "Good Entry",
			Published: now,
			Updated:   now,
			FirstSeen: now,
		},
		{
			FeedID:    feedID,
			EntryID:   "", // Invalid - empty entry_id
			Title:     "Bad Entry",
			Published: now,
			Updated:   now,
			FirstSeen: now,
		},
	}

	_, err = repo.UpsertEntriesBatch(context.Background(), badBatch)
	if err == nil {
		t.Fatal("UpsertEntriesBatch() should fail with invalid entry in batch")
	}

	// Verify none of the bad batch entries were persisted (rollback)
	total, err := repo.CountEntries(context.Background())
	if err != nil {
		t.Fatalf("CountEntries() error = %v", err)
	}
	if total != 1 {
		t.Errorf("CountEntries() = %d, want 1 (only the pre-existing entry)", total)
	}
}

func TestUpsertEntriesBatch_UpdatesExisting(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	feedID, err := repo.AddFeed(context.Background(), "https://example.com/feed", "Test Feed")
	if err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	now := time.Now()

	// Insert initial entries
	initial := []*Entry{
		{
			FeedID:    feedID,
			EntryID:   "entry-1",
			Title:     "Original Title 1",
			Published: now,
			Updated:   now,
			FirstSeen: now,
		},
		{
			FeedID:    feedID,
			EntryID:   "entry-2",
			Title:     "Original Title 2",
			Published: now,
			Updated:   now,
			FirstSeen: now,
		},
	}

	count, err := repo.UpsertEntriesBatch(context.Background(), initial)
	if err != nil {
		t.Fatalf("UpsertEntriesBatch() initial error = %v", err)
	}
	if count != 2 {
		t.Errorf("Initial count = %d, want 2", count)
	}

	// Update existing entries via batch
	updated := []*Entry{
		{
			FeedID:    feedID,
			EntryID:   "entry-1",
			Title:     "Updated Title 1",
			Published: now,
			Updated:   now.Add(time.Hour),
			FirstSeen: now,
		},
		{
			FeedID:    feedID,
			EntryID:   "entry-2",
			Title:     "Updated Title 2",
			Published: now,
			Updated:   now.Add(time.Hour),
			FirstSeen: now,
		},
		{
			FeedID:    feedID,
			EntryID:   "entry-3",
			Title:     "New Entry 3",
			Published: now,
			Updated:   now,
			FirstSeen: now,
		},
	}

	count, err = repo.UpsertEntriesBatch(context.Background(), updated)
	if err != nil {
		t.Fatalf("UpsertEntriesBatch() update error = %v", err)
	}
	if count != 3 {
		t.Errorf("Update count = %d, want 3", count)
	}

	// Verify total count is 3 (2 updated + 1 new)
	total, err := repo.CountEntries(context.Background())
	if err != nil {
		t.Fatalf("CountEntries() error = %v", err)
	}
	if total != 3 {
		t.Errorf("CountEntries() = %d, want 3", total)
	}

	// Verify updates took effect
	var title string
	err = repo.db.QueryRow("SELECT title FROM entries WHERE entry_id = 'entry-1'").Scan(&title)
	if err != nil {
		t.Fatalf("Query error: %v", err)
	}
	if title != "Updated Title 1" {
		t.Errorf("Entry 1 title = %q, want %q", title, "Updated Title 1")
	}
}

func TestUpsertEntriesBatch_ContextCancellation(t *testing.T) {
	t.Parallel()
	repo, _ := setupTestDB(t)
	defer func() { _ = repo.Close() }()

	feedID, err := repo.AddFeed(context.Background(), "https://example.com/feed", "Test Feed")
	if err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	now := time.Now()
	entries := make([]*Entry, 5)
	for i := 0; i < 5; i++ {
		entries[i] = &Entry{
			FeedID:    feedID,
			EntryID:   fmt.Sprintf("ctx-entry-%d", i),
			Title:     fmt.Sprintf("Entry %d", i),
			Published: now,
			Updated:   now,
			FirstSeen: now,
		}
	}

	// Use an already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = repo.UpsertEntriesBatch(ctx, entries)
	if err == nil {
		t.Error("UpsertEntriesBatch() should fail with cancelled context")
	}
}
