//go:build network
// +build network

package crawler

import (
	"context"
	"testing"
	"time"

	"github.com/adewale/rogue_planet/pkg/normalizer"
	"github.com/adewale/rogue_planet/pkg/repository"
)

// These tests require network access and fetch from real URLs.
// Run with: go test -tags=network ./pkg/crawler

// TestLiveFetchRealWorldFeeds tests fetching and parsing from actual live URLs
func TestLiveFetchRealWorldFeeds(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("Skipping live network test in short mode")
	}

	tests := []struct {
		name          string
		url           string
		expectedTitle string
		minEntries    int
	}{
		{
			name:          "Daring Fireball",
			url:           "https://daringfireball.net/feeds/main",
			expectedTitle: "Daring Fireball",
			minEntries:    10,
		},
		{
			name:          "Asymco",
			url:           "https://asymco.com/feed/",
			expectedTitle: "Asymco",
			minEntries:    5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create crawler
			c := New()

			// Fetch feed with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			resp, err := c.Fetch(ctx, tt.url, FeedCache{})
			if err != nil {
				t.Fatalf("Failed to fetch %s: %v", tt.name, err)
			}

			if resp.NotModified {
				t.Fatalf("Unexpected 304 Not Modified on first fetch")
			}

			// Verify we got data
			if len(resp.Body) == 0 {
				t.Fatal("Empty response body")
			}

			// Parse the feed
			n := normalizer.New()
			metadata, entries, err := n.Parse(resp.Body, tt.url, resp.FetchTime)
			if err != nil {
				t.Fatalf("Failed to parse feed: %v", err)
			}

			// Verify metadata
			if metadata.Title != tt.expectedTitle {
				t.Errorf("Feed title = %q, want %q", metadata.Title, tt.expectedTitle)
			}

			if metadata.Link == "" {
				t.Error("Feed link should not be empty")
			}

			// Verify entries
			if len(entries) < tt.minEntries {
				t.Errorf("Got %d entries, want at least %d", len(entries), tt.minEntries)
			}

			// Verify first entry has required fields
			if len(entries) > 0 {
				entry := entries[0]

				if entry.ID == "" {
					t.Error("Entry ID should not be empty")
				}

				if entry.Title == "" {
					t.Error("Entry title should not be empty")
				}

				if entry.Link == "" {
					t.Error("Entry link should not be empty")
				}

				if entry.Published.IsZero() {
					t.Error("Entry published date should not be zero")
				}

				t.Logf("First entry: %q", entry.Title)
				t.Logf("  Link: %s", entry.Link)
				t.Logf("  Published: %s", entry.Published.Format(time.RFC3339))
			}

			// Verify cache headers were captured
			if resp.NewCache.ETag == "" && resp.NewCache.LastModified == "" {
				t.Log("Warning: No cache headers received (ETag or Last-Modified)")
			} else {
				t.Logf("Cache headers: ETag=%q, Last-Modified=%q",
					resp.NewCache.ETag, resp.NewCache.LastModified)
			}

			t.Logf("✓ Successfully fetched and parsed %d entries from %s", len(entries), tt.name)
		})
	}
}

// TestLiveFetchConditionalRequest tests that conditional requests work with real servers
func TestLiveFetchConditionalRequest(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("Skipping live network test in short mode")
	}

	url := "https://daringfireball.net/feeds/main"
	c := New()

	// First fetch
	ctx1, cancel1 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel1()

	resp1, err := c.Fetch(ctx1, url, FeedCache{})
	if err != nil {
		t.Fatalf("First fetch failed: %v", err)
	}

	// Second fetch with cache headers
	ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel2()

	cache := FeedCache{
		URL:          url,
		ETag:         resp1.NewCache.ETag,
		LastModified: resp1.NewCache.LastModified,
		LastFetched:  resp1.FetchTime,
	}

	resp2, err := c.Fetch(ctx2, url, cache)
	if err != nil {
		t.Fatalf("Second fetch failed: %v", err)
	}

	// The server might return 304 Not Modified or fresh content
	// Both are valid responses
	if resp2.NotModified {
		t.Log("✓ Server returned 304 Not Modified (cache hit)")
	} else {
		t.Log("✓ Server returned fresh content (cache miss or no caching)")
		if len(resp2.Body) == 0 {
			t.Error("Expected response body when not using cached version")
		}
	}
}

