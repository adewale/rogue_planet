# Rogue Planet - Comprehensive Test Plan

## Overview

This test plan ensures Rogue Planet maintains high quality through comprehensive unit testing (with branch coverage) and diverse integration testing scenarios. Target: >75% code coverage across all packages, with >90% branch coverage for critical security paths.

## Test Categories

1. **Unit Tests**: Test individual functions and methods in isolation
2. **Integration Tests**: Test component interactions and full workflows
3. **Security Tests**: Test XSS prevention, SSRF prevention, SQL injection prevention
4. **Edge Case Tests**: Test boundary conditions and error handling
5. **Performance Tests**: Test concurrent operations and resource limits
6. **Real-World Tests**: Test with actual feed data and network conditions

---

## 1. Unit Tests by Package

### 1.1 pkg/crawler (HTTP Fetching)

#### Test Coverage Goals
- **Target**: >85% line coverage, >90% branch coverage
- **Critical Paths**: SSRF validation, conditional requests, error handling

#### Test Cases

**ValidateURL() - SSRF Prevention (Critical)**
- ✅ Valid HTTP URL: `http://example.com/feed.xml`
- ✅ Valid HTTPS URL: `https://example.com/feed.xml`
- ❌ Invalid scheme - FTP: `ftp://example.com/feed.xml` → ErrInvalidScheme
- ❌ Invalid scheme - File: `file:///etc/passwd` → ErrInvalidScheme
- ❌ Invalid scheme - Data URI: `data:text/html,<script>alert(1)</script>` → ErrInvalidScheme
- ❌ Localhost by name: `http://localhost/feed.xml` → ErrPrivateIP
- ❌ Localhost by IPv4: `http://127.0.0.1/feed.xml` → ErrPrivateIP
- ❌ Localhost by IPv6: `http://[::1]/feed.xml` → ErrPrivateIP
- ❌ All zeros: `http://0.0.0.0/feed.xml` → ErrPrivateIP
- ❌ Private IPv4 (10.x): `http://10.0.0.1/feed.xml` → ErrPrivateIP
- ❌ Private IPv4 (192.168.x): `http://192.168.1.1/feed.xml` → ErrPrivateIP
- ❌ Private IPv4 (172.16-31.x): `http://172.16.0.1/feed.xml` → ErrPrivateIP
- ❌ Link-local IPv4: `http://169.254.0.1/feed.xml` → ErrPrivateIP
- ❌ Link-local IPv6: `http://[fe80::1]/feed.xml` → ErrPrivateIP
- ❌ Malformed URL: `http://[invalid` → ErrInvalidURL
- ✅ Valid domain with port: `http://example.com:8080/feed.xml`
- ✅ Valid IPv4 (public): `http://93.184.216.34/feed.xml` (example.com)

**Fetch() - HTTP Conditional Requests (Critical)**
- ✅ First fetch (no cache) → Returns full body
- ✅ Second fetch with ETag (304 response) → Returns NotModified=true
- ✅ Second fetch with Last-Modified (304 response) → Returns NotModified=true
- ✅ Second fetch with both headers (304 response) → Returns NotModified=true
- ✅ ETag with quotes: `"abc123"` → Sent as-is in If-None-Match
- ✅ ETag without quotes: `abc123` → Sent as-is in If-None-Match
- ✅ Weak ETag: `W/"abc123"` → Sent as-is in If-None-Match
- ✅ Last-Modified RFC 1123 format → Sent as-is in If-Modified-Since
- ✅ Cache updated on 200 response → NewCache contains new headers
- ✅ Cache preserved on 304 response → NewCache contains old headers
- ❌ SSRF attempt (skipSSRFCheck=false) → ErrPrivateIP
- ✅ SSRF allowed in test mode (skipSSRFCheck=true) → Success

**Fetch() - Response Handling**
- ✅ 200 OK with body → Returns body and new cache headers
- ✅ 304 Not Modified → Returns NotModified=true
- ❌ 404 Not Found → Error with status code
- ❌ 429 Too Many Requests → Error with status code
- ❌ 500 Server Error → Error with status code
- ✅ Gzip-encoded response → Automatically decompressed
- ✅ Deflate-encoded response → Automatically decompressed
- ✅ No encoding → Body returned as-is
- ❌ Invalid gzip data → Error
- ✅ Response with User-Agent set → Correct User-Agent header sent
- ✅ Custom User-Agent (NewWithUserAgent) → Custom UA sent

**Fetch() - Size and Timeout Limits**
- ✅ Response under 10MB → Success
- ❌ Response over 10MB → ErrMaxSizeExceeded
- ✅ Response exactly 10MB → Success
- ❌ Request timeout (context deadline) → Context deadline exceeded
- ❌ Request cancelled (context cancel) → Context cancelled
- ✅ Request within timeout → Success

**Fetch() - Redirects**
- ✅ 301 Permanent Redirect → Follows and returns final URL
- ✅ 302 Temporary Redirect → Follows and returns final URL
- ✅ 307 Temporary Redirect → Follows and returns final URL
- ✅ 3 redirects → Success, returns final URL
- ❌ 6 redirects (exceeds MaxRedirects=5) → Error "stopped after 5 redirects"
- ✅ Redirect loop detection → Error after MaxRedirects

**FetchWithRetry() - Retry Logic**
- ✅ Success on first try → No retries, returns immediately
- ✅ Transient error, success on second try → 1 retry with backoff
- ✅ Transient error, success on third try → 2 retries with exponential backoff
- ❌ Max retries exceeded → Error "max retries exceeded"
- ❌ Non-retryable error (ErrInvalidURL) → Returns immediately, no retry
- ❌ Non-retryable error (ErrPrivateIP) → Returns immediately, no retry
- ❌ Non-retryable error (ErrMaxSizeExceeded) → Returns immediately, no retry
- ❌ 400 Bad Request → No retry (client error)
- ❌ 403 Forbidden → No retry (client error)
- ✅ 429 Too Many Requests → Retries (special case)
- ❌ 500 Server Error → Retries up to max
- ✅ Context cancelled during retry → Returns immediately

