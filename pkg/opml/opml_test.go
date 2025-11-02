package opml

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adewale/rogue_planet/pkg/timeprovider"
)

// Test parsing valid OPML 2.0
func TestParse_OPML20(t *testing.T) {
	t.Parallel()
	opmlData := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head>
    <title>My Feeds</title>
    <dateCreated>Mon, 12 Oct 2025 10:00:00 -0700</dateCreated>
    <ownerName>John Doe</ownerName>
  </head>
  <body>
    <outline text="Daring Fireball" title="Daring Fireball" type="rss" xmlUrl="https://daringfireball.net/feeds/main" htmlUrl="https://daringfireball.net/"/>
  </body>
</opml>`

	opml, err := Parse([]byte(opmlData))
	if err != nil {
		t.Fatalf("Failed to parse OPML 2.0: %v", err)
	}

	if opml.Version != "2.0" {
		t.Errorf("Expected version 2.0, got %s", opml.Version)
	}

	if opml.Head.Title != "My Feeds" {
		t.Errorf("Expected title 'My Feeds', got %s", opml.Head.Title)
	}

	if len(opml.Body.Outlines) != 1 {
		t.Fatalf("Expected 1 outline, got %d", len(opml.Body.Outlines))
	}

	outline := opml.Body.Outlines[0]
	if outline.Text != "Daring Fireball" {
		t.Errorf("Expected text 'Daring Fireball', got %s", outline.Text)
	}

	if outline.XMLUrl != "https://daringfireball.net/feeds/main" {
		t.Errorf("Expected xmlUrl 'https://daringfireball.net/feeds/main', got %s", outline.XMLUrl)
	}
}

// Test parsing OPML 1.0 with 'url' attribute
func TestParse_OPML10_UrlAttribute(t *testing.T) {
	t.Parallel()
	opmlData := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <head>
    <title>My Feeds</title>
  </head>
  <body>
    <outline text="Example Feed" type="rss" url="https://example.com/feed.xml"/>
  </body>
</opml>`

	opml, err := Parse([]byte(opmlData))
	if err != nil {
		t.Fatalf("Failed to parse OPML 1.0: %v", err)
	}

	feeds := opml.ExtractFeeds()
	if len(feeds) != 1 {
		t.Fatalf("Expected 1 feed, got %d", len(feeds))
	}

	if feeds[0].FeedURL != "https://example.com/feed.xml" {
		t.Errorf("Expected feed URL 'https://example.com/feed.xml', got %s", feeds[0].FeedURL)
	}
}

// Test handling both text and title attributes
func TestParse_TextAndTitle(t *testing.T) {
	t.Parallel()
	opmlData := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <body>
    <outline text="Text Only" xmlUrl="https://example.com/1"/>
    <outline text="Has Both" title="Title Wins" xmlUrl="https://example.com/2"/>
  </body>
</opml>`

	opml, err := Parse([]byte(opmlData))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	feeds := opml.ExtractFeeds()
	if len(feeds) != 2 {
		t.Fatalf("Expected 2 feeds, got %d", len(feeds))
	}

	// Feed with text only should use text
	if feeds[0].Title != "Text Only" {
		t.Errorf("Expected title 'Text Only', got %s", feeds[0].Title)
	}

	// Feed with both should prefer title
	if feeds[1].Title != "Title Wins" {
		t.Errorf("Expected title 'Title Wins', got %s", feeds[1].Title)
	}
}

// Test nested outlines (categories)
func TestParse_NestedOutlines(t *testing.T) {
	t.Parallel()
	opmlData := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <body>
    <outline text="Tech Blogs">
      <outline text="Feed 1" xmlUrl="https://example.com/1"/>
      <outline text="Feed 2" xmlUrl="https://example.com/2"/>
    </outline>
    <outline text="News">
      <outline text="Feed 3" xmlUrl="https://example.com/3"/>
    </outline>
  </body>
</opml>`

	opml, err := Parse([]byte(opmlData))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	feeds := opml.ExtractFeeds()
	if len(feeds) != 3 {
		t.Fatalf("Expected 3 feeds (flattened), got %d", len(feeds))
	}

	// Verify all feeds were extracted
	urls := []string{}
	for _, feed := range feeds {
		urls = append(urls, feed.FeedURL)
	}

	expected := []string{
		"https://example.com/1",
		"https://example.com/2",
		"https://example.com/3",
	}

	for _, exp := range expected {
		found := false
		for _, url := range urls {
			if url == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected URL %s not found in extracted feeds", exp)
		}
	}
}

