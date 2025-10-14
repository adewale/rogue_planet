# OPML (Outline Processor Markup Language) Research Report

## Executive Summary

OPML (Outline Processor Markup Language) is an XML-based format for representing hierarchical outlines, primarily used by feed aggregators for importing and exporting subscription lists. This report provides comprehensive technical details for implementing 100% compliant OPML import/export functionality in Rogue Planet.

**Key Findings:**
- OPML 2.0 (2006) is the current and final specification
- Real-world OPML files frequently violate the specification
- The `text` vs `title` attribute is a common source of confusion and errors
- Major aggregators have varying levels of OPML compliance
- Robust parsing requires handling numerous edge cases

---

## 1. Official Specifications

### OPML 2.0 Specification (Current Standard)

**Official Source:** http://opml.org/spec2.opml

**Version Declaration:** OPML 2.0 is the last version of OPML. Any further development will take place in namespaces.

#### Document Structure

```xml
<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head>
    <!-- Metadata about the document -->
  </head>
  <body>
    <!-- One or more outline elements -->
  </body>
</opml>
```

#### Required Elements

- `<opml>` root element with `version` attribute (must be "2.0")
- `<head>` element (can be empty but must be present)
- `<body>` element (must contain at least one `<outline>` element)

#### Optional `<head>` Elements

All head elements are optional and can appear at most once:

- `<title>` - Document title
- `<dateCreated>` - Creation timestamp (RFC 822 format)
- `<dateModified>` - Last modification timestamp (RFC 822 format)
- `<ownerName>` - Document owner's name
- `<ownerEmail>` - Owner's email address
- `<ownerId>` - URL of page with author contact information
- `<docs>` - URL pointing to documentation
- `<expansionState>` - Comma-separated list of expanded line numbers
- `<vertScrollState>` - Pixel position of top displayed line
- Window position elements: `<windowTop>`, `<windowLeft>`, `<windowBottom>`, `<windowRight>`

#### `<outline>` Element

**Required Attributes:**
- `text` - The text to display for this outline node (REQUIRED)

**Common Optional Attributes:**
- `type` - Determines interpretation of other attributes (case-insensitive)
- `isComment` - "true" or "false" (default: "false")
- `isBreakpoint` - "true" or "false" (default: "false")
- `created` - RFC 822 timestamp of node creation
- `category` - Comma-separated list of categories

**For RSS/Atom Feeds (when `type="rss"`):**
- `xmlUrl` - URL of the feed (RSS/Atom XML) - REQUIRED for feeds
- `htmlUrl` - URL of the website (human-readable home page)
- `title` - Feed title (optional, derived from feed)
- `description` - Feed description
- `language` - Feed language (e.g., "en-us")
- `version` - Must be "RSS", "RSS1", "RSS2", or "scriptingNews"

**Nesting:**
- `<outline>` elements can be nested to arbitrary depth for hierarchical organization
- Empty outline elements (without nested children) represent leaf nodes

#### Date Format

All dates MUST conform to RFC 822 with the modification that the year should be expressed as four digits.

**RFC 822 Date Examples:**
```
Mon, 01 Jan 2024 12:00:00 GMT
Tue, 15 Oct 2024 09:30:45 -0500
Wed, 20 Dec 2023 14:15:00 +0000
```

**Format:** `Day, DD Mon YYYY HH:MM:SS +/-ZZZZ`

#### Character Encoding

- Default encoding: UTF-8
- XML declaration should specify encoding: `<?xml version="1.0" encoding="UTF-8"?>`
- All text must be properly XML-escaped

#### Extensibility

- New attributes/elements MUST use XML namespaces
- Processors should ignore unrecognized attributes
- New `type` values should clearly define their required and optional attributes

### OPML 1.0 Specification (Legacy)

**Official Source:** http://2005.opml.org/spec1.html

**Key Differences from 2.0:**
- No category attribute
- Less detailed specification for date formats
- Pre-1.0 documents used `<outlineDocument>` as root element instead of `<opml>`
- Otherwise largely compatible with OPML 2.0

**Recommendation:** Generate OPML 2.0, accept both 1.0 and 2.0 on import.

---

## 2. Feed Aggregator-Specific Usage

### Standard Feed Subscription Outline

For feed aggregators, an outline representing an RSS or Atom subscription typically has:

```xml
<outline
  type="rss"
  text="Go Blog"
  title="The Go Blog"
  xmlUrl="https://blog.golang.org/feed.atom"
  htmlUrl="https://blog.golang.org/"
  description="The official Go blog"
/>
```

### Attribute Definitions for Feeds

| Attribute | Required | Description | Notes |
|-----------|----------|-------------|-------|
| `text` | **YES** | Display name for outline | What will be shown in UI |
| `xmlUrl` | **YES** | Feed URL (RSS/Atom) | Actual feed to fetch |
| `type` | Recommended | Feed type | Usually "rss" (works for Atom too) |
| `title` | No | Feed title | Often same as `text` |
| `htmlUrl` | No | Website URL | Human-readable site |
| `description` | No | Feed description | Optional metadata |
| `language` | No | Language code | e.g., "en-us" |
| `version` | No | Feed version | "RSS", "RSS1", "RSS2", "scriptingNews" |

### Categories and Folders

Nested outline elements represent categories or folders:

```xml
<outline text="Technology">
  <outline type="rss" text="Hacker News" xmlUrl="..." />
  <outline type="rss" text="Ars Technica" xmlUrl="..." />

  <outline text="Programming">
    <outline type="rss" text="Go Blog" xmlUrl="..." />
    <outline type="rss" text="Rust Blog" xmlUrl="..." />
  </outline>
</outline>
```

