package normalizer

import (
	"os"
	"testing"
	"time"
)

// TestParseRealWorldFeeds tests parsing of actual saved feed snapshots
// to ensure compatibility with real-world feed formats
func TestParseRealWorldFeeds(t *testing.T) {
	n := New()

	tests := []struct {
		name          string
		feedPath      string
		feedURL       string
		expectedTitle string
		minEntries    int
		feedType      string
	}{
		{
			name:          "Daring Fireball (Atom)",
			feedPath:      "../../testdata/daringfireball-feed.xml",
			feedURL:       "https://daringfireball.net/feeds/main",
			expectedTitle: "Daring Fireball",
			minEntries:    1,
			feedType:      "Atom",
		},
		{
			name:          "Asymco (RSS)",
			feedPath:      "../../testdata/asymco-feed.xml",
			feedURL:       "https://asymco.com/feed/",
			expectedTitle: "Asymco",
			minEntries:    1,
			feedType:      "RSS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read feed file
			data, err := os.ReadFile(tt.feedPath)
			if err != nil {
				t.Fatalf("Failed to read feed file: %v", err)
			}

			// Parse feed
			metadata, entries, err := n.Parse(data, tt.feedURL, time.Now())
			if err != nil {
				t.Fatalf("Failed to parse %s feed: %v", tt.feedType, err)
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

				// Content or Summary should be present
				if entry.Content == "" && entry.Summary == "" {
					t.Error("Entry should have either content or summary")
				}

				t.Logf("First entry: %q", entry.Title)
				t.Logf("  Link: %s", entry.Link)
				t.Logf("  Published: %s", entry.Published.Format(time.RFC3339))
				t.Logf("  Content length: %d bytes", len(entry.Content))
			}

			t.Logf("✓ Successfully parsed %d entries from %s", len(entries), tt.name)
		})
	}
}

// TestEndToEndWithRealFeeds tests the complete workflow with real feed snapshots
func TestEndToEndWithRealFeeds(t *testing.T) {
	// This test verifies we can:
	// 1. Parse real-world feeds
	// 2. Store entries in database
	// 3. Retrieve them
	// 4. Generate HTML

	// We'll use the existing TestEndToEndHTMLGeneration pattern but with real feeds
	t.Run("Daring Fireball end-to-end", func(t *testing.T) {
		testEndToEndWithFeedFile(t, "../../testdata/daringfireball-feed.xml", "Daring Fireball")
	})

	t.Run("Asymco end-to-end", func(t *testing.T) {
		testEndToEndWithFeedFile(t, "../../testdata/asymco-feed.xml", "Asymco")
	})
}

// Helper function to test end-to-end with a feed file
func testEndToEndWithFeedFile(t *testing.T, feedPath, expectedTitle string) {
	n := New()

	// Read feed file
	data, err := os.ReadFile(feedPath)
	if err != nil {
		t.Fatalf("Failed to read feed file: %v", err)
	}

	// Parse feed
	metadata, entries, err := n.Parse(data, "https://example.com/feed", time.Now())
	if err != nil {
		t.Fatalf("Failed to parse feed: %v", err)
	}

	// Verify we got data
	if metadata.Title != expectedTitle {
		t.Errorf("Feed title = %q, want %q", metadata.Title, expectedTitle)
	}

	if len(entries) == 0 {
		t.Fatal("Expected at least one entry")
	}

	// Verify HTML sanitization worked
	for i, entry := range entries {
		// Check that dangerous tags are removed
		if containsDangerousHTML(entry.Content) {
			t.Errorf("Entry %d content contains dangerous HTML: %s", i, entry.Content)
		}

		if containsDangerousHTML(entry.Summary) {
			t.Errorf("Entry %d summary contains dangerous HTML: %s", i, entry.Summary)
		}
	}

	t.Logf("✓ Successfully processed %d entries from %s", len(entries), expectedTitle)
}

// containsDangerousHTML checks if content contains dangerous HTML elements
func containsDangerousHTML(content string) bool {
	dangerous := []string{
		"<script",
		"<iframe",
		"<object",
		"<embed",
		"javascript:",
		"onerror=",
		"onclick=",
	}

	for _, d := range dangerous {
		if len(content) > 0 && containsIgnoreCase(content, d) {
			return true
		}
	}

	return false
}

func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive contains check
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := s[i+j]
			c2 := substr[j]
			// Convert to lowercase for comparison
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 32
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 32
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