**Constructor Variants**
- ✅ New() → Default settings (30s timeout, SSRF checks enabled)
- ✅ NewWithUserAgent(customUA) → Custom UA, other defaults
- ✅ NewWithUserAgent("") → Falls back to default UA
- ✅ NewForTesting() → SSRF checks disabled

**Branch Coverage Focus**
- Error path when gzip reader creation fails
- Error path when response body read fails
- Both ETag and Last-Modified present (test both branches)
- Only ETag present
- Only Last-Modified present
- Neither ETag nor Last-Modified present
- Status code branches (200, 304, 4xx, 5xx)

---

### 1.2 pkg/normalizer (Feed Parsing and Sanitization)

#### Test Coverage Goals
- **Target**: >80% line coverage, >90% branch coverage for sanitization
- **Critical Paths**: HTML sanitization, ID generation, XSS prevention

#### Test Cases

**Parse() - Feed Format Support**
- ✅ Valid RSS 2.0 feed → Parsed correctly
- ✅ Valid RSS 1.0 (RDF) feed → Parsed correctly
- ✅ Valid Atom 1.0 feed → Parsed correctly
- ✅ Valid JSON Feed → Parsed correctly
- ❌ Invalid XML → ErrInvalidFeed
- ❌ Empty feed data → ErrInvalidFeed
- ❌ Malformed XML (unclosed tags) → ErrInvalidFeed
- ✅ Feed with no entries → Success, returns empty entries slice
- ✅ Feed with 1 entry → Success, returns 1 entry
- ✅ Feed with 100 entries → Success, returns 100 entries
- ✅ Feed with UTF-8 content → Correctly parsed
- ✅ Feed with HTML entities → Entities decoded

**Parse() - Feed Metadata Extraction**
- ✅ Feed with title → Metadata.Title set
- ✅ Feed without title → Metadata.Title empty
- ✅ Feed with link → Metadata.Link set
- ✅ Feed without link → Metadata.Link empty
- ✅ Feed with updated date → Metadata.Updated set
- ✅ Feed without updated date → Metadata.Updated = fetchTime

**normalizeEntry() - ID Generation (Critical)**
- ✅ Entry has GUID → Use GUID as ID
- ✅ Entry has no GUID, has link → Use link as ID
- ✅ Entry has no GUID/link, has title → Generate SHA256 hash from feedURL+title+date
- ✅ Entry has no GUID/link/title → Generate SHA256 hash from feedURL+content
- ✅ Two entries with same title but different dates → Different IDs
- ✅ Two entries with same title and date → Same ID (deduplication)
- ✅ Entry ID stability → Same entry generates same ID across parses

**extractAuthor() - Author Extraction**
- ✅ Entry has author.Name → Use entry author
- ✅ Entry has multiple authors → Use first author
- ✅ Entry has no author, feed has author → Use feed author
- ✅ Neither entry nor feed has author → Empty string
- ✅ Author is empty string → Empty string returned

**extractPublished() - Date Extraction with Fallbacks**
- ✅ Entry has PublishedParsed → Use entry published date
- ✅ Entry has no published, has UpdatedParsed → Use entry updated date
- ✅ Entry has no dates, feed has UpdatedParsed → Use feed updated date
- ✅ No dates anywhere → Use fetchTime as fallback
- ✅ Entry has both published and updated → Use published
- ✅ Date is zero value → Try next fallback
- ✅ Future date → Accept (no filtering at this layer)

**extractUpdated() - Updated Date**
- ✅ Entry has UpdatedParsed → Use it
- ✅ Entry has no updated → Use published date

**sanitizeHTML() - XSS Prevention (Critical Security)**
- ✅ Safe HTML (paragraph) → `<p>Hello</p>` → Unchanged
- ✅ Safe HTML (links) → `<a href="http://example.com">Link</a>` → Unchanged
- ✅ Safe HTML (images) → `<img src="https://example.com/img.jpg" alt="Image">` → Unchanged
- ❌ Script tag → `<script>alert(1)</script>` → Removed
- ❌ Script in img src → `<img src="javascript:alert(1)">` → Script removed
- ❌ Script in a href → `<a href="javascript:alert(1)">Click</a>` → Script removed
- ❌ Data URI in img → `<img src="data:text/html,<script>alert(1)</script>">` → Removed
- ❌ Data URI in a href → `<a href="data:text/html,<script>alert(1)</script>">` → Removed
- ❌ onclick handler → `<div onclick="alert(1)">Click</div>` → onclick removed
- ❌ onerror handler → `<img src="x" onerror="alert(1)">` → onerror removed
- ❌ onload handler → `<body onload="alert(1)">` → onload removed
- ❌ onmouseover handler → `<div onmouseover="alert(1)">` → onmouseover removed
- ❌ Iframe tag → `<iframe src="evil.com"></iframe>` → Removed
- ❌ Object tag → `<object data="evil.swf"></object>` → Removed
- ❌ Embed tag → `<embed src="evil.swf">` → Removed
- ❌ Base tag → `<base href="http://evil.com">` → Removed
- ❌ Meta refresh → `<meta http-equiv="refresh" content="0;url=evil.com">` → Removed
- ❌ Link to stylesheet (potential CSS injection) → Removed or sanitized
- ✅ Valid style attribute (safe CSS) → Allowed (if UGCPolicy allows)
- ❌ Expression in style → `<div style="background:expression(alert(1))">` → Removed
- ✅ Bold, italic, underline → `<b><i><u>Text</u></i></b>` → Unchanged
- ✅ Lists → `<ul><li>Item</li></ul>` → Unchanged
- ✅ Blockquote → `<blockquote>Quote</blockquote>` → Unchanged
- ✅ Code blocks → `<pre><code>code</code></pre>` → Unchanged
- ✅ Tables → `<table><tr><td>Cell</td></tr></table>` → Unchanged
- ❌ Unclosed tags → `<p>Text<p>` → Fixed by bluemonday
- ❌ Nested scripts → `<div><script>alert(1)</script></div>` → Script removed
- ❌ Obfuscated script → `<scr<script>ipt>alert(1)</script>` → Handled by bluemonday

**sanitizeHTML() - URL Scheme Validation**
- ✅ http:// URLs → Allowed
- ✅ https:// URLs → Allowed
- ❌ ftp:// URLs → Removed
- ❌ file:// URLs → Removed
- ❌ mailto: URLs → Removed (or allowed based on policy)
- ✅ Protocol-relative URLs → `//example.com/image.jpg` → Handled correctly