**Important:** Category outline elements typically do NOT have `type` or `xmlUrl` attributes - they're just organizational containers.

### Major Aggregators' OPML Structure

#### Google Reader (Historical Reference)

- Used nested `<outline>` elements for folders
- Always included both `text` and `title` attributes (usually identical)
- Included `htmlUrl` for most feeds
- Used `type="rss"` for all feeds (including Atom)

#### Feedly

- Supports nested categories
- Flattens deeply nested hierarchies: "Category 1 - Sub A" as a single folder
- Preserves `text`, `title`, `xmlUrl`, and `htmlUrl`
- Uses `type="rss"` consistently

#### NewsBlur

- Export available at: `newsblur.com/import/opml_export`
- Supports category nesting
- Similar structure to Google Reader exports

#### Inoreader

- Full OPML import/export support
- Supports OPML subscriptions (subscribe to OPML URL for dynamic updates)
- Preserves folder structure

#### FreshRSS

- Standard OPML support with namespace extensions
- Custom namespace: `https://freshrss.org/opml`
- Extended attributes for web scraping:
  - HTML+XPath parsing
  - JSON+DotNotation parsing
  - Custom XPath/JSON extraction

### Type Attribute Values

While the specification allows any value, common values in feed aggregators:

- `type="rss"` - RSS or Atom feed (most common)
- `type="atom"` - Atom feed (rarely used, `rss` works for both)
- `type="link"` - Simple link (not a feed)
- No type attribute - Category/folder container

**Important:** Most feed readers treat Atom feeds as `type="rss"` for compatibility.

---

## 3. Edge Cases and Quirks

### The `text` vs `title` Attribute Problem

**The Specification:** The `text` attribute is REQUIRED for all outline elements. For `type="rss"` nodes, `title` is optional and represents the feed's title.

**The Reality:** Many early RSS readers incorrectly used `title` instead of `text`, and this error has propagated throughout the ecosystem.

