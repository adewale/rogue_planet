package crawler

import (
	"context"
)

// FeedCrawler defines the interface for HTTP feed fetching operations.
// This interface enables dependency injection and makes testing easier by allowing
// mock implementations to be used in place of the concrete Crawler.
type FeedCrawler interface {
	// FetchWithRetry fetches a feed with exponential backoff retry logic
	// Returns a FeedResponse containing the feed data, HTTP status, caching info, and redirect details
	FetchWithRetry(ctx context.Context, feedURL string, cache FeedCache, maxRetries int) (*FeedResponse, error)
}

// Ensure Crawler implements FeedCrawler interface
var _ FeedCrawler = (*Crawler)(nil)
