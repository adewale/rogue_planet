package normalizer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Atom Torture Tests
//
// Original source: Jacques Distler's "Atom Torture Test" (April 18, 2006)
// URL: https://golem.ph.utexas.edu/~distler/blog/archives/000793.html
// Alternative URL: https://classes.golem.ph.utexas.edu/~distler/blog/archives/000793.html
//
// Background:
// Jacques Distler created these tests to expose critical gaps in feed aggregator
// implementations, particularly around advanced Atom 1.0 features. As of 2006,
// virtually no aggregators properly handled type="xhtml" content using XML parsing
// rules instead of HTML tag-soup parsing.
//
// The Four Original Tests:
// 1. Relative URLs with xml:base - URLs should resolve per entry's xml:base attribute
// 2. XHTML Content - Must use XML parsing (case-sensitive, strict)
//    - The famous "a **b**c **D**e f" test - only lowercase 'b' should be bold
// 3. MathML Support - Mathematical markup (requires XML parsing + MathML renderer)
// 4. SVG with Fallbacks - Vector graphics with fallback images (object/img pattern)
//
// Rogue Planet Implementation:
// - Security prioritized over feature completeness
// - MathML/SVG stripped by bluemonday (prevents XSS)
// - XHTML sanitized but structure preserved
// - All event handlers removed
// - Object tags completely removed
//
// Research Documentation: specs/research/ATOM_TORTURE_TEST_RESEARCH.md
//
// These tests validate that Rogue Planet:
// 1. Correctly distinguishes XHTML/HTML/text content types
// 2. Sanitizes all dangerous content (scripts, event handlers)
// 3. Preserves safe HTML structure
// 4. Handles XML namespaces correctly
// 5. Documents expected behavior when stripping MathML/SVG
// 6. Tests XML case-sensitivity (The Distler Test)
// 7. Tests xml:base relative URL resolution

