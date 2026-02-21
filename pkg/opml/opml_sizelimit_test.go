package opml

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Tests for L15: OPML file size limit

func TestMaxOPMLFileSize_Constant(t *testing.T) {
	t.Parallel()

	// Verify the constant is 10MB
	expected := int64(10 * 1024 * 1024)
	if MaxOPMLFileSize != expected {
		t.Errorf("MaxOPMLFileSize = %d, want %d (10MB)", MaxOPMLFileSize, expected)
	}
}

func TestParseFile_RejectsOversizedFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	filePath := filepath.Join(dir, "huge.opml")

	// Create a file just over MaxOPMLFileSize
	// Write a valid OPML header, then pad with spaces
	header := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head><title>Huge</title></head>
  <body>
    <outline text="Feed 1" xmlUrl="https://example.com/1"/>
  </body>
</opml>`

	padding := strings.Repeat(" ", int(MaxOPMLFileSize)-len(header)+1)
	data := []byte(header + padding)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		t.Fatalf("Failed to write oversized file: %v", err)
	}

	_, err := ParseFile(context.Background(), filePath)
	if err == nil {
		t.Error("Expected error for oversized OPML file, got nil")
	}

	if !strings.Contains(err.Error(), "exceeds maximum") {
		t.Errorf("Error should mention exceeds maximum size, got: %v", err)
	}
}

func TestParseFile_AcceptsFileJustUnderLimit(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	filePath := filepath.Join(dir, "ok.opml")

	// Create a valid OPML file well under the limit
	opmlData := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head><title>Normal Size</title></head>
  <body>
    <outline text="Feed 1" xmlUrl="https://example.com/1"/>
  </body>
</opml>`

	if err := os.WriteFile(filePath, []byte(opmlData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	opml, err := ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("ParseFile should accept file under size limit: %v", err)
	}

	feeds := opml.ExtractFeeds()
	if len(feeds) != 1 {
		t.Errorf("Expected 1 feed, got %d", len(feeds))
	}
}

func TestParse_LimitReader(t *testing.T) {
	t.Parallel()

	// Create data that's well within limits - should parse fine
	opmlData := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head><title>Test</title></head>
  <body>
    <outline text="Feed" xmlUrl="https://example.com/feed"/>
  </body>
</opml>`

	opml, err := Parse([]byte(opmlData))
	if err != nil {
		t.Fatalf("Parse should succeed for normal data: %v", err)
	}

	if opml.Head.Title != "Test" {
		t.Errorf("Title = %q, want %q", opml.Head.Title, "Test")
	}
}

func TestParse_RejectsOversizedData(t *testing.T) {
	t.Parallel()

	// Create data larger than MaxOPMLFileSize
	header := []byte(`<?xml version="1.0" encoding="UTF-8"?><opml version="2.0"><body>`)
	padding := bytes.Repeat([]byte(" "), int(MaxOPMLFileSize)+1)
	footer := []byte(`</body></opml>`)

	data := make([]byte, 0, len(header)+len(padding)+len(footer))
	data = append(data, header...)
	data = append(data, padding...)
	data = append(data, footer...)

	_, err := Parse(data)
	if err == nil {
		t.Error("Expected error for oversized OPML data, got nil")
	}

	if !strings.Contains(err.Error(), "exceeds maximum") {
		t.Errorf("Error should mention exceeds maximum size, got: %v", err)
	}
}
