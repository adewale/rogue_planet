package main

import (
	"strings"
	"testing"
)

func TestParseInitFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		wantFile  string
		wantError bool
	}{
		{
			name:      "no flags",
			args:      []string{},
			wantFile:  "",
			wantError: false,
		},
		{
			name:      "with feeds file",
			args:      []string{"-f", "feeds.txt"},
			wantFile:  "feeds.txt",
			wantError: false,
		},
		{
			name:      "invalid flag",
			args:      []string{"-invalid"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := parseInitFlags(tt.args)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if opts.FeedsFile != tt.wantFile {
				t.Errorf("FeedsFile = %q, want %q", opts.FeedsFile, tt.wantFile)
			}
			if opts.ConfigPath != "config.ini" {
				t.Errorf("ConfigPath = %q, want %q", opts.ConfigPath, "config.ini")
			}
		})
	}
}

func TestParseAddFeedFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		wantURL    string
		wantConfig string
		wantError  bool
	}{
		{
			name:       "url only",
			args:       []string{"https://example.com/feed.xml"},
			wantURL:    "https://example.com/feed.xml",
			wantConfig: "./config.ini",
			wantError:  false,
		},
		{
			name:       "url with custom config",
			args:       []string{"-config", "/tmp/config.ini", "https://example.com/feed.xml"},
			wantURL:    "https://example.com/feed.xml",
			wantConfig: "/tmp/config.ini",
			wantError:  false,
		},
		{
			name:      "missing url",
			args:      []string{},
			wantError: true,
		},
		{
			name:      "invalid flag",
			args:      []string{"-bad", "value"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := parseAddFeedFlags(tt.args)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if opts.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", opts.URL, tt.wantURL)
			}
			if opts.ConfigPath != tt.wantConfig {
				t.Errorf("ConfigPath = %q, want %q", opts.ConfigPath, tt.wantConfig)
			}
		})
	}
}

func TestParseAddAllFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		wantFile   string
		wantConfig string
		wantError  bool
	}{
		{
			name:       "feeds file only",
			args:       []string{"-f", "feeds.txt"},
			wantFile:   "feeds.txt",
			wantConfig: "./config.ini",
			wantError:  false,
		},
		{
			name:       "feeds file with custom config",
			args:       []string{"-f", "feeds.txt", "-config", "/tmp/config.ini"},
			wantFile:   "feeds.txt",
			wantConfig: "/tmp/config.ini",
			wantError:  false,
		},
		{
			name:      "missing feeds file",
			args:      []string{},
			wantError: true,
		},
		{
			name:      "empty feeds file flag",
			args:      []string{"-f", ""},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := parseAddAllFlags(tt.args)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "feeds file") {
					t.Errorf("error should mention 'feeds file', got: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if opts.FeedsFile != tt.wantFile {
				t.Errorf("FeedsFile = %q, want %q", opts.FeedsFile, tt.wantFile)
			}
			if opts.ConfigPath != tt.wantConfig {
				t.Errorf("ConfigPath = %q, want %q", opts.ConfigPath, tt.wantConfig)
			}
		})
	}
}

func TestParseRemoveFeedFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		wantURL    string
		wantConfig string
		wantForce  bool
		wantError  bool
	}{
		{
			name:       "url only",
			args:       []string{"https://example.com/feed.xml"},
			wantURL:    "https://example.com/feed.xml",
			wantConfig: "./config.ini",
			wantForce:  false,
			wantError:  false,
		},
		{
			name:       "url with force",
			args:       []string{"-force", "https://example.com/feed.xml"},
			wantURL:    "https://example.com/feed.xml",
			wantConfig: "./config.ini",
			wantForce:  true,
			wantError:  false,
		},
		{
			name:       "url with custom config and force",
			args:       []string{"-config", "/tmp/config.ini", "-force", "https://example.com/feed.xml"},
			wantURL:    "https://example.com/feed.xml",
			wantConfig: "/tmp/config.ini",
			wantForce:  true,
			wantError:  false,
		},
		{
			name:      "missing url",
			args:      []string{"-force"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := parseRemoveFeedFlags(tt.args)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if opts.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", opts.URL, tt.wantURL)
			}
			if opts.ConfigPath != tt.wantConfig {
				t.Errorf("ConfigPath = %q, want %q", opts.ConfigPath, tt.wantConfig)
			}
			if opts.Force != tt.wantForce {
				t.Errorf("Force = %v, want %v", opts.Force, tt.wantForce)
			}
		})
	}
}

func TestParseGenerateFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		wantDays   int
		wantConfig string
		wantError  bool
	}{
		{
			name:       "no flags",
			args:       []string{},
			wantDays:   0,
			wantConfig: "./config.ini",
			wantError:  false,
		},
		{
			name:       "with days",
			args:       []string{"-days", "14"},
			wantDays:   14,
			wantConfig: "./config.ini",
			wantError:  false,
		},
		{
			name:       "with days and config",
			args:       []string{"-days", "30", "-config", "/tmp/config.ini"},
			wantDays:   30,
			wantConfig: "/tmp/config.ini",
			wantError:  false,
		},
		{
			name:      "invalid days value",
			args:      []string{"-days", "invalid"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := parseGenerateFlags(tt.args)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if opts.Days != tt.wantDays {
				t.Errorf("Days = %d, want %d", opts.Days, tt.wantDays)
			}
			if opts.ConfigPath != tt.wantConfig {
				t.Errorf("ConfigPath = %q, want %q", opts.ConfigPath, tt.wantConfig)
			}
		})
	}
}

func TestParsePruneFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		wantDays   int
		wantDryRun bool
		wantConfig string
		wantError  bool
	}{
		{
			name:       "default days",
			args:       []string{},
			wantDays:   90,
			wantDryRun: false,
			wantConfig: "./config.ini",
			wantError:  false,
		},
		{
			name:       "custom days",
			args:       []string{"-days", "180"},
			wantDays:   180,
			wantDryRun: false,
			wantConfig: "./config.ini",
			wantError:  false,
		},
		{
			name:       "with dry-run",
			args:       []string{"-dry-run"},
			wantDays:   90,
			wantDryRun: true,
			wantConfig: "./config.ini",
			wantError:  false,
		},
		{
			name:       "all flags",
			args:       []string{"-days", "365", "-dry-run", "-config", "/tmp/config.ini"},
			wantDays:   365,
			wantDryRun: true,
			wantConfig: "/tmp/config.ini",
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := parsePruneFlags(tt.args)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if opts.Days != tt.wantDays {
				t.Errorf("Days = %d, want %d", opts.Days, tt.wantDays)
			}
			if opts.DryRun != tt.wantDryRun {
				t.Errorf("DryRun = %v, want %v", opts.DryRun, tt.wantDryRun)
			}
			if opts.ConfigPath != tt.wantConfig {
				t.Errorf("ConfigPath = %q, want %q", opts.ConfigPath, tt.wantConfig)
			}
		})
	}
}

func TestParseUpdateFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		wantVerbose bool
		wantConfig  string
		wantError   bool
	}{
		{
			name:        "no flags",
			args:        []string{},
			wantVerbose: false,
			wantConfig:  "./config.ini",
			wantError:   false,
		},
		{
			name:        "with verbose",
			args:        []string{"-verbose"},
			wantVerbose: true,
			wantConfig:  "./config.ini",
			wantError:   false,
		},
		{
			name:        "verbose with custom config",
			args:        []string{"-verbose", "-config", "/tmp/config.ini"},
			wantVerbose: true,
			wantConfig:  "/tmp/config.ini",
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := parseUpdateFlags(tt.args)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if opts.Verbose != tt.wantVerbose {
				t.Errorf("Verbose = %v, want %v", opts.Verbose, tt.wantVerbose)
			}
			if opts.ConfigPath != tt.wantConfig {
				t.Errorf("ConfigPath = %q, want %q", opts.ConfigPath, tt.wantConfig)
			}
			if opts.Logger == nil {
				t.Error("Logger should not be nil")
			}
		})
	}
}

func TestParseImportOPMLFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		wantFile   string
		wantDryRun bool
		wantConfig string
		wantError  bool
	}{
		{
			name:       "file only",
			args:       []string{"feeds.opml"},
			wantFile:   "feeds.opml",
			wantDryRun: false,
			wantConfig: "./config.ini",
			wantError:  false,
		},
		{
			name:       "file with dry-run",
			args:       []string{"-dry-run", "feeds.opml"},
			wantFile:   "feeds.opml",
			wantDryRun: true,
			wantConfig: "./config.ini",
			wantError:  false,
		},
		{
			name:       "all flags",
			args:       []string{"-dry-run", "-config", "/tmp/config.ini", "feeds.opml"},
			wantFile:   "feeds.opml",
			wantDryRun: true,
			wantConfig: "/tmp/config.ini",
			wantError:  false,
		},
		{
			name:      "missing file",
			args:      []string{"-dry-run"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := parseImportOPMLFlags(tt.args)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if opts.OPMLFile != tt.wantFile {
				t.Errorf("OPMLFile = %q, want %q", opts.OPMLFile, tt.wantFile)
			}
			if opts.DryRun != tt.wantDryRun {
				t.Errorf("DryRun = %v, want %v", opts.DryRun, tt.wantDryRun)
			}
			if opts.ConfigPath != tt.wantConfig {
				t.Errorf("ConfigPath = %q, want %q", opts.ConfigPath, tt.wantConfig)
			}
		})
	}
}

func TestParseExportOPMLFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		wantOutput string
		wantConfig string
		wantError  bool
	}{
		{
			name:       "no flags (stdout)",
			args:       []string{},
			wantOutput: "",
			wantConfig: "./config.ini",
			wantError:  false,
		},
		{
			name:       "with output file",
			args:       []string{"-output", "feeds.opml"},
			wantOutput: "feeds.opml",
			wantConfig: "./config.ini",
			wantError:  false,
		},
		{
			name:       "with output and config",
			args:       []string{"-output", "feeds.opml", "-config", "/tmp/config.ini"},
			wantOutput: "feeds.opml",
			wantConfig: "/tmp/config.ini",
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := parseExportOPMLFlags(tt.args)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if opts.OutputFile != tt.wantOutput {
				t.Errorf("OutputFile = %q, want %q", opts.OutputFile, tt.wantOutput)
			}
			if opts.ConfigPath != tt.wantConfig {
				t.Errorf("ConfigPath = %q, want %q", opts.ConfigPath, tt.wantConfig)
			}
		})
	}
}

// Test simple parsers with less complexity
func TestParseListFeedsFlags(t *testing.T) {
	t.Parallel()

	opts, err := parseListFeedsFlags([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.ConfigPath != "./config.ini" {
		t.Errorf("ConfigPath = %q, want %q", opts.ConfigPath, "./config.ini")
	}

	opts, err = parseListFeedsFlags([]string{"-config", "/tmp/config.ini"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.ConfigPath != "/tmp/config.ini" {
		t.Errorf("ConfigPath = %q, want %q", opts.ConfigPath, "/tmp/config.ini")
	}
}

func TestParseStatusFlags(t *testing.T) {
	t.Parallel()

	opts, err := parseStatusFlags([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.ConfigPath != "./config.ini" {
		t.Errorf("ConfigPath = %q, want %q", opts.ConfigPath, "./config.ini")
	}
}

func TestParseVerifyFlags(t *testing.T) {
	t.Parallel()

	opts, err := parseVerifyFlags([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.ConfigPath != "./config.ini" {
		t.Errorf("ConfigPath = %q, want %q", opts.ConfigPath, "./config.ini")
	}
}

func TestParseFetchFlags(t *testing.T) {
	t.Parallel()

	opts, err := parseFetchFlags([]string{"-verbose"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.Verbose {
		t.Error("Verbose should be true")
	}
	if opts.Logger == nil {
		t.Error("Logger should not be nil")
	}
}
