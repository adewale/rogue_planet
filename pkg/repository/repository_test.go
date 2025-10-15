package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	repo, _ := setupTestDB(t)
	defer repo.Close()

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
	repo, _ := setupTestDB(t)
	defer repo.Close()

	id, err := repo.AddFeed("https://example.com/feed", "Test Feed")
	if err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	if id == 0 {
		t.Error("Expected non-zero ID")
	}

	// Verify feed was added
	feed, err := repo.GetFeedByURL("https://example.com/feed")
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
	repo, _ := setupTestDB(t)
	defer repo.Close()

	_, err := repo.AddFeed("https://example.com/feed", "Test Feed")
	if err != nil {
		t.Fatalf("AddFeed() error = %v", err)
	}

	// Try to add duplicate
	_, err = repo.AddFeed("https://example.com/feed", "Test Feed 2")
	if err == nil {
		t.Error("Expected error for duplicate feed, got nil")
	}
}

func TestUpdateFeed(t *testing.T) {
	repo, _ := setupTestDB(t)
	defer repo.Close()

	id, _ := repo.AddFeed("https://example.com/feed", "Old Title")

	updated := time.Now()
	err := repo.UpdateFeed(id, "New Title", "https://example.com", updated)
	if err != nil {
		t.Fatalf("UpdateFeed() error = %v", err)
	}

	feed, _ := repo.GetFeedByURL("https://example.com/feed")

	if feed.Title != "New Title" {
		t.Errorf("Title = %q, want %q", feed.Title, "New Title")
	}

	if feed.Link != "https://example.com" {
		t.Errorf("Link = %q, want %q", feed.Link, "https://example.com")
	}
}

func TestUpdateFeedCache(t *testing.T) {
	repo, _ := setupTestDB(t)
	defer repo.Close()

	id, _ := repo.AddFeed("https://example.com/feed", "Test Feed")

	fetchTime := time.Now()
	err := repo.UpdateFeedCache(id, `"etag123"`, "Mon, 02 Jan 2006 15:04:05 GMT", fetchTime)
	if err != nil {
		t.Fatalf("UpdateFeedCache() error = %v", err)
	}

	feed, _ := repo.GetFeedByURL("https://example.com/feed")

	if feed.ETag != `"etag123"` {
		t.Errorf("ETag = %q, want %q", feed.ETag, `"etag123"`)
	}

	if feed.LastModified != "Mon, 02 Jan 2006 15:04:05 GMT" {
		t.Errorf("LastModified = %q, want %q", feed.LastModified, "Mon, 02 Jan 2006 15:04:05 GMT")
	}
}

func TestGetFeeds(t *testing.T) {
	repo, _ := setupTestDB(t)
	defer repo.Close()

	repo.AddFeed("https://example.com/feed1", "Feed 1")
	repo.AddFeed("https://example.com/feed2", "Feed 2")

	feeds, err := repo.GetFeeds(false)
	if err != nil {
		t.Fatalf("GetFeeds() error = %v", err)
	}

	if len(feeds) != 2 {
		t.Errorf("len(feeds) = %d, want 2", len(feeds))
	}
}

func TestGetFeedByURL(t *testing.T) {
	repo, _ := setupTestDB(t)
	defer repo.Close()

	repo.AddFeed("https://example.com/feed", "Test Feed")

	feed, err := repo.GetFeedByURL("https://example.com/feed")
	if err != nil {
		t.Fatalf("GetFeedByURL() error = %v", err)
	}

	if feed.URL != "https://example.com/feed" {
		t.Errorf("URL = %q, want %q", feed.URL, "https://example.com/feed")
	}

	// Test non-existent feed
	_, err = repo.GetFeedByURL("https://example.com/nonexistent")
	if err != ErrFeedNotFound {
		t.Errorf("Expected ErrFeedNotFound, got %v", err)
	}
}

