// Package generator provides static HTML generation for feed aggregation.
//
// The generator renders feed entries using Go's html/template with automatic
// HTML escaping. It supports custom templates, date grouping, feed sidebars,
// and responsive layouts. The default template follows classic Planet Planet design.
package generator

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/adewale/rogue_planet/pkg/timeprovider"
)

// TemplateData contains all data needed for template rendering
type TemplateData struct {
	Title       string
	Subtitle    string
	Link        string
	Updated     time.Time
	Generator   string
	OwnerName   string
	OwnerEmail  string
	Entries     []EntryData
	GroupByDate bool
	DateGroups  []DateGroup
	Feeds       []FeedData // For sidebar
}

// FeedData represents a feed for sidebar display
type FeedData struct {
	Title       string
	Link        string
	URL         string
	Subscribers int
	LastUpdated time.Time
	ErrorCount  int
}

// EntryData represents an entry for template rendering
type EntryData struct {
	Title             template.HTML
	Link              string
	Author            string
	FeedTitle         string
	FeedLink          string
	Published         time.Time
	Updated           time.Time
	Content           template.HTML // Already sanitized, safe to render
	Summary           template.HTML
	PublishedRelative string
}

// DateGroup groups entries by date
type DateGroup struct {
	Date    time.Time
	DateStr string
	Entries []EntryData
}

// Generator handles static HTML generation
type Generator struct {
	template     *template.Template
	templatePath string // Path to template file (if custom template)
	timeProvider timeprovider.TimeProvider
}

// New creates a new Generator with the default template and real system time
func New() (*Generator, error) {
	g := &Generator{
		timeProvider: timeprovider.WallClock{},
	}

	tmpl, err := template.New("default").Funcs(g.templateFuncs()).Parse(defaultTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse default template: %w", err)
	}

	g.template = tmpl
	return g, nil
}

// NewWithTimeProvider creates a new Generator with the default template and custom TimeProvider.
// This is primarily for testing with FakeClock.
func NewWithTimeProvider(tp timeprovider.TimeProvider) (*Generator, error) {
	g := &Generator{
		timeProvider: tp,
	}

	tmpl, err := template.New("default").Funcs(g.templateFuncs()).Parse(defaultTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse default template: %w", err)
	}

	g.template = tmpl
	return g, nil
}

// NewWithTemplate creates a Generator with a custom template and real system time
func NewWithTemplate(templatePath string) (*Generator, error) {
	g := &Generator{
		templatePath: templatePath,
		timeProvider: timeprovider.WallClock{},
	}

	tmpl, err := template.New(filepath.Base(templatePath)).Funcs(g.templateFuncs()).ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	g.template = tmpl
	return g, nil
}

// Generate generates HTML and writes it to the specified writer
func (g *Generator) Generate(ctx context.Context, w io.Writer, data TemplateData) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Add version info
	data.Generator = "Rogue Planet v0.1"
	data.Updated = g.timeProvider.Now()

	// Calculate relative dates using the time provider
	for i := range data.Entries {
		data.Entries[i].PublishedRelative = relativeTime(data.Entries[i].Published, g.timeProvider)
	}

	// Group by date if requested
	if data.GroupByDate {
		data.DateGroups = groupEntriesByDate(data.Entries, g.timeProvider)
	}

	// Execute template
	if err := g.template.Execute(w, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}

// GenerateToFile generates HTML and writes it to a file
func (g *Generator) GenerateToFile(ctx context.Context, outputPath string, data TemplateData) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Create output directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	// Create output file
	f, createErr := os.Create(outputPath)
	if createErr != nil {
		return fmt.Errorf("create output file: %w", createErr)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close output file: %w", closeErr)
		}
	}()

	// Generate HTML
	if err := g.Generate(ctx, f, data); err != nil {
		return err
	}

	// Copy static assets if using custom template
	if g.templatePath != "" {
		if err := g.CopyStaticAssets(ctx, dir); err != nil {
			return fmt.Errorf("copy static assets: %w", err)
		}
	}

	return nil
}

// CopyStaticAssets copies static assets from template directory to output directory
func (g *Generator) CopyStaticAssets(ctx context.Context, outputDir string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if g.templatePath == "" {
		return nil // No custom template, no static assets
	}

	// Find static directory relative to template
	templateDir := filepath.Dir(g.templatePath)
	staticSrc := filepath.Join(templateDir, "static")

	// Check if static directory exists
	if _, err := os.Stat(staticSrc); os.IsNotExist(err) {
		return nil // No static directory, nothing to copy
	}

	// Destination static directory
	staticDst := filepath.Join(outputDir, "static")

	// Remove existing static directory
	if err := os.RemoveAll(staticDst); err != nil {
		return fmt.Errorf("remove existing static directory: %w", err)
	}

	// Copy static directory
	if err := copyDir(ctx, staticSrc, staticDst); err != nil {
		return fmt.Errorf("copy static directory: %w", err)
	}

	return nil
}

