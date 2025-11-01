# Remove Feed Specification

## Overview

The `remove-feed` command removes a feed from the Rogue Planet aggregator. When a feed is removed, all its associated entries are automatically deleted via CASCADE DELETE, and the feed will no longer be fetched or displayed in the generated output.

## Command Syntax

```bash
rp remove-feed <url> [--config <path>] [--force]
```

## Parameters

### Required

- `<url>` - The exact URL of the feed to remove
  - Must match the URL stored in the database exactly
  - Case-sensitive
  - Example: `https://blog.golang.org/feed.atom`
  - **Note**: If a feed has been redirected (301), use the current URL shown by `rp list-feeds`

### Optional Flags

- `--config <path>` - Path to configuration file (default: `./config.ini`)
- `--force` - Skip confirmation prompt (useful for scripting)

## Behavior

### Interactive Mode (Default)

When run without `--force`, the command prompts for confirmation:

1. **Feed Lookup**: Find the feed by URL in the database
2. **Entry Count**: Query the number of entries associated with the feed
3. **Confirmation Prompt**: Display feed details and ask for confirmation
4. **User Response**:
   - If user enters `y` or `yes` (case-insensitive, whitespace trimmed): Proceed with removal
   - If user enters anything else or just presses Enter: Cancel operation
   - If stdin is not a terminal (piped/scripted): Error and exit
5. **Cascade Delete**: Remove the feed record (automatically deletes all associated entries via `ON DELETE CASCADE`)
6. **Confirmation**: Display success message with entry count
7. **Exit**: Return exit code 0

**Example Output:**
```
Feed: https://example.com/feed.xml
Title: Example Blog
Entries: 152

Remove this feed and all 152 entries? (y/N): y
✓ Removed feed: https://example.com/feed.xml (152 entries deleted)
```

**Input Handling:**
- Accepted inputs (case-insensitive): `y`, `yes`, ` y `, ` YES `
- Leading/trailing whitespace is trimmed
- Default (Enter with no input): Cancel
- Any other input: Cancel

### Force Mode (`--force`)

Skips confirmation prompt (useful for scripting):

1. **Feed Lookup**: Find the feed by URL in the database
2. **Entry Count**: Query the number of entries associated with the feed (for success message)
3. **Cascade Delete**: Immediately remove the feed and entries
4. **Confirmation**: Display success message with entry count
5. **Exit**: Return exit code 0

**Example Output:**
```
✓ Removed feed: https://example.com/feed.xml (152 entries deleted)
```

### Error Cases

#### 1. Missing URL Parameter

**Command:** `rp remove-feed`

**Error:** URL is required

**Exit Code:** 1

#### 2. Feed Not Found

**Command:** `rp remove-feed https://nonexistent.com/feed.xml`

**Error:** Feed not found: [database error details]

**Exit Code:** 1

**Note:** This error occurs when:
- The URL doesn't exist in the database
- The URL format doesn't match exactly (e.g., http vs https, trailing slash differences)
- The feed was already removed
- **The feed URL was updated due to a 301 redirect** (see Redirect Handling section)

#### 3. User Cancelled

**Command:** `rp remove-feed https://example.com/feed.xml` (then enters 'n')

**Output:**
```
Feed: https://example.com/feed.xml
Title: Example Blog
Entries: 152

Remove this feed and all 152 entries? (y/N): n
Cancelled.
```

**Exit Code:** 1 (operation not completed)

**Note:** Exit code 1 prevents subsequent commands in chains from running:
```bash
rp remove-feed https://example.com/feed.xml && rp generate
# User enters 'n'
# Cancelled. (exit 1)
# generate will NOT run ✓
```

#### 4. Non-Interactive Without --force

**Command:** `echo "y" | rp remove-feed https://example.com/feed.xml`

**Error:** Cannot prompt for confirmation in non-interactive mode. Use --force to skip confirmation.

**Exit Code:** 1

**Note:** This prevents accidental deletions in scripts that don't explicitly use `--force`

#### 5. Database Error

**Command:** `rp remove-feed https://example.com/feed.xml`

**Error:** Failed to remove feed: [database error details]

**Exit Code:** 1

**Note:** This error can occur due to:
- Database file is locked
- Database file is corrupted
- Insufficient permissions
- Disk full

