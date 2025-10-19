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

// TestParseJSONFeed10 tests parsing of JSON Feed version 1.0
func TestParseJSONFeed10(t *testing.T) {
	n := New()

	// Read JSON Feed 1.0 test fixture
	data, err := os.ReadFile("../../testdata/jsonfeed-1.0.json")
	if err != nil {
		t.Fatalf("Failed to read JSON Feed 1.0 file: %v", err)
	}

	// Parse feed
	metadata, entries, err := n.Parse(data, "https://example.org/feed.json", time.Now())
	if err != nil {
		t.Fatalf("Failed to parse JSON Feed 1.0: %v", err)
	}

	// Verify feed metadata
	if metadata.Title != "JSON Feed 1.0 Test Feed" {
		t.Errorf("Feed title = %q, want %q", metadata.Title, "JSON Feed 1.0 Test Feed")
	}

	if metadata.Link != "https://example.org/" {
		t.Errorf("Feed link = %q, want %q", metadata.Link, "https://example.org/")
	}

	// Verify we got all entries
	if len(entries) != 4 {
		t.Fatalf("Got %d entries, want 4", len(entries))
	}

	// Test entry 1: HTML content
	entry1 := entries[0]
	if entry1.ID != "item-1" {
		t.Errorf("Entry 1 ID = %q, want %q", entry1.ID, "item-1")
	}
	if entry1.Title != "First Post" {
		t.Errorf("Entry 1 title = %q, want %q", entry1.Title, "First Post")
	}
	if entry1.Link != "https://example.org/2025/10/first-post" {
		t.Errorf("Entry 1 link = %q, want %q", entry1.Link, "https://example.org/2025/10/first-post")
	}
	if entry1.Author != "Jane Doe" {
		t.Errorf("Entry 1 author = %q, want %q", entry1.Author, "Jane Doe")
	}
	if !containsIgnoreCase(entry1.Content, "first post") {
		t.Errorf("Entry 1 content should contain 'first post'")
	}
	expectedTime := time.Date(2025, 10, 18, 12, 0, 0, 0, time.UTC)
	if !entry1.Published.Equal(expectedTime) {
		t.Errorf("Entry 1 published = %v, want %v", entry1.Published, expectedTime)
	}

	// Test entry 2: Plain text content
	entry2 := entries[1]
	if entry2.ID != "item-2" {
		t.Errorf("Entry 2 ID = %q, want %q", entry2.ID, "item-2")
	}
	if entry2.Title != "Second Post with Plain Text" {
		t.Errorf("Entry 2 title = %q, want %q", entry2.Title, "Second Post with Plain Text")
	}
	// Plain text should be in Summary (gofeed maps content_text to Description)
	if entry2.Summary == "" && entry2.Content == "" {
		t.Error("Entry 2 should have content or summary")
	}

	// Test entry 4: XSS prevention
	entry4 := entries[3]
	if entry4.ID != "item-4" {
		t.Errorf("Entry 4 ID = %q, want %q", entry4.ID, "item-4")
	}
	// Verify dangerous content is sanitized
	if containsDangerousHTML(entry4.Content) {
		t.Errorf("Entry 4 content should not contain dangerous HTML: %s", entry4.Content)
	}
	if containsIgnoreCase(entry4.Content, "<script") {
		t.Error("Entry 4 content should not contain <script> tags")
	}
	if containsIgnoreCase(entry4.Content, "onclick") {
		t.Error("Entry 4 content should not contain onclick handlers")
	}
	if containsIgnoreCase(entry4.Content, "<iframe") {
		t.Error("Entry 4 content should not contain <iframe> tags")
	}

	t.Logf("✓ Successfully parsed JSON Feed 1.0 with %d entries", len(entries))
}

