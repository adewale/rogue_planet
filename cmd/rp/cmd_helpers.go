package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/adewale/rogue_planet/pkg/config"
	"github.com/adewale/rogue_planet/pkg/crawler"
	"github.com/adewale/rogue_planet/pkg/fetcher"
	"github.com/adewale/rogue_planet/pkg/generator"
	"github.com/adewale/rogue_planet/pkg/logging"
	"github.com/adewale/rogue_planet/pkg/normalizer"
	"github.com/adewale/rogue_planet/pkg/ratelimit"
	"github.com/adewale/rogue_planet/pkg/repository"
)

// loadConfig loads configuration from file, falling back to defaults if file doesn't exist
func loadConfig(path string) (*config.Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return config.Default(), nil
	}
	return config.LoadFromFile(path)
}

// setVerboseLogging configures log output to include file and line numbers
func setVerboseLogging(verbose bool) {
	if verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}
}

// openConfigAndRepo loads config and opens database, returning both along with a cleanup function
// The cleanup function should be called with defer to ensure the repository is closed
func openConfigAndRepo(configPath string) (*config.Config, *repository.Repository, func(), error) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	repo, err := repository.New(cfg.Database.Path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	cleanup := func() { repo.Close() }
	return cfg, repo, cleanup, nil
}

// importFeedsFromURLs adds a list of feed URLs to the repository with progress reporting
// Returns the number of successfully added feeds
func importFeedsFromURLs(repo *repository.Repository, feedURLs []string, output io.Writer) int {
	addedCount := 0
	for i, url := range feedURLs {
		fmt.Fprintf(output, "  [%d/%d] Adding %s\n", i+1, len(feedURLs), url)
		id, err := repo.AddFeed(url, "")
		if err != nil {
			log.Printf("         Warning: Failed to add feed: %v", err)
			continue
		}
		fmt.Fprintf(output, "         ✓ Added (ID: %d)\n", id)
		addedCount++
	}
	return addedCount
}

func fetchFeeds(cfg *config.Config, logger logging.Logger) error {
	// Set log level from config if logger supports it
	if stdLogger, ok := logger.(*logging.StandardLogger); ok {
		stdLogger.SetLevel(cfg.Planet.LogLevel)
	}

	repo, err := repository.New(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer repo.Close()

	// Get feeds from database
	feeds, err := repo.GetFeeds(true)
	if err != nil {
		return fmt.Errorf("get feeds: %w", err)
	}

	if len(feeds) == 0 {
		fmt.Println("No feeds to fetch. Add feeds with 'rp add-feed <url>'")
		return nil
	}

	logger.Info("Fetching %d feeds with concurrency=%d", len(feeds), cfg.Planet.ConcurrentFetch)

	// Create crawler with custom configuration
	c := crawler.NewWithConfig(crawler.CrawlerConfig{
		UserAgent:                    cfg.Planet.UserAgent,
		MaxIdleConns:                 cfg.Planet.MaxIdleConns,
		MaxIdleConnsPerHost:          cfg.Planet.MaxIdleConnsPerHost,
		MaxConnsPerHost:              cfg.Planet.MaxConnsPerHost,
		IdleConnTimeoutSeconds:       cfg.Planet.IdleConnTimeoutSeconds,
		HTTPTimeoutSeconds:           cfg.Planet.HTTPTimeoutSeconds,
		DialTimeoutSeconds:           cfg.Planet.DialTimeoutSeconds,
		TLSHandshakeTimeoutSeconds:   cfg.Planet.TLSHandshakeTimeoutSeconds,
		ResponseHeaderTimeoutSeconds: cfg.Planet.ResponseHeaderTimeoutSeconds,
	})
	n := normalizer.New()

	// Create rate limiter for per-domain rate limiting
	rateLimiter := ratelimit.New(cfg.Planet.RequestsPerMinute, cfg.Planet.RateLimitBurst)
	logger.Debug("Rate limiter configured: %d requests/min, burst=%d", cfg.Planet.RequestsPerMinute, cfg.Planet.RateLimitBurst)

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Handle signals in background
	go func() {
		sig, ok := <-sigChan
		if !ok {
			// Channel closed, normal shutdown
			return
		}
		logger.Info("Received signal %v, cancelling fetches...", sig)
		cancel()
	}()

	// Use semaphore pattern for concurrency control
	concurrency := cfg.Planet.ConcurrentFetch
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > len(feeds) {
		concurrency = len(feeds)
	}

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex // Protects repo writes

	// Create fetcher with dependencies (passes mutex for database protection)
	feedFetcher := fetcher.New(c, n, repo, &mu, logger, cfg.Planet.MaxRetries)

	// Fetch feeds concurrently
	for i, feed := range feeds {
		wg.Add(1)
		go func(index int, f repository.Feed) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				logger.Debug("Skipping %s (cancelled)", f.URL)
				return
			}

			// Check if already cancelled before starting
			select {
			case <-ctx.Done():
				logger.Debug("Skipping %s (cancelled)", f.URL)
				return
			default:
			}

			fmt.Printf("  [%d/%d] Fetching %s\n", index+1, len(feeds), f.URL)

			// Apply rate limiting before fetching (use parent context)
			fetchCtx, fetchCancel := context.WithTimeout(ctx, 30*time.Second)
			defer fetchCancel()

			if err := rateLimiter.Wait(fetchCtx, f.URL); err != nil {
				if err == context.Canceled {
					logger.Debug("Fetch cancelled for %s", f.URL)
				} else {
					logger.Error("Rate limiter error for %s: %v", f.URL, err)
				}
				return
			}

			// Fetch and process feed (fetcher handles mutex internally for database writes)
			result := feedFetcher.FetchFeed(fetchCtx, f)

			// Report results
			if result.Error != nil {
				// Error already logged by fetcher
				return
			}

			if result.NotModified {
				fmt.Printf("    Not modified (cached)\n")
				return
			}

			fmt.Printf("    Stored %d entries\n", result.StoredEntries)
		}(i, feed)
	}

	// Wait for all fetches to complete
	wg.Wait()

	// Stop listening for signals
	signal.Stop(sigChan)
	close(sigChan)

	// Check if we were cancelled
	select {
	case <-ctx.Done():
		logger.Info("Fetch operation cancelled")
		return fmt.Errorf("operation cancelled by user")
	default:
		logger.Info("Completed fetching all feeds")
	}

	return nil
}

