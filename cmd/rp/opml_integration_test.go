package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adewale/rogue_planet/pkg/opml"
)

func TestOPMLRoundTrip(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	configPath := filepath.Join(dir, "config.ini")

	// Initialize
	initOpts := InitOptions{
		FeedsFile:  "",
		ConfigPath: configPath,
		Output:     &bytes.Buffer{},
	}
	if err := cmdInit(initOpts); err != nil {
		t.Fatalf("cmdInit failed: %v", err)
	}

	// Add some feeds manually
	feeds := []struct {
		url   string
		title string
	}{
		{"https://go.dev/blog/feed.atom", "Go Blog"},
		{"https://blog.golang.org/feed.atom", "Golang Blog"},
		{"https://example.com/feed.xml", "Example Feed"},
	}

	for _, feed := range feeds {
		addOpts := AddFeedOptions{
			URL:        feed.url,
			ConfigPath: configPath,
			Output:     &bytes.Buffer{},
		}
		if err := cmdAddFeed(addOpts); err != nil {
			t.Fatalf("cmdAddFeed failed for %s: %v", feed.url, err)
		}
	}

	// Export to OPML
	exportPath := filepath.Join(dir, "export.opml")
	exportOpts := ExportOPMLOptions{
		ConfigPath: configPath,
		OutputFile: exportPath,
		Output:     &bytes.Buffer{},
	}
	if err := cmdExportOPML(exportOpts); err != nil {
		t.Fatalf("cmdExportOPML failed: %v", err)
	}

	// Verify OPML file was created
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Fatal("Export OPML file should exist")
	}

	// Parse exported OPML
	opmlDoc, err := opml.ParseFile(context.Background(), exportPath)
	if err != nil {
		t.Fatalf("Failed to parse exported OPML: %v", err)
	}

	exportedFeeds := opmlDoc.ExtractFeeds()
	if len(exportedFeeds) != len(feeds) {
		t.Errorf("Expected %d feeds in export, got %d", len(feeds), len(exportedFeeds))
	}

	// Create a new database for import test
	dir2 := t.TempDir()
	configPath2 := filepath.Join(dir2, "config.ini")

	initOpts2 := InitOptions{
		FeedsFile:  "",
		ConfigPath: configPath2,
		Output:     &bytes.Buffer{},
	}
	if err := cmdInit(initOpts2); err != nil {
		t.Fatalf("cmdInit failed for import test: %v", err)
	}

	// Import the exported OPML
	importOpts := ImportOPMLOptions{
		OPMLFile:   exportPath,
		ConfigPath: configPath2,
		DryRun:     false,
		Output:     &bytes.Buffer{},
	}
	if err := cmdImportOPML(importOpts); err != nil {
		t.Fatalf("cmdImportOPML failed: %v", err)
	}

	// List feeds from new database
	var listBuf bytes.Buffer
	listOpts := ListFeedsOptions{
		ConfigPath: configPath2,
		Output:     &listBuf,
	}
	if err := cmdListFeeds(listOpts); err != nil {
		t.Fatalf("cmdListFeeds failed: %v", err)
	}

	listOutput := listBuf.String()

	// Verify all feeds were imported
	for _, feed := range feeds {
		if !strings.Contains(listOutput, feed.url) {
			t.Errorf("Imported feeds should contain %s", feed.url)
		}
	}
}

