package normalizer

import (
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

func TestParse(t *testing.T) {
	t.Parallel()
	t.Run("RSS 2.0 feed", func(t *testing.T) {
		feedData := `<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>https://example.com</link>
    <item>
      <title>Test Entry</title>
      <link>https://example.com/post1</link>
      <guid>post-1</guid>
      <pubDate>Mon, 02 Jan 2006 15:04:05 GMT</pubDate>
      <description>Test content</description>
    </item>
  </channel>
</rss>`

		n := New()
		metadata, entries, err := n.Parse([]byte(feedData), "https://example.com/feed", time.Now())

		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		if metadata.Title != "Test Feed" {
			t.Errorf("Title = %q, want %q", metadata.Title, "Test Feed")
		}

		if len(entries) != 1 {
			t.Fatalf("len(entries) = %d, want 1", len(entries))
		}

		entry := entries[0]
		if entry.Title != "Test Entry" {
			t.Errorf("entry.Title = %q, want %q", entry.Title, "Test Entry")
		}

		if entry.ID != "post-1" {
			t.Errorf("entry.ID = %q, want %q", entry.ID, "post-1")
		}
	})

	t.Run("Atom feed", func(t *testing.T) {
		feedData := `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Test Atom Feed</title>
  <link href="https://example.com"/>
  <entry>
    <title>Atom Entry</title>
    <link href="https://example.com/atom-post"/>
    <id>atom-1</id>
    <updated>2006-01-02T15:04:05Z</updated>
    <content>Atom content</content>
  </entry>
</feed>`

		n := New()
		metadata, entries, err := n.Parse([]byte(feedData), "https://example.com/feed", time.Now())

		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		if metadata.Title != "Test Atom Feed" {
			t.Errorf("Title = %q, want %q", metadata.Title, "Test Atom Feed")
		}

		if len(entries) != 1 {
			t.Fatalf("len(entries) = %d, want 1", len(entries))
		}

		entry := entries[0]
		if entry.Title != "Atom Entry" {
			t.Errorf("entry.Title = %q, want %q", entry.Title, "Atom Entry")
		}
	})

	t.Run("empty feed", func(t *testing.T) {
		feedData := `<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>Empty Feed</title>
  </channel>
</rss>`

		n := New()
		metadata, entries, err := n.Parse([]byte(feedData), "https://example.com/feed", time.Now())

		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		if len(entries) != 0 {
			t.Errorf("len(entries) = %d, want 0", len(entries))
		}

		if metadata.Title != "Empty Feed" {
			t.Errorf("Title = %q, want %q", metadata.Title, "Empty Feed")
		}
	})

	t.Run("invalid feed", func(t *testing.T) {
		feedData := `not a valid feed`

		n := New()
		_, _, err := n.Parse([]byte(feedData), "https://example.com/feed", time.Now())

		if err == nil {
			t.Error("Expected error for invalid feed, got nil")
		}
	})
}

func TestSanitizeHTML(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		input       string
		contains    []string
		notContains []string
	}{
		{
			name:     "safe HTML preserved",
			input:    "<p>Hello <strong>world</strong></p>",
			contains: []string{"<p>", "Hello", "<strong>", "world", "</strong>", "</p>"},
		},
		{
			name:        "script tag removed",
			input:       "<p>Test</p><script>alert('xss')</script>",
			contains:    []string{"<p>", "Test", "</p>"},
			notContains: []string{"<script>", "alert", "xss"},
		},
		{
			name:        "javascript: URI removed",
			input:       `<a href="javascript:alert(1)">click</a>`,
			notContains: []string{"javascript:", "alert"},
		},
		{
			name:        "onclick removed",
			input:       `<div onclick="alert(1)">click</div>`,
			notContains: []string{"onclick", "alert"},
		},
		{
			name:        "iframe removed",
			input:       `<iframe src="evil.com"></iframe>`,
			notContains: []string{"<iframe", "evil.com"},
		},
		{
			name:        "object tag removed",
			input:       `<object data="evil.com"></object>`,
			notContains: []string{"<object", "evil.com"},
		},
		{
			name:        "embed tag removed",
			input:       `<embed src="evil.com">`,
			notContains: []string{"<embed", "evil.com"},
		},
		{
			name:     "safe link preserved",
			input:    `<a href="https://example.com">link</a>`,
			contains: []string{"<a", "href", "https://example.com", "link", "</a>"},
		},
		{
			name:     "safe image preserved",
			input:    `<img src="https://example.com/pic.jpg" alt="test">`,
			contains: []string{"<img", "src", "https://example.com/pic.jpg", "alt", "test"},
		},
		{
			name:        "data: URI removed",
			input:       `<img src="data:text/html,<script>alert(1)</script>">`,
			notContains: []string{"data:", "script", "alert"},
		},
	}

	n := New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := n.SanitizeHTML(tt.input)

			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("Result should contain %q, got: %s", s, result)
				}
			}

			for _, s := range tt.notContains {
				if strings.Contains(result, s) {
					t.Errorf("Result should NOT contain %q, got: %s", s, result)
				}
			}
		})
	}
}

