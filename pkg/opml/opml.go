// Package opml provides OPML (Outline Processor Markup Language) parsing and generation.
//
// The package supports OPML 1.0 and 2.0 formats for importing and exporting feed lists.
// It handles both text/title attribute variations and xmlUrl/url variations for maximum
// compatibility with different OPML readers and writers.
package opml

import (
	"encoding/xml"
	"fmt"
	"os"
	"time"

	"github.com/adewale/rogue_planet/pkg/timeprovider"
)

// OPML represents the root OPML structure
type OPML struct {
	XMLName xml.Name `xml:"opml"`
	Version string   `xml:"version,attr"`
	Head    Head     `xml:"head"`
	Body    Body     `xml:"body"`
}

// Head contains metadata
type Head struct {
	Title       string `xml:"title"`
	DateCreated string `xml:"dateCreated,omitempty"` // RFC 822 format per OPML spec
	OwnerName   string `xml:"ownerName,omitempty"`
	OwnerEmail  string `xml:"ownerEmail,omitempty"`
}

// Body contains outlines (feeds)
type Body struct {
	Outlines []Outline `xml:"outline"`
}

// Outline represents a feed or category
type Outline struct {
	Text    string `xml:"text,attr"`              // Required by OPML spec
	Title   string `xml:"title,attr,omitempty"`   // CRITICAL: Many readers expect this
	Type    string `xml:"type,attr,omitempty"`    // "rss", "atom", etc.
	XMLUrl  string `xml:"xmlUrl,attr,omitempty"`  // Feed URL (OPML 2.0)
	Url     string `xml:"url,attr,omitempty"`     // Feed URL (OPML 1.0 compatibility)
	HTMLUrl string `xml:"htmlUrl,attr,omitempty"` // Website URL

	// For nested categories
	Outlines []Outline `xml:"outline,omitempty"`
}

// Feed represents an extracted feed
type Feed struct {
	Title   string
	FeedURL string
	WebURL  string
}

// Metadata for OPML generation
type Metadata struct {
	Title      string
	OwnerName  string
	OwnerEmail string
}

// Parse parses an OPML file from bytes
func Parse(data []byte) (*OPML, error) {
	var opml OPML
	if err := xml.Unmarshal(data, &opml); err != nil {
		return nil, fmt.Errorf("parse OPML: %w", err)
	}
	return &opml, nil
}

// ParseFile parses an OPML file from disk
func ParseFile(path string) (*OPML, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return Parse(data)
}

// ExtractFeeds extracts all feed URLs from OPML (flattens nested outlines)
func (o *OPML) ExtractFeeds() []Feed {
	feeds := []Feed{}
	o.extractOutlines(o.Body.Outlines, &feeds)
	return feeds
}

// extractOutlines recursively extracts feeds from outlines
func (o *OPML) extractOutlines(outlines []Outline, feeds *[]Feed) {
	for _, outline := range outlines {
		// Get feed URL (try xmlUrl first, then url for OPML 1.0 compatibility)
		feedURL := outline.XMLUrl
		if feedURL == "" {
			feedURL = outline.Url
		}

		// If we have a feed URL, extract it
		if feedURL != "" {
			// Prefer title, fall back to text
			title := outline.Title
			if title == "" {
				title = outline.Text
			}

			*feeds = append(*feeds, Feed{
				Title:   title,
				FeedURL: feedURL,
				WebURL:  outline.HTMLUrl,
			})
		}

		// Recursively process nested outlines
		if len(outline.Outlines) > 0 {
			o.extractOutlines(outline.Outlines, feeds)
		}
	}
}

// Generate creates OPML from feed list using the current system time
func Generate(feeds []Feed, metadata Metadata) (*OPML, error) {
	return GenerateWithTimeProvider(feeds, metadata, timeprovider.WallClock{})
}

// GenerateWithTimeProvider creates OPML from feed list with a custom TimeProvider.
// This is primarily for testing with FakeClock.
func GenerateWithTimeProvider(feeds []Feed, metadata Metadata, tp timeprovider.TimeProvider) (*OPML, error) {
	outlines := make([]Outline, 0, len(feeds))

	for _, feed := range feeds {
		// Set both text and title for maximum compatibility
		title := feed.Title
		if title == "" {
			title = feed.FeedURL
		}

		outlines = append(outlines, Outline{
			Text:    title,
			Title:   title,
			Type:    "rss",
			XMLUrl:  feed.FeedURL,
			HTMLUrl: feed.WebURL,
		})
	}

	opml := &OPML{
		Version: "2.0",
		Head: Head{
			Title:       metadata.Title,
			DateCreated: FormatRFC822(tp.Now()),
			OwnerName:   metadata.OwnerName,
			OwnerEmail:  metadata.OwnerEmail,
		},
		Body: Body{
			Outlines: outlines,
		},
	}

	return opml, nil
}

// Marshal serializes OPML to XML bytes
func (o *OPML) Marshal() ([]byte, error) {
	output, err := xml.MarshalIndent(o, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal OPML: %w", err)
	}

	// Add XML declaration
	result := []byte(xml.Header + string(output))
	return result, nil
}

// Write writes OPML to file
func (o *OPML) Write(path string) error {
	data, err := o.Marshal()
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// FormatRFC822 converts time.Time to RFC 822 format (OPML spec requirement)
// Example: "Mon, 02 Jan 2006 15:04:05 -0700"
func FormatRFC822(t time.Time) string {
	return t.Format(time.RFC1123Z)
}

// ParseRFC822 parses RFC 822 date string to time.Time
func ParseRFC822(s string) (time.Time, error) {
	// Try RFC1123Z first (with numeric zone)
	t, err := time.Parse(time.RFC1123Z, s)
	if err == nil {
		return t, nil
	}

	// Try RFC1123 (with named zone like "EST")
	t, err = time.Parse(time.RFC1123, s)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("parse RFC 822 date: %w", err)
}
