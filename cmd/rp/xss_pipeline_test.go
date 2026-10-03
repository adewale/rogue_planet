package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/adewale/rogue_planet/internal/htmlsafety"
	"github.com/adewale/rogue_planet/pkg/crawler"
	"github.com/adewale/rogue_planet/pkg/fetcher"
	"github.com/adewale/rogue_planet/pkg/logging"
	"github.com/adewale/rogue_planet/pkg/normalizer"
	"github.com/adewale/rogue_planet/pkg/repository"
	"golang.org/x/net/html"
)

// TestHostileFeedProducesSafePage drives a hostile feed through the real
// fetch -> normalize -> SQLite -> generate path and inspects the parsed page.
// Every feed field that reaches the page (title, link, content) is attacked,
// because a field that skips sanitization is rendered as template.HTML.
func TestHostileFeedProducesSafePage(t *testing.T) {
	feedXML, err := os.ReadFile("../../testdata/hostile-feed.xml")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write(feedXML)
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "planet.db")
	outputDir := filepath.Join(tmpDir, "public")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(tmpDir, "config.ini")
	config := "[planet]\nname = Test Planet\nlink = https://planet.example.com\noutput_dir = " +
		outputDir + "\ndays = 36500\n\n[database]\npath = " + dbPath + "\n"
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	repo, err := repository.New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	feedURL := srv.URL + "/feed.xml"
	if _, err := repo.AddFeed(ctx, feedURL, ""); err != nil {
		t.Fatal(err)
	}
	feed, err := repo.GetFeedByURL(ctx, feedURL)
	if err != nil {
		t.Fatal(err)
	}
	// NewForTesting only lifts the SSRF block so the loopback server is reachable.
	f := fetcher.New(crawler.NewForTesting(), normalizer.New(), repo, &sync.Mutex{},
		logging.New("error"), 0)
	if res := f.FetchFeed(ctx, *feed); res.Error != nil || res.StoredEntries != 3 {
		t.Fatalf("FetchFeed = %+v, want 3 stored entries", res)
	}
	repo.Close()

	if err := cmdGenerate(ctx, GenerateOptions{ConfigPath: configPath, Output: io.Discard}); err != nil {
		t.Fatalf("cmdGenerate: %v", err)
	}
	page, err := os.ReadFile(filepath.Join(outputDir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	doc, err := html.Parse(bytes.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	body := htmlsafety.Find(doc, "body")
	if body == nil {
		t.Fatal("generated page has no <body>")
	}

	for _, v := range htmlsafety.Violations(body) {
		t.Errorf("unsafe markup in generated page: %s", v)
	}

	// The sanitizer must not throw away the benign parts of the same fields.
	text := htmlsafety.Text(body)
	for _, marker := range []string{
		"Title Marker One", "Title Marker Two", "Title Marker Three",
		"Body Marker One", "Body Marker Two", "Body Marker Three",
	} {
		if !strings.Contains(text, marker) {
			t.Errorf("generated page lost benign text %q", marker)
		}
	}
	if !bytes.Contains(page, []byte(`src="https://img.example.com/a.png"`)) {
		t.Error("generated page lost the safe https image")
	}
}
