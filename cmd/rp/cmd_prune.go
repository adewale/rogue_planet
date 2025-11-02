package main

import (
	"context"
	"fmt"
)

func cmdPrune(ctx context.Context, opts PruneOptions) error {
	_, repo, cleanup, err := openConfigAndRepo(opts.ConfigPath)
	if err != nil {
		return err
	}
	defer cleanup()

	if opts.DryRun {
		fmt.Fprintf(opts.Output, "Dry run: would delete entries older than %d days\n", opts.Days)
		// In a real implementation, we'd query and show what would be deleted
		return nil
	}

	deleted, err := repo.PruneOldEntries(ctx, opts.Days)
	if err != nil {
		return fmt.Errorf("failed to prune entries: %w", err)
	}

	fmt.Fprintf(opts.Output, "✓ Deleted %d old entries\n", deleted)
	return nil
}
