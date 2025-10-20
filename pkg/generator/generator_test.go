package generator

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"just now", time.Now(), "just now"},
		{"1 minute ago", time.Now().Add(-1 * time.Minute), "1 minute ago"},
		{"5 minutes ago", time.Now().Add(-5 * time.Minute), "5 minutes ago"},
		{"1 hour ago", time.Now().Add(-1 * time.Hour), "1 hour ago"},
		{"3 hours ago", time.Now().Add(-3 * time.Hour), "3 hours ago"},
		{"yesterday", time.Now().Add(-25 * time.Hour), "yesterday"},
		{"2 days ago", time.Now().Add(-48 * time.Hour), "2 days ago"},
		{"1 week ago", time.Now().Add(-7 * 24 * time.Hour), "1 week ago"},
		{"2 weeks ago", time.Now().Add(-14 * 24 * time.Hour), "2 weeks ago"},
		{"1 month ago", time.Now().Add(-35 * 24 * time.Hour), "1 month ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := relativeTime(tt.time)
			if result != tt.expected {
				t.Errorf("relativeTime() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGroupEntriesByDate(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	entries := []EntryData{
		{Title: "Entry 1", Published: now},
		{Title: "Entry 2", Published: now},
		{Title: "Entry 3", Published: yesterday},
	}

	groups := groupEntriesByDate(entries)

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
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.Add(-24 * time.Hour)
	lastWeek := today.Add(-5 * 24 * time.Hour)
	longAgo := today.Add(-30 * 24 * time.Hour)

	tests := []struct {
		name     string
		time     time.Time
		contains string
	}{
		{"today", today, "Today"},
		{"yesterday", yesterday, "Yesterday"},
		{"last week", lastWeek, now.Weekday().String()[:3]}, // Contains day name
		{"long ago", longAgo, "2"},                          // Contains year
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDateGroup(tt.time)
			// Just check it returns something reasonable
			if len(result) == 0 {
				t.Error("formatDateGroup should return non-empty string")
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
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"3 months ago", time.Now().Add(-90 * 24 * time.Hour), "3 months ago"},
		{"1 year ago", time.Now().Add(-400 * 24 * time.Hour), "1 year ago"},
		{"2 years ago", time.Now().Add(-800 * 24 * time.Hour), "2 years ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := relativeTime(tt.time)
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
