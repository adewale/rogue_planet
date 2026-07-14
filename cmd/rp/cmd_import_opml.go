package main

import (
	"context"
	"fmt"

	"github.com/adewale/rogue_planet/pkg/crawler"
	"github.com/adewale/rogue_planet/pkg/opml"
)

func cmdImportOPML(opts ImportOPMLOptions) error {
	if opts.OPMLFile == "" {
		return fmt.Errorf("OPML file is required")
	}

	// Parse OPML file
	opmlDoc, err := opml.ParseFile(context.Background(), opts.OPMLFile)
	if err != nil {
		return fmt.Errorf("failed to parse OPML file: %w", err)
	}

	// Extract feeds
	feeds := opmlDoc.ExtractFeeds()

	if len(feeds) == 0 {
		_, _ = fmt.Fprintln(opts.Output, "No feeds found in OPML file")
		return nil
	}

	ctx := context.Background()

	if opts.DryRun {
		_, _ = fmt.Fprintf(opts.Output, "DRY RUN: Importing feeds from %s...\n\n", opts.OPMLFile)
		_, _ = fmt.Fprintf(opts.Output, "Found %d feeds in OPML file\n\n", len(feeds))

		// Load config and database to check for duplicates
		_, repo, cleanup, err := openConfigAndRepo(opts.ConfigPath)
		if err != nil {
			// Database might not exist yet, just show what would be imported
			for i, feed := range feeds {
				_, _ = fmt.Fprintf(opts.Output, "  [%d/%d] Would add: %s (%s)\n", i+1, len(feeds), feed.FeedURL, feed.Title)
			}
			_, _ = fmt.Fprintf(opts.Output, "\nDRY RUN: Would import %d feeds\n", len(feeds))
			return nil
		}
		defer cleanup()

		// Check which feeds already exist
		skipCount := 0
		for i, feed := range feeds {
			_, err := repo.GetFeedByURL(ctx, feed.FeedURL)
			if err == nil {
				_, _ = fmt.Fprintf(opts.Output, "  [%d/%d] Would skip: %s (already exists)\n", i+1, len(feeds), feed.FeedURL)
				skipCount++
			} else {
				_, _ = fmt.Fprintf(opts.Output, "  [%d/%d] Would add: %s (%s)\n", i+1, len(feeds), feed.FeedURL, feed.Title)
			}
		}

		_, _ = fmt.Fprintf(opts.Output, "\nDRY RUN: Would import %d/%d feeds (%d duplicates skipped)\n", len(feeds)-skipCount, len(feeds), skipCount)
		return nil
	}

	// Real import
	_, _ = fmt.Fprintf(opts.Output, "Importing feeds from %s...\n\n", opts.OPMLFile)
	_, _ = fmt.Fprintf(opts.Output, "Found %d feeds in OPML file\n\n", len(feeds))

	_, repo, cleanup, err := openConfigAndRepo(opts.ConfigPath)
	if err != nil {
		return err
	}
	defer cleanup()

	// Import each feed
	addedCount := 0
	skippedCount := 0

	for i, feed := range feeds {
		// Check if feed already exists
		_, err := repo.GetFeedByURL(ctx, feed.FeedURL)
		if err == nil {
			_, _ = fmt.Fprintf(opts.Output, "  [%d/%d] %s\n", i+1, len(feeds), feed.FeedURL)
			_, _ = fmt.Fprintln(opts.Output, "         ⚠ Skipped (already exists)")
			skippedCount++
			continue
		}

		// Validate URL
		if err := crawler.ValidateURL(feed.FeedURL); err != nil {
			_, _ = fmt.Fprintf(opts.Output, "  [%d/%d] %s\n", i+1, len(feeds), feed.FeedURL)
			_, _ = fmt.Fprintf(opts.Output, "         ✗ Skipped (invalid URL: %v)\n", err)
			skippedCount++
			continue
		}

		// Add feed
		title := feed.Title
		if title == "" {
			title = feed.FeedURL
		}

		id, err := repo.AddFeed(ctx, feed.FeedURL, title)
		if err != nil {
			_, _ = fmt.Fprintf(opts.Output, "  [%d/%d] %s\n", i+1, len(feeds), feed.FeedURL)
			_, _ = fmt.Fprintf(opts.Output, "         ✗ Failed: %v\n", err)
			skippedCount++
			continue
		}

		_, _ = fmt.Fprintf(opts.Output, "  [%d/%d] Adding %s (%s)\n", i+1, len(feeds), feed.FeedURL, title)
		_, _ = fmt.Fprintf(opts.Output, "         ✓ Added (ID: %d)\n", id)
		addedCount++
	}

	_, _ = fmt.Fprintf(opts.Output, "\n✓ Successfully imported %d/%d feeds\n", addedCount, len(feeds))
	_, _ = fmt.Fprintf(opts.Output, "  - %d added\n", addedCount)
	_, _ = fmt.Fprintf(opts.Output, "  - %d skipped (duplicates or invalid)\n", skippedCount)

	return nil
}
