package normalizer

import (
	"time"
)

// FeedNormalizer defines the interface for feed parsing and content normalization.
// This interface enables dependency injection and makes testing easier by allowing
// mock implementations to be used in place of the concrete Normalizer.
type FeedNormalizer interface {
	// Parse parses and normalizes a feed from raw bytes
	// Returns feed metadata, normalized entries, and any parsing errors
	Parse(feedData []byte, feedURL string, fetchTime time.Time) (*FeedMetadata, []Entry, error)
}

// Ensure Normalizer implements FeedNormalizer interface
var _ FeedNormalizer = (*Normalizer)(nil)
