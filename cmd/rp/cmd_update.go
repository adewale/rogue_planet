package main

import "fmt"

func cmdUpdate(opts UpdateOptions) error {
	setVerboseLogging(opts.Verbose)

	// Load config
	cfg, err := loadConfig(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Fetch feeds
	fmt.Fprintln(opts.Output, "Fetching feeds...")
	if err := fetchFeeds(cfg, opts.Logger); err != nil {
		return fmt.Errorf("failed to fetch feeds: %w", err)
	}

	// Generate site
	fmt.Fprintln(opts.Output, "Generating site...")
	if err := generateSite(cfg); err != nil {
		return fmt.Errorf("failed to generate site: %w", err)
	}

	fmt.Fprintln(opts.Output, "✓ Update complete")
	return nil
}
