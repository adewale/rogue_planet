# Files to Push to GitHub

## Summary
This document lists all files that should be pushed to the GitHub repository for Rogue Planet v0.1.0 (Initial Development Release).

## Core Application (cmd/rp/)
```
cmd/rp/main.go
cmd/rp/commands.go
cmd/rp/logger_test.go
cmd/rp/commands_test.go
cmd/rp/integration_test.go
cmd/rp/realworld_integration_test.go
```

## Package Files (pkg/)
```
pkg/config/config.go
pkg/config/config_test.go

pkg/crawler/crawler.go
pkg/crawler/crawler_test.go
pkg/crawler/crawler_live_test.go
pkg/crawler/crawler_comprehensive_test.go
pkg/crawler/crawler_user_agent_test.go

pkg/generator/generator.go
pkg/generator/generator_test.go
pkg/generator/generator_integration_test.go

pkg/normalizer/normalizer.go
pkg/normalizer/normalizer_test.go
pkg/normalizer/normalizer_realworld_test.go

pkg/repository/repository.go
pkg/repository/repository_test.go
```

## Documentation (Root Level)
```
README.md              ✅ Already staged
QUICKSTART.md
WORKFLOWS.md
THEMES.md              (NEW - comprehensive theme guide)
CONTRIBUTING.md
CHANGELOG.md
CLAUDE.md
TODO.md
TESTING.md
CONSISTENCY_CHECK_REPORT.md
CONSISTENCY_REVIEW.md
CLEANUP_SUMMARY.md
TEST_FAILURES.md       (optional - may want to review/remove)
```

## Specifications (specs/)
```
specs/rogue-planet-spec.md    ✅ Already staged
specs/research.md
specs/testing-plan.md
```

## Configuration & Build Files
```
.gitignore
go.mod
go.sum
Makefile
LICENSE                (⚠️  NEEDS TO BE CREATED)
```

## Examples (examples/)
```
examples/README.md
examples/config.ini
examples/feeds.txt
examples/themes/classic/
examples/themes/elegant/
examples/themes/dark/
examples/themes/flexoki/      (NEW - automatic light/dark mode theme)
```

## Test Data (testdata/)
```
testdata/test-feed.xml
testdata/asymco-feed.xml
testdata/daringfireball-feed.xml
testdata/feeds/*       (if contains files)
testdata/expected/*    (if contains files)
testdata/malicious/*   (if contains files)
```

## Documentation Archives (docs/)
```
docs/archive/*         (optional - historical docs, consider excluding)
```

---

## Files Excluded by .gitignore (DO NOT PUSH)
- `bin/` - Build artifacts
- `rp` - Binary executable
- `*.exe`, `*.dll`, `*.so`, `*.dylib` - Platform binaries
- `*.test` - Test binaries
- `*.out` - Go coverage files
- `coverage.html` - HTML coverage reports
- `*.db`, `*.db-shm`, `*.db-wal` - Database files
- `public/`, `output/` - Generated HTML output
- `config.ini` - Local configuration (keep `config.example.ini` if exists)
- `tmp/`, `cache/` - Temporary files
- `.vscode/`, `.idea/` - IDE files
- `.DS_Store`, `Thumbs.db` - OS files

---

## Quick Commands

### Stage all project files:
```bash
# Core app
git add cmd/

# Packages
git add pkg/

# Documentation
git add README.md QUICKSTART.md WORKFLOWS.md CONTRIBUTING.md
git add CHANGELOG.md CLAUDE.md TODO.md TESTING.md

# Specs
git add specs/

# Config & build
git add .gitignore go.mod go.sum Makefile LICENSE

# Examples
git add examples/

# Test data
git add testdata/

# Optional: archived docs (review first)
# git add docs/
```

### Verify what will be committed:
```bash
git status
git diff --cached
```

### Create initial commit:
```bash
git commit -m "Initial development release: Rogue Planet v0.1.0

- Complete feed aggregator implementation
- RSS 1.0, RSS 2.0, Atom 1.0 support
- HTML sanitization (CVE-2009-2937 prevention)
- HTTP conditional requests (ETag/Last-Modified)
- SSRF protection
- SQLite storage with WAL mode
- Static HTML generation
- 80%+ test coverage
- All core features complete

🤖 Generated with Claude Code (https://claude.com/claude-code)"
```

### Push to GitHub:
```bash
# Add remote (replace with your repo URL)
git remote add origin https://github.com/YOUR_USERNAME/rogue-planet.git

# Push to main branch
git push -u origin main

# Create and push release tag
git tag -a v0.1.0 -m "Release v0.1.0 - Initial Development Release"
git push origin v0.1.0
```

---

## Pre-Push Checklist

- [ ] **LICENSE file created** (required before public release)
- [ ] **All tests passing** (`make check`)
- [ ] **Documentation reviewed** (README, QUICKSTART, etc.)
- [ ] **No secrets in code** (API keys, passwords, etc.)
- [ ] **Version number correct** (check cmd/rp/main.go)
- [ ] **User-Agent includes contact URL** (will update after repo created)
- [ ] **.gitignore is comprehensive** (no build artifacts committed)
- [ ] **Example configs have no secrets** (examples/config.ini)

---

## Notes

1. **LICENSE file**: You must create this before pushing. Common choices:
   - MIT License (permissive, simple)
   - Apache 2.0 (permissive, includes patent grant)
   - GPL v3 (copyleft, requires derivatives to be open source)

2. **docs/archive/**: Consider whether to include archived documentation. These appear to be historical test failure notes that may not be relevant for public release.

3. **CLEANUP_SUMMARY.md, CONSISTENCY_REVIEW.md, TEST_FAILURES.md**: Review these files - they may be internal development notes that don't need to be public.

4. **User-Agent string**: After creating the GitHub repo, update the User-Agent in `pkg/crawler/crawler.go` to include the repo URL for contact information.

5. **Binary builds**: Don't commit the `rp` binary or `bin/` directory - these are generated files. Users will build from source or download from releases.

---

## Development Status

**Current Version**: 0.1.0 (Initial Development Release)
- All core features implemented and tested
- Ready for community feedback and real-world testing
- Production use possible but marked as pre-1.0 development release

**Path to 1.0.0**:
- Gather community feedback
- Additional real-world testing
- Performance optimization
- Extended documentation based on user questions

---

*Generated: 2025-10-10*
*For: Rogue Planet v0.1.0 Initial Development Release*