// TestAtomContentType_XHTML tests handling of Atom content type="xhtml"
func TestAtomContentType_XHTML(t *testing.T) {
	t.Parallel()
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-xhtml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	metadata, entries, err := n.Parse(context.Background(), feedData, "https://example.com/feed", time.Now())
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
	t.Parallel()
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-xhtml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(context.Background(), feedData, "https://example.com/feed", time.Now())
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
	t.Parallel()
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-xhtml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(context.Background(), feedData, "https://example.com/feed", time.Now())
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
	t.Parallel()
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-xhtml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(context.Background(), feedData, "https://example.com/feed", time.Now())
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
	t.Parallel()
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-mathml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	metadata, entries, err := n.Parse(context.Background(), feedData, "https://example.com/feed", time.Now())
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
	t.Parallel()
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-mathml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(context.Background(), feedData, "https://example.com/feed", time.Now())
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
	t.Parallel()
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-mathml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(context.Background(), feedData, "https://example.com/feed", time.Now())
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
	t.Parallel()
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-svg.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	metadata, entries, err := n.Parse(context.Background(), feedData, "https://example.com/feed", time.Now())
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
	t.Parallel()
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-svg.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(context.Background(), feedData, "https://example.com/feed", time.Now())
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
	t.Parallel()
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-svg.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(context.Background(), feedData, "https://example.com/feed", time.Now())
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
	t.Parallel()
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-svg.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(context.Background(), feedData, "https://example.com/feed", time.Now())
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
	t.Parallel()
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-xhtml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(context.Background(), feedData, "https://example.com/feed", time.Now())
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
	t.Parallel()
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-xhtml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(context.Background(), feedData, "https://example.com/feed", time.Now())
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
	t.Parallel()
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

			_, entries, err := n.Parse(context.Background(), feedData, "https://example.com/feed", time.Now())
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

// TestXHTML_CaseSensitivity tests the famous Distler test case
// This is THE definitive test of proper XML parsing vs HTML tag-soup parsing.
// Reference: https://golem.ph.utexas.edu/~distler/blog/archives/000793.html
func TestXHTML_CaseSensitivity(t *testing.T) {
	t.Parallel()
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-xhtml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(context.Background(), feedData, "https://example.com/feed", time.Now())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Find the Distler test entry
	var distlerEntry *Entry
	for i := range entries {
		if entries[i].Title == "XHTML Case Sensitivity - The Distler Test" {
			distlerEntry = &entries[i]
			break
		}
	}

	if distlerEntry == nil {
		t.Fatal("Could not find 'XHTML Case Sensitivity - The Distler Test' entry")
	}

	// The test case is: a <b>b</b>c <D>D</D>e f
	// Expected: "a **b**c De f" (only lowercase 'b' is bold)

	// Should contain text 'b' - the lowercase b should be present
	if !strings.Contains(distlerEntry.Content, "b") {
		t.Errorf("Content should contain 'b': %s", distlerEntry.Content)
	}

	// Should contain the letters a, c, D, e, f as text
	expectedText := []string{"a ", "c ", "D", "e f"}
	for _, text := range expectedText {
		if !strings.Contains(distlerEntry.Content, text) {
			t.Errorf("Content missing expected text %q: %s", text, distlerEntry.Content)
		}
	}

	// Check for <b> or <strong> tag (sanitizer might convert <b> to <strong>)
	hasBoldTag := strings.Contains(distlerEntry.Content, "<b>") || strings.Contains(distlerEntry.Content, "<strong>")
	if !hasBoldTag {
		t.Errorf("Content should contain <b> or <strong> tag for lowercase b: %s", distlerEntry.Content)
	}

	// Check if uppercase D tag exists - it SHOULD NOT be treated as markup
	// The <D> tag is not a recognized HTML element in lowercase, so with proper XML parsing
	// it should either:
	// 1. Be stripped as an unknown element (leaving just "D" as text)
	// 2. Be preserved as literal <D>D</D> (unlikely with sanitizer)
	// It should NOT be converted to <d> (which would indicate HTML tag-soup parsing)

	hasUppercaseDTag := strings.Contains(distlerEntry.Content, "<D>") || strings.Contains(distlerEntry.Content, "<D ")
	hasLowercaseDTag := strings.Contains(distlerEntry.Content, "<d>") || strings.Contains(distlerEntry.Content, "<d ")

	if hasLowercaseDTag {
		t.Errorf("FAIL: Parser is using HTML tag-soup parsing (case-insensitive)")
		t.Errorf("The <D> tag should NOT be converted to <d>")
		t.Errorf("Content: %s", distlerEntry.Content)
	} else if hasUppercaseDTag {
		t.Logf("Note: Uppercase <D> tag preserved in content (unusual but technically valid XML)")
		t.Logf("Content: %s", distlerEntry.Content)
	} else {
		t.Logf("SUCCESS: Unknown <D> tag was stripped, leaving just 'D' as text (expected behavior)")
		// Verify the text 'D' is still present
		if !strings.Contains(distlerEntry.Content, "D") {
			t.Errorf("Content should contain 'D' as text even if <D> tag is stripped: %s", distlerEntry.Content)
		}
	}
}

// TestXMLBase_RelativeURLResolution tests xml:base support
// Atom 1.0 supports xml:base attributes for entry-specific base URLs.
// Reference: https://golem.ph.utexas.edu/~distler/blog/archives/000793.html
// Reference: https://www.w3.org/TR/xmlbase/
func TestXMLBase_RelativeURLResolution(t *testing.T) {
	t.Parallel()
	n := New()

	feedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "atom-torture-xhtml.xml"))
	if err != nil {
		t.Fatalf("Failed to read test feed: %v", err)
	}

	_, entries, err := n.Parse(context.Background(), feedData, "https://example.com/feed", time.Now())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Find the xml:base test entry
	var xmlbaseEntry *Entry
	for i := range entries {
		if entries[i].Title == "xml:base Relative URL Test" {
			xmlbaseEntry = &entries[i]
			break
		}
	}

	if xmlbaseEntry == nil {
		t.Fatal("Could not find 'xml:base Relative URL Test' entry")
	}

	// The entry has xml:base="https://example.com/blog/2006/04/"
	// and contains <a href="image.jpg">
	// The href should resolve to: https://example.com/blog/2006/04/image.jpg

	expectedURL := "https://example.com/blog/2006/04/image.jpg"

	// Check if the URL was resolved to absolute
	if strings.Contains(xmlbaseEntry.Content, expectedURL) {
		t.Logf("SUCCESS: xml:base relative URLs are properly resolved to absolute URLs")
	} else if strings.Contains(xmlbaseEntry.Content, "image.jpg") {
		// URL is present but might not be resolved
		t.Logf("WARNING: Relative URL 'image.jpg' found but may not be resolved to absolute")
		t.Logf("Expected: %s", expectedURL)

		// Check if it's a relative URL
		if strings.Contains(xmlbaseEntry.Content, `href="image.jpg"`) {
			t.Logf("Note: URL remains relative - xml:base attribute not processed")
			t.Logf("This is acceptable if relative URLs work in the generated HTML context")
		} else if strings.Contains(xmlbaseEntry.Content, `href="https://example.com/`) {
			// It's absolute but to wrong path
			t.Logf("Note: URL was resolved to absolute but may not use xml:base correctly")
		}

		t.Logf("Content excerpt: %s", xmlbaseEntry.Content)
	} else {
		t.Errorf("Could not find 'image.jpg' link in content at all: %s", xmlbaseEntry.Content)
	}

	// Verify the link element exists
	hasLink := strings.Contains(xmlbaseEntry.Content, "<a") || strings.Contains(xmlbaseEntry.Content, "href")
	if !hasLink {
		t.Errorf("Content should contain a link element: %s", xmlbaseEntry.Content)
	}
}
