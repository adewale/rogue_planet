# Testing Guide

Rogue Planet has comprehensive test coverage with 375 tests across all packages.

**Test Status**: ✅ All 375 tests passing (100%)
**Coverage**: 78% overall (target: >75%)
**Last Updated**: 2025-10-10

---

## Quick Reference

### Run Tests

```bash
# All tests
go test ./...

# With verbose output
go test ./... -v

# Specific package
go test ./pkg/crawler -v

# With coverage
make coverage

# With race detector
go test -race ./...

# Integration tests only
make test-integration

# Network tests (requires internet)
go test -tags=network ./pkg/crawler -v
```

### Makefile Targets

```bash
make test           # Run all tests (excludes network tests)
make test-race      # Run with race detector
make coverage       # Generate HTML coverage report
make quick          # Format + test + build
make check          # Full quality checks (fmt + vet + test + race)
```

---

## Test Structure

### Unit Tests by Package

| Package | Tests | Coverage | Status |
|---------|-------|----------|--------|
| **cmd/rp** | 24 | ~75% | ✅ All passing |
| **pkg/config** | 27 | ~90% | ✅ All passing |
| **pkg/crawler** | 165 | ~85% | ✅ All passing |
| **pkg/generator** | 12 | ~60% | ✅ All passing |
| **pkg/normalizer** | 118 | ~80% | ✅ All passing |
| **pkg/repository** | 15 | ~85% | ✅ All passing |
| **Total** | **361** | **~78%** | **✅ 100%** |

**Note**: Generator coverage is below target (60% vs 75%). Consider adding more template tests.

### Test Categories

#### 1. Security Tests (Critical)

**XSS Prevention** (100 tests)
- Script tag removal
- Event handler blocking
- JavaScript/data URI blocking
- OWASP XSS vectors
- Malformed HTML handling
- **Status**: ✅ 100% passing

