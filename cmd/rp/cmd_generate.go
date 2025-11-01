package main

import "fmt"

func cmdGenerate(opts GenerateOptions) error {
	cfg, err := loadConfig(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if opts.Days > 0 {
		cfg.Planet.Days = opts.Days
	}

	fmt.Fprintln(opts.Output, "Generating site...")
	if err := generateSite(cfg); err != nil {
		return fmt.Errorf("failed to generate site: %w", err)
	}

	fmt.Fprintln(opts.Output, "✓ Generate complete")
	return nil
}
