# Rogue Planet - Unfinished Tasks Summary

**Generated**: 2025-10-30
**Current Branch**: 0.4.0
**Status**: In active development for v0.4.0 release

---

## ✅ Recently Completed

### P2.6 - Atom Torture Test Research (COMPLETE)
- ✅ Comprehensive research document created
- ✅ Added 2 missing test cases (XHTML case-sensitivity, xml:base resolution)
- ✅ Updated all test fixture headers with references
- ✅ Updated test suite documentation
- ✅ All 20 torture tests now passing
- **Files**: `specs/research/ATOM_TORTURE_TEST_RESEARCH.md`, test fixtures, test code
- **Status**: Pushed to remote, ready for PR

---

## 🔴 Priority 1: BLOCKING ISSUES

These prevent users from getting started or break documented features.

### P1.1 - Missing examples/config.ini
- **Status**: ❌ NOT STARTED
- **Issue**: Referenced in Makefile (lines 182, 191, 204, 217) but doesn't exist
- **Impact**: `make examples` target fails completely
- **Estimate**: 1 hour
- **Files**: Create `examples/config.ini`

### P1.2 - macOS-specific sed in Makefile
- **Status**: ❌ NOT STARTED
- **Issue**: `sed -i ''` syntax fails on Linux
- **Impact**: `make examples` fails on Linux systems
- **Estimate**: 30 minutes
- **Files**: `Makefile` (lines 194, 207, 220)

### P1.3 - Missing QUICKSTART.md
- **Status**: ❌ NOT STARTED
- **Issue**: Referenced in examples/README.md:100, marked complete in TODO.md:88
- **Decision Needed**: Create it or remove references?
- **Estimate**: 2 hours (create) OR 15 minutes (remove)
- **Files**: `QUICKSTART.md` or `examples/README.md`, `specs/TODO.md`

### P1.4 - Missing THEMES.md
- **Status**: ❌ NOT STARTED
- **Issue**: Referenced in examples/README.md:42,99, marked complete in TODO.md:88
- **Decision Needed**: Create it or remove references?
- **Estimate**: 3 hours (create) OR 15 minutes (remove)
- **Files**: `THEMES.md` or `examples/README.md`, `specs/TODO.md`

### P1.5 - Missing setup-example-planet.sh
- **Status**: ❌ NOT STARTED
- **Issue**: Referenced in Makefile:172, file doesn't exist
- **Recommendation**: Remove from Makefile (redundant with `make run-example`)
- **Estimate**: 5 minutes
- **Files**: `Makefile`

---

## 📝 Priority 2: DOCUMENTATION GAPS