#### 4. Configuration Error

**Command:** `rp remove-feed https://example.com/feed.xml --config /nonexistent/config.ini`

**Error:** [configuration loading error]

**Exit Code:** 1

## Redirect Handling

### Overview

Rogue Planet automatically updates feed URLs when it encounters **301 Moved Permanently** redirects during fetch operations. This creates scenarios where the URL you originally added differs from the URL stored in the database, which affects feed removal.

### How Redirects Are Handled

#### During Fetch Operations

1. **301 Permanent Redirect**: The crawler detects 301 responses and updates the feed URL in the database
   - Old URL: `http://blog.example.com/rss`
   - Server returns: `301 Moved Permanently` → `https://blog.example.com/feed`
   - Database updated: Stores `https://blog.example.com/feed`
   - Cache headers (ETag, Last-Modified) are cleared (associated with old URL)

2. **302 Temporary Redirect**: The URL is NOT updated
   - Feed URL remains as originally added
   - Redirects are followed for fetching but don't modify the database

#### Impact on Remove-Feed Command

The `remove-feed` command requires an **exact URL match**. After a 301 redirect:

**❌ This will fail:**
```bash
$ rp add-feed http://blog.example.com/rss
✓ Added feed: http://blog.example.com/rss (ID: 1)

$ rp update
# ... 301 redirect detected, URL updated to https://blog.example.com/feed

$ rp remove-feed http://blog.example.com/rss
Error: feed not found: feed not found
```

**✓ This will succeed:**
```bash
$ rp list-feeds
Configured feeds (1):

  [1] https://blog.example.com/feed
      Status: active

$ rp remove-feed https://blog.example.com/feed
Feed: https://blog.example.com/feed
Title: Example Blog
Entries: 152

Remove this feed and all 152 entries? (y/N): y
✓ Removed feed: https://blog.example.com/feed (152 entries deleted)
```

### Redirect Scenarios

#### Scenario 1: Feed URL Changed After Adding (Most Common)

**Timeline:**
1. Day 1: User adds feed with URL `http://old.example.com/feed`
2. Day 2: `rp update` runs, detects 301 redirect to `https://new.example.com/feed`
3. Day 3: User wants to remove feed using original URL

**Result:** Removal fails with "feed not found"

**Solution:**
1. Run `rp list-feeds` to see current URLs
2. Use the current URL shown in the list

#### Scenario 2: Multiple Redirects in Chain

**Server behavior:**
- `http://blog.com/rss` → 301 → `http://blog.com/feed` → 301 → `https://blog.com/feed`

**Rogue Planet behavior:**
- Follows all redirects (up to 5 maximum)
- Stores only the **final** URL: `https://blog.com/feed`
- All intermediate URLs are lost

**Impact:** Can only remove using the final URL

#### Scenario 3: Redirect After Some Time

**Timeline:**
1. Feed works fine for months with URL `http://example.com/feed.xml`
2. Site owner migrates to HTTPS
3. Next update detects 301 redirect to `https://example.com/feed.xml`
4. Database updated automatically

**User experience:** May be surprised that their original URL doesn't work for removal

### Best Practices for Feed Management

1. **Always check current URLs before removing:**
   ```bash
   rp list-feeds  # Shows current URLs
   rp remove-feed <url-from-list>
   ```

2. **Use dry-run to verify before removing:**
   ```bash
   rp remove-feed https://example.com/feed --dry-run
   # Confirms you're removing the right feed
   ```

3. **Export OPML before bulk removals:**
   ```bash
   rp export-opml --output backup.opml
   # Then perform removals
   ```

4. **Check update logs for URL changes:**
   ```bash
   rp update --verbose
   # Look for: "Feed https://old.url permanently redirected to https://new.url (301)"
   ```

### Why Not Store URL History?

The current implementation does NOT store old URLs because:

1. **Simplicity**: Single URL per feed keeps schema simple
2. **Storage**: Avoid unbounded growth from repeated redirects
3. **Clarity**: Database reflects current state, not history
4. **Performance**: No need to search multiple URL tables

**Trade-off:** Users must use `list-feeds` to find current URLs for removal.

### Future Enhancements for Redirect Handling

Potential improvements (not in initial version):

