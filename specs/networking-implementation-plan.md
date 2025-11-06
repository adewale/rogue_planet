# Networking Implementation Plan
# Comprehensive Plan for HTTP, Retry, Rate Limiting, Connection Pooling, and Intelligent Scheduling

**Created**: 2025-10-29
**Status**: Planning
**Scope**: All networking and performance features from v0.4.0 and v1.0.0 plans

---

## Overview

This document consolidates all networking-related work from v0.4.0-plan.md and v1.0.0-plan.md into a coherent implementation plan. Features are organized by dependency and complexity for efficient implementation.

**Goals**:
1. Improve HTTP performance (connection pooling, retries)
2. Be a good netizen (rate limiting, respect Cache-Control)
3. Reduce bandwidth (smart scheduling, 301 redirect handling)
4. Scale gracefully (adaptive polling, exponential backoff)

---

## Current State Assessment

### ✅ Already Implemented
- HTTP conditional requests (ETag/Last-Modified)
- SSRF prevention (localhost, private IP blocking)
- Gzip decompression
- Response size limiting (10MB)
- Timeout handling (30s default)
- Redirect following (301/302)
- User-Agent headers
- `FetchWithRetry()` method with exponential backoff

### ❌ Missing/Incomplete
- Connection pooling (new connection per fetch)
- Rate limiting (no per-domain throttling)
- 301 redirect URL updates (followed but not stored)
- 308 redirect handling (not tracked as permanent)
- Jitter for exponential backoff (deterministic delays)
- Intelligent scheduling (all feeds fetched every time)
- Cache-Control header support
- FetchWithRetry unused in production code
- Retry logic not used in cmd/rp

### 📊 Performance Impact
- Creating new connections: ~50-200ms overhead per fetch
- No connection reuse: 2-3x slower than with pooling
- No rate limiting: risk of being blocked by servers
- No smart scheduling: wasted bandwidth on slow feeds

---

## Implementation Phases

### Phase 1: Foundation (Quick Wins) 🏃
**Effort**: 1 day | **Dependencies**: None

These are "quick wins" that improve performance immediately with minimal risk.

#### 1.1 HTTP Connection Pooling
**Source**: v0.4.0 P5.1
**File**: `pkg/crawler/crawler.go`
**Benefit**: 10-20% faster fetching

**Implementation**:
```go
// In Crawler.New()
transport := &http.Transport{
    MaxIdleConns:        100,  // Total idle connections
    MaxIdleConnsPerHost: 10,   // Per-host limit (good netizen)
    IdleConnTimeout:     90 * time.Second,
    DisableKeepAlives:   false,
    MaxConnsPerHost:     0,    // No limit (use MaxIdleConnsPerHost)
}

client := &http.Client{
    Transport: transport,
    Timeout:   DefaultTimeout,
    // ... existing CheckRedirect
}
```

**Testing**:
- Unit test: verify Transport configured correctly
- Integration test: measure fetch time improvement (benchmark)
- Test connection reuse with multiple fetches to same domain

**Rationale**: Reusing TCP connections eliminates handshake overhead. 10 connections per host is conservative (polite to servers).

---

#### 1.2 Use FetchWithRetry in Production
**Source**: v0.4.0 P4.3
**File**: `cmd/rp/commands.go` (fetchFeeds function)
**Benefit**: More reliable fetching, handles transient failures

**Current Problem**:
```go
// cmd/rp/commands.go:842
resp, err := c.Fetch(ctx, f.URL, cache)  // No retry logic!
```

**Solution**:
```go
// Use existing FetchWithRetry (already implemented and tested)
resp, err := c.FetchWithRetry(ctx, f.URL, cache, 3)  // 3 retries
```

**Configuration**:
```go
const (
    DefaultMaxRetries = 3
    RetryDelay        = 2 * time.Second  // Already in FetchWithRetry
)
```

**Testing**:
- Integration test: flaky server (50% failure rate) succeeds with retries
- Integration test: permanent failures don't retry forever
- Test that 429 (rate limit) is retried with backoff

**Rationale**: FetchWithRetry already exists (lines 266-310 in crawler.go), tested, and working. Just needs to be used.

---

#### 1.3 Connection Pool Configuration
**Source**: New (makes 1.1 configurable)
**File**: `pkg/config/config.go`
**Benefit**: Users can tune for their feed count

**Add to config**:
```ini
[planet]
max_idle_connections = 100
max_connections_per_host = 10
connection_timeout = 90
```

