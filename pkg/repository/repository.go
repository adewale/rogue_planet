// Package repository provides SQLite database operations for feed storage.
//
// The repository handles all database interactions including feed management,
// entry storage with deduplication, and intelligent querying with fallback logic.
// It uses WAL mode for better concurrency and prepared statements for security.
package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrFeedNotFound  = errors.New("feed not found")
	ErrEntryNotFound = errors.New("entry not found")
)

// Feed represents a feed in the database
type Feed struct {
	ID              int64
	URL             string
	Title           string
	Link            string
	Updated         time.Time
	LastFetched     time.Time
	ETag            string
	LastModified    string
	FetchError      string
	FetchErrorCount int
	NextFetch       time.Time
	Active          bool
	FetchInterval   int // seconds
}

// Entry represents a feed entry in the database
type Entry struct {
	ID          int64
	FeedID      int64
	EntryID     string
	Title       string
	Link        string
	Author      string
	Published   time.Time
	Updated     time.Time
	Content     string
	ContentType string
	Summary     string
	FirstSeen   time.Time
}

// Repository handles database operations
type Repository struct {
	db *sql.DB
}

// New creates a new Repository and initializes the database
func New(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL: %w", err)
	}

	// Enable foreign keys (required for CASCADE DELETE)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	repo := &Repository{db: db}

	// Initialize schema
	if err := repo.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return repo, nil
}

// Close closes the database connection
func (r *Repository) Close() error {
	return r.db.Close()
}

// initSchema creates the database schema
func (r *Repository) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS feeds (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL UNIQUE,
		title TEXT,
		link TEXT,
		updated TEXT,
		last_fetched TEXT,
		etag TEXT,
		last_modified TEXT,
		fetch_error TEXT,
		fetch_error_count INTEGER DEFAULT 0,
		next_fetch TEXT,
		active INTEGER DEFAULT 1,
		fetch_interval INTEGER DEFAULT 3600
	);

	CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		feed_id INTEGER NOT NULL,
		entry_id TEXT NOT NULL,
		title TEXT,
		link TEXT,
		author TEXT,
		published TEXT,
		updated TEXT,
		content TEXT,
		content_type TEXT DEFAULT 'html',
		summary TEXT,
		first_seen TEXT,
		FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
		UNIQUE(feed_id, entry_id)
	);

	CREATE INDEX IF NOT EXISTS idx_entries_published ON entries(published DESC);
	CREATE INDEX IF NOT EXISTS idx_entries_updated ON entries(updated DESC);
	CREATE INDEX IF NOT EXISTS idx_entries_feed_id ON entries(feed_id);
	CREATE INDEX IF NOT EXISTS idx_entries_first_seen ON entries(first_seen DESC);
	CREATE INDEX IF NOT EXISTS idx_feeds_active ON feeds(active);
	CREATE INDEX IF NOT EXISTS idx_feeds_next_fetch ON feeds(next_fetch);
	`

	_, err := r.db.Exec(schema)
	if err != nil {
		return err
	}

	// Backfill first_seen for entries that don't have it
	// This ensures backwards compatibility with databases created before this feature
	// Uses COALESCE to handle NULL values and falls back to current time if both are NULL
	_, err = r.db.Exec(`
		UPDATE entries
		SET first_seen = COALESCE(
			NULLIF(first_seen, ''),
			published,
			updated,
			datetime('now')
		)
		WHERE first_seen IS NULL OR first_seen = ''
	`)
	if err != nil {
		return fmt.Errorf("backfill first_seen: %w", err)
	}

	return nil
}

// AddFeed adds a new feed to the database
func (r *Repository) AddFeed(url, title string) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO feeds (url, title, next_fetch)
		VALUES (?, ?, ?)
	`, url, title, time.Now().Format(time.RFC3339))

	if err != nil {
		return 0, fmt.Errorf("insert feed: %w", err)
	}

	return result.LastInsertId()
}