1. **Fuzzy URL Matching**: Search for similar URLs when exact match fails
2. **URL History Table**: Store previous URLs for lookup
3. **Better Error Messages**: "Feed not found. Did you mean https://new.url?"
4. **Undo URL Update**: Command to revert to previous URL if redirect was temporary
5. **Remove by Feed Title**: `rp remove-feed --title "Example Blog"`

## Database Impact

### CASCADE DELETE Behavior

The database schema uses `ON DELETE CASCADE` for the foreign key relationship:

```sql
CREATE TABLE entries (
    ...
    feed_id INTEGER NOT NULL,
    ...
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);
```

When a feed is deleted:
- **Automatic**: All entries with matching `feed_id` are deleted
- **Transactional**: If entry deletion fails, feed deletion is rolled back
- **Efficient**: No manual cleanup required

### Example Impact

Feed with 100 entries:
- **Before removal**: Database contains 1 feed record, 100 entry records
- **After removal**: Database contains 0 feed records, 0 entry records (for that feed)
- **Query**: Single `DELETE FROM feeds WHERE id = ?` triggers cascade

## Usage Examples

### Example 1: Interactive Removal (Default)

```bash
$ rp list-feeds
Configured feeds (3):

  [1] https://blog.golang.org/feed.atom
      Status: active

  [2] https://example.com/feed.xml
      Status: active

  [3] https://another.com/rss
      Status: active

$ rp remove-feed https://example.com/feed.xml
Feed: https://example.com/feed.xml
Title: Example Blog
Entries: 152

Remove this feed and all 152 entries? (y/N): y
✓ Removed feed: https://example.com/feed.xml (152 entries deleted)

$ rp list-feeds
Configured feeds (2):

  [1] https://blog.golang.org/feed.atom
      Status: active

  [2] https://another.com/rss
      Status: active
```

### Example 2: Force Mode for Scripting

```bash
#!/bin/bash
# Remove multiple feeds without prompts
for url in spam1.com spam2.com spam3.com; do
  rp remove-feed "https://$url/feed.xml" --force
done
```

Output:
```
✓ Removed feed: https://spam1.com/feed.xml (43 entries deleted)
✓ Removed feed: https://spam2.com/feed.xml (87 entries deleted)
✓ Removed feed: https://spam3.com/feed.xml (12 entries deleted)
```

### Example 3: User Cancels Removal

```bash
$ rp remove-feed https://important-blog.com/feed.xml
Feed: https://important-blog.com/feed.xml
Title: Important Blog
Entries: 523

Remove this feed and all 523 entries? (y/N): n
Cancelled.

$ rp list-feeds
# Feed still present
```

### Example 4: Handle Feed After Redirect

```bash
# Original URL no longer works
$ rp remove-feed http://blog.example.com/rss
Error: feed not found: feed not found

# Check current URLs
$ rp list-feeds
Configured feeds (1):

  [1] https://blog.example.com/feed  ← URL was updated by 301 redirect
      Status: active

# Use current URL
$ rp remove-feed https://blog.example.com/feed
Feed: https://blog.example.com/feed
Title: Example Blog
Entries: 152

Remove this feed and all 152 entries? (y/N): y
✓ Removed feed: https://blog.example.com/feed (152 entries deleted)
```

### Example 5: Non-Interactive Error

```bash
$ echo "y" | rp remove-feed https://example.com/feed.xml
Error: Cannot prompt for confirmation in non-interactive mode. Use --force to skip confirmation.

$ echo $?
1

# Correct usage in scripts:
$ rp remove-feed https://example.com/feed.xml --force
✓ Removed feed: https://example.com/feed.xml (152 entries deleted)
```

### Example 6: Remove with Custom Config

```bash
$ rp remove-feed https://example.com/feed.xml --config ~/planet/config.ini --force
✓ Removed feed: https://example.com/feed.xml (152 entries deleted)
```

## Implementation Details

### Command Flow

#### Interactive Mode (Default)

1. **Validate Input**
   - Check URL parameter is provided
   - Parse flags: `--force`, `--config`

2. **Open Configuration and Database**
   - Load config from `--config` path or default `./config.ini`
   - Open database connection
   - Defer cleanup (close database)

3. **Lookup Feed**
   - Call `repo.GetFeedByURL(url)`
   - Return error if not found

