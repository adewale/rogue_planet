package main

import (
	"context"
	"fmt"
)

func cmdStatus(opts StatusOptions) error {
	cfg, repo, cleanup, err := openConfigAndRepo(opts.ConfigPath)
	if err != nil {
		return err
	}
	defer cleanup()

	ctx := context.Background()

	// Get feed counts
	feeds, err := repo.GetFeeds(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get feeds: %w", err)
	}

	activeFeeds := 0
	for _, feed := range feeds {
		if feed.Active {
			activeFeeds++
		}
	}

	// Get entry count
	totalEntries, err := repo.CountEntries(ctx)
	if err != nil {
		return fmt.Errorf("failed to count entries: %w", err)
	}

	// Get recent entry count (based on config days)
	recentEntries, err := repo.CountRecentEntries(ctx, cfg.Planet.Days)
	if err != nil {
		return fmt.Errorf("failed to count recent entries: %w", err)
	}

	// Display status
	fmt.Fprintln(opts.Output, "Rogue Planet Status")
	fmt.Fprintln(opts.Output, "===================")
	fmt.Fprintln(opts.Output)
	fmt.Fprintf(opts.Output, "Feeds:           %d total (%d active, %d inactive)\n", len(feeds), activeFeeds, len(feeds)-activeFeeds)
	fmt.Fprintf(opts.Output, "Entries:         %d total\n", totalEntries)
	fmt.Fprintf(opts.Output, "Recent entries:  %d (last %d days)\n", recentEntries, cfg.Planet.Days)
	fmt.Fprintln(opts.Output)
	fmt.Fprintf(opts.Output, "Output:          %s/index.html\n", cfg.Planet.OutputDir)
	fmt.Fprintf(opts.Output, "Database:        %s\n", cfg.Database.Path)

	return nil
}
