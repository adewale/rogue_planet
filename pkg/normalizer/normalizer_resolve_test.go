package normalizer

import (
	"context"
	"strings"
	"testing"
	"time"
)

// Tests for L12: Relative URL resolution in HTML content

func TestResolveRelativeURLs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		html        string
		baseURL     string
		contains    []string
		notContains []string
	}{
		{
			name:     "absolute URL unchanged",
			html:     `<a href="https://example.com/page">link</a>`,
			baseURL:  "https://base.com/feed",
			contains: []string{`href="https://example.com/page"`},
		},
		{
			name:     "relative path resolved",
			html:     `<a href="/page">link</a>`,
			baseURL:  "https://base.com/feed",
			contains: []string{`href="https://base.com/page"`},
		},
		{
			name:     "relative path without leading slash resolved",
			html:     `<a href="page.html">link</a>`,
			baseURL:  "https://base.com/blog/feed",
			contains: []string{`href="https://base.com/blog/page.html"`},
		},
		{
			name:     "image src resolved",
			html:     `<img src="/images/photo.jpg">`,
			baseURL:  "https://example.com/feed",
			contains: []string{`src="https://example.com/images/photo.jpg"`},
		},
		{
			name:     "protocol-relative URL resolved",
			html:     `<a href="//cdn.example.com/page">link</a>`,
			baseURL:  "https://base.com/feed",
			contains: []string{`href="https://cdn.example.com/page"`},
		},
		{
			name:     "empty baseURL skips resolution",
			html:     `<a href="/page">link</a>`,
			baseURL:  "",
			contains: []string{`href="/page"`},
		},
		{
			name:     "malformed baseURL returns original",
			html:     `<a href="/page">link</a>`,
			baseURL:  "ht!tp://invalid url with spaces",
			contains: []string{`/page`},
		},
		{
			name:     "multiple links resolved",
			html:     `<a href="/page1">link1</a><a href="/page2">link2</a>`,
			baseURL:  "https://example.com/feed",
			contains: []string{`href="https://example.com/page1"`, `href="https://example.com/page2"`},
		},
		{
			name:     "mixed absolute and relative",
			html:     `<a href="https://other.com/page">abs</a><img src="/img.png">`,
			baseURL:  "https://example.com/feed",
			contains: []string{`href="https://other.com/page"`, `src="https://example.com/img.png"`},
		},
		{
			name:     "empty href left alone",
			html:     `<a href="">link</a>`,
			baseURL:  "https://example.com/feed",
			contains: []string{"link"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveRelativeURLs(tt.html, tt.baseURL)
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

func TestResolveRelativeURLs_MalformedHTML(t *testing.T) {
	t.Parallel()

	// Malformed HTML should not panic, should return something reasonable
	input := `<a href="/page">link<p>unclosed`
	result := resolveRelativeURLs(input, "https://example.com/feed")

	// Should still contain the link text
	if !strings.Contains(result, "link") {
		t.Errorf("Result should contain link text, got: %s", result)
	}
}

func TestSanitizeHTML_ResolvesRelativeURLs(t *testing.T) {
	t.Parallel()
	n := New()

	// Test that sanitizeHTML resolves relative URLs in content
	html := `<p>Check <a href="/page">this page</a> and <img src="/img.jpg" alt="photo"></p>`
	result := n.sanitizeHTML(html, "https://example.com/feed")

	if !strings.Contains(result, "https://example.com/page") {
		t.Errorf("sanitizeHTML should resolve relative href, got: %s", result)
	}
	if !strings.Contains(result, "https://example.com/img.jpg") {
		t.Errorf("sanitizeHTML should resolve relative src, got: %s", result)
	}
}

func TestParse_ContentRelativeURLsResolved(t *testing.T) {
	t.Parallel()
	n := New()

	feedData := `<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>https://example.com</link>
    <item>
      <title>Test Entry</title>
      <link>https://example.com/post1</link>
      <guid>post-1</guid>
      <description>&lt;p&gt;Check &lt;a href="/page"&gt;this&lt;/a&gt; and &lt;img src="/img.jpg"&gt;&lt;/p&gt;</description>
      <content:encoded>&lt;p&gt;Full content with &lt;a href="/other"&gt;link&lt;/a&gt;&lt;/p&gt;</content:encoded>
    </item>
  </channel>
</rss>`

	_, entries, err := n.Parse(context.Background(), []byte(feedData), "https://example.com/feed", time.Now())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}

	entry := entries[0]

	// Content should have resolved URLs
	if !strings.Contains(entry.Content, "https://example.com/other") {
		t.Errorf("Content should have resolved relative URL, got: %s", entry.Content)
	}

	// Summary should have resolved URLs too
	if entry.Summary != "" && strings.Contains(entry.Summary, `href="/page"`) {
		t.Errorf("Summary should have resolved relative URL, got: %s", entry.Summary)
	}
}

