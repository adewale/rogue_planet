package main

import (
	"fmt"

	"github.com/adewale/rogue_planet/pkg/config"
)

func cmdAddAll(opts AddAllOptions) error {
	if opts.FeedsFile == "" {
		return fmt.Errorf("feeds file is required")
	}

	_, repo, cleanup, err := openConfigAndRepo(opts.ConfigPath)
	if err != nil {
		return err
	}
	defer cleanup()

	// Load feeds from file
	feedURLs, err := config.LoadFeedsFile(opts.FeedsFile)
	if err != nil {
		return fmt.Errorf("failed to load feeds file: %w", err)
	}

	if len(feedURLs) == 0 {
		fmt.Fprintln(opts.Output, "No feeds found in file")
		return nil
	}

	fmt.Fprintf(opts.Output, "Adding %d feeds from %s...\n", len(feedURLs), opts.FeedsFile)

	// Add each feed to database
	addedCount := importFeedsFromURLs(repo, feedURLs, opts.Output)

	fmt.Fprintf(opts.Output, "\n✓ Added %d/%d feeds\n", addedCount, len(feedURLs))
	return nil
}
