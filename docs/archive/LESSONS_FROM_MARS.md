# Lessons from Planet Mars

This document analyzes **two different tools both named "Planet Mars"** to extract lessons for Rogue Planet's development.

## IMPORTANT: Two Different Tools

There are **two completely separate projects** both called "Planet Mars":

### 1. Sam Ruby's Mars (2007-2008)
- **Language**: Ruby
- **Purpose**: Short-lived experimental Ruby rewrite while Venus was being developed
- **Outcome**: Abandoned after ~1 year; Venus (Python) continued as the main project
- **Architecture**: Sophisticated pipeline with 8+ stages (Fido → Spider → Harvest → Transmogrify → Sift → Splice → Formatter)
- **Status**: Abandoned in 2008
- **Timeline**: Blog post Dec 19, 2007 → GitHub repo created Apr 4, 2008 → Last change 2008
- **Repository**: https://github.com/rubys/mars
- **Key Innovation**: Clean architectural separation of concerns (later incorporated into Venus)

### 2. Rob Galanakis's Mars (2013-2014)
- **Language**: Python
- **Purpose**: "Ruthlessly simple" alternative to the now-complex Venus
- **Outcome**: Standalone simple aggregator for common use cases
- **Architecture**: Minimal pipeline (Config → Fetch → Parse → Cache → Render)
- **Status**: Maintained for simplicity
- **Repository**: https://github.com/rgalanakis/planet-mars
- **Philosophy**: "Worse is better" - deliberately limited features

## Executive Summary

**These are NOT the same tool.** They taught us different lessons:

- **Ruby Mars (Sam Ruby, 2007-2008)** → Short-lived experiment that explored clean architecture patterns; ideas incorporated into Venus
- **Python Mars (Rob Galanakis, 2013-2014)** → Taught us the value of simplicity, the dangers of feature creep, when to limit scope

**IMPORTANT CHRONOLOGY**:
- **~2006**: Sam Ruby begins Venus (Python) as refactoring of Planet 2.0
- **Dec 2007**: Mars (Ruby) experiment announced as potential Venus alternative
- **2008**: Mars abandoned; Venus continues as primary project
- **2010**: Venus GitHub repository created (project already active for ~4 years)
- **2013-2014**: Rob Galanakis creates different "Mars" as simpler alternative to now-complex Venus

This analysis extracts lessons from both implementations.

## Why Planet Mars Was Created

### The Problem with Venus

By 2013, Planet Venus had become:
- Overly complex for common use cases
- Difficult to understand and modify
- Heavyweight for simple aggregation needs
- Hard to install and configure

### Mars's Philosophy

> "My goal for Planet Mars is to make it as simple as possible, to be the best aggregator for the very common use case."
> — Rob Galanakis

**Core Principles**:
- **Ruthless simplicity** over comprehensive features
- **Easy to understand** codebase
- **Quick to set up** and deploy
- **Good enough** for 90% of use cases

This represents the "worse is better" philosophy applied to feed aggregation.

## Architecture Comparison

### Technology Stack

| Component | Mars (Python) | Mars (Ruby) | Rogue Planet |
|-----------|---------------|-------------|--------------|
| Language | Python 2.6+ | Ruby | Go |
| Feed Parsing | feedparser | FeedNormalizer | gofeed |
| Templates | Jinja2, htmltmpl | HAML, XSLT | html/template |
| Storage | File cache | File cache | SQLite |
| Concurrency | ThreadPool | Threads | Goroutines |
| HTTP Library | urllib2 | Net::HTTP | net/http |

### Pipeline Architecture

**Ruby Mars** (more sophisticated):
```
Config → Fido → Spider → Harvest → Transmogrify → Sift → Splice → Formatter
         (fetch) (crawl)  (parse)   (normalize)   (filter) (merge)  (render)
```

**Python Mars** (simpler):
```
Config → Planet.run() → Feed Fetcher → Parser → Normalizer → Cache → Renderer
```

