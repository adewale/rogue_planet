package normalizer

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

func FuzzParseFeed(f *testing.F) {
	seeds := [][]byte{
		[]byte(`<?xml version="1.0"?><rss version="2.0"><channel><title>RSS</title><item><guid>one</guid><title>Entry</title><description><![CDATA[<p>safe</p>]]></description></item></channel></rss>`),
		[]byte(`<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"><title>Atom</title><entry><id>one</id><title>Entry</title><updated>2026-01-02T03:04:05Z</updated><content type="html">&lt;p&gt;safe&lt;/p&gt;</content></entry></feed>`),
		[]byte(`{"version":"https://jsonfeed.org/version/1.1","title":"JSON","items":[{"id":"one","content_html":"<p>safe</p>"}]}`),
		[]byte(`<?xml version="1.0"?><rss version="2.0"><channel><title>XSS</title><item><guid>x</guid><description><![CDATA[<script>alert(1)</script><a href="javascript:alert(2)">link</a>]]></description></item></channel></rss>`),
		[]byte(`not a feed`),
		{},
		[]byte(`<rss version="2.0"><channel><title>Deep</title><item><guid>deep</guid><description><![CDATA[` + strings.Repeat("<div>", 600) + "safe" + strings.Repeat("</div>", 600) + `]]></description></item></channel></rss>`),
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
		// Known-valid seeds must not silently become rejected or empty feeds.
		// Arbitrary malformed mutations may still legitimately fail to parse.
		for i, text := range []string{"safe", "safe", "safe", "link"} {
			if bytes.Equal(data, seeds[i]) {
				if err != nil || metadata == nil || len(entries) != 1 {
					t.Fatalf("valid seed %d lost its entry: metadata=%v entries=%d err=%v", i, metadata, len(entries), err)
				}
				if !strings.Contains(entries[0].Content, text) {
					t.Fatalf("valid seed %d lost benign content %q", i, text)
				}
			}
		}
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
			for _, field := range []struct{ name, value string }{{"content", entry.Content}, {"summary", entry.Summary}} {
				if err := checkFeedHTML(field.value); err != nil {
					t.Fatalf("entry %d unsafe %s: %v", i, field.name, err)
				}
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

// A fixed point is not a security oracle: an unsafe policy can preserve script
// on both passes. Inspect HTML tokens independently of the sanitizer policy.
// This catches specified active-markup regressions, not every browser XSS vector.
func checkFeedHTML(markup string) error {
	// Tokenize all output, including head markup. Building an AST would reject
	// harmless >512-deep nesting admitted by the production sanitizer.
	tokens := html.NewTokenizer(strings.NewReader(markup))
	for {
		switch tokens.Next() {
		case html.ErrorToken:
			if err := tokens.Err(); err != io.EOF {
				return err
			}
			return nil
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokens.Token()
			switch token.Data {
			case "script", "iframe", "frame", "frameset", "object", "embed", "applet", "base", "meta", "link", "style", "form", "svg", "math":
				return fmt.Errorf("active element <%s>", token.Data)
			}
			for _, attr := range token.Attr {
				key := strings.ToLower(attr.Key)
				if strings.HasPrefix(key, "on") || key == "style" || key == "srcdoc" {
					return fmt.Errorf("active attribute %s", key)
				}
				switch key {
				case "href", "src", "action", "formaction", "background", "dynsrc", "lowsrc", "poster":
					// Tokenization decodes entities; browsers also ignore tabs and
					// newlines in URL schemes and trim leading control characters.
					value := strings.Map(func(r rune) rune {
						if r == '\t' || r == '\n' || r == '\r' {
							return -1
						}
						return r
					}, strings.TrimFunc(attr.Val, func(r rune) bool { return r <= ' ' }))
					u, err := url.Parse(value)
					if err != nil {
						return fmt.Errorf("invalid %s URL: %w", key, err)
					}
					switch strings.ToLower(u.Scheme) {
					case "", "http", "https", "mailto": // Existing UGC policy permits mailto.
					default:
						return fmt.Errorf("unsafe %s URL scheme %q", key, u.Scheme)
					}
				}
			}
		}
	}
}

func TestFeedHTMLOracle(t *testing.T) {
	for _, markup := range []string{
		`<script>alert(1)</script>`,
		`<head><meta http-equiv="refresh" content="0;url=https://evil.example"></head>`,
		`<p onclick="alert(1)">text</p>`,
		`<a href="java&#x09;script:alert(1)">link</a>`,
		`<a href="&#x01;javascript:alert(1)">link</a>`,
		`<svg><a xlink:href="javascript:alert(1)">link</a></svg>`,
		strings.Repeat("<div>", 600) + `<script>alert(1)</script>` + strings.Repeat("</div>", 600),
	} {
		if err := checkFeedHTML(markup); err == nil {
			t.Errorf("oracle accepted active markup %q", markup)
		}
	}
	for _, markup := range []string{
		`<p>&lt;script&gt; is text, not executable markup</p>`,
		`<a href="/relative">relative</a><a href="https://example.com">secure</a>`,
		`<a href="mailto:person@example.com">email</a>`,
		strings.Repeat("<div>", 600) + "safe" + strings.Repeat("</div>", 600),
		``,
	} {
		if err := checkFeedHTML(markup); err != nil {
			t.Errorf("oracle rejected benign markup %q: %v", markup, err)
		}
	}
}
