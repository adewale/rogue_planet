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
		// Count entries that would be deleted (total - recent = old)
		totalEntries, err := repo.CountEntries(ctx)
		if err != nil {
			return fmt.Errorf("failed to count total entries: %w", err)
		}
		recentEntries, err := repo.CountRecentEntries(ctx, opts.Days)
		if err != nil {
			return fmt.Errorf("failed to count recent entries: %w", err)
		}
		oldEntries := totalEntries - recentEntries

		fmt.Fprintf(opts.Output, "Dry run: would delete %d entries older than %d days\n", oldEntries, opts.Days)
		fmt.Fprintf(opts.Output, "         (%d total entries, %d recent entries kept)\n", totalEntries, recentEntries)
		return nil
	}

	deleted, err := repo.PruneOldEntries(ctx, opts.Days)
	if err != nil {
		return fmt.Errorf("failed to prune entries: %w", err)
	}

	fmt.Fprintf(opts.Output, "✓ Deleted %d old entries\n", deleted)
	return nil
}