// Test RFC 822 date parsing
func TestParseRFC822(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		wantYear int
	}{
		{
			name:     "RFC1123Z format",
			input:    "Mon, 12 Oct 2025 10:00:00 -0700",
			wantErr:  false,
			wantYear: 2025,
		},
		{
			name:     "RFC1123 format",
			input:    "Mon, 12 Oct 2025 10:00:00 EST",
			wantErr:  false,
			wantYear: 2025,
		},
		{
			name:    "Invalid format",
			input:   "2025-10-12",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseRFC822(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRFC822() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && parsed.Year() != tt.wantYear {
				t.Errorf("Expected year %d, got %d", tt.wantYear, parsed.Year())
			}
		})
	}
}

// Test RFC 822 date formatting
func TestFormatRFC822(t *testing.T) {
	t.Parallel()
	testTime := time.Date(2025, 10, 12, 10, 0, 0, 0, time.FixedZone("MST", -7*3600))
	formatted := FormatRFC822(testTime)

	// Should be in RFC1123Z format
	if !strings.Contains(formatted, "12 Oct 2025") {
		t.Errorf("Expected formatted date to contain '12 Oct 2025', got %s", formatted)
	}

	// Should be parseable
	parsed, err := ParseRFC822(formatted)
	if err != nil {
		t.Errorf("Formatted date should be parseable: %v", err)
	}

	if parsed.Year() != 2025 || parsed.Month() != 10 || parsed.Day() != 12 {
		t.Errorf("Round-trip failed: expected 2025-10-12, got %v", parsed)
	}
}

// Test round-trip: Generate → Parse → Extract
func TestRoundTrip(t *testing.T) {
	t.Parallel()
	originalFeeds := []Feed{
		{Title: "Feed 1", FeedURL: "https://example.com/1", WebURL: "https://example.com"},
		{Title: "Feed 2", FeedURL: "https://example.com/2"},
	}

	metadata := Metadata{
		Title:      "My Feeds",
		OwnerName:  "Test User",
		OwnerEmail: "test@example.com",
	}

	// Generate OPML
	opml, err := Generate(originalFeeds, metadata)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Marshal to XML
	xmlData, err := opml.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Parse back
	parsed, err := Parse(xmlData)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Extract feeds
	extractedFeeds := parsed.ExtractFeeds()

	if len(extractedFeeds) != len(originalFeeds) {
		t.Fatalf("Expected %d feeds, got %d", len(originalFeeds), len(extractedFeeds))
	}

	for i, extracted := range extractedFeeds {
		if extracted.FeedURL != originalFeeds[i].FeedURL {
			t.Errorf("Feed %d: expected URL %s, got %s", i, originalFeeds[i].FeedURL, extracted.FeedURL)
		}
		if extracted.Title != originalFeeds[i].Title {
			t.Errorf("Feed %d: expected title %s, got %s", i, originalFeeds[i].Title, extracted.Title)
		}
	}
}

// Test generating OPML with both text and title
func TestGenerate_TextAndTitle(t *testing.T) {
	t.Parallel()
	feeds := []Feed{
		{Title: "My Feed", FeedURL: "https://example.com/feed"},
	}

	opml, err := Generate(feeds, Metadata{Title: "Test"})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(opml.Body.Outlines) != 1 {
		t.Fatalf("Expected 1 outline, got %d", len(opml.Body.Outlines))
	}

	outline := opml.Body.Outlines[0]
	if outline.Text != "My Feed" {
		t.Errorf("Expected text 'My Feed', got %s", outline.Text)
	}
	if outline.Title != "My Feed" {
		t.Errorf("Expected title 'My Feed', got %s", outline.Title)
	}
}

