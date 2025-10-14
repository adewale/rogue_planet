# Rogue Planet Wishlist

This document captures potential features and improvements for Rogue Planet, many inspired by real issues encountered by Planet Venus users over the years.

## Sources

- Planet Venus GitHub issues: https://github.com/rubys/venus/issues
- Real-world usage patterns from Venus deployments
- Lessons learned from 20+ years of feed aggregation

## High Priority

### HTML Escaping in Titles (Venus Issue #24)

**Problem**: When feed titles contain HTML tags (like `<dialog>`, `<foo>`), they should be displayed as literal text, not stripped or interpreted as HTML.

**Current Behavior**: HTML sanitization strips tags from all fields including titles.

**Desired Behavior**:
- Titles, author names, and feed names should have HTML entities escaped (show `&lt;dialog&gt;` as `<dialog>`)
- Content and summary fields should continue using full HTML sanitization (allow safe subset, block dangerous tags)

**Rationale**: Data accuracy - users want to see exactly what's in the feed title, not have content silently removed.

**Implementation**: Use different sanitization policies for metadata (escape all HTML) vs content (allow safe HTML subset).

---

### Responsive Image Constraints (Venus Issue #36)

**Problem**: Large images in feed content overflow page layout, breaking responsive design.

**Current Behavior**: Images render at their natural size, potentially wider than the page content area.

**Desired Behavior**: Images should be constrained to the content width and scale proportionally.

**Rationale**: Basic UX issue affecting readability on all devices, especially mobile.

**Implementation**:
```css
.entry-content img {
    max-width: 100%;
    height: auto;
}
```

Add to default template CSS and document for custom templates.

---

### Prevent Entry Spam on New Feed Addition

**Problem**: When you add a new feed to a planet, it fetches ALL historical entries (often 50-100 entries), and they all appear in the generated HTML at once based on their original published dates. This:
- Spams the planet's RSS feed with old entries
- Pollutes the chronological timeline with backdated content
- Annoys readers whose feed readers show "100 new items"
- Was a major complaint about Venus/PlanetPlanet

**Current Status**: ❌ **Not Implemented**

**Priority**: **HIGH** - This addresses a major user pain point that drove people away from Venus

**Detailed Analysis**: See [ENTRY_SPAM.md](ENTRY_SPAM.md) for complete implementation proposal including:
- Problem statement and rationale
- Proposed solution with config options
- Implementation details
- Comprehensive testing strategy (time-independent, network-independent)
- Edge cases and migration path

**Quick Summary**:
- Database already has `first_seen` field ✅
- Index on `first_seen` already exists ✅
- Need to add config options: `filter_by_first_seen` and `sort_by`
- Need to update `GetRecentEntries()` to support these options
- Includes complete test suite design

---

### Stable Sort Dates (Venus Issue #15)

**Problem**: When feeds update entries (fixing typos, adding corrections), the entry's published/updated date changes, affecting chronological sort order. This creates instability in the aggregated timeline.

**Note**: This is related to the entry spam issue above but can be implemented independently.

**Current Behavior**: Entries sorted by published date (with fallback to updated/fetched time).

**Desired Behavior**: Option to sort by "first seen" date - the timestamp when Rogue Planet first fetched the entry - which never changes even if the entry is later updated.

**Rationale**:
- Predictable chronological ordering
- Entries don't "jump" in timeline when authors make corrections
- Better matches user expectations for "river of news" style aggregation

**Implementation**: Covered in [ENTRY_SPAM.md](ENTRY_SPAM.md) as part of the `sort_by` config option

---

## Medium Priority

### Auto-Reactivate Inactive Feeds (Venus Issue #34)

**Problem**: When a feed goes inactive (author stops posting), and later resumes, it stays marked as inactive in the database. Requires manual intervention to reactivate.

**Current Behavior**: Feeds marked inactive are skipped during fetches. Must be manually reactivated via database update.

**Desired Behavior**: When an inactive feed has new entries (detected by newer published dates than last fetch), automatically mark it active again.

**Rationale**: Blogs often go dormant and resume. Automatic reactivation improves user experience.

**Implementation**:
```
1. Fetch inactive feeds periodically (e.g., daily instead of hourly)
2. If new entries found: set active=1, fetch_error_count=0
3. Log reactivation for visibility
```

---

### Allow data-* Attributes (Venus PR #19)

**Problem**: HTML5 `data-*` attributes are harmless metadata but get stripped by sanitization, breaking third-party scripts and widgets that users add to templates.

**Current Behavior**: bluemonday strips data-* attributes from all elements.

**Desired Behavior**: Allow data-* attributes on safe elements (div, span, article, section, etc.).

**Rationale**:
- Enables template customization with JavaScript libraries
- data-* attributes are client-side only, no XSS risk
- Standard HTML5 practice for storing custom data

**Implementation**: Configure bluemonday policy to allow data-* attributes:
```go
policy.AllowDataAttributes()
```

---

### Future Dates Configuration

**Problem**: Some feeds contain entries with future dates (scheduled posts, timezone errors). Need policy for handling these.

