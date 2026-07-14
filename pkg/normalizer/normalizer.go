// Package normalizer provides feed parsing and HTML sanitization for RSS/Atom feeds.
//
// The normalizer parses multiple feed formats (RSS 1.0, RSS 2.0, Atom, JSON Feed)
// and converts them to a canonical internal format. It implements HTML sanitization
// to prevent XSS attacks (CVE-2009-2937), handles missing dates and IDs gracefully,
// and resolves relative URLs to absolute.
package normalizer

import (
	"bytes"
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
	"golang.org/x/net/html"
)

var (
	ErrInvalidFeed = errors.New("invalid feed data")
	ErrNoEntries   = errors.New("feed contains no entries")
)

const (
	// MaxEntryContentSize is the maximum size in bytes for a single entry's content
	MaxEntryContentSize = 1024 * 1024 // 1MB
	// MaxEntriesPerFeed is the maximum number of entries to process from a single feed
	MaxEntriesPerFeed = 500
	// FutureDateTolerance is the maximum amount of time in the future a date can be
	// before it is clamped to fetchTime. Dates more than this duration in the future
	// relative to fetchTime are clamped.
	FutureDateTolerance = 1 * time.Hour
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

	// Block mailto: scheme which UGCPolicy allows by default via AllowStandardURLs().
	// The spec requires "Only allow http/https URL schemes."
	policy.AllowURLSchemeWithCustomPolicy("mailto", func(u *url.URL) bool {
		return false // Reject all mailto: URLs
	})

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

	// Limit number of entries to prevent memory exhaustion
	items := feed.Items
	if len(items) > MaxEntriesPerFeed {
		items = items[:MaxEntriesPerFeed]
	}

	entries := make([]Entry, 0, len(items))
	for _, item := range items {
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
	entry.Title = n.sanitizeTitle(item.Title)

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

	// Extract dates and clamp future dates
	entry.Published = n.extractPublished(item, feed, fetchTime)
	entry.Published = clampFutureDate(entry.Published, fetchTime)
	entry.Updated = n.extractUpdated(item, entry.Published)
	entry.Updated = clampFutureDate(entry.Updated, fetchTime)

	// Extract content (prefer full content over summary)
	// Truncate before sanitization to prevent memory exhaustion
	if item.Content != "" {
		content := item.Content
		if len(content) > MaxEntryContentSize {
			content = content[:MaxEntryContentSize]
		}
		entry.Content = n.sanitizeHTML(content, feedURL)
		entry.ContentType = "html"
	} else if item.Description != "" {
		desc := item.Description
		if len(desc) > MaxEntryContentSize {
			desc = desc[:MaxEntryContentSize]
		}
		entry.Content = n.sanitizeHTML(desc, feedURL)
		entry.ContentType = "html"
	}

	// Extract summary
	if item.Description != "" && item.Content != "" {
		desc := item.Description
		if len(desc) > MaxEntryContentSize {
			desc = desc[:MaxEntryContentSize]
		}
		entry.Summary = n.sanitizeHTML(desc, feedURL)
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
		return hex.EncodeToString(hash.Sum(nil))
	}

	// Last resort: hash of content + link + date to prevent collisions
	hash := sha256.New()
	hash.Write([]byte(feedURL))
	hash.Write([]byte(item.Description))
	hash.Write([]byte(item.Content))
	hash.Write([]byte(item.Link))
	if item.PublishedParsed != nil {
		hash.Write([]byte(item.PublishedParsed.String()))
	}
	return hex.EncodeToString(hash.Sum(nil))
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

// clampFutureDate returns fetchTime if the given date is more than
// FutureDateTolerance in the future relative to fetchTime.
// This prevents content ordering manipulation via far-future dates.
func clampFutureDate(date time.Time, fetchTime time.Time) time.Time {
	if date.After(fetchTime.Add(FutureDateTolerance)) {
		return fetchTime
	}
	return date
}

// sanitizeHTML sanitizes HTML content and resolves relative URLs
func (n *Normalizer) sanitizeHTML(rawHTML string, baseURL string) string {
	// Sanitize HTML to remove dangerous content
	sanitized := n.sanitizer.Sanitize(rawHTML)

	// Resolve relative URLs after sanitization
	sanitized = resolveRelativeURLs(sanitized, baseURL)

	return strings.TrimSpace(sanitized)
}

// resolveRelativeURLs parses HTML content, finds src and href attributes,
// and resolves relative URLs to absolute using the provided base URL.
// If baseURL is empty or invalid, the original HTML is returned unchanged.
// If the HTML is malformed, it attempts best-effort resolution.
func resolveRelativeURLs(htmlContent string, baseURL string) string {
	if baseURL == "" {
		return htmlContent
	}

	base, err := url.Parse(baseURL)
	if err != nil || base.Scheme == "" {
		return htmlContent
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	// Walk the HTML tree and resolve relative URLs in href and src attributes
	var resolve func(*html.Node)
	resolve = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i, attr := range n.Attr {
				if attr.Key == "href" || attr.Key == "src" {
					if attr.Val == "" {
						continue
					}
					ref, err := url.Parse(attr.Val)
					if err != nil {
						continue
					}
					// Only resolve if the URL is relative (no scheme)
					if ref.Scheme == "" {
						n.Attr[i].Val = base.ResolveReference(ref).String()
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			resolve(c)
		}
	}
	resolve(doc)

	// Re-serialize the HTML
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return htmlContent
	}

	// html.Parse wraps content in <html><head></head><body>...</body></html>
	// We need to extract just the body content
	rendered := buf.String()
	bodyStart := strings.Index(rendered, "<body>")
	bodyEnd := strings.LastIndex(rendered, "</body>")
	if bodyStart >= 0 && bodyEnd > bodyStart {
		rendered = rendered[bodyStart+len("<body>") : bodyEnd]
	}

	return rendered
}

// sanitizeTitle strips all HTML tags from titles.
// Titles should be plain text, not HTML. This prevents XSS when
// titles are cast to template.HTML for rendering.
func (n *Normalizer) sanitizeTitle(title string) string {
	// Use bluemonday StrictPolicy to strip ALL HTML tags
	strict := bluemonday.StrictPolicy()
	cleaned := strict.Sanitize(title)
	return strings.TrimSpace(cleaned)
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