// UpdateFeed updates feed metadata
func (r *Repository) UpdateFeed(id int64, title, link string, updated time.Time) error {
	_, err := r.db.Exec(`
		UPDATE feeds
		SET title = ?, link = ?, updated = ?
		WHERE id = ?
	`, title, link, updated.Format(time.RFC3339), id)

	if err != nil {
		return fmt.Errorf("update feed: %w", err)
	}

	return nil
}

// UpdateFeedCache updates the HTTP cache headers for a feed
func (r *Repository) UpdateFeedCache(id int64, etag, lastModified string, lastFetched time.Time) error {
	_, err := r.db.Exec(`
		UPDATE feeds
		SET etag = ?, last_modified = ?, last_fetched = ?, fetch_error = NULL, fetch_error_count = 0
		WHERE id = ?
	`, etag, lastModified, lastFetched.Format(time.RFC3339), id)

	if err != nil {
		return fmt.Errorf("update feed cache: %w", err)
	}

	return nil
}

// UpdateFeedError records a fetch error for a feed
func (r *Repository) UpdateFeedError(id int64, errorMsg string) error {
	_, err := r.db.Exec(`
		UPDATE feeds
		SET fetch_error = ?, fetch_error_count = fetch_error_count + 1, last_fetched = ?
		WHERE id = ?
	`, errorMsg, time.Now().Format(time.RFC3339), id)

	if err != nil {
		return fmt.Errorf("update feed error: %w", err)
	}

	return nil
}

// GetFeeds returns all feeds, optionally filtering by active status
func (r *Repository) GetFeeds(activeOnly bool) ([]Feed, error) {
	query := "SELECT id, url, title, link, updated, last_fetched, etag, last_modified, fetch_error, fetch_error_count, next_fetch, active, fetch_interval FROM feeds"
	if activeOnly {
		query += " WHERE active = 1"
	}
	query += " ORDER BY id"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query feeds: %w", err)
	}
	defer rows.Close()

	return scanFeeds(rows)
}

// GetFeedByURL returns a feed by its URL
func (r *Repository) GetFeedByURL(url string) (*Feed, error) {
	row := r.db.QueryRow(`
		SELECT id, url, title, link, updated, last_fetched, etag, last_modified, fetch_error, fetch_error_count, next_fetch, active, fetch_interval
		FROM feeds
		WHERE url = ?
	`, url)

	feed := &Feed{}
	err := scanFeed(row, feed)
	if err == sql.ErrNoRows {
		return nil, ErrFeedNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query feed: %w", err)
	}

	return feed, nil
}

// RemoveFeed removes a feed and all its entries
func (r *Repository) RemoveFeed(id int64) error {
	_, err := r.db.Exec("DELETE FROM feeds WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete feed: %w", err)
	}

	return nil
}

// UpsertEntry inserts or updates an entry.
// On conflict (duplicate feed_id + entry_id), updates content fields but preserves
// first_seen to maintain the original discovery timestamp for spam prevention.
func (r *Repository) UpsertEntry(entry *Entry) error {
	_, err := r.db.Exec(`
		INSERT INTO entries (feed_id, entry_id, title, link, author, published, updated, content, content_type, summary, first_seen)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(feed_id, entry_id) DO UPDATE SET
			title = excluded.title,
			link = excluded.link,
			author = excluded.author,
			updated = excluded.updated,
			content = excluded.content,
			summary = excluded.summary
			-- first_seen deliberately NOT updated to preserve original discovery time
	`, entry.FeedID, entry.EntryID, entry.Title, entry.Link, entry.Author,
		entry.Published.Format(time.RFC3339), entry.Updated.Format(time.RFC3339),
		entry.Content, entry.ContentType, entry.Summary, entry.FirstSeen.Format(time.RFC3339))

	if err != nil {
		return fmt.Errorf("upsert entry: %w", err)
	}

	return nil
}

