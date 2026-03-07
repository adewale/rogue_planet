package main

import (
	"context"
	"fmt"

	"github.com/adewale/rogue_planet/pkg/crawler"
)

func cmdAddFeed(opts AddFeedOptions) error {
	if opts.URL == "" {
		return fmt.Errorf("URL is required")
	}

	// Validate URL for SSRF prevention
	if err := crawler.ValidateURL(opts.URL); err != nil {
		return fmt.Errorf("invalid feed URL: %w", err)
	}

	_, repo, cleanup, err := openConfigAndRepo(opts.ConfigPath)
	if err != nil {
		return err
	}
	defer cleanup()

	ctx := context.Background()

	// Add feed
	id, err := repo.AddFeed(ctx, opts.URL, "")
	if err != nil {
		return fmt.Errorf("failed to add feed: %w", err)
	}

	_, _ = fmt.Fprintf(opts.Output, "✓ Added feed: %s (ID: %d)\n", opts.URL, id)
	return nil
}
