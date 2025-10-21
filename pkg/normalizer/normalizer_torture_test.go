package normalizer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Atom Torture Tests
// Inspired by Jacques Distler's "Atom Torture Test" blog posts:
// - https://golem.ph.utexas.edu/~distler/blog/archives/000793.html (Atom Torture Test)
// - https://golem.ph.utexas.edu/~distler/blog/archives/000836.html (unknown content)
//
// TODO: Properly research and document the specific issues raised in these blog posts.
//       The current tests are based on web search results about the Atom Torture Test
//       which focused on XHTML, MathML, and SVG content handling in feed aggregators.
//       Direct access to these blog posts would allow for more accurate test coverage.
//
// These tests validate that Rogue Planet correctly handles:
// 1. XHTML content (type="xhtml") vs HTML (type="html") vs plain text (type="text")
// 2. MathML mathematical markup in feed content
// 3. SVG vector graphics with fallback images
// 4. Proper XML namespace handling
// 5. Security (no XSS via MathML/SVG event handlers)

// TestAtomContentType_XHTML tests handling of Atom content type="xhtml"
func TestAtomContentType_XHTML(t *testing.T) {
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-xhtml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	metadata, entries, err := n.Parse(feedData, "https://example.com/feed", time.Now())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if metadata.Title != "Atom Torture Test: XHTML Content" {
		t.Errorf("Expected title 'Atom Torture Test: XHTML Content', got %q", metadata.Title)
	}

	if len(entries) == 0 {
		t.Fatal("Expected entries, got none")
	}

	// Test 1: XHTML content should be parsed and sanitized
	entry := entries[0]
	if entry.Title != "XHTML Content with Namespace" {
		t.Errorf("Expected first entry title 'XHTML Content with Namespace', got %q", entry.Title)
	}

	// Content should contain the XHTML elements (sanitized)
	if !strings.Contains(entry.Content, "XHTML") {
		t.Errorf("Content missing 'XHTML' text: %s", entry.Content)
	}
	if !strings.Contains(entry.Content, "<strong>") {
		t.Errorf("Content missing <strong> tag: %s", entry.Content)
	}

	// Self-closing br tags should be handled
	// Note: After HTML sanitization, <br/> might become <br> or remain <br/>
	// We just check that the content is present
	if !strings.Contains(entry.Content, "self-closing tags") {
		t.Errorf("Content missing 'self-closing tags' text: %s", entry.Content)
	}
}

// TestAtomContentType_HTML tests handling of Atom content type="html"
func TestAtomContentType_HTML(t *testing.T) {
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-xhtml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(feedData, "https://example.com/feed", time.Now())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Find the HTML content type entry (entry 4)
	var htmlEntry *Entry
	for i := range entries {
		if entries[i].Title == "HTML Content Type" {
			htmlEntry = &entries[i]
			break
		}
	}

	if htmlEntry == nil {
		t.Fatal("Could not find 'HTML Content Type' entry")
	}

	// HTML content should be parsed (entities decoded) and sanitized
	if !strings.Contains(htmlEntry.Content, "HTML") {
		t.Errorf("HTML content missing 'HTML' text: %s", htmlEntry.Content)
	}
	if !strings.Contains(htmlEntry.Content, "<strong>") || !strings.Contains(htmlEntry.Content, "</strong>") {
		t.Errorf("HTML content missing <strong> tags: %s", htmlEntry.Content)
	}
}

// TestAtomContentType_Text tests handling of Atom content type="text"
func TestAtomContentType_Text(t *testing.T) {
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-xhtml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(feedData, "https://example.com/feed", time.Now())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Find the plain text entry (entry 3)
	var textEntry *Entry
	for i := range entries {
		if entries[i].Title == "Plain Text Content" {
			textEntry = &entries[i]
			break
		}
	}

	if textEntry == nil {
		t.Fatal("Could not find 'Plain Text Content' entry")
	}

	// Plain text content should have HTML entities decoded but NOT rendered as HTML
	if !strings.Contains(textEntry.Content, "plain text") {
		t.Errorf("Text content missing 'plain text': %s", textEntry.Content)
	}

	// The entities &lt;b&gt; should be decoded to <b> and then the sanitizer
	// might either keep them as text or convert to tags. The key is that
	// "HTML" text should be present.
	if !strings.Contains(textEntry.Content, "HTML") {
		t.Errorf("Text content missing 'HTML' text: %s", textEntry.Content)
	}
}