**Implementation**:
```go
type PlanetConfig struct {
    // ... existing fields
    MaxIdleConnections     int
    MaxConnectionsPerHost  int
    ConnectionTimeout      int
}

// Default() sets sensible defaults
MaxIdleConnections:    100,
MaxConnectionsPerHost: 10,
ConnectionTimeout:     90,
```

**Testing**:
- Config parsing tests
- Verify applied to http.Transport
- Test with 0 (unlimited), negative (use defaults)

**Rationale**: Power users with 1000+ feeds may want higher limits. Conservative defaults protect servers.

---

#### 1.4 Retry-After Header Support (NEW - from audit)
**Source**: networking-features-audit.md #1 (HIGH PRIORITY)
**File**: `pkg/crawler/crawler.go`
**Benefit**: Respect server rate limits, prevent bans

**Current Problem**:
```go
// FetchWithRetry uses fixed exponential backoff
// Ignores Retry-After header from 429/503 responses
```

**Solution**:
```go
// Parse Retry-After header (RFC 7231 Section 7.1.3)
func ParseRetryAfter(header string) (time.Duration, error) {
    // Try parsing as seconds: "120"
    if seconds, err := strconv.Atoi(header); err == nil {
        return time.Duration(seconds) * time.Second, nil
    }

    // Try parsing as HTTP-date: "Wed, 21 Oct 2015 07:28:00 GMT"
    t, err := http.ParseTime(header)
    if err != nil {
        return 0, fmt.Errorf("invalid Retry-After format: %w", err)
    }

    delay := time.Until(t)
    if delay < 0 {
        delay = 0  // Already past, retry immediately
    }

    return delay, nil
}

// In FetchWithRetry (around line 295):
if resp.StatusCode == 429 || resp.StatusCode == 503 {
    if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
        delay, err := ParseRetryAfter(retryAfter)
        if err == nil {
            time.Sleep(delay)
            continue
        }
    }
    // Fall back to exponential backoff if no valid Retry-After
}
```

**Testing**:
- Unit test: Parse "120" → 120 seconds
- Unit test: Parse HTTP-date → duration until that time
- Unit test: Parse invalid format → error
- Unit test: Parse past date → 0 delay
- Integration test: 429 with Retry-After is honored
- Integration test: 503 with Retry-After is honored

**Rationale**: Critical for production. Servers explicitly say "wait N seconds" and we must respect that to avoid bans.

---

#### 1.5 Timeout Configuration (NEW - from audit)
**Source**: networking-features-audit.md #3 (HIGH PRIORITY)
**File**: `pkg/config/config.go`, `pkg/crawler/crawler.go`
**Benefit**: Tune for slow/fast servers, better control

**Current Problem**:
```go
// Hardcoded timeout in crawler.go:25
const DefaultTimeout = 30 * time.Second
```

**Solution**:
```ini
# Add to config.ini
[planet]
http_timeout = 30         # Total request timeout (seconds)
connect_timeout = 10      # TCP connection timeout (seconds)
read_timeout = 30         # Response read timeout (seconds)
```

**Implementation**:
```go
// pkg/config/config.go
type PlanetConfig struct {
    // ... existing fields
    HTTPTimeout     int  // Total timeout (seconds)
    ConnectTimeout  int  // TCP connect timeout (seconds)
    ReadTimeout     int  // Response read timeout (seconds)
}

// Default()
HTTPTimeout:    30,
ConnectTimeout: 10,
ReadTimeout:    30,

// pkg/crawler/crawler.go - NewWithConfig()
func NewWithConfig(config PlanetConfig) *Crawler {
    transport := &http.Transport{
        DialContext: (&net.Dialer{
            Timeout: time.Duration(config.ConnectTimeout) * time.Second,
        }).DialContext,
        ResponseHeaderTimeout: time.Duration(config.ReadTimeout) * time.Second,
        // ... connection pooling settings
    }

    client := &http.Client{
        Transport: transport,
        Timeout:   time.Duration(config.HTTPTimeout) * time.Second,
        // ... existing CheckRedirect
    }

    return &Crawler{
        client:    client,
        userAgent: config.UserAgent,
        maxSize:   MaxFeedSize,
    }
}
```

**Testing**:
- Config parsing tests
- Unit test: Verify timeouts applied to transport
- Integration test: Connect timeout triggers
- Integration test: Read timeout triggers
- Test with 0 (use defaults)

**Rationale**: Production needs vary. Some feeds are slow (increase timeout), some are fast (decrease timeout to fail fast).

---

#### 1.6 Add Jitter to Exponential Backoff (NEW - moved from Phase 4)
**Source**: Phase 4.4 (simplified for immediate use)
**File**: `pkg/crawler/crawler.go`
**Benefit**: Prevent thundering herd on synchronized retries

