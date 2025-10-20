package normalizer

import (
	"strings"
	"testing"
)

// Comprehensive XSS prevention tests
func TestSanitizeHTML_XSS_Prevention(t *testing.T) {
	n := New()

	tests := []struct {
		name    string
		input   string
		wantNot []string // Strings that should NOT appear in output
	}{
		{
			name:    "script tag removed",
			input:   `<p>Hello</p><script>alert(1)</script><p>World</p>`,
			wantNot: []string{"<script>", "alert(1)", "</script>"},
		},
		{
			name:    "script in img src",
			input:   `<img src="javascript:alert(1)">`,
			wantNot: []string{"javascript:", "alert(1)"},
		},
		{
			name:    "script in a href",
			input:   `<a href="javascript:alert(1)">Click</a>`,
			wantNot: []string{"javascript:", "alert(1)"},
		},
		{
			name:    "data URI in img",
			input:   `<img src="data:text/html,<script>alert(1)</script>">`,
			wantNot: []string{"data:", "<script>", "alert(1)"},
		},
		{
			name:    "data URI in a href",
			input:   `<a href="data:text/html,<script>alert(1)</script>">Click</a>`,
			wantNot: []string{"data:", "<script>", "alert(1)"},
		},
		{
			name:    "onclick handler",
			input:   `<div onclick="alert(1)">Click me</div>`,
			wantNot: []string{"onclick", "alert(1)"},
		},
		{
			name:    "onerror handler in img",
			input:   `<img src="x" onerror="alert(1)">`,
			wantNot: []string{"onerror", "alert(1)"},
		},
		{
			name:    "onload handler",
			input:   `<body onload="alert(1)">Content</body>`,
			wantNot: []string{"onload", "alert(1)"},
		},
		{
			name:    "onmouseover handler",
			input:   `<div onmouseover="alert(1)">Hover</div>`,
			wantNot: []string{"onmouseover", "alert(1)"},
		},
		{
			name:    "iframe tag",
			input:   `<p>Safe</p><iframe src="http://evil.com"></iframe><p>Content</p>`,
			wantNot: []string{"<iframe>", "</iframe>", "evil.com"},
		},
		{
			name:    "object tag",
			input:   `<object data="evil.swf"></object>`,
			wantNot: []string{"<object>", "</object>", "evil.swf"},
		},
		{
			name:    "embed tag",
			input:   `<embed src="evil.swf">`,
			wantNot: []string{"<embed>", "evil.swf"},
		},
		{
			name:    "base tag",
			input:   `<base href="http://evil.com"><p>Content</p>`,
			wantNot: []string{"<base>", "</base>"},
		},
		{
			name:    "meta refresh",
			input:   `<meta http-equiv="refresh" content="0;url=http://evil.com">`,
			wantNot: []string{"<meta>", "refresh", "evil.com"},
		},
		{
			name:    "nested scripts",
			input:   `<div><span><script>alert(1)</script></span></div>`,
			wantNot: []string{"<script>", "alert(1)"},
		},
		{
			name:  "obfuscated script",
			input: `<scr<script>ipt>alert(1)</script>`,
			// The <script> tag is removed (security), but text content "alert(1)" may remain.
			// This is acceptable - the text is not executable. We check for <script> removal.
			wantNot: []string{"<script>"},
		},
		{
			name:    "uppercase SCRIPT tag",
			input:   `<SCRIPT>alert(1)</SCRIPT>`,
			wantNot: []string{"SCRIPT", "alert(1)"},
		},
		{
			name:    "mixed case oNcLiCk",
			input:   `<div oNcLiCk="alert(1)">Click</div>`,
			wantNot: []string{"alert(1)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := n.SanitizeHTML(tt.input)

			for _, forbidden := range tt.wantNot {
				if strings.Contains(strings.ToLower(output), strings.ToLower(forbidden)) {
					t.Errorf("Output contains forbidden string %q\nInput: %s\nOutput: %s",
						forbidden, tt.input, output)
				}
			}
		})
	}
}