// copyDir recursively copies a directory
func copyDir(ctx context.Context, src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read directory contents
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}

		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(ctx, srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(ctx, srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(ctx context.Context, src, dst string) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Open source file
	srcFile, openErr := os.Open(src)
	if openErr != nil {
		return openErr
	}
	defer func() {
		if closeErr := srcFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	// Get source file info
	srcInfo, statErr := srcFile.Stat()
	if statErr != nil {
		return statErr
	}

	// Create destination file
	dstFile, createErr := os.Create(dst)
	if createErr != nil {
		return createErr
	}
	defer func() {
		if closeErr := dstFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	// Copy file contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Set file permissions
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return err
	}

	return nil
}

// templateFuncs returns custom template functions with access to the Generator's timeProvider
func (g *Generator) templateFuncs() template.FuncMap {
	return template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("January 2, 2006 at 3:04 PM")
		},
		"formatDateShort": func(t time.Time) string {
			return t.Format("Jan 2, 2006")
		},
		"formatDateISO": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
		"relativeTime": func(t time.Time) string {
			return relativeTime(t, g.timeProvider)
		},
	}
}

// relativeTime returns a human-readable relative time string
func relativeTime(t time.Time, tp timeprovider.TimeProvider) string {
	diff := tp.Since(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 48*time.Hour:
		return "yesterday"
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case diff < 365*24*time.Hour:
		months := int(diff.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(diff.Hours() / (24 * 365))
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

// groupEntriesByDate groups entries by their published date
func groupEntriesByDate(entries []EntryData, tp timeprovider.TimeProvider) []DateGroup {
	groups := make(map[string][]EntryData)
	// Pre-allocate with entries length as upper bound (worst case: one entry per day)
	dateOrder := make([]string, 0, len(entries))

	for _, entry := range entries {
		dateKey := entry.Published.Format("2006-01-02")
		if _, exists := groups[dateKey]; !exists {
			dateOrder = append(dateOrder, dateKey)
		}
		groups[dateKey] = append(groups[dateKey], entry)
	}

	result := make([]DateGroup, 0, len(dateOrder))
	for _, dateKey := range dateOrder {
		date, _ := time.Parse("2006-01-02", dateKey)
		result = append(result, DateGroup{
			Date:    date,
			DateStr: formatDateGroup(date, tp),
			Entries: groups[dateKey],
		})
	}

	return result
}

// formatDateGroup formats a date for group headers
func formatDateGroup(t time.Time, tp timeprovider.TimeProvider) string {
	now := tp.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)
	targetDate := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	if targetDate.Equal(today) {
		return "Today"
	} else if targetDate.Equal(yesterday) {
		return "Yesterday"
	} else if now.Sub(targetDate) < 7*24*time.Hour {
		return t.Format("Monday, January 2")
	}

	return t.Format("Monday, January 2, 2006")
}

// defaultTemplate is the built-in HTML template
const defaultTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Security-Policy" content="default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' https:; object-src 'none'; base-uri 'self';">
    <title>{{.Title}}</title>
    <meta name="generator" content="{{.Generator}}">
    <style>
        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background: #f5f5f5;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .layout {
            display: flex;
            gap: 0;
        }
        .main-content {
            flex: 1;
            padding: 40px;
            min-width: 0;
        }
        .sidebar {
            width: 280px;
            background: #f9f9f9;
            border-left: 1px solid #e0e0e0;
            padding: 30px 20px;
        }
        .sidebar h2 {
            font-size: 1.2em;
            margin-bottom: 15px;
            color: #333;
            border-bottom: 2px solid #ddd;
            padding-bottom: 8px;
        }
        .sidebar ul {
            list-style: none;
        }
        .sidebar li {
            margin-bottom: 12px;
            font-size: 0.9em;
        }
        .sidebar a {
            color: #0066cc;
            text-decoration: none;
            display: block;
        }
        .sidebar a:hover {
            text-decoration: underline;
        }
        .feed-meta {
            font-size: 0.8em;
            color: #999;
            margin-top: 3px;
        }
        .feed-error {
            color: #cc0000;
        }
        header {
            border-bottom: 3px solid #333;
            padding-bottom: 20px;
            margin-bottom: 40px;
        }
        h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
        }
        h1 a {
            color: #333;
            text-decoration: none;
        }
        h1 a:hover {
            color: #666;
        }
        .subtitle {
            color: #666;
            font-size: 1.1em;
        }
        .date-group {
            margin-bottom: 40px;
        }
        .date-group h2 {
            font-size: 1.5em;
            color: #666;
            border-bottom: 2px solid #eee;
            padding-bottom: 10px;
            margin-bottom: 20px;
        }
        .entry {
            margin-bottom: 40px;
            padding-bottom: 30px;
            border-bottom: 1px solid #eee;
        }
        .entry:last-child {
            border-bottom: none;
        }
        .entry h3 {
            font-size: 1.5em;
            margin-bottom: 10px;
        }
        .entry h3 a {
            color: #0066cc;
            text-decoration: none;
        }
        .entry h3 a:hover {
            text-decoration: underline;
        }
        .entry-meta {
            color: #666;
            font-size: 0.9em;
            margin-bottom: 15px;
        }
        .entry-meta a {
            color: #666;
            text-decoration: none;
        }
        .entry-meta a:hover {
            color: #333;
            text-decoration: underline;
        }
        .entry-content {
            margin-top: 15px;
        }
        .entry-content img {
            max-width: 100%;
            height: auto;
        }
        .entry-content pre {
            background: #f5f5f5;
            padding: 15px;
            overflow-x: auto;
            border-radius: 5px;
        }
        .entry-content code {
            background: #f5f5f5;
            padding: 2px 5px;
            border-radius: 3px;
            font-family: monospace;
        }
        .entry-content pre code {
            background: none;
            padding: 0;
        }
        .entry-content blockquote {
            border-left: 4px solid #ddd;
            padding-left: 20px;
            margin: 20px 0;
            color: #666;
        }
        footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #eee;
            text-align: center;
            color: #666;
            font-size: 0.9em;
        }
        footer a {
            color: #666;
        }
        @media (max-width: 968px) {
            .layout {
                flex-direction: column;
            }
            .sidebar {
                width: 100%;
                border-left: none;
                border-top: 1px solid #e0e0e0;
            }
        }
        @media (max-width: 768px) {
            body {
                padding: 10px;
            }
            .main-content {
                padding: 20px;
            }
            h1 {
                font-size: 2em;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="layout">
            <div class="main-content">
                <header>
                    <h1>{{if .Link}}<a href="{{.Link}}">{{.Title}}</a>{{else}}{{.Title}}{{end}}</h1>
                    {{if .Subtitle}}<p class="subtitle">{{.Subtitle}}</p>{{end}}
                </header>

                <main>
            {{if .GroupByDate}}
                {{range .DateGroups}}
                <div class="date-group">
                    <h2>{{.DateStr}}</h2>
                    {{range .Entries}}
                    <article class="entry">
                        <h3><a href="{{.Link}}">{{.Title}}</a></h3>
                        <div class="entry-meta">
                            {{if .Author}}By {{.Author}} &middot; {{end}}
                            <a href="{{.FeedLink}}">{{.FeedTitle}}</a> &middot;
                            <time datetime="{{formatDateISO .Published}}">{{.PublishedRelative}}</time>
                        </div>
                        <div class="entry-content">
                            {{.Content}}
                        </div>
                    </article>
                    {{end}}
                </div>
                {{end}}
            {{else}}
                {{range .Entries}}
                <article class="entry">
                    <h3><a href="{{.Link}}">{{.Title}}</a></h3>
                    <div class="entry-meta">
                        {{if .Author}}By {{.Author}} &middot; {{end}}
                        <a href="{{.FeedLink}}">{{.FeedTitle}}</a> &middot;
                        <time datetime="{{formatDateISO .Published}}">{{.PublishedRelative}}</time>
                    </div>
                    <div class="entry-content">
                        {{.Content}}
                    </div>
                </article>
                {{end}}
            {{end}}
                </main>

                <footer>
                    <p>Generated by {{.Generator}} on {{formatDate .Updated}}</p>
                    {{if .OwnerName}}<p>&copy; {{.Updated.Year}} {{.OwnerName}}</p>{{end}}
                </footer>
            </div>

            {{if .Feeds}}
            <aside class="sidebar">
                <h2>Subscriptions</h2>
                <ul>
                {{range .Feeds}}
                    <li>
                        <a href="{{.Link}}" title="{{.URL}}">{{.Title}}</a>
                        {{if .LastUpdated}}
                        <div class="feed-meta">
                            Updated {{relativeTime .LastUpdated}}
                        </div>
                        {{end}}
                        {{if gt .ErrorCount 0}}
                        <div class="feed-meta feed-error">
                            {{.ErrorCount}} errors
                        </div>
                        {{end}}
                    </li>
                {{end}}
                </ul>
            </aside>
            {{end}}
        </div>
    </div>
</body>
</html>
`