**resolveURL() - Relative URL Resolution**
- ✅ Absolute URL unchanged → `http://example.com/path` → Unchanged
- ✅ Relative path → `/path/to/page` + `http://example.com/feed` → `http://example.com/path/to/page`
- ✅ Relative path with ./ → `./image.jpg` + base → Resolved correctly
- ✅ Relative path with ../ → `../image.jpg` + base → Resolved correctly
- ✅ Query string → `?query=1` + base → Resolved correctly
- ✅ Fragment → `#section` + base → Resolved correctly
- ❌ Invalid base URL → Error returned
- ❌ Invalid href → Error returned

**SanitizeHTML() - Public API**
- ✅ Same behavior as private sanitizeHTML()
- ✅ Can be called independently for testing

**normalizeEntry() - Full Entry Processing**
- ✅ Complete entry with all fields → All fields normalized
- ✅ Minimal entry (only title) → Other fields default/generated
- ✅ Entry with relative image URLs → Images resolved to absolute
- ✅ Entry with relative link → Link resolved to absolute
- ✅ Entry with content and description → Content used, summary set
- ✅ Entry with description only → Content set from description
- ✅ Entry with neither content nor description → Content empty

**Branch Coverage Focus**
- Each fallback in ID generation (GUID → link → title hash → content hash)
- Each fallback in date extraction (entry published → entry updated → feed updated → fetch time)
- Each fallback in author extraction (entry author → authors[0] → feed author → empty)
- HTML sanitization: safe content vs dangerous content branches
- URL resolution: absolute vs relative URL branches
- Content vs description logic: has content, has description, has both, has neither

---

### 1.3 pkg/repository (Database Operations)

#### Test Coverage Goals
- **Target**: >85% line coverage, >95% branch coverage for CRUD operations
- **Critical Paths**: UPSERT logic, foreign key cascades, NULL handling

#### Test Cases

**New() - Database Initialization**
- ✅ Create new database file → Success, schema created
- ✅ Open existing database → Success, schema already exists
- ❌ Invalid path (read-only directory) → Error
- ✅ WAL mode enabled → PRAGMA journal_mode=WAL effective
- ✅ Foreign keys enabled → PRAGMA foreign_keys=ON effective
- ✅ Schema tables created → feeds and entries tables exist
- ✅ Schema indices created → All 6 indices exist

**AddFeed() - Feed Creation**
- ✅ Add new feed with URL and title → Success, returns feed ID
- ✅ Add feed with URL only (no title) → Success
- ❌ Add duplicate URL → Error (UNIQUE constraint)
- ✅ Add multiple feeds → Success, different IDs assigned
- ✅ next_fetch set to current time → Correct timestamp

**GetFeedByURL() - Feed Retrieval**
- ✅ Get existing feed → Returns feed with all fields
- ❌ Get non-existent feed → ErrFeedNotFound
- ✅ Feed with NULL title → Title is empty string
- ✅ Feed with NULL etag → ETag is empty string
- ✅ Feed with NULL last_modified → LastModified is empty string
- ✅ Feed with active=1 → Active=true
- ✅ Feed with active=0 → Active=false
- ✅ Feed with timestamps → Parsed correctly to time.Time

**GetFeeds() - List All Feeds**
- ✅ No feeds in database → Empty slice
- ✅ One feed → Returns slice with 1 feed
- ✅ Multiple feeds → Returns all feeds in order
- ✅ activeOnly=false → Returns all feeds (active and inactive)
- ✅ activeOnly=true → Returns only active feeds (active=1)
- ✅ Mix of active/inactive feeds → Filtering works correctly
- ✅ Feeds ordered by ID → Ascending order

**UpdateFeed() - Feed Metadata Update**
- ✅ Update title → Title changed
- ✅ Update link → Link changed
- ✅ Update updated timestamp → Updated changed
- ✅ Update all three → All changed
- ❌ Update non-existent feed ID → No error, 0 rows affected
- ✅ NULL values handled → Can update to empty string

**UpdateFeedCache() - HTTP Cache Headers**
- ✅ Update ETag → ETag stored exactly as received
- ✅ Update Last-Modified → LastModified stored exactly as received
- ✅ Update both → Both stored
- ✅ ETag with quotes `"abc123"` → Stored with quotes
- ✅ Update also sets last_fetched → Correct timestamp
- ✅ Update clears fetch_error → Error set to NULL
- ✅ Update resets fetch_error_count → Count set to 0
- ❌ Update non-existent feed → No error, 0 rows affected

**UpdateFeedError() - Error Tracking**
- ✅ Record first error → Error stored, count = 1
- ✅ Record second error → Count incremented to 2
- ✅ Record multiple errors → Count increments each time
- ✅ Update last_fetched → Timestamp updated
- ✅ Error message with special characters → Stored correctly
- ❌ Update non-existent feed → No error, 0 rows affected

**RemoveFeed() - Feed Deletion with Cascade**
- ✅ Remove feed with no entries → Feed deleted
- ✅ Remove feed with entries → Feed deleted, entries cascade-deleted
- ✅ Remove feed with 100 entries → All deleted due to CASCADE
- ❌ Remove non-existent feed → No error, 0 rows affected
- ✅ Foreign key constraint enforced → Can't have orphaned entries

**UpsertEntry() - Insert or Update Entry**
- ✅ Insert new entry → Entry added
- ✅ Insert duplicate (same feed_id + entry_id) → Entry updated, not duplicated
- ✅ Update changes title → Title updated
- ✅ Update changes content → Content updated
- ✅ Update changes link, author, updated → All updated
- ✅ Update does NOT change published → Published preserved
- ✅ Update does NOT change first_seen → FirstSeen preserved
- ✅ Insert entry with all fields → All stored correctly
- ✅ Insert entry with NULL author → NULL stored
- ✅ Insert entry with NULL summary → NULL stored
- ❌ Insert with invalid feed_id → Foreign key constraint error
- ✅ RFC3339 timestamps → Stored and retrieved correctly

