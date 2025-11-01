package generator

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adewale/rogue_planet/pkg/timeprovider"
)

func TestNew(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if gen.template == nil {
		t.Error("template should not be nil")
	}
}

func TestGenerate(t *testing.T) {
	gen, _ := New()

	data := TemplateData{
		Title:    "Test Planet",
		Subtitle: "A collection of feeds",
		Link:     "https://example.com",
		Entries: []EntryData{
			{
				Title:     "Test Entry",
				Link:      "https://example.com/post1",
				Author:    "John Doe",
				FeedTitle: "Example Feed",
				FeedLink:  "https://example.com/feed",
				Published: time.Now().Add(-2 * time.Hour),
				Updated:   time.Now(),
				Content:   template.HTML("<p>This is test content</p>"),
				Summary:   template.HTML("Test summary"),
			},
		},
	}

	var buf bytes.Buffer
	err := gen.Generate(&buf, data)

	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	output := buf.String()

	// Check that HTML was generated
	if len(output) == 0 {
		t.Error("Generated HTML should not be empty")
	}

	// Check for key elements
	tests := []struct {
		name     string
		contains string
	}{
		{"doctype", "<!DOCTYPE html>"},
		{"title", "<title>Test Planet</title>"},
		{"subtitle", "A collection of feeds"},
		{"entry title", "Test Entry"},
		{"author", "John Doe"},
		{"feed title", "Example Feed"},
		{"content", "This is test content"},
		{"CSP header", "Content-Security-Policy"},
		{"generator", "Rogue Planet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(output, tt.contains) {
				t.Errorf("Output should contain %q", tt.contains)
			}
		})
	}
}

func TestGenerateGroupByDate(t *testing.T) {
	gen, _ := New()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	data := TemplateData{
		Title:       "Test Planet",
		GroupByDate: true,
		Entries: []EntryData{
			{
				Title:     "Today's Entry",
				Link:      "https://example.com/today",
				FeedTitle: "Feed 1",
				FeedLink:  "https://example.com/feed1",
				Published: now,
				Content:   template.HTML("<p>Today</p>"),
			},
			{
				Title:     "Yesterday's Entry",
				Link:      "https://example.com/yesterday",
				FeedTitle: "Feed 1",
				FeedLink:  "https://example.com/feed1",
				Published: yesterday,
				Content:   template.HTML("<p>Yesterday</p>"),
			},
		},
	}

	var buf bytes.Buffer
	err := gen.Generate(&buf, data)

	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	output := buf.String()

	// Check for date grouping
	if !strings.Contains(output, "date-group") {
		t.Error("Output should contain date-group class")
	}

	if !strings.Contains(output, "Today") || !strings.Contains(output, "Yesterday") {
		t.Error("Output should contain date group headers")
	}
}

func TestGenerateToFile(t *testing.T) {
	gen, _ := New()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output", "index.html")

	data := TemplateData{
		Title: "Test Planet",
		Entries: []EntryData{
			{
				Title:     "Test Entry",
				Link:      "https://example.com/post1",
				FeedTitle: "Example Feed",
				FeedLink:  "https://example.com/feed",
				Published: time.Now(),
				Content:   template.HTML("<p>Test content</p>"),
			},
		},
	}

	err := gen.GenerateToFile(outputPath, data)
	if err != nil {
		t.Fatalf("GenerateToFile() error = %v", err)
	}

	// Check that file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file should exist")
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !strings.Contains(string(content), "Test Planet") {
		t.Error("Output file should contain title")
	}
}

func TestNewWithTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "custom.html")

	customTemplate := `<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>
<h1>{{.Title}}</h1>
{{range .Entries}}
<div>{{.Title}}</div>
{{end}}
</body>
</html>`

	err := os.WriteFile(templatePath, []byte(customTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	gen, err := NewWithTemplate(templatePath)
	if err != nil {
		t.Fatalf("NewWithTemplate() error = %v", err)
	}

	data := TemplateData{
		Title: "Custom Template Test",
		Entries: []EntryData{
			{
				Title:     "Entry 1",
				Published: time.Now(),
			},
		},
	}

	var buf bytes.Buffer
	err = gen.Generate(&buf, data)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Custom Template Test") {
		t.Error("Should use custom template")
	}
}