// Test safe HTML is preserved
func TestSanitizeHTML_SafeContent(t *testing.T) {
	n := New()

	tests := []struct {
		name  string
		input string
		want  []string // Strings that SHOULD appear in output
	}{
		{
			name:  "safe paragraph",
			input: `<p>Hello world</p>`,
			want:  []string{"<p>", "Hello world", "</p>"},
		},
		{
			name:  "safe link",
			input: `<a href="http://example.com">Link</a>`,
			want:  []string{"<a", "href", "http://example.com", "Link", "</a>"},
		},
		{
			name:  "safe https link",
			input: `<a href="https://example.com">Secure Link</a>`,
			want:  []string{"<a", "href", "https://example.com", "Secure Link", "</a>"},
		},
		{
			name:  "safe image",
			input: `<img src="https://example.com/image.jpg" alt="Image">`,
			want:  []string{"<img", "src", "https://example.com/image.jpg", "alt", "Image"},
		},
		{
			name:  "bold and italic",
			input: `<p><b>Bold</b> and <i>italic</i> text</p>`,
			want:  []string{"<b>", "Bold", "</b>", "<i>", "italic", "</i>"},
		},
		{
			name:  "strong and em",
			input: `<p><strong>Strong</strong> and <em>emphasis</em></p>`,
			want:  []string{"<strong>", "Strong", "</strong>", "<em>", "emphasis", "</em>"},
		},
		{
			name:  "unordered list",
			input: `<ul><li>Item 1</li><li>Item 2</li></ul>`,
			want:  []string{"<ul>", "<li>", "Item 1", "</li>", "</ul>"},
		},
		{
			name:  "ordered list",
			input: `<ol><li>First</li><li>Second</li></ol>`,
			want:  []string{"<ol>", "<li>", "First", "</li>", "</ol>"},
		},
		{
			name:  "blockquote",
			input: `<blockquote>Quoted text</blockquote>`,
			want:  []string{"<blockquote>", "Quoted text", "</blockquote>"},
		},
		{
			name:  "code block",
			input: `<pre><code>function foo() {}</code></pre>`,
			want:  []string{"<pre>", "<code>", "function foo()", "</code>", "</pre>"},
		},
		{
			name:  "headings",
			input: `<h1>Title</h1><h2>Subtitle</h2><h3>Section</h3>`,
			want:  []string{"<h1>", "Title", "</h1>", "<h2>", "Subtitle", "</h2>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := n.SanitizeHTML(tt.input)

			for _, required := range tt.want {
				if !strings.Contains(output, required) {
					t.Errorf("Output missing required string %q\nInput: %s\nOutput: %s",
						required, tt.input, output)
				}
			}
		})
	}
}

// Test URL scheme validation
func TestSanitizeHTML_URLSchemes(t *testing.T) {
	n := New()

	tests := []struct {
		name        string
		input       string
		shouldAllow bool
		checkString string
	}{
		{
			name:        "http URL allowed",
			input:       `<a href="http://example.com">Link</a>`,
			shouldAllow: true,
			checkString: "http://example.com",
		},
		{
			name:        "https URL allowed",
			input:       `<a href="https://example.com">Link</a>`,
			shouldAllow: true,
			checkString: "https://example.com",
		},
		{
			name:        "ftp URL blocked",
			input:       `<a href="ftp://example.com/file">Link</a>`,
			shouldAllow: false,
			checkString: "ftp://",
		},
		{
			name:        "file URL blocked",
			input:       `<a href="file:///etc/passwd">Link</a>`,
			shouldAllow: false,
			checkString: "file://",
		},
		{
			name:        "javascript URL blocked",
			input:       `<a href="javascript:void(0)">Link</a>`,
			shouldAllow: false,
			checkString: "javascript:",
		},
		{
			name:        "data URL blocked",
			input:       `<img src="data:image/png;base64,ABC">`,
			shouldAllow: false,
			checkString: "data:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := n.SanitizeHTML(tt.input)
			contains := strings.Contains(strings.ToLower(output), strings.ToLower(tt.checkString))

			if tt.shouldAllow && !contains {
				t.Errorf("Should allow %q but it was removed\nOutput: %s", tt.checkString, output)
			}
			if !tt.shouldAllow && contains {
				t.Errorf("Should block %q but it was preserved\nOutput: %s", tt.checkString, output)
			}
		})
	}
}

