# MathJax Support Research - NOT RECOMMENDED FOR IMPLEMENTATION

**Date**: 2025-10-21
**Status**: ❌ **NOT RECOMMENDED FOR ROGUE PLANET**
**Reason**: Conflicts with core architectural principles

---

## Executive Summary

After deep research into mathematical notation rendering in feed aggregators, **MathJax support should NOT be added to Rogue Planet** because:

1. **Violates security-first principle** - Requires relaxing Content Security Policy
2. **Adds external dependency** - Requires CDN or bundled 100KB+ JavaScript library
3. **Conflicts with static simplicity** - Adds client-side rendering complexity
4. **Breaks "good netizen" behavior** - Requires external CDN requests on every page load
5. **Minimal actual benefit** - Very few feeds use mathematical notation; those that do already handle it themselves

**Recommendation**: Let feed publishers handle math rendering in their own content. If aggregating math-heavy feeds becomes critical, use native MathML (which requires zero JavaScript) instead of MathJax.

---

## Why This Conflicts with Rogue Planet's Core Principles

### 1. Security-First Architecture

**From LESSONS_LEARNED.md, Lesson #12**:
> CVE-2009-2937: Never Trust Feed Content - Defense Strategy includes CSP headers in generated HTML

**Current CSP** (`pkg/generator/generator.go:378`):
```html
script-src 'self';
```

**Required for MathJax**:
```html
script-src 'self' https://cdn.jsdelivr.net;
```

This violates the principle of **defense in depth**. We specifically lock down script sources to prevent XSS, and MathJax would require punching a hole in that policy.

### 2. Simple is Sustainable

**From LESSONS_LEARNED.md, Lesson #3**:
> Fewer features = fewer bugs = longer project life
> Moonmoon's "stupidly simple" approach outlasted complex alternatives

Adding MathJax would:
- Add ~100KB JavaScript dependency
- Require CDN management or local bundling
- Add configuration complexity
- Create maintenance burden (security updates)
- Increase page load time

**Current**: 6 packages, ~6,000 lines of code, 88/100 Go standards compliance
**With MathJax**: Additional complexity for edge case feature

### 3. Static Output is King

**From LESSONS_LEARNED.md, Lesson #1**:
> Generate static HTML files, not dynamic pages
> Why: Survives traffic spikes, no database queries on reads

MathJax adds **client-side rendering**:
- JavaScript must execute to render math
- Doesn't work without JavaScript enabled
- Requires external network request (CDN)
- Page incomplete until JavaScript loads

This degrades the "works everywhere" property of static HTML.

### 4. Good Netizen Behavior

**From LESSONS_LEARNED.md, Lesson #27**:
> Be a Good Netizen - Clear User-Agent, Respect Cache-Control

Adding MathJax CDN means:
- Every visitor's browser hits cdn.jsdelivr.net
- Dependency on third-party service uptime
- Privacy implications (external requests)
- Would need to document CDN dependency

### 5. Minimal Actual Use Case

**Reality**: Very few feeds contain mathematical notation:
- Most blogs use plain text or images for math
- Academic blogs that need math already handle it (MathML or pre-rendered)
- RSS readers don't support math rendering anyway (it's a fundamental limitation)

**Rogue Planet's mission**: Aggregate feeds and display them chronologically
**Not our mission**: Solve the RSS math rendering problem that's existed for 12+ years

---

## Research Findings (For Reference)

### The Fundamental RSS/Feed Problem

**Core Issue**: RSS doesn't allow JavaScript, which creates fundamental incompatibility with MathJax.

- Feed readers don't pass through JavaScript
- MathJax won't render in most RSS readers
- Raw LaTeX appears as text: `$\alpha$` instead of α
- This is **unsolvable at the feed level**

### Historical Context

This problem has existed since at least 2012. The community consensus:
- Feed aggregators leave math as-is (LaTeX delimiters or MathML)
- Users install browser extensions (MathJax bookmarklet)
- Feed publishers pre-render to images if universal support needed