func TestIDGeneration(t *testing.T) {
	t.Parallel()
	n := New()

	t.Run("uses existing GUID", func(t *testing.T) {
		item := &gofeed.Item{
			GUID:  "existing-guid",
			Title: "Test",
			Link:  "https://example.com/post",
		}

		id := n.extractID(item, "https://example.com/feed")

		if id != "existing-guid" {
			t.Errorf("ID = %q, want %q", id, "existing-guid")
		}
	})

	t.Run("falls back to link", func(t *testing.T) {
		item := &gofeed.Item{
			Title: "Test",
			Link:  "https://example.com/post",
		}

		id := n.extractID(item, "https://example.com/feed")

		if id != "https://example.com/post" {
			t.Errorf("ID = %q, want %q", id, "https://example.com/post")
		}
	})

	t.Run("generates hash when no GUID or link", func(t *testing.T) {
		item := &gofeed.Item{
			Title: "Test Title",
		}

		id := n.extractID(item, "https://example.com/feed")

		// Should generate a hash
		if id == "" {
			t.Error("ID should not be empty")
		}

		if len(id) != 16 {
			t.Errorf("Generated ID length = %d, want 16", len(id))
		}
	})

	t.Run("same input generates same ID", func(t *testing.T) {
		item := &gofeed.Item{
			Title: "Consistent Title",
		}

		id1 := n.extractID(item, "https://example.com/feed")
		id2 := n.extractID(item, "https://example.com/feed")

		if id1 != id2 {
			t.Errorf("IDs should be consistent: %q != %q", id1, id2)
		}
	})
}