**Rogue Planet** (balanced):
```
Config → Crawler → Normalizer → Repository → Generator
         (fetch)   (parse+sanitize) (SQLite)    (render)
```

## What Mars Got Right

### 1. HTTP Conditional Requests (Ruby Mars)

**Outstanding implementation** in `fido.rb`:

```ruby
headers['If-None-Match'] = cached['etag'] if cached['etag']
headers['If-Modified-Since'] = cached['modified'] if cached['modified']

# Handles 304 Not Modified correctly
# Caches 200, 301, 410 responses
```

**Why This Is Gold Standard**:
- Stores ETags and Last-Modified exactly as received
- Sends both headers on subsequent requests
- Properly handles 304 responses (uses cached content)
- Caches headers for future requests
- Respects HTTP semantics perfectly

**Rogue Planet Status**: ✅ **Already implemented** in pkg/crawler/crawler.go:233-235

**Reference**: specs/rogue-planet-spec.md lines 64-156

### 2. Feed Format Normalization

**Brilliant approach** in Ruby Mars's `transmogrify.rb`:

Transform all formats (RSS 1.0, RSS 2.0, Atom) to canonical Atom format **before** processing:

```
RSS 2.0 → Atom:
  <item> → <entry>
  <description> → <summary>
  <pubDate> → <published>
```

**Why This Is Smart**:
- Single code path for all feed formats
- Normalization at input boundary
- Downstream code never needs to know original format
- Easy to add new formats (just write one transformer)

**Rogue Planet Status**: ✅ **Already implemented** via gofeed in pkg/normalizer/

**Lesson**: Mars validates Rogue Planet's design choice.

### 3. Parallel Feed Fetching

**Simple, effective** in Python Mars:

```python
from multiprocessing.pool import ThreadPool

if config.get('spider_threads'):
    pool = ThreadPool(int(config['spider_threads']))
    pool.map(channel.update_entries, channels)
```

**Documentation notes**: "Can have a significant effect" on performance.

**Why This Works**:
- Opt-in complexity (defaults to sequential)
- Graceful fallback if threading unavailable
- No complex coordination needed

**Rogue Planet Status**: ✅ **Already implemented** with goroutine worker pool (more sophisticated)

**Lesson**: All three implementations recognize concurrent fetching is essential.

### 4. Intelligent Date Normalization

**Sophisticated fallback chain** in Python Mars:

```python
# Try multiple date sources in order:
1. entry.published_parsed
2. entry.updated_parsed
3. channel.updated_parsed
4. current_fetch_time
```

**Why Essential**: Real-world feeds have inconsistent date handling.

**Rogue Planet Status**: ✅ **Already implemented** in pkg/normalizer/normalizer.go lines 150-179

**Lesson**: Validates Rogue Planet's approach is industry best practice.

### 5. Configuration Simplicity

**Mars format**: Simple INI with feed URLs as section headers.

**Key Insight**: Most users need:
- Feed URL
- Optional feed-specific settings
- Simple is better than comprehensive

**Rogue Planet Status**: ✅ **Already supported** via pkg/config/

**Lesson**: Both simple feeds.txt and extended config.ini are valuable.

### 6. The "Ruthlessly Simple" Philosophy

**Mars Principle**: Deliberately limit features for maintainability.

**Application**:
- Don't add features just because Venus has them
- Keep CLI focused on core use cases
- Resist feature creep
- Document what you *won't* do

**Rogue Planet Status**: ✅ **Already following** - simple CLI, focused features

**Lesson**: Simplicity is a feature, not a limitation. Guard against scope creep.

## What Mars Got Wrong

### 1. File-Based Caching Is Insufficient

**Mars Limitation**: File-based cache, no relational queries.

**Problems**:
- Can't efficiently query by date range
- Can't count entries per feed
- Can't implement `prune --days N` efficiently
- Hard to detect duplicates across feeds
- No atomic operations
- Doesn't scale beyond ~100 feeds

