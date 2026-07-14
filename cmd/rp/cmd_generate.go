package main

import (
	"context"
	"fmt"
)

func cmdGenerate(ctx context.Context, opts GenerateOptions) error {
	cfg, err := loadConfig(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if opts.Days > 0 {
		cfg.Planet.Days = opts.Days
	}

	_, _ = fmt.Fprintln(opts.Output, "Generating site...")
	if err := generateSite(ctx, cfg); err != nil {
		return fmt.Errorf("failed to generate site: %w", err)
	}

	_, _ = fmt.Fprintln(opts.Output, "✓ Generate complete")
	return nil
}
