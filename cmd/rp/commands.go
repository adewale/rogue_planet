package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adewale/rogue_planet/pkg/config"
	"github.com/adewale/rogue_planet/pkg/crawler"
	"github.com/adewale/rogue_planet/pkg/fetcher"
	"github.com/adewale/rogue_planet/pkg/generator"
	"github.com/adewale/rogue_planet/pkg/normalizer"
	"github.com/adewale/rogue_planet/pkg/opml"
	"github.com/adewale/rogue_planet/pkg/ratelimit"
	"github.com/adewale/rogue_planet/pkg/repository"
)

// Logger wraps standard logger with level support
type Logger struct {
	level int
}

const (
	LogLevelError = 0
	LogLevelWarn  = 1
	LogLevelInfo  = 2
	LogLevelDebug = 3
)

var globalLogger = &Logger{level: LogLevelInfo}

func (l *Logger) SetLevel(level string) {
	switch level {
	case "error":
		l.level = LogLevelError
	case "warn", "warning":
		l.level = LogLevelWarn
	case "info":
		l.level = LogLevelInfo
	case "debug":
		l.level = LogLevelDebug
	default:
		l.level = LogLevelInfo
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.level >= LogLevelError {
		log.Printf("ERROR: "+format, v...)
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level >= LogLevelWarn {
		log.Printf("WARN: "+format, v...)
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.level >= LogLevelInfo {
		log.Printf("INFO: "+format, v...)
	}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level >= LogLevelDebug {
		log.Printf("DEBUG: "+format, v...)
	}
}

// Command options structures for testability

type InitOptions struct {
	FeedsFile  string
	ConfigPath string
	Output     io.Writer
}

type AddFeedOptions struct {
	URL        string
	ConfigPath string
	Output     io.Writer
}

type AddAllOptions struct {
	FeedsFile  string
	ConfigPath string
	Output     io.Writer
}

type RemoveFeedOptions struct {
	URL        string
	ConfigPath string
	Output     io.Writer
}

type ListFeedsOptions struct {
	ConfigPath string
	Output     io.Writer
}

type StatusOptions struct {
	ConfigPath string
	Output     io.Writer
}

type UpdateOptions struct {
	ConfigPath string
	Verbose    bool
	Output     io.Writer
}

type FetchOptions struct {
	ConfigPath string
	Verbose    bool
	Output     io.Writer
}

type GenerateOptions struct {
	ConfigPath string
	Days       int
	Output     io.Writer
}

type PruneOptions struct {
	ConfigPath string
	Days       int
	DryRun     bool
	Output     io.Writer
}

type VerifyOptions struct {
	ConfigPath string
	Output     io.Writer
}

type ImportOPMLOptions struct {
	OPMLFile   string
	ConfigPath string
	DryRun     bool
	Output     io.Writer
}

type ExportOPMLOptions struct {
	OutputFile string
	ConfigPath string
	Output     io.Writer
}

// Command implementations

func cmdInit(opts InitOptions) error {
	fmt.Fprintln(opts.Output, "Initializing Rogue Planet...")

	// Create directories
	dirs := []string{"data", "public"}
	for _, dir := range dirs {
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

func cmdAddFeed(opts AddFeedOptions) error {
	if opts.URL == "" {
		return fmt.Errorf("URL is required")
	}

	_, repo, cleanup, err := openConfigAndRepo(opts.ConfigPath)
	if err != nil {
		return err
	}
	defer cleanup()

	// Add feed
	id, err := repo.AddFeed(opts.URL, "")
	if err != nil {
		return fmt.Errorf("failed to add feed: %w", err)
	}

	fmt.Fprintf(opts.Output, "✓ Added feed: %s (ID: %d)\n", opts.URL, id)
	return nil
}

func cmdAddAll(opts AddAllOptions) error {
	if opts.FeedsFile == "" {
		return fmt.Errorf("feeds file is required")
	}

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

	if len(feedURLs) == 0 {
		fmt.Fprintln(opts.Output, "No feeds found in file")
		return nil
	}

	fmt.Fprintf(opts.Output, "Adding %d feeds from %s...\n", len(feedURLs), opts.FeedsFile)

	// Add each feed to database
	addedCount := importFeedsFromURLs(repo, feedURLs, opts.Output)

	fmt.Fprintf(opts.Output, "\n✓ Added %d/%d feeds\n", addedCount, len(feedURLs))
	return nil
}

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

	// Remove feed
	if err := repo.RemoveFeed(feed.ID); err != nil {
		return fmt.Errorf("failed to remove feed: %w", err)
	}

	fmt.Fprintf(opts.Output, "✓ Removed feed: %s\n", opts.URL)
	return nil
}

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

func cmdStatus(opts StatusOptions) error {
	cfg, repo, cleanup, err := openConfigAndRepo(opts.ConfigPath)
	if err != nil {
		return err
	}
	defer cleanup()

	// Get feed counts
	feeds, err := repo.GetFeeds(false)
	if err != nil {
		return fmt.Errorf("failed to get feeds: %w", err)
	}

	activeFeeds := 0
	for _, feed := range feeds {
		if feed.Active {
			activeFeeds++
		}
	}

	// Get entry count
	totalEntries, err := repo.CountEntries()
	if err != nil {
		return fmt.Errorf("failed to count entries: %w", err)
	}

	// Get recent entry count (based on config days)
	recentEntries, err := repo.CountRecentEntries(cfg.Planet.Days)
	if err != nil {
		return fmt.Errorf("failed to count recent entries: %w", err)
	}

	// Display status
	fmt.Fprintln(opts.Output, "Rogue Planet Status")
	fmt.Fprintln(opts.Output, "===================")
	fmt.Fprintln(opts.Output)
	fmt.Fprintf(opts.Output, "Feeds:           %d total (%d active, %d inactive)\n", len(feeds), activeFeeds, len(feeds)-activeFeeds)
	fmt.Fprintf(opts.Output, "Entries:         %d total\n", totalEntries)
	fmt.Fprintf(opts.Output, "Recent entries:  %d (last %d days)\n", recentEntries, cfg.Planet.Days)
	fmt.Fprintln(opts.Output)
	fmt.Fprintf(opts.Output, "Output:          %s/index.html\n", cfg.Planet.OutputDir)
	fmt.Fprintf(opts.Output, "Database:        %s\n", cfg.Database.Path)

	return nil
}

func cmdUpdate(opts UpdateOptions) error {
	setVerboseLogging(opts.Verbose)

	// Load config
	cfg, err := loadConfig(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Fetch feeds
	fmt.Fprintln(opts.Output, "Fetching feeds...")
	if err := fetchFeeds(cfg); err != nil {
		return fmt.Errorf("failed to fetch feeds: %w", err)
	}

	// Generate site
	fmt.Fprintln(opts.Output, "Generating site...")
	if err := generateSite(cfg); err != nil {
		return fmt.Errorf("failed to generate site: %w", err)
	}

	fmt.Fprintln(opts.Output, "✓ Update complete")
	return nil
}

func cmdFetch(opts FetchOptions) error {
	setVerboseLogging(opts.Verbose)

	cfg, err := loadConfig(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Fprintln(opts.Output, "Fetching feeds...")
	if err := fetchFeeds(cfg); err != nil {
		return fmt.Errorf("failed to fetch feeds: %w", err)
	}

	fmt.Fprintln(opts.Output, "✓ Fetch complete")
	return nil
}

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

func cmdPrune(opts PruneOptions) error {
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

	deleted, err := repo.PruneOldEntries(opts.Days)
	if err != nil {
		return fmt.Errorf("failed to prune entries: %w", err)
	}

	fmt.Fprintf(opts.Output, "✓ Deleted %d old entries\n", deleted)
	return nil
}

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

func cmdImportOPML(opts ImportOPMLOptions) error {
	if opts.OPMLFile == "" {
		return fmt.Errorf("OPML file is required")
	}

	// Parse OPML file
	opmlDoc, err := opml.ParseFile(opts.OPMLFile)
	if err != nil {
		return fmt.Errorf("failed to parse OPML file: %w", err)
	}

	// Extract feeds
	feeds := opmlDoc.ExtractFeeds()

	if len(feeds) == 0 {
		fmt.Fprintln(opts.Output, "No feeds found in OPML file")
		return nil
	}

	if opts.DryRun {
		fmt.Fprintf(opts.Output, "DRY RUN: Importing feeds from %s...\n\n", opts.OPMLFile)
		fmt.Fprintf(opts.Output, "Found %d feeds in OPML file\n\n", len(feeds))

		// Load config and database to check for duplicates
		_, repo, cleanup, err := openConfigAndRepo(opts.ConfigPath)
		if err != nil {
			// Database might not exist yet, just show what would be imported
			for i, feed := range feeds {
				fmt.Fprintf(opts.Output, "  [%d/%d] Would add: %s (%s)\n", i+1, len(feeds), feed.FeedURL, feed.Title)
			}
			fmt.Fprintf(opts.Output, "\nDRY RUN: Would import %d feeds\n", len(feeds))
			return nil
		}
		defer cleanup()

		// Check which feeds already exist
		skipCount := 0
		for i, feed := range feeds {
			_, err := repo.GetFeedByURL(feed.FeedURL)
			if err == nil {
				fmt.Fprintf(opts.Output, "  [%d/%d] Would skip: %s (already exists)\n", i+1, len(feeds), feed.FeedURL)
				skipCount++
			} else {
				fmt.Fprintf(opts.Output, "  [%d/%d] Would add: %s (%s)\n", i+1, len(feeds), feed.FeedURL, feed.Title)
			}
		}

		fmt.Fprintf(opts.Output, "\nDRY RUN: Would import %d/%d feeds (%d duplicates skipped)\n", len(feeds)-skipCount, len(feeds), skipCount)
		return nil
	}

	// Real import
	fmt.Fprintf(opts.Output, "Importing feeds from %s...\n\n", opts.OPMLFile)
	fmt.Fprintf(opts.Output, "Found %d feeds in OPML file\n\n", len(feeds))

	_, repo, cleanup, err := openConfigAndRepo(opts.ConfigPath)
	if err != nil {
		return err
	}
	defer cleanup()

	// Import each feed
	addedCount := 0
	skippedCount := 0

	for i, feed := range feeds {
		// Check if feed already exists
		_, err := repo.GetFeedByURL(feed.FeedURL)
		if err == nil {
			fmt.Fprintf(opts.Output, "  [%d/%d] %s\n", i+1, len(feeds), feed.FeedURL)
			fmt.Fprintln(opts.Output, "         ⚠ Skipped (already exists)")
			skippedCount++
			continue
		}

		// Validate URL
		if err := crawler.ValidateURL(feed.FeedURL); err != nil {
			fmt.Fprintf(opts.Output, "  [%d/%d] %s\n", i+1, len(feeds), feed.FeedURL)
			fmt.Fprintf(opts.Output, "         ✗ Skipped (invalid URL: %v)\n", err)
			skippedCount++
			continue
		}

		// Add feed
		title := feed.Title
		if title == "" {
			title = feed.FeedURL
		}

		id, err := repo.AddFeed(feed.FeedURL, title)
		if err != nil {
			fmt.Fprintf(opts.Output, "  [%d/%d] %s\n", i+1, len(feeds), feed.FeedURL)
			fmt.Fprintf(opts.Output, "         ✗ Failed: %v\n", err)
			skippedCount++
			continue
		}

		fmt.Fprintf(opts.Output, "  [%d/%d] Adding %s (%s)\n", i+1, len(feeds), feed.FeedURL, title)
		fmt.Fprintf(opts.Output, "         ✓ Added (ID: %d)\n", id)
		addedCount++
	}

	fmt.Fprintf(opts.Output, "\n✓ Successfully imported %d/%d feeds\n", addedCount, len(feeds))
	fmt.Fprintf(opts.Output, "  - %d added\n", addedCount)
	fmt.Fprintf(opts.Output, "  - %d skipped (duplicates or invalid)\n", skippedCount)

	return nil
}

func cmdExportOPML(opts ExportOPMLOptions) error {
	// Load config
	cfg, repo, cleanup, err := openConfigAndRepo(opts.ConfigPath)
	if err != nil {
		return err
	}
	defer cleanup()

	// Get all feeds
	repoFeeds, err := repo.GetFeeds(false)
	if err != nil {
		return fmt.Errorf("failed to get feeds: %w", err)
	}

	if len(repoFeeds) == 0 {
		fmt.Fprintln(opts.Output, "No feeds to export")
		return nil
	}

	// Convert to OPML feeds
	opmlFeeds := make([]opml.Feed, 0, len(repoFeeds))
	for _, feed := range repoFeeds {
		title := feed.Title
		if title == "" {
			title = feed.URL
		}

		opmlFeeds = append(opmlFeeds, opml.Feed{
			Title:   title,
			FeedURL: feed.URL,
			WebURL:  feed.Link,
		})
	}

	// Generate OPML
	metadata := opml.Metadata{
		Title:      cfg.Planet.Name + " Feed List",
		OwnerName:  cfg.Planet.OwnerName,
		OwnerEmail: cfg.Planet.OwnerEmail,
	}

	opmlDoc, err := opml.Generate(opmlFeeds, metadata)
	if err != nil {
		return fmt.Errorf("failed to generate OPML: %w", err)
	}

	// Marshal to XML
	xmlData, err := opmlDoc.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal OPML: %w", err)
	}

	// Write to file or stdout
	if opts.OutputFile != "" {
		if err := os.WriteFile(opts.OutputFile, xmlData, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		fmt.Fprintf(opts.Output, "✓ Exported %d feeds to %s\n", len(opmlFeeds), opts.OutputFile)
	} else {
		fmt.Fprint(opts.Output, string(xmlData))
	}

	return nil
}

// Helper functions (unchanged from main.go)

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

func fetchFeeds(cfg *config.Config) error {
	// Set log level from config
	globalLogger.SetLevel(cfg.Planet.LogLevel)

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

	globalLogger.Info("Fetching %d feeds with concurrency=%d", len(feeds), cfg.Planet.ConcurrentFetch)

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
	globalLogger.Debug("Rate limiter configured: %d requests/min, burst=%d", cfg.Planet.RequestsPerMinute, cfg.Planet.RateLimitBurst)

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
	feedFetcher := fetcher.New(c, n, repo, &mu, globalLogger, cfg.Planet.MaxRetries)

	// Fetch feeds concurrently
	for i, feed := range feeds {
		wg.Add(1)
		go func(index int, f repository.Feed) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			fmt.Printf("  [%d/%d] Fetching %s\n", index+1, len(feeds), f.URL)

			// Apply rate limiting before fetching
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := rateLimiter.Wait(ctx, f.URL); err != nil {
				globalLogger.Error("Rate limiter error for %s: %v", f.URL, err)
				cancel()
				return
			}

			// Fetch and process feed (fetcher handles mutex internally for database writes)
			result := feedFetcher.FetchFeed(ctx, f)
			cancel()

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
	globalLogger.Info("Completed fetching all feeds")

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
