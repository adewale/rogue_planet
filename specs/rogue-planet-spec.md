# Rogue Planet (rp) - Specification

## Overview

Rogue Planet is a modern feed aggregator written in Go, inspired by Venus/Planet. It downloads RSS and Atom feeds from multiple sources and aggregates them into a single chronological stream published as a static HTML page.

## Core Principles

- **Simplicity**: Single binary, minimal configuration
- **Modern Go**: Use contemporary Go patterns and standard library where possible
- **Data Normalization**: All feeds normalized to a consistent format before storage
- **Static Output**: Generate simple, fast-loading HTML pages

## Architecture

```
┌─────────────┐
│   Config    │
│  (feeds.txt)│
└──────┬──────┘
       │
       ▼
┌─────────────┐      ┌──────────────┐      ┌────────────┐
│  Crawler    │─────▶│  Normaliser  │─────▶│ Repository │
│             │      │              │      │  (SQLite)  │
└─────────────┘      └──────────────┘      └─────┬──────┘
                                                  │
                                                  ▼
                                           ┌──────────────┐
                                           │     Site     │
                                           │  Generator   │
                                           └──────┬───────┘
                                                  │
                                                  ▼
                                           ┌──────────────┐
                                           │  index.html  │
                                           └──────────────┘
```

## Component Specifications

### 1. Crawler

**Purpose**: Fetch feed content from configured URLs **efficiently** to minimize server load

**Responsibilities**:
- Read list of feed URLs from configuration
- Fetch feed content via HTTP/HTTPS with proper caching headers
- Handle common HTTP scenarios (redirects, timeouts, caching)
- **CRITICAL: Implement HTTP conditional requests properly**
- Pass raw feed data to normalizer
- Support concurrent fetching with configurable worker pool
- Respect server resources and implement good netizen behavior

**Input**: 
- List of feed URLs (from config file or database)
- Previously stored ETag and Last-Modified headers from database

**Output**: 
- Raw feed data (XML/JSON bytes) - only when content has changed
- Metadata (fetch time, HTTP status, response headers)
- Updated ETag and Last-Modified headers for storage

**HTTP Conditional Request Implementation (CRITICAL)**:

Rogue Planet MUST implement proper HTTP conditional requests to be a good netizen. This is **non-negotiable** for a feed aggregator designed to run on cron jobs.

**The Problem** (from real-world feed reader failures):
1. Many feed readers fetch the entire feed on every request, wasting bandwidth
2. Some readers send fake/made-up conditional headers like "Wed, 01 Jan 1800 00:00:00 GMT"
3. Some readers parse body content to detect changes instead of using proper HTTP headers
4. These behaviors lead to being rate-limited or banned (429 Too Many Requests)

**The Solution** (proper implementation):

```go
// Pseudocode for conditional requests
type FeedCache struct {
    URL          string
    ETag         string    // Store EXACTLY as received, including quotes
    LastModified string    // Store EXACTLY as received  
    LastFetched  time.Time
}

func FetchFeed(url string, cache FeedCache) (*FeedResponse, error) {
    req, _ := http.NewRequest("GET", url, nil)
    
    // Set User-Agent with contact info
    req.Header.Set("User-Agent", "RoguePlanet/1.0 (+https://example.com/about)")
    
    // CRITICAL: Send conditional headers if we have them
    if cache.LastModified != "" {
        // Use If-Modified-Since with the EXACT value from Last-Modified
        req.Header.Set("If-Modified-Since", cache.LastModified)
    }
    
    if cache.ETag != "" {
        // Use If-None-Match with the EXACT value from ETag (including quotes!)
        req.Header.Set("If-None-Match", cache.ETag)
    }
    
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // Handle 304 Not Modified
    if resp.StatusCode == http.StatusNotModified {
        return &FeedResponse{NotModified: true}, nil
    }
    
    // CRITICAL: Always update cache with NEW headers from response
    newCache := FeedCache{
        URL:          url,
        ETag:         resp.Header.Get("ETag"),         // May include quotes
        LastModified: resp.Header.Get("Last-Modified"), // Exact value
        LastFetched:  time.Now(),
    }
    
    // Read body only if status is 200
    if resp.StatusCode == http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return &FeedResponse{
            Body:     body,
            NewCache: newCache,
        }, nil
    }
    
    return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
}
```