### P2.1 - TODO.md Out of Sync
- **Status**: ❌ NOT STARTED
- **Issue**: Claims features complete that don't exist
  - Line 88: "✅ QUICKSTART.md" (doesn't exist)
  - Line 88: "✅ THEMES.md" (doesn't exist)
  - Line 204: "✅ golang.org/x/time/rate" (not in go.mod)
- **Estimate**: 30 minutes
- **Files**: `specs/TODO.md`

### P2.2 - Undocumented Config Options
- **Status**: ❌ NOT STARTED
- **Issue**: `filter_by_first_seen` and `sort_by` implemented but not in README
- **Impact**: Users don't know these features exist
- **Estimate**: 1 hour
- **Files**: `README.md`

### P2.3 - Rate Limiting Documentation
- **Status**: ❌ NOT STARTED
- **Issue**: CLAUDE.md claims it's a dependency but it's not implemented
- **Fix**: Remove from dependencies, add to "Future Features"
- **Estimate**: 15 minutes
- **Files**: `CLAUDE.md` (lines 119, 185)

### P2.4 - 301 Redirect Documentation
- **Status**: ❌ NOT STARTED
- **Issue**: CLAUDE.md:188 claims URLs are auto-updated on 301
- **Reality**: Redirects followed but DB never updated
- **Estimate**: 10 minutes
- **Files**: `CLAUDE.md`

### P2.5 - Coverage Reporting Clarity
- **Status**: ❌ NOT STARTED
- **Issue**: TODO.md reports 88.4% but excludes cmd/rp (26.6%)
- **Fix**: Include cmd/rp in average or note exclusion
- **Estimate**: 5 minutes
- **Files**: `specs/TODO.md`

---

## ⚠️ Priority 3: INCOMPLETE FEATURES

### P3.1 - Dry-Run Prune is a Stub
- **Status**: ❌ NOT STARTED
- **Issue**: `rp prune --dry-run` doesn't show what would be deleted
- **Current**: Just prints "Dry run: would delete entries..."
- **Expected**: Should query and preview entries
- **Estimate**: 1 hour
- **Files**: `cmd/rp/commands.go`

### P3.2 - 301 Redirect URL Updating
- **Status**: ❌ NOT STARTED
- **Issue**: FinalURL field exists but feature not implemented
- **Impact**: Permanent redirects waste work on every fetch
- **Decision Needed**: Implement or defer to v1.0?
- **Estimate**: 3 hours (implement) OR 1 hour (remove field)
- **Files**: `pkg/repository/repository.go`, `cmd/rp/commands.go`

### P3.3 - Feed Pause/Activate Commands
- **Status**: ❌ NOT STARTED
- **Issue**: Database has `active` field but no commands to change it
- **Current**: Only `rp remove-feed` (permanent deletion)
- **Wanted**: `rp pause-feed <url>` and `rp activate-feed <url>`
- **Estimate**: 2 hours
- **Files**: `pkg/repository/repository.go`, `cmd/rp/commands.go`, `cmd/rp/main.go`

### P3.4 - Rate Limiting
- **Status**: ⏸️ DEFERRED to v1.0
- **Issue**: Documented but not implemented
- **Recommendation**: Fix documentation (P2.3), defer implementation

---

## 🧹 Priority 4: CODE QUALITY

### P4.1 - Unused Subscribers Field
- **Status**: ❌ NOT STARTED
- **Issue**: `FeedData.Subscribers` always zero
- **Fix**: Remove field or add TODO comment
- **Estimate**: 10 minutes
- **Files**: `pkg/generator/generator.go`

### P4.2 - SQL Injection Safety Comment
- **Status**: ❌ NOT STARTED
- **Issue**: `repository.go:369-376` looks unsafe but is safe
- **Fix**: Add comment explaining validation
- **Estimate**: 5 minutes
- **Files**: `pkg/repository/repository.go`

### P4.3 - FetchWithRetry Unused
- **Status**: ❌ NOT STARTED
- **Issue**: Method exists but never used in production
- **Recommendation**: Keep as public API, add comment
- **Estimate**: 5 minutes
- **Files**: `pkg/crawler/crawler.go`

### P4.4 - next_fetch/fetch_interval Unused
- **Status**: ❌ NOT STARTED
- **Issue**: DB fields exist but never consulted
- **Recommendation**: Keep for v1.0, add comment
- **Estimate**: 10 minutes
- **Files**: `pkg/repository/repository.go`

---

## 🔧 Priority 5: QUICK WINS (Optional)

These are technical improvements that can be done in v0.4.0 or deferred.

### P5.1 - HTTP Connection Pooling
- **Status**: ❌ NOT STARTED
- **Benefit**: 10-20% faster HTTP requests
- **Estimate**: 2 hours
- **Files**: `pkg/crawler/crawler.go`

### P5.2 - Config Validation on Load
- **Status**: ❌ NOT STARTED
- **Benefit**: Fail fast with clear errors
- **Estimate**: 2 hours
- **Files**: `pkg/config/config.go`

### P5.3 - Export Sentinel Errors
- **Status**: ❌ NOT STARTED
- **Benefit**: Better error handling with `errors.Is()`
- **Estimate**: 4 hours
- **Files**: `pkg/normalizer/errors.go`, `pkg/generator/errors.go`

### P5.4 - Improve cmd/rp Test Coverage
- **Status**: ❌ NOT STARTED
- **Current**: 26.6% coverage
- **Target**: >40% coverage
- **Estimate**: 2 days
- **Files**: `cmd/rp/*_test.go`

---

## 📦 Untracked Files (Needs Decision)

These files exist locally but aren't committed:

```
docs/CODEBASE_AUDIT_REPORT.md
docs/CODE_QUALITY_SUPPLEMENT.md
specs/AUDIT_HEURISTICS.md
specs/COMPREHENSIVE_AUDIT_2025-10-19.md
specs/INDUSTRY_AUDIT_HEURISTICS_RESEARCH.md
specs/TEST_ASSERTION_FIX_RESULTS.md
specs/TEST_ASSERTION_QUALITY_AUDIT.md
```

**Decision Needed**: Commit these or move to separate branch/archive?

---

## 🚀 v1.0.0 Planned Features

Deferred until after v0.4.0 release:

### P0.1 - Feed Autodiscovery
- **Status**: 📋 PLANNED for v1.0
- **Feature**: Parse HTML `<link rel="alternate">` to find feeds
- **Commands**: `rp discover <url>`, `--auto-discover` flag
- **Estimate**: 2 days

### P0.2 - Intelligent Feed Scheduling
- **Status**: 📋 PLANNED for v1.0
- **Feature**: Adaptive polling based on feed update frequency
- **Database**: Fields exist but not used yet
- **Estimate**: 5 days

### Additional v1.0 Features
- Production deployment documentation
- Binary distribution packages
- GitHub repository configuration
- Community announcement

---

## 📊 v0.4.0 Release Criteria

Version 0.4.0 can be released when:

- [ ] All P1 (blocking) issues resolved
- [ ] All P2 (documentation) issues resolved
- [ ] At least 2 of 4 P3 (incomplete features) resolved
- [ ] P4 code cleanup complete
- [ ] All tests passing
- [ ] CHANGELOG.md updated
- [ ] README.md reflects actual features
- [ ] TODO.md synchronized with reality
- [ ] `make examples` works on both macOS and Linux

**Optional**:
- [ ] P5.1-P5.3 quick wins (adds 1 day)

---

## 📅 Estimated Timeline

### Minimum v0.4.0 (Required Tasks Only):
- **P1 (Blocking)**: 1 day (assuming remove QUICKSTART/THEMES)
- **P2 (Documentation)**: 1 day
- **P3 (Features)**: 1 day (2 of 4 features)
- **P4 (Code Quality)**: 0.5 days
- **Total**: 3.5 days

### Full v0.4.0 (With Optional Quick Wins):
- **Base**: 3.5 days
- **P5 Quick Wins**: 1 day
- **Total**: 4.5 days

### v1.0.0 (After v0.4.0):
- **Feed Autodiscovery**: 2 days
- **Intelligent Scheduling**: 5 days
- **Documentation & Packaging**: 3 days
- **Total**: 10 days (2 weeks)

---

## 🎯 Recommended Next Steps

1. **Immediate** (This session):
   - Decide on QUICKSTART.md and THEMES.md (create or remove?)
   - Decide on untracked audit files (commit or archive?)

2. **Phase 1** (1 day):
   - Create examples/config.ini (P1.1)
   - Fix Makefile sed compatibility (P1.2)
   - Execute QUICKSTART/THEMES decision (P1.3, P1.4)
   - Remove setup-example-planet.sh reference (P1.5)
   - Update TODO.md (P2.1)

3. **Phase 2** (1 day):
   - Document config options (P2.2)
   - Fix rate limiting docs (P2.3)
   - Fix 301 redirect docs (P2.4)
   - Fix coverage reporting (P2.5)

4. **Phase 3** (1 day):
   - Implement dry-run prune preview (P3.1)
   - Choose 2 of: 301 redirect, pause/activate, or defer

5. **Phase 4** (0.5 days):
   - All P4 code cleanup tasks

6. **Release v0.4.0**

---

## 📝 Notes

- P2.6 (Atom Torture Test) completed this session
- All core v0.3.0 features are complete and working
- Most issues are documentation/polish, not broken features
- Test coverage is good (>75% on libraries)
- Security implementation is solid

---

**Last Updated**: 2025-10-30
**Updated By**: Claude Code
