package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/adewale/rogue_planet/pkg/config"
)

func cmdInit(opts InitOptions) error {
	fmt.Fprintln(opts.Output, "Initializing Rogue Planet...")

	// Create directories
	dirs := []string{"data", "public"}
	for _, dir := range dirs {
		// Safety check: reject parent directory references to prevent path traversal
		if strings.Contains(dir, "..") {
			return fmt.Errorf("invalid directory path: %s (contains parent directory reference)", dir)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create example config file
	configContent := `[planet]
name = My Planet
link = https://example.com
owner_name = Your Name
owner_email = you@example.com
output_dir = ./public
days = 7
log_level = info
concurrent_fetches = 5
group_by_date = true

[database]
path = ./data/planet.db
`

	if err := os.WriteFile(opts.ConfigPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create config.ini: %w", err)
	}

	fmt.Fprintln(opts.Output, "✓ Created config.ini")
	fmt.Fprintln(opts.Output, "✓ Created data/ directory")
	fmt.Fprintln(opts.Output, "✓ Created public/ directory")

	// Import feeds if -f flag provided
	if opts.FeedsFile != "" {
		fmt.Fprintf(opts.Output, "\nImporting feeds from %s...\n", opts.FeedsFile)

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

		// Add each feed to database
		addedCount := importFeedsFromURLs(repo, feedURLs, opts.Output)

		fmt.Fprintf(opts.Output, "\n✓ Imported %d/%d feeds\n", addedCount, len(feedURLs))

		fmt.Fprintln(opts.Output, "\nNext steps:")
		fmt.Fprintln(opts.Output, "  1. Edit config.ini with your planet details")
		fmt.Fprintln(opts.Output, "  2. Run 'rp update' to fetch feeds and generate your planet")
	} else {
		fmt.Fprintln(opts.Output, "\nNext steps:")
		fmt.Fprintln(opts.Output, "  1. Edit config.ini with your planet details")
		fmt.Fprintln(opts.Output, "  2. Add feeds with 'rp add-feed <url>'")
		fmt.Fprintln(opts.Output, "  3. Run 'rp update' to fetch feeds and generate your planet")
	}

	return nil
}
