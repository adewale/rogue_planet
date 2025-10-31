// Package fetcher provides business logic for fetching and processing individual feeds.
// It separates the core feed processing logic from the orchestration concerns
// (concurrency, rate limiting, etc.) to improve testability.
package fetcher

import (
	"context"
	"fmt"
	"sync"

	"github.com/adewale/rogue_planet/pkg/crawler"
	"github.com/adewale/rogue_planet/pkg/normalizer"
	"github.com/adewale/rogue_planet/pkg/repository"
)

// Logger interface for structured logging
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// Fetcher handles the business logic for fetching and processing a single feed.
// It coordinates between the crawler (HTTP fetching), normalizer (parsing),
// and repository (storage) components.
//
// The repoMutex protects concurrent database access. HTTP fetching and feed
// parsing operations run concurrently without locks for maximum performance.
type Fetcher struct {
	crawler    crawler.FeedCrawler
	normalizer normalizer.FeedNormalizer
	repo       repository.FeedRepository
	repoMutex  sync.Locker // Protects repository operations only
	logger     Logger
	maxRetries int
}

// New creates a new Fetcher with the provided dependencies
//
// The repoMutex protects concurrent access to the repository. Pass a shared
// mutex when multiple goroutines will call FetchFeed concurrently. Pass nil
// for single-threaded usage or when the repository is already thread-safe.
func New(
	c crawler.FeedCrawler,
	n normalizer.FeedNormalizer,
	r repository.FeedRepository,
	repoMutex sync.Locker,
	logger Logger,
	maxRetries int,
) *Fetcher {
	return &Fetcher{
		crawler:    c,
		normalizer: n,
		repo:       r,
		repoMutex:  repoMutex,
		logger:     logger,
		maxRetries: maxRetries,
	}
}

// FetchResult contains the result of a feed fetch operation
type FetchResult struct {
	StoredEntries int
	NotModified   bool
	Error         error
}

// FetchFeed fetches and processes a single feed.
// This method contains all the business logic for:
// - Fetching the feed with retries (NO LOCK - runs concurrently)
// - Handling redirects (301) (WITH LOCK - database write)
// - Handling cached responses (304) (WITH LOCK - database write)
// - Parsing and normalizing feed content (NO LOCK - runs concurrently)
// - Storing feed metadata and entries (WITH LOCK - database writes)
//
// HTTP fetching and feed parsing run without locks for maximum concurrency.
// Only database operations are protected by the mutex.
//
// The caller is responsible for:
// - Rate limiting
// - Spawning goroutines for concurrency
// - Progress reporting
func (f *Fetcher) FetchFeed(ctx context.Context, feed repository.Feed) FetchResult {
	f.logger.Debug("Starting fetch for %s (ID: %d)", feed.URL, feed.ID)

	// Prepare cache
	cache := crawler.FeedCache{
		URL:          feed.URL,
		ETag:         feed.ETag,
		LastModified: feed.LastModified,
		LastFetched:  feed.LastFetched,
	}

	// Fetch feed with retry logic (exponential backoff) - NO LOCK (concurrent HTTP)
	resp, err := f.crawler.FetchWithRetry(ctx, feed.URL, cache, f.maxRetries)
	if err != nil {
		f.logger.Error("Error fetching %s: %v", feed.URL, err)
		// Database write - WITH LOCK
		f.lock()
		if updateErr := f.repo.UpdateFeedError(feed.ID, err.Error()); updateErr != nil {
			f.logger.Error("Failed to update feed error for %s: %v", feed.URL, updateErr)
		}
		f.unlock()
		return FetchResult{Error: fmt.Errorf("fetch failed: %w", err)}
	}

	// Handle 301 permanent redirect - update feed URL in database
	if resp.PermanentRedirect && resp.FinalURL != feed.URL {
		f.logger.Info("Feed %s permanently redirected to %s (301)", feed.URL, resp.FinalURL)
		// Database write - WITH LOCK
		f.lock()
		if updateErr := f.repo.UpdateFeedURL(feed.ID, resp.FinalURL); updateErr != nil {
			f.logger.Error("Failed to update feed URL for %s: %v", feed.URL, updateErr)
		} else {
			f.logger.Info("Updated feed URL from %s to %s", feed.URL, resp.FinalURL)
		}
		f.unlock()
	}

	// Handle 304 Not Modified
	if resp.NotModified {
		f.logger.Debug("%s returned 304 Not Modified", feed.URL)
		// Database write - WITH LOCK
		f.lock()
		if updateErr := f.repo.UpdateFeedCache(feed.ID, resp.NewCache.ETag, resp.NewCache.LastModified, resp.FetchTime); updateErr != nil {
			f.logger.Error("Failed to update feed cache for %s: %v", feed.URL, updateErr)
		}
		f.unlock()
		return FetchResult{NotModified: true}
	}

	// Parse and normalize feed - NO LOCK (concurrent parsing)
	metadata, entries, err := f.normalizer.Parse(resp.Body, feed.URL, resp.FetchTime)
	if err != nil {
		f.logger.Error("Error parsing %s: %v", feed.URL, err)
		// Database write - WITH LOCK
		f.lock()
		if updateErr := f.repo.UpdateFeedError(feed.ID, err.Error()); updateErr != nil {
			f.logger.Error("Failed to update feed error for %s: %v", feed.URL, updateErr)
		}
		f.unlock()
		return FetchResult{Error: fmt.Errorf("parse failed: %w", err)}
	}

	f.logger.Debug("Parsed %d entries from %s", len(entries), feed.URL)

	// Database writes - WITH LOCK (entire section)
	f.lock()

	// Update feed metadata and cache
	if updateErr := f.repo.UpdateFeed(feed.ID, metadata.Title, metadata.Link, metadata.Updated); updateErr != nil {
		f.logger.Error("Failed to update feed metadata for %s: %v", feed.URL, updateErr)
	}
	if updateErr := f.repo.UpdateFeedCache(feed.ID, resp.NewCache.ETag, resp.NewCache.LastModified, resp.FetchTime); updateErr != nil {
		f.logger.Error("Failed to update feed cache for %s: %v", feed.URL, updateErr)
	}

	// Store entries
	storedCount := 0
	for _, entry := range entries {
		repoEntry := &repository.Entry{
			FeedID:      feed.ID,
			EntryID:     entry.ID,
			Title:       entry.Title,
			Link:        entry.Link,
			Author:      entry.Author,
			Published:   entry.Published,
			Updated:     entry.Updated,
			Content:     entry.Content,
			ContentType: entry.ContentType,
			Summary:     entry.Summary,
			FirstSeen:   entry.FirstSeen,
		}

		if err := f.repo.UpsertEntry(repoEntry); err != nil {
			f.logger.Warn("Error storing entry from %s: %v", feed.URL, err)
		} else {
			storedCount++
		}
	}

	f.unlock()

	f.logger.Info("Successfully processed %s: %d entries", feed.URL, storedCount)

	return FetchResult{StoredEntries: storedCount}
}

// lock acquires the repository mutex if one was provided
func (f *Fetcher) lock() {
	if f.repoMutex != nil {
		f.repoMutex.Lock()
	}
}

// unlock releases the repository mutex if one was provided
func (f *Fetcher) unlock() {
	if f.repoMutex != nil {
		f.repoMutex.Unlock()
	}
}
