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
	_, _ = fmt.Fprintln(opts.Output, "Fetching feeds...")
	if err := fetchFeeds(ctx, cfg, opts.Logger); err != nil {
		return fmt.Errorf("failed to fetch feeds: %w", err)
	}

	// Generate site
	_, _ = fmt.Fprintln(opts.Output, "Generating site...")
	if err := generateSite(ctx, cfg); err != nil {
		return fmt.Errorf("failed to generate site: %w", err)
	}

	_, _ = fmt.Fprintln(opts.Output, "✓ Update complete")
	return nil
}