// Test empty OPML
func TestParse_Empty(t *testing.T) {
	t.Parallel()
	opmlData := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head><title>Empty</title></head>
  <body></body>
</opml>`

	opml, err := Parse([]byte(opmlData))
	if err != nil {
		t.Fatalf("Failed to parse empty OPML: %v", err)
	}

	feeds := opml.ExtractFeeds()
	if len(feeds) != 0 {
		t.Errorf("Expected 0 feeds from empty OPML, got %d", len(feeds))
	}
}

// Test malformed XML
func TestParse_Malformed(t *testing.T) {
	t.Parallel()
	opmlData := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head><title>Malformed</head>
  <body>
</opml>`

	_, err := Parse([]byte(opmlData))
	if err == nil {
		t.Error("Expected error for malformed XML, got nil")
	}
}

// Test special characters in feed titles
func TestParse_SpecialCharacters(t *testing.T) {
	t.Parallel()
	opmlData := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <body>
    <outline text="Feed &amp; More" xmlUrl="https://example.com/feed"/>
    <outline text="Feed &lt;HTML&gt;" xmlUrl="https://example.com/feed2"/>
  </body>
</opml>`

	opml, err := Parse([]byte(opmlData))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	feeds := opml.ExtractFeeds()
	if len(feeds) != 2 {
		t.Fatalf("Expected 2 feeds, got %d", len(feeds))
	}

	if feeds[0].Title != "Feed & More" {
		t.Errorf("Expected title 'Feed & More', got %s", feeds[0].Title)
	}

	if feeds[1].Title != "Feed <HTML>" {
		t.Errorf("Expected title 'Feed <HTML>', got %s", feeds[1].Title)
	}
}

// Test ParseFile
func TestParseFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.opml")

	opmlData := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head><title>Test</title></head>
  <body>
    <outline text="Feed 1" xmlUrl="https://example.com/1"/>
  </body>
</opml>`

	if err := os.WriteFile(filePath, []byte(opmlData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	opml, err := ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if opml.Head.Title != "Test" {
		t.Errorf("Expected title 'Test', got %s", opml.Head.Title)
	}
}

// Test Write method
func TestWrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	filePath := filepath.Join(dir, "output.opml")

	feeds := []Feed{
		{Title: "Test Feed", FeedURL: "https://example.com/feed"},
	}

	opml, err := Generate(feeds, Metadata{Title: "Test"})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if err := opml.Write(context.Background(), filePath); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify file was written
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File was not written")
	}

	// Verify content is valid OPML
	parsed, err := ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("Failed to parse written file: %v", err)
	}

	if parsed.Head.Title != "Test" {
		t.Errorf("Expected title 'Test', got %s", parsed.Head.Title)
	}
}

// Test XML declaration in output
func TestMarshal_XMLDeclaration(t *testing.T) {
	t.Parallel()
	feeds := []Feed{{Title: "Test", FeedURL: "https://example.com/feed"}}
	opml, _ := Generate(feeds, Metadata{Title: "Test"})

	xmlData, err := opml.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	xmlStr := string(xmlData)
	if !strings.HasPrefix(xmlStr, "<?xml") {
		t.Error("Expected XML declaration at start of output")
	}
}

// Test OPML version in output
func TestGenerate_Version(t *testing.T) {
	t.Parallel()
	feeds := []Feed{{Title: "Test", FeedURL: "https://example.com/feed"}}
	opml, _ := Generate(feeds, Metadata{Title: "Test"})

	if opml.Version != "2.0" {
		t.Errorf("Expected OPML version 2.0, got %s", opml.Version)
	}
}

// Test metadata in generated OPML
func TestGenerate_Metadata(t *testing.T) {
	t.Parallel()
	feeds := []Feed{{Title: "Test", FeedURL: "https://example.com/feed"}}
	metadata := Metadata{
		Title:      "My Feed List",
		OwnerName:  "John Doe",
		OwnerEmail: "john@example.com",
	}

	// Use FakeClock for deterministic testing
	fixedTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	clock := timeprovider.NewFakeClock(fixedTime)

	opml, err := GenerateWithTimeProvider(feeds, metadata, clock)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if opml.Head.Title != "My Feed List" {
		t.Errorf("Expected title 'My Feed List', got %s", opml.Head.Title)
	}

	if opml.Head.OwnerName != "John Doe" {
		t.Errorf("Expected owner 'John Doe', got %s", opml.Head.OwnerName)
	}

	if opml.Head.OwnerEmail != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got %s", opml.Head.OwnerEmail)
	}

	// DateCreated should match our exact fixed time
	expectedDate := FormatRFC822(fixedTime)
	if opml.Head.DateCreated != expectedDate {
		t.Errorf("Expected DateCreated %q, got %q", expectedDate, opml.Head.DateCreated)
	}

	// Verify the exact timestamp format
	if opml.Head.DateCreated != "Wed, 15 Jan 2025 14:30:00 +0000" {
		t.Errorf("Expected DateCreated 'Wed, 15 Jan 2025 14:30:00 +0000', got %q", opml.Head.DateCreated)
	}
}