**Current Problem**:
```go
// FetchWithRetry line 454-455
// Exponential backoff: 1s, 2s, 4s, 8s...
backoff = time.Duration(1<<uint(attempt-1)) * time.Second
// No jitter - all clients retry at same intervals!
```

**Solution**:
```go
import "math/rand"

// In FetchWithRetry around line 454
backoff = time.Duration(1<<uint(attempt-1)) * time.Second

// Add ±10% jitter to prevent synchronized retries
jitterRange := float64(backoff) * 0.1
jitter := time.Duration((rand.Float64()*2-1) * jitterRange)
backoff += jitter
```

**Testing**:
- Unit test: Verify jitter is within ±10% of base backoff
- Unit test: Test multiple retries produce different delays
- Unit test: Test with rand seed for deterministic testing
- Benchmark: Verify jitter overhead is negligible

**Rationale**: Industry best practice (AWS, Google SRE books). Without jitter, if 100 feeds fail at the same instant, they all retry at exactly 1s, 2s, 4s intervals - overwhelming the server. Jitter spreads retries across a range.

---

#### 1.7 Add HTTP 308 Permanent Redirect Support (NEW)
**Source**: HTTP 308 investigation (RFC 7538)
**File**: `pkg/crawler/crawler.go`
**Benefit**: Handle modern permanent redirects correctly

**Current Problem**:
```go
// CheckRedirect line 304
if req.Response.StatusCode == http.StatusMovedPermanently {
    sawPermanentRedirect = true
}
// Only detects 301, ignores 308!
```

**Solution**:
```go
// Detect both 301 and 308 as permanent
if req.Response.StatusCode == http.StatusMovedPermanently ||
   req.Response.StatusCode == http.StatusPermanentRedirect {
    sawPermanentRedirect = true
}
```

**Testing**:
- Unit test: 308 redirect marked as permanent
- Unit test: 308 triggers URL update in database
- Unit test: Multiple redirects (301→308) marked permanent
- Integration test: 308 equivalent to 301 for feed fetching

**Rationale**: Modern servers use 308 for permanent redirects (WordPress, many CMSs). For GET requests (feed fetching), 308 behaves identically to 301. This is a one-line change with zero downside.

---

### Phase 2: 301/308 Redirect Handling 🔄
**Effort**: 0.5 days | **Dependencies**: Phase 1

Permanent redirects should update the stored feed URL to avoid extra hop on every fetch.

**Source**: v0.4.0 P3.2, v1.0.0 notes

#### 2.1 Track 301 vs 302 Redirects
**File**: `pkg/crawler/crawler.go`

**Current State**:
- FeedResponse has `FinalURL` field (line 53)
- Populated after redirects (line 233)
- Never used!

**Enhancement**:
```go
type FeedResponse struct {
    Body            []byte
    StatusCode      int
    NotModified     bool
    NewCache        FeedCache
    FinalURL        string
    FetchTime       time.Time
    PermanentRedirect bool  // NEW: True if 301 encountered
}

// Track redirects in CheckRedirect
client := &http.Client{
    CheckRedirect: func(req *http.Request, via []*http.Request) error {
        if len(via) >= MaxRedirects {
            return fmt.Errorf("stopped after %d redirects", MaxRedirects)
        }
        // Track if any redirect was 301
        for _, r := range via {
            if r.Response.StatusCode == 301 {
                // Store in closure variable
                permanentRedirect = true
            }
        }
        return nil
    },
}
```

**Testing**:
- Test server returns 301 → PermanentRedirect = true
- Test server returns 302 → PermanentRedirect = false
- Test multiple redirects (301 → 302) → true
- Test no redirects → false

---

#### 2.2 Add UpdateFeedURL to Repository
**File**: `pkg/repository/repository.go`

**Implementation**:
```go
// UpdateFeedURL changes a feed's URL (for 301 permanent redirects)
func (r *Repository) UpdateFeedURL(oldURL, newURL string) error {
    result, err := r.db.Exec(`
        UPDATE feeds
        SET url = ?, updated_at = CURRENT_TIMESTAMP
        WHERE url = ? AND active = 1
    `, newURL, oldURL)

    if err != nil {
        return fmt.Errorf("update feed URL: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("check rows affected: %w", err)
    }

    if rows == 0 {
        return ErrFeedNotFound
    }

    return nil
}
```