**SSRF Prevention** (25 tests)
- Localhost blocking (127.0.0.1, ::1, localhost)
- Private IP ranges (RFC 1918)
- Link-local addresses
- Invalid schemes (file://, ftp://, data:)
- **Status**: ✅ 100% passing

**SQL Injection Prevention**
- All queries use prepared statements
- No string concatenation in SQL
- **Status**: ✅ Verified via code review

#### 2. HTTP Functionality Tests (50+ tests)

**Conditional Requests** (16 tests)
- ETag handling (quoted, unquoted, weak)
- Last-Modified headers
- 304 Not Modified responses
- Cache preservation
- **Status**: ✅ 100% passing

**Size Limits** (3 tests)
- Under 10MB: ✅ Accepts
- Exactly 10MB: ✅ Accepts
- Over 10MB: ✅ Rejects
- **Status**: ✅ All boundary conditions correct

**Retry Logic** (7 tests)
- Exponential backoff
- Non-retryable errors
- Context cancellation
- **Status**: ✅ 100% passing

#### 3. Integration Tests (24 tests)

**Real-World Feeds**
- Daring Fireball (Atom)
- Asymco (RSS)
- Full pipeline: Parse → Store → Retrieve → Generate HTML
- **Status**: ✅ Time-invariant tests

**Full Workflow Tests**
- Init → Add feeds → Fetch → Generate
- HTML verification
- XSS prevention in output
- **Status**: ✅ 100% passing

---

## Test Patterns

### Table-Driven Tests

All tests use Go's table-driven pattern:

```go
func TestSomeFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"description", "input", "expected", false},
        // ... more cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Best Practices Used

✅ Use `t.TempDir()` for automatic cleanup
✅ Use `httptest.NewServer()` for HTTP mocking
✅ Test both success and error paths
✅ Descriptive test names
✅ Keep tests fast (<100ms per test)
✅ No implementation changes during testing

---

## Network Tests

Tests marked with `// +build network` require internet access:

```bash
go test -tags=network ./pkg/crawler -v
```

These tests verify:
- Live fetching from real URLs
- Gzip-encoded responses
- HTTP conditional requests
- Complete pipeline with real feeds

**Note**: Network tests are excluded from regular test runs to maintain reliability.

---

## Test Files

### Core Test Files

**Crawler**:
- `pkg/crawler/crawler_test.go` - Basic functionality
- `pkg/crawler/crawler_comprehensive_test.go` - Security & edge cases (140+ tests)
- `pkg/crawler/crawler_user_agent_test.go` - User agent handling
- `pkg/crawler/crawler_live_test.go` - Network tests

**Normalizer**:
- `pkg/normalizer/normalizer_test.go` - Basic functionality
- `pkg/normalizer/normalizer_xss_test.go` - XSS prevention (100+ tests)
- `pkg/normalizer/normalizer_realworld_test.go` - Real feed parsing

**Repository**:
- `pkg/repository/repository_test.go` - Database operations

**Generator**:
- `pkg/generator/generator_test.go` - Template rendering
- `pkg/generator/generator_integration_test.go` - Full generation

**Config**:
- `pkg/config/config_test.go` - Configuration parsing

**CLI**:
- `cmd/rp/commands_test.go` - Command implementations
- `cmd/rp/integration_test.go` - Full workflow tests
- `cmd/rp/realworld_integration_test.go` - Real feed integration

---

## Running Specific Tests

```bash
# Security tests
go test ./pkg/crawler -run TestValidateURL -v
go test ./pkg/normalizer -run TestSanitizeHTML_XSS -v

# Boundary conditions
go test ./pkg/crawler -run TestFetch_SizeLimits -v

# Real-world feeds
go test ./cmd/rp -run TestRealWorldFeedsFullPipeline -v

# Specific test case
go test ./pkg/crawler -run TestValidateURL/localhost -v
```

---

## Coverage Reports

Generate detailed coverage reports:

```bash
# HTML report (opens in browser)
make coverage

# Terminal summary
go test ./... -cover

# Per-package breakdown
go test ./pkg/crawler -cover
go test ./pkg/normalizer -cover
go test ./pkg/repository -cover
```

Coverage reports are generated in `coverage/` directory:
- `coverage.out` - Raw coverage data
- `coverage.html` - Interactive HTML report

---

## CI/CD Integration

### Pre-commit Hook

```bash
#!/bin/bash
make quick || exit 1
```

### PR Validation

```bash
#!/bin/bash
make check           # Format, vet, test, race
make coverage        # Ensure >75% coverage
```

### Nightly Build

```bash
#!/bin/bash
go test -race ./...
go test -tags=network ./...
```

---

## Test Maintenance

### Adding New Tests

1. Follow table-driven pattern
2. Use descriptive test names
3. Test both success and error cases
4. Keep tests isolated (no shared state)
5. Use `t.TempDir()` for file operations

### Updating Tests

When modifying functionality:
1. Update tests FIRST (TDD)
2. Ensure tests fail without changes
3. Implement changes
4. Verify tests pass
5. Check coverage didn't decrease

### Test Fixtures

Real-world feed snapshots in `testdata/`:
- `daringfireball-feed.xml` - Atom feed
- `asymco-feed.xml` - RSS feed

These ensure tests remain stable without network dependencies.

---

## Common Test Issues

### Test Fails with "build failed"
**Solution**: Check for syntax errors in test files

### Test Hangs
**Solution**: Check for missing timeouts in httptest servers

### Race Detector Fails
**Solution**: Run without race detector to identify the test, then fix concurrency issues

### Coverage Not Generated
**Solution**: Ensure coverage directory exists: `mkdir -p coverage`

---

## Test Metrics

**Total Tests**: 375
**Test Runtime**: ~12 seconds (all tests)
**Average per Test**: ~30ms
**Slowest Package**: pkg/crawler (retry tests with backoff)
**Fastest Package**: pkg/config (pure logic)

---

## Historical Context

For historical information about test development and resolved issues, see:
- `docs/archive/TEST_PLAN.md` - Original comprehensive test plan
- `docs/archive/TEST_IMPLEMENTATION_SUMMARY.md` - Implementation results
- `docs/archive/TEST_FAILURE_INSIGHTS.md` - Analysis of initial failures
- `docs/archive/TEST_FAILURES.md` - All issues resolved as of 2025-10-10

**All identified issues have been fixed**. The codebase maintains 100% test pass rate.

---

## Contributing Tests

When contributing:

1. **Security tests are mandatory** for:
   - Any user input handling
   - HTML/URL processing
   - Database queries

2. **Integration tests** for:
   - New commands
   - End-to-end workflows

3. **Coverage target**: >75% for all packages

4. **Test documentation**:
   - Use descriptive names
   - Document complex test scenarios
   - Explain expected behaviour

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

## Test Philosophy

> **"Good tests find problems. Great tests teach you about your code."**

Rogue Planet's tests serve multiple purposes:
- **Verification**: Ensure code works correctly
- **Documentation**: Demonstrate how to use APIs
- **Safety net**: Catch regressions
- **Design feedback**: Highlight complexity
- **Confidence**: Enable fearless refactoring

Our test suite achieves all five goals.
