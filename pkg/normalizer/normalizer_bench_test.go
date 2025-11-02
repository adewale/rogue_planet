package normalizer

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

// BenchmarkSanitizeHTML benchmarks HTML sanitization performance
func BenchmarkSanitizeHTML(b *testing.B) {
	n := New()
	html := `<p>This is a <strong>test</strong> with <a href="https://example.com">links</a> and <img src="https://example.com/image.jpg" alt="image">.</p>
<script>alert('xss')</script>
<div onclick="alert('xss')">Click me</div>
<p>More content with <em>emphasis</em> and <code>code blocks</code>.</p>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = n.sanitizeHTML(html, "https://example.com")
	}
}

// BenchmarkSanitizeHTMLLarge benchmarks sanitization of large HTML content
func BenchmarkSanitizeHTMLLarge(b *testing.B) {
	n := New()
	// Create a large HTML document
	var sb strings.Builder
	sb.WriteString("<article>")
	for i := 0; i < 100; i++ {
		sb.WriteString("<p>This is paragraph ")
		sb.WriteString(string(rune(i)))
		sb.WriteString(" with <strong>formatting</strong> and <a href=\"https://example.com/")
		sb.WriteString(string(rune(i)))
		sb.WriteString("\">links</a>.</p>")
	}
	sb.WriteString("</article>")
	html := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = n.sanitizeHTML(html, "https://example.com")
	}
}

// BenchmarkParse benchmarks full feed parsing and normalization
func BenchmarkParse(b *testing.B) {
	n := New()
	feedXML := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
	<title>Test Feed</title>
	<link>https://example.com</link>
	<description>A test feed</description>
	<item>
		<title>Test Entry</title>
		<link>https://example.com/entry1</link>
		<description><![CDATA[<p>Test content with <strong>HTML</strong>.</p>]]></description>
		<pubDate>Mon, 01 Jan 2024 00:00:00 GMT</pubDate>
		<guid>entry-1</guid>
	</item>
</channel>
</rss>`

	feedURL := "https://example.com/feed"
	fetchTime := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = n.Parse(context.Background(), []byte(feedXML), feedURL, fetchTime)
	}
}

// BenchmarkExtractID benchmarks ID extraction logic
func BenchmarkExtractID(b *testing.B) {
	n := New()
	// Use a real gofeed.Item for benchmarking
	item := &gofeed.Item{
		GUID: "",
		Link: "https://example.com/entry/12345",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = n.extractID(item, "https://example.com/feed")
	}
}