**Testing**:
- Update existing feed URL → success
- Update non-existent feed → ErrFeedNotFound
- Update inactive feed → ErrFeedNotFound
- Update to existing URL → unique constraint error

---

#### 2.3 Auto-Update on 301 in fetchFeeds
**File**: `cmd/rp/commands.go`

**Implementation**:
```go
// After successful fetch in fetchFeeds()
if resp.PermanentRedirect && resp.FinalURL != f.URL {
    logger.Info("301 redirect detected: %s → %s", f.URL, resp.FinalURL)

    err := repo.UpdateFeedURL(f.URL, resp.FinalURL)
    if err != nil {
        logger.Warn("Failed to update feed URL: %v", err)
        // Don't fail fetch - just log warning
    } else {
        logger.Info("Updated feed URL in database")
    }
}
```

**Configuration**:
```ini
[planet]
auto_update_redirects = true  # Default: true
```

**Testing**:
- Integration test: feed returns 301 → URL updated in DB
- Integration test: next fetch uses new URL directly
- Integration test: config=false → URL not updated
- Test logging output

**Rationale**: Following redirects forever wastes bandwidth. HTTP→HTTPS migrations are permanent.

---

### Phase 3: Rate Limiting 🚦
**Effort**: 1 day | **Dependencies**: Phase 1

Prevent overwhelming servers with too many requests. Critical for being a good netizen.

**Source**: v0.4.0 P3.4 (deferred to v1.0), CLAUDE.md references

#### 3.1 Per-Domain Rate Limiter
**File**: `pkg/crawler/ratelimiter.go` (new file)

**Implementation**:
```go
package crawler

import (
    "sync"
    "time"
    "net/url"
    "golang.org/x/time/rate"
)

// DomainRateLimiter manages per-domain rate limits
type DomainRateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex

    // Default: 1 request per second per domain
    requestsPerSecond rate.Limit
    burst             int
}

// NewDomainRateLimiter creates a rate limiter
func NewDomainRateLimiter(requestsPerSecond float64, burst int) *DomainRateLimiter {
    return &DomainRateLimiter{
        limiters:          make(map[string]*rate.Limiter),
        requestsPerSecond: rate.Limit(requestsPerSecond),
        burst:             burst,
    }
}

// Wait blocks until request can proceed for given URL
func (d *DomainRateLimiter) Wait(ctx context.Context, urlStr string) error {
    limiter := d.getLimiter(urlStr)
    return limiter.Wait(ctx)
}

// getLimiter returns rate limiter for URL's domain
func (d *DomainRateLimiter) getLimiter(urlStr string) *rate.Limiter {
    u, err := url.Parse(urlStr)
    if err != nil {
        // Fallback to whole URL if parse fails
        return d.getLimiterForDomain(urlStr)
    }

    domain := u.Hostname()
    return d.getLimiterForDomain(domain)
}

func (d *DomainRateLimiter) getLimiterForDomain(domain string) *rate.Limiter {
    d.mu.RLock()
    limiter, exists := d.limiters[domain]
    d.mu.RUnlock()

    if exists {
        return limiter
    }

    d.mu.Lock()
    defer d.mu.Unlock()

    // Double-check after acquiring write lock
    limiter, exists = d.limiters[domain]
    if exists {
        return limiter
    }

    // Create new limiter for this domain
    limiter = rate.NewLimiter(d.requestsPerSecond, d.burst)
    d.limiters[domain] = limiter
    return limiter
}
```

**Testing**:
- Unit test: same domain rate-limited
- Unit test: different domains not rate-limited
- Unit test: burst allows immediate requests
- Unit test: context cancellation works
- Benchmark: overhead of rate limiting

---

#### 3.2 Integrate Rate Limiter in Crawler
**File**: `pkg/crawler/crawler.go`

**Implementation**:
```go
type Crawler struct {
    client        *http.Client
    userAgent     string
    maxSize       int64
    skipSSRFCheck bool
    rateLimiter   *DomainRateLimiter  // NEW
}

// NewWithRateLimiter creates crawler with rate limiting
func NewWithRateLimiter(requestsPerSecond float64, burst int) *Crawler {
    c := New()
    c.rateLimiter = NewDomainRateLimiter(requestsPerSecond, burst)
    return c
}

// Fetch uses rate limiter if configured
func (c *Crawler) Fetch(ctx context.Context, feedURL string, cache FeedCache) (*FeedResponse, error) {
    // Rate limit before making request
    if c.rateLimiter != nil {
        if err := c.rateLimiter.Wait(ctx, feedURL); err != nil {
            return nil, fmt.Errorf("rate limit wait: %w", err)
        }
    }

    // ... existing fetch logic
}
```