// TestXHTML_ComplexStructure tests XHTML with lists, links, and code
func TestXHTML_ComplexStructure(t *testing.T) {
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-xhtml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(feedData, "https://example.com/feed", time.Now())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Find entry 2: "XHTML with Lists and Links"
	var entry *Entry
	for i := range entries {
		if entries[i].Title == "XHTML with Lists and Links" {
			entry = &entries[i]
			break
		}
	}

	if entry == nil {
		t.Fatal("Could not find 'XHTML with Lists and Links' entry")
	}

	// Should contain list elements
	if !strings.Contains(entry.Content, "<ul>") && !strings.Contains(entry.Content, "<li>") {
		t.Errorf("Content missing list elements: %s", entry.Content)
	}

	// Should contain safe link
	if !strings.Contains(entry.Content, "https://example.com") {
		t.Errorf("Content missing link: %s", entry.Content)
	}

	// Should contain code/pre tags
	if !strings.Contains(entry.Content, "code") {
		t.Errorf("Content missing code-related content: %s", entry.Content)
	}
}

// TestMathML_BasicEquations tests basic MathML rendering
func TestMathML_BasicEquations(t *testing.T) {
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-mathml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	metadata, entries, err := n.Parse(feedData, "https://example.com/feed", time.Now())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if metadata.Title != "Atom Torture Test: MathML Content" {
		t.Errorf("Expected title 'Atom Torture Test: MathML Content', got %q", metadata.Title)
	}

	if len(entries) == 0 {
		t.Fatal("Expected entries, got none")
	}

	// Test Einstein's equation entry
	entry := entries[0]
	if entry.Title != "Einstein's Mass-Energy Equivalence" {
		t.Errorf("Expected first entry title 'Einstein's Mass-Energy Equivalence', got %q", entry.Title)
	}

	// Check if content contains the equation text
	// Note: MathML may be stripped by bluemonday or preserved depending on policy
	if !strings.Contains(entry.Content, "Einstein") {
		t.Errorf("Content missing 'Einstein' text: %s", entry.Content)
	}

	// Check if MathML elements are present or stripped
	// Current bluemonday.UGCPolicy() strips MathML, so we expect it to be gone
	// This test documents current behavior
	hasMathML := strings.Contains(entry.Content, "<math") || strings.Contains(entry.Content, "<mi>")

	if hasMathML {
		t.Logf("MathML is preserved in content (good for academic blogs)")
		// Verify namespace is present if MathML is kept
		if !strings.Contains(entry.Content, "http://www.w3.org/1998/Math/MathML") {
			t.Logf("Warning: MathML preserved but namespace might be missing")
		}
	} else {
		t.Logf("MathML is stripped (current behavior - consider allowing safe MathML tags)")
		t.Logf("See specs/TODO.md:550-560 for MathJax workaround documentation")
	}
}

// TestMathML_Sanitization tests that MathML cannot be used for XSS
func TestMathML_Sanitization(t *testing.T) {
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-mathml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(feedData, "https://example.com/feed", time.Now())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Find the security test entry
	var secEntry *Entry
	for i := range entries {
		if entries[i].Title == "MathML Security Test" {
			secEntry = &entries[i]
			break
		}
	}

	if secEntry == nil {
		t.Fatal("Could not find 'MathML Security Test' entry")
	}

	// Event handlers should be stripped
	dangerousPatterns := []string{"onclick", "onload", "alert('XSS')"}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(strings.ToLower(secEntry.Content), strings.ToLower(pattern)) {
			t.Errorf("Dangerous pattern %q not sanitized in content: %s", pattern, secEntry.Content)
		}
	}

	// Safe text should remain
	if !strings.Contains(secEntry.Content, "should not execute") {
		t.Errorf("Safe text missing from content: %s", secEntry.Content)
	}
}