4. **Count Entries**
   - Call `repo.GetEntryCountForFeed(feed.ID)`
   - Store count for display

5. **Handle Interactive Confirmation**
   - If `--force` flag is NOT set:
     - Check if stdin is a terminal:
       ```go
       stat, _ := os.Stdin.Stat()
       isTerminal := (stat.Mode() & os.ModeCharDevice) != 0
       ```
     - If not a terminal: Return error "Cannot prompt in non-interactive mode. Use --force to skip confirmation."
     - Display feed info: URL, Title, Entry count
     - Prompt: "Remove this feed and all N entries? (y/N): "
     - Read user input from stdin
     - Trim whitespace and convert to lowercase
     - If input is "y" or "yes": Continue to step 6
     - Otherwise: Print "Cancelled." and return exit code 1

6. **Remove Feed**
   - Call `repo.RemoveFeed(feed.ID)`
   - CASCADE DELETE automatically removes all entries
   - Return error if deletion fails

7. **Confirm Success**
   - Print success message with URL and entry count
   - Format: "✓ Removed feed: <url> (N entries deleted)"
   - Return exit code 0

#### Force Mode Flow

Same as interactive but skips step 5 (confirmation prompt)

### Code Location

- **Command Handler**: `cmd/rp/commands.go:cmdRemoveFeed()`
- **Repository Method**: `pkg/repository/repository.go:RemoveFeed()`
- **Database Schema**: `pkg/repository/repository.go:initSchema()`
- **Tests**:
  - `cmd/rp/commands_test.go:TestCmdRemoveFeed()`
  - `pkg/repository/repository_test.go:TestRemoveFeed()`
  - `pkg/repository/repository_test.go:TestRemoveFeedCascade()`

## Testing

### Required Test Coverage

#### Unit Tests

1. **Missing URL Test**
   - Verifies error when URL parameter is empty
   - Location: `cmd/rp/commands_test.go:TestCmdRemoveFeed()`

2. **Basic Removal Test**
   - Adds a feed, removes it, verifies it's gone
   - Location: `pkg/repository/repository_test.go:TestRemoveFeed()`

3. **Cascade Delete Test**
   - Adds feed with entries, removes feed, verifies entries deleted
   - Location: `pkg/repository/repository_test.go:TestRemoveFeedCascade()`

4. **Entry Count Test** (NEW)
   - Verifies correct count is returned for feeds with varying entry counts
   - Should test: 0 entries, 1 entry, many entries
   - Location: `pkg/repository/repository_test.go:TestGetEntryCountForFeed()`

5. **Interactive Confirmation Test** (NEW)
   - Mock stdin with "y" input → should proceed with removal
   - Mock stdin with "n" input → should cancel
   - Mock stdin with "yes" input → should proceed
   - Mock stdin with "no" input → should cancel
   - Location: `cmd/rp/commands_test.go:TestCmdRemoveFeedConfirmation()`

6. **Force Flag Test** (NEW)
   - Verifies `--force` skips confirmation
   - Should not read from stdin
   - Location: `cmd/rp/commands_test.go:TestCmdRemoveFeedForce()`

7. **Non-Interactive Error Test** (NEW)
   - Verifies error when stdin is not a terminal and no `--force`
   - Mock non-terminal stdin
   - Location: `cmd/rp/commands_test.go:TestCmdRemoveFeedNonInteractive()`

#### Integration Tests

8. **Redirect Then Remove Test** (NEW - CRITICAL)
   - Add feed with URL A
   - Mock 301 redirect to URL B
   - Fetch feed (triggers URL update)
   - Attempt remove with URL A → should fail
   - Attempt remove with URL B → should succeed
   - Location: `cmd/rp/integration_test.go:TestRemoveFeedAfterRedirect()`

9. **Update Feed URL Test** (NEW - CRITICAL)
   - Test `UpdateFeedURL` repository method
   - Verify old URL no longer findable
   - Verify new URL is findable
   - Verify ETag and Last-Modified cleared
   - Location: `pkg/repository/repository_test.go:TestUpdateFeedURL()`

### Manual Testing Checklist