**Rogue Planet Solution**: ✅ SQLite with proper indexes

**Reference**: pkg/repository/repository.go lines 63, 69, 267-314

**Lesson**: SQLite hits the sweet spot - simple deployment, powerful queries, scales well.

### 2. No CLI Commands for Management

**Mars Limitation**: Only one command - run the aggregator.

**Missing**:
- No `add-feed` (must edit config file manually)
- No `remove-feed`
- No `list-feeds` to see configuration
- No `status` to check health
- No `prune` to clean old entries
- No `verify` to test configuration

**Rogue Planet Solution**: ✅ 13 CLI subcommands

**Commands**: init, add-feed, add-all, remove-feed, list-feeds, status, update, fetch, generate, prune, import-opml, export-opml, verify

**Lesson**: Rich CLI dramatically improves UX. Manual config editing is error-prone.

### 3. No Intelligent Scheduling

**Mars Behavior**: Fetches all feeds every run, relies on HTTP 304.

**Problems**:
- Still makes HTTP request even if feed hasn't changed
- No backoff for failing feeds
- No per-feed fetch intervals
- Wastes bandwidth and server resources

**Rogue Planet Solution**: ✅ Database tracks `fetch_error_count` and `next_fetch`

**Implementation**: Exponential backoff for failing feeds

**Reference**: specs/rogue-planet-spec.md line 290

**Lesson**: Intelligent scheduling is essential for production use.

### 4. Limited Error Handling

**Mars Approach**: Log errors but no persistent tracking.

**Problems**:
- No visibility into which feeds are failing
- No automatic disabling of dead feeds
- No exponential backoff for failures
- Hard to debug feed issues

**Rogue Planet Solution**: ✅ Database stores `fetch_error` and `fetch_error_count`

**Reference**: pkg/repository/repository.go

**Lesson**: Persistent error state enables better observability and reliability.

### 5. No OPML Import/Export

**Mars Status**: Python version mentions OPML but unclear if implemented.

**Why This Matters**: OPML is the standard for feed list interchange.

**Use Cases**:
- Import from other aggregators (Feedly, Inoreader, etc.)
- Export for backup or migration
- Share curated feed lists

**Rogue Planet Solution**: ✅ Full OPML 1.0/2.0 support

**Commands**: `rp import-opml`, `rp export-opml`

**Reference**: pkg/opml/ (91.8% test coverage)

**Lesson**: OPML support is table stakes for modern feed aggregators.

### 6. Minimal Testing

**Mars Status**: Few tests visible in both implementations.

**Problems**:
- Hard to refactor safely
- No regression protection
- Difficult to verify feed format handling
- Unclear test coverage

**Rogue Planet Solution**: ✅ Comprehensive test suite

**Coverage**:
- Unit tests per package (>75% coverage)
- Integration tests for full workflows
- Real-world feed snapshots in testdata/
- Network tests with build tags
- Race detector tests

**Reference**: TESTING.md, Makefile

**Lesson**: Tests are investment in maintainability. Worth the effort.

## Security Comparison

| Feature | Mars | Rogue Planet | Winner |
|---------|------|--------------|--------|
| **HTML Sanitization** | ⚠️ Unclear (Python), ✅ Explicit (Ruby) | ✅ bluemonday | Rogue Planet |
| **SSRF Prevention** | ❌ None visible | ✅ URL validation | Rogue Planet |
| **XSS Prevention** | ⚠️ Template escaping only | ✅ Sanitize + escape | Rogue Planet |
| **CSP Headers** | ❌ No | ✅ Yes | Rogue Planet |
| **SQL Injection** | N/A (no database) | ✅ Prepared statements | Rogue Planet |

**Rogue Planet Security**:
- Sanitizes HTML at input (bluemonday)
- Escapes at output (html/template)
- Validates URLs before fetching (blocks localhost, private IPs)
- CSP headers in generated HTML
- Prepared statements for all SQL

**Reference**: specs/rogue-planet-spec.md lines 548-762

