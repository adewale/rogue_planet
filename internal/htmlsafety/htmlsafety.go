// Package htmlsafety is test support: it inspects parsed HTML for markup that
// can run script or redirect the page, so tests can assert on the structure of
// sanitized output instead of searching it for substrings.
package htmlsafety

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/html"
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
// in n and its descendants. Relative URLs and http/https URLs are allowed.
func Violations(n *html.Node) []string {
	var out []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if forbiddenElements[n.Data] {
				out = append(out, fmt.Sprintf("element <%s>", n.Data))
			}
			for _, a := range n.Attr {
				key := strings.ToLower(a.Key)
				if a.Namespace != "" {
					key = strings.ToLower(a.Namespace) + ":" + key
				}
				if strings.HasPrefix(key, "on") {
					out = append(out, fmt.Sprintf("<%s %s=%q>", n.Data, key, a.Val))
				}
				if key == "style" {
					out = append(out, fmt.Sprintf("<%s style=%q>", n.Data, a.Val))
				}
				if urlAttributes[key] && !safeURL(a.Val) {
					out = append(out, fmt.Sprintf("<%s %s=%q>", n.Data, key, a.Val))
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return out
}

// Fragment parses s as the body of an HTML document and returns the <body>
// node, the way a browser would interpret it.
func Fragment(s string) (*html.Node, error) {
	doc, err := html.Parse(strings.NewReader("<!DOCTYPE html><html><head></head><body>" + s + "</body></html>"))
	if err != nil {
		return nil, err
	}
	return Find(doc, "body"), nil
}

// Find returns the first element named tag in n, or nil.
func Find(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if f := Find(c, tag); f != nil {
			return f
		}
	}
	return nil
}

// Text returns the concatenated text content of n.
func Text(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return b.String()
}

func safeURL(raw string) bool {
	// Browsers strip leading/trailing whitespace and control characters, and
	// ignore tabs and newlines inside the scheme ("java\tscript:").
	cleaned := strings.Map(func(r rune) rune {
		if r == '\t' || r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, strings.TrimSpace(raw))
	u, err := url.Parse(cleaned)
	if err != nil {
		return false
	}
	switch strings.ToLower(u.Scheme) {
	case "", "http", "https":
		return true
	}
	return false
}