```bash
# Setup test environment
rp init
rp add-feed https://example.com/feed1.xml
rp add-feed https://example.com/feed2.xml
rp update

# Test 1: Interactive confirmation (accept)
rp remove-feed https://example.com/feed1.xml
# Enter: y
# Verify: Feed removed

# Test 2: Interactive confirmation (reject)
rp remove-feed https://example.com/feed2.xml
# Enter: n
# Verify: Feed NOT removed

# Test 3: Force mode
rp add-feed https://example.com/feed3.xml
rp remove-feed https://example.com/feed2.xml --force
# Verify: Feed removed without prompt

# Test 4: Non-interactive error
echo "y" | rp remove-feed https://example.com/feed3.xml
# Verify: Error about non-interactive mode

# Test 5: Non-interactive with force
echo "y" | rp remove-feed https://example.com/feed3.xml --force
# Verify: Feed removed

# Test 6: Error cases
rp remove-feed  # Should error: missing URL
rp remove-feed https://never-added.com/feed.xml  # Should error: not found

# Test 7: Redirect handling (manual mock)
# Add feed, manually update URL in database, try to remove with old URL
# Verify: Error "feed not found"
# Try with new URL: Success
```

## Security Considerations

### 1. URL Validation

- URLs are not validated or sanitized before lookup
- Database uses parameterized queries (safe from SQL injection)
- No SSRF risk (only database lookup, no network requests)

### 2. Authorization

- No authentication/authorization mechanism
- Any user with filesystem access can remove feeds
- **Recommendation**: Restrict config file permissions (0600)

### 3. Data Loss Prevention

- **Destructive Operation**: Requires explicit confirmation by default
- **Irreversible**: Removed feeds and entries cannot be recovered
- **Protections**:
  - Interactive confirmation prompt shows entry count
  - `--dry-run` flag for safe preview before removal
  - `--force` must be explicitly provided for scripting
  - Non-interactive detection prevents accidental pipe usage
- **Mitigation**:
  - Users should backup database before bulk removals
  - Export OPML before major changes: `rp export-opml --output backup.opml`
  - Use `--dry-run` to verify what will be deleted

## Integration Points

### 1. Database (pkg/repository)

- **Reads**: `GetFeedByURL()` - Lookup feed by URL
- **Writes**: `RemoveFeed()` - Delete feed record
- **Side Effects**: CASCADE DELETE removes all associated entries

### 2. Configuration (pkg/config)

- **Reads**: Config file location (for database path)
- **No Modifications**: Config file itself is not modified

### 3. Generated Output

- **No Direct Impact**: Doesn't regenerate HTML automatically
- **Recommendation**: Run `rp generate` after removing feeds to update output

## Future Enhancements

### Potential Improvements (Not in Initial Version)

1. **Bulk Removal from File**
   ```bash
   rp remove-feeds -f remove-list.txt
   # Remove multiple feeds from a file
   # Each line should contain a URL
   ```

2. **Soft Delete / Archive**
   - Mark feed as inactive instead of deleting
   - Keep entries but don't fetch updates
   - Allow restoration with `rp restore-feed`
   - Useful for temporary subscription pauses

3. **Remove by ID**
   ```bash
   rp remove-feed --id 5
   # Remove feed by database ID instead of URL
   # Useful when exact URL is unknown
   ```

4. **Remove by Title**
   ```bash
   rp remove-feed --title "Example Blog"
   # Remove feed by matching title
   # Useful for fuzzy matching
   ```

5. **URL History / Fuzzy Matching**
   - Store previous URLs when 301 redirects occur
   - Allow removal by any historical URL
   - Suggest similar URLs when exact match fails

6. **Batch Operations with Confirmation**
   ```bash
   rp remove-feed pattern*.com --pattern
   # Remove all feeds matching pattern
   # Single confirmation for all matches
   ```

7. **Undo Support**
   ```bash
   rp remove-feed https://example.com/feed.xml
   # ... realize mistake ...
   rp undo
   # Restores last removed feed and its entries
   ```

## Related Commands

- `rp add-feed` - Add a feed to the planet
- `rp list-feeds` - List all configured feeds
- `rp status` - Show feed and entry counts
- `rp prune` - Remove old entries (keeps feeds)
- `rp import-opml` - Import feeds from OPML file
- `rp export-opml` - Export current feeds to OPML

## Comparison: remove-feed vs prune