func TestOPMLImportRealWorldFile(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	configPath := filepath.Join(dir, "config.ini")

	// Initialize
	initOpts := InitOptions{
		FeedsFile:  "",
		ConfigPath: configPath,
		Output:     &bytes.Buffer{},
	}
	if err := cmdInit(initOpts); err != nil {
		t.Fatalf("cmdInit failed: %v", err)
	}

	// Create a realistic OPML file (similar to Feedly export)
	opmlPath := filepath.Join(dir, "feedly-export.opml")
	opmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head>
    <title>Feedly Export</title>
    <dateCreated>Mon, 12 Oct 2025 10:00:00 +0000</dateCreated>
  </head>
  <body>
    <outline text="Tech" title="Tech">
      <outline text="Hacker News" title="Hacker News" type="rss"
               xmlUrl="https://news.ycombinator.com/rss"
               htmlUrl="https://news.ycombinator.com/"/>
      <outline text="Ars Technica" title="Ars Technica" type="rss"
               xmlUrl="http://feeds.arstechnica.com/arstechnica/index"
               htmlUrl="https://arstechnica.com"/>
    </outline>
    <outline text="Design" title="Design">
      <outline text="A List Apart" title="A List Apart" type="rss"
               xmlUrl="https://alistapart.com/main/feed/"
               htmlUrl="https://alistapart.com"/>
    </outline>
  </body>
</opml>`

	if err := os.WriteFile(opmlPath, []byte(opmlContent), 0644); err != nil {
		t.Fatalf("Failed to write OPML file: %v", err)
	}

	// Import OPML
	var importBuf bytes.Buffer
	importOpts := ImportOPMLOptions{
		OPMLFile:   opmlPath,
		ConfigPath: configPath,
		DryRun:     false,
		Output:     &importBuf,
	}
	if err := cmdImportOPML(importOpts); err != nil {
		t.Fatalf("cmdImportOPML failed: %v", err)
	}

	importOutput := importBuf.String()

	// Should have imported 3 feeds (from nested structure)
	if !strings.Contains(importOutput, "3/3 feeds") {
		t.Errorf("Should import 3 feeds, got output: %s", importOutput)
	}

	// List feeds to verify
	var listBuf bytes.Buffer
	listOpts := ListFeedsOptions{
		ConfigPath: configPath,
		Output:     &listBuf,
	}
	if err := cmdListFeeds(listOpts); err != nil {
		t.Fatalf("cmdListFeeds failed: %v", err)
	}

	listOutput := listBuf.String()

	expectedFeeds := []string{
		"news.ycombinator.com",
		"arstechnica.com",
		"alistapart.com",
	}

	for _, expected := range expectedFeeds {
		if !strings.Contains(listOutput, expected) {
			t.Errorf("Should contain feed from %s", expected)
		}
	}
}

func TestOPMLImportDryRun(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	configPath := filepath.Join(dir, "config.ini")

	// Initialize
	initOpts := InitOptions{
		FeedsFile:  "",
		ConfigPath: configPath,
		Output:     &bytes.Buffer{},
	}
	if err := cmdInit(initOpts); err != nil {
		t.Fatalf("cmdInit failed: %v", err)
	}

	// Create OPML file
	opmlPath := filepath.Join(dir, "test.opml")
	opmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head>
    <title>Test</title>
  </head>
  <body>
    <outline text="Feed 1" type="rss" xmlUrl="https://example.com/feed1"/>
    <outline text="Feed 2" type="rss" xmlUrl="https://example.com/feed2"/>
  </body>
</opml>`

	if err := os.WriteFile(opmlPath, []byte(opmlContent), 0644); err != nil {
		t.Fatalf("Failed to write OPML file: %v", err)
	}

	// Dry run import
	var dryRunBuf bytes.Buffer
	dryRunOpts := ImportOPMLOptions{
		OPMLFile:   opmlPath,
		ConfigPath: configPath,
		DryRun:     true,
		Output:     &dryRunBuf,
	}
	if err := cmdImportOPML(dryRunOpts); err != nil {
		t.Fatalf("cmdImportOPML dry run failed: %v", err)
	}

	dryRunOutput := dryRunBuf.String()

	// Check for dry run indicators
	if !strings.Contains(dryRunOutput, "DRY RUN") {
		t.Error("Dry run output should contain 'DRY RUN'")
	}

	if !strings.Contains(dryRunOutput, "Would add") {
		t.Error("Dry run output should contain 'Would add'")
	}

	// Verify feeds were NOT actually added
	var listBuf bytes.Buffer
	listOpts := ListFeedsOptions{
		ConfigPath: configPath,
		Output:     &listBuf,
	}
	if err := cmdListFeeds(listOpts); err != nil {
		t.Fatalf("cmdListFeeds failed: %v", err)
	}

	listOutput := listBuf.String()

	if strings.Contains(listOutput, "example.com/feed1") {
		t.Error("Dry run should not actually add feeds")
	}
}