func TestURLResolution(t *testing.T) {
	t.Parallel()
	n := New()

	tests := []struct {
		name        string
		href        string
		baseURL     string
		expected    string
		expectError bool
	}{
		{
			name:     "absolute URL unchanged",
			href:     "https://example.com/page",
			baseURL:  "https://base.com/feed",
			expected: "https://example.com/page",
		},
		{
			name:     "relative URL resolved",
			href:     "/page",
			baseURL:  "https://base.com/feed",
			expected: "https://base.com/page",
		},
		{
			name:     "relative path resolved",
			href:     "page.html",
			baseURL:  "https://base.com/blog/feed",
			expected: "https://base.com/blog/page.html",
		},
		{
			name:     "protocol-relative URL",
			href:     "//example.com/page",
			baseURL:  "https://base.com/feed",
			expected: "https://example.com/page",
		},
		{
			name:        "invalid base URL returns error",
			href:        "relative-link.html",
			baseURL:     "ht!tp://invalid url with spaces",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid item URL returns error",
			href:        "ht!tp://invalid url",
			baseURL:     "https://example.com/feed",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := n.resolveURL(tt.href, tt.baseURL)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for invalid URL, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("resolveURL() error = %v", err)
			}

			if result != tt.expected {
				t.Errorf("resolveURL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Edge case tests for branch coverage

func TestParse_MalformedFeed(t *testing.T) {
	t.Parallel()
	// Test Parse error path (line 100: err != nil)
	feedData := `this is not valid XML at all!`

	n := New()
	metadata, entries, err := n.Parse([]byte(feedData), "https://example.com/feed", time.Now())

	if err == nil {
		t.Error("Expected error for malformed feed")
	}

	if metadata != nil {
		t.Error("Expected nil metadata for malformed feed")
	}

	if entries != nil {
		t.Error("Expected nil entries for malformed feed")
	}
}

// TABLE-DRIVEN: Missing fields edge cases
func TestNormalizeEntry_MissingFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		feedXML     string
		wantTitle   string
		wantLink    string
		wantAuthor  string
		wantID      string
		checkIDFunc func(*testing.T, string) // For hash-based IDs
	}{
		{
			name: "missing link",
			feedXML: `<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>https://example.com</link>
    <item>
      <title>Entry Without Link</title>
      <guid>entry-1</guid>
      <description>Content without a link</description>
    </item>
  </channel>
</rss>`,
			wantTitle:  "Entry Without Link",
			wantLink:   "",
			wantAuthor: "",
			wantID:     "entry-1",
		},
		{
			name: "missing title - uses link as ID",
			feedXML: `<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>https://example.com</link>
    <item>
      <link>https://example.com/post1</link>
      <description>Content without a title</description>
    </item>
  </channel>
</rss>`,
			wantTitle:  "",
			wantLink:   "https://example.com/post1",
			wantAuthor: "",
			wantID:     "https://example.com/post1",
		},
		{
			name: "missing title and link - generates hash ID",
			feedXML: `<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>https://example.com</link>
    <item>
      <description>Content without title or link</description>
    </item>
  </channel>
</rss>`,
			wantTitle:  "",
			wantLink:   "",
			wantAuthor: "",
			checkIDFunc: func(t *testing.T, id string) {
				if len(id) != 16 {
					t.Errorf("Expected 16-char hash ID, got %q (len %d)", id, len(id))
				}
			},
		},
		{
			name: "missing author",
			feedXML: `<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>https://example.com</link>
    <item>
      <title>Entry Without Author</title>
      <link>https://example.com/post1</link>
      <description>Content without author</description>
    </item>
  </channel>
</rss>`,
			wantTitle:  "Entry Without Author",
			wantLink:   "https://example.com/post1",
			wantAuthor: "",
			wantID:     "https://example.com/post1", // Link is used as ID when no GUID
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := New()
			_, entries, err := n.Parse([]byte(tt.feedXML), "https://example.com/feed", time.Now())

			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(entries) != 1 {
				t.Fatalf("len(entries) = %d, want 1", len(entries))
			}

			entry := entries[0]

			if entry.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", entry.Title, tt.wantTitle)
			}

			if entry.Link != tt.wantLink {
				t.Errorf("Link = %q, want %q", entry.Link, tt.wantLink)
			}

			if entry.Author != tt.wantAuthor {
				t.Errorf("Author = %q, want %q", entry.Author, tt.wantAuthor)
			}

			if tt.checkIDFunc != nil {
				tt.checkIDFunc(t, entry.ID)
			} else if entry.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", entry.ID, tt.wantID)
			}
		})
	}
}

func TestNormalizeEntry_WithPublishedDate(t *testing.T) {
	t.Parallel()
	// Test branch where item.PublishedParsed is non-nil (line 173: item.PublishedParsed != nil tested as true)
	feedData := `<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>https://example.com</link>
    <item>
      <title>Entry With Date</title>
      <link>https://example.com/post1</link>
      <guid>entry-1</guid>
      <pubDate>Mon, 02 Jan 2006 15:04:05 GMT</pubDate>
      <description>Content</description>
    </item>
  </channel>
</rss>`

	n := New()
	_, entries, err := n.Parse([]byte(feedData), "https://example.com/feed", time.Now())

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}

	entry := entries[0]

	// Verify published date was used
	expectedTime := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	if !entry.Published.Equal(expectedTime) {
		t.Errorf("Published = %v, want %v", entry.Published, expectedTime)
	}
}

func TestExtractPublished_ZeroTime(t *testing.T) {
	t.Parallel()
	// Test branch where PublishedParsed.IsZero() is true (line 210)
	feedData := `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Test Feed</title>
  <link href="https://example.com"/>
  <updated>2006-01-02T15:04:05Z</updated>
  <entry>
    <title>Entry With Zero Published Time</title>
    <link href="https://example.com/post1"/>
    <id>entry-1</id>
    <content>Content</content>
  </entry>
</feed>`

	fetchTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	n := New()
	_, entries, err := n.Parse([]byte(feedData), "https://example.com/feed", fetchTime)

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}

	entry := entries[0]

	// Should fall back to feed.UpdatedParsed, not fetchTime
	feedUpdated := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	if !entry.Published.Equal(feedUpdated) {
		t.Errorf("Published = %v, want %v (feed updated)", entry.Published, feedUpdated)
	}
}