func generateSite(cfg *config.Config) error {
	repo, err := repository.New(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer repo.Close()

	// Get recent entries
	entries, err := repo.GetRecentEntriesWithOptions(cfg.Planet.Days, cfg.Planet.FilterByFirstSeen, cfg.Planet.SortBy)
	if err != nil {
		return fmt.Errorf("get entries: %w", err)
	}

	// Get feeds for metadata
	feeds, err := repo.GetFeeds(true)
	if err != nil {
		return fmt.Errorf("get feeds: %w", err)
	}

	feedMap := make(map[int64]*repository.Feed)
	for i := range feeds {
		feedMap[feeds[i].ID] = &feeds[i]
	}

	// Convert to generator format
	genEntries := make([]generator.EntryData, 0, len(entries))
	for _, entry := range entries {
		feed := feedMap[entry.FeedID]
		if feed == nil {
			continue
		}

		// SAFETY: Content was sanitized by normalizer.Parse() before storage.
		// See pkg/normalizer/normalizer.go:56-69 for HTML sanitization using bluemonday.
		// Title, Content, and Summary are safe for template.HTML after sanitization:
		// - XSS vectors removed (script tags, event handlers, javascript: URLs)
		// - Only http/https schemes allowed in links
		// - Dangerous tags stripped (object, embed, iframe, base)
		genEntries = append(genEntries, generator.EntryData{
			Title:     template.HTML(entry.Title),
			Link:      entry.Link,
			Author:    entry.Author,
			FeedTitle: feed.Title,
			FeedLink:  feed.Link,
			Published: entry.Published,
			Updated:   entry.Updated,
			Content:   template.HTML(entry.Content),
			Summary:   template.HTML(entry.Summary),
		})
	}

	// Create generator
	var gen *generator.Generator
	if cfg.Planet.Template != "" {
		gen, err = generator.NewWithTemplate(cfg.Planet.Template)
		if err != nil {
			return fmt.Errorf("create generator with template: %w", err)
		}
	} else {
		gen, err = generator.New()
		if err != nil {
			return fmt.Errorf("create generator: %w", err)
		}
	}

	// Convert feeds for sidebar
	genFeeds := make([]generator.FeedData, 0, len(feeds))
	for _, feed := range feeds {
		genFeeds = append(genFeeds, generator.FeedData{
			Title:       feed.Title,
			Link:        feed.Link,
			URL:         feed.URL,
			LastUpdated: feed.LastFetched,
			ErrorCount:  feed.FetchErrorCount,
		})
	}

	// Generate HTML
	data := generator.TemplateData{
		Title:       cfg.Planet.Name,
		Link:        cfg.Planet.Link,
		OwnerName:   cfg.Planet.OwnerName,
		OwnerEmail:  cfg.Planet.OwnerEmail,
		Entries:     genEntries,
		GroupByDate: cfg.Planet.GroupByDate,
		Feeds:       genFeeds,
	}

	outputPath := filepath.Join(cfg.Planet.OutputDir, "index.html")
	if err := gen.GenerateToFile(outputPath, data); err != nil {
		return fmt.Errorf("generate file: %w", err)
	}

	fmt.Printf("  Generated %s with %d entries\n", outputPath, len(entries))
	return nil
}
