package normalizer

import (
	"context"
	"os"
	"testing"
	"time"
)

// FuzzNormalizerParse feeds random bytes to the normalizer's Parse() method.
// It should not panic on any input.
func FuzzNormalizerParse(f *testing.F) {
	// Add seed corpus from existing testdata files
	seedFiles := []string{
		"../../testdata/daringfireball-feed.xml",
		"../../testdata/asymco-feed.xml",
		"../../testdata/test-feed.xml",
		"../../testdata/jsonfeed-1.0.json",
		"../../testdata/jsonfeed-1.1.json",
		"../../testdata/jsonfeed-edge-cases.json",
	}

	for _, path := range seedFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			continue // Skip missing files
		}
		f.Add(data)
	}

	// Add some basic seed inputs
	f.Add([]byte(`<?xml version="1.0"?><rss version="2.0"><channel><title>T</title><item><title>I</title></item></channel></rss>`))
	f.Add([]byte(`<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"><title>T</title><entry><id>1</id><title>E</title></entry></feed>`))
	f.Add([]byte(`{"version":"https://jsonfeed.org/version/1","title":"T","items":[{"id":"1","content_text":"C"}]}`))
	f.Add([]byte(`not a valid feed at all`))
	f.Add([]byte(`<html><body>not a feed</body></html>`))
	f.Add([]byte{})

	n := New()
	fetchTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	f.Fuzz(func(t *testing.T, data []byte) {
		// Should not panic on any input
		_, _, _ = n.Parse(context.Background(), data, "https://example.com/feed", fetchTime)
	})
}
