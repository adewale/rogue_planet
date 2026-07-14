package opml

import (
	"testing"
)

// FuzzOPMLParse feeds random bytes to opml.Parse().
// It should not panic on any input.
func FuzzOPMLParse(f *testing.F) {
	// Add seed corpus with valid and malformed OPML
	f.Add([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head><title>Test</title></head>
  <body>
    <outline text="Feed" xmlUrl="https://example.com/feed"/>
  </body>
</opml>`))

	f.Add([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <head><title>OPML 1.0</title></head>
  <body>
    <outline text="Feed" url="https://example.com/feed"/>
  </body>
</opml>`))

	f.Add([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <body>
    <outline text="Category">
      <outline text="Nested Feed" xmlUrl="https://example.com/nested"/>
    </outline>
  </body>
</opml>`))

	// Malformed inputs
	f.Add([]byte(`not xml at all`))
	f.Add([]byte(`<opml><body><outline/></body></opml>`))
	f.Add([]byte(`<?xml version="1.0"?><opml></opml>`))
	f.Add([]byte{})
	f.Add([]byte(`<opml version="2.0"><head><title>` + string(make([]byte, 1000)) + `</title></head><body></body></opml>`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Should not panic on any input
		result, err := Parse(data)
		if err != nil {
			return // Parse errors are expected for random input
		}

		// If parsing succeeded, extraction should not panic
		if result != nil {
			_ = result.ExtractFeeds()
		}
	})
}
