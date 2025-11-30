// Package normalizer provides feed parsing and HTML sanitization for RSS/Atom feeds.
//
// The normalizer parses multiple feed formats (RSS 1.0, RSS 2.0, Atom, JSON Feed)
// and converts them to a canonical internal format. It implements HTML sanitization
// to prevent XSS attacks (CVE-2009-2937), handles missing dates and IDs gracefully,
// and resolves relative URLs to absolute.
package normalizer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
)

// GeneratedIDLength is the length of generated entry IDs (hex characters).
// Using 16 hex chars (64 bits) from SHA256 provides sufficient uniqueness
// while keeping IDs reasonably short.
const GeneratedIDLength = 16

var (
	ErrInvalidFeed = errors.New("invalid feed data")
	ErrNoEntries   = errors.New("feed contains no entries")
)

// Entry represents a normalized feed entry
type Entry struct {
	ID          string // Unique ID (GUID or generated)
	Title       string
	Link        string // Permalink to original article
	Author      string
	Published   time.Time // RFC 3339 timestamp
	Updated     time.Time // RFC 3339 timestamp
	Content     string    // Sanitized HTML content
	ContentType string    // "html" or "text"
	Summary     string    // Sanitized summary
	FirstSeen   time.Time // When first crawled
}

// FeedMetadata contains feed-level information
type FeedMetadata struct {
	Title   string
	Link    string
	Updated time.Time
}

// Normalizer handles feed parsing and content normalization
type Normalizer struct {
	parser    *gofeed.Parser
	sanitizer *bluemonday.Policy
}

// New creates a new Normalizer with default settings
func New() *Normalizer {
	// Create strict sanitization policy
	policy := bluemonday.UGCPolicy()

	// Only allow http and https schemes
	policy.AllowURLSchemes("http", "https")

	// Additional safe attributes
	policy.AllowAttrs("alt", "title").OnElements("img")
	policy.AllowAttrs("href", "title").OnElements("a")

	return &Normalizer{
		parser:    gofeed.NewParser(),
		sanitizer: policy,
	}
}

// Parse parses and normalizes a feed
func (n *Normalizer) Parse(ctx context.Context, feedData []byte, feedURL string, fetchTime time.Time) (*FeedMetadata, []Entry, error) {
	// Check context before expensive parsing
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	// Parse feed
	feed, err := n.parser.ParseString(string(feedData))
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidFeed, err)
	}

	// Extract feed metadata
	metadata := FeedMetadata{
		Title: feed.Title,
		Link:  feed.Link,
	}

	if feed.UpdatedParsed != nil {
		metadata.Updated = *feed.UpdatedParsed
	} else {
		metadata.Updated = fetchTime
	}

	// Normalize entries
	if len(feed.Items) == 0 {
		return &metadata, []Entry{}, nil
	}

	entries := make([]Entry, 0, len(feed.Items))
	for _, item := range feed.Items {
		entry, err := n.normalizeEntry(item, feed, feedURL, fetchTime)
		if err != nil {
			// Log error but continue processing other entries
			continue
		}
		entries = append(entries, entry)
	}

	return &metadata, entries, nil
}

// normalizeEntry converts a feed item to a normalized Entry
func (n *Normalizer) normalizeEntry(item *gofeed.Item, feed *gofeed.Feed, feedURL string, fetchTime time.Time) (Entry, error) {
	entry := Entry{
		FirstSeen: fetchTime,
	}

	// Extract ID (or generate one)
	entry.ID = n.extractID(item, feedURL)

	// Extract title
	entry.Title = strings.TrimSpace(item.Title)

	// Extract link and resolve to absolute URL
	if item.Link != "" {
		absURL, err := n.resolveURL(item.Link, feedURL)
		if err == nil {
			entry.Link = absURL
		} else {
			entry.Link = item.Link // Use as-is if resolution fails
		}
	}

	// Extract author
	entry.Author = n.extractAuthor(item, feed)

	// Extract dates
	entry.Published = n.extractPublished(item, feed, fetchTime)
	entry.Updated = n.extractUpdated(item, entry.Published)

	// Extract content (prefer full content over summary)
	if item.Content != "" {
		entry.Content = n.sanitizeHTML(item.Content, feedURL)
		entry.ContentType = "html"
	} else if item.Description != "" {
		entry.Content = n.sanitizeHTML(item.Description, feedURL)
		entry.ContentType = "html"
	}

	// Extract summary
	if item.Description != "" && item.Content != "" {
		entry.Summary = n.sanitizeHTML(item.Description, feedURL)
	}

	return entry, nil
}

