package main

import "fmt"

func cmdFetch(opts FetchOptions) error {
	setVerboseLogging(opts.Verbose)

	cfg, err := loadConfig(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Fprintln(opts.Output, "Fetching feeds...")
	if err := fetchFeeds(cfg, opts.Logger); err != nil {
		return fmt.Errorf("failed to fetch feeds: %w", err)
	}

	fmt.Fprintln(opts.Output, "✓ Fetch complete")
	return nil
}
