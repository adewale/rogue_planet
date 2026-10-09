package htmlsafety

import (
	"strings"
	"testing"
)

// The checker is an oracle for other tests, so it needs inputs it must flag
// and near-misses it must not flag.
func TestViolations(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int // number of violations
	}{
		{"plain paragraph", `<p>hello</p>`, 0},
		{"http link", `<a href="http://example.com/">x</a>`, 0},
		{"https image", `<img src="https://example.com/a.png" alt="a">`, 0},
		{"existing mailto policy", `<a href="mailto:person@example.com">email</a>`, 0},
		{"relative link", `<a href="/about">x</a>`, 0},
		{"fragment link", `<a href="#top">x</a>`, 0},
		{"text that mentions script", `<p>&lt;script&gt;alert(1)&lt;/script&gt; onerror=x javascript:</p>`, 0},
		{"attribute named like a handler value", `<a title="onclick" href="https://e.com">x</a>`, 0},

		{"script element", `<script>alert(1)</script>`, 1},
		{"iframe element", `<iframe src="https://e.com"></iframe>`, 1},
		{"object element", `<object data="x.swf"></object>`, 1},
		{"embed element", `<embed src="x.swf">`, 1},
		{"base element", `<base href="https://evil.example.com/">`, 1},
		{"meta refresh", `<meta http-equiv="refresh" content="0;url=https://e.com">`, 1},
		{"style element", `<style>body{}</style>`, 1},
		{"form element", `<form action="https://e.com"></form>`, 1},
		{"event handler", `<img src="https://e.com/a.png" onerror="alert(1)">`, 1},
		{"uppercase event handler", `<div ONCLICK="alert(1)">x</div>`, 1},
		{"style attribute", `<div style="background:url(x)">x</div>`, 1},
		{"srcdoc attribute", `<div srcdoc="<script>alert(1)</script>">x</div>`, 1},
		{"javascript href", `<a href="javascript:alert(1)">x</a>`, 1},
		{"mixed case javascript href", `<a href="JaVaScRiPt:alert(1)">x</a>`, 1},
		{"tab-split javascript href", "<a href=\"java\tscript:alert(1)\">x</a>", 1},
		{"leading space javascript href", `<a href=" javascript:alert(1)">x</a>`, 1},
		{"leading control javascript href", `<a href="&#x01;javascript:alert(1)">x</a>`, 1},
		{"data src", `<img src="data:image/png;base64,AAAA">`, 1},
		{"vbscript href", `<a href="vbscript:msgbox(1)">x</a>`, 1},
		{"svg onload", `<svg onload="alert(1)"></svg>`, 2},
		{"svg animation to javascript", `<svg><a><set attributeName="href" to="javascript:alert(1)"/></a></svg>`, 1},
		{"svg xlink href", `<svg><a xlink:href="javascript:alert(1)">x</a></svg>`, 2},
		{"mathml element", `<math><mi>x</mi></math>`, 1},
		{"deep harmless nesting", strings.Repeat("<div>", 600) + "safe" + strings.Repeat("</div>", 600), 0},
		{"markup after body boundary", `</body><head><meta http-equiv="refresh" content="0;url=https://e.com"></head>`, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarkupViolations(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != tt.want {
				t.Errorf("Violations(%s) = %q, want %d violation(s)", tt.input, got, tt.want)
			}
		})
	}
}

func TestText(t *testing.T) {
	body, err := Fragment(`<p>one <b>two</b></p><p>three</p>`)
	if err != nil {
		t.Fatal(err)
	}
	if got := Text(body); got != "one twothree" {
		t.Errorf("Text = %q, want %q", got, "one twothree")
	}
}
