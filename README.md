# Rogue Planet

[![Build Status](https://github.com/adewale/rogue_planet/actions/workflows/go.yml/badge.svg)](https://github.com/adewale/rogue_planet/actions/workflows/go.yml)

**Development Release v0.3.0** - This project is in active development. While core features are implemented and tested, it has not yet been deployed to production. The v1.0 production release will add feed autodiscovery, 301 redirect handling, and intelligent feed scheduling.

A modern feed aggregator written in Go, inspired by Planet Planet, Planet Venus and friends. Rogue Planet downloads RSS and Atom feeds from multiple sources and aggregates them into a single reverse-chronological stream published as a static HTML page.

## Features

- **Modern Go Implementation**: Clean, well-tested codebase using contemporary Go patterns
- **Multiple Feed Formats**: Supports RSS 1.0, RSS 2.0, Atom 1.0, and JSON Feed
- **HTTP Conditional Requests**: Implements proper ETag/Last-Modified caching to minimize bandwidth
- **Security First**:
  - XSS prevention via HTML sanitization (prevents CVE-2009-2937)
  - SSRF protection (blocks private IPs and localhost)
  - SQL injection prevention via prepared statements
  - Content Security Policy headers in generated output
- **Static Output**: Generates fast-loading HTML files that can be served by any web server
- **SQLite Database**: Efficient storage with proper indexing and WAL mode
- **Responsive Design**: Mobile-friendly default template with classic Planet Planet sidebar
- **Feed Sidebar**: Lists all subscribed feeds with last updated times and health status
- **Single Binary**: No dependencies, easy deployment

## Installation

### Option 1: Install via Go (Recommended)

```bash
go install github.com/adewale/rogue_planet/cmd/rp@latest
```

This installs `rp` to your `$GOPATH/bin` directory.

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/adewale/rogue_planet
cd rogue_planet

# Install using Make
make install

# Or build to local directory
make build

# Or use Go directly
go build -o rp ./cmd/rp
```

## Quick Start

> **🚀 New to Rogue Planet? See [WORKFLOWS.md](WORKFLOWS.md) for detailed setup and usage guides!**

1. **Initialize a new planet**:
   ```bash
   # Option 1: Initialize empty, add feeds manually
   rp init

   # Option 2: Initialize with feeds from a file
   rp init -f feeds.txt
   ```

2. **Edit `config.ini`** with your planet details:
   ```ini
   [planet]
   name = My Planet
   link = https://planet.example.com
   owner_name = Your Name
   ```

3. **Add feeds** (if you didn't use `-f` during init):
   ```bash
   # Add feeds one at a time (supports RSS, Atom, and JSON Feed)
   rp add-feed https://blog.golang.org/feed.atom
   rp add-feed https://github.blog/feed/
   rp add-feed https://username.micro.blog/feed.json

   # Or add multiple feeds from a file
   rp add-all -f feeds.txt
   ```

4. **Update your planet** (fetch feeds and generate HTML):
   ```bash
   rp update
   ```

5. **View the result**: Open `public/index.html` in your browser

For step-by-step workflows covering setup, daily operations, troubleshooting, and deployment, see **[WORKFLOWS.md](WORKFLOWS.md)**.

## Commands

### Core Commands
- `rp init [-f FILE]` - Initialise a new planet in the current directory
- `rp add-feed <url>` - Add a feed to the planet
- `rp add-all -f FILE` - Add multiple feeds from a file
- `rp remove-feed <url>` - Remove a feed from the planet
- `rp list-feeds` - List all configured feeds
- `rp status` - Show planet status (feed and entry counts)

### Operation Commands
- `rp update [--config FILE]` - Fetch all feeds and regenerate site
- `rp fetch [--config FILE]` - Fetch feeds without generating HTML
- `rp generate [--config FILE] [--days N]` - Generate HTML without fetching feeds
- `rp prune --days N [--config FILE] [--dry-run]` - Remove old entries from database

### Import/Export Commands
- `rp import-opml <file> [--dry-run]` - Import feeds from OPML file
- `rp export-opml [--output FILE]` - Export feeds to OPML format (stdout by default)

### Utility Commands
- `rp verify` - Validate configuration and environment
- `rp version` - Show version information

**Global Flags**:
- `--config <path>` - Path to config file (default: ./config.ini)

**Note**: All commands support the `--config` flag to specify a non-default configuration file.

## Configuration

Configuration is stored in `config.ini` (INI format):

```ini
[planet]
name = My Planet
link = https://planet.example.com
owner_name = Your Name
owner_email = you@example.com
output_dir = ./public
days = 7                    # Days of entries to include
log_level = info
concurrent_fetches = 5      # Parallel feed fetching (1-50)
group_by_date = true        # Group entries by date in output

[database]
path = ./data/planet.db
```

**Smart Content Display**: The `days` setting controls how many days back to look for entries. However, if no entries are found within that time window (e.g., feeds haven't updated recently), Rogue Planet automatically falls back to showing the most recent 50 entries regardless of age. This ensures your planet always has content to display, even if feeds go stale.

## Architecture

Rogue Planet follows a clear pipeline architecture:

```
Config → Crawler → Normaliser → Repository → Generator → HTML Output
```

- **Crawler**: Fetches feeds via HTTP with proper conditional request support, gzip decompression, and configurable user agent
- **Normaliser**: Parses feeds and sanitises HTML content
- **Repository**: Stores entries in SQLite with intelligent caching
- **Generator**: Creates static HTML using Go templates with responsive sidebar
- **Concurrent Fetching**: Configurable parallel feed fetching (1-50 concurrent requests)
- **Flexible Logging**: Configurable log levels (ERROR, WARN, INFO, DEBUG)

## Security Features

### XSS Prevention (CVE-2009-2937)

Rogue Planet sanitises all HTML content from feeds using [bluemonday](https://github.com/microcosm-cc/bluemonday) to prevent XSS attacks. The Planet Venus aggregator suffered from CVE-2009-2937, which allowed attackers to inject JavaScript via malicious feed content. Rogue Planet prevents this by:

- Stripping `<script>`, `<iframe>`, `<object>`, and `<embed>` tags
- Removing all event handlers (onclick, onerror, etc.)
- Blocking `javascript:` and `data:` URIs
- Only allowing http/https URL schemes
- Adding Content Security Policy headers to generated HTML

### SSRF Prevention

Prevents Server-Side Request Forgery attacks by:
- Blocking localhost, 127.0.0.1, ::1
- Blocking private IP ranges (RFC 1918)
- Blocking link-local addresses
- Only allowing http/https schemes

### Good Netizen Behavior

Implements proper HTTP conditional requests to minimize server load:
- Stores ETag and Last-Modified headers exactly as received
- Sends If-None-Match and If-Modified-Since on subsequent requests
- Handles 304 Not Modified responses correctly
- Never fabricates or modifies cache headers

## Test Coverage

All core packages have comprehensive test coverage (>75%).

Run tests:
```bash
# Run all tests (excludes network tests)
make test

# Run integration tests only
make test-integration

# Run tests with coverage report
make test-coverage

# Run tests with race detector
make test-race

# Run live network tests (requires internet connection)
go test -tags=network ./pkg/crawler -v

# Or use go directly
go test ./...
go test ./... -cover
```

**Integration Tests**: The project includes comprehensive integration tests that:
- Use `t.TempDir()` for automatic cleanup (no manual temp directory management)
- Test full end-to-end workflows (init → add feeds → fetch → generate HTML)
- Verify HTML content contains feed titles, entry titles, and proper structure
- Use mock HTTP servers for reliable, fast testing
- Test real-world feed parsing with saved snapshots from Daring Fireball (Atom) and Asymco (RSS)
- Verify HTML sanitisation prevents XSS attacks in real feed content

**Live Network Tests**: Tests marked with `// +build network` fetch from real URLs and require internet access. These tests verify:
- Live fetching from Daring Fireball and Asymco feeds
- Proper handling of gzip-encoded responses
- HTTP conditional request support (ETag, Last-Modified)
- Complete pipeline from fetch → parse → store → generate

## Development

### Project Structure

```
rogue_planet/
├── cmd/rp/              # CLI entry point
├── pkg/
│   ├── crawler/         # HTTP fetching with conditional requests
│   ├── normalizer/      # Feed parsing and HTML sanitisation (American spelling for package name)
│   ├── repository/      # SQLite database operations
│   ├── generator/       # Static HTML generation
│   └── config/          # Configuration parsing
├── specs/               # Specifications and testing plan
├── testdata/            # Test fixtures
└── CLAUDE.md            # Development guidance
```

### Building from Source

Using the Makefile (recommended):
```bash
# Install dependencies
make deps

# Quick build and test
make quick

# Full build with all checks
make check
```

Using Go directly:
```bash
# Install dependencies
go mod download

# Build
go build -o rp ./cmd/rp

# Build for production (smaller binary)
CGO_ENABLED=1 go build -ldflags="-s -w" -o rp ./cmd/rp

# Run tests
go test ./...

# Check coverage
go test ./... -cover

# Run specific package tests
go test ./pkg/crawler -v
```

### Makefile Targets

Run `make help` to see all available targets:

**Development:**
- `make quick` - Quick build for development (fmt + test + build)
- `make check` - Run all quality checks (fmt, vet, test, race)
- `make fmt` - Format code
- `make vet` - Run go vet
- `make lint` - Run linters

**Building:**
- `make build` - Build for current platform
- `make install` - Install to GOPATH/bin

**Testing:**
- `make test` - Run all tests
- `make coverage` - Generate HTML coverage report
- `make test-race` - Run with race detector
- `make bench` - Run benchmarks

**Other:**
- `make clean` - Remove build artifacts
- `make deps` - Download dependencies
- `make run-example` - Create example planet

## Recommended Workflows

> **📖 For detailed workflow examples, see [WORKFLOWS.md](WORKFLOWS.md)**

This section covers the most common workflows. For more detailed examples, troubleshooting guides, and advanced usage patterns, see the complete [Workflows Guide](WORKFLOWS.md).

### Daily Operation (Automated)

The recommended workflow for production use is to run `rp update` on a schedule:

```bash
# Cron job: Update every 30 minutes
*/30 * * * * cd /path/to/planet && ./rp update >> update.log 2>&1
```

This single command:
1. Fetches all active feeds (with proper HTTP caching)
2. Parses and sanitises new entries
3. Stores them in the database
4. Regenerates the HTML output

### Initial Setup Workflow

**Option 1: Start with a feeds file**
```bash
# Create feeds.txt with your feed URLs
cat > feeds.txt <<EOF
https://blog.golang.org/feed.atom
https://github.blog/feed/
https://daringfireball.net/feeds/main
EOF

# Initialize planet with feeds
rp init -f feeds.txt

# Edit config (optional)
vim config.ini

# Fetch and generate
rp update

# View result
open public/index.html
```

**Option 2: Add feeds interactively**
```bash
# Initialize empty planet
rp init

# Edit config first
vim config.ini

# Add feeds one by one
rp add-feed https://blog.golang.org/feed.atom
rp add-feed https://github.blog/feed/
rp add-feed https://daringfireball.net/feeds/main

# Fetch and generate
rp update

# View result
open public/index.html
```

### Development Workflow

When testing or developing your planet configuration:

```bash
# Check status
rp status

# List all feeds
rp list-feeds

# Fetch without regenerating HTML (test feed connectivity)
rp fetch

# Regenerate HTML without fetching (test template changes)
rp generate

# Full update
rp update
```

### Maintenance Workflow

**Adding new feeds:**
```bash
# Add single feed
rp add-feed https://example.com/feed.xml

# Or add multiple feeds from file
rp add-all -f new-feeds.txt

# Update to fetch new feeds
rp update
```

**Managing feeds:**
```bash
# List all feeds to see their status
rp list-feeds

# Remove a problematic feed
rp remove-feed https://broken-feed.example.com

# Check overall status
rp status
```

**Database maintenance:**
```bash
# Remove entries older than 90 days (keeps database small)
rp prune --days 90

# Check database size
ls -lh data/planet.db

# Regenerate HTML after pruning
rp generate
```

### Troubleshooting Workflow

**If a feed isn't updating:**
```bash
# 1. Check feed list and status
rp list-feeds

# 2. Try fetching manually
rp fetch

# 3. Check if feed URL is accessible
curl -I https://problem-feed.example.com/feed.xml

# 4. Check database for error counts
sqlite3 data/planet.db "SELECT url, fetch_error, fetch_error_count FROM feeds WHERE fetch_error != '';"

# 5. Remove and re-add if needed
rp remove-feed https://problem-feed.example.com/feed.xml
rp add-feed https://problem-feed.example.com/feed.xml
```

**If HTML isn't updating:**
```bash
# 1. Check if entries exist
rp status

# 2. Check database directly
sqlite3 data/planet.db "SELECT COUNT(*) FROM entries WHERE published >= datetime('now', '-7 days');"

# 3. Try regenerating HTML
rp generate

# 4. Check output file was updated
ls -l public/index.html
```

**If database gets too large:**
```bash
# Prune old entries
rp prune --days 30

# Vacuum database to reclaim space
sqlite3 data/planet.db "VACUUM;"

# Check new size
ls -lh data/planet.db
```

### Migration Workflow

**Moving from another aggregator (Venus, Planet):**
```bash
# 1. Export your feed list from old aggregator
# (Usually in a config file like config.ini or feeds.txt)

# 2. Initialize Rogue Planet with your feeds
rp init -f old-feeds.txt

# 3. Configure to match your old setup
vim config.ini

# 4. Do initial fetch
rp update

# 5. Compare output and adjust configuration
# 6. Set up cron job when satisfied
```

**Backing up your planet:**
```bash
# Backup database and config
tar czf planet-backup-$(date +%Y%m%d).tar.gz \
    config.ini \
    data/planet.db \
    feeds.txt

# Restore from backup
tar xzf planet-backup-20250101.tar.gz
rp generate  # Regenerate HTML from database
```

### Custom Template Workflow

> **📚 For complete theme documentation, see [THEMES.md](THEMES.md)**

Rogue Planet includes 5 built-in themes (Default, Classic, Elegant, Dark, Flexoki) and supports custom templates.

**Quick theme setup:**

```bash
# 1. Copy a theme
cp -r examples/themes/elegant themes/

# 2. Update config.ini
vim config.ini
# Add: template = ./themes/elegant/template.html

# 3. Generate with theme
rp generate

# 4. View result
open public/index.html
```

**Creating a custom theme:**

```bash
# 1. Create theme directory
mkdir -p themes/custom/static

# 2. Create template
vim themes/custom/template.html

# 3. Add styles
vim themes/custom/static/style.css

# 4. Configure and generate
echo "template = ./themes/custom/template.html" >> config.ini
rp generate
```

See [THEMES.md](THEMES.md) for:
- Complete template variables reference
- Theme creation guide
- Customization examples
- Template functions documentation

### Performance Tuning Workflow

For planets with many feeds (100+):

```bash
# 1. Edit config.ini
vim config.ini

# Increase concurrent fetches (default: 5)
# concurrent_fetches = 10

# Reduce days to keep (default: 7)
# days = 3

# 2. Test performance
time rp update

# 3. Monitor database size
ls -lh data/planet.db

# 4. Set up regular pruning
# Add to cron: 0 0 * * 0 cd /path/to/planet && ./rp prune --days 30
```

## Deployment

Since Rogue Planet generates static HTML, deployment is simple:

1. **Run on a schedule** (e.g., cron):
   ```cron
   */30 * * * * cd /path/to/planet && ./rp update
   ```

2. **Serve with any web server**:
   ```nginx
   server {
       listen 80;
       server_name planet.example.com;
       root /path/to/planet/public;
       index index.html;
   }
   ```

3. **Or use GitHub Pages**:
   - Commit generated `public/index.html` to repository
   - Enable GitHub Pages from repository settings

## Comparison with Venus/Planet

Rogue Planet improves upon classic feed aggregators:

| Feature | Venus/Planet | Rogue Planet |
|---------|--------------|--------------|
| Language | Python | Go |
| Dependencies | Many | Zero (single binary) |
| Security | CVE-2009-2937 | XSS prevention built-in |
| HTTP Caching | Often broken | RFC-compliant implementation |
| Deployment | virtualenv, system packages | Single binary |
| Database | pickle files | SQLite with indexes |
| Performance | Single-threaded | Concurrent feed fetching |
| Test Coverage | Minimal | >75% across core packages |

## Contributing

Contributions are welcome! Please:

1. Read the specifications in `specs/`
2. Follow the testing plan in `specs/testing-plan.md`
3. Maintain test coverage above 75%
4. Run `go fmt` and `go vet` before committing

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

Inspired by:
- [Planet Planet](https://web.archive.org/web/20171029175722/http://www.planetplanet.org/)
- [Planet Venus](https://github.com/rubys/venus)
- The decades of work by the feed aggregator community
- The Spin Doctors for the name: https://www.youtube.com/watch?v=GrQCro68sRU

Special thanks to the lessons learned from 20+ years of feed aggregator development, documented in `specs/rogue-planet-spec.md`.