**Configuration**:
```ini
[planet]
enable_rate_limiting = true
rate_limit_per_second = 1.0   # 1 req/sec per domain
rate_limit_burst = 3          # Allow 3 immediate requests
```

**Testing**:
- Integration test: multiple fetches to same domain delayed
- Integration test: disabled rate limiting works
- Integration test: burst allows quick requests
- Test with config.Config

**Dependencies**: Adds `golang.org/x/time/rate` to go.mod

---

#### 3.3 Update go.mod and Documentation
**Files**: `go.mod`, `specs/TODO.md`, `CLAUDE.md`

**go.mod**:
```
require (
    golang.org/x/time v0.7.0  // NEW: Rate limiting
)
```

**TODO.md line 200-204**: Already fixed in P2.1, but now add:
```markdown
- `golang.org/x/time/rate` - Per-domain rate limiting
```

**CLAUDE.md**: Update dependencies section

---

### Phase 4: Intelligent Scheduling 🧠
**Effort**: 3-4 days | **Dependencies**: Phases 1-3

Fetch feeds at appropriate intervals based on update frequency. Most complex feature.

**Source**: v1.0.0 Section 2 (Intelligent Feed Scheduling)

#### 4.1 Database Schema Verification
**File**: `pkg/repository/repository.go`

**Current State**:
- `next_fetch` column EXISTS (line 256 of repository.go)
- `fetch_interval` column EXISTS
- `last_updated` column EXISTS
- Index on `next_fetch` EXISTS
- Fields populated but NEVER consulted!

**Verification**:
```sql
-- Verify schema
SELECT
    id, url, next_fetch, fetch_interval, last_updated
FROM feeds
WHERE active = 1;

-- Check index
PRAGMA index_list(feeds);
PRAGMA index_info(idx_feeds_next_fetch);
```

**No migration needed** - schema ready!

---

#### 4.2 GetFeedsReadyForFetch Query
**File**: `pkg/repository/repository.go`

**Implementation**:
```go
// GetFeedsReadyForFetch returns active feeds due for fetching
func (r *Repository) GetFeedsReadyForFetch(now time.Time) ([]Feed, error) {
    rows, err := r.db.Query(`
        SELECT
            id, url, title, link, etag, last_modified,
            last_fetched, fetch_error, fetch_error_count,
            next_fetch, fetch_interval, active
        FROM feeds
        WHERE active = 1
          AND (next_fetch IS NULL OR next_fetch <= ?)
        ORDER BY next_fetch ASC
    `, now.Format(time.RFC3339))

    if err != nil {
        return nil, fmt.Errorf("query ready feeds: %w", err)
    }
    defer rows.Close()

    // ... scan rows (same as GetFeeds)
}
```

**Testing**:
- Test next_fetch = NULL → included
- Test next_fetch < now → included
- Test next_fetch > now → excluded
- Test inactive feeds excluded
- Test ordering by next_fetch

---

#### 4.3 UpdateNextFetch Method
**File**: `pkg/repository/repository.go`

**Implementation**:
```go
// UpdateNextFetch sets when feed should be fetched next
func (r *Repository) UpdateNextFetch(feedID int64, nextFetch time.Time, interval int) error {
    _, err := r.db.Exec(`
        UPDATE feeds
        SET next_fetch = ?,
            fetch_interval = ?,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = ?
    `, nextFetch.Format(time.RFC3339), interval, feedID)

    if err != nil {
        return fmt.Errorf("update next fetch: %w", err)
    }

    return nil
}
```

**Testing**:
- Update existing feed → success
- Update non-existent feed → no error (0 rows)
- Verify values stored correctly
- Test with various intervals

---

#### 4.4 Adaptive Interval Calculator
**File**: `pkg/scheduler/scheduler.go` (new package)