**GetRecentEntries() - Query with Smart Fallback (Critical)**
- ✅ Entries within time window → Returns those entries
- ✅ No entries in time window, but 50+ total → Returns 50 most recent
- ✅ No entries in time window, only 10 total → Returns all 10
- ✅ Empty database → Returns empty slice
- ✅ days=7, entries exist → Returns last 7 days
- ✅ days=1, entries exist → Returns last 1 day
- ✅ days=30, no entries → Falls back to most recent 50
- ✅ Only inactive feeds have entries → Returns empty (active=1 filter)
- ✅ Mix of active/inactive feeds → Returns only active feed entries
- ✅ Entries sorted by published DESC → Newest first
- ✅ Fallback query limit=50 → Exactly 50 entries if available
- ✅ Exactly on cutoff date → Included (>=, not >)

**CountEntries() - Total Count**
- ✅ Empty database → 0
- ✅ One entry → 1
- ✅ Multiple entries → Correct count
- ✅ Includes entries from inactive feeds → Total count

**CountRecentEntries() - Count in Time Window**
- ✅ No entries → 0
- ✅ All entries within window → Count = total
- ✅ Some entries outside window → Count < total
- ✅ days=7 → Correct count
- ✅ days=0 (today only) → Correct count

**PruneOldEntries() - Deletion by Age**
- ✅ Prune with no old entries → 0 deleted
- ✅ Prune with all old entries → All deleted
- ✅ Prune with mix → Only old entries deleted
- ✅ days=90 → Entries older than 90 days deleted
- ✅ days=1 → Entries older than 1 day deleted
- ✅ Returns correct count → RowsAffected is accurate
- ✅ Feed metadata preserved → Only entries deleted

**Transaction Support**
- ✅ Batch insert in transaction → All committed or all rolled back
- ✅ Transaction rollback on error → No partial data
- ✅ Concurrent reads during write → WAL mode allows this

**NULL Handling (Critical for Branch Coverage)**
- ✅ scanFeed with NULL strings → Empty strings returned
- ✅ scanFeed with valid strings → Strings returned
- ✅ scanFeed with NULL timestamps → Zero time.Time
- ✅ scanFeed with valid timestamps → Parsed correctly
- ✅ scanEntries with NULL content → Empty string
- ✅ scanEntries with NULL author → Empty string

**Close() - Resource Cleanup**
- ✅ Close database → No error
- ✅ Operations after close → Error

**Branch Coverage Focus**
- GetRecentEntries: time window branch vs fallback branch
- scanFeed: each NULL check (title, link, etag, last_modified, etc.)
- scanEntries: each NULL check (title, link, author, content, etc.)
- UpsertEntry: INSERT branch vs UPDATE branch
- GetFeeds: activeOnly true vs false branches
- Foreign key cascade: with entries vs without entries

---

### 1.4 pkg/generator (HTML Generation)

#### Test Coverage Goals
- **Target**: >80% line coverage, >85% branch coverage
- **Critical Paths**: Template execution, XSS prevention, date grouping

#### Test Cases

**New() - Default Template**
- ✅ Create generator → Template compiled successfully
- ✅ Template functions registered → formatDate, relativeTime available
- ❌ Template syntax error (if default template broken) → Error

**NewWithTemplate() - Custom Template**
- ✅ Load valid custom template → Success
- ❌ Load non-existent template → Error
- ❌ Load template with syntax error → Error
- ✅ Custom template with custom HTML → Renders correctly

**Generate() - Basic Rendering**
- ✅ Empty data (no entries) → Valid HTML generated
- ✅ One entry → Entry rendered
- ✅ Multiple entries → All entries rendered
- ✅ Generator version added → "Rogue Planet v1.0" in output
- ✅ Updated timestamp added → Current time in output
- ✅ Template execution error → Error returned

**Generate() - Template Data Binding**
- ✅ Title bound → {{.Title}} renders correctly
- ✅ Subtitle bound → {{.Subtitle}} renders correctly
- ✅ Link bound → {{.Link}} renders correctly
- ✅ OwnerName bound → {{.OwnerName}} renders correctly
- ✅ OwnerEmail bound → {{.OwnerEmail}} renders correctly
- ✅ Entries bound → {{range .Entries}} iterates correctly
- ✅ Feeds bound → {{range .Feeds}} renders sidebar

**Generate() - Entry Data**
- ✅ Entry title (template.HTML) → Rendered unescaped (already sanitized)
- ✅ Entry content (template.HTML) → Rendered unescaped
- ✅ Entry link (string) → Rendered with escaping
- ✅ Entry author → Rendered correctly
- ✅ Entry FeedTitle → Rendered correctly
- ✅ Entry FeedLink → Rendered correctly
- ✅ Entry Published (time.Time) → Formatted correctly
- ✅ Entry PublishedRelative → "2 hours ago" etc.

**Generate() - Date Grouping**
- ✅ GroupByDate=false → All entries in single list
- ✅ GroupByDate=true → Entries grouped by date
- ✅ Entries on same date → Grouped together
- ✅ Entries on different dates → Separate groups
- ✅ Date groups ordered → Newest date first
- ✅ Entries within group ordered → Newest first
- ✅ Date group headers → "Today", "Yesterday", formatted dates

**Generate() - Feed Sidebar**
- ✅ No feeds → Sidebar not rendered
- ✅ One feed → Sidebar with 1 feed
- ✅ Multiple feeds → All feeds listed
- ✅ Feed title → Rendered correctly
- ✅ Feed link → Rendered correctly
- ✅ Feed URL in title attribute → Rendered
- ✅ Feed LastUpdated → "Updated X ago"
- ✅ Feed with ErrorCount=0 → No error display
- ✅ Feed with ErrorCount>0 → Error count displayed

**GenerateToFile() - File Output**
- ✅ Create output file → File created
- ✅ Create output directory → Directory created if not exists
- ✅ Write HTML to file → Content matches Generate() output
- ❌ Read-only directory → Error
- ❌ Invalid path → Error
- ✅ Overwrite existing file → Success

