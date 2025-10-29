# Rogue Planet Quick Start

Get your planet aggregator running in 5 minutes!

## Installation

```bash
go install github.com/adewale/rogue_planet/cmd/rp@latest
```

## First-Time Setup

```bash
# 1. Create a directory
mkdir ~/my-planet && cd ~/my-planet

# 2. Create feeds.txt with your favorite blogs
cat > feeds.txt <<'EOF'
https://blog.golang.org/feed.atom
https://github.blog/feed/
https://daringfireball.net/feeds/main
EOF

# 3. Initialize
rp init -f feeds.txt

# 4. Fetch and generate
rp update

# 5. View result
open public/index.html
```

## Daily Commands

| Command | What It Does |
|---------|-------------|
| `rp update` | Fetch all feeds and regenerate HTML (use this most) |
| `rp status` | Show feed and entry counts |
| `rp list-feeds` | Show all configured feeds |
| `rp add-feed <url>` | Add a new feed |
| `rp remove-feed <url>` | Remove a feed |
| `rp import-opml <file>` | Import feeds from OPML file |
| `rp export-opml` | Export feeds to OPML format |
| `rp verify` | Validate configuration and environment |
| `rp fetch` | Fetch feeds without generating HTML |
| `rp generate` | Regenerate HTML without fetching |
| `rp prune --days 90` | Remove old entries from database |

## Automated Updates (Cron)

```bash
# Edit crontab
crontab -e

# Add this line (updates every 30 minutes):
*/30 * * * * cd /home/user/my-planet && rp update >> update.log 2>&1
```

## Configuration

Edit `config.ini` to customize your planet:

```ini
[planet]
name = My Planet
link = https://planet.example.com
owner_name = Your Name
days = 7                    # Days of entries to show
concurrent_fetches = 5      # Parallel feed fetching
group_by_date = true        # Group entries by date

[database]
path = ./data/planet.db
```

## Troubleshooting

**Feed not updating?**
```bash
rp list-feeds              # Check feed status
rp fetch                   # Try manual fetch
rp remove-feed <url>       # Remove problematic feed
rp add-feed <url>          # Re-add it
```

**HTML not updating?**
```bash
rp status                  # Check if entries exist
rp generate                # Force regeneration
ls -l public/index.html    # Check file timestamp
```

**Database too large?**
```bash
rp prune --days 30         # Remove old entries
sqlite3 data/planet.db "VACUUM;"  # Reclaim space
```

## File Structure

```
my-planet/
├── config.ini             # Configuration
├── feeds.txt              # Feed list (optional)
├── data/
│   └── planet.db         # SQLite database
└── public/
    └── index.html        # Generated HTML
```

## Next Steps

- **More examples**: See [WORKFLOWS.md](WORKFLOWS.md)
- **Full documentation**: See [README.md](README.md)
- **Development guide**: See [CLAUDE.md](CLAUDE.md)
- **Report issues**: https://github.com/adewale/rogue_planet/issues

## Common Workflows

**Add multiple feeds at once:**
```bash
cat > new-feeds.txt <<'EOF'
https://example.com/feed1.xml
https://example.com/feed2.xml
EOF
rp add-all -f new-feeds.txt
```

**Import feeds from another reader:**
```bash
# Most feed readers export to OPML (Feedly, Inoreader, NewsBlur, etc.)
rp import-opml feedly-export.opml  # Preview first with --dry-run
rp update  # Fetch the new feeds
```

**Backup and restore your feed list:**
```bash
# Export to OPML
rp export-opml --output backup.opml

# Later, restore from backup
rp import-opml backup.opml
```

**Validate your configuration:**
```bash
rp verify  # Checks config, database, output directory, template
```

**Check for broken feeds:**
```bash
sqlite3 data/planet.db \
  "SELECT url, fetch_error_count FROM feeds WHERE fetch_error_count > 0;"
```

**Monitor update logs:**
```bash
tail -f update.log
```

**Backup your planet:**
```bash
tar czf planet-backup.tar.gz config.ini data/planet.db feeds.txt
```

---

That's it! You now have a working feed aggregator. Run `rp update` regularly (via cron) to keep it fresh.