func TestRemoveFeed(t *testing.T) {
	repo, _ := setupTestDB(t)
	defer repo.Close()

	id, _ := repo.AddFeed("https://example.com/feed", "Test Feed")

	err := repo.RemoveFeed(id)
	if err != nil {
		t.Fatalf("RemoveFeed() error = %v", err)
	}

	// Verify feed was removed
	_, err = repo.GetFeedByURL("https://example.com/feed")
	if err != ErrFeedNotFound {
		t.Error("Feed should have been removed")
	}
}

func TestUpsertEntry(t *testing.T) {
	repo, _ := setupTestDB(t)
	defer repo.Close()

	feedID, _ := repo.AddFeed("https://example.com/feed", "Test Feed")

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

	err := repo.UpsertEntry(entry)
	if err != nil {
		t.Fatalf("UpsertEntry() error = %v", err)
	}

	// Test update
	entry.Title = "Updated Title"
	err = repo.UpsertEntry(entry)
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

func TestGetRecentEntries(t *testing.T) {
	repo, _ := setupTestDB(t)
	defer repo.Close()

	feedID, _ := repo.AddFeed("https://example.com/feed", "Test Feed")

	// Add recent entry
	recentEntry := &Entry{
		FeedID:    feedID,
		EntryID:   "recent",
		Title:     "Recent Entry",
		Published: time.Now(),
		Updated:   time.Now(),
		FirstSeen: time.Now(),
	}
	repo.UpsertEntry(recentEntry)

	// Add old entry
	oldEntry := &Entry{
		FeedID:    feedID,
		EntryID:   "old",
		Title:     "Old Entry",
		Published: time.Now().AddDate(0, 0, -10),
		Updated:   time.Now().AddDate(0, 0, -10),
		FirstSeen: time.Now().AddDate(0, 0, -10),
	}
	repo.UpsertEntry(oldEntry)

	// Get recent entries (last 7 days)
	entries, err := repo.GetRecentEntries(7)
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
	repo, _ := setupTestDB(t)
	defer repo.Close()

	feedID, _ := repo.AddFeed("https://example.com/feed", "Test Feed")

	// Add recent entry
	recentEntry := &Entry{
		FeedID:    feedID,
		EntryID:   "recent",
		Title:     "Recent Entry",
		Published: time.Now(),
		Updated:   time.Now(),
		FirstSeen: time.Now(),
	}
	repo.UpsertEntry(recentEntry)

	// Add old entry
	oldEntry := &Entry{
		FeedID:    feedID,
		EntryID:   "old",
		Title:     "Old Entry",
		Published: time.Now().AddDate(0, 0, -100),
		Updated:   time.Now().AddDate(0, 0, -100),
		FirstSeen: time.Now().AddDate(0, 0, -100),
	}
	repo.UpsertEntry(oldEntry)

	// Prune entries older than 90 days
	deleted, err := repo.PruneOldEntries(90)
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
	repo, _ := setupTestDB(t)
	defer repo.Close()

	feedID, _ := repo.AddFeed("https://example.com/feed", "Test Feed")

	// Add entry
	entry := &Entry{
		FeedID:    feedID,
		EntryID:   "entry-1",
		Title:     "Test Entry",
		Published: time.Now(),
		Updated:   time.Now(),
		FirstSeen: time.Now(),
	}
	repo.UpsertEntry(entry)

	// Remove feed
	repo.RemoveFeed(feedID)

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
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "persist.db")

	// Create repository and add data
	repo1, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	repo1.AddFeed("https://example.com/feed", "Test Feed")
	repo1.Close()

	// Reopen database
	repo2, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to reopen repository: %v", err)
	}
	defer repo2.Close()

	// Verify data persisted
	feeds, _ := repo2.GetFeeds(false)
	if len(feeds) != 1 {
		t.Errorf("Data did not persist: len(feeds) = %d, want 1", len(feeds))
	}

	// Verify database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file should exist")
	}
}