**CopyStaticAssets() - Asset Copying**
- ✅ No custom template → No assets copied (no error)
- ✅ Custom template, no static dir → No assets copied (no error)
- ✅ Custom template with static dir → Assets copied to output/static/
- ✅ Static dir with subdirectories → Recursive copy
- ✅ File permissions preserved → Correct permissions on copied files
- ✅ Existing static dir in output → Removed and replaced

**templateFuncs() - Template Functions**
- ✅ formatDate → "January 2, 2006 at 3:04 PM"
- ✅ formatDateShort → "Jan 2, 2006"
- ✅ formatDateISO → RFC3339 format
- ✅ relativeTime (see below)

**relativeTime() - Relative Date Formatting**
- ✅ <1 minute ago → "just now"
- ✅ 1 minute ago → "1 minute ago"
- ✅ 5 minutes ago → "5 minutes ago"
- ✅ 1 hour ago → "1 hour ago"
- ✅ 3 hours ago → "3 hours ago"
- ✅ Yesterday (24-48h) → "yesterday"
- ✅ 2 days ago → "2 days ago"
- ✅ 1 week ago → "1 week ago"
- ✅ 2 weeks ago → "2 weeks ago"
- ✅ 1 month ago → "1 month ago"
- ✅ 6 months ago → "6 months ago"
- ✅ 1 year ago → "1 year ago"
- ✅ 5 years ago → "5 years ago"

**groupEntriesByDate() - Date Grouping Logic**
- ✅ Empty entries → Empty groups
- ✅ All entries same date → 1 group
- ✅ Entries on 3 different dates → 3 groups
- ✅ Date boundaries → Entries at 23:59 and 00:01 in different groups
- ✅ Groups maintain entry order → Newest first within group
- ✅ Date order → Newest date group first

**formatDateGroup() - Date Header Formatting**
- ✅ Today → "Today"
- ✅ Yesterday → "Yesterday"
- ✅ This week (Mon-Sun) → "Monday, January 2"
- ✅ Older than 1 week → "Monday, January 2, 2006"
- ✅ Date in future → Handled correctly

**copyDir() / copyFile() - File Operations**
- ✅ Copy single file → Success
- ✅ Copy directory → Success
- ✅ Copy nested directories → Recursive copy
- ✅ Preserve file permissions → Permissions match source
- ❌ Source doesn't exist → Error
- ❌ Destination not writable → Error

**XSS Prevention in Templates (Critical)**
- ✅ template.HTML used for sanitized content → Rendered unescaped
- ✅ Regular strings auto-escaped → HTML entities escaped
- ✅ Content already sanitized by normalizer → No double-escaping
- ✅ User input (titles, authors) → Properly escaped

**Branch Coverage Focus**
- GroupByDate true vs false branches
- Feed sidebar rendering: no feeds vs has feeds
- Date grouping: today/yesterday/this week/older branches
- relativeTime: all time range branches
- CopyStaticAssets: has template vs no template, has static vs no static
- Feed ErrorCount: =0 vs >0 display branches

---

### 1.5 pkg/config (Configuration Parsing)

#### Test Coverage Goals
- **Target**: >80% line coverage, >85% branch coverage
- **Critical Paths**: INI parsing, default values, validation

#### Test Cases

**Load() - Basic Config Loading**
- ✅ Valid config.ini → All values loaded
- ✅ Config with all sections → planet, database, feeds sections parsed
- ❌ Non-existent file → Error
- ❌ Malformed INI (syntax error) → Error
- ✅ Empty config file → Defaults used
- ✅ Minimal config (only required fields) → Defaults for optional fields

**Planet Section**
- ✅ name → Loaded correctly
- ✅ link → Loaded correctly
- ✅ owner_name → Loaded correctly
- ✅ owner_email → Loaded correctly
- ✅ output_dir → Loaded correctly
- ✅ days (integer) → Parsed as int
- ✅ log_level → Loaded correctly
- ✅ concurrent_fetches → Parsed as int
- ✅ group_by_date (boolean) → Parsed as bool
- ✅ Missing optional fields → Defaults applied
- ❌ Invalid days (not a number) → Error or default
- ❌ Invalid concurrent_fetches → Error or default
- ❌ Invalid log_level → Error or default

**Database Section**
- ✅ path → Loaded correctly
- ✅ Missing path → Default "./data/planet.db"
- ✅ Relative path → Resolved correctly
- ✅ Absolute path → Used as-is

**Feed-Specific Overrides**
- ✅ [feed "URL"] sections → Parsed correctly
- ✅ Per-feed title override → Applied
- ✅ Per-feed ignore_in_feed → Applied
- ✅ Per-feed future_dates setting → Applied
- ✅ Multiple feed sections → All parsed

**Default Values**
- ✅ Default days = 7
- ✅ Default output_dir = "./public"
- ✅ Default database path = "./data/planet.db"
- ✅ Default concurrent_fetches = 5
- ✅ Default log_level = "info"
- ✅ Default group_by_date = false

**Validation**
- ❌ days < 1 → Error or default
- ❌ concurrent_fetches < 1 → Error or default
- ❌ concurrent_fetches > 50 → Error or default
- ✅ output_dir with spaces → Handled correctly
- ✅ email validation (optional) → Valid email format

**Branch Coverage Focus**
- Missing value vs present value for each config key
- Default value application branches
- Integer parsing success vs failure
- Boolean parsing success vs failure
- Feed section present vs absent

---

### 1.6 cmd/rp/commands.go (CLI Commands)

#### Test Coverage Goals
- **Target**: >75% line coverage, >80% branch coverage
- **Critical Paths**: Error handling, user-facing messages, database operations

#### Test Cases (per command)

**cmdInit()**
- ✅ Init in empty directory → Success, creates config.ini and database
- ✅ Init with -f feeds.txt → Imports feeds from file
- ❌ Init in directory with existing config.ini → Error or warning
- ✅ Creates output directory → ./public created
- ✅ Creates database directory → ./data created
- ✅ Default config content → Contains sensible defaults
- ❌ Invalid feeds file path → Error
- ❌ Malformed feeds file → Error

**cmdAddFeed()**
- ✅ Add valid URL → Feed added to database
- ✅ Add feed with autodiscovery → Title fetched from feed
- ❌ Add duplicate URL → Error message
- ❌ Add invalid URL → Error
- ❌ Add non-feed URL (HTML page) → Error or warning
- ❌ Network error during fetch → Error, but feed added to database
- ✅ Output message → "Added feed: <title>"