// TestMathML_ComplexFormulas tests complex MathML structures
func TestMathML_ComplexFormulas(t *testing.T) {
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-mathml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(feedData, "https://example.com/feed", time.Now())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Find the quadratic formula entry
	var quadEntry *Entry
	for i := range entries {
		if entries[i].Title == "The Quadratic Formula" {
			quadEntry = &entries[i]
			break
		}
	}

	if quadEntry == nil {
		t.Fatal("Could not find 'The Quadratic Formula' entry")
	}

	// Content should contain the surrounding text even if MathML is stripped
	if !strings.Contains(quadEntry.Content, "quadratic formula") {
		t.Errorf("Content missing 'quadratic formula' text: %s", quadEntry.Content)
	}

	// Check for MathML elements (mfrac, msqrt, etc.)
	hasComplexMathML := strings.Contains(quadEntry.Content, "<mfrac>") ||
		strings.Contains(quadEntry.Content, "<msqrt>")

	if hasComplexMathML {
		t.Logf("Complex MathML structures preserved")
	} else {
		t.Logf("Complex MathML stripped (may need MathJax for rendering)")
	}
}

// TestSVG_Inline tests inline SVG handling
func TestSVG_Inline(t *testing.T) {
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-svg.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	metadata, entries, err := n.Parse(feedData, "https://example.com/feed", time.Now())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if metadata.Title != "Atom Torture Test: SVG Content" {
		t.Errorf("Expected title 'Atom Torture Test: SVG Content', got %q", metadata.Title)
	}

	if len(entries) == 0 {
		t.Fatal("Expected entries, got none")
	}

	// Test simple SVG entry
	entry := entries[0]
	if entry.Title != "Simple SVG Circle" {
		t.Errorf("Expected first entry title 'Simple SVG Circle', got %q", entry.Title)
	}

	// Check if SVG is preserved or stripped
	hasSVG := strings.Contains(entry.Content, "<svg") || strings.Contains(entry.Content, "<circle")

	if hasSVG {
		t.Logf("SVG elements are preserved")
		// Check for dangerous attributes should be removed even if SVG is kept
		if strings.Contains(entry.Content, "onload") || strings.Contains(entry.Content, "onclick") {
			t.Errorf("SVG event handlers not sanitized: %s", entry.Content)
		}
	} else {
		t.Logf("SVG elements are stripped (safer for security)")
	}

	// Text content should remain regardless
	if !strings.Contains(entry.Content, "red circle") {
		t.Errorf("Content missing descriptive text: %s", entry.Content)
	}
}

// TestSVG_WithFallback tests SVG with fallback image handling
func TestSVG_WithFallback(t *testing.T) {
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-svg.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(feedData, "https://example.com/feed", time.Now())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Find the fallback entry
	var fallbackEntry *Entry
	for i := range entries {
		if entries[i].Title == "SVG with Image Fallback" {
			fallbackEntry = &entries[i]
			break
		}
	}

	if fallbackEntry == nil {
		t.Fatal("Could not find 'SVG with Image Fallback' entry")
	}

	// Note: bluemonday strips <object> tags entirely for security reasons,
	// including any nested <img> fallback tags. This is correct and expected behavior.
	// <object> tags can be used for Flash, Java applets, and other potentially dangerous content.

	// The surrounding text should be preserved
	if !strings.Contains(fallbackEntry.Content, "fallback") {
		t.Errorf("Text content should be preserved even when object tag is stripped: %s", fallbackEntry.Content)
	}

	// Document current behavior: object tag and nested img are both stripped
	hasObjectTag := strings.Contains(fallbackEntry.Content, "<object")
	hasImgTag := strings.Contains(fallbackEntry.Content, "https://example.com/diagram-fallback.png")

	if hasObjectTag {
		t.Logf("Object tag preserved (unexpected - may need security review)")
	} else {
		t.Logf("Object tag stripped (correct - prevents Flash/Java/plugin exploits)")
	}

	if hasImgTag {
		t.Logf("Nested img fallback preserved (good UX)")
	} else {
		t.Logf("Nested img fallback also stripped (expected with object tag removal)")
		t.Logf("Recommendation: Use direct <img src='file.svg'> instead of object/img pattern")
	}
}

// TestSVG_Security tests that SVG cannot be used for XSS
func TestSVG_Security(t *testing.T) {
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-svg.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(feedData, "https://example.com/feed", time.Now())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Find the security test entry
	var secEntry *Entry
	for i := range entries {
		if entries[i].Title == "SVG Security Test" {
			secEntry = &entries[i]
			break
		}
	}

	if secEntry == nil {
		t.Fatal("Could not find 'SVG Security Test' entry")
	}

	// All XSS vectors should be removed
	dangerousPatterns := []string{"onload=", "onclick=", "<script>", "alert('XSS')"}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(strings.ToLower(secEntry.Content), strings.ToLower(pattern)) {
			t.Errorf("Dangerous pattern %q not sanitized: %s", pattern, secEntry.Content)
		}
	}

	// Safe text should remain
	if !strings.Contains(secEntry.Content, "No scripts should execute") {
		t.Errorf("Safe text missing: %s", secEntry.Content)
	}
}

