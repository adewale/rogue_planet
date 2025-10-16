package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adewale/rogue_planet/pkg/repository"
)

func TestEntrySpamPrevention(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	// Initialize planet with filter_by_first_seen
	initOpts := InitOptions{
		ConfigPath: "./config.ini",
		FeedsFile:  "",
		Output:     io.Discard,
	}
	if err := cmdInit(initOpts); err != nil {
		t.Fatalf("cmdInit() error = %v", err)
	}

	// Update config to enable first_seen filtering
	configContent := `[planet]
name = Test Planet
link = https://test.example.com
owner_name = Test Owner
filter_by_first_seen = true
sort_by = first_seen
days = 7

[database]
path = ./data/planet.db
`
	if err := os.WriteFile(filepath.Join(dir, "config.ini"), []byte(configContent), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Open repository directly to inject test data
	repo, err := repository.New(filepath.Join(dir, "data/planet.db"))
	if err != nil {
		t.Fatalf("repository.New() error = %v", err)
	}
	defer repo.Close()

	// Simulate adding a feed with old entries
	feedID, _ := repo.AddFeed("https://example.com/feed", "Example Feed")

	baseTime := time.Now()

	// Feed has 3 entries:
	// 1. Published 30 days ago, first_seen 10 days ago (old entry, should be filtered out)
	// 2. Published 2 days ago, first_seen today (recent entry, should appear)
	// 3. Published 3 days ago, first_seen 10 days ago (existing entry, should be filtered)
	entries := []repository.Entry{
		{
			FeedID:    feedID,
			EntryID:   "old-entry",
			Title:     "Old Entry (should be filtered)",
			Link:      "https://example.com/old",
			Published: baseTime.AddDate(0, 0, -30),
			FirstSeen: baseTime.AddDate(0, 0, -10), // Discovered 10 days ago
		},
		{
			FeedID:    feedID,
			EntryID:   "recent-entry",
			Title:     "Recent Entry (should appear)",
			Link:      "https://example.com/recent",
			Published: baseTime.AddDate(0, 0, -2),
			FirstSeen: baseTime, // Just discovered
		},
		{
			FeedID:    feedID,
			EntryID:   "existing-entry",
			Title:     "Existing Entry (should be filtered)",
			Link:      "https://example.com/existing",
			Published: baseTime.AddDate(0, 0, -3),
			FirstSeen: baseTime.AddDate(0, 0, -10), // Discovered long ago
		},
	}

	for _, e := range entries {
		if err := repo.UpsertEntry(&e); err != nil {
			t.Fatalf("UpsertEntry() error = %v", err)
		}
	}

	// Generate HTML
	genOpts := GenerateOptions{
		ConfigPath: "./config.ini",
		Output:     io.Discard,
	}
	if err := cmdGenerate(genOpts); err != nil {
		t.Fatalf("cmdGenerate() error = %v", err)
	}

	// Read generated HTML
	htmlContent, err := os.ReadFile(filepath.Join(dir, "public/index.html"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	html := string(htmlContent)

	// Verify only the recent entry appears
	if !strings.Contains(html, "Recent Entry (should appear)") {
		t.Error("HTML should contain 'Recent Entry (should appear)'")
	}

	// Verify old entries are filtered out
	if strings.Contains(html, "Old Entry (should be filtered)") {
		t.Error("HTML should NOT contain 'Old Entry (should be filtered)' - entry spam not prevented!")
	}

	if strings.Contains(html, "Existing Entry (should be filtered)") {
		t.Error("HTML should NOT contain 'Existing Entry (should be filtered)' - first_seen filter failed!")
	}
}

func TestBackwardsCompatibility(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	// Initialize with default config (filter_by_first_seen = false)
	initOpts := InitOptions{
		ConfigPath: "./config.ini",
		Output:     io.Discard,
	}
	if err := cmdInit(initOpts); err != nil {
		t.Fatalf("cmdInit() error = %v", err)
	}

	repo, _ := repository.New(filepath.Join(dir, "data/planet.db"))
	defer repo.Close()

	feedID, _ := repo.AddFeed("https://example.com/feed", "Example Feed")
	baseTime := time.Now()

	// Add entry published recently but first_seen long ago
	if err := repo.UpsertEntry(&repository.Entry{
		FeedID:    feedID,
		EntryID:   "test-entry",
		Title:     "Test Entry",
		Published: baseTime.AddDate(0, 0, -2),
		FirstSeen: baseTime.AddDate(0, 0, -30), // Seen 30 days ago
	}); err != nil {
		t.Fatalf("UpsertEntry() error = %v", err)
	}

	// Generate with default config (should filter by published)
	if err := cmdGenerate(GenerateOptions{ConfigPath: "./config.ini", Output: io.Discard}); err != nil {
		t.Fatalf("cmdGenerate() error = %v", err)
	}

	htmlContent, _ := os.ReadFile(filepath.Join(dir, "public/index.html"))
	html := string(htmlContent)

	// Should appear because published date is recent (default behavior)
	if !strings.Contains(html, "Test Entry") {
		t.Error("Default behavior should filter by published date")
	}
}