**Lesson**: Mars was created before SSRF and some XSS vectors were widely known. Modern aggregators need defense-in-depth.

## Feature Comparison

| Feature | Mars | Rogue Planet | Advantage |
|---------|------|--------------|-----------|
| HTTP Conditional Requests | ✅ (Ruby) | ✅ | Tie |
| Feed Formats | RSS, Atom | RSS, Atom, JSON Feed | Rogue Planet |
| HTML Sanitization | ⚠️ Partial | ✅ Comprehensive | Rogue Planet |
| Storage | File cache | SQLite | Rogue Planet |
| Parallel Fetching | ✅ | ✅ | Tie |
| CLI Commands | 1 | 13 | Rogue Planet |
| OPML Support | ❌/⚠️ | ✅ | Rogue Planet |
| Feed Scheduling | ❌ | ✅ | Rogue Planet |
| Error Tracking | ❌ | ✅ | Rogue Planet |
| Template Engines | Multiple | Go templates | Mars (variety) |
| Testing | Minimal | Comprehensive | Rogue Planet |
| Documentation | Basic | Extensive | Rogue Planet |
| Date Normalization | ✅ | ✅ | Tie |
| Offline Mode | ✅ (Python) | ❌ | Mars |
| Multiple Outputs | ✅ | ❌ | Mars |

## Recommendations for Rogue Planet

### High Priority - Consider Adding

#### 1. Prevent Entry Spam on New Feed Addition

**From**: Reddit complaint about Venus/PlanetPlanet behavior

**Problem**: When you add a new feed, all historical entries (50-100) appear at once based on their original published dates, spamming the planet's timeline and RSS feed. This was a **major complaint** that drove users away from Venus.

**Use Case**: Clean feed addition without polluting timeline with old content.

**Implementation**:
```ini
[planet]
filter_by_first_seen = true   # Only show entries first seen within time window
sort_by = "first_seen"         # Sort by when aggregator first saw entry
```

**Technical Details**:
- Database already has `first_seen` field ✅
- Index on `first_seen` exists ✅
- Need to update `GetRecentEntries()` to filter/sort by `first_seen` instead of `published`

**Effort**: Low - modify one SQL query based on config.

**Value**: **VERY HIGH** - addresses the #1 user complaint about Venus/Planet.

**Reference**: See WISHLIST.md for complete implementation details.

#### 2. Offline Mode

**From**: Python Mars

**Feature**: Regenerate HTML from cache without network requests.

**Use Case**: Development, testing, bandwidth-limited environments.

**Implementation**:
```bash
rp generate --offline  # Use cached data, no network
```

**Effort**: Low - just skip fetch step, use existing database entries.

**Value**: Medium - useful for development and debugging.

#### 2. Multiple Output Formats

**From**: Both Mars implementations

**Feature**: Generate multiple formats in one run (HTML + RSS + Atom + OPML).

**Use Case**: Provide feeds-of-feeds, multiple presentation formats.

**Implementation**:
```bash
rp generate --format=html,rss,atom
```

**Effort**: Medium - need RSS/Atom feed templates.

**Value**: Medium - some users want to provide feed output.

#### 3. Template Debugging

**From**: Mars verbose logging

**Feature**: Show template variables during generation.

**Implementation**:
```bash
rp generate --debug  # Show all template data
```

**Effort**: Low - add debug printing.

**Value**: High - makes template development much easier.

### Medium Priority - Future Enhancements

#### 4. Feed Health Dashboard

**Feature**: Track fetch success rate over time, not just last error.

**Implementation**:
```bash
rp status --detailed  # Show per-feed health metrics
```

**Effort**: Medium - needs historical tracking in database.

**Value**: Medium - better observability for planet operators.

#### 5. Feed URL Update on 301

**From**: Ruby Mars's fido.rb

**Feature**: Automatically update feed URL in database on permanent redirect.

**Effort**: Low - detect 301, update database.

**Value**: High - prevents unnecessary redirects on every fetch.