// GetRecentEntries returns entries from the last N days.
// If no entries are found in that time window, it falls back to returning
// the most recent 50 entries to ensure the page always has content.
func (r *Repository) GetRecentEntries(days int) ([]Entry, error) {
	cutoff := time.Now().AddDate(0, 0, -days)

	// First, try to get entries from the last N days
	rows, err := r.db.Query(`
		SELECT e.id, e.feed_id, e.entry_id, e.title, e.link, e.author, e.published, e.updated, e.content, e.content_type, e.summary, e.first_seen
		FROM entries e
		JOIN feeds f ON e.feed_id = f.id
		WHERE f.active = 1 AND e.published >= ?
		ORDER BY e.published DESC
	`, cutoff.Format(time.RFC3339))

	if err != nil {
		return nil, fmt.Errorf("query entries: %w", err)
	}
	defer rows.Close()

	entries, err := scanEntries(rows)
	if err != nil {
		return nil, err
	}

	// If we found entries in the time window, return them
	if len(entries) > 0 {
		return entries, nil
	}

	// Otherwise, fall back to the most recent 50 entries regardless of date
	// This ensures the page always has content even if feeds are stale
	rows, err = r.db.Query(`
		SELECT e.id, e.feed_id, e.entry_id, e.title, e.link, e.author, e.published, e.updated, e.content, e.content_type, e.summary, e.first_seen
		FROM entries e
		JOIN feeds f ON e.feed_id = f.id
		WHERE f.active = 1
		ORDER BY e.published DESC
		LIMIT 50
	`)

	if err != nil {
		return nil, fmt.Errorf("query fallback entries: %w", err)
	}
	defer rows.Close()

	return scanEntries(rows)
}

// GetRecentEntriesWithOptions returns entries based on filtering and sorting preferences.
// If filterByFirstSeen is true, only entries first seen within the time window are returned.
// sortBy determines the ordering: "published" or "first_seen".
// Falls back to the most recent 50 entries if none found in the time window.
func (r *Repository) GetRecentEntriesWithOptions(days int, filterByFirstSeen bool, sortBy string) ([]Entry, error) {
	// Validate sortBy parameter to prevent SQL injection
	if sortBy != "published" && sortBy != "first_seen" {
		return nil, fmt.Errorf("invalid sortBy value: %s (must be 'published' or 'first_seen')", sortBy)
	}

	cutoff := time.Now().AddDate(0, 0, -days)

	// Choose filter field (validated by boolean type)
	filterField := "e.published"
	if filterByFirstSeen {
		filterField = "e.first_seen"
	}

	// Choose sort field (validated above)
	sortField := "e.published"
	if sortBy == "first_seen" {
		sortField = "e.first_seen"
	}

	query := fmt.Sprintf(`
		SELECT e.id, e.feed_id, e.entry_id, e.title, e.link, e.author,
		       e.published, e.updated, e.content, e.content_type, e.summary, e.first_seen
		FROM entries e
		JOIN feeds f ON e.feed_id = f.id
		WHERE f.active = 1 AND %s >= ?
		ORDER BY %s DESC
	`, filterField, sortField)

	rows, err := r.db.Query(query, cutoff.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("query entries: %w", err)
	}
	defer rows.Close()

	entries, err := scanEntries(rows)
	if err != nil {
		return nil, err
	}

	// If we found entries, return them
	if len(entries) > 0 {
		return entries, nil
	}

	// Fallback to most recent 50 entries (use same sort field)
	query = fmt.Sprintf(`
		SELECT e.id, e.feed_id, e.entry_id, e.title, e.link, e.author,
		       e.published, e.updated, e.content, e.content_type, e.summary, e.first_seen
		FROM entries e
		JOIN feeds f ON e.feed_id = f.id
		WHERE f.active = 1
		ORDER BY %s DESC
		LIMIT 50
	`, sortField)

	rows, err = r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query fallback entries: %w", err)
	}
	defer rows.Close()

	return scanEntries(rows)
}

// CountEntries returns the total number of entries in the database
func (r *Repository) CountEntries() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM entries").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count entries: %w", err)
	}
	return count, nil
}

// CountRecentEntries returns the number of entries published within the last N days
func (r *Repository) CountRecentEntries(days int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -days)

	var count int64
	err := r.db.QueryRow(`
		SELECT COUNT(*)
		FROM entries
		WHERE published >= ?
	`, cutoff.Format(time.RFC3339)).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("count recent entries: %w", err)
	}
	return count, nil
}