**Implementation**:
```go
package scheduler

import (
    "time"
    "math"
)

const (
    MinInterval     = 15 * time.Minute  // Don't hammer servers
    MaxInterval     = 7 * 24 * time.Hour  // Check dormant feeds weekly
    DefaultInterval = 1 * time.Hour       // For new feeds
)

// Config controls scheduling behavior
type Config struct {
    MinInterval          time.Duration
    MaxInterval          time.Duration
    DefaultInterval      time.Duration
    AdaptiveScheduling   bool
    RespectCacheControl  bool
    JitterPercent        int
}

// CalculateNextFetch determines when to fetch feed next
func CalculateNextFetch(
    config Config,
    lastFetched time.Time,
    lastUpdated time.Time,
    fetchErrorCount int,
    cacheControlMaxAge int,
) (nextFetch time.Time, interval time.Duration) {

    // Start with default
    interval = config.DefaultInterval

    // 1. Exponential backoff for errors
    if fetchErrorCount > 0 {
        backoff := time.Duration(math.Pow(2, float64(fetchErrorCount))) * time.Hour
        interval = backoff
    }

    // 2. Adaptive based on update frequency (if enabled)
    if config.AdaptiveScheduling && fetchErrorCount == 0 {
        // If feed hasn't been updated in a while, slow down
        timeSinceUpdate := time.Since(lastUpdated)

        if timeSinceUpdate > 7*24*time.Hour {
            // Dormant feed (no updates in week) → check daily
            interval = 24 * time.Hour
        } else if timeSinceUpdate > 24*time.Hour {
            // Slow feed (daily updates) → check every 4 hours
            interval = 4 * time.Hour
        } else if timeSinceUpdate < time.Hour {
            // Fast feed (hourly updates) → check every 30 min
            interval = 30 * time.Minute
        }
        // else: default 1 hour (multiple times per day)
    }

    // 3. Respect Cache-Control max-age (if enabled)
    if config.RespectCacheControl && cacheControlMaxAge > 0 {
        cacheControlDuration := time.Duration(cacheControlMaxAge) * time.Second
        if cacheControlDuration > interval {
            interval = cacheControlDuration
        }
    }

    // 4. Enforce bounds
    if interval < config.MinInterval {
        interval = config.MinInterval
    }
    if interval > config.MaxInterval {
        interval = config.MaxInterval
    }

    // 5. Add jitter (±10% randomness)
    if config.JitterPercent > 0 {
        jitterRange := float64(interval) * float64(config.JitterPercent) / 100.0
        jitter := time.Duration((rand.Float64()*2-1) * jitterRange)
        interval += jitter
    }

    // Calculate next fetch time
    nextFetch = time.Now().Add(interval)

    return nextFetch, interval
}
```

**Testing**:
- Test error backoff: 1 error → 2h, 2 errors → 4h, 3 errors → 8h
- Test adaptive: dormant → 24h, slow → 4h, fast → 30min
- Test cache-control: max-age=7200 → 2h minimum
- Test jitter: results vary but within bounds
- Test bounds enforcement

---

#### 4.5 Integrate Scheduling in fetchFeeds
**File**: `cmd/rp/commands.go`

**Current**:
```go
feeds, err := repo.GetFeeds(true)  // All active feeds
```

**Updated**:
```go
// Only fetch feeds that are due
feeds, err := repo.GetFeedsReadyForFetch(time.Now())

// After successful fetch, calculate next fetch time
nextFetch, interval := scheduler.CalculateNextFetch(
    schedulerConfig,
    time.Now(),
    lastUpdatedFromFeed,  // Extract from feed metadata
    feed.FetchErrorCount,
    cacheControlMaxAge,   // Parse from response headers
)

err = repo.UpdateNextFetch(feed.ID, nextFetch, int(interval.Seconds()))
```

**Configuration**:
```ini
[planet]
min_fetch_interval = 900      # 15 minutes
max_fetch_interval = 604800   # 7 days
default_fetch_interval = 3600 # 1 hour
adaptive_scheduling = true
respect_cache_control = true
jitter_percent = 10
```

**Testing**:
- Integration test: only due feeds fetched
- Integration test: intervals adjust over time
- Integration test: disabled scheduling fetches all
- Test with various config values

---

#### 4.6 Parse Cache-Control Headers
**File**: `pkg/crawler/crawler.go`

**Implementation**:
```go
// ParseCacheControl extracts max-age from Cache-Control header
func ParseCacheControl(cacheControl string) (maxAge int) {
    // Parse: "max-age=3600, public"
    parts := strings.Split(cacheControl, ",")
    for _, part := range parts {
        part = strings.TrimSpace(part)
        if strings.HasPrefix(part, "max-age=") {
            ageStr := strings.TrimPrefix(part, "max-age=")
            age, err := strconv.Atoi(ageStr)
            if err == nil && age > 0 {
                return age
            }
        }
    }
    return 0  // No max-age or invalid
}

// Add to FeedResponse
type FeedResponse struct {
    // ... existing fields
    CacheControlMaxAge int  // Seconds from Cache-Control: max-age=N
}

// Populate in Fetch()
resp.CacheControlMaxAge = ParseCacheControl(httpResp.Header.Get("Cache-Control"))
```

