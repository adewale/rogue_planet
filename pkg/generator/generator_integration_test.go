package generator

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adewale/rogue_planet/pkg/crawler"
	"github.com/adewale/rogue_planet/pkg/normalizer"
	"github.com/adewale/rogue_planet/pkg/repository"
	"golang.org/x/net/html"
)

// TestEndToEndHTMLGeneration tests the full pipeline from fetching to HTML generation
func TestEndToEndHTMLGeneration(t *testing.T) {
	t.Parallel()
	// Setup temporary directory
	tmpDir := t.TempDir()

	// Read test feed data
	testFeedData, err := os.ReadFile("../../testdata/test-feed.xml")
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	// Create mock HTTP server serving the test feed
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		if _, err := w.Write(testFeedData); err != nil {
			t.Errorf("Write error: %v", err)
		}
	}))
	defer server.Close()

	// Create test database
	dbPath := filepath.Join(tmpDir, "test.db")
	repo, err := repository.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	// Add feed
	feedID, err := repo.AddFeed(server.URL, "Test Blog")
	if err != nil {
		t.Fatalf("Failed to add feed: %v", err)
	}

	// Fetch feed
	c := crawler.NewForTesting() // Skip SSRF check for localhost
	n := normalizer.New()

	ctx := context.Background()
	resp, err := c.Fetch(ctx, server.URL, crawler.FeedCache{})
	if err != nil {
		t.Fatalf("Failed to fetch feed: %v", err)
	}

	// Parse feed
	metadata, entries, err := n.Parse(resp.Body, server.URL, time.Now())
	if err != nil {
		t.Fatalf("Failed to parse feed: %v", err)
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

	// Get entries from database
	dbEntries, err := repo.GetRecentEntries(7)
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}

	if len(dbEntries) == 0 {
		t.Fatal("No entries in database")
	}

	// Convert to generator format
	genEntries := make([]EntryData, 0, len(dbEntries))
	for _, entry := range dbEntries {
		genEntries = append(genEntries, EntryData{
			Title:     template.HTML(entry.Title),
			Link:      entry.Link,
			Author:    entry.Author,
			FeedTitle: metadata.Title,
			FeedLink:  metadata.Link,
			Published: entry.Published,
			Updated:   entry.Updated,
			Content:   template.HTML(entry.Content),
			Summary:   template.HTML(entry.Summary),
		})
	}

	// Generate HTML
	gen, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "index.html")
	data := TemplateData{
		Title:       "Test Planet",
		Link:        "https://planet.example.com",
		OwnerName:   "Test Owner",
		OwnerEmail:  "test@example.com",
		Entries:     genEntries,
		GroupByDate: true,
	}

	if err := gen.GenerateToFile(outputPath, data); err != nil {
		t.Fatalf("Failed to generate HTML: %v", err)
	}

	// Verify HTML was generated
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("HTML file was not generated")
	}

	// Parse and verify HTML content
	doc, err := parseHTMLFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to parse generated HTML: %v", err)
	}

	// Test 1: Verify feed title appears
	if !containsText(doc, "Test Blog") {
		t.Error("Generated HTML does not contain feed title 'Test Blog'")
	}

	// Test 2: Verify entry titles appear
	if !containsText(doc, "Test Entry 1") {
		t.Error("Generated HTML does not contain 'Test Entry 1'")
	}
	if !containsText(doc, "Test Entry 2") {
		t.Error("Generated HTML does not contain 'Test Entry 2'")
	}

	// Test 3: Verify entry content appears
	if !containsText(doc, "This is the first test entry content") {
		t.Error("Generated HTML does not contain entry 1 content")
	}

	// Test 4: Verify planet title appears
	if !containsText(doc, "Test Planet") {
		t.Error("Generated HTML does not contain planet title")
	}

	// Test 5: Verify CSP meta tag exists
	if !hasCSPHeader(doc) {
		t.Error("Generated HTML does not have Content-Security-Policy meta tag")
	}

	// Test 6: Verify links are present
	if !hasLink(doc, "https://example.com/entry1") {
		t.Error("Generated HTML does not contain link to entry 1")
	}

	t.Logf("Successfully generated and verified HTML with %d entries", len(genEntries))
}