**Critical Rules for Conditional Requests**:
1. **ALWAYS store ETag and Last-Modified EXACTLY as received** - don't modify, hash, or "improve" them
2. **ETag values often include double quotes** - these quotes are PART of the value per RFC 7232
3. **Send BOTH If-Modified-Since AND If-None-Match if you have both** - let the server decide
4. **NEVER make up values** - if you don't have a previous value, don't send the header
5. **Update stored values on EVERY response** - even if content seems unchanged
6. **Don't hash body content** - use server-provided headers, not your own heuristics
7. **The previous response's Last-Modified becomes the next request's If-Modified-Since**
8. **The previous response's ETag becomes the next request's If-None-Match**

**Key Features**:
- Proper User-Agent with version and contact URL
- Timeout handling (default: 30s, configurable)
- Respect robots.txt (optional but recommended)
- Rate limiting per domain (avoid hammering servers)
- Retry logic with exponential backoff (with maximum retries)
- **Mandatory HTTP conditional request support**
- Handle 304 Not Modified responses correctly
- Handle 429 Too Many Requests with backoff
- Handle 301/302 redirects and update feed URL in database
- gzip/deflate compression support via Accept-Encoding
- Connection pooling with appropriate timeouts

**Fetch Scheduling**:
- Default fetch interval: 1 hour (configurable per feed)
- Respect Cache-Control: max-age header if present
- Implement jitter to avoid thundering herd (don't fetch all feeds at exact same time)
- Back off on repeated errors (exponential backoff for failing feeds)

**Good Netizen Behaviors**:
- Identify clearly in User-Agent: "RoguePlanet/1.0 (+https://yoursite.com/bot-info)"
- Include contact information in User-Agent or separate header
- Honor 429 rate limit responses with Retry-After header
- Don't fetch more frequently than necessary
- Implement maximum fetch frequency (e.g., never faster than every 15 minutes)
- Stop fetching feeds that consistently fail (after N attempts)

**Go Libraries**:
- `net/http` for fetching (stdlib)
- `context` for timeout/cancellation (stdlib)
- `golang.org/x/time/rate` for rate limiting
- Consider: connection pool tuning for efficiency

### 2. Normaliser

**Purpose**: Transform feeds into a consistent, clean format following Venus normalization principles

**Responsibilities**:
- Parse RSS 1.0, RSS 2.0, Atom 1.0, and JSON Feed formats
- Convert all feeds to a canonical internal format (Atom-style)
- Sanitize HTML content (remove dangerous tags/attributes)
- Resolve relative URLs to absolute
- Fix common encoding issues
- Extract and normalize dates to RFC 3339
- Generate IDs for entries that lack them
- Handle malformed HTML with proper parsing

**Normalization Rules** (following Venus):

1. **Character Encoding**:
   - Convert all text to UTF-8
   - Handle common charset detection issues
   - Replace invalid characters with Unicode replacement character (U+FFFD)
   - Convert HTML entities to Unicode

2. **HTML Sanitization**:
   - Remove dangerous tags: `<script>`, `<object>`, `<embed>`, `<iframe>`
   - Remove event handlers: `onclick`, `onerror`, etc.
   - Allow safe tags: `<p>`, `<a>`, `<img>`, `<strong>`, `<em>`, `<ul>`, `<ol>`, `<li>`, `<blockquote>`, `<pre>`, `<code>`, etc.
   - Allow safe attributes: `href`, `src`, `alt`, `title`, `class`
   - Close unmatched tags properly
   - Support safe subset of MathML (optional)

3. **Link Resolution**:
   - Convert all relative URLs to absolute using xml:base or feed URL
   - Apply to: links, images, embedded content

4. **Date Normalization**:
   - Parse various date formats (RFC 822, RFC 3339, ISO 8601, etc.)
   - Convert to RFC 3339 format
   - Handle timezone conversions
   - Use feed date if entry date missing
   - Use fetch time if no dates present
   - Handle future dates (configurable: ignore or accept)

5. **ID Generation**:
   - Use existing GUID/ID if present
   - Otherwise synthesize from: permalink, title+date hash, or content hash
   - Ensure IDs are stable across fetches

6. **Content Normalization**:
   - Prefer full content over summary
   - Extract author information
   - Normalize multiple author formats
   - Handle feed-level vs entry-level metadata

**Input**: 
- Raw feed data (bytes)
- Feed source URL
- Feed configuration overrides

**Output**: 
- Normalized entry objects ready for storage

**Configuration Overrides**:
```
ignore_in_feed: [list of elements/attributes to ignore]
title_type: override content type
summary_type: override content type  
content_type: override content type
future_dates: "ignore" or "ignore_entry" or "accept"
xml_base: override xml:base value
```

**Go Libraries**:
- Feed parsing: `github.com/mmcdole/gofeed` or custom parser
- HTML sanitization: `golang.org/x/net/html` or `github.com/microcosm-cc/bluemonday`
- Character encoding: `golang.org/x/text/encoding`

### 3. Repository

**Purpose**: Persist normalized entries in SQLite database

**Database Schema**:

```sql
-- Feeds table
CREATE TABLE feeds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT NOT NULL UNIQUE,
    title TEXT,
    link TEXT,
    updated TEXT,  -- RFC 3339 timestamp from feed
    last_fetched TEXT,  -- RFC 3339 timestamp of last fetch attempt
    etag TEXT,  -- ETag from last successful fetch (stored exactly as received)
    last_modified TEXT,  -- Last-Modified from last successful fetch (stored exactly)
    fetch_error TEXT,  -- Last error message if fetch failed
    fetch_error_count INTEGER DEFAULT 0,  -- Consecutive error count for backoff
    next_fetch TEXT,  -- RFC 3339 timestamp of next scheduled fetch
    active INTEGER DEFAULT 1,  -- 0 = disabled, 1 = active
    fetch_interval INTEGER DEFAULT 3600  -- Seconds between fetches (default 1 hour)
);

-- Entries table
CREATE TABLE entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_id INTEGER NOT NULL,
    entry_id TEXT NOT NULL,  -- Original GUID/ID from feed (must be unique per feed)
    title TEXT,
    link TEXT,
    author TEXT,
    published TEXT,  -- RFC 3339 timestamp
    updated TEXT,    -- RFC 3339 timestamp
    content TEXT,    -- Normalized and sanitized HTML
    content_type TEXT DEFAULT 'html',
    summary TEXT,    -- Sanitized summary text
    first_seen TEXT,  -- RFC 3339 timestamp when first crawled
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
    UNIQUE(feed_id, entry_id)
);

-- Indices
CREATE INDEX idx_entries_published ON entries(published DESC);
CREATE INDEX idx_entries_updated ON entries(updated DESC);
CREATE INDEX idx_entries_feed_id ON entries(feed_id);
CREATE INDEX idx_entries_first_seen ON entries(first_seen DESC);
CREATE INDEX idx_feeds_active ON feeds(active);
CREATE INDEX idx_feeds_next_fetch ON feeds(next_fetch);
```

**Schema Notes**:
- `etag` and `last_modified` in feeds table enable HTTP conditional requests
- `fetch_error_count` enables exponential backoff for failing feeds
- `next_fetch` allows intelligent scheduling with jitter
- `fetch_interval` allows per-feed customization
- `ON DELETE CASCADE` ensures cleanup when feeds are removed
- `UNIQUE(feed_id, entry_id)` prevents duplicate entries
- Multiple indices on dates optimize common queries

**Operations**:
- `AddFeed(url, title) error`
- `UpdateFeed(id, metadata) error`
- `GetFeeds(activeOnly bool) ([]Feed, error)`
- `UpsertEntry(entry Entry) error` - Insert or update if exists
- `GetRecentEntries(days int) ([]Entry, error)`
- `GetEntriesSince(timestamp time.Time) ([]Entry, error)`
- `PruneOldEntries(days int) error` - Optional cleanup

**Key Features**:
- Connection pooling
- Prepared statements
- Transaction support for batch inserts
- Handle UNIQUE constraint violations gracefully

**Go Libraries**:
- `database/sql`
- `github.com/mattn/go-sqlite3` (CGO required) or `modernc.org/sqlite` (pure Go)

### 4. Site Generator

**Purpose**: Generate static HTML page from recent entries

**Responsibilities**:
- Query repository for entries from last N days
- Sort entries by date (newest first)
- Apply HTML template
- Write output to file
- Generate feed metadata (feed list, dates, etc.)

**Template Variables**:
```
{{.Title}}           - Site title
{{.Subtitle}}        - Site subtitle  
{{.Link}}           - Site URL
{{.Updated}}        - Last updated timestamp
{{.Generator}}      - "Rogue Planet vX.X.X"
{{.OwnerName}}      - Site owner
{{.OwnerEmail}}     - Owner email
{{.Entries}}        - Array of entries
  {{.Title}}        - Entry title
  {{.Link}}         - Entry permalink
  {{.Author}}       - Entry author
  {{.FeedTitle}}    - Source feed title
  {{.FeedLink}}     - Source feed URL
  {{.Published}}    - Published date
  {{.Updated}}      - Updated date
  {{.Content}}      - Sanitized HTML content (safe to output)
  {{.Summary}}      - Summary if available
```

**Output Format**:
- Single HTML file (`index.html`)
- Embedded CSS (optional external stylesheet)
- Responsive design
- Semantic HTML5
- Valid HTML5 output

**Configuration**:
```
[site]
title = "My Planet"
link = "https://planet.example.com"
owner_name = "John Doe"
owner_email = "john@example.com"
days = 7
output_file = "public/index.html"
template_file = "templates/default.html"
```

**Template Engine**:
- Use Go's `html/template` (automatic HTML escaping)
- Support for custom template files
- Provide default template

**Default Template Features**:
- Clean, readable design
- Group entries by date
- Show feed source for each entry
- Responsive layout
- Minimal JavaScript (or none)

## Configuration File Format

**feeds.txt** (simple line-by-line format):
```
https://blog.golang.org/feed.atom
https://example.com/rss
# Comments supported
https://another-blog.com/feed
```

**config.ini** (extended configuration):
```ini
[planet]
name = My Planet
link = https://planet.example.com
owner_name = John Doe
owner_email = john@example.com
cache_directory = ./cache
output_dir = ./public
days = 7
log_level = info
concurrent_fetches = 5
user_agent = RoguePlanet/1.0

[database]
path = ./data/planet.db

[feed "https://blog.golang.org/feed.atom"]
name = Go Blog
ignore_in_feed = author

[feed "https://example.com/rss"]
future_dates = ignore
```

## CLI Interface

```bash
# Fetch all feeds and regenerate site
rp update

# Add a new feed
rp add-feed https://example.com/feed.xml

# Remove a feed
rp remove-feed https://example.com/feed.xml

# List all feeds
rp list-feeds

# Generate site without fetching
rp generate

# Fetch feeds without generating
rp fetch

# Initialize new planet
rp init [directory]

# Show version
rp version

# Prune old entries
rp prune --days 90
```

## Command Flags

```
Global Flags:
  --config, -c     Path to config file (default: ./config.ini)
  --verbose, -v    Verbose output
  --quiet, -q      Quiet mode (errors only)

Update Command:
  --force, -f      Force fetch even if cached
  --feeds          Comma-separated list of specific feeds to update

Generate Command:
  --days, -d       Number of days to include (default: from config)
  --template, -t   Path to template file

Prune Command:
  --days           Keep entries newer than N days
  --dry-run        Show what would be deleted
```

## Data Flow

1. **Fetch Phase**:
   ```
   Read config → Load feeds from DB → Fetch in parallel → 
   Pass to normalizer → Store in DB
   ```

2. **Generate Phase**:
   ```
   Query DB for recent entries → Sort by date → 
   Apply template → Write HTML file
   ```

## Error Handling

- Log all fetch errors but continue processing other feeds
- Store fetch errors in database with timestamp
- Report errors in generated HTML (optional section)
- Graceful degradation: show what data is available
- Retry failed feeds on next update

## Performance Considerations

- Concurrent feed fetching (configurable worker pool)
- Connection pooling for database
- Prepared statements for queries
- Conditional HTTP requests (ETag, Last-Modified)
- Only regenerate site if new entries exist (optional)
- Index database appropriately

## Future Enhancements

- Multiple output formats (Atom feed, RSS feed, JSON Feed)
- Multiple page templates
- Plugin system for custom filters
- Web UI for configuration
- WebSub (PubSubHubbub) support
- OPML import/export
- Feed discovery
- Archive pages (by month/year)
- Tag/category support
- Full-text search

## Dependencies

**Required**:
- Go 1.21+
- SQLite 3

**Go Modules**:
- Feed parsing library
- HTML sanitization library  
- SQLite driver
- Configuration parser (optional)

## Testing Strategy

- Unit tests for each component
- Integration tests for full pipeline
- Test with various feed formats
- Test with malformed feeds
- Benchmark concurrent fetch performance
- Test HTML sanitization with XSS attempts

## Security Considerations

**CRITICAL: CVE-2009-2937 and XSS Prevention**

Planet Venus and Planet 2.0 suffered from CVE-2009-2937, a cross-site scripting vulnerability that allowed remote attackers to inject arbitrary web script or HTML via malicious content in feeds. This vulnerability arose from insufficient sanitization of feed content, particularly in IMG src attributes.

**The Attack Vector**:
```html
<!-- Malicious feed content that executed JavaScript -->
<img src="javascript:alert(1);" >
```

**Rogue Planet's Defense Strategy**:

### 1. **Comprehensive HTML Sanitization** (MANDATORY)

All HTML content from feeds MUST be sanitized before storage and display. This is the #1 security concern.

**Use a battle-tested sanitization library**:
- `github.com/microcosm-cc/bluemonday` (recommended - widely used, actively maintained)
- Configure with strict policies

**Sanitization Rules**:
```go
// Example using bluemonday
import "github.com/microcosm-cc/bluemonday"

// Create a strict policy
policy := bluemonday.StrictPolicy()
// Or use UGCPolicy (User Generated Content) for basic formatting
policy := bluemonday.UGCPolicy()

// Additional rules:
// 1. Strip ALL JavaScript
policy.AllowAttrs("src").OnlyIfNotPresent("javascript:").OnElements("img")
policy.AllowAttrs("href").OnlyIfNotPresent("javascript:").OnElements("a")

// 2. Strip event handlers
// Already handled by bluemonday, but document it
// onclick, onerror, onload, onmouseover, etc. - ALL REMOVED

// 3. Remove dangerous elements
// <script>, <object>, <embed>, <iframe>, <frame> - ALL REMOVED
// <base> tag - REMOVED (can redirect all relative links)

// 4. Remove data: URIs in src/href (can contain embedded JS)
policy.AllowURLSchemes("http", "https") // Only allow http/https

// 5. Sanitize CSS (if allowing style attributes)
// bluemonday does this, but be aware of expression() and other CSS attacks
```

**Safe Tags to Allow** (after sanitization):
- Text formatting: `<p>`, `<br>`, `<strong>`, `<em>`, `<b>`, `<i>`, `<u>`, `<strike>`
- Lists: `<ul>`, `<ol>`, `<li>`
- Headings: `<h1>` through `<h6>`
- Quotes: `<blockquote>`, `<q>`, `<cite>`
- Code: `<pre>`, `<code>`, `<tt>`, `<kbd>`, `<samp>`, `<var>`
- Tables: `<table>`, `<thead>`, `<tbody>`, `<tr>`, `<th>`, `<td>`, `<caption>`
- Links: `<a>` (with href restricted to http/https)
- Images: `<img>` (with src restricted to http/https)
- Divisions: `<div>`, `<span>`

**Safe Attributes to Allow** (after validation):
- `href` (on `<a>` only, URL-validated)
- `src` (on `<img>` only, URL-validated)
- `alt`, `title` (descriptive text)
- `class` (for styling - but restrict to known-safe classes if possible)
- NO `style` attributes (or sanitize very carefully)
- NO event handlers (onclick, etc.)
- NO `id` attributes (can cause conflicts)

### 2. **Defense in Depth**

**Layer 1: Sanitize on Input** (when normalizing feed)
- Clean HTML immediately after parsing feed
- Store sanitized content in database

**Layer 2: Sanitize on Output** (when generating HTML)
- Use Go's `html/template` which auto-escapes by default
- Mark pre-sanitized content as safe using `template.HTML()` ONLY after sanitization

**Layer 3: Content Security Policy Headers** (in generated HTML)
```html
<meta http-equiv="Content-Security-Policy" 
      content="default-src 'self'; 
               script-src 'self'; 
               style-src 'self' 'unsafe-inline'; 
               img-src 'self' https:; 
               object-src 'none'; 
               base-uri 'self';">
```

### 3. **URL Validation**

**Prevent Server-Side Request Forgery (SSRF)**:
```go
func ValidateFeedURL(url string) error {
    parsed, err := url.Parse(url)
    if err != nil {
        return err
    }
    
    // Only allow http and https
    if parsed.Scheme != "http" && parsed.Scheme != "https" {
        return errors.New("only http/https allowed")
    }
    
    // Prevent localhost/internal network access
    host := parsed.Hostname()
    ip := net.ParseIP(host)
    if ip != nil {
        // Block localhost
        if ip.IsLoopback() {
            return errors.New("localhost not allowed")
        }
        // Block private networks (RFC 1918)
        if ip.IsPrivate() {
            return errors.New("private IPs not allowed")
        }
        // Block link-local
        if ip.IsLinkLocalUnicast() {
            return errors.New("link-local not allowed")
        }
    }
    
    // Block known internal hostnames
    internalHosts := []string{"localhost", "127.0.0.1", "::1"}
    for _, blocked := range internalHosts {
        if strings.EqualFold(host, blocked) {
            return errors.New("internal host not allowed")
        }
    }
    
    return nil
}
```

### 4. **SQL Injection Prevention**

- **ALWAYS use prepared statements / parameterized queries**
- Never concatenate user input into SQL
- Use `database/sql` with placeholder syntax (`?` or `$1`, etc.)

```go
// GOOD - parameterized query
db.Exec("INSERT INTO feeds (url, title) VALUES (?, ?)", url, title)

// BAD - NEVER DO THIS
db.Exec(fmt.Sprintf("INSERT INTO feeds (url, title) VALUES ('%s', '%s')", url, title))
```

### 5. **Resource Limits**

**Prevent Denial of Service**:
```go
// Limit response body size (e.g., 10MB max)
const MaxFeedSize = 10 * 1024 * 1024
resp.Body = http.MaxBytesReader(nil, resp.Body, MaxFeedSize)

// Timeout all requests
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Limit concurrent fetches
semaphore := make(chan struct{}, maxConcurrent)
```

### 6. **Input Validation**

- Validate all feed URLs before fetching
- Sanitize feed titles and descriptions before storage
- Limit string lengths (titles, authors, etc.) to reasonable values
- Validate date formats and handle malformed dates gracefully

### 7. **Template Safety**

- Use `html/template` (not `text/template`) for HTML generation
- Auto-escaping is enabled by default
- Only mark content as safe (`template.HTML()`) AFTER sanitization
- Never pass unsanitized feed content directly to templates

### 8. **Error Handling**

- Don't expose internal paths or stack traces in error messages
- Log detailed errors server-side
- Return generic errors to users
- Handle panics gracefully

### 9. **Dependency Management**

- Keep all dependencies up to date
- Use `go mod` for dependency management
- Regularly check for security updates
- Pin dependency versions in production

### 10. **File System Safety**

- Never allow user input to determine file paths
- Use absolute paths for output directories
- Validate output paths don't escape intended directory
- Set appropriate file permissions (e.g., 0644 for output files)

**Security Checklist** (before release):
- [ ] HTML sanitization implemented and tested
- [ ] XSS attack testing performed
- [ ] SQL injection testing performed  
- [ ] SSRF protection implemented
- [ ] URL validation implemented
- [ ] Resource limits configured
- [ ] CSP headers added to output
- [ ] Input validation comprehensive
- [ ] Error messages don't leak information
- [ ] Dependencies up to date
- [ ] Security review completed

## Lessons Learned from Planet History

Based on 20+ years of feed aggregator development (Planet, Venus, Mars, Pluto, Moonmoon), here are critical lessons:

### Architecture Lessons

**1. Static Output is Superior**
- **Lesson**: Venus/Planet generate static HTML files, not dynamic pages
- **Why**: Survives traffic spikes (HN, Reddit), no database queries on reads, can be served by any web server
- **Implementation**: Generate complete HTML files that can be served by nginx/Apache/CDN

**2. Separate Concerns Cleanly**
- **Lesson**: Keep fetch, normalize, store, and generate as distinct phases
- **Why**: Easier to debug, test, and maintain; can run phases independently
- **Implementation**: Four separate components that communicate through the database

**3. Caching is Critical**
- **Lesson**: Cache everything - feeds, parsed content, generated HTML
- **Why**: Reduces server load, enables offline development, faster regeneration
- **Implementation**: SQLite as both database and cache, etag/last-modified tracking

**4. Simple is Better**
- **Lesson**: Moonmoon's "stupidly simple" approach vs Venus's complexity
- **Why**: Easier to maintain, fewer bugs, more contributors, faster to deploy
- **Implementation**: No comments, no voting, no user accounts - just aggregation

### Feed Handling Lessons

**5. Proper HTTP Conditional Requests are Mandatory**
- **Lesson**: Many aggregators got this wrong and were banned/rate-limited
- **Why**: Feed publishers will block aggregators that waste bandwidth
- **Implementation**: Store and send ETag/Last-Modified exactly as received (see Crawler section)
- **Evidence**: rachelbythebay.com documented extensive problems with poorly-behaved readers

**6. Feed URLs Change and Redirect**
- **Lesson**: Feeds move, domains change, HTTP → HTTPS redirects happen
- **Why**: Permanent failures vs temporary moves need different handling
- **Implementation**: 
  - Follow 301 (permanent) redirects and update feed URL in database
  - Follow 302/307 (temporary) redirects but keep original URL
  - Handle redirect loops (max 5 redirects)

**7. Feeds are Messy**
- **Lesson**: Real-world feeds violate specs, have encoding issues, broken HTML
- **Why**: Feed publishers use varied tools and don't always validate
- **Implementation**: Robust parsing (Venus's normalization), graceful degradation, error logging

**8. Some Feeds are Hostile (Intentionally or Not)**
- **Lesson**: Malformed XML, huge files, incorrect content-types, missing dates
- **Why**: Bugs, malice, or ignorance on publisher side
- **Implementation**: Timeouts, size limits, strict validation, error isolation

### Normalization Lessons

**9. Normalize Everything to a Canonical Format**
- **Lesson**: Venus normalizes RSS 1.0/2.0, Atom to a single internal format
- **Why**: Template generation is simpler, only one output code path
- **Implementation**: Parse all formats → canonical internal format → store

**10. HTML Must Be Sanitized (CVE-2009-2937)**
- **Lesson**: XSS via malicious feed content is a real threat
- **Why**: Attackers can inject JavaScript through feeds
- **Implementation**: Strict HTML sanitization with whitelist approach (see Security section)

**11. Character Encoding is Hard**
- **Lesson**: Feeds claim one encoding but use another, or use Windows-1252 claiming UTF-8
- **Why**: Publishers make mistakes, copy from broken sources
- **Implementation**: 
  - Try declared encoding first
  - Fall back to detection heuristics
  - Replace invalid characters with Unicode replacement character (U+FFFD)
  - Always output UTF-8

**12. Dates are a Mess**
- **Lesson**: Many date formats, timezones, missing dates, future dates
- **Why**: Specs allow flexibility, publishers make errors
- **Implementation**:
  - Parse multiple date formats (RFC 822, RFC 3339, ISO 8601, etc.)
  - Use feed date if entry date missing
  - Use fetch time as last resort
  - Handle future dates with configuration (ignore vs accept)

### ID and Deduplication Lessons

**13. Entry IDs are Often Missing or Change**
- **Lesson**: Not all feeds provide stable GUIDs/IDs
- **Why**: Spec violations, CMS bugs, feed regeneration
- **Implementation**:
  - Use provided GUID/ID if present and stable
  - Generate from permalink if available
  - Fall back to content hash (but warn about duplicates on edits)
  - Never use publish date as ID (not unique)

**14. Detect and Handle Duplicates**
- **Lesson**: Same entry appears in multiple feeds or with different IDs
- **Why**: Syndication, feed URL changes, publisher errors
- **Implementation**: 
  - Unique constraint on (feed_id, entry_id)
  - Update existing entries rather than creating duplicates
  - Log when entries change significantly

### Performance Lessons

**15. Concurrent Fetching with Limits**
- **Lesson**: Fetch feeds in parallel, but not too many at once
- **Why**: Faster updates, but need to avoid overwhelming servers or self
- **Implementation**: Worker pool with 5-20 workers, rate limiting per domain

**16. Don't Fetch Too Frequently**
- **Lesson**: Even with conditional requests, polling every minute wastes resources
- **Why**: Most blogs update hourly or daily, not every minute
- **Implementation**: 
  - Default: 1 hour between fetches
  - Respect Cache-Control: max-age if present
  - Add jitter to avoid thundering herd
  - Back off on repeated failures

**17. Database Indexes Matter**
- **Lesson**: Queries by date, by feed, and sorting are common
- **Why**: Site generation queries recent entries frequently
- **Implementation**: Index on published date, updated date, feed_id

### User Experience Lessons

**18. Group by Date, Not by Feed**
- **Lesson**: "River of news" format (chronological) is more engaging
- **Why**: Users want to see latest content, not per-feed organization
- **Implementation**: Single chronological stream, newest first, optionally grouped by day

**19. Show Source Feed Information**
- **Lesson**: Users want to know which feed an entry came from
- **Why**: Context, credibility, ability to follow back to original
- **Implementation**: Display feed title/icon next to each entry

**20. Link to Original Post**
- **Lesson**: Always link to the original article on the source site
- **Why**: Attribution, full content, comments, respect for publisher
- **Implementation**: Prominent link on entry title and/or "read more" link

### Operational Lessons

**21. User-Agent Must Identify the Bot**
- **Lesson**: Anonymous user-agents get blocked
- **Why**: Publishers want to identify and contact bot operators
- **Implementation**: "RoguePlanet/1.0 (+https://yoursite.com/bot-info)" with contact page

**22. Provide Bot Information Page**
- **Lesson**: Publishers need a way to contact you or request removal
- **Why**: Professional courtesy, debugging, avoiding blocks
- **Implementation**: Web page explaining your bot, contact info, how to opt-out

**23. Log Errors but Don't Stop**
- **Lesson**: One failing feed shouldn't break the whole aggregation
- **Why**: Feeds fail temporarily; resilience is key
- **Implementation**: 
  - Try/catch around each feed fetch
  - Log errors with context
  - Continue processing other feeds
  - Store error state in database

**24. Clean Up Old Data**
- **Lesson**: Database grows forever if not pruned
- **Why**: Performance degrades, disk fills up
- **Implementation**: Prune entries older than N days (configurable), keep feed metadata

### Template and Output Lessons

**25. Provide Default Theme, Allow Customization**
- **Lesson**: Users want both ease-of-use and customization
- **Why**: Different communities have different aesthetics
- **Implementation**: Include default template, document template variables

**26. Responsive Design is Expected**
- **Lesson**: Mobile traffic is significant
- **Why**: Users read aggregated content on phones/tablets
- **Implementation**: Mobile-first CSS, readable on all screen sizes

**27. Keep JavaScript Minimal or Absent**
- **Lesson**: Static HTML works everywhere and is fast
- **Why**: Accessibility, performance, works with JS disabled
- **Implementation**: Pure HTML/CSS generation, JS only for enhancements

### Configuration Lessons

**28. Configuration Should be Simple**
- **Lesson**: Venus's INI format is easier than XML or complex config
- **Why**: Lower barrier to entry, easier to edit by hand
- **Implementation**: Simple INI or TOML format with sensible defaults

**29. Per-Feed Overrides are Valuable**
- **Lesson**: Some feeds need special handling (encoding, date handling, etc.)
- **Why**: Not all feeds are well-behaved; customization avoids forking
- **Implementation**: Feed-specific config sections to override global settings

### Modern Improvements Over Venus

**30. Use Modern Feed Parsing Libraries**
- **Lesson**: gofeed (Go) is more maintained than Universal Feed Parser (Python)
- **Why**: Better error handling, active maintenance, more formats
- **Implementation**: Use `github.com/mmcdole/gofeed` in Go

**31. SQLite is Perfect for This Use Case**
- **Lesson**: No need for PostgreSQL/MySQL for a feed aggregator
- **Why**: Simple, portable, fast enough, single-file database
- **Implementation**: Use SQLite with proper indexes and WAL mode

**32. Single Binary Deployment**
- **Lesson**: Go's static compilation makes deployment trivial
- **Why**: No Python dependencies, virtualenvs, or system packages
- **Implementation**: Compile to single binary, ship anywhere

### Project Sustainability Lessons

**33. Document Everything**
- **Lesson**: Many Planet instances died due to lack of documentation
- **Why**: New maintainers need clear instructions
- **Implementation**: README with setup, usage, troubleshooting

**34. Make it Easy to Contribute**
- **Lesson**: Simple codebases attract contributors
- **Why**: More contributors = more sustainability
- **Implementation**: Clean code, good comments, clear architecture

**35. Test with Real Feeds**
- **Lesson**: Unit tests aren't enough; real feeds are messy
- **Why**: Edge cases appear in production
- **Implementation**: Integration tests with real feed URLs, snapshot testing

### Summary of Key Principles

1. **Be a good netizen**: Proper caching, identification, respect for servers
2. **Security first**: Sanitize everything, assume feeds are hostile
3. **Simple is sustainable**: Fewer features = fewer bugs = longer life
4. **Static output**: Fast, scalable, survivable
5. **Graceful degradation**: One bad feed doesn't break everything
6. **Data normalization**: Canonical format makes everything simpler
7. **Clear separation**: Fetch → Normalize → Store → Generate
8. **Default to working**: Sensible defaults, minimal configuration required