**Testing**:
- Test "max-age=3600" → 3600
- Test "max-age=3600, public" → 3600
- Test "public, max-age=7200" → 7200
- Test no Cache-Control → 0
- Test invalid value → 0

---

#### 4.7 Migration for Existing Feeds
**File**: `cmd/rp/commands.go` (in cmdUpdate or new migration command)

**Implementation**:
```go
// InitializeScheduling sets initial next_fetch for existing feeds
func InitializeScheduling(repo *repository.Repository) error {
    feeds, err := repo.GetFeeds(true)
    if err != nil {
        return err
    }

    now := time.Now()

    for _, feed := range feeds {
        // Skip if already scheduled
        if !feed.NextFetch.IsZero() {
            continue
        }

        // Set next_fetch = now (fetch immediately on first run)
        // Set fetch_interval = 3600 (1 hour default)
        err := repo.UpdateNextFetch(feed.ID, now, 3600)
        if err != nil {
            return fmt.Errorf("initialize feed %d: %w", feed.ID, err)
        }
    }

    return nil
}
```

**Run once**: Automatically in cmdUpdate if any feed has NULL next_fetch

**Testing**:
- Test all NULL next_fetch → initialized
- Test already scheduled → unchanged
- Test new feed added → gets default

---

#### 4.8 Add rp status --show-schedule Command
**File**: `cmd/rp/commands.go`

**Implementation**:
```go
// cmdStatus with --show-schedule flag
func cmdStatus(opts StatusOptions) error {
    // ... existing status output

    if opts.ShowSchedule {
        feeds, err := repo.GetFeeds(true)
        if err != nil {
            return err
        }

        fmt.Fprintln(opts.Output, "\nFeed Schedule:")
        fmt.Fprintln(opts.Output, "----------------------------------------")

        for _, feed := range feeds {
            nextFetch := "not scheduled"
            if !feed.NextFetch.IsZero() {
                nextFetch = feed.NextFetch.Format("2006-01-02 15:04:05")

                // Show relative time
                untilNext := time.Until(feed.NextFetch)
                if untilNext > 0 {
                    nextFetch += fmt.Sprintf(" (in %s)", untilNext.Round(time.Minute))
                } else {
                    nextFetch += " (overdue)"
                }
            }

            interval := time.Duration(feed.FetchInterval) * time.Second

            fmt.Fprintf(opts.Output, "  %s\n", feed.Title)
            fmt.Fprintf(opts.Output, "    Next fetch: %s\n", nextFetch)
            fmt.Fprintf(opts.Output, "    Interval:   %s\n", interval)
            fmt.Fprintf(opts.Output, "    Errors:     %d\n\n", feed.FetchErrorCount)
        }
    }

    return nil
}
```

**Usage**:
```bash
rp status --show-schedule
```

**Output**:
```
Feed Schedule:
----------------------------------------
  Go Blog
    Next fetch: 2025-10-29 15:30:00 (in 15m)
    Interval:   1h0m0s
    Errors:     0

  GitHub Blog
    Next fetch: 2025-10-29 16:00:00 (in 45m)
    Interval:   1h0m0s
    Errors:     0
```

**Testing**:
- Test output format
- Test overdue feeds
- Test future feeds
- Test relative time display

---

### Phase 5: Testing & Documentation 📝
**Effort**: 1 day | **Dependencies**: Phases 1-4

#### 5.1 Comprehensive Testing
- Unit tests for all new functions
- Integration tests for end-to-end flows
- Benchmark tests for connection pooling
- Race detector on all tests
- Coverage verification (>75% all packages)

#### 5.2 Documentation Updates
- README.md: New config options
- CLAUDE.md: Architecture updates
- WORKFLOWS.md: New commands
- CHANGELOG.md: All changes listed
- specs/TODO.md: Mark features complete

#### 5.3 Performance Benchmarks
```bash
# Before and after measurements
make bench

# Expected improvements:
# - Connection pooling: 10-20% faster
# - Rate limiting: <5% overhead
# - Smart scheduling: 50-80% fewer fetches
```

---

## Implementation Timeline

### Quick Path (Minimum Viable)
**Total**: 3 days
1. Phase 1 (Foundation): 1 day
2. Phase 2 (301 Redirects): 0.5 days
3. Phase 3 (Rate Limiting): 1 day
4. Phase 5 (Testing): 0.5 days

**Result**: Connection pooling, retries, 301 handling, rate limiting

---

### Full Path (All Features)
**Total**: 7-8 days
1. Phase 1 (Foundation): 1 day
2. Phase 2 (301 Redirects): 0.5 days
3. Phase 3 (Rate Limiting): 1 day
4. Phase 4 (Scheduling): 3-4 days
5. Phase 5 (Testing & Docs): 1 day

