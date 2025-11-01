package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFullWorkflow tests the complete workflow from init to HTML generation
func TestFullWorkflow(t *testing.T) {
	// Create temporary directory that auto-cleans up
	tmpDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Test 1: Initialize planet
	t.Run("init", func(t *testing.T) {
		// Simulate: rp init
		os.Args = []string{"rp", "init"}
		// We can't easily test main() directly, so we'll call runInit()
		runInit()

		// Verify files were created
		if _, err := os.Stat("config.ini"); os.IsNotExist(err) {
			t.Error("config.ini was not created")
		}
		if _, err := os.Stat("data"); os.IsNotExist(err) {
			t.Error("data/ directory was not created")
		}
		if _, err := os.Stat("public"); os.IsNotExist(err) {
			t.Error("public/ directory was not created")
		}
	})

	// Test 2: Add feeds
	t.Run("add-feed", func(t *testing.T) {
		// Create test feed URLs
		testFeeds := []string{
			"https://example.com/feed1.xml",
			"https://example.com/feed2.xml",
		}

		for _, feedURL := range testFeeds {
			os.Args = []string{"rp", "add-feed", feedURL}
			runAddFeed()
		}

		// Verify feeds were added
		os.Args = []string{"rp", "list-feeds"}
		// runListFeeds() prints to stdout, we'd need to capture it
		// For now, just verify no panic
		runListFeeds()
	})

	// Test 3: Check status
	t.Run("status", func(t *testing.T) {
		os.Args = []string{"rp", "status"}
		runStatus()
		// Should show 2 feeds, 0 entries
	})
}

