# Rogue Planet Themes & Templates

Complete guide to using existing themes and creating your own custom templates for Rogue Planet.

## Table of Contents

- [Built-in Themes](#built-in-themes)
- [Using a Theme](#using-a-theme)
- [Creating Your Own Theme](#creating-your-own-theme)
- [Template Variables Reference](#template-variables-reference)
- [Template Functions](#template-functions)
- [Theme Structure](#theme-structure)
- [Examples](#examples)

---

## Built-in Themes

Rogue Planet includes **five themes** out of the box:

### 1. Default Theme (Built-in)
**Modern responsive design with clean typography**

- ✅ No configuration needed - works automatically
- ✅ Responsive flexbox layout
- ✅ System font stack for optimal performance
- ✅ Clean, minimal design
- ✅ Feed sidebar with health status
- ✅ Mobile-friendly responsive breakpoints

**Best for**: General use, quick setup, modern aesthetics

### 2. Classic Theme
**Faithful recreation of Planet Venus classic_fancy theme**

- 🎨 Right sidebar with RSS feed icons
- 🎨 Bitstream Vera Sans typography
- 🎨 Blue/purple color scheme (#200080 purple, #a0c0ff blue)
- 🎨 Lowercase headers with negative letter spacing
- 🎨 Nostalgic Planet Planet aesthetic from 2000s-2010s

**Best for**: Nostalgia, traditional Planet sites, text-focused content

See [examples/themes/classic/README.md](examples/themes/classic/README.md) for details.

### 3. Elegant Theme
**Modern design system with sophisticated typography**

- 🎯 CSS custom properties for easy customization
- 🎯 Modular typography scale
- 🎯 Semantic color palette
- 🎯 Georgia serif body text for readability
- 🎯 Print-optimized styles
- 🎯 Generous spacing and elegant layouts

**Best for**: Professional sites, content-heavy planets, readable long-form posts

See [examples/themes/elegant/README.md](examples/themes/elegant/README.md) for details.

### 4. Dark Theme
**Cutting-edge CSS with lush dark aesthetic**

- 🌙 OKLCH color space for perceptually uniform vibrant colors
- 🌙 CSS cascade layers for maintainable architecture
- 🌙 Fluid typography with clamp() functions
- 🌙 Glassmorphism effects with backdrop-filter
- 🌙 Container queries and modern selectors
- 🌙 Electric accent colors (cyan, violet, pink, emerald)

**Best for**: Modern browsers, showcasing CSS features, vibrant dark mode

See [examples/themes/dark/README.md](examples/themes/dark/README.md) for details.

### 5. Flexoki Theme
**Automatic light/dark mode with warm, inky aesthetic**

- 📖 Designed for reading with warm paper tones
- 📖 Automatic light/dark mode switching (no JavaScript)
- 📖 Oklab color space for perceptual consistency
- 📖 Serif typography optimized for long-form content
- 📖 8 vibrant accent colors (red, orange, yellow, green, cyan, blue, purple, magenta)
- 📖 Respects system color scheme preference

**Best for**: Reading-focused sites, automatic theming, users who switch between light/dark modes

See [examples/themes/flexoki/README.md](examples/themes/flexoki/README.md) for details.

---

## Using a Theme

### Quick Start

1. **Copy the theme** to your planet directory:

```bash
# Choose one:
mkdir -p themes
cp -r examples/themes/classic themes/
# or
cp -r examples/themes/elegant themes/
# or
cp -r examples/themes/dark themes/
# or
cp -r examples/themes/flexoki themes/
```

2. **Update your config.ini**:

```ini
[planet]
template = ./themes/classic/template.html
# or
template = ./themes/elegant/template.html
# or
template = ./themes/dark/template.html
# or
template = ./themes/flexoki/template.html
```

3. **Generate your site**:

```bash
rp generate
```

Static assets (CSS, images, SVG files) are **automatically copied** from `themes/*/static/` to `public/static/` when you generate.

### Using an Absolute Path

You can also reference themes with absolute paths:

```ini
[planet]
template = /full/path/to/rogue_planet/examples/themes/classic/template.html
```

---

## Creating Your Own Theme

### Method 1: Modify an Existing Theme

The easiest way to create a custom theme is to start with an existing one:

```bash
# 1. Copy an existing theme
cp -r examples/themes/elegant themes/mytheme

# 2. Customize the template
vim themes/mytheme/template.html

# 3. Customize the styles
vim themes/mytheme/static/style.css

# 4. Use your theme
echo "template = ./themes/mytheme/template.html" >> config.ini

# 5. Generate and preview
rp generate
open public/index.html
```

### Method 2: Create from Scratch

**Minimal theme structure:**

```
themes/mytheme/
├── template.html        # HTML template (required)
├── static/              # Static assets (optional)
│   ├── style.css       # CSS stylesheet
│   ├── images/         # Images
│   └── icons/          # Icons, SVG files
└── README.md           # Theme documentation (recommended)
```

**Basic template example:**

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="static/style.css">
</head>
<body>
    <header>
        <h1>{{.Title}}</h1>
        {{if .Subtitle}}<p>{{.Subtitle}}</p>{{end}}
    </header>

    <main>
        {{range .Entries}}
        <article>
            <h2><a href="{{.Link}}">{{.Title}}</a></h2>
            <p class="meta">
                By {{.Author}} from <a href="{{.FeedLink}}">{{.FeedTitle}}</a>
                • {{.PublishedRelative}}
            </p>
            <div class="content">{{.Content}}</div>
        </article>
        {{end}}
    </main>

    <aside>
        <h2>Feeds</h2>
        <ul>
        {{range .Feeds}}
            <li><a href="{{.Link}}">{{.Title}}</a></li>
        {{end}}
        </ul>
    </aside>

    <footer>
        Generated by {{.Generator}} on {{formatDate .Updated}}
    </footer>
</body>
</html>
```

### Date Grouping

To group entries by date ("Today", "Yesterday", etc.), use the `DateGroups` variable:

```html
{{if .GroupByDate}}
    {{range .DateGroups}}
    <section class="date-group">
        <h2>{{.DateStr}}</h2>
        {{range .Entries}}
        <article>
            <h3><a href="{{.Link}}">{{.Title}}</a></h3>
            <div>{{.Content}}</div>
        </article>
        {{end}}
    </section>
    {{end}}
{{else}}
    <!-- Non-grouped entries -->
    {{range .Entries}}
    <article>...</article>
    {{end}}
{{end}}
```

Enable in config.ini:

```ini
[planet]
group_by_date = true
```

---

## Template Variables Reference

All variables available in your templates:

### Site-Level Variables

| Variable | Type | Description |
|----------|------|-------------|
| `{{.Title}}` | string | Site title |
| `{{.Subtitle}}` | string | Site subtitle (optional) |
| `{{.Link}}` | string | Site URL |
| `{{.Updated}}` | time.Time | Last generated timestamp |
| `{{.Generator}}` | string | Generator name and version ("Rogue Planet v0.1") |
| `{{.OwnerName}}` | string | Planet owner name |
| `{{.OwnerEmail}}` | string | Planet owner email |
| `{{.GroupByDate}}` | bool | Whether entries are grouped by date |

### Entry Variables

Access via `{{range .Entries}}...{{end}}`

| Variable | Type | Description |
|----------|------|-------------|
| `{{.Title}}` | HTML | Entry title (sanitized) |
| `{{.Link}}` | string | Entry permalink URL |
| `{{.Author}}` | string | Entry author name |
| `{{.FeedTitle}}` | string | Source feed title |
| `{{.FeedLink}}` | string | Source feed website URL |
| `{{.Published}}` | time.Time | Published date |
| `{{.Updated}}` | time.Time | Last updated date |
| `{{.Content}}` | HTML | Full entry content (sanitized HTML) |
| `{{.Summary}}` | HTML | Entry summary (sanitized HTML) |
| `{{.PublishedRelative}}` | string | Relative time ("2 hours ago", "yesterday") |

### Date Group Variables

Access via `{{range .DateGroups}}...{{end}}` (when `GroupByDate` is enabled)

| Variable | Type | Description |
|----------|------|-------------|
| `{{.Date}}` | time.Time | Date for this group |
| `{{.DateStr}}` | string | Formatted date string ("Today", "Yesterday", "Monday, January 2") |
| `{{.Entries}}` | []EntryData | Entries for this date |

### Feed Variables

Access via `{{range .Feeds}}...{{end}}` (for sidebar)

| Variable | Type | Description |
|----------|------|-------------|
| `{{.Title}}` | string | Feed title |
| `{{.Link}}` | string | Feed website URL |
| `{{.URL}}` | string | Feed XML/RSS/Atom URL |
| `{{.LastUpdated}}` | time.Time | Last successful fetch time |
| `{{.ErrorCount}}` | int | Number of consecutive fetch errors |

---

## Template Functions

Built-in functions you can use in templates:

### Date Formatting Functions

```go
{{formatDate .Updated}}
// Output: "January 2, 2006 at 3:04 PM"

{{formatDateShort .Published}}
// Output: "Jan 2, 2006"

{{formatDateISO .Published}}
// Output: "2006-01-02T15:04:05Z07:00" (RFC3339)

{{relativeTime .Published}}
// Output: "2 hours ago", "yesterday", "3 days ago"
```

### Conditional Logic

```html
{{if .Subtitle}}
    <p>{{.Subtitle}}</p>
{{end}}

{{if .Author}}
    By {{.Author}}
{{else}}
    By Anonymous
{{end}}

{{if gt .ErrorCount 0}}
    <span class="error">{{.ErrorCount}} errors</span>
{{end}}
```

### Loops

```html
<!-- Basic loop -->
{{range .Entries}}
    <article>{{.Title}}</article>
{{end}}

<!-- Loop with index -->
{{range $index, $entry := .Entries}}
    <article data-index="{{$index}}">{{$entry.Title}}</article>
{{end}}

<!-- Empty check -->
{{if .Entries}}
    {{range .Entries}}...{{end}}
{{else}}
    <p>No entries yet.</p>
{{end}}
```

### Comparison Operators

```html
{{if eq .Title "My Planet"}}         <!-- Equal -->
{{if ne .Author ""}}                <!-- Not equal -->
{{if gt .ErrorCount 5}}             <!-- Greater than -->
{{if lt .ErrorCount 3}}             <!-- Less than -->
{{if ge .ErrorCount 1}}             <!-- Greater or equal -->
{{if le .ErrorCount 10}}            <!-- Less or equal -->
```

---

## Theme Structure

### Recommended Directory Layout

```
themes/mytheme/
├── template.html          # Main HTML template (required)
├── static/                # Static assets directory (optional)
│   ├── style.css         # Main stylesheet
│   ├── print.css         # Print-specific styles (optional)
│   ├── fonts/            # Custom fonts (optional)
│   │   └── custom.woff2
│   ├── images/           # Images (optional)
│   │   └── logo.png
│   └── icons/            # SVG icons (optional)
│       └── feed-icon.svg
├── README.md             # Theme documentation
└── LICENSE               # License information (if distributing)
```

### Static Asset Handling

**Automatic Copying**: When you use a custom template, Rogue Planet automatically copies `static/` from your theme directory to `public/static/` in the output.

**In your template**, reference static assets like this:

```html
<link rel="stylesheet" href="static/style.css">
<img src="static/images/logo.png" alt="Logo">
<script src="static/script.js"></script>
```

**Output structure** after generation:

```
public/
├── index.html
└── static/             # Automatically copied from themes/mytheme/static/
    ├── style.css
    ├── images/
    │   └── logo.png
    └── icons/
        └── feed-icon.svg
```

### CSS Best Practices

**External CSS** (recommended):

```html
<!-- In template.html -->
<link rel="stylesheet" href="static/style.css">
```

**Inline CSS** (for simple themes):

```html
<style>
    body { font-family: sans-serif; }
    .entry { margin: 20px 0; }
</style>
```

**CSS custom properties** (modern approach):

```css
:root {
  --primary-color: #0066cc;
  --text-color: #333;
  --spacing: 1.5rem;
}

body {
  color: var(--text-color);
}

.entry {
  margin-bottom: var(--spacing);
}
```

---

## Examples

### Example 1: Minimal Theme

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    <style>
        body { max-width: 800px; margin: 40px auto; font-family: sans-serif; }
        article { margin: 30px 0; padding-bottom: 20px; border-bottom: 1px solid #eee; }
    </style>
</head>
<body>
    <h1>{{.Title}}</h1>
    {{range .Entries}}
    <article>
        <h2><a href="{{.Link}}">{{.Title}}</a></h2>
        <p><small>{{.FeedTitle}} • {{.PublishedRelative}}</small></p>
        {{.Content}}
    </article>
    {{end}}
</body>
</html>
```

### Example 2: Feed Sidebar with Health Status

```html
<aside class="sidebar">
    <h2>Subscriptions ({{len .Feeds}})</h2>
    <ul>
    {{range .Feeds}}
        <li>
            <a href="{{.Link}}">{{.Title}}</a>
            {{if .LastUpdated}}
                <small>Updated {{relativeTime .LastUpdated}}</small>
            {{end}}
            {{if gt .ErrorCount 0}}
                <span class="error">⚠ {{.ErrorCount}} errors</span>
            {{end}}
        </li>
    {{end}}
    </ul>
</aside>
```

### Example 3: Date Grouped Entries

```html
{{if .GroupByDate}}
    {{range .DateGroups}}
    <section class="date-group">
        <h2 class="date-header">{{.DateStr}}</h2>
        {{range .Entries}}
        <article>
            <h3><a href="{{.Link}}">{{.Title}}</a></h3>
            <p class="meta">
                {{if .Author}}{{.Author}} • {{end}}
                <a href="{{.FeedLink}}">{{.FeedTitle}}</a>
            </p>
            <div class="content">{{.Content}}</div>
        </article>
        {{end}}
    </section>
    {{end}}
{{end}}
```

### Example 4: Responsive Layout

```css
/* Mobile-first responsive */
.container {
    padding: 20px;
}

.layout {
    display: flex;
    flex-direction: column;
}

/* Desktop */
@media (min-width: 768px) {
    .layout {
        flex-direction: row;
        gap: 40px;
    }

    .main-content {
        flex: 1;
    }

    .sidebar {
        width: 280px;
        flex-shrink: 0;
    }
}
```

### Example 5: Print Styles

```css
@media print {
    /* Hide sidebar when printing */
    .sidebar {
        display: none;
    }

    /* Remove backgrounds */
    body {
        background: white;
    }

    /* Add page breaks */
    article {
        page-break-inside: avoid;
    }

    /* Show URLs */
    a[href]::after {
        content: " (" attr(href) ")";
        font-size: 0.8em;
    }
}
```

---

## Security Considerations

### Content is Pre-Sanitized

**Important**: All entry content (`{{.Content}}` and `{{.Summary}}`) is **already sanitized** by Rogue Planet before reaching your template. You can safely output it without additional escaping.

```html
<!-- This is SAFE - content is pre-sanitized -->
<div class="entry-content">
    {{.Content}}
</div>
```

### Use html/template Auto-Escaping

Rogue Planet uses Go's `html/template` which **automatically escapes** HTML in most contexts:

```html
<!-- Automatically escaped -->
<title>{{.Title}}</title>
<meta name="author" content="{{.OwnerName}}">

<!-- NO escaping needed for pre-sanitized HTML content -->
<div>{{.Content}}</div>
```

### Content Security Policy

Include a Content Security Policy to prevent XSS:

```html
<meta http-equiv="Content-Security-Policy"
      content="default-src 'self';
               script-src 'self';
               style-src 'self' 'unsafe-inline';
               img-src 'self' https:;
               object-src 'none';">
```

---

## Troubleshooting

### Theme Not Loading

**Problem**: Template errors or theme doesn't apply

**Solutions**:
1. Check template path in config.ini is correct
2. Verify template.html exists in theme directory
3. Check for Go template syntax errors
4. Look at rp generate output for error messages

```bash
# Test your theme
rp generate -v
```

### Static Assets Not Copying

**Problem**: CSS/images not appearing in output

**Solutions**:
1. Verify `static/` directory exists in theme folder
2. Reference assets as `static/style.css` (not `/static/` or `./static/`)
3. Check file permissions on static directory
4. Regenerate site: `rp generate`

### Template Variable Not Found

**Problem**: `{{.Variable}}` shows nothing or errors

**Solutions**:
1. Check variable name spelling (case-sensitive)
2. Verify variable exists in [Template Variables Reference](#template-variables-reference)
3. Use `{{if .Variable}}` to check if it exists before using
4. Check Go template syntax

### Styles Not Applying

**Problem**: Page loads but looks unstyled

**Solutions**:
1. Check browser developer tools for 404 errors
2. Verify CSS path in `<link>` tag
3. Ensure static assets were copied (check `public/static/`)
4. Clear browser cache
5. Check CSS syntax for errors

---

## Contributing Themes

Have you created a great theme? Consider contributing it!

1. Create a theme directory in `examples/themes/`
2. Include comprehensive README.md
3. Add LICENSE file with proper attribution
4. Test on multiple browsers
5. Ensure accessibility (WCAG AA contrast ratios)
6. Submit a pull request

---

## Additional Resources

- **Theme Examples**: `examples/themes/`
- **Default Template Source**: `pkg/generator/generator.go` (see `defaultTemplate` constant)
- **Quick Start Guide**: [QUICKSTART.md](QUICKSTART.md)
- **Complete Workflows**: [WORKFLOWS.md](WORKFLOWS.md)
- **Development Guide**: [CLAUDE.md](CLAUDE.md)

---

## Quick Reference

### Essential Template Snippet

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
</head>
<body>
    <h1>{{.Title}}</h1>
    {{range .Entries}}
    <article>
        <h2><a href="{{.Link}}">{{.Title}}</a></h2>
        <p>{{.FeedTitle}} • {{.PublishedRelative}}</p>
        <div>{{.Content}}</div>
    </article>
    {{end}}
</body>
</html>
```

### Essential Config

```ini
[planet]
name = My Planet
link = https://planet.example.com
template = ./themes/mytheme/template.html
group_by_date = true
```

### Essential Commands

```bash
# Copy a theme
cp -r examples/themes/classic themes/mytheme

# Edit your theme
vim themes/mytheme/template.html
vim themes/mytheme/static/style.css

# Generate and preview
rp generate
open public/index.html
```

---

*Last Updated: 2025-10-10*
*Rogue Planet v0.1.0*