// PruneOldEntries deletes entries older than N days
func (r *Repository) PruneOldEntries(days int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -days)

	result, err := r.db.Exec(`
		DELETE FROM entries
		WHERE published < ?
	`, cutoff.Format(time.RFC3339))

	if err != nil {
		return 0, fmt.Errorf("prune entries: %w", err)
	}

	return result.RowsAffected()
}

// Helper functions for scanning rows

func scanFeed(row interface{ Scan(...interface{}) error }, feed *Feed) error {
	var title, link, updated, lastFetched, etag, lastModified, fetchError, nextFetch sql.NullString
	var active sql.NullInt64

	err := row.Scan(
		&feed.ID, &feed.URL, &title, &link,
		&updated, &lastFetched,
		&etag, &lastModified,
		&fetchError, &feed.FetchErrorCount,
		&nextFetch, &active, &feed.FetchInterval,
	)

	if err != nil {
		return err
	}

	// Handle NULL strings
	if title.Valid {
		feed.Title = title.String
	}
	if link.Valid {
		feed.Link = link.String
	}
	if etag.Valid {
		feed.ETag = etag.String
	}
	if lastModified.Valid {
		feed.LastModified = lastModified.String
	}
	if fetchError.Valid {
		feed.FetchError = fetchError.String
	}
	if active.Valid {
		feed.Active = active.Int64 == 1
	}

	// Parse times
	if updated.Valid {
		var err error
		feed.Updated, err = time.Parse(time.RFC3339, updated.String)
		if err != nil {
			return fmt.Errorf("invalid updated timestamp %q: %w", updated.String, err)
		}
	}
	if lastFetched.Valid {
		var err error
		feed.LastFetched, err = time.Parse(time.RFC3339, lastFetched.String)
		if err != nil {
			return fmt.Errorf("invalid last_fetched timestamp %q: %w", lastFetched.String, err)
		}
	}
	if nextFetch.Valid {
		var err error
		feed.NextFetch, err = time.Parse(time.RFC3339, nextFetch.String)
		if err != nil {
			return fmt.Errorf("invalid next_fetch timestamp %q: %w", nextFetch.String, err)
		}
	}

	return nil
}

func scanFeeds(rows *sql.Rows) ([]Feed, error) {
	var feeds []Feed

	for rows.Next() {
		var feed Feed
		if err := scanFeed(rows, &feed); err != nil {
			return nil, err
		}
		feeds = append(feeds, feed)
	}

	return feeds, rows.Err()
}

func scanEntries(rows *sql.Rows) ([]Entry, error) {
	var entries []Entry

	for rows.Next() {
		var entry Entry
		var title, link, author, content, contentType, summary sql.NullString
		var published, updated, firstSeen string

		err := rows.Scan(
			&entry.ID, &entry.FeedID, &entry.EntryID,
			&title, &link, &author,
			&published, &updated,
			&content, &contentType, &summary,
			&firstSeen,
		)

		if err != nil {
			return nil, err
		}

		// Handle NULL strings
		if title.Valid {
			entry.Title = title.String
		}
		if link.Valid {
			entry.Link = link.String
		}
		if author.Valid {
			entry.Author = author.String
		}
		if content.Valid {
			entry.Content = content.String
		}
		if contentType.Valid {
			entry.ContentType = contentType.String
		}
		if summary.Valid {
			entry.Summary = summary.String
		}

		// Parse times (required fields in database)
		entry.Published, err = time.Parse(time.RFC3339, published)
		if err != nil {
			return nil, fmt.Errorf("invalid published timestamp %q for entry %s: %w", published, entry.EntryID, err)
		}
		entry.Updated, err = time.Parse(time.RFC3339, updated)
		if err != nil {
			return nil, fmt.Errorf("invalid updated timestamp %q for entry %s: %w", updated, entry.EntryID, err)
		}
		entry.FirstSeen, err = time.Parse(time.RFC3339, firstSeen)
		if err != nil {
			return nil, fmt.Errorf("invalid first_seen timestamp %q for entry %s: %w", firstSeen, entry.EntryID, err)
		}

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}