**cmdAddAll()**
- ✅ Import 5 feeds from file → All added
- ✅ File with comments → Comments ignored
- ✅ File with blank lines → Blank lines ignored
- ❌ File with invalid URLs → Errors reported, valid feeds added
- ✅ File with mix of valid/invalid → Partial success
- ✅ Empty file → No feeds added, no error
- ❌ File doesn't exist → Error

**cmdRemoveFeed()**
- ✅ Remove existing feed → Feed deleted
- ✅ Remove feed with entries → Feed and entries deleted (cascade)
- ❌ Remove non-existent feed → Error message
- ✅ Output message → "Removed feed: <url>"

**cmdListFeeds()**
- ✅ No feeds → "No feeds configured"
- ✅ One feed → Feed listed with title and URL
- ✅ Multiple feeds → All feeds listed
- ✅ Feed with error → Error count displayed
- ✅ Feed never fetched → "Never fetched" displayed
- ✅ Feed recently fetched → "Last fetched: X ago"

**cmdStatus()**
- ✅ Show feed count → "Feeds: 5"
- ✅ Show entry count → "Entries: 123"
- ✅ Show recent entry count → "Recent (7 days): 45"
- ✅ Empty database → "Feeds: 0, Entries: 0"

**cmdUpdate()**
- ✅ Fetch and generate → Both operations performed
- ✅ Verbose mode → Detailed output
- ✅ Quiet mode → Minimal output
- ✅ Some feeds fail → Errors logged, others succeed
- ✅ All feeds fail → Error summary
- ✅ No new entries → "No new entries"
- ✅ New entries → "Generated index.html with X entries"

**cmdFetch()**
- ✅ Fetch all feeds → All feeds fetched
- ✅ Feeds with 304 response → "Not modified" message
- ✅ Feeds with new content → "Fetched X new entries"
- ✅ Verbose mode → Shows each feed fetch
- ✅ Concurrent fetching → Multiple feeds fetched in parallel
- ❌ Network errors → Errors logged, continues with other feeds

**cmdGenerate()**
- ✅ Generate with default days → Uses config value
- ✅ Generate with --days flag → Overrides config
- ✅ No entries → "No entries to display"
- ✅ Entries exist → "Generated index.html with X entries"
- ❌ Template error → Error message
- ❌ Output directory not writable → Error

**cmdPrune()**
- ✅ Prune old entries → Entries deleted
- ✅ --dry-run → Shows what would be deleted, doesn't delete
- ✅ --days 90 → Deletes entries older than 90 days
- ✅ No old entries → "No entries to prune"
- ✅ Output → "Pruned X entries"

**Option Structs**
- ✅ All Options structs have Output field → Used for testing
- ✅ ConfigPath is configurable → Can test with temp config
- ✅ Verbose/Quiet flags → Affect output verbosity

**Branch Coverage Focus**
- Each command: success vs error branches
- Verbose vs quiet output branches
- Empty database vs populated database branches
- Network success vs failure branches
- File exists vs doesn't exist branches

---

## 2. Integration Tests

### 2.1 Full Pipeline Integration Tests

**Test: Complete End-to-End Workflow**
```
init → add-feed → fetch → generate → verify HTML
```
- ✅ Initialize empty planet
- ✅ Add 3 feeds (Atom, RSS, JSON Feed)
- ✅ Fetch all feeds (mock HTTP server)
- ✅ Generate HTML
- ✅ Verify HTML contains entries from all 3 feeds
- ✅ Verify sidebar lists all 3 feeds
- ✅ Verify HTML structure is valid
- ✅ Verify CSP header present in HTML

**Test: Update Cycle**
```
init → add-feed → fetch → generate → fetch again (304) → generate again
```
- ✅ First fetch downloads content
- ✅ Second fetch gets 304 Not Modified
- ✅ No duplicate entries in database
- ✅ Generated HTML unchanged (or minimal changes)

**Test: Feed Management Workflow**
```
init → add 5 feeds → list → remove 2 → list → fetch → generate
```
- ✅ Add multiple feeds
- ✅ List shows all feeds
- ✅ Remove feeds
- ✅ Removed feeds not fetched
- ✅ Removed feed entries not in generated HTML

**Test: Error Recovery Workflow**
```
init → add good feed + bad URL → fetch → generate
```
- ✅ Good feed fetched successfully
- ✅ Bad feed error logged
- ✅ HTML generated with good feed content
- ✅ Bad feed shown in sidebar with error count

**Test: Prune Workflow**
```
init → add-feed → fetch → (time passes) → prune → generate
```
- ✅ Initial entries stored
- ✅ Old entries pruned
- ✅ Recent entries remain
- ✅ HTML generated with only recent entries

**Test: Custom Template Workflow**
```
init → create custom template → add-feed → fetch → generate with custom template
```
- ✅ Custom template loaded
- ✅ Generated HTML uses custom template
- ✅ Static assets copied

**Test: Concurrent Fetch Workflow**
```
init → add 20 feeds → fetch with concurrent_fetches=10
```
- ✅ Multiple feeds fetched concurrently
- ✅ No race conditions
- ✅ All feeds processed
- ✅ No deadlocks

**Test: Large Feed Workflow**
```
init → add feed with 100 entries → fetch → generate
```
- ✅ All 100 entries stored
- ✅ All 100 entries in database
- ✅ HTML generated with all entries (if within time window)

**Test: Stale Feed Fallback**
```
init → add-feed → fetch → (feeds go stale) → generate
```
- ✅ No entries in configured time window (7 days)
- ✅ Fallback activates: shows 50 most recent entries
- ✅ HTML not empty despite stale feeds

**Test: Database Migration (future)**
```
Old DB schema → run rp → auto-migrate
```
- Future test for schema upgrades

### 2.2 Real-World Feed Integration Tests

Use saved snapshots in testdata/ for reproducibility:

**Test: Daring Fireball (Atom)**
- ✅ Parse Daring Fireball feed snapshot
- ✅ Extract all entries
- ✅ Verify author is "John Gruber"
- ✅ Verify content is sanitized
- ✅ Verify links are absolute