// TestLiveEndToEndPipeline tests the complete pipeline with live fetching
func TestLiveEndToEndPipeline(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("Skipping live network test in short mode")
	}

	t.Run("Daring Fireball full pipeline", func(t *testing.T) {
		testLiveEndToEnd(t, "https://daringfireball.net/feeds/main", "Daring Fireball", 10)
	})

	t.Run("Asymco full pipeline", func(t *testing.T) {
		testLiveEndToEnd(t, "https://asymco.com/feed/", "Asymco", 5)
	})
}

func testLiveEndToEnd(t *testing.T, url, expectedTitle string, minEntries int) {
	// Setup temporary database
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	repo, err := repository.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	// Add feed
	feedID, err := repo.AddFeed(url, "")
	if err != nil {
		t.Fatalf("Failed to add feed: %v", err)
	}

	// Fetch feed
	c := New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.Fetch(ctx, url, FeedCache{})
	if err != nil {
		t.Fatalf("Failed to fetch feed: %v", err)
	}

	// Parse feed
	n := normalizer.New()
	metadata, entries, err := n.Parse(resp.Body, url, resp.FetchTime)
	if err != nil {
		t.Fatalf("Failed to parse feed: %v", err)
	}

	// Verify metadata
	if metadata.Title != expectedTitle {
		t.Errorf("Feed title = %q, want %q", metadata.Title, expectedTitle)
	}

	// Store entries
	for _, entry := range entries {
		repoEntry := &repository.Entry{
			FeedID:      feedID,
			EntryID:     entry.ID,
			Title:       entry.Title,
			Link:        entry.Link,
			Author:      entry.Author,
			Published:   entry.Published,
			Updated:     entry.Updated,
			Content:     entry.Content,
			ContentType: entry.ContentType,
			Summary:     entry.Summary,
			FirstSeen:   entry.FirstSeen,
		}

		if err := repo.UpsertEntry(repoEntry); err != nil {
			t.Fatalf("Failed to store entry: %v", err)
		}
	}

	// Retrieve entries
	dbEntries, err := repo.GetRecentEntries(7)
	if err != nil {
		t.Fatalf("Failed to retrieve entries: %v", err)
	}

	if len(dbEntries) < minEntries {
		t.Errorf("Got %d entries from database, want at least %d", len(dbEntries), minEntries)
	}

	// Update feed metadata
	if err := repo.UpdateFeed(feedID, metadata.Title, metadata.Link, metadata.Updated); err != nil {
		t.Fatalf("Failed to update feed: %v", err)
	}

	// Update cache
	if err := repo.UpdateFeedCache(feedID, resp.NewCache.ETag, resp.NewCache.LastModified, resp.FetchTime); err != nil {
		t.Fatalf("Failed to update cache: %v", err)
	}

	// Verify feed was updated
	feed, err := repo.GetFeedByURL(url)
	if err != nil {
		t.Fatalf("Failed to get feed: %v", err)
	}

	if feed.Title != expectedTitle {
		t.Errorf("Stored feed title = %q, want %q", feed.Title, expectedTitle)
	}

	t.Logf("✓ Successfully completed full pipeline: fetch -> parse -> store -> retrieve")
	t.Logf("  Entries: %d fetched, %d stored, %d retrieved", len(entries), len(entries), len(dbEntries))
	t.Logf("  Feed: %s", feed.Title)
	t.Logf("  Cache: ETag=%q, Last-Modified=%q", feed.ETag, feed.LastModified)
}
