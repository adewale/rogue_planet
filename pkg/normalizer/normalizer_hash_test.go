package normalizer

import (
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

// Tests for L13: Full SHA256 hash and cross-feed collision prevention

func TestExtractID_FullSHA256Hash(t *testing.T) {
	t.Parallel()
	n := New()

	// When generating a hash-based ID from title, it should use the full SHA256 (64 hex chars)
	item := &gofeed.Item{
		Title: "Test Title Without GUID or Link",
	}

	id := n.extractID(item, "https://example.com/feed")

	// Full SHA256 produces 64 hex characters
	if len(id) != 64 {
		t.Errorf("Generated ID length = %d, want 64 (full SHA256 hex), got %q", len(id), id)
	}
}

func TestExtractID_ContentHashFullLength(t *testing.T) {
	t.Parallel()
	n := New()

	// Hash from content (last resort) should also be full SHA256
	item := &gofeed.Item{
		Description: "Some content",
		Content:     "More content",
	}

	id := n.extractID(item, "https://example.com/feed")

	if len(id) != 64 {
		t.Errorf("Content hash ID length = %d, want 64, got %q", len(id), id)
	}
}

func TestExtractID_CrossFeedCollisionPrevention(t *testing.T) {
	t.Parallel()
	n := New()

	// Entries with identical titles from different feeds should get different IDs
	item := &gofeed.Item{
		Title: "Same Title",
	}

	id1 := n.extractID(item, "https://feed1.example.com/feed")
	id2 := n.extractID(item, "https://feed2.example.com/feed")

	if id1 == id2 {
		t.Errorf("IDs should differ for same title from different feeds: both = %q", id1)
	}
}

func TestExtractID_EmptyContentIncludesLinkAndDate(t *testing.T) {
	t.Parallel()
	n := New()

	// Entries with empty content but different links should get different IDs
	pubTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	item1 := &gofeed.Item{
		Description:     "",
		Content:         "",
		Link:            "https://example.com/post1",
		PublishedParsed: &pubTime,
	}

	item2 := &gofeed.Item{
		Description:     "",
		Content:         "",
		Link:            "https://example.com/post2",
		PublishedParsed: &pubTime,
	}

	// Note: both items have no GUID so they fall back to link as ID
	// But if both had no link either, the content hash would need to differ.
	// Let's test that scenario:
	item3 := &gofeed.Item{
		Description: "",
		Content:     "",
	}

	item4 := &gofeed.Item{
		Description: "",
		Content:     "",
	}

	// Items with truly empty content from the same feed get the same ID (expected),
	// but from different feeds should differ
	id3 := n.extractID(item3, "https://feed1.example.com/feed")
	id4 := n.extractID(item4, "https://feed2.example.com/feed")

	if id3 == id4 {
		t.Errorf("Empty content entries from different feeds should get different IDs: both = %q", id3)
	}

	// Items with no GUID but with a link should use the link
	id1 := n.extractID(item1, "https://example.com/feed")
	id2 := n.extractID(item2, "https://example.com/feed")

	// These use link as ID, so they should differ
	if id1 == id2 {
		t.Errorf("Items with different links should get different IDs")
	}
}

func TestExtractID_EmptyContentSameFeedCollision(t *testing.T) {
	t.Parallel()
	n := New()

	// Two entries with empty content from the same feed but with different
	// link and date should get different IDs via the hash that includes link+date
	pubTime1 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	pubTime2 := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	item1 := &gofeed.Item{
		Description:     "",
		Content:         "",
		PublishedParsed: &pubTime1,
	}

	item2 := &gofeed.Item{
		Description:     "",
		Content:         "",
		PublishedParsed: &pubTime2,
	}

	id1 := n.extractID(item1, "https://example.com/feed")
	id2 := n.extractID(item2, "https://example.com/feed")

	if id1 == id2 {
		t.Errorf("Empty content entries with different dates should get different IDs: both = %q", id1)
	}
}