// Test feed without title uses URL
func TestGenerate_NoTitle(t *testing.T) {
	t.Parallel()
	feeds := []Feed{{FeedURL: "https://example.com/feed"}}
	opml, _ := Generate(feeds, Metadata{Title: "Test"})

	if len(opml.Body.Outlines) != 1 {
		t.Fatalf("Expected 1 outline, got %d", len(opml.Body.Outlines))
	}

	outline := opml.Body.Outlines[0]
	if outline.Text != "https://example.com/feed" {
		t.Errorf("Expected text to be URL, got %s", outline.Text)
	}
}

// Test htmlUrl preservation
func TestExtractFeeds_HTMLUrl(t *testing.T) {
	t.Parallel()
	opmlData := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <body>
    <outline text="Feed" xmlUrl="https://example.com/feed.xml" htmlUrl="https://example.com"/>
  </body>
</opml>`

	opml, err := Parse([]byte(opmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	feeds := opml.ExtractFeeds()
	if len(feeds) != 1 {
		t.Fatalf("Expected 1 feed, got %d", len(feeds))
	}

	if feeds[0].WebURL != "https://example.com" {
		t.Errorf("Expected WebURL 'https://example.com', got %s", feeds[0].WebURL)
	}
}

// Test mixed xmlUrl and url attributes
func TestExtractFeeds_MixedURLAttributes(t *testing.T) {
	t.Parallel()
	opmlData := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <body>
    <outline text="Feed 1" xmlUrl="https://example.com/1"/>
    <outline text="Feed 2" url="https://example.com/2"/>
    <outline text="Feed 3" xmlUrl="https://example.com/3" url="https://ignored.com"/>
  </body>
</opml>`

	opml, err := Parse([]byte(opmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	feeds := opml.ExtractFeeds()
	if len(feeds) != 3 {
		t.Fatalf("Expected 3 feeds, got %d", len(feeds))
	}

	// Feed 1 should use xmlUrl
	if feeds[0].FeedURL != "https://example.com/1" {
		t.Errorf("Feed 1: expected URL from xmlUrl, got %s", feeds[0].FeedURL)
	}

	// Feed 2 should use url
	if feeds[1].FeedURL != "https://example.com/2" {
		t.Errorf("Feed 2: expected URL from url, got %s", feeds[1].FeedURL)
	}

	// Feed 3 should prefer xmlUrl over url
	if feeds[2].FeedURL != "https://example.com/3" {
		t.Errorf("Feed 3: expected xmlUrl to take precedence, got %s", feeds[2].FeedURL)
	}
}

// Test feed type preservation
func TestGenerate_FeedType(t *testing.T) {
	t.Parallel()
	feeds := []Feed{{Title: "Test", FeedURL: "https://example.com/feed"}}
	opml, _ := Generate(feeds, Metadata{Title: "Test"})

	if opml.Body.Outlines[0].Type != "rss" {
		t.Errorf("Expected type 'rss', got %s", opml.Body.Outlines[0].Type)
	}
}

// Test CDATA handling
func TestParse_CDATA(t *testing.T) {
	t.Parallel()
	opmlData := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <body>
    <outline text="Feed &lt;![CDATA[Title]]&gt;" xmlUrl="https://example.com/feed"/>
  </body>
</opml>`

	opml, err := Parse([]byte(opmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	feeds := opml.ExtractFeeds()
	if len(feeds) != 1 {
		t.Fatalf("Expected 1 feed, got %d", len(feeds))
	}
}