func TestOPMLImportDuplicateDetection(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	configPath := filepath.Join(dir, "config.ini")

	// Initialize
	initOpts := InitOptions{
		FeedsFile:  "",
		ConfigPath: configPath,
		Output:     &bytes.Buffer{},
	}
	if err := cmdInit(initOpts); err != nil {
		t.Fatalf("cmdInit failed: %v", err)
	}

	// Add a feed manually
	addOpts := AddFeedOptions{
		URL:        "https://example.com/feed",
		ConfigPath: configPath,
		Output:     &bytes.Buffer{},
	}
	if err := cmdAddFeed(addOpts); err != nil {
		t.Fatalf("cmdAddFeed failed: %v", err)
	}

	// Create OPML with duplicate and new feed
	opmlPath := filepath.Join(dir, "test.opml")
	opmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head>
    <title>Test</title>
  </head>
  <body>
    <outline text="Existing Feed" type="rss" xmlUrl="https://example.com/feed"/>
    <outline text="New Feed" type="rss" xmlUrl="https://example.org/feed"/>
  </body>
</opml>`

	if err := os.WriteFile(opmlPath, []byte(opmlContent), 0644); err != nil {
		t.Fatalf("Failed to write OPML file: %v", err)
	}

	// Import OPML
	var importBuf bytes.Buffer
	importOpts := ImportOPMLOptions{
		OPMLFile:   opmlPath,
		ConfigPath: configPath,
		DryRun:     false,
		Output:     &importBuf,
	}
	if err := cmdImportOPML(importOpts); err != nil {
		t.Fatalf("cmdImportOPML failed: %v", err)
	}

	importOutput := importBuf.String()

	// Should have added only 1 new feed, skipped 1 duplicate
	if !strings.Contains(importOutput, "1 added") {
		t.Errorf("Should add 1 feed, got output: %s", importOutput)
	}

	if !strings.Contains(importOutput, "1 skipped") {
		t.Errorf("Should skip 1 duplicate, got output: %s", importOutput)
	}

	if !strings.Contains(importOutput, "Skipped") && !strings.Contains(importOutput, "already exists") {
		t.Errorf("Should mention duplicate/already exists: %s", importOutput)
	}
}

func TestOPMLRFC822DateHandling(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Test various RFC 822 date formats
	dateFormats := []string{
		"Mon, 12 Oct 2025 10:00:00 +0000",
		"Mon, 12 Oct 2025 10:00:00 GMT",
		"12 Oct 2025 10:00:00 -0700",
	}

	for i, dateStr := range dateFormats {
		opmlPath := filepath.Join(dir, "test-"+string(rune(i+48))+".opml")
		opmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head>
    <title>Test</title>
    <dateCreated>` + dateStr + `</dateCreated>
  </head>
  <body>
    <outline text="Test Feed" type="rss" xmlUrl="https://example.com/feed"/>
  </body>
</opml>`

		if err := os.WriteFile(opmlPath, []byte(opmlContent), 0644); err != nil {
			t.Fatalf("Failed to write OPML file: %v", err)
		}

		// Parse OPML
		opmlDoc, err := opml.ParseFile(context.Background(), opmlPath)
		if err != nil {
			t.Fatalf("Failed to parse OPML with date %q: %v", dateStr, err)
		}

		if opmlDoc.Head.DateCreated == "" {
			t.Errorf("DateCreated should be preserved for format: %q", dateStr)
		}
	}
}

func TestOPMLExportToStdout(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	configPath := filepath.Join(dir, "config.ini")

	// Initialize and add a feed
	initOpts := InitOptions{
		FeedsFile:  "",
		ConfigPath: configPath,
		Output:     &bytes.Buffer{},
	}
	if err := cmdInit(initOpts); err != nil {
		t.Fatalf("cmdInit failed: %v", err)
	}

	addOpts := AddFeedOptions{
		URL:        "https://example-export-stdout.com/feed",
		ConfigPath: configPath,
		Output:     &bytes.Buffer{},
	}
	if err := cmdAddFeed(addOpts); err != nil {
		t.Fatalf("cmdAddFeed failed: %v", err)
	}

	// Export to stdout (no OutputFile specified)
	var exportBuf bytes.Buffer
	exportOpts := ExportOPMLOptions{
		ConfigPath: configPath,
		OutputFile: "", // stdout
		Output:     &exportBuf,
	}
	if err := cmdExportOPML(exportOpts); err != nil {
		t.Fatalf("cmdExportOPML failed: %v", err)
	}

	exportOutput := exportBuf.String()

	// Should contain valid OPML
	if !strings.Contains(exportOutput, "<?xml version") {
		t.Error("Export to stdout should contain XML declaration")
	}

	if !strings.Contains(exportOutput, "<opml version=\"2.0\">") {
		t.Error("Export to stdout should contain OPML root element")
	}

	if !strings.Contains(exportOutput, "example-export-stdout.com/feed") {
		t.Error("Export to stdout should contain the feed URL")
	}
}
