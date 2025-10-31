package repository

import "time"

// FeedRepository defines the interface for feed and entry storage operations.
// This interface enables dependency injection and makes testing easier by allowing
// mock implementations to be used in place of the concrete Repository.
type FeedRepository interface {
	// GetFeeds retrieves all feeds from the database
	// If activeOnly is true, only returns feeds where Active = true
	GetFeeds(activeOnly bool) ([]Feed, error)

	// AddFeed adds a new feed to the database
	AddFeed(url, title string) (int64, error)

	// GetFeedByURL retrieves a feed by its URL
	GetFeedByURL(url string) (*Feed, error)

	// UpdateFeed updates feed metadata (title, link, updated time)
	UpdateFeed(id int64, title, link string, updated time.Time) error

	// UpdateFeedURL updates the feed's URL (for 301 redirects)
	UpdateFeedURL(id int64, newURL string) error

	// UpdateFeedCache updates the feed's HTTP cache headers
	UpdateFeedCache(id int64, etag, lastModified string, lastFetched time.Time) error

	// UpdateFeedError records a fetch error for a feed
	UpdateFeedError(id int64, errorMsg string) error

	// RemoveFeed removes a feed and its entries from the database
	RemoveFeed(id int64) error

	// UpsertEntry inserts or updates an entry (deduplicates by feed_id + entry_id)
	UpsertEntry(entry *Entry) error

	// GetRecentEntries retrieves entries from the last N days
	GetRecentEntries(days int) ([]Entry, error)

	// GetRecentEntriesWithOptions retrieves entries with filter and sort options
	GetRecentEntriesWithOptions(days int, filterByFirstSeen bool, sortBy string) ([]Entry, error)

	// CountEntries returns the total number of entries in the database
	CountEntries() (int64, error)

	// CountRecentEntries returns the number of entries published within the last N days
	CountRecentEntries(days int) (int64, error)

	// PruneOldEntries deletes entries older than N days
	PruneOldEntries(days int) (int64, error)

	// Close closes the database connection
	Close() error
}

// Ensure Repository implements FeedRepository interface
var _ FeedRepository = (*Repository)(nil)
