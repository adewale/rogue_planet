package main

import (
	"fmt"
	"os"

	"github.com/adewale/rogue_planet/pkg/opml"
)

func cmdExportOPML(opts ExportOPMLOptions) error {
	// Load config
	cfg, repo, cleanup, err := openConfigAndRepo(opts.ConfigPath)
	if err != nil {
		return err
	}
	defer cleanup()

	// Get all feeds
	repoFeeds, err := repo.GetFeeds(false)
	if err != nil {
		return fmt.Errorf("failed to get feeds: %w", err)
	}

	if len(repoFeeds) == 0 {
		fmt.Fprintln(opts.Output, "No feeds to export")
		return nil
	}

	// Convert to OPML feeds
	opmlFeeds := make([]opml.Feed, 0, len(repoFeeds))
	for _, feed := range repoFeeds {
		title := feed.Title
		if title == "" {
			title = feed.URL
		}

		opmlFeeds = append(opmlFeeds, opml.Feed{
			Title:   title,
			FeedURL: feed.URL,
			WebURL:  feed.Link,
		})
	}

	// Generate OPML
	metadata := opml.Metadata{
		Title:      cfg.Planet.Name + " Feed List",
		OwnerName:  cfg.Planet.OwnerName,
		OwnerEmail: cfg.Planet.OwnerEmail,
	}

	opmlDoc, err := opml.Generate(opmlFeeds, metadata)
	if err != nil {
		return fmt.Errorf("failed to generate OPML: %w", err)
	}

	// Marshal to XML
	xmlData, err := opmlDoc.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal OPML: %w", err)
	}

	// Write to file or stdout
	if opts.OutputFile != "" {
		if err := os.WriteFile(opts.OutputFile, xmlData, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		fmt.Fprintf(opts.Output, "✓ Exported %d feeds to %s\n", len(opmlFeeds), opts.OutputFile)
	} else {
		fmt.Fprint(opts.Output, string(xmlData))
	}

	return nil
}
