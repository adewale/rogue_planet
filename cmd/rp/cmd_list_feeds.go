package main

import (
	"fmt"
	"time"
)

func cmdListFeeds(opts ListFeedsOptions) error {
	_, repo, cleanup, err := openConfigAndRepo(opts.ConfigPath)
	if err != nil {
		return err
	}
	defer cleanup()

	// Get feeds
	feeds, err := repo.GetFeeds(false)
	if err != nil {
		return fmt.Errorf("failed to get feeds: %w", err)
	}

	if len(feeds) == 0 {
		fmt.Fprintln(opts.Output, "No feeds configured.")
		return nil
	}

	fmt.Fprintf(opts.Output, "Configured feeds (%d):\n\n", len(feeds))
	for _, feed := range feeds {
		status := "active"
		if !feed.Active {
			status = "inactive"
		}

		fmt.Fprintf(opts.Output, "  [%d] %s\n", feed.ID, feed.URL)
		if feed.Title != "" {
			fmt.Fprintf(opts.Output, "      Title: %s\n", feed.Title)
		}
		fmt.Fprintf(opts.Output, "      Status: %s\n", status)
		if !feed.LastFetched.IsZero() {
			fmt.Fprintf(opts.Output, "      Last fetched: %s\n", feed.LastFetched.Format(time.RFC3339))
		}
		if feed.FetchError != "" {
			fmt.Fprintf(opts.Output, "      Error: %s\n", feed.FetchError)
		}
		fmt.Fprintln(opts.Output)
	}

	return nil
}