// TestHTMLGenerationWithNoEntries tests generating HTML when there are no entries
func TestHTMLGenerationWithNoEntries(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	gen, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "index.html")
	data := TemplateData{
		Title:       "Empty Planet",
		Link:        "https://planet.example.com",
		OwnerName:   "Test Owner",
		OwnerEmail:  "test@example.com",
		Entries:     []EntryData{},
		GroupByDate: false,
	}

	if err := gen.GenerateToFile(outputPath, data); err != nil {
		t.Fatalf("Failed to generate HTML: %v", err)
	}

	// Verify HTML was generated even with no entries
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("HTML file was not generated")
	}

	doc, err := parseHTMLFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to parse generated HTML: %v", err)
	}

	// Should still have planet title
	if !containsText(doc, "Empty Planet") {
		t.Error("Generated HTML does not contain planet title")
	}
}

// Helper functions for HTML parsing

func parseHTMLFile(path string) (*html.Node, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return html.Parse(f)
}

func containsText(n *html.Node, text string) bool {
	if n.Type == html.TextNode && strings.Contains(n.Data, text) {
		return true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if containsText(c, text) {
			return true
		}
	}

	return false
}

func hasCSPHeader(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "meta" {
		for _, attr := range n.Attr {
			if attr.Key == "http-equiv" && strings.Contains(attr.Val, "Content-Security-Policy") {
				return true
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if hasCSPHeader(c) {
			return true
		}
	}

	return false
}

func hasLink(n *html.Node, href string) bool {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" && attr.Val == href {
				return true
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if hasLink(c, href) {
			return true
		}
	}

	return false
}

// TestGeneratedHTMLStructure verifies the HTML structure is correct
func TestGeneratedHTMLStructure(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	gen, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Create test entry
	now := time.Now()
	entries := []EntryData{
		{
			Title:     template.HTML("Test Entry"),
			Link:      "https://example.com/entry",
			Author:    "Test Author",
			FeedTitle: "Test Feed",
			FeedLink:  "https://example.com",
			Published: now,
			Updated:   now,
			Content:   template.HTML("<p>Test content</p>"),
		},
	}

	outputPath := filepath.Join(tmpDir, "index.html")
	data := TemplateData{
		Title:       "Test Planet",
		Link:        "https://planet.example.com",
		OwnerName:   "Test Owner",
		OwnerEmail:  "test@example.com",
		Entries:     entries,
		GroupByDate: false,
	}

	if err := gen.GenerateToFile(outputPath, data); err != nil {
		t.Fatalf("Failed to generate HTML: %v", err)
	}

	doc, err := parseHTMLFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to parse generated HTML: %v", err)
	}

	// Verify basic HTML structure
	tests := []struct {
		name string
		test func(*html.Node) bool
		desc string
	}{
		{"has html tag", func(n *html.Node) bool { return hasTag(n, "html") }, "HTML tag"},
		{"has head tag", func(n *html.Node) bool { return hasTag(n, "head") }, "HEAD tag"},
		{"has body tag", func(n *html.Node) bool { return hasTag(n, "body") }, "BODY tag"},
		{"has title tag", func(n *html.Node) bool { return hasTag(n, "title") }, "TITLE tag"},
		{"has entry title", func(n *html.Node) bool { return containsText(n, "Test Entry") }, "Entry title"},
		{"has feed attribution", func(n *html.Node) bool { return containsText(n, "Test Feed") }, "Feed title"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.test(doc) {
				t.Errorf("Generated HTML missing: %s", tt.desc)
			}
		})
	}
}

func hasTag(n *html.Node, tag string) bool {
	if n.Type == html.ElementNode && n.Data == tag {
		return true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if hasTag(c, tag) {
			return true
		}
	}

	return false
}
