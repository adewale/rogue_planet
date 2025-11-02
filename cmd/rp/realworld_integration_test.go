package main

import (
	"context"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adewale/rogue_planet/pkg/generator"
	"github.com/adewale/rogue_planet/pkg/normalizer"
	"github.com/adewale/rogue_planet/pkg/repository"
	"golang.org/x/net/html"
)

// TestRealWorldFeedsFullPipeline tests the complete pipeline with real feed snapshots:
// Parse -> Store -> Retrieve -> Generate HTML -> Verify
func TestRealWorldFeedsFullPipeline(t *testing.T) {
	tests := []struct {
		name          string
		feedPath      string
		feedURL       string
		expectedTitle string
		minEntries    int // Minimum entries expected to parse from feed
	}{
		{
			name:          "Daring Fireball",
			feedPath:      "../../testdata/daringfireball-feed.xml",
			feedURL:       "https://daringfireball.net/feeds/main",
			expectedTitle: "Daring Fireball",
			minEntries:    10, // Should parse at least 10 entries from snapshot
		},
		{
			name:          "Asymco",
			feedPath:      "../../testdata/asymco-feed.xml",
			feedURL:       "https://asymco.com/feed/",
			expectedTitle: "Asymco",
			minEntries:    5, // Should parse at least 5 entries from snapshot
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tmpDir := t.TempDir()
			dbPath := filepath.Join(tmpDir, "test.db")

			// Create repository
			repo, err := repository.New(dbPath)
			if err != nil {
				t.Fatalf("Failed to create repository: %v", err)
			}
			defer repo.Close()

			// Add feed
			ctx := context.Background()

			feedID, err := repo.AddFeed(ctx, tt.feedURL, "")
			if err != nil {
				t.Fatalf("Failed to add feed: %v", err)
			}

			// Read and parse feed
			data, err := os.ReadFile(tt.feedPath)
			if err != nil {
				t.Fatalf("Failed to read feed file: %v", err)
			}

			n := normalizer.New()
			metadata, entries, err := n.Parse(ctx, data, tt.feedURL, time.Now())
			if err != nil {
				t.Fatalf("Failed to parse feed: %v", err)
			}

			// Verify metadata
			if metadata.Title != tt.expectedTitle {
				t.Errorf("Feed title = %q, want %q", metadata.Title, tt.expectedTitle)
			}

			// Verify we parsed some entries
			if len(entries) < tt.minEntries {
				t.Errorf("Parsed %d entries, want at least %d", len(entries), tt.minEntries)
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

				if err := repo.UpsertEntry(ctx, repoEntry); err != nil {
					t.Fatalf("Failed to store entry: %v", err)
				}
			}

			t.Logf("Stored %d entries", len(entries))

			// Retrieve entries from database
			// Note: GetRecentEntries() has smart fallback - if no recent entries,
			// it returns the 50 most recent regardless of date. This is time-invariant.
			dbEntries, err := repo.GetRecentEntries(ctx, 7)
			if err != nil {
				t.Fatalf("Failed to get entries: %v", err)
			}

			// Verify we got SOME entries back (either recent or fallback)
			if len(dbEntries) == 0 {
				t.Error("GetRecentEntries returned no entries (should use fallback)")
			}

			// Generate HTML
			gen, err := generator.New()
			if err != nil {
				t.Fatalf("Failed to create generator: %v", err)
			}

			genEntries := make([]generator.EntryData, 0, len(dbEntries))
			for _, entry := range dbEntries {
				genEntries = append(genEntries, generator.EntryData{
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

			outputPath := filepath.Join(tmpDir, "index.html")
			data2 := generator.TemplateData{
				Title:       "Test Planet",
				Link:        "https://planet.example.com",
				OwnerName:   "Test Owner",
				OwnerEmail:  "test@example.com",
				Entries:     genEntries,
				GroupByDate: true,
			}

			if err := gen.GenerateToFile(ctx, outputPath, data2); err != nil {
				t.Fatalf("Failed to generate HTML: %v", err)
			}

			// Verify HTML
			doc, err := parseHTMLFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			// Verify feed title appears
			if !containsText(doc, tt.expectedTitle) {
				t.Errorf("HTML does not contain feed title %q", tt.expectedTitle)
			}

			// Verify first entry title appears
			if len(entries) > 0 && !containsText(doc, entries[0].Title) {
				t.Errorf("HTML does not contain first entry title %q", entries[0].Title)
			}

			// Verify no dangerous HTML passed through
			htmlContent, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read HTML file: %v", err)
			}

			dangerous := []string{"<script", "javascript:", "onerror="}
			for _, d := range dangerous {
				if strings.Contains(strings.ToLower(string(htmlContent)), d) {
					t.Errorf("HTML contains dangerous content: %s", d)
				}
			}

			t.Logf("✓ Successfully processed %s: %d entries -> database -> HTML", tt.name, len(entries))
		})
	}
}

// Helper functions

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