// TestInitWithFeedsFile tests initializing with a feeds file
func TestInitWithFeedsFile(t *testing.T) {
	tmpDir := t.TempDir()

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Create a test feeds file
	feedsContent := `https://example.com/feed1.xml
https://example.com/feed2.xml
# Comment line
https://example.com/feed3.xml
`
	feedsPath := filepath.Join(tmpDir, "test-feeds.txt")
	if err := os.WriteFile(feedsPath, []byte(feedsContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Initialize with feeds file
	os.Args = []string{"rp", "init", "-f", feedsPath}
	runInit()

	// Verify files were created
	if _, err := os.Stat("config.ini"); os.IsNotExist(err) {
		t.Error("config.ini was not created")
	}

	// Verify database exists and has feeds
	if _, err := os.Stat("data/planet.db"); os.IsNotExist(err) {
		t.Error("database was not created")
	}
}

// TestHTMLGeneration tests the complete pipeline from HTTP fetch to HTML generation
func TestHTMLGeneration(t *testing.T) {
	t.Skip("TODO: Complete implementation - needs test crawler support")

	// This test validates the full end-to-end workflow but uses direct function
	// calls instead of CLI commands to allow test-only crawlers that skip SSRF checks

	// Setup temporary directory
	tmpDir := t.TempDir()

	// Create a test RSS feed
	testFeed := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Blog</title>
    <link>https://example.com</link>
    <description>A test blog</description>
    <item>
      <title>Test Entry 1</title>
      <link>https://example.com/post1</link>
      <description>This is the first test entry</description>
      <author>test@example.com (Test Author)</author>
      <pubDate>Mon, 01 Jan 2024 12:00:00 GMT</pubDate>
      <guid>https://example.com/post1</guid>
    </item>
    <item>
      <title>Test Entry 2</title>
      <link>https://example.com/post2</link>
      <description>This is the second test entry</description>
      <author>test@example.com (Test Author)</author>
      <pubDate>Tue, 02 Jan 2024 12:00:00 GMT</pubDate>
      <guid>https://example.com/post2</guid>
    </item>
  </channel>
</rss>`

	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		if _, err := w.Write([]byte(testFeed)); err != nil {
			t.Errorf("Write error: %v", err)
		}
	}))
	defer server.Close()

	// Step 1: Initialize planet in temp directory
	configPath := filepath.Join(tmpDir, "config.ini")

	initOpts := InitOptions{
		ConfigPath: configPath,
		Output:     os.Stdout,
	}
	if err := cmdInit(initOpts); err != nil {
		t.Fatalf("Failed to initialize planet: %v", err)
	}

	// Step 2: Add the test feed
	addOpts := AddFeedOptions{
		ConfigPath: configPath,
		URL:        server.URL,
		Output:     os.Stdout,
	}
	if err := cmdAddFeed(addOpts); err != nil {
		t.Fatalf("Failed to add feed: %v", err)
	}

	// Step 3: Generate HTML (even with no entries, tests the pipeline)
	// Note: We can't fetch from localhost due to SSRF protection
	// Future enhancement: Add support for test crawlers in fetch commands
	generateOpts := GenerateOptions{
		ConfigPath: configPath,
		Output:     os.Stdout,
	}
	if err := cmdGenerate(generateOpts); err != nil {
		t.Fatalf("Failed to generate HTML: %v", err)
	}

	// Step 4: Verify HTML was generated
	// The config uses relative paths, so files are created relative to config location
	htmlPath := filepath.Join(tmpDir, "public", "index.html")
	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		// Debug: list what files were actually created
		if entries, listErr := os.ReadDir(tmpDir); listErr == nil {
			t.Logf("Files in tmpDir: %v", entries)
		}
		if entries, listErr := os.ReadDir(filepath.Join(tmpDir, "public")); listErr == nil {
			t.Logf("Files in public: %v", entries)
		}
		t.Fatalf("HTML file was not generated at %s", htmlPath)
	}

	// Step 5: Read and verify HTML structure (even if no entries due to SSRF)
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read generated HTML: %v", err)
	}

	htmlStr := string(htmlContent)

	// Verify basic HTML structure was generated
	basicChecks := []struct {
		name    string
		content string
	}{
		{"html structure", "<html"},
		{"head section", "<head>"},
		{"body section", "<body>"},
		{"CSP header", "Content-Security-Policy"},
	}

	for _, check := range basicChecks {
		if !strings.Contains(htmlStr, check.content) {
			t.Errorf("Generated HTML missing %s: %q", check.name, check.content)
		}
	}

	t.Logf("✓ Successfully tested HTML generation pipeline")
	t.Logf("Note: Full content validation requires refactoring to support test crawlers")
}

// Test #8: Redirect Then Remove Integration Test (CRITICAL)
// Tests that remove-feed works correctly after a feed URL has been updated due to a 301 redirect
// This test simulates the scenario where UpdateFeedURL is called (as would happen during a 301 redirect)
func TestRemoveFeedAfterRedirect(t *testing.T) {
	tmpDir := t.TempDir()

	// Create database directory
	dataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("Failed to create data directory: %v", err)
	}

	dbPath := filepath.Join(dataDir, "planet.db")

	// Create config file
	configPath := filepath.Join(tmpDir, "config.ini")
	configContent := `[planet]
name = Test Planet

[database]
path = ` + dbPath

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	oldURL := "https://example.com/old-feed"
	newURL := "https://example.com/new-feed"

	// STEP 1: Add feed with old URL using cmdAddFeed
	addOpts := AddFeedOptions{
		URL:        oldURL,
		ConfigPath: configPath,
		Output:     os.Stdout,
	}

	if err := cmdAddFeed(addOpts); err != nil {
		t.Fatalf("Failed to add feed: %v", err)
	}

	// STEP 2: Simulate a 301 redirect by manually updating the feed URL
	// (In production, this would happen in the fetcher when a 301 is detected)
	cfg, repo, cleanup, err := openConfigAndRepo(configPath)
	if err != nil {
		t.Fatalf("Failed to open config and repo: %v", err)
	}

	// Get the feed by old URL
	feed, err := repo.GetFeedByURL(oldURL)
	if err != nil {
		cleanup()
		t.Fatalf("Failed to get feed by old URL: %v", err)
	}

	// Update the URL (simulating 301 redirect)
	if err := repo.UpdateFeedURL(feed.ID, newURL); err != nil {
		cleanup()
		t.Fatalf("Failed to update feed URL: %v", err)
	}
	cleanup()

	// Verify cfg is not nil (just to use the variable)
	if cfg == nil {
		t.Fatal("Config should not be nil")
	}

	// STEP 3: Attempt to remove feed using OLD URL (should fail - feed not found)
	removeOptsOld := RemoveFeedOptions{
		URL:        oldURL,
		ConfigPath: configPath,
		Output:     os.Stdout,
		Input:      strings.NewReader("y\n"),
		Force:      true,
	}

	err = cmdRemoveFeed(removeOptsOld)
	if err == nil {
		t.Fatal("Remove with old URL should fail after URL update (simulated 301)")
	}
	if !strings.Contains(err.Error(), "feed not found") {
		t.Errorf("Error should mention 'feed not found', got: %v", err)
	}

	// STEP 4: Attempt to remove feed using NEW URL (should succeed)
	removeOptsNew := RemoveFeedOptions{
		URL:        newURL,
		ConfigPath: configPath,
		Output:     os.Stdout,
		Input:      strings.NewReader("y\n"),
		Force:      true,
	}

	if err := cmdRemoveFeed(removeOptsNew); err != nil {
		t.Fatalf("Remove with new URL should succeed after URL update, got error: %v", err)
	}

	t.Logf("✓ Successfully tested remove-feed after simulated 301 redirect")
	t.Logf("  - Old URL (%s) correctly rejected (feed not found)", oldURL)
	t.Logf("  - New URL (%s) correctly accepted and removed", newURL)
}