**Test: Asymco (RSS)**
- ✅ Parse Asymco feed snapshot
- ✅ Extract entries with images
- ✅ Verify image URLs resolved correctly
- ✅ Verify HTML sanitization

**Test: Malformed Feed Recovery**
- ✅ Feed with invalid dates → Use fallback dates
- ✅ Feed with missing entry IDs → Generate IDs
- ✅ Feed with broken HTML → Sanitized and rendered
- ✅ Feed with relative URLs → Resolved to absolute
- ✅ Feed with incorrect encoding declaration → Handled gracefully

**Test: Mixed Feed Formats**
```
Add Atom + RSS 2.0 + RSS 1.0 + JSON Feed → fetch all → generate
```
- ✅ All formats parsed correctly
- ✅ All normalized to same internal format
- ✅ Chronological order in generated HTML

### 2.3 Security Integration Tests

**Test: XSS Prevention End-to-End**
```
Mock feed with malicious content → fetch → generate → verify HTML safe
```
- ✅ Script tags removed from feed content
- ✅ Event handlers removed
- ✅ javascript: URLs removed
- ✅ Generated HTML has CSP header
- ✅ No executable JavaScript in output

**Test: SSRF Prevention**
```
Try to add feed with localhost URL → error
Try to add feed with private IP → error
```
- ✅ localhost blocked
- ✅ 127.0.0.1 blocked
- ✅ 192.168.x.x blocked
- ✅ 10.x.x.x blocked
- ✅ Error message clear to user

**Test: SQL Injection Prevention**
```
Feed with SQL in title/content → fetch → no SQL execution
```
- ✅ SQL in feed title → Stored as text, not executed
- ✅ SQL in entry content → Stored as text
- ✅ All database operations use prepared statements

**Test: Path Traversal Prevention**
```
Config with output_dir="../../../etc" → generate → error or safe path
```
- ✅ Path traversal blocked
- ✅ Files written to safe location

### 2.4 Performance Integration Tests

**Test: Concurrent Fetch Performance**
```
Fetch 50 feeds concurrently (concurrent_fetches=10)
```
- ✅ Completes in reasonable time
- ✅ No race conditions
- ✅ Memory usage reasonable

**Test: Large Database Performance**
```
Database with 10,000 entries → generate
```
- ✅ Query performance acceptable (<1s)
- ✅ HTML generation completes
- ✅ Indices used effectively

**Test: Memory Usage**
```
Fetch feed with 1000 entries → monitor memory
```
- ✅ Memory usage reasonable
- ✅ No memory leaks

---

## 3. Edge Cases and Error Scenarios

### 3.1 Network Error Scenarios

**Test: Timeout**
- ✅ Feed server timeout → Error logged, continues with other feeds
- ✅ Partial response → Error handled gracefully

**Test: DNS Failures**
- ✅ DNS resolution fails → Error logged, feed marked as error

**Test: TLS Errors**
- ✅ Invalid TLS certificate → Error (or option to ignore)
- ✅ TLS handshake timeout → Error logged

**Test: HTTP Errors**
- ✅ 404 Not Found → Error logged, feed marked as failed
- ✅ 429 Too Many Requests → Backoff, retry
- ✅ 500 Server Error → Retry with backoff
- ✅ 503 Service Unavailable → Retry

**Test: Redirect Chains**
- ✅ 3 redirects → Followed, final URL used
- ✅ 6 redirects → Error (exceeds limit)
- ✅ Redirect loop → Error after max redirects

### 3.2 Feed Content Edge Cases

**Test: Empty Feeds**
- ✅ Feed with 0 entries → Stored, no entries added
- ✅ Feed with only future-dated entries → Handled per config

**Test: Encoding Issues**
- ✅ Feed declares UTF-8, actually Windows-1252 → Handled gracefully
- ✅ Feed with invalid UTF-8 → Replacement characters used
- ✅ Feed with HTML entities → Decoded correctly

**Test: Date Edge Cases**
- ✅ Missing dates → Fallback to fetch time
- ✅ Invalid date format → Fallback or error
- ✅ Future dates → Configurable behavior
- ✅ Year 1900 dates → Handled
- ✅ Year 2100 dates → Handled

**Test: Huge Content**
- ✅ 10MB feed → Accepted
- ✅ 11MB feed → Rejected (exceeds limit)
- ✅ Very long entry content → Stored and rendered

**Test: Special Characters**
- ✅ Feed with Unicode emoji → Rendered correctly
- ✅ Feed with RTL text → Rendered correctly
- ✅ Feed with mathematical symbols → Rendered correctly

### 3.3 Database Edge Cases

**Test: Concurrent Access**
- ✅ Fetch while generate running → Both succeed (WAL mode)
- ✅ Multiple fetch processes → Serialized by SQLite

**Test: Disk Full**
- ❌ Disk full during write → Error, no corruption
- ✅ WAL checkpoint fails → Error handled

**Test: Database Corruption**
- ❌ Corrupted DB file → Error message, suggest restore from backup

**Test: Very Long Strings**
- ✅ 100KB entry content → Stored successfully
- ✅ URL with 1000 characters → Handled

### 3.4 File System Edge Cases

**Test: Permissions**
- ❌ Output directory read-only → Error
- ❌ Database directory read-only → Error
- ❌ Config file not readable → Error

**Test: Path Issues**
- ✅ Output path with spaces → Handled correctly
- ✅ Output path with Unicode → Handled correctly
- ✅ Relative vs absolute paths → Resolved correctly

**Test: Symlinks**
- ✅ Config via symlink → Followed correctly
- ✅ Output dir is symlink → Followed correctly

---

## 4. Test Execution Strategy

### 4.1 Test Commands

```bash
# Run all unit tests (fast)
make test

# Run all tests with verbose output
make test ARGS=-v

# Run specific package
go test ./pkg/crawler -v

# Run with coverage
make coverage

# Run with race detector
make test-race

# Run integration tests only
make test-integration

# Run network tests (requires internet)
go test -tags=network ./pkg/crawler -v

# Run all tests including network
go test -tags=network ./...

# Benchmark
make bench
```

### 4.2 Test Organization