### MathJax Version Evolution

#### MathJax v2.x (Legacy)
- Required `script-src 'unsafe-eval'` in CSP ❌
- Required `style-src 'unsafe-inline'` in CSP ❌
- Used `<script type="math/tex">` tags
- **XSS vulnerabilities**: Nested script tags could execute (kramdown CVE-2020-14001)

#### MathJax v3+ (Modern)
- No longer requires `unsafe-eval` ✅
- Uses delimiter parsing: `\(...\)` for inline, `\[...\]` for display
- Still requires CDN in script-src ❌
- Still requires `style-src 'unsafe-inline'` ❌

### Bluemonday Sanitizer Conflicts

**Current implementation** (`pkg/normalizer/normalizer.go:57`):
```go
policy := bluemonday.UGCPolicy()
policy.AllowURLSchemes("http", "https")
```

**Problems**:
1. bluemonday strips `<script>` tags by default (security feature)
2. To allow `<script type="math/tex">`, would need `AllowUnsafe(true)` ❌
3. But allowing unsafe defeats entire purpose of sanitization
4. Can't distinguish between dangerous and "safe" script tags

### Math Markup Formats in Feeds

#### Format 1: LaTeX with Delimiters
```
Inline: $x^2$ or \(x^2\)
Display: $$x^2$$ or \[x^2\]
```

**Pros**: Human-readable, widely used in academic blogs
**Cons**: Requires MathJax on aggregated page, shows as raw code in RSS readers

#### Format 2: MathML (W3C Standard)
```xml
<math xmlns="http://www.w3.org/1998/Math/MathML">
  <msup><mi>x</mi><mn>2</mn></msup>
</math>
```

**Pros**:
- Native browser support (Chrome 109+, Firefox always, Safari now)
- W3C standard
- **No JavaScript required** ✅
- Survives sanitization

**Cons**:
- Verbose XML
- Rare in actual feeds (almost nobody publishes MathML)
- W3C Feed Validator flags as "non-HTML"

#### Format 3: Pre-rendered Images
```html
<img src="https://latex.codecogs.com/png.latex?x%5E2" alt="x^2">
```

**Pros**: Works everywhere, no JavaScript, survives sanitization
**Cons**: Not searchable, accessibility issues, external dependency

### Browser Support (2024-2025)

**Native MathML Support**:
- Firefox: Full native support (always had it)
- Chrome/Edge: Native support since Chrome 109 (January 2023)
- Safari: Native support in WebKit
- **All major browsers now support MathML Core natively**

This is significant: If we were to support math at all, **native MathML is the answer**, not JavaScript.

---

## If We Were to Implement (We Shouldn't)

### Least-Bad Approach: Native MathML Only

If absolutely required, the **only acceptable approach** would be:

**Configure bluemonday to allow MathML elements** (`pkg/normalizer/normalizer.go`):
```go
// Allow MathML elements (native browser support, no JavaScript)
policy.AllowElements("math", "mrow", "msup", "msub", "mi", "mn", "mo",
                     "mfrac", "msqrt", "mroot", "mtext", "mspace")
policy.AllowAttrs("xmlns").OnElements("math")
```

**Why this is acceptable**:
- ✅ Zero JavaScript
- ✅ No CSP changes required
- ✅ Native browser rendering (fast, accessible)
- ✅ W3C standard
- ✅ No external dependencies
- ✅ Maintains security posture

**Why this still isn't great**:
- ❌ Almost no feeds actually use MathML
- ❌ Doesn't help with LaTeX-formatted feeds (the common case)
- ❌ W3C Feed Validator complains
- ❌ Adds complexity for minimal benefit

**Effort**: ~30 minutes to implement and test

---

## Alternative Solutions

### 1. Document the Limitation
Add to README.md:
```markdown
## Mathematical Notation

Rogue Planet does not process mathematical notation. If you aggregate feeds
containing math equations:

- LaTeX delimiters (\(...\), \[...\]) will appear as plain text
- MathML may render in modern browsers (Chrome 109+, Firefox, Safari)
- For better math support, use a MathJax bookmarklet or browser extension
```

