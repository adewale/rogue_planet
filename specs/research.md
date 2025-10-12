# Research: Feed Aggregator Best Practices and History

This document captures the research conducted to inform the Rogue Planet specification, including all source URLs and key findings.

## Table of Contents

- [Primary Resources](#primary-resources)
- [Feed Fetching Best Practices](#feed-fetching-best-practices)
- [Historical Implementations](#historical-implementations)
- [Security Research](#security-research)
- [Active Planet Sites](#active-planet-sites)
- [Key Findings Summary](#key-findings-summary)

---

## Primary Resources

### Venus (Planet 2.0) - The Foundation

**Main Project**:
- GitHub: https://github.com/rubys/venus
- Documentation: https://intertwingly.net/code/venus/docs/
- Normalization Docs: https://intertwingly.net/code/venus/docs/normalization.html

**Key Developer**: Sam Ruby (rubys@intertwingly.net)

**Background**:
Venus was a major refactoring of Planet 2.0, started by Sam Ruby in 2006. It introduced:
- Comprehensive feed normalization
- HTML5 parsing with html5lib
- Universal Feed Parser integration
- Extensive character encoding handling
- Template-based output generation

**Blog Posts About Venus Development**:
- Venus Rising (2006): https://intertwingly.net/blog/2006/08/16/Venus-Rising
  - Announced the initial Venus refactoring
  - Discussed architectural improvements over Planet 2.0
  
- Yet Another Planet Refactoring (2007): https://intertwingly.net/blog/2007/12/19/Yet-Another-Planet-Refactoring
  - Further improvements to the architecture
  - Lessons learned from production deployments

**Wikipedia Entry**:
- https://en.wikipedia.org/wiki/Planet_(software)
- Lists Venus as one of the major successors to original Planet

---

## Feed Fetching Best Practices

### Critical Resource: Rachel by the Bay

**Primary Article**: https://rachelbythebay.com/w/2024/05/27/feed/
- ⚠️ Note: This URL was not directly accessible during research due to robots.txt restrictions
- However, related content was found documenting feed reader issues

**Related Articles Found**:

1. **Feed Score Documentation**: https://rachelbythebay.com/fs/help.html
   - Detailed explanation of proper conditional request implementation
   - How to use ETag and Last-Modified headers correctly
   - Common mistakes made by feed readers

2. **Feeds, Updates, and HTTP Status Codes** (2023): https://rachelbythebay.com/w/2023/01/18/http/
   - Explains the 200/304/429 status code handling
   - Documents why poorly-behaved readers get rate limited
   - Technical details on implementing If-Modified-Since and If-None-Match

3. **Feedback on Feed Issues** (2023): https://rachelbythebay.com/w/2023/06/03/feed/
   - Real examples of broken feed readers
   - Readers fetching every 2 minutes without conditional headers
   - How servers respond with 429 Too Many Requests

4. **Attack of the Broken Feed Reader** (2018): https://rachelbythebay.com/w/2018/04/12/badfeed/
   - Example of Tiny Tiny RSS making 101 HTTP requests in 17 seconds
   - Not using If-Modified-Since, requesting full content every time
   - Fetching individual posts in addition to the feed

5. **More Feedback on Feeds** (2018): https://rachelbythebay.com/w/2018/05/14/feedback/
   - Issues with URL redirects (www. vs non-www)
   - Feed readers not updating their stored URLs after 301 redirects
   - Problems with protocol-relative URLs in feeds

**Key Lessons from Rachel's Documentation**:
- Always send If-Modified-Since with the exact Last-Modified value received
- Always send If-None-Match with the exact ETag value received (including quotes)
- Update cached headers on EVERY response, even if content seems unchanged
- Never hash body content as a shortcut - use server headers
- Some readers send fake dates like "Wed, 01 Jan 1800 00:00:00 GMT" - don't do this
- 304 Not Modified is your friend - it saves bandwidth for everyone
- Respect 429 Too Many Requests with proper backoff

---

## Historical Implementations

### 1. Planet (Original Implementation)

**Project**: The original Planet feed aggregator
- Language: Python
- Developers: Jeff Waugh and Scott James Remnant
- License: Python License
- Status: Still maintained but superseded by Venus

**Key Technologies**:
- Mark Pilgrim's Universal Feed Parser (RSS, Atom, RDF)
- Tomas Styblo's htmltmpl templating engine
- Static file generation

**Historical Significance**:
- First major "planet" style aggregator
- Established the pattern of community feed aggregation
- Used by many FLOSS communities (GNOME, KDE, Ubuntu, etc.)

---

### 2. Venus (Planet 2.0)

See [Primary Resources](#primary-resources) section above.

**Improvements Over Planet**:
- Better feed normalization
- HTML5 parsing support
- Character encoding detection and correction
- More robust error handling
- Plugin/filter system

---

### 3. Mars

**Project**: https://github.com/rubys/mars
- Language: Ruby
- Developer: Sam Ruby (same as Venus)
- Status: Experimental, not widely adopted

**Purpose**:
- Experimental rewrite exploring different architectural approaches
- Testing ideas that might be backported to Venus
- Ruby implementation for comparison with Python Venus

**Key Features Explored**:
- Different templating approaches
- Alternative database backends
- Simplified configuration

---

### 4. Moonmoon

**Project**: https://github.com/moonmoon/moonmoon
- Official Site: https://moonmoon.org/
- Language: PHP
- Philosophy: "Stupidly simple"
- Status: Active, last release 2016

**Key Characteristics**:
- Web-based administration interface (no config file editing)
- No database required (uses file caching)
- SimplePie library for feed parsing
- Minimal features by design: no archives, no comments, no voting
- Single page output
- PHP 5.6+ required (PHP 7+ recommended)

**Security Issues Addressed**:
- CSRF protection added (#98)
- Only allow fetching pre-configured feeds (#84)
- Regular SimplePie library updates

**Lessons from Moonmoon**:
- Simplicity is a feature, not a limitation
- Web UI lowers barrier to entry
- PHP deployment is easier for many shared hosting scenarios
- Active community despite minimal feature set

**Real-World Usage**:
- Planet Judo used Moonmoon but experienced performance issues with 49 feeds
- Multiple small communities adopted it for ease of deployment

---

### 5. Pluto (hackNY Version)

**Project**: https://github.com/ChimeraCoder/pluto
- README: https://github.com/ChimeraCoder/pluto?tab=readme-ov-file
- Language: Go
- Developer: ChimeraCoder (Adrian Chifor)
- Focus: Static generation, hackNY community

**Example Site**: https://github.com/ChimeraCoder/Planet-hackNY
- Demonstrates Pluto in production use
- hackNY community aggregator

**Significance**:
- Early Go implementation of planet-style aggregator
- Proof of concept that Go works well for this use case
- Static file generation similar to Venus

---

### 6. Pluto (Gerald Bauer Version)

**Resources**:
- Links collection: https://github.com/web-work-tools/awesome-planet-pluto
- Multiple related gems: pluto, pluto.more.tools, pluto.starter, etc.

**Project**: Planet Pluto static feed reader
- Language: Ruby
- Developer: Gerald Bauer
- Distribution: RubyGems

**Features**:
- Static website generator from web feeds
- Template system
- OPML, FOAF, CSV, and config file support
- Integration with Jekyll and other static site generators

**Ecosystem**:
- Multiple gems for different purposes
- planet.rb for Jekyll integration
- news.rb for quick feed aggregation scripts

**Example Usage**:
```ruby
# Simple planet.rb script
require 'open-uri'
require 'feedparser'

FEED_URLS = [
  'http://vienna-rb.at/atom.xml',
  'http://weblog.rubyonrails.org/feed/atom.xml'
]

items = []
FEED_URLS.each do |url|
  feed = FeedParser::Parser.parse(open(url).read)
  items += feed.items
end
```

---

### 7. Feedreader

**Project**: https://github.com/feedreader/
- Organization with multiple feed-related projects
- Appears to be inactive or minimal activity

**Research Note**: Limited information available, appears to be a collection point for feed-related tools rather than a single implementation.

---

### 8. Mercury

**Project**: https://github.com/kgaughan/mercury
- Developer: Keith Gaughan

**Research Note**: Repository details not fully accessible, appears to be another feed aggregator implementation but limited public information available.

---

## Security Research

### CVE-2009-2937: XSS Vulnerability in Planet/Venus

**Official CVE**: https://nvd.nist.gov/vuln/detail/CVE-2009-2937

**Vulnerability Details**:
- **Severity**: Cross-site scripting (XSS) vulnerability
- **Affected**: Planet 2.0 and Planet Venus
- **Vector**: IMG element SRC attribute in feed content
- **CVSS Score**: Medium severity
- **CWE**: CWE-79 (Improper Neutralization of Input During Web Page Generation)

**The Attack**:
```html
<img src="javascript:alert(1);" >
```

This malicious content in a feed could execute JavaScript in the browser of anyone viewing the aggregated page.

**What Went Wrong**:
- Planet/Venus attempted to sanitize HTML but failed to catch all JavaScript injection vectors
- IMG src attributes were not properly validated
- JavaScript protocol URLs were not blocked

**Debian Bug Reports**:
- Planet: https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=546178
- Venus: https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=546179

**Fix Timeline**:
- Reported: September 2009
- Fixed in Venus: bzr116 (September 2009)
- Fixed in Debian: October 2009 (stable and unstable)

**Lessons Learned**:
1. HTML sanitization is critical and difficult to get right
2. Whitelist approach is safer than blacklist
3. Use established sanitization libraries, don't roll your own
4. Test with actual XSS attack vectors
5. All user-provided content (feeds) must be treated as untrusted
6. Defense in depth: sanitize on input AND output

**Modern Mitigation**:
- Use libraries like bluemonday (Go) or Bleach (Python)
- Implement Content Security Policy headers
- Never allow javascript: protocol in any attribute
- Block event handlers (onclick, onerror, etc.)
- Use html/template auto-escaping in Go

---

## Active Planet Sites

These sites demonstrate Planet/Venus in production and provide examples of successful deployments:

### 1. Planet GNOME
- **URL**: https://planet.gnome.org/
- **Community**: GNOME Desktop Environment developers
- **Status**: Active, regularly updated
- **Technology**: Venus
- **Features**: Clean design, chronological posts, source attribution

### 2. Planet KDE
- **URL**: https://planet.kde.org/
- **Community**: KDE Desktop Environment developers
- **Status**: Active
- **Technology**: Venus/Planet
- **Features**: Simple, functional design focusing on content

### 3. Planet Mozilla
- **URL**: https://planet.mozilla.org/
- **Community**: Mozilla contributors and community
- **Status**: Active
- **Technology**: Planet/Venus
- **Features**: Large-scale deployment with many feeds

### 4. Planet Ubuntu
- **URL**: https://planet.ubuntu.com/
- **Community**: Ubuntu developers and community
- **Status**: Active
- **Technology**: Venus
- **Features**: High-traffic site, multiple feeds from Ubuntu community

### 5. Planet Fedora
- **URL**: https://fedoraplanet.org/
- **Community**: Fedora Linux project
- **Status**: Active
- **Technology**: Venus
- **Features**: Community-focused, regular updates

### 6. Planet Gentoo
- **URL**: https://planet.gentoo.org/
- **Community**: Gentoo Linux developers
- **Status**: Active
- **Technology**: Venus/Planet
- **Features**: Technical content from Gentoo community

### 7. Planet Intertwingly
- **URL**: http://planet.intertwingly.net/
- **Owner**: Sam Ruby (Venus creator)
- **Status**: Active, demo/test site
- **Technology**: Venus (latest features)
- **Significance**: Reference implementation showing Venus capabilities
- **Features**: Meme tracker, advanced Venus features

### 8. Planet Performance (PerfPlanet)
- **URL**: https://feed.perfplanet.com/
- **Community**: Web performance professionals
- **Status**: Active
- **Technology**: Feed aggregation
- **Features**: Focused on web performance topics

### Common Patterns in Successful Planet Sites:
1. **Community Focus**: All serve specific technical communities
2. **Simple Design**: Content over fancy features
3. **Attribution**: Always show source blog/author
4. **Chronological**: River of news format, newest first
5. **Static Output**: Fast loading, survives traffic spikes
6. **Regular Updates**: Automated fetching (hourly typical)
7. **Minimal Maintenance**: Set-and-forget operation

---

## Key Findings Summary

### Feed Fetching (Most Critical)

**Best Practices**:
1. **Always** use HTTP conditional requests (If-Modified-Since, If-None-Match)
2. Store ETag and Last-Modified **exactly** as received (don't modify them)
3. Handle 304 Not Modified correctly (don't re-download)
4. Handle 429 Too Many Requests with exponential backoff
5. Update cached headers on **every** response
6. Never fetch more often than necessary (1 hour minimum recommended)
7. Identify clearly in User-Agent with contact information
8. Follow 301 redirects and update stored URLs
9. Add jitter to avoid thundering herd problem
10. Implement per-domain rate limiting

**Common Mistakes to Avoid**:
- Hashing body content instead of using server headers
- Making up fake header values
- Not updating cache after receiving new headers
- Fetching too frequently (every minute/every 2 minutes)
- Not handling 304 responses
- Anonymous or misleading User-Agent strings
- Ignoring 429 rate limit responses

### Security (Critical)

**Must-Have Protections**:
1. **HTML Sanitization**: Use battle-tested library (bluemonday)
2. **XSS Prevention**: Remove javascript:, data: URLs, event handlers
3. **SSRF Prevention**: Validate URLs, block localhost/private IPs
4. **SQL Injection**: Always use parameterized queries
5. **Resource Limits**: Max response size, timeouts
6. **CSP Headers**: Add Content-Security-Policy to output
7. **Input Validation**: Validate and sanitize all feed content

**Attack Vectors to Test**:
- `<img src="javascript:alert(1)">`
- `<a href="javascript:alert(1)">click</a>`
- `<img src=x onerror="alert(1)">`
- `http://localhost/admin` (SSRF)
- SQL injection in feed titles/content
- Extremely large feeds (DoS)

### Architecture

**Proven Patterns**:
1. **Static Output**: Fast, scalable, survives traffic spikes
2. **Four Phases**: Fetch → Normalize → Store → Generate
3. **SQLite Database**: Perfect for this use case
4. **Separation of Concerns**: Each component independent
5. **Cron/Timer Based**: Not a long-running service
6. **Single Binary**: Easy deployment (Go advantage)

### Normalization

**Critical Tasks**:
1. Convert all feeds to canonical format (Atom-style)
2. Fix character encoding issues (output UTF-8)
3. Sanitize HTML content
4. Resolve relative URLs to absolute
5. Parse multiple date formats → RFC 3339
6. Generate IDs when missing
7. Handle malformed content gracefully

### Operations

**Requirements**:
1. Comprehensive logging (fetch attempts, errors, timings)
2. Error isolation (one bad feed doesn't break everything)
3. Exponential backoff for failing feeds
4. Bot information page with contact details
5. Database maintenance (pruning, vacuuming)
6. Monitoring for consistent failures

### Design Philosophy

**From 20 Years of Planet History**:
1. **Simple is Sustainable**: Fewer features = fewer bugs = longer life
2. **Good Netizen**: Respect servers, use bandwidth efficiently
3. **Security First**: Assume all feed content is hostile
4. **Static Output**: Faster and more reliable than dynamic
5. **Graceful Degradation**: Handle failures without breaking
6. **Clear Attribution**: Always credit source feeds
7. **Community Focused**: Built for specific communities, not generic social media

---

## Additional Resources Referenced

### Go Feed Parsing Libraries
- **gofeed**: https://github.com/mmcdole/gofeed
  - Most popular Go feed parser
  - Supports RSS 1.0, 2.0, Atom, JSON Feed
  - Active maintenance
  - Clean API

### Go HTML Sanitization
- **bluemonday**: https://github.com/microcosm-cc/bluemonday
  - Battle-tested HTML sanitizer
  - Whitelist-based approach
  - Policies for different trust levels
  - Used by many production systems

### Go SQLite Drivers
- **go-sqlite3**: https://github.com/mattn/go-sqlite3 (CGO required)
- **modernc.org/sqlite**: Pure Go implementation (no CGO)

### Standards and Specifications
- **Universal Feed Parser**: https://pythonhosted.org/feedparser/
  - Python library that Venus built upon
  - Documents feed format quirks and normalization
  - ETag/Last-Modified documentation: https://pythonhosted.org/feedparser/http-etag.html

- **RFC 7232**: HTTP Conditional Requests
  - Defines ETag and If-None-Match
  - Explains 304 Not Modified

- **RFC 7231**: HTTP Semantics
  - Defines status codes including 429 Too Many Requests

---

## Research Methodology

This research was conducted by:
1. Reviewing historical Planet/Venus documentation
2. Examining multiple implementations (Venus, Mars, Moonmoon, Pluto)
3. Studying real-world feed fetching issues documented by blog operators
4. Analyzing CVE-2009-2937 security vulnerability
5. Surveying active Planet sites for production patterns
6. Reviewing Go ecosystem libraries for feed parsing and HTML sanitization

The goal was to understand both **what works** (from successful implementations) and **what fails** (from documented problems) to create a robust, modern feed aggregator specification.

---

## Conclusion

Rogue Planet builds on 20+ years of feed aggregator development, incorporating:
- **Technical best practices** from Venus's normalization and Mars's experiments
- **Security lessons** from CVE-2009-2937 and modern XSS prevention
- **Operational wisdom** from real-world feed fetching problems
- **Simplicity principles** from Moonmoon's "stupidly simple" philosophy
- **Modern tooling** from Go's ecosystem and current libraries

The result is a specification that respects both publisher servers (through proper HTTP conditional requests) and user security (through comprehensive sanitization), while maintaining the simplicity and reliability that has made Planet-style aggregators successful for two decades.