```
pkg/crawler/
  crawler_test.go              # Unit tests
  crawler_live_test.go         # Network tests (requires -tags=network)
  crawler_user_agent_test.go   # Specific feature tests

pkg/normalizer/
  normalizer_test.go           # Unit tests with mock feeds
  normalizer_realworld_test.go # Tests with real feed snapshots

cmd/rp/
  commands_test.go             # Integration tests for CLI commands
  integration_test.go          # Full pipeline integration tests
  realworld_integration_test.go # Real-world feed integration tests

testdata/
  daring-fireball.xml          # Feed snapshots for reproducible tests
  asymco.xml
  malicious-feed.xml           # XSS test cases
  malformed-feed.xml           # Error handling test cases
```

### 4.3 Coverage Targets

**Overall Target: >75% line coverage**

**Per-Package Targets:**
- pkg/crawler: >85% (critical path)
- pkg/normalizer: >80% (complex logic)
- pkg/repository: >85% (data integrity)
- pkg/generator: >80% (template rendering)
- pkg/config: >80% (parsing logic)
- cmd/rp: >75% (user-facing)

**Branch Coverage Targets:**
- Security-critical paths (SSRF, XSS, SQL): >95%
- Error handling paths: >85%
- Fallback logic: >90%

### 4.4 CI/CD Integration

**Pre-commit Checks:**
```bash
make quick  # fmt + test + build
```

**Pull Request Checks:**
```bash
make check  # fmt + vet + test + race
make coverage  # Generate coverage report
```

**Release Checks:**
```bash
make check
make test-integration
go test -tags=network ./...  # If CI has network access
make bench  # Performance regression check
```

### 4.5 Test Data Management

**Mock HTTP Server**: Use `httptest.NewServer()` for predictable tests
**Saved Snapshots**: Store real feed snapshots in testdata/
**Test Fixtures**: Use t.TempDir() for temporary files (auto-cleanup)
**Time Mocking**: Use fixed timestamps for date-dependent tests

---

## 5. Test Prioritization

### P0 (Critical - Must Pass)
- SSRF validation tests
- XSS sanitization tests
- SQL injection prevention tests
- HTTP conditional request tests
- Database UPSERT logic tests
- End-to-end integration test

### P1 (High Priority)
- All unit tests for core packages
- Error handling tests
- Fallback logic tests
- Concurrent fetch tests
- Feed format parsing tests

### P2 (Medium Priority)
- Edge case tests
- Real-world feed tests
- Performance tests
- CLI command tests

### P3 (Nice to Have)
- Benchmark tests
- Network tests (require internet)
- Stress tests (large databases)

---

## 6. Test Maintenance

**Adding New Tests:**
1. For new features, add tests BEFORE implementation (TDD)
2. For bug fixes, add regression test that reproduces the bug
3. Update this test plan when adding new test categories

**Reviewing Test Coverage:**
```bash
make coverage  # Generate HTML report
open coverage/coverage.html  # View in browser
```

**Identifying Missing Tests:**
- Look for uncovered lines in coverage report
- Focus on uncovered branches in conditionals
- Add tests for error paths
- Add tests for edge cases

**Test Code Quality:**
- Tests should be readable and maintainable
- Use descriptive test names: `TestFetch_WithETag_Returns304`
- Use table-driven tests for similar scenarios
- Don't repeat setup code - use helper functions
- Clean up resources with t.Cleanup() or defer

---

## 7. Example Table-Driven Test

```go
func TestValidateURL(t *testing.T) {
    tests := []struct {
        name    string
        url     string
        wantErr error
    }{
        {"valid http", "http://example.com/feed", nil},
        {"valid https", "https://example.com/feed", nil},
        {"localhost by name", "http://localhost/feed", ErrPrivateIP},
        {"localhost by ip", "http://127.0.0.1/feed", ErrPrivateIP},
        {"private ip", "http://192.168.1.1/feed", ErrPrivateIP},
        {"invalid scheme", "ftp://example.com/feed", ErrInvalidScheme},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateURL(tt.url)
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("ValidateURL(%q) error = %v, want %v",
                    tt.url, err, tt.wantErr)
            }
        })
    }
}
```

---

## 8. Continuous Improvement

**Quarterly Reviews:**
- Review test coverage reports
- Identify weak spots
- Add tests for production bugs
- Update test plan with new scenarios

**Metrics to Track:**
- Overall code coverage (target: >75%)
- Branch coverage for critical paths (target: >90%)
- Test execution time (target: <30s for unit tests)
- Number of flaky tests (target: 0)

**Test Performance:**
- Keep unit tests fast (<100ms per test)
- Use t.Parallel() for independent tests
- Mock external dependencies
- Use build tags for slow network tests

---

## Appendix: Testing Tools and Libraries

**Standard Library:**
- `testing` - Core testing framework
- `testing/httptest` - Mock HTTP servers
- `testing/iotest` - I/O error injection

**Third-Party (if needed):**
- `github.com/stretchr/testify/assert` - Assertion helpers (optional)
- `github.com/google/go-cmp/cmp` - Deep comparison (optional)

**Coverage Tools:**
- `go test -cover` - Basic coverage
- `go test -coverprofile` - Detailed coverage data
- `go tool cover -html` - HTML coverage report

**Race Detection:**
- `go test -race` - Detect race conditions

**Benchmarking:**
- `go test -bench` - Run benchmarks
- `go test -benchmem` - Memory allocation stats

---

## Summary

This test plan ensures Rogue Planet maintains high quality through:

1. **Comprehensive Unit Tests**: >75% coverage with focus on branch coverage
2. **Diverse Integration Tests**: Full workflows and real-world scenarios
3. **Security-First Testing**: >95% coverage of security-critical paths
4. **Edge Case Coverage**: Boundary conditions and error scenarios
5. **Performance Testing**: Concurrent operations and scalability
6. **Maintainable Test Code**: Clear organization and table-driven tests

**Next Steps:**
1. Implement remaining unit tests to reach coverage targets
2. Add integration tests for all CLI commands
3. Create test fixtures (feed snapshots) in testdata/
4. Set up CI/CD to run tests automatically
5. Monitor coverage reports and address gaps

This test plan should be treated as a living document, updated as the project evolves and new edge cases are discovered.
