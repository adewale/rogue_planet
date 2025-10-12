# Contributing to Rogue Planet

Thank you for your interest in contributing to Rogue Planet! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Process](#development-process)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Release Process](#release-process)

## Code of Conduct

Be respectful, constructive, and professional. We're all here to make great software together.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- SQLite3 (usually pre-installed)
- Make (optional, but recommended)

### Setting Up Development Environment

```bash
# 1. Fork the repository on GitHub

# 2. Clone your fork
git clone https://github.com/YOUR_USERNAME/rogue_planet
cd rogue_planet

# 3. Add upstream remote
git remote add upstream https://github.com/adewale/rogue_planet

# 4. Install dependencies
go mod download

# 5. Build and test
make build
make test

# 6. Create a feature branch
git checkout -b feature/your-feature-name
```

### Project Structure

```
rogue_planet/
├── cmd/rp/              # CLI application
├── pkg/
│   ├── config/          # Configuration parsing
│   ├── crawler/         # HTTP fetching
│   ├── generator/       # HTML generation
│   ├── normalizer/      # Feed parsing
│   └── repository/      # Database operations
├── examples/            # Example configs
├── specs/               # Specifications
├── testdata/            # Test fixtures
└── docs/                # Documentation
```

## Development Process

### Branching Strategy

- `main` - Production-ready code
- `feature/*` - New features
- `fix/*` - Bug fixes
- `docs/*` - Documentation updates

### Making Changes

1. **Create an issue** first to discuss significant changes
2. **Create a branch** from `main`
3. **Make your changes** with clear, atomic commits
4. **Add tests** for new functionality
5. **Update documentation** as needed
6. **Run all checks** before submitting

### Running Checks

```bash
# Format code
make fmt

# Run linters
make vet
make lint  # if golangci-lint is installed

# Run tests
make test

# Run tests with race detector
make test-race

# Check coverage
make test-coverage

# Run all checks
make check
```

## Coding Standards

### Go Style

Follow standard Go conventions:

- Use `gofmt` for formatting (automatic with `make fmt`)
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Naming Conventions

- **Packages**: Short, lowercase, single-word names (e.g., `crawler`, `repository`)
- **Files**: Lowercase with underscores (e.g., `crawler_test.go`)
- **Types**: PascalCase (e.g., `FeedCache`, `EntryData`)
- **Functions**: PascalCase for exported, camelCase for internal
- **Variables**: camelCase

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("fetch feed: %w", err)
}

// Good: Handle expected errors explicitly
if errors.Is(err, ErrNotFound) {
    return nil
}

// Bad: Ignore errors
_ = someFunction()  // Don't do this

// Bad: Generic error messages
return errors.New("error")  // Not helpful
```

### Comments

- Add package documentation in `doc.go` or at top of main file
- Document all exported types, functions, and constants
- Use complete sentences
- Explain *why*, not *what* (code shows what)

```go
// Good: Explains why
// Fetch uses HTTP conditional requests to minimize bandwidth and
// avoid being rate-limited by feed servers.
func (c *Crawler) Fetch(ctx context.Context, url string) (*Response, error) {

// Bad: Just repeats the code
// Fetch fetches a URL
func (c *Crawler) Fetch(ctx context.Context, url string) (*Response, error) {
```

## Testing

### Test Requirements

- **All new code must have tests**
- **Maintain >75% coverage** for all packages
- **Security-critical code must have 100% coverage**
- Use table-driven tests for multiple scenarios
- Test both success and error cases

### Writing Tests

```go
func TestFetch(t *testing.T) {
    tests := []struct {
        name    string
        url     string
        want    int
        wantErr bool
    }{
        {
            name: "valid feed",
            url:  "https://example.com/feed.xml",
            want: 200,
            wantErr: false,
        },
        {
            name: "invalid URL",
            url:  "not-a-url",
            want: 0,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Fetch(tt.url)
            if (err != nil) != tt.wantErr {
                t.Errorf("Fetch() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Fetch() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Test Organization

- Unit tests in `*_test.go` files alongside source
- Integration tests in `*_integration_test.go`
- Use `t.TempDir()` for temporary files (auto-cleanup)
- Use `httptest.NewServer()` for mock HTTP servers
- Build tags for network tests: `// +build network`

### Running Specific Tests

```bash
# Run all tests
go test ./...

# Run specific package
go test ./pkg/crawler

# Run specific test
go test ./pkg/crawler -run TestFetch

# Run with verbose output
go test ./... -v

# Run network tests
go test -tags=network ./pkg/crawler -v
```

## Submitting Changes

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style changes (formatting, no logic change)
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvement
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:
```
feat(crawler): add gzip decompression support

fix(repository): handle null values in cache fields

docs(workflows): add Docker deployment example

test(crawler): add live network tests
```

### Pull Request Process

1. **Update documentation** for any user-facing changes
2. **Add tests** for new functionality
3. **Update CHANGELOG.md** under "Unreleased" section
4. **Run all checks** (`make check`)
5. **Create pull request** with clear description

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tests added/updated
- [ ] All tests passing
- [ ] Coverage maintained

## Checklist
- [ ] Code follows style guidelines
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] All checks passing
```

### Review Process

- Maintainers will review your PR
- Address feedback promptly
- Keep discussion focused and professional
- Once approved, maintainers will merge

## Release Process

(For maintainers)

1. Update version in `cmd/rp/main.go`
2. Update `CHANGELOG.md` (move Unreleased to new version)
3. Create git tag: `git tag v1.x.x`
4. Push tag: `git push origin v1.x.x`
5. GitHub Actions will create release (once configured)
6. Or manually: `make release` then upload `dist/*` to GitHub

## Security

### Reporting Security Issues

**Do not open public issues for security vulnerabilities.**

Email security concerns to: [security contact - to be added]

### Security Requirements

- All HTML must be sanitized (use `bluemonday`)
- All database queries must use prepared statements
- All URLs must be validated for SSRF prevention
- Security-critical code requires 100% test coverage

## Documentation

### What to Document

- **README.md**: High-level overview, installation, quick start
- **WORKFLOWS.md**: Operational workflows and examples
- **Code comments**: All exported functions and types
- **CHANGELOG.md**: All user-visible changes

### Documentation Style

- Use clear, concise language
- Provide examples for complex features
- Link between related documentation
- Keep documentation in sync with code

## Getting Help

- **Questions**: Open a GitHub Discussion
- **Bugs**: Open a GitHub Issue
- **Features**: Open a GitHub Issue or Discussion
- **Documentation**: Check CLAUDE.md for development details

## Recognition

Contributors will be:
- Listed in git history
- Mentioned in release notes for significant contributions
- Added to CONTRIBUTORS file (if created)

Thank you for contributing to Rogue Planet! 🚀