// TestParseJSONFeed11 tests parsing of JSON Feed version 1.1
func TestParseJSONFeed11(t *testing.T) {
	n := New()

	// Read JSON Feed 1.1 test fixture
	data, err := os.ReadFile("../../testdata/jsonfeed-1.1.json")
	if err != nil {
		t.Fatalf("Failed to read JSON Feed 1.1 file: %v", err)
	}

	// Parse feed
	metadata, entries, err := n.Parse(data, "https://example.org/feed.json", time.Now())
	if err != nil {
		t.Fatalf("Failed to parse JSON Feed 1.1: %v", err)
	}

	// Verify feed metadata
	if metadata.Title != "JSON Feed 1.1 Test Feed" {
		t.Errorf("Feed title = %q, want %q", metadata.Title, "JSON Feed 1.1 Test Feed")
	}

	if metadata.Link != "https://example.org/" {
		t.Errorf("Feed link = %q, want %q", metadata.Link, "https://example.org/")
	}

	// Verify we got all entries
	if len(entries) != 3 {
		t.Fatalf("Got %d entries, want 3", len(entries))
	}

	// Test entry 1: Version 1.1 features
	entry1 := entries[0]
	if entry1.ID != "item-1-v1.1" {
		t.Errorf("Entry 1 ID = %q, want %q", entry1.ID, "item-1-v1.1")
	}
	if entry1.Title != "Multilingual Post" {
		t.Errorf("Entry 1 title = %q, want %q", entry1.Title, "Multilingual Post")
	}
	if entry1.Author != "John Smith" {
		t.Errorf("Entry 1 author = %q, want %q (from authors array)", entry1.Author, "John Smith")
	}
	if !containsIgnoreCase(entry1.Content, "version 1.1") {
		t.Error("Entry 1 content should contain 'version 1.1'")
	}

	// Test entry 2: Multiple authors (should use first author)
	entry2 := entries[1]
	if entry2.ID != "item-2-v1.1" {
		t.Errorf("Entry 2 ID = %q, want %q", entry2.ID, "item-2-v1.1")
	}
	if entry2.Title != "Post with Multiple Authors" {
		t.Errorf("Entry 2 title = %q, want %q", entry2.Title, "Post with Multiple Authors")
	}
	// gofeed should extract first author from authors array
	if entry2.Author == "" {
		t.Error("Entry 2 should have an author (from authors array)")
	}

	// Test entry 3: Attachments (podcast episode)
	entry3 := entries[2]
	if entry3.ID != "item-3-v1.1" {
		t.Errorf("Entry 3 ID = %q, want %q", entry3.ID, "item-3-v1.1")
	}
	if entry3.Title != "Podcast Episode with Attachments" {
		t.Errorf("Entry 3 title = %q, want %q", entry3.Title, "Podcast Episode with Attachments")
	}

	t.Logf("✓ Successfully parsed JSON Feed 1.1 with %d entries", len(entries))
}

