package main

import (
	"context"
	"fmt"
)

func cmdUpdate(ctx context.Context, opts UpdateOptions) error {
	setVerboseLogging(opts.Verbose)

	// Load config
	cfg, err := loadConfig(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Fetch feeds
	fmt.Fprintln(opts.Output, "Fetching feeds...")
	if err := fetchFeeds(ctx, cfg, opts.Logger, opts.Output); err != nil {
		return fmt.Errorf("failed to fetch feeds: %w", err)
	}

	// Generate site
	fmt.Fprintln(opts.Output, "Generating site...")
	if err := generateSite(ctx, cfg, opts.Output); err != nil {
		return fmt.Errorf("failed to generate site: %w", err)
	}

	fmt.Fprintln(opts.Output, "✓ Update complete")
	return nil
}