func TestGetRecentEntriesFallback(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	// Add a feed
	feedID, err := repo.AddFeed("https://example.com/feed", "Test Feed")
	if err != nil {
		t.Fatalf("Failed to add feed: %v", err)
	}

	// Add entries with old publish dates (100 days ago)
	oldDate := time.Now().AddDate(0, 0, -100)
	for i := 0; i < 10; i++ {
		entry := &Entry{
			FeedID:    feedID,
			EntryID:   fmt.Sprintf("entry-%d", i),
			Title:     fmt.Sprintf("Old Entry %d", i),
			Link:      fmt.Sprintf("https://example.com/entry%d", i),
			Published: oldDate.Add(time.Duration(i) * time.Hour),
			Updated:   oldDate.Add(time.Duration(i) * time.Hour),
			FirstSeen: time.Now(),
		}
		if err := repo.UpsertEntry(entry); err != nil {
			t.Fatalf("Failed to upsert entry: %v", err)
		}
	}

	// Test 1: Requesting entries from last 7 days should return 0 (within window)
	// But with fallback, should return the old entries
	entries, err := repo.GetRecentEntries(7)
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
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	// Add a feed
	feedID, err := repo.AddFeed("https://example.com/feed", "Test Feed")
	if err != nil {
		t.Fatalf("Failed to add feed: %v", err)
	}

	// Add recent entries (within last 7 days)
	now := time.Now()
	for i := 0; i < 5; i++ {
		entry := &Entry{
			FeedID:    feedID,
			EntryID:   fmt.Sprintf("recent-%d", i),
			Title:     fmt.Sprintf("Recent Entry %d", i),
			Link:      fmt.Sprintf("https://example.com/recent%d", i),
			Published: now.Add(time.Duration(-i) * 24 * time.Hour), // Last 5 days
			Updated:   now.Add(time.Duration(-i) * 24 * time.Hour),
			FirstSeen: now,
		}
		if err := repo.UpsertEntry(entry); err != nil {
			t.Fatalf("Failed to upsert entry: %v", err)
		}
	}

	// Add old entries (100 days ago)
	oldDate := time.Now().AddDate(0, 0, -100)
	for i := 0; i < 5; i++ {
		entry := &Entry{
			FeedID:    feedID,
			EntryID:   fmt.Sprintf("old-%d", i),
			Title:     fmt.Sprintf("Old Entry %d", i),
			Link:      fmt.Sprintf("https://example.com/old%d", i),
			Published: oldDate.Add(time.Duration(i) * time.Hour),
			Updated:   oldDate.Add(time.Duration(i) * time.Hour),
			FirstSeen: now,
		}
		if err := repo.UpsertEntry(entry); err != nil {
			t.Fatalf("Failed to upsert entry: %v", err)
		}
	}

	// Request entries from last 7 days
	entries, err := repo.GetRecentEntries(7)
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
	repo, _ := setupTestDB(t)
	defer repo.Close()

	id, _ := repo.AddFeed("https://example.com/feed", "Test Feed")

	err := repo.UpdateFeedError(id, "Connection timeout")
	if err != nil {
		t.Fatalf("UpdateFeedError() error = %v", err)
	}

	feed, _ := repo.GetFeedByURL("https://example.com/feed")

	if feed.FetchError != "Connection timeout" {
		t.Errorf("FetchError = %q, want %q", feed.FetchError, "Connection timeout")
	}

	if feed.FetchErrorCount != 1 {
		t.Errorf("FetchErrorCount = %d, want 1", feed.FetchErrorCount)
	}

	// Call again to increment error count
	repo.UpdateFeedError(id, "Another error")
	feed, _ = repo.GetFeedByURL("https://example.com/feed")

	if feed.FetchErrorCount != 2 {
		t.Errorf("FetchErrorCount = %d, want 2", feed.FetchErrorCount)
	}
}