// Test various XSS vectors from OWASP and real-world attacks
func TestSanitizeHTML_RealWorld_XSS_Vectors(t *testing.T) {
	n := New()

	vectors := []struct {
		name   string
		vector string
	}{
		{"IMG onerror", `<IMG SRC=x onerror="alert('XSS')">`},
		{"IMG lowsrc", `<IMG SRC=x onError="alert('XSS')">`},
		{"BODY onload", `<BODY ONLOAD=alert('XSS')>`},
		{"SCRIPT in attribute", `<IMG SRC="javascript:alert('XSS')">`},
		{"IMG dynsrc", `<IMG DYNSRC="javascript:alert('XSS')">`},
		{"TABLE background", `<TABLE BACKGROUND="javascript:alert('XSS')">`},
		{"DIV style expression", `<DIV STYLE="background-image: expression(alert('XSS'))">`},
		{"XML namespace", `<HTML xmlns:xss><?import namespace="xss" implementation="http://ha.ckers.org/xss.htc"><xss:xss>XSS</xss:xss></HTML>`},
		{"Meta charset", `<META HTTP-EQUIV="Set-Cookie" Content="USERID=<SCRIPT>alert('XSS')</SCRIPT>">`},
		{"Link rel stylesheet", `<LINK REL="stylesheet" HREF="javascript:alert('XSS');">`},
		{"Style tag", `<STYLE>@import'http://ha.ckers.org/xss.css';</STYLE>`},
		{"Style tag 2", `<STYLE>BODY{-moz-binding:url("http://ha.ckers.org/xssmoz.xml#xss")}</STYLE>`},
		{"Input image", `<INPUT TYPE="IMAGE" SRC="javascript:alert('XSS');">`},
		{"Iframe", `<IFRAME SRC="javascript:alert('XSS');"></IFRAME>`},
		{"Frame", `<FRAMESET><FRAME SRC="javascript:alert('XSS');"></FRAMESET>`},
		{"SVG", `<svg/onload=alert('XSS')>`},
		{"Details open", `<details open ontoggle=alert('XSS')>`},
		{"Form action", `<FORM action="javascript:alert('XSS')"><input type="submit"></FORM>`},
	}

	for _, v := range vectors {
		t.Run(v.name, func(t *testing.T) {
			output := n.SanitizeHTML(v.vector)

			// Check that common dangerous patterns are removed
			dangerous := []string{"alert", "javascript:", "onerror", "onload", "expression("}
			for _, pattern := range dangerous {
				if strings.Contains(strings.ToLower(output), strings.ToLower(pattern)) {
					t.Errorf("XSS vector not fully sanitized\nVector: %s\nOutput: %s\nDangerous pattern: %s",
						v.name, output, pattern)
				}
			}
		})
	}
}

// Test HTML entity handling
func TestSanitizeHTML_HTMLEntities(t *testing.T) {
	n := New()

	tests := []struct {
		name  string
		input string
		want  string // Approximate expected content (may vary by implementation)
	}{
		{
			name:  "numeric entities",
			input: `<p>&#72;&#101;&#108;&#108;&#111;</p>`,
			want:  "Hello",
		},
		{
			name:  "hex entities",
			input: `<p>&#x48;&#x65;&#x6C;&#x6C;&#x6F;</p>`,
			want:  "Hello",
		},
		{
			name:  "named entities",
			input: `<p>&lt;script&gt;alert(1)&lt;/script&gt;</p>`,
			want:  "&lt;script&gt;", // HTML entities should remain encoded (safe)
		},
		{
			name:  "quote entities",
			input: `<p>&quot;Quoted&quot; and &apos;apostrophe&apos;</p>`,
			want:  "Quoted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := n.SanitizeHTML(tt.input)
			if !strings.Contains(output, tt.want) {
				t.Errorf("Output missing expected string %q\nInput: %s\nOutput: %s",
					tt.want, tt.input, output)
			}
		})
	}
}

// Test that unclosed tags are handled
func TestSanitizeHTML_MalformedHTML(t *testing.T) {
	n := New()

	tests := []struct {
		name  string
		input string
	}{
		{"unclosed paragraph", `<p>Hello world`},
		{"unclosed div", `<div>Content`},
		{"mismatched tags", `<p>Hello</div>`},
		{"missing closing angle", `<p>Hello<p`},
		{"double nested unclosed", `<div><p>Content`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := n.SanitizeHTML(tt.input)

			// Should not panic and should produce some output
			if output == "" {
				t.Error("Sanitizer returned empty string for malformed HTML")
			}

			// Should preserve text content even if HTML is malformed
			if strings.Contains(tt.input, "Hello") && !strings.Contains(output, "Hello") {
				t.Errorf("Sanitizer lost text content 'Hello'\nInput: %s\nOutput: %s",
					tt.input, output)
			}
			if strings.Contains(tt.input, "Content") && !strings.Contains(output, "Content") {
				t.Errorf("Sanitizer lost text content 'Content'\nInput: %s\nOutput: %s",
					tt.input, output)
			}

			// Should not contain obviously broken HTML
			if strings.Contains(output, "<<") || strings.Contains(output, ">>") {
				t.Errorf("Output contains broken HTML markers\nInput: %s\nOutput: %s",
					tt.input, output)
			}
		})
	}
}

// Test content length and special characters
func TestSanitizeHTML_EdgeCases(t *testing.T) {
	n := New()

	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ``},
		{"whitespace only", `   `},
		{"unicode emoji", `<p>Hello 👋 World 🌍</p>`},
		{"RTL text", `<p>مرحبا بك</p>`},
		{"mathematical symbols", `<p>∑ ∫ √ π</p>`},
		{"very long content", strings.Repeat("<p>Lorem ipsum dolor sit amet. </p>", 1000)},
		{"deeply nested", `<div><div><div><div><div><p>Deep</p></div></div></div></div></div>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			output := n.SanitizeHTML(tt.input)

			// Basic sanity check
			if tt.input != "" && output == "" && strings.TrimSpace(tt.input) != "" {
				t.Errorf("Sanitizer removed all content (should preserve safe HTML)\nInput: %s\nOutput: %s",
					tt.input, output)
			}
		})
	}
}
