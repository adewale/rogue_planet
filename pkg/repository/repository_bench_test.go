package repository

import (
	"testing"
	"time"
)

// BenchmarkAddFeed benchmarks feed insertion
func BenchmarkAddFeed(b *testing.B) {
	repo, err := New(b.TempDir() + "/bench.db")
	if err != nil {
		b.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		url := "https://example.com/feed/" + string(rune(i%1000))
		_, _ = repo.AddFeed(url, "Test Feed")
	}
}

// BenchmarkUpsertEntry benchmarks entry insertion/update
func BenchmarkUpsertEntry(b *testing.B) {
	repo, err := New(b.TempDir() + "/bench.db")
	if err != nil {
		b.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	// Add a test feed
	feedID, err := repo.AddFeed("https://example.com/feed", "Test Feed")
	if err != nil {
		b.Fatalf("Failed to add feed: %v", err)
	}

	now := time.Now()
	entry := &Entry{
		FeedID:      feedID,
		EntryID:     "entry-1",
		Title:       "Test Entry",
		Link:        "https://example.com/entry1",
		Author:      "Test Author",
		Published:   now,
		Updated:     now,
		Content:     "<p>Test content</p>",
		ContentType: "html",
		Summary:     "Test summary",
		FirstSeen:   now,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry.EntryID = "entry-" + string(rune(i%1000))
		_ = repo.UpsertEntry(entry)
	}
}

// BenchmarkGetRecentEntries benchmarks entry retrieval
func BenchmarkGetRecentEntries(b *testing.B) {
	repo, err := New(b.TempDir() + "/bench.db")
	if err != nil {
		b.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	// Add test data
	feedID, _ := repo.AddFeed("https://example.com/feed", "Test Feed")
	now := time.Now()
	for i := 0; i < 100; i++ {
		entry := &Entry{
			FeedID:      feedID,
			EntryID:     "entry-" + string(rune(i)),
			Title:       "Test Entry",
			Link:        "https://example.com/entry" + string(rune(i)),
			Author:      "Test Author",
			Published:   now.AddDate(0, 0, -i),
			Updated:     now.AddDate(0, 0, -i),
			Content:     "<p>Test content</p>",
			ContentType: "html",
			Summary:     "Test summary",
			FirstSeen:   now,
		}
		_ = repo.UpsertEntry(entry)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetRecentEntries(30)
	}
}

// BenchmarkGetFeeds benchmarks feed retrieval
func BenchmarkGetFeeds(b *testing.B) {
	repo, err := New(b.TempDir() + "/bench.db")
	if err != nil {
		b.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	// Add test feeds
	for i := 0; i < 50; i++ {
		_, _ = repo.AddFeed("https://example.com/feed/"+string(rune(i)), "Test Feed "+string(rune(i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetFeeds(true)
	}
}

// BenchmarkUpdateFeedCache benchmarks cache header updates
func BenchmarkUpdateFeedCache(b *testing.B) {
	repo, err := New(b.TempDir() + "/bench.db")
	if err != nil {
		b.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	feedID, _ := repo.AddFeed("https://example.com/feed", "Test Feed")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		etag := "etag-" + string(rune(i%100))
		lastModified := "Mon, 01 Jan 2024 00:00:00 GMT"
		_ = repo.UpdateFeedCache(feedID, etag, lastModified, time.Now())
	}
}