func TestCountEntries(t *testing.T) {
	repo, _ := setupTestDB(t)
	defer repo.Close()

	// Initially should be 0
	count, err := repo.CountEntries()
	if err != nil {
		t.Fatalf("CountEntries() error = %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	// Add some entries
	feedID, _ := repo.AddFeed("https://example.com/feed", "Test Feed")
	for i := 0; i < 5; i++ {
		entry := &Entry{
			FeedID:    feedID,
			EntryID:   fmt.Sprintf("entry-%d", i),
			Title:     fmt.Sprintf("Entry %d", i),
			Published: time.Now(),
			Updated:   time.Now(),
			FirstSeen: time.Now(),
		}
		repo.UpsertEntry(entry)
	}

	count, err = repo.CountEntries()
	if err != nil {
		t.Fatalf("CountEntries() error = %v", err)
	}
	if count != 5 {
		t.Errorf("count = %d, want 5", count)
	}
}

func TestCountRecentEntries(t *testing.T) {
	repo, _ := setupTestDB(t)
	defer repo.Close()

	feedID, _ := repo.AddFeed("https://example.com/feed", "Test Feed")

	// Add recent entries (last 3 days)
	now := time.Now()
	for i := 0; i < 3; i++ {
		entry := &Entry{
			FeedID:    feedID,
			EntryID:   fmt.Sprintf("recent-%d", i),
			Title:     fmt.Sprintf("Recent %d", i),
			Published: now.Add(time.Duration(-i) * 24 * time.Hour),
			Updated:   now,
			FirstSeen: now,
		}
		repo.UpsertEntry(entry)
	}

	// Add old entries (100 days ago)
	oldDate := now.AddDate(0, 0, -100)
	for i := 0; i < 2; i++ {
		entry := &Entry{
			FeedID:    feedID,
			EntryID:   fmt.Sprintf("old-%d", i),
			Title:     fmt.Sprintf("Old %d", i),
			Published: oldDate,
			Updated:   oldDate,
			FirstSeen: now,
		}
		repo.UpsertEntry(entry)
	}

	// Count recent entries (last 7 days)
	count, err := repo.CountRecentEntries(7)
	if err != nil {
		t.Fatalf("CountRecentEntries() error = %v", err)
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}

	// Count last 200 days (should include all)
	count, err = repo.CountRecentEntries(200)
	if err != nil {
		t.Fatalf("CountRecentEntries() error = %v", err)
	}
	if count != 5 {
		t.Errorf("count = %d, want 5", count)
	}
}

func TestNewErrors(t *testing.T) {
	// Test with invalid path
	_, err := New("/invalid/path/to/nonexistent/dir/test.db")
	if err == nil {
		t.Error("New() should fail with invalid path")
	}
}

func TestGetRecentEntriesFilterByFirstSeen(t *testing.T) {
	repo, _ := setupTestDB(t)
	defer repo.Close()

	// Add a feed
	feedID, _ := repo.AddFeed("https://example.com/feed", "Test Feed")

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
		err := repo.UpsertEntry(&Entry{
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
	publishedFiltered, err := repo.GetRecentEntriesWithOptions(7, false, "published")
	if err != nil {
		t.Fatalf("GetRecentEntriesWithOptions() error = %v", err)
	}
	if len(publishedFiltered) != 2 {
		t.Errorf("Filter by published: got %d entries, want 2", len(publishedFiltered))
	}

	// Test 2: Filter by first_seen
	// Should return entries 0 and 1 (first_seen within 7 days)
	firstSeenFiltered, err := repo.GetRecentEntriesWithOptions(7, true, "published")
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
	repo, _ := setupTestDB(t)
	defer repo.Close()

	feedID, _ := repo.AddFeed("https://example.com/feed", "Test Feed")
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
		err := repo.UpsertEntry(&Entry{
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
	byPublished, _ := repo.GetRecentEntriesWithOptions(7, false, "published")
	if byPublished[0].Title != "Entry A" {
		t.Errorf("Sort by published: first entry = %s, want Entry A", byPublished[0].Title)
	}

	// Sort by first_seen
	byFirstSeen, _ := repo.GetRecentEntriesWithOptions(7, false, "first_seen")
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
	repo, _ := setupTestDB(t)
	defer repo.Close()

	feedID, _ := repo.AddFeed("https://example.com/feed", "Test Feed")
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
		repo.UpsertEntry(&Entry{
			FeedID:    feedID,
			EntryID:   fmt.Sprintf("entry-%d", i),
			Title:     e.title,
			Published: e.published,
			FirstSeen: e.firstSeen,
		})
	}

	// Filter by first_seen AND sort by first_seen
	results, _ := repo.GetRecentEntriesWithOptions(7, true, "first_seen")

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