**The Confusion:**
- `text` = What's displayed in the outline/UI
- `title` = The actual title of the feed (from the feed's metadata)
- In practice, they're often identical

**Best Practice for Parsing:**
```go
// When reading OPML
displayName := outline.Text
if displayName == "" {
    displayName = outline.Title  // Fallback for non-compliant OPML
}
if displayName == "" {
    displayName = "Untitled Feed"  // Last resort
}

// When writing OPML
outline.Text = feedTitle        // Required
outline.Title = feedTitle       // Optional but recommended for compatibility
```

**Recommendation:**
- **Reading:** Accept either `text` or `title` (prefer `text`)
- **Writing:** Include both with identical values for maximum compatibility

### Missing `xmlUrl`

Some OPML files have outline elements without `xmlUrl`:

```xml
<outline text="Some Blog" htmlUrl="https://example.com/" />
```

**Handling:**
- Skip outlines without `xmlUrl` when importing feeds
- These might be categories, bookmarks, or incomplete entries
- Log a warning for debugging

### Empty Outline Elements

```xml
<outline text="Empty Category" />
```

**Valid?** Yes, according to spec (outline can have no children)

**Handling:**
- If it has `xmlUrl`, treat as feed
- If no `xmlUrl`, treat as empty category (or skip)

### Nested Categories - Depth Limits

**Specification:** No depth limit defined

**Real-world:**
- Most aggregators support 2-3 levels comfortably
- Some aggregators flatten deep nesting
- Very deep nesting (>5 levels) is rare

**Recommendation:** Support arbitrary depth, but test with at least 5 levels

### Character Encoding Issues

**UTF-8 BOM (Byte Order Mark):**
- UTF-8 files may start with BOM bytes: `EF BB BF`
- XML parsers should handle this automatically
- If not, skip the BOM before parsing

**Encoding Declaration Mismatches:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!-- File is actually ISO-8859-1 -->
```

**Handling:**
- Trust the XML declaration first
- If parsing fails, try UTF-8 regardless of declaration
- As last resort, try common encodings (ISO-8859-1, Windows-1252)

### XML Escaping Issues

**Common Problems:**
- Unescaped `&` in URLs: `xmlUrl="https://example.com/feed?a=1&b=2"`
- Unescaped `<` and `>` in text
- Unescaped quotes in attribute values

**Valid Escaping:**
```xml
<outline
  text="Q&amp;A Blog"
  xmlUrl="https://example.com/feed?a=1&amp;b=2"
  description="Answers to &quot;common&quot; questions"
/>
```

**Handling:**
- Use XML parser that handles entity decoding
- For generation, ensure proper escaping (Go's `encoding/xml` does this automatically)

### Malformed XML

**Common Issues:**
- Missing closing tags
- Unclosed `<outline>` elements
- Invalid attribute syntax
- Comments in wrong places

**Handling Strategy:**
- Try strict XML parsing first
- If that fails, try lenient parsing (if available)
- Log errors with specific line numbers
- Skip malformed entries but continue processing

### Comments and Processing Instructions

```xml
<?xml version="1.0"?>
<!-- This is my feed list -->
<?xml-stylesheet type="text/xsl" href="style.xsl"?>
<opml version="2.0">
  <head>
    <!-- Generated on 2024-10-12 -->
    <title>My Feeds</title>
  </head>
  <body>
    <!-- Technology Section -->
    <outline text="Tech Feeds">
      <outline type="rss" text="Example" xmlUrl="..." />
    </outline>
  </body>
</opml>
```

**Valid?** Yes, XML comments and processing instructions are allowed

**Handling:** XML parsers ignore these by default

### Duplicate Feed URLs

```xml
<outline text="Tech News" xmlUrl="https://example.com/feed" />
<outline text="Example Feed" xmlUrl="https://example.com/feed" />
```

**Handling Options:**
1. Import both (allow duplicates with different display names)
2. Skip duplicates (import only first occurrence)
3. Merge (use first name, note duplicate in logs)

**Recommendation:** Option 3 - import once, log warning about duplicate

### Invalid URLs

```xml
<outline text="Broken" xmlUrl="htp://example.com/feed" />
<outline text="Relative" xmlUrl="/feed.xml" />
<outline text="Empty" xmlUrl="" />
```

**Handling:**
- Validate URLs during import
- Skip invalid URLs with error message
- Don't allow relative URLs (no base URL context)

### Future Dates in `dateCreated` / `dateModified`

```xml
<head>
  <dateModified>Mon, 01 Jan 2099 00:00:00 GMT</dateModified>
</head>
```

**Handling:**
- Accept dates as-is (metadata only, not critical)
- Optionally log warning for obviously wrong dates
- Don't reject entire file due to bad dates

### Missing `<body>` or `<head>` Elements

**Malformed:**
```xml
<opml version="2.0">
  <outline text="Feed 1" xmlUrl="..." />
</opml>
```

**Specification:** Both `<head>` and `<body>` are required

**Reality:** Some exporters generate files without `<head>`

**Handling:**
- Try to parse anyway if `<outline>` elements exist
- Log warning about non-compliant structure
- Be liberal in what you accept

### Namespace Handling

**With Namespaces:**
```xml
<opml version="2.0" xmlns:custom="http://example.com/ns">
  <head>
    <custom:metadata>Custom data</custom:metadata>
  </head>
  <body>
    <outline text="Feed" xmlUrl="..." custom:priority="high" />
  </body>
</opml>
```

**Handling:**
- Ignore unknown namespaces during import (per spec)
- Don't fail on unknown elements/attributes
- Preserve namespace declarations on round-trip if possible

### Case Sensitivity

**Specification:** XML is case-sensitive

**Reality:**
- Element names: `<opml>`, `<OPML>`, `<Opml>` are all different
- Attribute names: `text`, `TEXT`, `Text` are different
- Attribute values: `type="rss"`, `type="RSS"` are different

**But:** The spec says `type` attribute is case-insensitive for comparison

**Handling:**
- Parse XML case-sensitively
- Compare `type` attribute values case-insensitively
- Generate lowercase element/attribute names

---

## 4. Best Practices for Generation

### Minimal Valid OPML 2.0

```xml
<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head>
    <title>My Feeds</title>
    <dateCreated>Sat, 12 Oct 2024 10:00:00 GMT</dateCreated>
    <dateModified>Sat, 12 Oct 2024 10:00:00 GMT</dateModified>
    <ownerName>John Doe</ownerName>
  </head>
  <body>
    <outline text="Go Blog"
             type="rss"
             xmlUrl="https://blog.golang.org/feed.atom"
             htmlUrl="https://blog.golang.org/" />
  </body>
</opml>
```

### Complete Example with Categories

```xml
<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head>
    <title>My Feed Subscriptions</title>
    <dateCreated>Sat, 12 Oct 2024 10:00:00 GMT</dateCreated>
    <dateModified>Sat, 12 Oct 2024 10:00:00 GMT</dateModified>
    <ownerName>John Doe</ownerName>
    <ownerEmail>john@example.com</ownerEmail>
    <docs>http://opml.org/spec2.opml</docs>
  </head>
  <body>
    <outline text="Technology">
      <outline text="Go Blog"
               title="The Go Blog"
               type="rss"
               xmlUrl="https://blog.golang.org/feed.atom"
               htmlUrl="https://blog.golang.org/"
               description="The official Go programming language blog" />
      <outline text="Hacker News"
               title="Hacker News"
               type="rss"
               xmlUrl="https://news.ycombinator.com/rss"
               htmlUrl="https://news.ycombinator.com/" />
    </outline>

    <outline text="News">
      <outline text="BBC News"
               type="rss"
               xmlUrl="http://feeds.bbci.co.uk/news/rss.xml"
               htmlUrl="https://www.bbc.com/news" />
    </outline>

    <outline text="Uncategorized">
      <outline text="Example Blog"
               type="rss"
               xmlUrl="https://example.com/feed.xml"
               htmlUrl="https://example.com/" />
    </outline>
  </body>
</opml>
```

### Metadata Best Practices

1. **Always include:**
   - `<?xml version="1.0" encoding="UTF-8"?>`
   - `<opml version="2.0">`
   - `<title>` in head
   - `dateModified` (current timestamp)

2. **Recommended:**
   - `dateCreated` (if known)
   - `ownerName` (for attribution)
   - Both `text` and `title` on feed outlines (compatibility)

3. **Optional but useful:**
   - `ownerEmail`
   - `description` for feeds
   - `htmlUrl` for feeds

### Character Escaping

**Must Escape:**
- `&` → `&amp;`
- `<` → `&lt;`
- `>` → `&gt;`
- `"` → `&quot;` (in attribute values)
- `'` → `&apos;` (in attribute values)

**Go's `encoding/xml` handles this automatically.**

### Date Format Generation

```go
// Generate RFC 822 date with 4-digit year
func formatOPMLDate(t time.Time) string {
    // RFC1123 is RFC 822 with 4-digit year
    return t.UTC().Format(time.RFC1123)
}
// Example: "Sat, 12 Oct 2024 10:00:00 GMT"
```

### Indentation and Formatting

**Human-readable:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head>
    <title>My Feeds</title>
  </head>
  <body>
    <outline text="Category">
      <outline text="Feed" type="rss" xmlUrl="..." />
    </outline>
  </body>
</opml>
```

**Go's `encoding/xml` with `MarshalIndent()` produces properly formatted output.**

### Attribute Order

The specification doesn't mandate attribute order, but a consistent order improves readability:

**Recommended order:**
1. `text` (required, most important)
2. `title` (if present)
3. `type`
4. `xmlUrl`
5. `htmlUrl`
6. `description`
7. Other attributes alphabetically

**Note:** XML parsers don't care about order; this is purely for human readability.

---

## 5. Real-World Examples

### Typical Feedly Export Structure

```xml
<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <head>
    <title>Subscriptions from Feedly</title>
  </head>
  <body>
    <outline text="tech" title="tech">
      <outline type="rss" text="Hacker News"
               title="Hacker News"
               xmlUrl="https://news.ycombinator.com/rss"
               htmlUrl="https://news.ycombinator.com/" />
    </outline>
    <outline text="news" title="news">
      <outline type="rss" text="BBC News - Home"
               title="BBC News - Home"
               xmlUrl="http://feeds.bbci.co.uk/news/rss.xml" />
    </outline>
  </body>
</opml>
```

**Observations:**
- Uses OPML 1.0 (but compatible with 2.0 parsers)
- Categories have both `text` and `title` attributes
- Some feeds missing `htmlUrl`
- Feed outline elements also duplicate `text` and `title`

### Typical NewsBlur Export

```xml
<?xml version="1.0" encoding="utf-8"?>
<opml version="1.0">
  <head>
    <title>NewsBlur Feeds</title>
  </head>
  <body>
    <outline text="Technology">
      <outline text="Ars Technica"
               type="rss"
               xmlUrl="http://feeds.arstechnica.com/arstechnica/index"
               htmlUrl="https://arstechnica.com" />
    </outline>
  </body>
</opml>
```

### Google Reader Export (Historical)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <head>
    <title>Google Reader Subscriptions</title>
  </head>
  <body>
    <outline title="Technology" text="Technology">
      <outline text="Coding Horror"
               title="Coding Horror"
               type="rss"
               xmlUrl="http://feeds.feedburner.com/codinghorror"
               htmlUrl="http://www.codinghorror.com/blog/" />
    </outline>
  </body>
</opml>
```

### Minimal/Broken Examples Found in the Wild

**Missing `type` attribute:**
```xml
<outline text="Some Blog"
         xmlUrl="https://example.com/feed.xml" />
```
**Handling:** Assume `type="rss"` if `xmlUrl` is present

**Relative URLs:**
```xml
<outline text="Blog" xmlUrl="/feed.xml" />
```
**Handling:** Reject - no context to resolve relative URL

**Mixed content (feeds and links):**
```xml
<body>
  <outline type="rss" text="Blog" xmlUrl="..." />
  <outline type="link" text="Website" url="..." />
</body>
```
**Handling:** Skip non-RSS outlines, only import feed subscriptions

---

## 6. Test Cases

### Required Test Scenarios

#### Valid OPML Files

1. **Minimal valid OPML 2.0**
   ```xml
   <?xml version="1.0" encoding="UTF-8"?>
   <opml version="2.0">
     <head><title>Test</title></head>
     <body>
       <outline text="Feed" type="rss" xmlUrl="https://example.com/feed.xml" />
     </body>
   </opml>
   ```

2. **OPML 1.0 (backward compatibility)**
   ```xml
   <?xml version="1.0"?>
   <opml version="1.0">
     <head><title>Test</title></head>
     <body>
       <outline text="Feed" xmlUrl="https://example.com/feed.xml" />
     </body>
   </opml>
   ```

3. **With nested categories (2 levels)**
   ```xml
   <body>
     <outline text="Tech">
       <outline text="Programming">
         <outline text="Go Blog" type="rss" xmlUrl="..." />
       </outline>
     </outline>
   </body>
   ```

4. **With nested categories (5 levels - stress test)**

5. **Multiple feeds in one category**

6. **Multiple categories with multiple feeds**

7. **Feeds with all optional attributes**
   ```xml
   <outline text="Blog"
            title="The Blog"
            type="rss"
            xmlUrl="..."
            htmlUrl="..."
            description="A blog"
            language="en-us"
            version="RSS2" />
   ```

8. **With full head metadata**
   ```xml
   <head>
     <title>My Feeds</title>
     <dateCreated>Mon, 01 Jan 2024 00:00:00 GMT</dateCreated>
     <dateModified>Tue, 02 Jan 2024 00:00:00 GMT</dateModified>
     <ownerName>John Doe</ownerName>
     <ownerEmail>john@example.com</ownerEmail>
     <docs>http://opml.org/spec2.opml</docs>
   </head>
   ```

9. **Empty categories**

10. **Special characters in text (escaping test)**
    ```xml
    <outline text="Q&amp;A Blog &amp; More &quot;Quotes&quot;"
             xmlUrl="https://example.com/feed?a=1&amp;b=2" />
    ```

#### Edge Cases

11. **`title` but no `text` attribute (non-compliant)**
    ```xml
    <outline title="Feed" type="rss" xmlUrl="..." />
    ```
    **Expected:** Use `title` as fallback for display name

12. **Both `text` and `title` (different values)**
    ```xml
    <outline text="My Name" title="Official Name" type="rss" xmlUrl="..." />
    ```
    **Expected:** Use `text` for display

13. **Missing `xmlUrl` (non-feed outline)**
    ```xml
    <outline text="Just a Category" />
    ```
    **Expected:** Skip or treat as category

14. **Empty `xmlUrl`**
    ```xml
    <outline text="Broken" type="rss" xmlUrl="" />
    ```
    **Expected:** Skip with error

15. **Invalid URL in `xmlUrl`**
    ```xml
    <outline text="Bad" xmlUrl="not-a-url" />
    ```
    **Expected:** Skip with validation error

16. **Relative URL in `xmlUrl`**
    ```xml
    <outline text="Relative" xmlUrl="/feed.xml" />
    ```
    **Expected:** Skip with error (no base URL)

17. **Duplicate feed URLs**
    ```xml
    <outline text="Feed 1" xmlUrl="https://example.com/feed" />
    <outline text="Feed 2" xmlUrl="https://example.com/feed" />
    ```
    **Expected:** Import once, log warning

18. **No `type` attribute but has `xmlUrl`**
    ```xml
    <outline text="Feed" xmlUrl="https://example.com/feed" />
    ```
    **Expected:** Assume `type="rss"`

19. **`type` is uppercase**
    ```xml
    <outline text="Feed" type="RSS" xmlUrl="..." />
    ```
    **Expected:** Accept (case-insensitive comparison)

20. **`type="atom"` instead of `type="rss"`**
    ```xml
    <outline text="Feed" type="atom" xmlUrl="..." />
    ```
    **Expected:** Accept (treat like RSS)

21. **Unknown `type` value**
    ```xml
    <outline text="Something" type="unknown" xmlUrl="..." />
    ```
    **Expected:** Skip or attempt import with warning

#### Encoding and Format Issues

22. **UTF-8 with BOM**
    ```
    EF BB BF <?xml version="1.0"...
    ```
    **Expected:** Parse successfully (skip BOM)

23. **ISO-8859-1 encoding**
    ```xml
    <?xml version="1.0" encoding="ISO-8859-1"?>
    ```
    **Expected:** Parse successfully with correct encoding

24. **Encoding declaration mismatch** (declares UTF-8 but is actually ISO-8859-1)
    **Expected:** Try declared encoding first, fall back if parsing fails

25. **No XML declaration**
    ```xml
    <opml version="2.0">
    ```
    **Expected:** Parse successfully (assume UTF-8)

26. **CRLF line endings (Windows)**
    **Expected:** Parse successfully

27. **LF line endings (Unix)**
    **Expected:** Parse successfully

28. **Mixed line endings**
    **Expected:** Parse successfully

#### Malformed XML

29. **Unclosed outline tag**
    ```xml
    <outline text="Feed" type="rss" xmlUrl="...">
    ```
    **Expected:** Parsing error (reject file or skip outline)

30. **Missing closing `</body>` tag**
    **Expected:** Parsing error

31. **Unescaped `&` in URL**
    ```xml
    <outline xmlUrl="https://example.com/feed?a=1&b=2" />
    ```
    **Expected:** Parsing error (strict) or attempt to fix (lenient)

32. **Unescaped `<` in text**
    ```xml
    <outline text="Cost < $10" />
    ```
    **Expected:** Parsing error

33. **Self-closing vs explicit closing**
    ```xml
    <outline text="Feed" xmlUrl="..." />  <!-- Self-closing -->
    <outline text="Feed" xmlUrl="..."></outline>  <!-- Explicit -->
    ```
    **Expected:** Both valid, parse successfully

#### Whitespace Handling

34. **Extra whitespace in attributes**
    ```xml
    <outline text="  Feed  " xmlUrl=" https://example.com/feed " />
    ```
    **Expected:** Trim whitespace from attribute values

35. **Newlines in attribute values**
    ```xml
    <outline text="Feed
    Name" />
    ```
    **Expected:** Normalize whitespace or reject

36. **Indentation variations**
    **Expected:** Parse successfully regardless of indentation

#### Comments and Processing Instructions

37. **XML comments in various locations**
    ```xml
    <!-- Comment before opml -->
    <opml>
      <head><!-- Comment in head --></head>
      <body>
        <!-- Comment before outline -->
        <outline text="Feed" />
        <!-- Comment after outline -->
      </body>
    </opml>
    ```
    **Expected:** Ignore comments, parse successfully

38. **Processing instructions**
    ```xml
    <?xml-stylesheet type="text/xsl" href="style.xsl"?>
    ```
    **Expected:** Ignore, parse successfully

#### Date Format Variations

39. **Valid RFC 822 date**
    ```xml
    <dateCreated>Mon, 01 Jan 2024 12:00:00 GMT</dateCreated>
    ```
    **Expected:** Parse successfully

40. **RFC 822 with timezone offset**
    ```xml
    <dateCreated>Mon, 01 Jan 2024 12:00:00 -0500</dateCreated>
    ```
    **Expected:** Parse successfully

41. **2-digit year (old RFC 822)**
    ```xml
    <dateCreated>Mon, 01 Jan 24 12:00:00 GMT</dateCreated>
    ```
    **Expected:** Accept (convert to 2024 using heuristic)

42. **Invalid date format**
    ```xml
    <dateCreated>2024-01-01</dateCreated>
    ```
    **Expected:** Log warning, continue parsing

43. **Future date**
    ```xml
    <dateCreated>Mon, 01 Jan 2099 00:00:00 GMT</dateCreated>
    ```
    **Expected:** Accept (metadata only)

#### Namespace Handling

44. **Custom namespace**
    ```xml
    <opml version="2.0" xmlns:custom="http://example.com/ns">
      <body>
        <outline text="Feed" xmlUrl="..." custom:priority="high" />
      </body>
    </opml>
    ```
    **Expected:** Ignore custom attributes, parse standard attributes

45. **Default namespace**
    ```xml
    <opml version="2.0" xmlns="http://opml.org/ns">
    ```
    **Expected:** Parse successfully

#### Empty and Missing Elements

46. **Empty `<head>` element**
    ```xml
    <head></head>
    ```
    **Expected:** Valid, parse successfully

47. **Missing `<head>` element** (non-compliant)
    ```xml
    <opml version="2.0">
      <body>
        <outline text="Feed" xmlUrl="..." />
      </body>
    </opml>
    ```
    **Expected:** Lenient parsing - accept if body exists

48. **Empty `<body>` element**
    ```xml
    <body></body>
    ```
    **Expected:** Valid but useless (no feeds to import)

49. **`<body>` with only comments**
    ```xml
    <body>
      <!-- All feeds removed -->
    </body>
    ```
    **Expected:** Valid (no feeds to import)

#### Large Files (Performance/Stress Tests)

50. **1000 feeds in flat structure**
    **Expected:** Parse in reasonable time (<1 second)

51. **10-level deep category nesting**
    **Expected:** Parse successfully, handle deep recursion

52. **Very long attribute values** (e.g., 10KB description)
    **Expected:** Parse successfully (within reason)

---

## 7. Implementation Recommendations

### Parsing Strategy

#### Use Go's Standard XML Parser

```go
import "encoding/xml"

type OPML struct {
    XMLName xml.Name `xml:"opml"`
    Version string   `xml:"version,attr"`
    Head    Head     `xml:"head"`
    Body    Body     `xml:"body"`
}

type Head struct {
    Title        string `xml:"title,omitempty"`
    DateCreated  string `xml:"dateCreated,omitempty"`
    DateModified string `xml:"dateModified,omitempty"`
    OwnerName    string `xml:"ownerName,omitempty"`
    OwnerEmail   string `xml:"ownerEmail,omitempty"`
    OwnerID      string `xml:"ownerId,omitempty"`
    Docs         string `xml:"docs,omitempty"`
}

type Body struct {
    Outlines []Outline `xml:"outline"`
}

type Outline struct {
    Text        string    `xml:"text,attr"`
    Title       string    `xml:"title,attr,omitempty"`
    Type        string    `xml:"type,attr,omitempty"`
    XMLURL      string    `xml:"xmlUrl,attr,omitempty"`
    HTMLURL     string    `xml:"htmlUrl,attr,omitempty"`
    Description string    `xml:"description,attr,omitempty"`
    Language    string    `xml:"language,attr,omitempty"`
    Version     string    `xml:"version,attr,omitempty"`
    Outlines    []Outline `xml:"outline,omitempty"` // Nested outlines
}
```

#### Robust Parsing Function

```go
func ParseOPML(r io.Reader) (*OPML, error) {
    // Read all data to detect BOM
    data, err := io.ReadAll(r)
    if err != nil {
        return nil, fmt.Errorf("read error: %w", err)
    }

    // Skip UTF-8 BOM if present
    data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

    // Parse XML
    var opml OPML
    if err := xml.Unmarshal(data, &opml); err != nil {
        return nil, fmt.Errorf("parse error: %w", err)
    }

    // Validate structure
    if opml.Version == "" {
        return nil, errors.New("missing version attribute")
    }

    // Version check (accept 1.0 and 2.0)
    if opml.Version != "1.0" && opml.Version != "2.0" {
        return nil, fmt.Errorf("unsupported version: %s", opml.Version)
    }

    return &opml, nil
}
```

#### Feed Extraction with Validation

```go
type Feed struct {
    Title   string
    FeedURL string
    WebURL  string
    Category string
}

func ExtractFeeds(opml *OPML) ([]Feed, []error) {
    var feeds []Feed
    var errors []error

    // Recursive extraction
    var extract func(outlines []Outline, category string)
    extract = func(outlines []Outline, category string) {
        for _, outline := range outlines {
            // Determine display name (text vs title)
            displayName := strings.TrimSpace(outline.Text)
            if displayName == "" {
                displayName = strings.TrimSpace(outline.Title)
            }
            if displayName == "" {
                displayName = "Untitled"
            }

            // Check if this is a feed (has xmlUrl)
            xmlURL := strings.TrimSpace(outline.XMLURL)
            if xmlURL != "" {
                // This is a feed - validate and add
                if err := validateFeedURL(xmlURL); err != nil {
                    errors = append(errors, fmt.Errorf(
                        "invalid URL for '%s': %w", displayName, err))
                    continue
                }

                htmlURL := strings.TrimSpace(outline.HTMLURL)

                feeds = append(feeds, Feed{
                    Title:    displayName,
                    FeedURL:  xmlURL,
                    WebURL:   htmlURL,
                    Category: category,
                })
            } else if len(outline.Outlines) > 0 {
                // This is a category - recurse with updated category path
                newCategory := displayName
                if category != "" {
                    newCategory = category + " / " + displayName
                }
                extract(outline.Outlines, newCategory)
            }
            // Else: neither feed nor category with children - skip
        }
    }

    extract(opml.Body.Outlines, "")

    return feeds, errors
}

func validateFeedURL(urlStr string) error {
    parsed, err := url.Parse(urlStr)
    if err != nil {
        return fmt.Errorf("parse error: %w", err)
    }

    // Require absolute URL with scheme
    if !parsed.IsAbs() {
        return errors.New("relative URLs not supported")
    }

    // Only allow http and https
    if parsed.Scheme != "http" && parsed.Scheme != "https" {
        return fmt.Errorf("unsupported scheme: %s", parsed.Scheme)
    }

    return nil
}
```

### Generation Strategy

#### OPML Generation Function

```go
func GenerateOPML(feeds []Feed, title, ownerName, ownerEmail string) (*OPML, error) {
    opml := &OPML{
        Version: "2.0",
        Head: Head{
            Title:        title,
            DateModified: formatOPMLDate(time.Now()),
            OwnerName:    ownerName,
            OwnerEmail:   ownerEmail,
            Docs:         "http://opml.org/spec2.opml",
        },
    }

    // Group feeds by category
    categoryMap := make(map[string][]Feed)
    for _, feed := range feeds {
        category := feed.Category
        if category == "" {
            category = "Uncategorized"
        }
        categoryMap[category] = append(categoryMap[category], feed)
    }

    // Build outline structure
    for category, categoryFeeds := range categoryMap {
        categoryOutline := Outline{
            Text: category,
        }

        for _, feed := range categoryFeeds {
            feedOutline := Outline{
                Text:        feed.Title,
                Title:       feed.Title, // Duplicate for compatibility
                Type:        "rss",
                XMLURL:      feed.FeedURL,
                HTMLURL:     feed.WebURL,
            }
            categoryOutline.Outlines = append(categoryOutline.Outlines, feedOutline)
        }

        opml.Body.Outlines = append(opml.Body.Outlines, categoryOutline)
    }

    return opml, nil
}

func WriteOPML(w io.Writer, opml *OPML) error {
    // Write XML declaration
    if _, err := w.Write([]byte(xml.Header)); err != nil {
        return err
    }

    // Marshal with indentation
    encoder := xml.NewEncoder(w)
    encoder.Indent("", "  ")
    if err := encoder.Encode(opml); err != nil {
        return fmt.Errorf("encode error: %w", err)
    }

    // Add final newline
    if _, err := w.Write([]byte("\n")); err != nil {
        return err
    }

    return nil
}

func formatOPMLDate(t time.Time) string {
    // RFC1123 is RFC 822 with 4-digit year
    return t.UTC().Format(time.RFC1123)
}
```

### Error Handling

**Principle:** Be liberal in what you accept, conservative in what you generate.

**Import Errors:**
- Log all parsing errors but don't fail entire import
- Skip individual feeds with invalid URLs
- Report warnings for non-compliant OPML
- Return both successfully imported feeds AND list of errors

**Export:**
- Always generate valid OPML 2.0
- Include all recommended metadata
- Properly escape all XML entities
- Validate URLs before adding to OPML

### Testing Approach

1. **Unit Tests:**
   - Test parsing each edge case
   - Test generation with various feed combinations
   - Test round-trip (parse → generate → parse)

2. **Integration Tests:**
   - Test with real OPML files from major aggregators
   - Save snapshots of real exports as test fixtures

3. **Fuzzing:**
   - Use Go's fuzzing support to find edge cases
   - Generate random XML and test parser robustness

4. **Validation:**
   - Validate generated OPML against XML schema (if available)
   - Test with online OPML validators

---

## 8. Common Pitfalls to Avoid

### For Parsing

1. **Don't reject entire file for one bad feed**
   - Skip invalid feeds, import the rest

2. **Don't assume `type` attribute is present**
   - If `xmlUrl` exists, assume it's a feed

3. **Don't fail on unknown attributes**
   - Ignore gracefully (per specification)

4. **Don't require both `text` and `title`**
   - Accept either, prefer `text`

5. **Don't ignore encoding declaration**
   - Respect declared encoding

6. **Don't forget to trim whitespace**
   - URLs with leading/trailing spaces are common

7. **Don't allow relative URLs**
   - No context to resolve them

8. **Don't ignore validation**
   - Check URLs are well-formed before importing

### For Generation

1. **Don't forget XML declaration**
   - Always include `<?xml version="1.0" encoding="UTF-8"?>`

2. **Don't forget to escape XML entities**
   - Let library handle it (Go's `encoding/xml` does this)

3. **Don't use non-standard date formats**
   - Always use RFC 822 (RFC1123 in Go)

4. **Don't omit required attributes**
   - Every outline must have `text` attribute
   - Feed outlines must have `xmlUrl`

5. **Don't generate invalid URLs**
   - Validate before adding to OPML

6. **Don't forget `version` attribute on `<opml>`**
   - Required by specification

---

## 9. Summary and Quick Reference

### For Rogue Planet Implementation

#### Import Command: `rp import-opml <file>`

**Steps:**
1. Read OPML file
2. Parse XML (handle BOM, encoding issues)
3. Validate version (accept 1.0 and 2.0)
4. Recursively extract feeds:
   - Look for outlines with `xmlUrl`
   - Use `text` attribute (fallback to `title`)
   - Validate URLs
   - Track category path for nested outlines
5. Add feeds to database:
   - Skip duplicates (same feed URL)
   - Log warnings for invalid/skipped feeds
6. Report results:
   - Number of feeds imported
   - Number of errors/skipped feeds
   - List of errors

#### Export Command: `rp export-opml [file]`

**Steps:**
1. Query all active feeds from database
2. Group by category (if categories implemented)
3. Build OPML structure:
   - Set version="2.0"
   - Add head metadata (title, dates, owner)
   - Create category outlines
   - Create feed outlines with all attributes
4. Generate XML with proper formatting
5. Write to file or stdout

#### Validation Rules

**On Import:**
- Accept if parseable, even if not 100% spec-compliant
- Require: `<opml>`, `<body>`, at least one `<outline>`
- Optional: `<head>`, head metadata
- Feed outline must have: `text` OR `title`, and `xmlUrl`
- Validate URLs: must be absolute http/https URLs

**On Export:**
- Generate fully compliant OPML 2.0
- Include: XML declaration, version, head, body
- Use proper RFC 822 dates
- Escape all XML entities
- Include both `text` and `title` for compatibility

### Quick Reference Table

| Element/Attribute | Required? | Purpose | Notes |
|-------------------|-----------|---------|-------|
| `<opml>` | Yes | Root element | Must have `version` attribute |
| `version` | Yes | OPML version | Use "2.0" for generation |
| `<head>` | Yes | Metadata | Can be empty |
| `<body>` | Yes | Contains outlines | Must have ≥1 `<outline>` |
| `<outline>` | Yes | Feed or category | Can be nested |
| `text` | Yes* | Display name | *Required per spec |
| `title` | No | Feed title | Include for compatibility |
| `type` | No | Outline type | Use "rss" for feeds |
| `xmlUrl` | Yes* | Feed URL | *Required for feeds |
| `htmlUrl` | No | Website URL | Recommended |
| `description` | No | Feed description | Optional |

### Example Complete Implementation Flow

```
IMPORT:
File → Parse XML → Validate structure → Extract feeds recursively →
Validate URLs → Skip duplicates → Add to database → Report results

EXPORT:
Query database → Group by category → Build OPML tree →
Add metadata → Generate XML → Write file
```

---

## 10. Additional Resources

### Official Specifications
- OPML 2.0: http://opml.org/spec2.opml
- OPML 1.0: http://2005.opml.org/spec1.html

### Related Specifications
- RFC 822 (dates): https://www.ietf.org/rfc/rfc822.txt
- RFC 1123 (updated dates): https://www.ietf.org/rfc/rfc1123.txt
- XML 1.0: https://www.w3.org/TR/xml/

### Useful Tools
- OPML Validator: http://validator.opml.org/ (if available)
- Feed readers for testing: Feedly, NewsBlur, Inoreader
- XML validators

### Go Libraries
- Standard library: `encoding/xml`
- No external dependencies required for basic OPML support

---

## Appendix: Complete Test OPML Files

### A. Minimal Valid OPML

```xml
<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head>
    <title>Minimal Test</title>
  </head>
  <body>
    <outline text="Example Feed"
             type="rss"
             xmlUrl="https://example.com/feed.xml" />
  </body>
</opml>
```

### B. Complete OPML with All Features

```xml
<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head>
    <title>Complete Test Feed List</title>
    <dateCreated>Sat, 12 Oct 2024 10:00:00 GMT</dateCreated>
    <dateModified>Sat, 12 Oct 2024 12:30:00 GMT</dateModified>
    <ownerName>Test User</ownerName>
    <ownerEmail>test@example.com</ownerEmail>
    <ownerId>https://example.com/about</ownerId>
    <docs>http://opml.org/spec2.opml</docs>
  </head>
  <body>
    <!-- Technology Category -->
    <outline text="Technology">
      <!-- Programming Subcategory -->
      <outline text="Programming">
        <outline text="Go Blog"
                 title="The Go Blog"
                 type="rss"
                 xmlUrl="https://blog.golang.org/feed.atom"
                 htmlUrl="https://blog.golang.org/"
                 description="The official Go programming language blog"
                 language="en-us" />

        <outline text="Rust Blog"
                 title="Rust Blog"
                 type="rss"
                 xmlUrl="https://blog.rust-lang.org/feed.xml"
                 htmlUrl="https://blog.rust-lang.org/" />
      </outline>

      <!-- General Tech -->
      <outline text="Hacker News"
               title="Hacker News"
               type="rss"
               xmlUrl="https://news.ycombinator.com/rss"
               htmlUrl="https://news.ycombinator.com/" />
    </outline>

    <!-- News Category -->
    <outline text="News">
      <outline text="BBC News"
               type="rss"
               xmlUrl="http://feeds.bbci.co.uk/news/rss.xml"
               htmlUrl="https://www.bbc.com/news" />
    </outline>

    <!-- Uncategorized Feeds -->
    <outline text="Example Blog"
             type="rss"
             xmlUrl="https://example.com/feed.xml"
             htmlUrl="https://example.com/" />
  </body>
</opml>
```

### C. Edge Case Test OPML

```xml
<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head>
    <title>Edge Case Tests</title>
  </head>
  <body>
    <!-- Feed with only 'title' attribute (no 'text') - non-compliant but common -->
    <outline title="Title Only Feed"
             type="rss"
             xmlUrl="https://example.com/feed1.xml" />

    <!-- Feed with special characters requiring escaping -->
    <outline text="Q&amp;A Blog &amp; More &quot;Quotes&quot;"
             type="rss"
             xmlUrl="https://example.com/feed?a=1&amp;b=2&amp;c=3" />

    <!-- Feed with no type attribute -->
    <outline text="No Type Feed"
             xmlUrl="https://example.com/feed3.xml" />

    <!-- Empty category -->
    <outline text="Empty Category" />

    <!-- Category with one feed -->
    <outline text="Single Feed Category">
      <outline text="Lonely Feed"
               type="rss"
               xmlUrl="https://example.com/feed4.xml" />
    </outline>

    <!-- Deeply nested structure (5 levels) -->
    <outline text="Level 1">
      <outline text="Level 2">
        <outline text="Level 3">
          <outline text="Level 4">
            <outline text="Level 5">
              <outline text="Deep Feed"
                       type="rss"
                       xmlUrl="https://example.com/deep.xml" />
            </outline>
          </outline>
        </outline>
      </outline>
    </outline>
  </body>
</opml>
```

---

**End of OPML Research Report**

Generated: 2024-10-12
For: Rogue Planet Feed Aggregator
Version: 1.0
