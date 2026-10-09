// Package htmlsafety is test support: it inspects parsed HTML for markup that
// can run script or redirect the page, so tests can assert on the structure of
// sanitized output instead of searching it for substrings.
package htmlsafety

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// forbiddenElements can execute script, load active content, or change how
// the page resolves URLs or navigates (spec: CLAUDE.md security requirements).
var forbiddenElements = map[string]bool{
	"script": true, "iframe": true, "frame": true, "frameset": true,
	"object": true, "embed": true, "applet": true, "base": true,
	"meta": true, "link": true, "style": true, "form": true,
	// SVG and MathML carry their own script and animation vectors
	// (<set to="javascript:...">, xlink:href); sanitized output has neither.
	"svg": true, "math": true,
}

// urlAttributes hold URLs that a browser will load or navigate to.
var urlAttributes = map[string]bool{
	"href": true, "src": true, "action": true, "formaction": true,
	"background": true, "dynsrc": true, "lowsrc": true, "poster": true,
	"xlink:href": true,
}

// Violations returns one description per unsafe element or attribute found
// in n and its descendants. This is a scoped security oracle, not a browser
// exploit detector. It permits the sanitizer's existing http/https, relative
// and mailto URL policy.
func Violations(n *html.Node) []string {
	var out []string
	walk(n, func(n *html.Node) {
		if n.Type == html.ElementNode {
			out = append(out, elementViolations(n.Data, n.Attr)...)
		}
	})
	return out
}

// MarkupViolations scans every token, including head markup and deeply nested
// fragments which html.Parse rejects at 512 open elements. Unlike wrapping a
// fragment in a document and inspecting only its body, it cannot omit markup
// moved outside that body. It does not model every browser repair/XSS vector.
func MarkupViolations(s string) ([]string, error) {
	var out []string
	tokens := html.NewTokenizer(strings.NewReader(s))
	for {
		switch tokens.Next() {
		case html.ErrorToken:
			if err := tokens.Err(); err != io.EOF {
				return nil, err
			}
			return out, nil
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokens.Token()
			out = append(out, elementViolations(token.Data, token.Attr)...)
		}
	}
}

func elementViolations(tag string, attrs []html.Attribute) []string {
	var out []string
	if forbiddenElements[tag] {
		out = append(out, fmt.Sprintf("element <%s>", tag))
	}
	for _, a := range attrs {
		key := strings.ToLower(a.Key)
		if a.Namespace != "" {
			key = strings.ToLower(a.Namespace) + ":" + key
		}
		if strings.HasPrefix(key, "on") || key == "style" || key == "srcdoc" || (urlAttributes[key] && !safeURL(a.Val)) {
			out = append(out, fmt.Sprintf("<%s %s=%q>", tag, key, a.Val))
		}
	}
	return out
}

// Fragment parses s in a body context for text comparisons. Use
// MarkupViolations for security checks so parser depth limits are irrelevant.
func Fragment(s string) (*html.Node, error) {
	body := &html.Node{Type: html.ElementNode, Data: "body", DataAtom: atom.Body}
	nodes, err := html.ParseFragment(strings.NewReader(s), body)
	if err != nil {
		return nil, err
	}
	for _, n := range nodes {
		body.AppendChild(n)
	}
	return body, nil
}

// Find returns the first element named tag in n, or nil.
func Find(n *html.Node, tag string) *html.Node {
	var found *html.Node
	walk(n, func(n *html.Node) {
		if found == nil && n.Type == html.ElementNode && n.Data == tag {
			found = n
		}
	})
	return found
}

// Text returns the concatenated text content of n.
func Text(n *html.Node) string {
	var b strings.Builder
	walk(n, func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
	})
	return b.String()
}

func walk(n *html.Node, visit func(*html.Node)) {
	if n == nil {
		return
	}
	stack := []*html.Node{n}
	for len(stack) > 0 {
		n = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		visit(n)
		for c := n.LastChild; c != nil; c = c.PrevSibling {
			stack = append(stack, c)
		}
	}
}

func safeURL(raw string) bool {
	// Browsers strip leading/trailing whitespace and control characters, and
	// ignore tabs and newlines inside the scheme ("java\tscript:").
	cleaned := strings.Map(func(r rune) rune {
		if r == '\t' || r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, strings.TrimFunc(raw, func(r rune) bool { return r <= ' ' }))
	u, err := url.Parse(cleaned)
	if err != nil {
		return false
	}
	switch strings.ToLower(u.Scheme) {
	case "", "http", "https", "mailto":
		return true
	}
	return false
}
