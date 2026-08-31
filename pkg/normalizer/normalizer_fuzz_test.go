package normalizer

import (
	"reflect"
	"testing"
	"time"
)

func FuzzParseFeed(f *testing.F) {
	seeds := [][]byte{
		[]byte(`<?xml version="1.0"?><rss version="2.0"><channel><title>RSS</title><item><guid>one</guid><title>Entry</title><description><![CDATA[<p>safe</p>]]></description></item></channel></rss>`),
		[]byte(`<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"><title>Atom</title><entry><id>one</id><title>Entry</title><updated>2026-01-02T03:04:05Z</updated><content type="html">&lt;p&gt;safe&lt;/p&gt;</content></entry></feed>`),
		[]byte(`{"version":"https://jsonfeed.org/version/1.1","title":"JSON","items":[{"id":"one","content_html":"<p>safe</p>"}]}`),
		[]byte(`<?xml version="1.0"?><rss version="2.0"><channel><title>XSS</title><item><guid>x</guid><description><![CDATA[<script>alert(1)</script><a href="javascript:alert(2)">link</a>]]></description></item></channel></rss>`),
		[]byte(`not a feed`),
		{},
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	const maxInputBytes = 128 << 10
	fetchTime := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > maxInputBytes {
			t.Skip()
		}

		metadata, entries, err := New().Parse(t.Context(), data, "https://example.com/feed", fetchTime)
		if err != nil {
			return
		}
		if metadata == nil {
			t.Fatal("successful parse returned nil metadata")
		}

		metadataAgain, entriesAgain, err := New().Parse(t.Context(), data, "https://example.com/feed", fetchTime)
		if err != nil {
			t.Fatalf("same input failed on second parse: %v", err)
		}
		if !reflect.DeepEqual(metadata, metadataAgain) || !reflect.DeepEqual(entries, entriesAgain) {
			t.Fatal("normalization is not deterministic for a fixed input and fetch time")
		}

		outputBytes := len(metadata.Title) + len(metadata.Link)
		sanitizer := New()
		for i, entry := range entries {
			if entry.ID == "" {
				t.Fatalf("entry %d has no stable identity", i)
			}
			if entry.Published.IsZero() || entry.Updated.IsZero() || !entry.FirstSeen.Equal(fetchTime) {
				t.Fatalf("entry %d has invalid normalized timestamps", i)
			}
			if entry.ContentType != "" && entry.ContentType != "html" {
				t.Fatalf("entry %d has unsupported content type %q", i, entry.ContentType)
			}
			if sanitizer.SanitizeHTML(entry.Content) != entry.Content {
				t.Fatalf("entry %d content is not stable under sanitization", i)
			}
			if sanitizer.SanitizeHTML(entry.Summary) != entry.Summary {
				t.Fatalf("entry %d summary is not stable under sanitization", i)
			}
			outputBytes += len(entry.ID) + len(entry.Title) + len(entry.Link) + len(entry.Author)
			outputBytes += len(entry.Content) + len(entry.ContentType) + len(entry.Summary)
		}

		// Parsed output should remain proportional to the untrusted input. The
		// allowance covers generated IDs and metadata when the feed omits them.
		if outputBytes > 16*len(data)+4096 {
			t.Fatalf("normalized output grew unexpectedly: input=%d output=%d", len(data), outputBytes)
		}
	})
}
