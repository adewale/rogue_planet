package main

import (
	"context"
	"fmt"
)

func cmdAddFeed(ctx context.Context, opts AddFeedOptions) error {
	if opts.URL == "" {
		return fmt.Errorf("URL is required")
	}

	_, repo, cleanup, err := openConfigAndRepo(opts.ConfigPath)
	if err != nil {
		return err
	}
	defer cleanup()

	// Add feed
	id, err := repo.AddFeed(ctx, opts.URL, "")
	if err != nil {
		return fmt.Errorf("failed to add feed: %w", err)
	}

	fmt.Fprintf(opts.Output, "✓ Added feed: %s (ID: %d)\n", opts.URL, id)
	return nil
}
