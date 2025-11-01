package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adewale/rogue_planet/pkg/config"
	"github.com/adewale/rogue_planet/pkg/repository"
)

func cmdVerify(opts VerifyOptions) error {
	errors := []string{}

	// 1. Load and validate config file
	cfg, err := config.LoadFromFile(opts.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file not found: %s", opts.ConfigPath)
		}
		return fmt.Errorf("invalid config file: %w", err)
	}

	// Validate config values
	if err := cfg.Validate(); err != nil {
		errors = append(errors, fmt.Sprintf("Invalid config value: %v", err))
	}

	// 2. Check database accessibility and schema
	if _, err := os.Stat(cfg.Database.Path); os.IsNotExist(err) {
		errors = append(errors, "Database does not exist → run 'rp init' to create")
	} else {
		// Try to open database
		repo, err := repository.New(cfg.Database.Path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Database error: %v", err))
		} else {
			// Try a simple query to verify schema
			_, err := repo.GetFeeds(false)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Database schema error: %v", err))
			}
			repo.Close()
		}
	}

	// 3. Check output directory
	if _, err := os.Stat(cfg.Planet.OutputDir); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("Output directory does not exist → mkdir -p %s", cfg.Planet.OutputDir))
	} else {
		// Check if writable
		testFile := filepath.Join(cfg.Planet.OutputDir, ".write_test")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			errors = append(errors, fmt.Sprintf("Output directory not writable → chmod 755 %s", cfg.Planet.OutputDir))
		} else {
			os.Remove(testFile)
		}
	}

	// 4. Check custom template if specified
	if cfg.Planet.Template != "" {
		if _, err := os.Stat(cfg.Planet.Template); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("Template file not found: %s", cfg.Planet.Template))
		}
	}

	// 5. Report results
	if len(errors) > 0 {
		fmt.Fprintln(opts.Output, "✗ Configuration validation failed")
		fmt.Fprintln(opts.Output)
		for _, e := range errors {
			fmt.Fprintf(opts.Output, "- %s\n", e)
		}
		fmt.Fprintln(opts.Output)
		fmt.Fprintf(opts.Output, "Found %d errors.\n", len(errors))
		return fmt.Errorf("validation failed")
	}

	// Success - get feed/entry counts if database exists
	repo, err := repository.New(cfg.Database.Path)
	if err == nil {
		defer repo.Close()
		feeds, _ := repo.GetFeeds(false)
		entries, _ := repo.CountEntries()
		fmt.Fprintf(opts.Output, "✓ Configuration valid (%d feeds, %d entries)\n", len(feeds), entries)
	} else {
		fmt.Fprintln(opts.Output, "✓ Configuration valid")
	}

	return nil
}