#### 6. Per-Feed Fetch Intervals

**Feature**: Different update frequencies per feed.

**Implementation**:
```ini
[https://example.com/feed.xml]
fetch_interval = 6h  # Slow-updating blog

[https://news.com/feed.xml]
fetch_interval = 15m  # Breaking news
```

**Effort**: Medium - add per-feed scheduling logic.

**Value**: High - reduces load, respects publisher update patterns.

### Low Priority - Nice to Have

#### 7. Feed Filtering

**From**: Mars's sift.rb shows explicit filtering stage.

**Feature**: Filter entries per feed.

**Implementation**:
```ini
[https://example.com/feed.xml]
max_entries = 10
exclude_title = "regex pattern"
```

**Effort**: Medium - add filtering logic to normalizer.

**Value**: Low - niche use case.

#### 8. Multiple Planets from One Database

**Feature**: Run multiple planets from same Rogue Planet installation.

**Implementation**:
```bash
rp generate --config=planet1.ini --output=/var/www/planet1
rp generate --config=planet2.ini --output=/var/www/planet2
```

**Effort**: Low - already mostly supported.

**Value**: Low - users can run multiple instances.

### Do NOT Implement

#### ❌ Multiple Template Engines

**Reason**: Go templates are sufficient. Adding Jinja2 would require embedding Python or complex reimplementation. Not worth complexity.

#### ❌ Web UI

**Reason**: Against "static output" philosophy. Adds attack surface, complexity, maintenance burden.

#### ❌ Database Migration to PostgreSQL/MySQL

**Reason**: SQLite is perfect for this use case. Other databases add deployment complexity with no benefit.

#### ❌ Feed Content Modification

**Reason**: Aggregator should aggregate, not modify. Content transformation belongs elsewhere.

## Key Architectural Insights

### 1. The Database Decision

**Mars**: No database, file-based cache.
**Rationale**: Zero dependencies, simple deployment, easy to inspect.

**Rogue Planet**: SQLite database with indexes.
**Rationale**: Efficient queries, scales better, enables advanced features.

**Verdict**: For 2025, SQLite is the right choice. It's as ubiquitous as files but far more powerful.

**Lesson**: Sometimes "simpler" (no database) is actually more complex (implement your own indexing, locking, queries). SQLite hits the sweet spot.

### 2. Single vs. Multiple Template Engines

**Mars**: Supports Jinja2, HAML, htmltmpl, XSLT.
**Rationale**: Flexibility attracts more users.

**Rogue Planet**: Only Go html/template.
**Rationale**: Consistent, auto-escaping, zero dependencies.

**Verdict**: For a single-purpose tool, one well-chosen engine is enough.

**Lesson**: Premature generalization adds complexity. Go templates have proven sufficient for Hugo and other projects.

### 3. Concurrency Models

**Mars**: Optional ThreadPool (Python), threaded (Ruby).
**Rogue Planet**: Always concurrent via goroutines.

**Insight**: All recognize parallel fetching is essential. Goroutines are so lightweight that "opt-in" concurrency isn't necessary in Go like it is in Python.

**Lesson**: Leverage your language's strengths. Go's concurrency primitives make parallelism nearly free.

### 4. Command-Line Interface

**Mars**: Single command that does everything.
**Rogue Planet**: 13 subcommands for different operations.

**Trade-off**:
- Mars: Simpler mental model
- Rogue Planet: More flexible, better for automation

**Verdict**: Rich CLI wins for production use. Automation requires separation (fetch vs. generate).

**Lesson**: "Unix philosophy" of single-purpose tools applies even within an application.

## Conclusion

### What Planet Mars Taught Us

Planet Mars validated many of Rogue Planet's design decisions while showing pitfalls to avoid:

**Validated Decisions** (Keep Doing):
- ✅ HTTP conditional requests are essential
- ✅ Feed format normalization at input boundary
- ✅ Parallel fetching is mandatory for performance
- ✅ Date handling needs robust fallback chain
- ✅ Simple configuration is better than complex
- ✅ Ruthless simplicity prevents feature creep