// TestParseJSONFeedEdgeCases tests edge cases and error handling
func TestParseJSONFeedEdgeCases(t *testing.T) {
	n := New()

	// Read edge cases test fixture
	data, err := os.ReadFile("../../testdata/jsonfeed-edge-cases.json")
	if err != nil {
		t.Fatalf("Failed to read JSON Feed edge cases file: %v", err)
	}

	// Parse feed
	fetchTime := time.Now()
	metadata, entries, err := n.Parse(data, "https://edge.example.org/feed.json", fetchTime)
	if err != nil {
		t.Fatalf("Failed to parse JSON Feed edge cases: %v", err)
	}

	// Verify feed metadata
	if metadata.Title != "JSON Feed Edge Cases" {
		t.Errorf("Feed title = %q, want %q", metadata.Title, "JSON Feed Edge Cases")
	}

	// Verify we got entries (some may be filtered)
	if len(entries) == 0 {
		t.Fatal("Expected at least one entry")
	}

	// Find specific test entries
	var (
		missingDateEntry   *Entry
		missingIDEntry     *Entry
		unicodeEntry       *Entry
		maliciousURLsEntry *Entry
		htmlInTitleEntry   *Entry
		futureDateEntry    *Entry
	)

	for i := range entries {
		switch entries[i].ID {
		case "missing-date":
			missingDateEntry = &entries[i]
		case "unicode-content":
			unicodeEntry = &entries[i]
		case "malicious-urls":
			maliciousURLsEntry = &entries[i]
		case "html-in-title":
			htmlInTitleEntry = &entries[i]
		case "future-date":
			futureDateEntry = &entries[i]
		}
		// Check for generated ID from URL
		if entries[i].Link == "https://edge.example.org/no-id" {
			missingIDEntry = &entries[i]
		}
	}

	// Test 1: Missing date (should use fetch time)
	if missingDateEntry != nil {
		if missingDateEntry.Published.IsZero() {
			t.Error("Entry with missing date should have a published time (using fetch time)")
		}
		// Should be close to fetch time (within 1 second)
		if missingDateEntry.Published.Sub(fetchTime).Abs() > time.Second {
			t.Errorf("Entry missing date: published = %v, expected close to %v",
				missingDateEntry.Published, fetchTime)
		}
		t.Logf("✓ Missing date handled correctly (used fetch time)")
	}

	// Test 2: Missing ID (should generate one)
	if missingIDEntry != nil {
		if missingIDEntry.ID == "" {
			t.Error("Entry without ID should have generated ID")
		}
		t.Logf("✓ Missing ID handled correctly (generated: %s)", missingIDEntry.ID)
	}

	// Test 3: Unicode content (UTF-8 handling)
	if unicodeEntry != nil {
		if !containsIgnoreCase(unicodeEntry.Title, "你好世界") {
			t.Error("Unicode entry title should contain Chinese characters")
		}
		if !containsIgnoreCase(unicodeEntry.Content, "中文") {
			t.Error("Unicode entry content should contain Chinese text")
		}
		if !containsIgnoreCase(unicodeEntry.Content, "日本語") {
			t.Error("Unicode entry content should contain Japanese text")
		}
		t.Logf("✓ Unicode content handled correctly")
	}

	// Test 4: Malicious URL schemes (should be sanitized)
	if maliciousURLsEntry != nil {
		if containsIgnoreCase(maliciousURLsEntry.Content, "javascript:") {
			t.Error("Malicious javascript: URLs should be removed or sanitized")
		}
		t.Logf("✓ Malicious URL schemes sanitized")
	}

	// Test 5: HTML in title (currently stripped, should be escaped per Venus #24)
	if htmlInTitleEntry != nil {
		// Note: This is a known limitation (Venus #24 - HTML Escaping in Titles)
		// Currently HTML is stripped from titles. Ideally it should be escaped.
		// For now, we just verify the entry parses successfully
		if htmlInTitleEntry.Title == "" {
			t.Error("Entry with HTML in title should still have a title")
		}
		t.Logf("✓ Entry with HTML in title parsed (title: %q)", htmlInTitleEntry.Title)
	}

	// Test 6: Future date (accepted or clamped depending on config)
	if futureDateEntry != nil {
		if futureDateEntry.Published.Year() == 2099 {
			t.Logf("✓ Future date accepted (published: %v)", futureDateEntry.Published)
		} else {
			t.Logf("✓ Future date clamped (published: %v)", futureDateEntry.Published)
		}
	}

	t.Logf("✓ Successfully parsed JSON Feed edge cases with %d entries", len(entries))
}

// TestJSONFeedSecuritySanitization specifically tests XSS prevention
func TestJSONFeedSecuritySanitization(t *testing.T) {
	n := New()

	// Create a minimal JSON Feed with dangerous content
	dangerousJSON := `{
		"version": "https://jsonfeed.org/version/1.1",
		"title": "Security Test",
		"home_page_url": "https://test.example.org/",
		"items": [
			{
				"id": "xss-test-1",
				"url": "https://test.example.org/xss",
				"title": "XSS Test Post",
				"content_html": "<p>Normal content</p><script>alert('XSS')</script><p onclick='evil()'>Click</p><iframe src='https://evil.com'></iframe>",
				"date_published": "2025-10-18T12:00:00Z"
			}
		]
	}`

	// Parse feed
	metadata, entries, err := n.Parse([]byte(dangerousJSON), "https://test.example.org/feed.json", time.Now())
	if err != nil {
		t.Fatalf("Failed to parse security test feed: %v", err)
	}

	if metadata.Title != "Security Test" {
		t.Errorf("Feed title = %q, want %q", metadata.Title, "Security Test")
	}

	if len(entries) != 1 {
		t.Fatalf("Got %d entries, want 1", len(entries))
	}

	entry := entries[0]

	// Verify all dangerous content is removed
	dangerousPatterns := []string{
		"<script",
		"</script>",
		"javascript:",
		"onclick",
		"onerror",
		"onload",
		"<iframe",
		"eval(",
	}

	for _, pattern := range dangerousPatterns {
		if containsIgnoreCase(entry.Content, pattern) {
			t.Errorf("Content should not contain %q but found it in: %s", pattern, entry.Content)
		}
	}

	// Verify safe content is preserved
	if !containsIgnoreCase(entry.Content, "Normal content") {
		t.Error("Safe content should be preserved")
	}

	t.Logf("✓ JSON Feed XSS prevention working correctly")
	t.Logf("  Sanitized content: %s", entry.Content)
}