**Result**: All networking features complete

---

## Dependencies & Order

**Must Do First**:
1. Phase 1.1 (Connection Pooling) - No dependencies
2. Phase 1.2 (Use FetchWithRetry) - No dependencies

**Can Do in Parallel**:
- Phase 2 (301 Redirects) - Independent of rate limiting
- Phase 3 (Rate Limiting) - Independent of redirects

**Must Do Last**:
- Phase 4 (Scheduling) - Depends on Phases 1-3 complete
- Phase 5 (Testing) - After all features

---

## Configuration Summary

**New config.ini options**:
```ini
[planet]
# Connection pooling (Phase 1)
max_idle_connections = 100
max_connections_per_host = 10
connection_timeout = 90

# 301 Redirects (Phase 2)
auto_update_redirects = true

# Rate limiting (Phase 3)
enable_rate_limiting = true
rate_limit_per_second = 1.0
rate_limit_burst = 3

# Intelligent scheduling (Phase 4)
min_fetch_interval = 900       # 15 minutes
max_fetch_interval = 604800    # 7 days
default_fetch_interval = 3600  # 1 hour
adaptive_scheduling = true
respect_cache_control = true
jitter_percent = 10
```

---

## Success Metrics

### Phase 1 (Foundation)
- ✅ HTTP connections reused (visible in logs/metrics)
- ✅ 10-20% faster fetch times (benchmark)
- ✅ FetchWithRetry used in production
- ✅ Transient failures recovered automatically

### Phase 2 (301 Redirects)
- ✅ 301 redirects detected and logged
- ✅ Feed URLs updated in database
- ✅ No extra HTTP hop on subsequent fetches

### Phase 3 (Rate Limiting)
- ✅ Same domain fetches delayed appropriately
- ✅ Different domains not blocked
- ✅ Configurable per-domain limits
- ✅ <5% performance overhead

### Phase 4 (Scheduling)
- ✅ Only due feeds fetched (not all every time)
- ✅ Fast feeds checked more often
- ✅ Slow feeds checked less often
- ✅ Failing feeds backed off exponentially
- ✅ 50-80% reduction in unnecessary fetches
- ✅ Cache-Control headers respected
- ✅ rp status --show-schedule works

---

## Risk Assessment

### Low Risk (Phases 1-2)
- Connection pooling: Well-understood, standard practice
- FetchWithRetry: Already implemented and tested
- 301 redirects: Simple database update

### Medium Risk (Phase 3)
- Rate limiting: New dependency (golang.org/x/time/rate)
- Per-domain tracking: Need thread-safe map
- Configuration complexity: More options to test

### High Risk (Phase 4)
- Scheduling logic: Complex with many edge cases
- Adaptive intervals: Requires tuning and validation
- Migration: Existing feeds need initialization
- User impact: Change in fetch behavior (breaking change if not careful)

---

## Rollout Strategy

### v0.5.0: Foundation
- Phase 1 (Connection pooling, retries)
- Phase 2 (301 redirects)
- Low risk, immediate benefits

### v0.6.0: Rate Limiting
- Phase 3 (Rate limiting)
- Optional feature (can be disabled)
- Validates approach before scheduling

### v1.0.0: Intelligent Scheduling
- Phase 4 (Scheduling)
- Major feature, needs thorough testing
- Requires v0.5.0 and v0.6.0 stable

---

## Open Questions

1. **Rate limit default**: 1 req/sec too conservative? Test with popular feeds.
2. **Scheduling UI**: Show next fetch times in rp list-feeds?
3. **Manual override**: Add rp fetch --force to ignore schedule?
4. **Metrics**: Track cache hits, rate limit delays, schedule efficiency?
5. **Error handling**: How many consecutive 301s before warning user?

---

## References

- **v0.4.0-plan.md**: P3.2 (301 redirects), P3.4 (rate limiting), P4.3 (retry), P5.1 (pooling)
- **v1.0.0-plan.md**: Section 2 (intelligent scheduling)
- **CLAUDE.md**: Lines 119, 185 (dependencies), 188 (redirects)
- **pkg/crawler/crawler.go**: Lines 266-310 (FetchWithRetry), 53 (FinalURL)
- **pkg/repository/repository.go**: Lines 256+ (next_fetch schema)

---

**Status**: Ready for implementation after v0.4.0 P1-P2 complete
**Next Steps**: Review with stakeholders, prioritize phases, begin Phase 1