| Feature | remove-feed | prune |
|---------|------------|-------|
| **Purpose** | Remove specific feeds | Remove old entries |
| **Target** | Feed + all entries | Entries only (all feeds) |
| **Selector** | By URL | By age (days) |
| **Scope** | Single feed | All feeds |
| **Feed Impact** | Feed deleted | Feeds preserved |
| **Reversible** | No | No |
| **Use Case** | Unsubscribe from feed | Clean up old content |

## Appendix: Database Schema Reference

```sql
-- Feeds table
CREATE TABLE feeds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT UNIQUE NOT NULL,
    title TEXT,
    link TEXT,
    last_fetched TEXT,
    etag TEXT,
    last_modified TEXT,
    fetch_error TEXT,
    fetch_error_count INTEGER DEFAULT 0,
    next_fetch TEXT,
    active INTEGER DEFAULT 1
);

-- Entries table with CASCADE DELETE
CREATE TABLE entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_id INTEGER NOT NULL,
    entry_id TEXT NOT NULL,
    title TEXT,
    link TEXT,
    author TEXT,
    published TEXT,
    updated TEXT,
    content TEXT,
    content_type TEXT,
    summary TEXT,
    first_seen TEXT,
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
    UNIQUE(feed_id, entry_id)
);
```

When `DELETE FROM feeds WHERE id = ?` executes:
1. SQLite identifies all entries where `feed_id = ?`
2. SQLite deletes those entries automatically
3. SQLite deletes the feed record
4. Transaction commits (or rolls back if any step fails)

---

## Summary

### Key Features (Initial Version)

✅ **Interactive Confirmation**
- Prompts user before deletion
- Shows feed title and entry count
- Accepts `y/yes` to proceed, anything else to cancel

✅ **Force Mode (`--force`)**
- Skips confirmation for scripting
- Must be explicitly provided
- Shows entry count in success message

✅ **Dry-Run Mode (`--dry-run`)**
- Preview what would be deleted
- Shows feed details and entry count
- No changes made to database

✅ **Redirect Awareness**
- Documents behavior with 301 redirects
- Comprehensive error handling
- Best practices for URL management

✅ **Safety Protections**
- Non-interactive detection (prevents accidental pipe usage)
- Entry count display (shows impact)
- CASCADE DELETE (automatic cleanup)

### Critical Scenarios Addressed

1. **301 Redirect URL Changes**
   - User aware that URL may have changed
   - `list-feeds` shows current URLs
   - Error message references redirect possibility

2. **Data Loss Prevention**
   - Confirmation required by default
   - Dry-run for safe preview
   - Entry count visible before deletion

3. **Scripting Support**
   - `--force` flag for automation
   - Predictable exit codes
   - Non-interactive detection with helpful error

4. **User Experience**
   - Clear confirmation prompts
   - Informative error messages
   - Cancellation is easy (default is NO)

### Implementation Checklist

- [ ] Add `--force` and `--dry-run` flags to command parser
- [ ] Implement `GetEntryCountForFeed(feedID)` repository method
- [ ] Add interactive confirmation with stdin detection
- [ ] Update success message to include entry count
- [ ] Handle cancellation (exit code 0)
- [ ] Detect non-interactive mode without `--force`
- [ ] Write comprehensive tests (10 new tests required)
- [ ] Update CLI help text
- [ ] Document redirect scenarios in user docs

### Testing Priorities

**High Priority (Must Have)**
1. Interactive confirmation (y/n/yes/no)
2. Force flag (skips confirmation)
3. Dry-run mode (no deletion)
4. Non-interactive detection
5. Redirect then remove integration test
6. UpdateFeedURL repository test

**Medium Priority (Should Have)**
7. Entry count accuracy
8. Cancellation behavior
9. Multiple redirect chains
10. Error message clarity

**Low Priority (Nice to Have)**
- Edge cases (empty strings, special characters in input)
- Concurrent removal attempts
- Very large entry counts (>10,000)

### Known Limitations

1. **No URL History**: Old URLs after redirects are not stored
2. **Exact Match Required**: Fuzzy matching not supported
3. **No Undo**: Deletion is permanent
4. **Single Feed**: No bulk removal command
5. **No Soft Delete**: Feed is completely removed, not archived

These limitations are documented and can be addressed in future versions if user demand exists.