### 2. Let Feed Publishers Handle It
Feeds that need math can:
- Pre-render equations to images
- Use MathML (works in Rogue Planet with zero changes)
- Accept that aggregated view shows raw LaTeX

### 3. User Browser Extensions
Users who read math-heavy feeds can install:
- MathJax bookmarklet
- Feedbro Feed Reader (has MathJax option)
- Browser extensions for math rendering

---

## Security Considerations Discovered

### CVE-2018-1999024: MathJax XSS Vulnerability
MathJax itself has had XSS vulnerabilities, including in the `\unicode{}` macro.

### kramdown CVE-2020-14001: Script Tag Injection
Nested script tags inside `<script type="math/tex">` caused code execution:
```html
<script type="math/tex">
  % <![CDATA[
  <script>alert('XSS')</script>
  %]]>
</script>
```

**Fix**: kramdown now strips ALL `<script>` tags from math statements.

### General LaTeX Injection Risks
LaTeX itself can be a security risk. While MathJax sandboxes most operations, keeping math processing simple (or absent) is safer.

---

## Performance Considerations

### KaTeX vs MathJax
If we were choosing a library (we're not):

**KaTeX**:
- Much faster than MathJax v2
- Comparable to MathJax v3
- Better CSP compatibility
- Limited LaTeX support

**MathJax v3**:
- ~2x faster than v2
- More complete LaTeX support
- Handles TeX/LaTeX, MathML, AsciiMath
- ~100KB+ library size

Both require relaxing CSP and adding JavaScript.

---

## Conclusion

**DO NOT IMPLEMENT MATHJAX SUPPORT**

The costs outweigh the benefits:

| **Cost** | **Benefit** |
|----------|-------------|
| Security policy relaxation | Support for edge case feeds |
| External CDN dependency | Render math equations |
| 100KB+ JavaScript | (that most feeds don't have) |
| Client-side rendering complexity | (that users could handle themselves) |
| Maintenance burden | (with browser extensions) |
| Violates core principles | |

**Rogue Planet's strength is simplicity, security, and static output.** Adding MathJax undermines all three.

**If math support becomes critical**: Allow native MathML through sanitizer (30 min change, zero JavaScript, maintains security). But even this is questionable given minimal use case.

**Best recommendation**: Document that Rogue Planet doesn't process math. Users who need it can use browser extensions. Feed publishers who need it can pre-render or use MathML.

---

## References

### Web Research Conducted
- MathJax CSP requirements and evolution (v2 → v3)
- bluemonday sanitizer limitations with script tags
- RSS/Atom feed math rendering challenges (2012-2024)
- kramdown security vulnerabilities (CVE-2020-14001)
- MathJax security issues (CVE-2018-1999024)
- Native MathML browser support status (2024-2025)
- KaTeX vs MathJax performance comparisons
- Feed reader math rendering solutions

### Historical Context
- 12+ years of "math in RSS" being an unsolved problem
- Planet Venus had no special math support
- Academic blog aggregators expect browser extensions
- Community consensus: Not the aggregator's job

### Files Analyzed
- `pkg/generator/generator.go` (CSP at line 378)
- `pkg/normalizer/normalizer.go` (bluemonday at line 57)
- `specs/LESSONS_LEARNED.md` (core principles)
- `CLAUDE.md` (project architecture)

---

## Archive Note

This document is archived as research for future reference. The conclusion is **clear**: MathJax support should not be added to Rogue Planet as it conflicts with the project's core architectural principles of simplicity, security, and static output.

If mathematical notation becomes a genuine requirement in the future, the only acceptable path would be **native MathML support** (zero JavaScript, no CSP changes, minimal code). But even this is not recommended given the minimal use case.

**Date Archived**: 2025-10-21
**Research Conducted By**: Claude Code Agent (Explore mode)
**Decision**: Do Not Implement