**Avoided Mistakes** (Different Path):
- ✅ SQLite better than file cache
- ✅ Rich CLI better than single command
- ✅ Persistent error tracking is essential
- ✅ OPML support is table stakes
- ✅ Comprehensive testing pays off
- ✅ Modern security (SSRF, XSS, CSP) is mandatory

**Consider Adding** (From Mars):
- ⚠️ Offline mode for development
- ⚠️ Multiple output formats
- ⚠️ Template debugging
- ⚠️ Feed URL updates on 301 redirects

### Rogue Planet's Synthesis

Rogue Planet stands on 20+ years of Planet evolution:

```
Planet (2002) → Venus (2006) → Mars Ruby (2007-08) → Mars Python (2013) → Rogue Planet (2025)
   ↓               ↓               ↓ (abandoned)        ↓                      ↓
Simple         Powerful        Experiment          Ruthlessly Simple      Best of All
File cache     File cache      File cache          File cache             SQLite
Python 2       Python 2        Ruby                Python 2               Go
No tests       Some tests      Few tests           Minimal tests          Extensive tests
No OPML        OPML           No OPML              Unclear OPML           OPML + more
Basic CLI      Moderate CLI    Minimal CLI         Single command         Rich CLI
```

**The Synthesis**:
- Learned from Venus's comprehensive features
- Adopted Mars's simplicity philosophy
- Added modern security practices
- Built on Go's strengths (concurrency, single binary, strong typing)
- Comprehensive testing from day one
- Rich CLI for production use

### Final Assessment

**Sam Ruby's Mars (2007-2008)** was a valuable experiment that explored Ruby-based architecture patterns, even though it was ultimately abandoned in favor of continuing Venus development in Python. Its clean pipeline architecture influenced Venus's evolution.

**Rob Galanakis's Mars (2013-2014)** was an important historical step showing that "simpler than Venus" was viable. Its file-based approach and minimal CLI worked for common use cases.

But for 2025, **Rogue Planet represents the evolution beyond both**:
- Keeps the simplicity philosophy from Python Mars
- Learns from Ruby Mars's architectural patterns
- Adopts Venus's comprehensive features where valuable
- Adds SQLite for scalability beyond file caching
- Provides rich CLI for production operations
- Implements modern security practices
- Comprehensive testing and documentation from day one

**Both Mars projects showed us what to avoid as much as what to embrace.** That makes them valuable teachers.

## References

### Source Code
- https://github.com/rubys/mars (Ruby implementation by Sam Ruby)
- https://github.com/rgalanakis/planet-mars (Python implementation by Rob Galanakis)

### Blog Posts
- https://intertwingly.net/blog/2007/12/19/Yet-Another-Planet-Refactoring (Sam Ruby, Dec 19, 2007 - Mars announcement)
- https://www.robg3d.com/2014/03/planet-mars-the-simple-feed-aggregator/ (Rob Galanakis, 2014)

### Timeline Sources
- **Sam Ruby's Mars**: Blog post Dec 19, 2007; GitHub repo created Apr 4, 2008; last change 2008
- **Planet Venus**: Project started ~2006; GitHub repo created May 13, 2010 (after 4 years of development)
- **Wikipedia**: Venus described as starting in 2006 as "radical refactoring of Planet 2.0"
- **GitHub metadata**: Verified via repository creation timestamps

### Rogue Planet References
- specs/rogue-planet-spec.md (architectural specification)
- CLAUDE.md (development guide)
- pkg/crawler/crawler.go (HTTP fetching implementation)
- pkg/normalizer/normalizer.go (feed normalization)
- pkg/repository/repository.go (SQLite storage)
- pkg/generator/generator.go (HTML generation)

---

**Document Version**: 1.1
**Date**: 2025-10-16
**Last Updated**: Corrected chronology - Venus preceded Mars Ruby experiment, not vice versa
