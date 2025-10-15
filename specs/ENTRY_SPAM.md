# Entry Spam Prevention

## Problem Statement

When adding a new feed to a planet aggregator, all historical entries from that feed (typically 50-100 entries) appear at once in the generated HTML, sorted by their original published dates. This creates several problems:

1. **Timeline Pollution**: Old entries backdated into the chronological stream
2. **RSS Feed Spam**: Readers see "100 new items" notification
3. **User Confusion**: Mix of genuinely new content with old content
4. **Social Media Spam**: If planet RSS is auto-posted to Twitter/Mastodon, causes spam

This was the **#1 complaint** about Planet Venus from users.

**Original complaint**: [Reddit comment by /u/fgilcher](https://www.reddit.com/r/rust/comments/mvm4r2/comment/gvcwxq0/):

> "One of the most annoying things about Venus/PlanetPlanet is how any change to a blog or addition of a new blog adds a ton of entries all at once, spamming people's RSS reader, as well as the twitter account that reposts the feed"

## Current Behavior

```
User adds new feed: https://example.com/feed.xml
Feed has 50 entries from last 30 days
Rogue Planet fetches all 50 entries
All 50 entries appear in HTML (because published dates are within window)
Result: Timeline flooded with 50 "new" entries that are actually old
```

**Example Scenario**:
```bash
$ rp add-feed https://newblog.com/feed.xml
✓ Added feed: https://newblog.com/feed.xml (ID: 5)

$ rp update
Fetching 5 feeds...
  [5/5] https://newblog.com/feed.xml
  ✓ Fetched 47 entries

$ rp generate
Generated public/index.html with 147 entries

# Problem: 47 entries just appeared in timeline, backdated to their
# original published dates, polluting the chronological stream
```

## Existing Infrastructure

Rogue Planet **already has** the necessary database fields:

```sql
CREATE TABLE entries (
    ...
    published TEXT,      -- Original published date from feed
    first_seen TEXT,     -- When Rogue Planet first fetched this entry
    ...
);

CREATE INDEX idx_entries_published ON entries(published DESC);
CREATE INDEX idx_entries_first_seen ON entries(first_seen DESC);  -- Already exists!
```

**Key Insight**: We track both dates but only use `published` for filtering/sorting.

## Proposed Solution

Add two configuration options:

```ini
[planet]
filter_by_first_seen = false   # Default: backwards compatible
sort_by = "published"           # Default: "published" or "first_seen"
```

### Option 1: Filter by First Seen

**Behavior**: Only show entries first seen within the time window.

```sql
-- Current query (line 279-285 in repository.go)
WHERE f.active = 1 AND e.published >= ?
ORDER BY e.published DESC

-- With filter_by_first_seen = true
WHERE f.active = 1 AND e.first_seen >= ?
ORDER BY e.published DESC  -- Still sort by published for chronology
```

**Effect**: When adding a new feed, only entries fetched in the last N days appear, regardless of their original published date.

### Option 2: Sort by First Seen

**Behavior**: Order entries by when aggregator first saw them.

```sql
-- With sort_by = "first_seen"
WHERE f.active = 1 AND e.published >= ?
ORDER BY e.first_seen DESC  -- Show newest-to-planet first
```

**Effect**: "River of news" ordering - entries appear in the order the planet discovered them, not their original published dates.

### Combined Usage

```ini
[planet]
filter_by_first_seen = true   # Only show recently discovered entries
sort_by = "first_seen"         # Sort by discovery order
```

**Result**: Perfect for preventing entry spam. New feeds only contribute entries discovered within the time window, sorted by discovery time.

## Implementation

### 1. Config Changes

**File**: `pkg/config/config.go`

```go
type Config struct {
    Planet struct {
        // Existing fields...
        FilterByFirstSeen bool   `ini:"filter_by_first_seen"`
        SortBy            string `ini:"sort_by"`  // "published" or "first_seen"
    }
}

func (c *Config) Validate() error {
    // Existing validations...

    if c.Planet.SortBy == "" {
        c.Planet.SortBy = "published"  // Default
    }
    if c.Planet.SortBy != "published" && c.Planet.SortBy != "first_seen" {
        return fmt.Errorf("sort_by must be 'published' or 'first_seen', got: %s", c.Planet.SortBy)
    }

    return nil
}
```

### 2. Repository Changes

**File**: `pkg/repository/repository.go`

Add new method:

```go
// GetRecentEntriesWithOptions returns entries based on filtering and sorting preferences
func (r *Repository) GetRecentEntriesWithOptions(days int, filterByFirstSeen bool, sortBy string) ([]Entry, error) {
    cutoff := time.Now().AddDate(0, 0, -days)

    // Choose filter field
    filterField := "e.published"
    if filterByFirstSeen {
        filterField = "e.first_seen"
    }

    // Choose sort field
    sortField := "e.published"
    if sortBy == "first_seen" {
        sortField = "e.first_seen"
    }

    query := fmt.Sprintf(`
        SELECT e.id, e.feed_id, e.entry_id, e.title, e.link, e.author,
               e.published, e.updated, e.content, e.content_type, e.summary, e.first_seen
        FROM entries e
        JOIN feeds f ON e.feed_id = f.id
        WHERE f.active = 1 AND %s >= ?
        ORDER BY %s DESC
    `, filterField, sortField)

    rows, err := r.db.Query(query, cutoff.Format(time.RFC3339))
    if err != nil {
        return nil, fmt.Errorf("query entries: %w", err)
    }
    defer rows.Close()

    entries, err := scanEntries(rows)
    if err != nil {
        return nil, err
    }

    // If we found entries, return them
    if len(entries) > 0 {
        return entries, nil
    }

    // Fallback to most recent 50 entries (use same sort field)
    query = fmt.Sprintf(`
        SELECT e.id, e.feed_id, e.entry_id, e.title, e.link, e.author,
               e.published, e.updated, e.content, e.content_type, e.summary, e.first_seen
        FROM entries e
        JOIN feeds f ON e.feed_id = f.id
        WHERE f.active = 1
        ORDER BY %s DESC
        LIMIT 50
    `, sortField)

    rows, err = r.db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("query fallback entries: %w", err)
    }
    defer rows.Close()

    return scanEntries(rows)
}
```

**Note**: Using `fmt.Sprintf` for field names is safe here because the values are controlled by config validation, not user input.

### 3. Generator Changes

**File**: `pkg/generator/generator.go`

Update to pass config options:

```go
func (g *Generator) Generate(cfg *config.Config, repo *repository.Repository) (string, error) {
    // Existing code...

    entries, err := repo.GetRecentEntriesWithOptions(
        cfg.Planet.Days,
        cfg.Planet.FilterByFirstSeen,
        cfg.Planet.SortBy,
    )

    // Rest of generation...
}
```

### 4. Command Changes

**File**: `cmd/rp/commands.go`

Update commands that generate HTML to pass config:

```go
func cmdGenerate(opts GenerateOptions) error {
    cfg, err := config.LoadFromFile(opts.ConfigPath)
    // ...

    html, err := gen.Generate(cfg, repo)  // Now passes full config
    // ...
}
```

## Testing Strategy

### Unit Tests (Time-Independent, No Network)

#### Test 1: Filter by First Seen

**File**: `pkg/repository/repository_test.go`

```go
func TestGetRecentEntriesFilterByFirstSeen(t *testing.T) {
    repo, cleanup := setupTestDB(t)
    defer cleanup()

    // Add a feed
    feedID, _ := repo.AddFeed("https://example.com/feed", "Test Feed")

    // Create entries with different published and first_seen dates
    // Use fixed time for deterministic testing
    baseTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

    entries := []struct {
        published time.Time  // Original published date
        firstSeen time.Time  // When aggregator saw it
    }{
        {baseTime.AddDate(0, 0, -30), baseTime.AddDate(0, 0, -1)}, // Old entry, recently seen
        {baseTime.AddDate(0, 0, -2), baseTime.AddDate(0, 0, -2)},  // Recent entry, recently seen
        {baseTime.AddDate(0, 0, -3), baseTime.AddDate(0, 0, -10)}, // Recent entry, seen long ago
    }

    for i, e := range entries {
        err := repo.UpsertEntry(&repository.Entry{
            FeedID:    feedID,
            EntryID:   fmt.Sprintf("entry-%d", i),
            Title:     fmt.Sprintf("Entry %d", i),
            Published: e.published,
            FirstSeen: e.firstSeen,
        })
        if err != nil {
            t.Fatalf("UpsertEntry() error = %v", err)
        }
    }

    // Test 1: Filter by published (default behavior)
    // Should return entries 1 and 2 (published within 7 days)
    publishedFiltered, err := repo.GetRecentEntriesWithOptions(7, false, "published")
    if err != nil {
        t.Fatalf("GetRecentEntriesWithOptions() error = %v", err)
    }
    if len(publishedFiltered) != 2 {
        t.Errorf("Filter by published: got %d entries, want 2", len(publishedFiltered))
    }

    // Test 2: Filter by first_seen
    // Should return entries 0 and 1 (first_seen within 7 days)
    firstSeenFiltered, err := repo.GetRecentEntriesWithOptions(7, true, "published")
    if err != nil {
        t.Fatalf("GetRecentEntriesWithOptions() error = %v", err)
    }
    if len(firstSeenFiltered) != 2 {
        t.Errorf("Filter by first_seen: got %d entries, want 2", len(firstSeenFiltered))
    }

    // Verify which entries were returned
    titles := make(map[string]bool)
    for _, e := range firstSeenFiltered {
        titles[e.Title] = true
    }

    if !titles["Entry 0"] || !titles["Entry 1"] {
        t.Errorf("Filter by first_seen returned wrong entries: %v", titles)
    }
    if titles["Entry 2"] {
        t.Errorf("Filter by first_seen should not include Entry 2 (first_seen too old)")
    }
}
```

#### Test 2: Sort by First Seen

```go
func TestGetRecentEntriesSortByFirstSeen(t *testing.T) {
    repo, cleanup := setupTestDB(t)
    defer cleanup()

    feedID, _ := repo.AddFeed("https://example.com/feed", "Test Feed")
    baseTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

    // Create entries where first_seen order differs from published order
    entries := []struct {
        title     string
        published time.Time
        firstSeen time.Time
    }{
        {"Entry A", baseTime.AddDate(0, 0, -1), baseTime.AddDate(0, 0, -3)}, // Published recently, seen first
        {"Entry B", baseTime.AddDate(0, 0, -2), baseTime.AddDate(0, 0, -2)}, // Published middle, seen second
        {"Entry C", baseTime.AddDate(0, 0, -3), baseTime.AddDate(0, 0, -1)}, // Published oldest, seen last
    }

    for i, e := range entries {
        err := repo.UpsertEntry(&repository.Entry{
            FeedID:    feedID,
            EntryID:   fmt.Sprintf("entry-%d", i),
            Title:     e.title,
            Published: e.published,
            FirstSeen: e.firstSeen,
        })
        if err != nil {
            t.Fatalf("UpsertEntry() error = %v", err)
        }
    }

    // Sort by published (default)
    byPublished, _ := repo.GetRecentEntriesWithOptions(7, false, "published")
    if byPublished[0].Title != "Entry A" {
        t.Errorf("Sort by published: first entry = %s, want Entry A", byPublished[0].Title)
    }

    // Sort by first_seen
    byFirstSeen, _ := repo.GetRecentEntriesWithOptions(7, false, "first_seen")
    if byFirstSeen[0].Title != "Entry C" {
        t.Errorf("Sort by first_seen: first entry = %s, want Entry C", byFirstSeen[0].Title)
    }
    if byFirstSeen[1].Title != "Entry B" {
        t.Errorf("Sort by first_seen: second entry = %s, want Entry B", byFirstSeen[1].Title)
    }
    if byFirstSeen[2].Title != "Entry A" {
        t.Errorf("Sort by first_seen: third entry = %s, want Entry A", byFirstSeen[2].Title)
    }
}
```

#### Test 3: Combined Filter and Sort

```go
func TestGetRecentEntriesFilterAndSortByFirstSeen(t *testing.T) {
    repo, cleanup := setupTestDB(t)
    defer cleanup()

    feedID, _ := repo.AddFeed("https://example.com/feed", "Test Feed")
    baseTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

    entries := []struct {
        title     string
        published time.Time
        firstSeen time.Time
    }{
        {"Recent discovery", baseTime.AddDate(0, 0, -30), baseTime.AddDate(0, 0, -1)}, // Old content, just discovered
        {"Recent post", baseTime.AddDate(0, 0, -1), baseTime.AddDate(0, 0, -1)},       // Recent content, recently discovered
        {"Old discovery", baseTime.AddDate(0, 0, -2), baseTime.AddDate(0, 0, -10)},    // Should be filtered out
    }

    for i, e := range entries {
        repo.UpsertEntry(&repository.Entry{
            FeedID:    feedID,
            EntryID:   fmt.Sprintf("entry-%d", i),
            Title:     e.title,
            Published: e.published,
            FirstSeen: e.firstSeen,
        })
    }

    // Filter by first_seen AND sort by first_seen
    results, _ := repo.GetRecentEntriesWithOptions(7, true, "first_seen")

    // Should have 2 entries (first_seen within 7 days)
    if len(results) != 2 {
        t.Fatalf("got %d entries, want 2", len(results))
    }

    // Should be sorted by first_seen DESC (both on day -1, order may vary by insert time)
    // Just verify the old discovery is not included
    for _, e := range results {
        if e.Title == "Old discovery" {
            t.Errorf("Old discovery should be filtered out (first_seen too old)")
        }
    }
}
```

### Integration Tests (Time-Independent, No Network)

#### Test 4: Entry Spam Prevention Workflow

**File**: `cmd/rp/entry_spam_integration_test.go`

```go
func TestEntrySpamPrevention(t *testing.T) {
    dir, cleanup := setupTestDir(t)
    defer cleanup()

    // Initialize planet with filter_by_first_seen
    initOpts := InitOptions{
        ConfigPath: "./config.ini",
        FeedsFile:  "",
        Output:     io.Discard,
    }
    if err := cmdInit(initOpts); err != nil {
        t.Fatalf("cmdInit() error = %v", err)
    }

    // Update config to enable first_seen filtering
    configContent := `[planet]
name = Test Planet
link = https://test.example.com
owner_name = Test Owner
filter_by_first_seen = true
sort_by = first_seen
days = 7

[database]
path = ./data/planet.db
`
    if err := os.WriteFile(filepath.Join(dir, "config.ini"), []byte(configContent), 0644); err != nil {
        t.Fatalf("WriteFile() error = %v", err)
    }

    // Open repository directly to inject test data
    repo, err := repository.New(filepath.Join(dir, "data/planet.db"))
    if err != nil {
        t.Fatalf("repository.New() error = %v", err)
    }
    defer repo.Close()

    // Simulate adding a feed with old entries
    feedID, _ := repo.AddFeed("https://example.com/feed", "Example Feed")

    baseTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

    // Feed has 3 entries:
    // 1. Published 30 days ago, first_seen today (simulates old entry from new feed)
    // 2. Published 2 days ago, first_seen today
    // 3. Published 3 days ago, first_seen 10 days ago (existing entry)
    entries := []repository.Entry{
        {
            FeedID:    feedID,
            EntryID:   "old-entry",
            Title:     "Old Entry (should be filtered)",
            Link:      "https://example.com/old",
            Published: baseTime.AddDate(0, 0, -30),
            FirstSeen: baseTime, // Just discovered
        },
        {
            FeedID:    feedID,
            EntryID:   "recent-entry",
            Title:     "Recent Entry (should appear)",
            Link:      "https://example.com/recent",
            Published: baseTime.AddDate(0, 0, -2),
            FirstSeen: baseTime, // Just discovered
        },
        {
            FeedID:    feedID,
            EntryID:   "existing-entry",
            Title:     "Existing Entry (should be filtered)",
            Link:      "https://example.com/existing",
            Published: baseTime.AddDate(0, 0, -3),
            FirstSeen: baseTime.AddDate(0, 0, -10), // Discovered long ago
        },
    }

    for _, e := range entries {
        if err := repo.UpsertEntry(&e); err != nil {
            t.Fatalf("UpsertEntry() error = %v", err)
        }
    }

    // Generate HTML
    genOpts := GenerateOptions{
        ConfigPath: "./config.ini",
        Output:     io.Discard,
    }
    if err := cmdGenerate(genOpts); err != nil {
        t.Fatalf("cmdGenerate() error = %v", err)
    }

    // Read generated HTML
    htmlContent, err := os.ReadFile(filepath.Join(dir, "public/index.html"))
    if err != nil {
        t.Fatalf("ReadFile() error = %v", err)
    }

    html := string(htmlContent)

    // Verify only the recent entry appears
    if !strings.Contains(html, "Recent Entry (should appear)") {
        t.Error("HTML should contain 'Recent Entry (should appear)'")
    }

    // Verify old entries are filtered out
    if strings.Contains(html, "Old Entry (should be filtered)") {
        t.Error("HTML should NOT contain 'Old Entry (should be filtered)' - entry spam not prevented!")
    }

    if strings.Contains(html, "Existing Entry (should be filtered)") {
        t.Error("HTML should NOT contain 'Existing Entry (should be filtered)' - first_seen filter failed!")
    }
}
```

#### Test 5: Backwards Compatibility

```go
func TestBackwardsCompatibility(t *testing.T) {
    dir, cleanup := setupTestDir(t)
    defer cleanup()

    // Initialize with default config (filter_by_first_seen = false)
    initOpts := InitOptions{
        ConfigPath: "./config.ini",
        Output:     io.Discard,
    }
    cmdInit(initOpts)

    repo, _ := repository.New(filepath.Join(dir, "data/planet.db"))
    defer repo.Close()

    feedID, _ := repo.AddFeed("https://example.com/feed", "Example Feed")
    baseTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

    // Add entry published recently but first_seen long ago
    repo.UpsertEntry(&repository.Entry{
        FeedID:    feedID,
        EntryID:   "test-entry",
        Title:     "Test Entry",
        Published: baseTime.AddDate(0, 0, -2),
        FirstSeen: baseTime.AddDate(0, 0, -30), // Seen 30 days ago
    })

    // Generate with default config (should filter by published)
    cmdGenerate(GenerateOptions{ConfigPath: "./config.ini", Output: io.Discard})

    htmlContent, _ := os.ReadFile(filepath.Join(dir, "public/index.html"))
    html := string(htmlContent)

    // Should appear because published date is recent (default behavior)
    if !strings.Contains(html, "Test Entry") {
        t.Error("Default behavior should filter by published date")
    }
}
```

### Key Testing Patterns

#### 1. Time Independence

**Problem**: Tests that depend on `time.Now()` are non-deterministic.

**Solution**: Use fixed times in test data:

```go
baseTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
recentEntry := baseTime.AddDate(0, 0, -2)  // 2 days ago from baseTime
oldEntry := baseTime.AddDate(0, 0, -30)    // 30 days ago from baseTime
```

**Alternative**: Mock time (more complex):

```go
// In production code, use time.Now via interface
type Clock interface {
    Now() time.Time
}

// In tests, use fixed clock
type FixedClock struct {
    time time.Time
}

func (c *FixedClock) Now() time.Time {
    return c.time
}
```

#### 2. Network Independence

**Problem**: Integration tests shouldn't fetch real feeds.

**Solution**: Inject test data directly into database:

```go
// Don't use rp fetch (requires network)
// Instead, directly insert test entries with UpsertEntry()

repo.UpsertEntry(&repository.Entry{
    FeedID:    feedID,
    EntryID:   "test-id",
    Title:     "Test Entry",
    Published: testTime,
    FirstSeen: testTime,
})
```

#### 3. Isolation

**Problem**: Tests interfere with each other.

**Solution**: Use `t.TempDir()` and `setupTestDir()`:

```go
func TestSomething(t *testing.T) {
    dir, cleanup := setupTestDir(t)  // Creates temp dir, chdirs to it
    defer cleanup()                   // Restores original dir, cleans up

    // All file operations are in isolated temp directory
    // Database is ./data/planet.db in temp dir
}
```

## Edge Cases to Consider

### 1. Clock Skew

**Scenario**: Entry has `first_seen` in the future due to clock skew.

**Handling**: Should still work - future entries are simply very recent.

**Test**:
```go
futureEntry := repository.Entry{
    FirstSeen: time.Now().Add(1 * time.Hour), // 1 hour in future
}
// Should appear in results (within time window)
```

### 2. Bulk Import

**Scenario**: User imports 100 feeds via OPML, all fetch at once.

**Desired Behavior**: With `filter_by_first_seen = true`, all entries appear but only if their `first_seen` is recent.

**Test**: Import OPML with multiple feeds, verify only recent discoveries appear.

### 3. Feed Updates Entry

**Scenario**: Entry is updated (corrected typo), `first_seen` stays the same.

**Desired Behavior**: Entry stays in same position in timeline (stable sort).

**Implementation**: `UpsertEntry()` already preserves `first_seen` on updates:
```sql
ON CONFLICT(feed_id, entry_id) DO UPDATE SET
    -- first_seen is NOT updated
```

### 4. Zero Entries Case

**Scenario**: All entries filtered out by `first_seen` filter.

**Desired Behavior**: Fall back to 50 most recent entries (by `first_seen`).

**Already Implemented**: `GetRecentEntriesWithOptions()` has fallback logic.

## Migration Path

### For Existing Planets

Users upgrading from earlier versions won't have `first_seen` data for existing entries.

**Solution**: Backfill `first_seen` from `published`:

```sql
-- Migration query (run once)
UPDATE entries
SET first_seen = published
WHERE first_seen IS NULL;
```

**Where to Run**: In `repository.initSchema()` after creating tables:

```go
func (r *Repository) initSchema() error {
    // Create tables...

    // Backfill first_seen for entries that don't have it
    _, err := r.db.Exec(`
        UPDATE entries
        SET first_seen = published
        WHERE first_seen IS NULL OR first_seen = ''
    `)
    if err != nil {
        return fmt.Errorf("backfill first_seen: %w", err)
    }

    return nil
}
```

## Documentation Updates

### 1. Config Documentation

Add to `config.ini` example:

```ini
[planet]
# Prevent entry spam when adding new feeds
# When true, only show entries first discovered within the time window
# Default: false (backwards compatible)
filter_by_first_seen = false

# Sort entries by published date or first seen date
# "published": Original published date from feed (default)
# "first_seen": When Rogue Planet first discovered the entry
# Default: published
sort_by = published
```

### 2. QUICKSTART.md

Add section:

```markdown
## Preventing Entry Spam

When you add a new feed, you may not want all its historical entries to flood your planet's timeline. Enable first-seen filtering:

```bash
# Edit config.ini
[planet]
filter_by_first_seen = true   # Only show recently discovered entries
sort_by = first_seen           # Optional: sort by discovery order

# Now when you add feeds, only entries discovered within your
# time window (e.g., last 7 days) will appear
rp add-feed https://example.com/feed.xml
rp update
```

### 3. WORKFLOWS.md

Add to "Feed Management" section:

```markdown
**Adding feeds without timeline spam:**
```bash
# Configure before adding feeds
vim config.ini
# Set: filter_by_first_seen = true

# Add new feed
rp add-feed https://newblog.com/feed.xml

# Update - only entries discovered today will appear
rp update
```

## Success Criteria

Implementation is successful when:

1. ✅ Config options validate correctly
2. ✅ `filter_by_first_seen = true` filters by `first_seen` date
3. ✅ `sort_by = "first_seen"` sorts by `first_seen` date
4. ✅ Default behavior unchanged (backwards compatible)
5. ✅ All unit tests pass (time-independent, no network)
6. ✅ All integration tests pass (time-independent, no network)
7. ✅ Existing entries backfilled with `first_seen = published`
8. ✅ Documentation updated

## References

- **Original complaint**: [Reddit comment by /u/fgilcher](https://www.reddit.com/r/rust/comments/mvm4r2/comment/gvcwxq0/) about Venus/Planet spam
- **Database schema**: `pkg/repository/repository.go` lines 114-134
- **Current query**: `pkg/repository/repository.go` lines 272-311
- **Config validation**: `pkg/config/config.go`
- **Test patterns**: `cmd/rp/opml_integration_test.go` (setupTestDir example)

---

**Document Version**: 1.0
**Date**: 2025-10-14
**Status**: Proposed Implementation