// extractID generates or extracts a unique ID for an entry
func (n *Normalizer) extractID(item *gofeed.Item, feedURL string) string {
	// Use existing GUID if present
	if item.GUID != "" {
		return item.GUID
	}

	// Fallback to link
	if item.Link != "" {
		return item.Link
	}

	// Fallback to hash of title + date
	if item.Title != "" {
		hash := sha256.New()
		hash.Write([]byte(feedURL))
		hash.Write([]byte(item.Title))
		if item.PublishedParsed != nil {
			hash.Write([]byte(item.PublishedParsed.String()))
		}
		return hex.EncodeToString(hash.Sum(nil))[:GeneratedIDLength]
	}

	// Last resort: hash of content
	hash := sha256.New()
	hash.Write([]byte(feedURL))
	hash.Write([]byte(item.Description))
	hash.Write([]byte(item.Content))
	return hex.EncodeToString(hash.Sum(nil))[:GeneratedIDLength]
}

// extractAuthor gets the author name from entry or feed level
func (n *Normalizer) extractAuthor(item *gofeed.Item, feed *gofeed.Feed) string {
	// Try item-level author
	if item.Author != nil && item.Author.Name != "" {
		return item.Author.Name
	}

	// Try multiple authors
	if len(item.Authors) > 0 && item.Authors[0].Name != "" {
		return item.Authors[0].Name
	}

	// Fallback to feed-level author
	if feed.Author != nil && feed.Author.Name != "" {
		return feed.Author.Name
	}

	return ""
}

// extractPublished extracts the published date
func (n *Normalizer) extractPublished(item *gofeed.Item, feed *gofeed.Feed, fetchTime time.Time) time.Time {
	// Use item published date
	if item.PublishedParsed != nil && !item.PublishedParsed.IsZero() {
		return *item.PublishedParsed
	}

	// Use item updated date
	if item.UpdatedParsed != nil && !item.UpdatedParsed.IsZero() {
		return *item.UpdatedParsed
	}

	// Use feed updated date
	if feed.UpdatedParsed != nil && !feed.UpdatedParsed.IsZero() {
		return *feed.UpdatedParsed
	}

	// Use fetch time as last resort
	return fetchTime
}

// extractUpdated extracts the updated date
func (n *Normalizer) extractUpdated(item *gofeed.Item, published time.Time) time.Time {
	if item.UpdatedParsed != nil && !item.UpdatedParsed.IsZero() {
		return *item.UpdatedParsed
	}
	return published
}

// sanitizeHTML sanitizes HTML content and resolves relative URLs
func (n *Normalizer) sanitizeHTML(html string, baseURL string) string {
	// First, resolve relative URLs (simplified - real implementation would use html parser)
	// For now, just sanitize

	// Sanitize HTML to remove dangerous content
	sanitized := n.sanitizer.Sanitize(html)

	return strings.TrimSpace(sanitized)
}

// resolveURL converts a relative URL to absolute using the feed URL as base
func (n *Normalizer) resolveURL(href string, baseURL string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	ref, err := url.Parse(href)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(ref).String(), nil
}

// SanitizeHTML provides public access to HTML sanitization
func (n *Normalizer) SanitizeHTML(html string) string {
	return n.sanitizer.Sanitize(html)
}