func TestRelativeTime(t *testing.T) {
	// Fixed current time for deterministic testing
	currentTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		publishedAt time.Time
		expected    string
	}{
		// "just now" - less than 1 minute
		{
			name:        "just now - 0 seconds",
			publishedAt: currentTime,
			expected:    "just now",
		},
		{
			name:        "just now - 30 seconds",
			publishedAt: currentTime.Add(-30 * time.Second),
			expected:    "just now",
		},
		{
			name:        "just now - 59 seconds (boundary)",
			publishedAt: currentTime.Add(-59 * time.Second),
			expected:    "just now",
		},

		// Minutes - 1 to 59 minutes
		{
			name:        "1 minute ago - 60 seconds exact (boundary)",
			publishedAt: currentTime.Add(-60 * time.Second),
			expected:    "1 minute ago",
		},
		{
			name:        "1 minute ago - 90 seconds",
			publishedAt: currentTime.Add(-90 * time.Second),
			expected:    "1 minute ago",
		},
		{
			name:        "5 minutes ago",
			publishedAt: currentTime.Add(-5 * time.Minute),
			expected:    "5 minutes ago",
		},
		{
			name:        "30 minutes ago",
			publishedAt: currentTime.Add(-30 * time.Minute),
			expected:    "30 minutes ago",
		},
		{
			name:        "59 minutes ago (boundary)",
			publishedAt: currentTime.Add(-59 * time.Minute),
			expected:    "59 minutes ago",
		},

		// Hours - 1 to 23 hours
		{
			name:        "1 hour ago - 60 minutes exact (boundary)",
			publishedAt: currentTime.Add(-60 * time.Minute),
			expected:    "1 hour ago",
		},
		{
			name:        "3 hours ago",
			publishedAt: currentTime.Add(-3 * time.Hour),
			expected:    "3 hours ago",
		},
		{
			name:        "12 hours ago",
			publishedAt: currentTime.Add(-12 * time.Hour),
			expected:    "12 hours ago",
		},
		{
			name:        "23 hours ago (boundary)",
			publishedAt: currentTime.Add(-23 * time.Hour),
			expected:    "23 hours ago",
		},

		// Yesterday - 24 to 47 hours
		{
			name:        "yesterday - 24 hours exact (boundary)",
			publishedAt: currentTime.Add(-24 * time.Hour),
			expected:    "yesterday",
		},
		{
			name:        "yesterday - 25 hours",
			publishedAt: currentTime.Add(-25 * time.Hour),
			expected:    "yesterday",
		},
		{
			name:        "yesterday - 47 hours (boundary)",
			publishedAt: currentTime.Add(-47 * time.Hour),
			expected:    "yesterday",
		},

		// Days - 2 to 6 days
		{
			name:        "2 days ago - 48 hours exact (boundary)",
			publishedAt: currentTime.Add(-48 * time.Hour),
			expected:    "2 days ago",
		},
		{
			name:        "3 days ago",
			publishedAt: currentTime.Add(-72 * time.Hour),
			expected:    "3 days ago",
		},
		{
			name:        "6 days ago (boundary)",
			publishedAt: currentTime.Add(-6 * 24 * time.Hour),
			expected:    "6 days ago",
		},

		// Weeks - 7 to 29 days
		{
			name:        "1 week ago - 7 days exact (boundary)",
			publishedAt: currentTime.Add(-7 * 24 * time.Hour),
			expected:    "1 week ago",
		},
		{
			name:        "10 days ago",
			publishedAt: currentTime.Add(-10 * 24 * time.Hour),
			expected:    "1 week ago",
		},
		{
			name:        "2 weeks ago - 14 days exact",
			publishedAt: currentTime.Add(-14 * 24 * time.Hour),
			expected:    "2 weeks ago",
		},
		{
			name:        "3 weeks ago - 21 days",
			publishedAt: currentTime.Add(-21 * 24 * time.Hour),
			expected:    "3 weeks ago",
		},
		{
			name:        "4 weeks ago - 29 days (boundary)",
			publishedAt: currentTime.Add(-29 * 24 * time.Hour),
			expected:    "4 weeks ago",
		},

		// Months - 30 to 364 days
		{
			name:        "1 month ago - 30 days exact (boundary)",
			publishedAt: currentTime.Add(-30 * 24 * time.Hour),
			expected:    "1 month ago",
		},
		{
			name:        "1 month ago - 35 days",
			publishedAt: currentTime.Add(-35 * 24 * time.Hour),
			expected:    "1 month ago",
		},
		{
			name:        "2 months ago - 60 days",
			publishedAt: currentTime.Add(-60 * 24 * time.Hour),
			expected:    "2 months ago",
		},
		{
			name:        "6 months ago - 180 days",
			publishedAt: currentTime.Add(-180 * 24 * time.Hour),
			expected:    "6 months ago",
		},
		{
			name:        "12 months ago - 364 days (boundary)",
			publishedAt: currentTime.Add(-364 * 24 * time.Hour),
			expected:    "12 months ago",
		},

		// Years - 365+ days
		{
			name:        "1 year ago - 365 days exact (boundary)",
			publishedAt: currentTime.Add(-365 * 24 * time.Hour),
			expected:    "1 year ago",
		},
		{
			name:        "1 year ago - 400 days",
			publishedAt: currentTime.Add(-400 * 24 * time.Hour),
			expected:    "1 year ago",
		},
		{
			name:        "2 years ago - 730 days",
			publishedAt: currentTime.Add(-730 * 24 * time.Hour),
			expected:    "2 years ago",
		},
		{
			name:        "5 years ago - 1825 days",
			publishedAt: currentTime.Add(-1825 * 24 * time.Hour),
			expected:    "5 years ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock := timeprovider.NewFakeClock(currentTime)
			result := relativeTime(tt.publishedAt, clock)
			if result != tt.expected {
				t.Errorf("relativeTime() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGroupEntriesByDate(t *testing.T) {
	// Use fixed time for deterministic testing
	currentTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	clock := timeprovider.NewFakeClock(currentTime)

	now := currentTime
	yesterday := now.Add(-24 * time.Hour)

	entries := []EntryData{
		{Title: "Entry 1", Published: now},
		{Title: "Entry 2", Published: now},
		{Title: "Entry 3", Published: yesterday},
	}

	groups := groupEntriesByDate(entries, clock)

	if len(groups) != 2 {
		t.Fatalf("Expected 2 groups, got %d", len(groups))
	}

	// First group should have 2 entries (today)
	if len(groups[0].Entries) != 2 {
		t.Errorf("First group should have 2 entries, got %d", len(groups[0].Entries))
	}

	// Second group should have 1 entry (yesterday)
	if len(groups[1].Entries) != 1 {
		t.Errorf("Second group should have 1 entry, got %d", len(groups[1].Entries))
	}
}

func TestFormatDateGroup(t *testing.T) {
	// Use fixed time for deterministic testing
	// Wednesday, January 15, 2025 at 14:30 UTC
	currentTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	clock := timeprovider.NewFakeClock(currentTime)

	today := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	yesterday := time.Date(2025, 1, 14, 0, 0, 0, 0, time.UTC)
	lastWeek := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)    // 5 days ago (Friday)
	longAgo := time.Date(2024, 12, 16, 0, 0, 0, 0, time.UTC)    // 30 days ago

	tests := []struct {
		name     string
		date     time.Time
		expected string
	}{
		{
			name:     "today",
			date:     today,
			expected: "Today",
		},
		{
			name:     "yesterday",
			date:     yesterday,
			expected: "Yesterday",
		},
		{
			name:     "last week (within 7 days)",
			date:     lastWeek,
			expected: "Friday, January 10",
		},
		{
			name:     "long ago (more than 7 days)",
			date:     longAgo,
			expected: "Monday, December 16, 2024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDateGroup(tt.date, clock)
			if result != tt.expected {
				t.Errorf("formatDateGroup() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestHTMLSanitization(t *testing.T) {
	gen, _ := New()

	// Content should already be sanitized by normalizer
	// but template should render it as-is (as template.HTML)
	data := TemplateData{
		Title: "Test Planet",
		Entries: []EntryData{
			{
				Title:     "Test",
				Link:      "https://example.com",
				FeedTitle: "Feed",
				FeedLink:  "https://example.com/feed",
				Published: time.Now(),
				Content:   template.HTML("<p>Safe HTML</p>"),
			},
		},
	}

	var buf bytes.Buffer
	err := gen.Generate(&buf, data)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	output := buf.String()

	// Should render the HTML as-is (not escaped)
	if !strings.Contains(output, "<p>Safe HTML</p>") {
		t.Error("Should render template.HTML without escaping")
	}
}

func TestCSPHeader(t *testing.T) {
	gen, _ := New()

	data := TemplateData{
		Title:   "Test",
		Entries: []EntryData{},
	}

	var buf bytes.Buffer
	if err := gen.Generate(&buf, data); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	output := buf.String()

	// Check for CSP header
	if !strings.Contains(output, "Content-Security-Policy") {
		t.Error("Should include CSP header")
	}

	// Check for strict CSP directives
	cspChecks := []string{
		"default-src 'self'",
		"script-src 'self'",
		"object-src 'none'",
		"base-uri 'self'",
	}

	for _, check := range cspChecks {
		if !strings.Contains(output, check) {
			t.Errorf("CSP should contain %q", check)
		}
	}
}

func TestResponsiveDesign(t *testing.T) {
	gen, _ := New()

	data := TemplateData{
		Title:   "Test",
		Entries: []EntryData{},
	}

	var buf bytes.Buffer
	if err := gen.Generate(&buf, data); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	output := buf.String()

	// Check for viewport meta tag
	if !strings.Contains(output, `<meta name="viewport"`) {
		t.Error("Should include viewport meta tag for responsive design")
	}

	// Check for media query in CSS
	if !strings.Contains(output, "@media") {
		t.Error("Should include media queries for responsive design")
	}
}

func TestGenerateWithFeeds(t *testing.T) {
	gen, _ := New()

	data := TemplateData{
		Title: "Test Planet",
		Feeds: []FeedData{
			{
				Title:       "Test Feed 1",
				Link:        "https://example.com",
				URL:         "https://example.com/feed",
				LastUpdated: time.Now().Add(-2 * time.Hour),
				ErrorCount:  0,
			},
			{
				Title:       "Test Feed 2",
				Link:        "https://example.org",
				URL:         "https://example.org/feed",
				LastUpdated: time.Now().Add(-1 * time.Hour),
				ErrorCount:  3,
			},
		},
		Entries: []EntryData{},
	}

	var buf bytes.Buffer
	err := gen.Generate(&buf, data)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	output := buf.String()

	// Check sidebar rendering
	if !strings.Contains(output, "sidebar") {
		t.Error("Should include sidebar when feeds provided")
	}

	if !strings.Contains(output, "Subscriptions") {
		t.Error("Should include subscriptions header")
	}

	if !strings.Contains(output, "Test Feed 1") {
		t.Error("Should include feed 1 title")
	}

	if !strings.Contains(output, "Test Feed 2") {
		t.Error("Should include feed 2 title")
	}

	if !strings.Contains(output, "3 errors") {
		t.Error("Should show error count for feed with errors")
	}
}

func TestGenerateWithOwnerInfo(t *testing.T) {
	gen, _ := New()

	data := TemplateData{
		Title:      "Test Planet",
		OwnerName:  "Jane Doe",
		OwnerEmail: "jane@example.com",
		Entries:    []EntryData{},
	}

	var buf bytes.Buffer
	err := gen.Generate(&buf, data)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Jane Doe") {
		t.Error("Should include owner name")
	}
}

func TestTemplateFuncs(t *testing.T) {
	gen, _ := New()

	testTime := time.Date(2025, 10, 12, 15, 30, 0, 0, time.UTC)

	data := TemplateData{
		Title: "Test",
		Entries: []EntryData{
			{
				Title:     "Test Entry",
				Link:      "https://example.com",
				FeedTitle: "Feed",
				FeedLink:  "https://example.com/feed",
				Published: testTime,
				Content:   template.HTML("<p>Content</p>"),
			},
		},
	}

	var buf bytes.Buffer
	err := gen.Generate(&buf, data)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	output := buf.String()

	// Check ISO date format in datetime attribute
	if !strings.Contains(output, "2025-10-12T15:30:00Z") {
		t.Error("Should include ISO formatted date")
	}
}

func TestRelativeTimeEdgeCases(t *testing.T) {
	// Fixed current time for deterministic testing
	currentTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		publishedAt time.Time
		expected    string
	}{
		{
			name:        "3 months ago - 90 days",
			publishedAt: currentTime.Add(-90 * 24 * time.Hour),
			expected:    "3 months ago",
		},
		{
			name:        "1 year ago - 400 days",
			publishedAt: currentTime.Add(-400 * 24 * time.Hour),
			expected:    "1 year ago",
		},
		{
			name:        "2 years ago - 800 days",
			publishedAt: currentTime.Add(-800 * 24 * time.Hour),
			expected:    "2 years ago",
		},
		{
			name:        "future time (negative duration)",
			publishedAt: currentTime.Add(1 * time.Hour),
			expected:    "just now", // Future times handled gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock := timeprovider.NewFakeClock(currentTime)
			result := relativeTime(tt.publishedAt, clock)
			if result != tt.expected {
				t.Errorf("relativeTime() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNewWithTemplateError(t *testing.T) {
	_, err := NewWithTemplate("/nonexistent/template.html")
	if err == nil {
		t.Error("NewWithTemplate() should error for non-existent file")
	}
}

func TestNewWithTemplateBadSyntax(t *testing.T) {
	t.Run("unclosed template action", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "bad.html")

		// Write template with unclosed action - this is a parse error
		badTemplate := `<!DOCTYPE html><html><body>{{.Title</body></html>`
		err := os.WriteFile(templatePath, []byte(badTemplate), 0644)
		if err != nil {
			t.Fatalf("Failed to write template: %v", err)
		}

		_, err = NewWithTemplate(templatePath)
		if err == nil {
			t.Error("Expected error for template with unclosed action, got nil")
		}
	})

	t.Run("undefined function", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "bad.html")

		// Write template with undefined function - this may fail at execution time
		badTemplate := `<!DOCTYPE html><html><body>{{.Title}} {{undefinedFunc .Title}}</body></html>`
		err := os.WriteFile(templatePath, []byte(badTemplate), 0644)
		if err != nil {
			t.Fatalf("Failed to write template: %v", err)
		}

		gen, err := NewWithTemplate(templatePath)
		if err != nil {
			// Parse-time error is acceptable
			return
		}

		// Try to execute - should fail
		data := TemplateData{Title: "Test"}
		var buf bytes.Buffer
		err = gen.Generate(&buf, data)
		if err == nil {
			t.Error("Expected error for undefined function, got nil")
		}
	})

	t.Run("invalid field access", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "bad.html")

		// Write template accessing non-existent field
		badTemplate := `<!DOCTYPE html><html><body>{{.NonExistentField}}</body></html>`
		err := os.WriteFile(templatePath, []byte(badTemplate), 0644)
		if err != nil {
			t.Fatalf("Failed to write template: %v", err)
		}

		gen, err := NewWithTemplate(templatePath)
		if err != nil {
			t.Fatalf("Template parsing should succeed, got: %v", err)
		}

		// Execute with empty data
		data := TemplateData{Title: "Test"}
		var buf bytes.Buffer
		err = gen.Generate(&buf, data)
		if err == nil {
			t.Error("Expected error for accessing non-existent field, got nil")
		}
	})
}

func TestCopyStaticAssets(t *testing.T) {
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	staticSrc := filepath.Join(templateDir, "static")
	outputDir := filepath.Join(tmpDir, "output")

	// Create template directory structure
	if err := os.MkdirAll(staticSrc, 0755); err != nil {
		t.Fatalf("Failed to create static dir: %v", err)
	}

	// Create some static files
	cssFile := filepath.Join(staticSrc, "style.css")
	if err := os.WriteFile(cssFile, []byte("body { color: red; }"), 0644); err != nil {
		t.Fatalf("Failed to write CSS file: %v", err)
	}

	// Create subdirectory with file
	jsDir := filepath.Join(staticSrc, "js")
	if err := os.MkdirAll(jsDir, 0755); err != nil {
		t.Fatalf("Failed to create js dir: %v", err)
	}

	jsFile := filepath.Join(jsDir, "app.js")
	if err := os.WriteFile(jsFile, []byte("console.log('test');"), 0644); err != nil {
		t.Fatalf("Failed to write JS file: %v", err)
	}

	// Create template
	templatePath := filepath.Join(templateDir, "template.html")
	if err := os.WriteFile(templatePath, []byte("<html><head></head><body>{{.Title}}</body></html>"), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	// Create generator with custom template
	gen, err := NewWithTemplate(templatePath)
	if err != nil {
		t.Fatalf("NewWithTemplate() error = %v", err)
	}

	// Copy static assets
	if err := gen.CopyStaticAssets(outputDir); err != nil {
		t.Fatalf("CopyStaticAssets() error = %v", err)
	}

	// Verify files were copied
	copiedCSS := filepath.Join(outputDir, "static", "style.css")
	if _, err := os.Stat(copiedCSS); os.IsNotExist(err) {
		t.Error("CSS file should be copied")
	}

	copiedJS := filepath.Join(outputDir, "static", "js", "app.js")
	if _, err := os.Stat(copiedJS); os.IsNotExist(err) {
		t.Error("JS file should be copied")
	}

	// Verify content
	cssContent, _ := os.ReadFile(copiedCSS)
	if !strings.Contains(string(cssContent), "color: red") {
		t.Error("CSS content should match original")
	}
}

func TestCopyStaticAssetsNoStaticDir(t *testing.T) {
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	outputDir := filepath.Join(tmpDir, "output")

	// Create template without static directory
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("Failed to create template dir: %v", err)
	}

	templatePath := filepath.Join(templateDir, "template.html")
	if err := os.WriteFile(templatePath, []byte("<html><body>{{.Title}}</body></html>"), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	gen, err := NewWithTemplate(templatePath)
	if err != nil {
		t.Fatalf("NewWithTemplate() error = %v", err)
	}

	// Should not error when no static directory exists
	if err := gen.CopyStaticAssets(outputDir); err != nil {
		t.Errorf("CopyStaticAssets() should not error when static dir doesn't exist: %v", err)
	}
}

func TestCopyStaticAssetsDefaultTemplate(t *testing.T) {
	gen, _ := New()

	tmpDir := t.TempDir()

	// Should not copy anything for default template
	err := gen.CopyStaticAssets(tmpDir)
	if err != nil {
		t.Errorf("CopyStaticAssets() should not error for default template: %v", err)
	}
}

func TestGenerateToFileWithStaticAssets(t *testing.T) {
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	staticSrc := filepath.Join(templateDir, "static")
	outputPath := filepath.Join(tmpDir, "output", "index.html")

	// Create static directory with file
	if err := os.MkdirAll(staticSrc, 0755); err != nil {
		t.Fatalf("Failed to create static dir: %v", err)
	}

	cssFile := filepath.Join(staticSrc, "style.css")
	if err := os.WriteFile(cssFile, []byte("body { }"), 0644); err != nil {
		t.Fatalf("Failed to write CSS: %v", err)
	}

	// Create template
	templatePath := filepath.Join(templateDir, "template.html")
	if err := os.WriteFile(templatePath, []byte("<html><body>{{.Title}}</body></html>"), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	gen, err := NewWithTemplate(templatePath)
	if err != nil {
		t.Fatalf("NewWithTemplate() error = %v", err)
	}

	data := TemplateData{Title: "Test"}

	// Generate to file (should also copy static assets)
	if err := gen.GenerateToFile(outputPath, data); err != nil {
		t.Fatalf("GenerateToFile() error = %v", err)
	}

	// Verify HTML file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output HTML file should exist")
	}

	// Verify static assets were copied
	copiedCSS := filepath.Join(tmpDir, "output", "static", "style.css")
	if _, err := os.Stat(copiedCSS); os.IsNotExist(err) {
		t.Error("Static CSS file should be copied")
	}
}

func TestGenerateWithSubtitle(t *testing.T) {
	gen, _ := New()

	data := TemplateData{
		Title:    "My Planet",
		Subtitle: "A collection of interesting blogs",
		Link:     "https://planet.example.com",
		Entries:  []EntryData{},
	}

	var buf bytes.Buffer
	err := gen.Generate(&buf, data)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "A collection of interesting blogs") {
		t.Error("Should include subtitle")
	}

	// Check that title is a link
	if !strings.Contains(output, `<a href="https://planet.example.com">My Planet</a>`) {
		t.Error("Title should be a link when Link is provided")
	}
}