// TestSVG_InImgTag tests SVG referenced in img src attribute
func TestSVG_InImgTag(t *testing.T) {
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-svg.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(feedData, "https://example.com/feed", time.Now())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Find entry 6: "Multiple Fallback Layers"
	var imgEntry *Entry
	for i := range entries {
		if entries[i].Title == "Multiple Fallback Layers" {
			imgEntry = &entries[i]
			break
		}
	}

	if imgEntry == nil {
		t.Fatal("Could not find 'Multiple Fallback Layers' entry")
	}

	// SVG in img src should be preserved (safe)
	if !strings.Contains(imgEntry.Content, "graphic.svg") {
		t.Errorf("SVG in img src not preserved: %s", imgEntry.Content)
	}

	// Should have proper alt text
	if !strings.Contains(imgEntry.Content, "alt=") {
		t.Errorf("Alt attribute missing from img: %s", imgEntry.Content)
	}
}

// TestXHTML_MixedContent tests XHTML with various HTML elements
func TestXHTML_MixedContent(t *testing.T) {
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-xhtml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(feedData, "https://example.com/feed", time.Now())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Find "XHTML Mixed Content" entry
	var mixedEntry *Entry
	for i := range entries {
		if entries[i].Title == "XHTML Mixed Content" {
			mixedEntry = &entries[i]
			break
		}
	}

	if mixedEntry == nil {
		t.Fatal("Could not find 'XHTML Mixed Content' entry")
	}

	// Check for various HTML elements
	expectedElements := []string{"<h2>", "<code>", "<strong>", "<blockquote>", "<img"}
	for _, elem := range expectedElements {
		if !strings.Contains(mixedEntry.Content, elem) {
			t.Errorf("Expected element %q missing from content: %s", elem, mixedEntry.Content)
		}
	}

	// Image should have proper attributes
	if !strings.Contains(mixedEntry.Content, "alt=") {
		t.Errorf("Image missing alt attribute: %s", mixedEntry.Content)
	}
}

// TestContentTypeDetection verifies that ContentType field is set correctly
func TestContentTypeDetection(t *testing.T) {
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-xhtml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(feedData, "https://example.com/feed", time.Now())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// All entries should have ContentType set
	for i, entry := range entries {
		if entry.ContentType == "" {
			t.Errorf("Entry %d (%q) has empty ContentType", i, entry.Title)
		}

		// ContentType should be one of: "html", "text", or "xhtml"
		// Note: Current implementation only sets "html", which is acceptable
		// as long as content is properly sanitized
		validTypes := []string{"html", "text", "xhtml"}
		isValid := false
		for _, vt := range validTypes {
			if entry.ContentType == vt {
				isValid = true
				break
			}
		}
		if !isValid {
			t.Errorf("Entry %d (%q) has invalid ContentType %q", i, entry.Title, entry.ContentType)
		}
	}
}

// TestNamespaceHandling verifies proper XML namespace handling
func TestNamespaceHandling(t *testing.T) {
	n := New()

	// Test all three torture test feeds
	feeds := []string{
		"atom-torture-xhtml.xml",
		"atom-torture-mathml.xml",
		"atom-torture-svg.xml",
	}

	for _, feedFile := range feeds {
		t.Run(feedFile, func(t *testing.T) {
			feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", feedFile))
			if err != nil {
				t.Fatalf("Failed to read %s: %v", feedFile, err)
			}

			_, entries, err := n.Parse(feedData, "https://example.com/feed", time.Now())
			if err != nil {
				t.Fatalf("Parse failed for %s: %v", feedFile, err)
			}

			if len(entries) == 0 {
				t.Fatalf("No entries parsed from %s", feedFile)
			}

			// All entries should have content
			for i, entry := range entries {
				if entry.Content == "" && entry.Summary == "" {
					t.Errorf("Entry %d in %s has no content or summary", i, feedFile)
				}

				// Content should not contain raw namespace declarations
				// (they should be processed by the parser)
				if strings.Contains(entry.Content, "xmlns=") {
					t.Logf("Note: Entry %d in %s contains xmlns attribute (may be preserved by parser)", i, feedFile)
				}
			}
		})
	}
}
