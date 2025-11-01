package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func cmdRemoveFeed(opts RemoveFeedOptions) error {
	if opts.URL == "" {
		return fmt.Errorf("URL is required")
	}

	_, repo, cleanup, err := openConfigAndRepo(opts.ConfigPath)
	if err != nil {
		return err
	}
	defer cleanup()

	// Find feed
	feed, err := repo.GetFeedByURL(opts.URL)
	if err != nil {
		return fmt.Errorf("feed not found: %w", err)
	}

	// Get entry count for this feed
	entryCount, err := repo.GetEntryCountForFeed(feed.ID)
	if err != nil {
		return fmt.Errorf("failed to count entries: %w", err)
	}

	// Interactive confirmation unless --force is specified
	if !opts.Force {
		// Check if input is a terminal (for production use with os.Stdin)
		// For testing, non-os.File inputs (like strings.Reader) are allowed for mocking
		if inputFile, isFile := opts.Input.(*os.File); isFile {
			// Only check terminal status for actual files (production)
			stat, err := inputFile.Stat()
			if err != nil {
				return fmt.Errorf("cannot determine terminal status: %w", err)
			}
			isTerminal := (stat.Mode() & os.ModeCharDevice) != 0
			if !isTerminal {
				return fmt.Errorf("cannot prompt for confirmation in non-interactive mode. Use --force to skip confirmation")
			}
		}
		// Non-file inputs (e.g., strings.Reader in tests) are allowed to proceed

		// Display feed information
		feedTitle := feed.Title
		if feedTitle == "" {
			feedTitle = "(no title)"
		}

		fmt.Fprintf(opts.Output, "Feed: %s\n", feed.URL)
		fmt.Fprintf(opts.Output, "Title: %s\n", feedTitle)
		fmt.Fprintf(opts.Output, "Entries: %d\n\n", entryCount)

		// Prompt for confirmation (matches spec exactly)
		fmt.Fprintf(opts.Output, "Remove this feed and all %d entries? (y/N): ", entryCount)

		// Read user input
		reader := bufio.NewReader(opts.Input)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		// Parse response
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Fprintln(opts.Output, "Cancelled.")
			return &ErrUserCancelled{"operation cancelled by user"}
		}
	}

	// Remove feed (CASCADE DELETE will remove all entries)
	if err := repo.RemoveFeed(feed.ID); err != nil {
		return fmt.Errorf("failed to remove feed: %w", err)
	}

	fmt.Fprintf(opts.Output, "✓ Removed feed: %s (%d entries deleted)\n", opts.URL, entryCount)
	return nil
}
