# Rogue Planet Workflows

This document provides detailed workflows for common Rogue Planet operations.

## Table of Contents

- [Getting Started](#getting-started)
- [Daily Operations](#daily-operations)
- [Feed Management](#feed-management)
- [Troubleshooting](#troubleshooting)
- [Advanced Usage](#advanced-usage)
- [Production Deployment](#production-deployment)

## Getting Started

### First-Time Setup (5 minutes)

This workflow takes you from zero to a working planet aggregator:

```bash
# 1. Install Rogue Planet
go install github.com/adewale/rogue_planet/cmd/rp@latest

# 2. Create a directory for your planet
mkdir ~/my-planet
cd ~/my-planet

# 3. Create a feeds file with your favorite blogs
cat > feeds.txt <<'EOF'
https://blog.golang.org/feed.atom
https://github.blog/feed/
https://daringfireball.net/feeds/main
https://asymco.com/feed/
https://xkcd.com/rss.xml
EOF

# 4. Initialize the planet
rp init -f feeds.txt

# 5. Customize your planet (optional)
cat > config.ini <<'EOF'
[planet]
name = My Tech Planet
link = https://planet.example.com
owner_name = Your Name
owner_email = you@example.com
output_dir = ./public
days = 7
log_level = info
concurrent_fetches = 5
group_by_date = true

[database]
path = ./data/planet.db
EOF

# 6. Fetch feeds and generate your planet
rp update

# 7. View your planet
open public/index.html  # macOS
# Or: xdg-open public/index.html  # Linux
# Or: start public/index.html     # Windows
```

**What happened:**
- `rp init -f feeds.txt` created the directory structure and imported 5 feeds
- `rp update` fetched all 5 feeds, parsed them, and generated HTML
- You now have a working planet aggregator!

**Next steps:**
- Add more feeds with `rp add-feed <url>`
- Set up a cron job to update automatically (see [Daily Operations](#daily-operations))
- Customize the HTML template (see [Advanced Usage](#advanced-usage))

## Daily Operations

### Automated Updates (Recommended)

The standard workflow is to run `rp update` on a schedule using cron:

**Every 30 minutes (recommended for personal planets):**
```bash
# Edit crontab
crontab -e

# Add this line:
*/30 * * * * cd /home/user/my-planet && /usr/local/bin/rp update >> /home/user/my-planet/update.log 2>&1
```

**Every hour (for public planets):**
```cron
0 * * * * cd /home/user/my-planet && /usr/local/bin/rp update >> /home/user/my-planet/update.log 2>&1
```

**Every 15 minutes (for high-traffic planets):**
```cron
*/15 * * * * cd /home/user/my-planet && /usr/local/bin/rp update >> /home/user/my-planet/update.log 2>&1
```

**With log rotation:**
```bash
# Install logrotate config
sudo cat > /etc/logrotate.d/rogue-planet <<'EOF'
/home/user/my-planet/update.log {
    daily
    rotate 7
    compress
    missingok
    notifempty
}
EOF
```

### Manual Update Workflow

When you want to update immediately:

```bash
cd ~/my-planet

# Quick update
rp update

# Or: Update with verbose output
rp fetch    # Shows each feed being fetched
rp generate # Regenerates HTML
```

### Monitoring Your Planet

**Quick health check:**
```bash
cd ~/my-planet

# Check overall status
rp status

# Output:
# Rogue Planet Status
# ===================
#
# Feeds:           15 total (14 active, 1 inactive)
# Entries:         245 total
# Recent entries:  47 (last 7 days)
#
# Output:          ./public/index.html
# Database:        ./data/planet.db
```

**Detailed feed inspection:**
```bash
# List all feeds with status
rp list-feeds

# Check when HTML was last updated
ls -lh public/index.html

# Check database size
ls -lh data/planet.db

# View recent log entries
tail -20 update.log
```

**Check for problematic feeds:**
```bash
# Query database for feeds with errors
sqlite3 data/planet.db <<'EOF'
SELECT
    url,
    fetch_error_count,
    fetch_error,
    datetime(last_fetched)
FROM feeds
WHERE fetch_error_count > 0
ORDER BY fetch_error_count DESC;
EOF
```

## Feed Management

### Adding Feeds

**Add a single feed:**
```bash
rp add-feed https://example.com/feed.xml
rp update  # Fetch the new feed
```

**Add multiple feeds at once:**
```bash
# Create a file with new feeds
cat > new-feeds.txt <<'EOF'
https://blog.example.com/rss
https://news.example.com/feed.atom
https://podcast.example.com/feed.xml
EOF

# Import them
rp add-all -f new-feeds.txt
rp update
```

**Discover feed URLs:**
```bash
# Many sites have feed autodiscovery
curl -s https://example.com | grep -i "rss\|atom\|feed"

# Common feed locations to try:
# https://example.com/feed
# https://example.com/rss
# https://example.com/atom.xml
# https://example.com/index.xml
# https://example.com/feed.xml
```

### Removing Feeds

**Remove a single feed:**
```bash
# List feeds to find the URL
rp list-feeds

# Remove by URL
rp remove-feed https://example.com/feed.xml

# Regenerate HTML
rp generate
```

**Remove multiple feeds:**
```bash
# Remove feeds matching a pattern
for url in $(rp list-feeds | grep 'example.com' | awk '{print $2}'); do
    rp remove-feed "$url"
done

rp generate
```

### Managing Inactive Feeds

When feeds have repeated errors, you may want to temporarily disable them:

```bash
# Check for problematic feeds
sqlite3 data/planet.db "SELECT url, fetch_error_count FROM feeds WHERE fetch_error_count > 5;"

# Mark feed as inactive
sqlite3 data/planet.db "UPDATE feeds SET active = 0 WHERE url = 'https://broken-feed.example.com';"

# Later, reactivate it
sqlite3 data/planet.db "UPDATE feeds SET active = 1, fetch_error_count = 0 WHERE url = 'https://broken-feed.example.com';"

# Regenerate HTML
rp generate
```

## Troubleshooting

### Problem: Feed Not Showing New Entries

**Diagnosis workflow:**
```bash
# 1. Check if feed is active
rp list-feeds | grep "problem-feed.com"

# 2. Check if feed can be fetched
curl -I https://problem-feed.com/feed.xml

# 3. Try manual fetch
rp fetch

# 4. Check database for errors
sqlite3 data/planet.db <<'EOF'
SELECT
    url,
    datetime(last_fetched),
    fetch_error,
    fetch_error_count
FROM feeds
WHERE url LIKE '%problem-feed.com%';
EOF

# 5. Check if entries were stored
sqlite3 data/planet.db <<'EOF'
SELECT
    COUNT(*),
    MIN(datetime(published)),
    MAX(datetime(published))
FROM entries e
JOIN feeds f ON e.feed_id = f.id
WHERE f.url LIKE '%problem-feed.com%';
EOF
```

**Solution: Reset feed:**
```bash
# Remove and re-add the feed
rp remove-feed https://problem-feed.com/feed.xml
rp add-feed https://problem-feed.com/feed.xml
rp update
```

### Problem: HTML Not Updating

**Diagnosis workflow:**
```bash
# 1. Check if cron job is running
grep rp /var/log/syslog  # Linux
tail -f ~/my-planet/update.log  # Your log

# 2. Check file permissions
ls -la public/index.html

# 3. Check disk space
df -h .

# 4. Try manual regeneration
cd ~/my-planet
rp generate

# 5. Check if entries exist
rp status
```

**Solution: Force regeneration:**
```bash
# Regenerate from database
rp generate

# If that doesn't work, fetch everything fresh
rp fetch
rp generate

# Check output
ls -lh public/index.html
```

### Problem: Database Growing Too Large

**Diagnosis:**
```bash
# Check database size
ls -lh data/planet.db

# Count entries
sqlite3 data/planet.db "SELECT COUNT(*) FROM entries;"

# Check oldest entries
sqlite3 data/planet.db "SELECT datetime(MIN(published)) FROM entries;"
```

**Solution: Prune old entries:**
```bash
# Remove entries older than 30 days
rp prune --days 30

# Vacuum database to reclaim space
sqlite3 data/planet.db "VACUUM;"

# Check new size
ls -lh data/planet.db

# Set up automatic pruning (weekly)
cat >> cleanup.sh <<'EOF'
#!/bin/bash
cd /home/user/my-planet
./rp prune --days 90
sqlite3 data/planet.db "VACUUM;"
EOF
chmod +x cleanup.sh

# Add to crontab (every Sunday at 2 AM)
# 0 2 * * 0 /home/user/my-planet/cleanup.sh >> /home/user/my-planet/cleanup.log 2>&1
```

### Problem: Slow Updates

**Diagnosis:**
```bash
# Time an update
time rp update

# Check feed count
rp list-feeds | wc -l

# Check if specific feeds are slow
rp fetch  # Watch which feeds take longest
```

**Solution: Tune concurrency:**
```bash
# Edit config.ini
vim config.ini

# Increase concurrent fetches
# [planet]
# concurrent_fetches = 10  # Was 5

# Reduce retention period
# days = 3  # Was 7

# Test performance improvement
time rp update
```

### Problem: Cron Job Not Running

**Diagnosis:**
```bash
# Check if cron is running
systemctl status cron  # Linux
ps aux | grep cron     # Alternative

# Check crontab
crontab -l

# Check logs
grep rp /var/log/syslog  # Linux
tail -f ~/my-planet/update.log  # Your log

# Test command manually
cd /home/user/my-planet && /usr/local/bin/rp update
```

**Solution: Fix cron job:**
```bash
# Use absolute paths
crontab -e

# Correct format:
*/30 * * * * cd /home/user/my-planet && /usr/local/bin/rp update >> /home/user/my-planet/update.log 2>&1

# Verify it works
# Wait 30 minutes and check:
tail update.log
```

## Advanced Usage

### Custom Templates and Themes

**Using a pre-built theme:**

Rogue Planet includes five themes out of the box:

1. **Default Theme**: Built-in responsive theme (no configuration needed)
2. **Classic Theme**: Planet Venus-inspired design with right sidebar and nostalgic aesthetic
3. **Elegant Theme**: Modern design system with sophisticated typography and CSS custom properties
4. **Dark Theme**: Cutting-edge dark theme with OKLCH colors, cascade layers, and glassmorphism
5. **Flexoki Theme**: Automatic light/dark mode with warm, inky aesthetic inspired by analog paper

```bash
# Copy a theme to your planet directory
mkdir -p themes
cp -r examples/themes/classic themes/
# or
cp -r examples/themes/elegant themes/
# or
cp -r examples/themes/dark themes/
# or
cp -r examples/themes/flexoki themes/

# Update config.ini to use the theme
echo "template = ./themes/classic/template.html" >> config.ini
# or
echo "template = ./themes/elegant/template.html" >> config.ini
# or
echo "template = ./themes/dark/template.html" >> config.ini
# or
echo "template = ./themes/flexoki/template.html" >> config.ini

# Generate with the theme
rp generate
open public/index.html
```

See the theme documentation for customization:
- [Complete Theme Guide](THEMES.md) - Comprehensive guide to all themes and creating your own
- [Classic Theme README](examples/themes/classic/README.md)
- [Elegant Theme README](examples/themes/elegant/README.md)
- [Dark Theme README](examples/themes/dark/README.md)
- [Flexoki Theme README](examples/themes/flexoki/README.md)

**Creating a custom template:**

```bash
# 1. Create template directory with static assets
mkdir -p themes/custom/static

# 2. Create your template
cat > themes/custom/template.html <<'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="static/style.css">
</head>
<body>
    <h1>{{.Title}}</h1>
    {{range .Entries}}
    <article>
        <h2><a href="{{.Link}}">{{.Title}}</a></h2>
        <p>From: <a href="{{.FeedLink}}">{{.FeedTitle}}</a></p>
        <time>{{formatDate .Published}}</time>
        <div>{{.Content}}</div>
    </article>
    {{end}}
</body>
</html>
EOF

# 3. Create your CSS file
cat > themes/custom/static/style.css <<'EOF'
body {
    font-family: Georgia, serif;
    max-width: 800px;
    margin: 0 auto;
    padding: 2rem;
}
/* Your custom CSS here */
EOF

# 4. Update config.ini
echo "template = ./themes/custom/template.html" >> config.ini

# 5. Generate (static assets are copied automatically)
rp generate
open public/index.html
```

### Filtering and Curating Content

**Show only specific feeds in output:**

```bash
# Create custom SQL query to filter entries
cat > generate-filtered.sh <<'EOF'
#!/bin/bash
# Get entries only from specific feeds
sqlite3 data/planet.db <<SQL | rp generate --stdin
SELECT e.*
FROM entries e
JOIN feeds f ON e.feed_id = f.id
WHERE f.url IN (
    'https://blog.golang.org/feed.atom',
    'https://github.blog/feed/'
)
AND e.published >= datetime('now', '-7 days')
ORDER BY e.published DESC;
SQL
EOF
chmod +x generate-filtered.sh
```

### Multi-Planet Setup

Run multiple planets from the same installation:

```bash
# Create multiple planet directories
mkdir -p ~/planets/tech ~/planets/news ~/planets/personal

# Set up each planet
cd ~/planets/tech
rp init -f tech-feeds.txt

cd ~/planets/news
rp init -f news-feeds.txt

cd ~/planets/personal
rp init -f personal-feeds.txt

# Create update script for all planets
cat > ~/update-all-planets.sh <<'EOF'
#!/bin/bash
for planet in ~/planets/*; do
    echo "Updating $planet..."
    cd "$planet"
    rp update
done
EOF
chmod +x ~/update-all-planets.sh

# Add to cron
# */30 * * * * /home/user/update-all-planets.sh >> /home/user/planets.log 2>&1
```

### Integration with Static Site Generators

**Hugo integration:**

```bash
# Generate planet into Hugo content directory
cat >> config.ini <<'EOF'
[planet]
output_dir = ../hugo-site/static/planet
EOF

rp update

# Now public/index.html is available at yoursite.com/planet/
```

**Jekyll integration:**

```bash
# Similar approach
cat >> config.ini <<'EOF'
[planet]
output_dir = ../jekyll-site/planet
EOF

rp update
```

## Production Deployment

### Simple Nginx Setup

```bash
# 1. Set up planet in production location
cd /var/www
sudo mkdir planet
sudo chown $USER:$USER planet
cd planet

# 2. Initialize with your feeds
rp init -f feeds.txt

# 3. Configure Nginx
sudo cat > /etc/nginx/sites-available/planet <<'EOF'
server {
    listen 80;
    server_name planet.example.com;

    root /var/www/planet/public;
    index index.html;

    location / {
        try_files $uri $uri/ =404;
    }

    # Optional: Add caching headers
    location ~* \.(html|css|js)$ {
        expires 5m;
        add_header Cache-Control "public, must-revalidate";
    }
}
EOF

sudo ln -s /etc/nginx/sites-available/planet /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

# 4. Set up cron job
crontab -e
# Add: */30 * * * * cd /var/www/planet && rp update >> /var/www/planet/update.log 2>&1

# 5. Test
rp update
curl -I http://planet.example.com
```

### GitHub Pages Deployment

```bash
# 1. Create a git repository
cd ~/my-planet
git init
git add config.ini data/.gitkeep
git commit -m "Initial planet setup"

# 2. Add GitHub remote
git remote add origin git@github.com:username/planet.git
git branch -M main
git push -u origin main

# 3. Set up GitHub Actions
mkdir -p .github/workflows
cat > .github/workflows/update.yml <<'EOF'
name: Update Planet

on:
  schedule:
    - cron: '*/30 * * * *'  # Every 30 minutes
  workflow_dispatch:  # Allow manual trigger

jobs:
  update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install Rogue Planet
        run: go install github.com/adewale/rogue_planet/cmd/rp@latest

      - name: Update planet
        run: rp update

      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./public
EOF

git add .github/
git commit -m "Add GitHub Actions workflow"
git push

# 4. Enable GitHub Pages
# Go to: Settings → Pages → Source → gh-pages branch
```

### Docker Deployment

```bash
# 1. Create Dockerfile
cat > Dockerfile <<'EOF'
FROM golang:1.21-alpine AS builder
WORKDIR /app
RUN go install github.com/adewale/rogue_planet/cmd/rp@latest

FROM alpine:latest
RUN apk add --no-cache ca-certificates sqlite tzdata
COPY --from=builder /go/bin/rp /usr/local/bin/rp
WORKDIR /planet
VOLUME ["/planet/data", "/planet/public"]
CMD ["rp", "update"]
EOF

# 2. Create docker-compose.yml
cat > docker-compose.yml <<'EOF'
version: '3'
services:
  planet:
    build: .
    volumes:
      - ./data:/planet/data
      - ./public:/planet/public
      - ./config.ini:/planet/config.ini:ro
      - ./feeds.txt:/planet/feeds.txt:ro
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./public:/usr/share/nginx/html:ro
    restart: unless-stopped
EOF

# 3. Build and run
docker-compose up -d

# 4. Set up cron in container
docker-compose exec planet sh -c 'echo "*/30 * * * * rp update" | crontab -'
```

### Backup Strategy

**Automated backup script:**

```bash
cat > backup-planet.sh <<'EOF'
#!/bin/bash
PLANET_DIR="/var/www/planet"
BACKUP_DIR="/var/backups/planet"
DATE=$(date +%Y%m%d-%H%M%S)

mkdir -p "$BACKUP_DIR"

# Backup database and config
tar czf "$BACKUP_DIR/planet-$DATE.tar.gz" \
    -C "$PLANET_DIR" \
    config.ini \
    data/planet.db \
    feeds.txt

# Keep only last 30 days of backups
find "$BACKUP_DIR" -name "planet-*.tar.gz" -mtime +30 -delete

echo "Backup completed: $BACKUP_DIR/planet-$DATE.tar.gz"
EOF

chmod +x backup-planet.sh

# Add to cron (daily at 3 AM)
# 0 3 * * * /var/www/planet/backup-planet.sh >> /var/log/planet-backup.log 2>&1
```

### Monitoring and Alerting

**Simple monitoring script:**

```bash
cat > monitor-planet.sh <<'EOF'
#!/bin/bash
PLANET_DIR="/var/www/planet"
ALERT_EMAIL="admin@example.com"

cd "$PLANET_DIR"

# Check if HTML file was updated in last hour
if [ $(find public/index.html -mmin +60) ]; then
    echo "WARNING: Planet HTML not updated in last hour" | \
        mail -s "Planet Alert" "$ALERT_EMAIL"
fi

# Check database size
DB_SIZE=$(du -m data/planet.db | cut -f1)
if [ "$DB_SIZE" -gt 1000 ]; then
    echo "WARNING: Database size is ${DB_SIZE}MB" | \
        mail -s "Planet Alert" "$ALERT_EMAIL"
fi

# Check for feeds with high error counts
ERROR_COUNT=$(sqlite3 data/planet.db \
    "SELECT COUNT(*) FROM feeds WHERE fetch_error_count > 10;")
if [ "$ERROR_COUNT" -gt 0 ]; then
    echo "WARNING: $ERROR_COUNT feeds have high error counts" | \
        mail -s "Planet Alert" "$ALERT_EMAIL"
fi
EOF

chmod +x monitor-planet.sh

# Add to cron (hourly)
# 0 * * * * /var/www/planet/monitor-planet.sh
```

## Summary

**Most Common Workflows:**

1. **Initial setup**: `rp init -f feeds.txt && rp update`
2. **Daily operation**: Cron job running `rp update` every 30 minutes
3. **Add feed**: `rp add-feed <url> && rp update`
4. **Troubleshoot**: `rp status`, `rp list-feeds`, check logs
5. **Maintenance**: `rp prune --days 90` periodically

**Best Practices:**

- Run `rp update` on a schedule (cron)
- Monitor `rp status` regularly
- Prune old entries periodically
- Back up `data/planet.db` and `config.ini`
- Keep logs for troubleshooting
- Use `rp fetch` and `rp generate` separately when debugging

**Getting Help:**

- Check logs: `tail -f update.log`
- Run commands manually to see errors
- Use `sqlite3 data/planet.db` to inspect database
- See README.md for configuration options
- Report issues at https://github.com/adewale/rogue_planet/issues