**Current Behavior**: Not explicitly handled - accepted as-is.

**Desired Behavior**: Configuration option to ignore, accept, or clamp future-dated entries.

**Rationale**: Already in spec (line 244), needs implementation. Prevents timeline pollution from incorrectly dated entries.

**Implementation**:
```ini
[planet]
future_dates = accept | ignore_entry | clamp_to_now
```

- `accept`: Keep as-is (default)
- `ignore_entry`: Skip entries with pub date > now
- `clamp_to_now`: Set pub date to fetch time if in future

---

## Low Priority

### MathJax Support (Venus Issue #33)

**Problem**: Mathematical equations in feeds (especially from Blogger, academic blogs) don't render properly. MathML and TeX markup display as raw code.

**Current Behavior**: Math markup passes through as text.

**Desired Behavior**: Option to include MathJax library in templates for equation rendering.

**Rationale**: Important for academic/technical planet sites aggregating math/science blogs.

**Implementation**:
1. Document how to add MathJax to custom templates
2. Ensure sanitization allows MathML tags or TeX delimiters
3. Optional: Add `enable_mathjax` config option to include in default template

---

### Gravatar Support (Venus PR #10)

**Problem**: Planet pages are text-heavy. Adding author avatars improves visual appeal and recognition.

**Current Behavior**: No avatar support.

**Desired Behavior**: Display Gravatar images for feed authors based on their email addresses.

**Rationale**:
- Visual enhancement
- Helps readers identify authors quickly
- Standard practice on many community sites

**Implementation**:
1. Add optional `gravatar_email` field to feeds table
2. Add template function to generate Gravatar URL from email
3. Update default template to show avatars in sidebar and entries
4. Support default avatar options (identicon, monsterid, etc.)

---

### Enhanced Media Tag Support (Venus PR #18)

**Problem**: HTML5 video/audio tags need attributes like `preload`, `poster`, `controls` to work well, but sanitization strips them.

**Current Behavior**: Only basic video/audio tags allowed, missing important attributes.

**Desired Behavior**: Allow safe media attributes that improve user experience.

**Rationale**: Better embedding of multimedia content from feeds.

**Implementation**: Configure bluemonday to allow:
- `<video>`: preload, poster, controls, width, height
- `<audio>`: preload, controls
- `<source>`: src, type

---

### GitHub Gist Embeds (Venus Issue #28)

**Problem**: GitHub Gist embeds (iframe-based) don't display because iframes are blocked by sanitization.

**Current Behavior**: All iframes stripped for security.

**Desired Behavior**: Allow iframes from trusted domains (GitHub, CodePen, JSFiddle, etc.).

**Rationale**: Common use case for developer planet sites.

**Implementation**:
1. Add `trusted_iframe_domains` config option
2. Extend sanitization to check iframe src against whitelist
3. Only allow iframes with https URLs from approved domains
4. Set strict sandbox attribute: `sandbox="allow-scripts allow-same-origin"`

**Security Note**: Requires careful implementation to avoid SSRF and XSS risks.

---

## Already Implemented

These Planet Venus issues are already addressed in Rogue Planet:

### ✅ User-Agent Headers (Venus Issue #29)
- **Status**: Implemented in spec (lines 89, 164-165)
- **Implementation**: Rogue Planet sends proper User-Agent header on all requests
- Configuration: `user_agent` in config.ini

### ✅ XSS Prevention (CVE-2009-2937)
- **Status**: Core security feature (spec lines 548-762)
- **Implementation**: Full HTML sanitization using bluemonday
- Strips script tags, event handlers, dangerous URIs
- Uses html/template for auto-escaping

### ✅ Video Autoplay Filtering (Venus Issue #1)
- **Status**: Handled by bluemonday sanitization
- **Implementation**: Dangerous attributes like autoplay stripped from all tags

### ✅ HTTP Conditional Requests
- **Status**: Core feature (spec lines 64-156)
- **Implementation**: Proper ETag/Last-Modified support
- Reduces bandwidth and server load

### ✅ OPML Import/Export
- **Status**: v0.2.0 feature
- **Implementation**: Full OPML 1.0/2.0 support
- Commands: `rp import-opml`, `rp export-opml`

---

## Not Relevant to Rogue Planet

These Venus issues don't apply due to architectural differences:

- **Python 2 EOL issues**: Rogue Planet is written in Go
- **Template engine issues**: Venus uses Genshi, Rogue Planet uses Go html/template
- **Plugin system issues**: Rogue Planet doesn't have a plugin architecture (yet?)
- **Dependency management**: Different ecosystems (Python vs Go)
- **GeoRSS support**: Niche use case, low priority

---

## Contributing

Have ideas for Rogue Planet? See [CONTRIBUTING.md](CONTRIBUTING.md) for how to propose features.

Before implementing wishlist items:
1. Check if already implemented in current version
2. Discuss approach in GitHub issues
3. Ensure test coverage >75%
4. Update documentation
5. Follow patterns in CLAUDE.md

---

## Version History

- **2025-10-14**: Initial wishlist based on Planet Venus issue analysis
