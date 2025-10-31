package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adewale/rogue_planet/pkg/repository"
)

func TestCmdAddFeed(t *testing.T) {
	tests := []struct {
		name       string
		opts       AddFeedOptions
		wantErr    bool
		wantOutput string
	}{
		{
			name: "missing URL",
			opts: AddFeedOptions{
				URL:        "",
				ConfigPath: "./config.ini",
				Output:     &bytes.Buffer{},
			},
			wantErr:    true,
			wantOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmdAddFeed(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("cmdAddFeed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCmdAddAll(t *testing.T) {
	tests := []struct {
		name       string
		opts       AddAllOptions
		wantErr    bool
		wantOutput string
	}{
		{
			name: "missing feeds file",
			opts: AddAllOptions{
				FeedsFile:  "",
				ConfigPath: "./config.ini",
				Output:     &bytes.Buffer{},
			},
			wantErr:    true,
			wantOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmdAddAll(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("cmdAddAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCmdRemoveFeed(t *testing.T) {
	tests := []struct {
		name       string
		opts       RemoveFeedOptions
		wantErr    bool
		wantOutput string
	}{
		{
			name: "missing URL",
			opts: RemoveFeedOptions{
				URL:        "",
				ConfigPath: "./config.ini",
				Output:     &bytes.Buffer{},
			},
			wantErr:    true,
			wantOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmdRemoveFeed(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("cmdRemoveFeed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCmdInit(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	opts := InitOptions{
		FeedsFile:  "",
		ConfigPath: "config.ini",
		Output:     &buf,
	}

	if err := cmdInit(opts); err != nil {
		t.Fatalf("cmdInit() error = %v", err)
	}

	// Check that directories were created
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		t.Error("data/ directory not created")
	}

	if _, err := os.Stat("public"); os.IsNotExist(err) {
		t.Error("public/ directory not created")
	}

	// Check that config file was created
	if _, err := os.Stat("config.ini"); os.IsNotExist(err) {
		t.Error("config.ini not created")
	}

	// Check output contains success messages
	output := buf.String()
	if !strings.Contains(output, "✓ Created config.ini") {
		t.Error("Output missing config.ini success message")
	}
}

func TestCmdInitWithFeedsFile(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Create a feeds file
	feedsPath := filepath.Join(tmpDir, "feeds.txt")
	feedsContent := "https://blog.golang.org/feed.atom\nhttps://github.blog/feed/\n"
	if err := os.WriteFile(feedsPath, []byte(feedsContent), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	opts := InitOptions{
		FeedsFile:  feedsPath,
		ConfigPath: "config.ini",
		Output:     &buf,
	}

	if err := cmdInit(opts); err != nil {
		t.Fatalf("cmdInit() error = %v", err)
	}

	// Check output
	output := buf.String()
	if !strings.Contains(output, "Importing feeds") {
		t.Error("Output missing feeds import message")
	}
}

func TestCmdPrune(t *testing.T) {
	var buf bytes.Buffer
	opts := PruneOptions{
		ConfigPath: "/nonexistent/config.ini",
		Days:       90,
		DryRun:     true,
		Output:     &buf,
	}

	// With dry-run, should print message and return without error
	err := cmdPrune(opts)
	if err == nil {
		output := buf.String()
		if !strings.Contains(output, "Dry run") {
			t.Error("cmdPrune() dry-run should print 'Dry run' message")
		}
	} else {
		// If it fails, it's because the database doesn't exist, which is also acceptable
		if !strings.Contains(err.Error(), "database") && !strings.Contains(err.Error(), "open") {
			t.Errorf("cmdPrune() unexpected error = %v", err)
		}
	}
}

// Note: Tests for database creation and command behavior are covered by:
// - TestFullWorkflow (integration_test.go) - tests init → add-feed → status
// - TestInitWithFeedsFile (integration_test.go) - tests init with feeds
// - TestCmdInit - tests basic init functionality
//
// The current design auto-creates databases with default configs when needed.
// Previous tests that expected failures with nonexistent databases have been
// removed as they don't match the actual behavior (auto-creation with defaults).

func TestCmdVerify(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) (configPath string, cleanup func())
		wantErr    bool
		wantOutput string
	}{
		{
			name: "database does not exist",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.ini")
				dbPath := filepath.Join(tmpDir, "nonexistent.db")
				outputDir := filepath.Join(tmpDir, "public")
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					t.Fatal(err)
				}

				configContent := `[planet]
name = Test Planet
link = https://example.com
owner_name = Test
owner_email = test@example.com
output_dir = ` + outputDir + `
days = 7

[database]
path = ` + dbPath + `
`
				if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
					t.Fatal(err)
				}
				return configPath, func() {}
			},
			wantErr:    true,
			wantOutput: "Database does not exist",
		},
		{
			name: "output directory does not exist",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.ini")
				dbPath := filepath.Join(tmpDir, "planet.db")
				outputDir := filepath.Join(tmpDir, "nonexistent_public")

				configContent := `[planet]
name = Test Planet
link = https://example.com
owner_name = Test
owner_email = test@example.com
output_dir = ` + outputDir + `
days = 7

[database]
path = ` + dbPath + `
`
				if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
					t.Fatal(err)
				}

				// Create database so we get past that check
				repo, err := repository.New(dbPath)
				if err != nil {
					t.Fatal(err)
				}
				repo.Close()

				return configPath, func() {}
			},
			wantErr:    true,
			wantOutput: "Output directory does not exist",
		},
		{
			name: "template file not found",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.ini")
				dbPath := filepath.Join(tmpDir, "planet.db")
				outputDir := filepath.Join(tmpDir, "public")
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					t.Fatal(err)
				}

				configContent := `[planet]
name = Test Planet
link = https://example.com
owner_name = Test
owner_email = test@example.com
output_dir = ` + outputDir + `
days = 7
template = /nonexistent/template.html

[database]
path = ` + dbPath + `
`
				if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
					t.Fatal(err)
				}

				// Create database so we get past that check
				repo, err := repository.New(dbPath)
				if err != nil {
					t.Fatal(err)
				}
				repo.Close()

				return configPath, func() {}
			},
			wantErr:    true,
			wantOutput: "Template file not found",
		},
		{
			name: "valid configuration",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.ini")
				dbPath := filepath.Join(tmpDir, "planet.db")
				outputDir := filepath.Join(tmpDir, "public")
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					t.Fatal(err)
				}

				configContent := `[planet]
name = Test Planet
link = https://example.com
owner_name = Test
owner_email = test@example.com
output_dir = ` + outputDir + `
days = 7

[database]
path = ` + dbPath + `
`
				if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
					t.Fatal(err)
				}

				// Create database
				repo, err := repository.New(dbPath)
				if err != nil {
					t.Fatal(err)
				}
				repo.Close()

				return configPath, func() {}
			},
			wantErr:    false,
			wantOutput: "✓ Configuration valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath, cleanup := tt.setup(t)
			defer cleanup()

			var buf bytes.Buffer
			opts := VerifyOptions{
				ConfigPath: configPath,
				Output:     &buf,
			}

			err := cmdVerify(opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("cmdVerify() error = %v, wantErr %v", err, tt.wantErr)
			}

			output := buf.String()
			if tt.wantOutput != "" && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("cmdVerify() output = %q, want to contain %q", output, tt.wantOutput)
			}
		})
	}
}

func TestCmdUpdate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (configPath string, cleanup func())
		wantErr bool
	}{
		{
			name: "with valid config",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.ini")
				dbPath := filepath.Join(tmpDir, "planet.db")
				outputDir := filepath.Join(tmpDir, "public")
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					t.Fatal(err)
				}

				configContent := `[planet]
name = Test Planet
link = https://example.com
owner_name = Test
owner_email = test@example.com
output_dir = ` + outputDir + `
days = 7

[database]
path = ` + dbPath + `
`
				if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
					t.Fatal(err)
				}

				return configPath, func() {}
			},
			wantErr: false, // Should auto-create database and succeed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath, cleanup := tt.setup(t)
			defer cleanup()

			var buf bytes.Buffer
			opts := UpdateOptions{
				ConfigPath: configPath,
				Verbose:    false,
				Output:     &buf,
			}

			err := cmdUpdate(opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("cmdUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check that output contains success messages
			output := buf.String()
			if !tt.wantErr && !strings.Contains(output, "complete") {
				t.Errorf("cmdUpdate() output should contain 'complete', got: %q", output)
			}
		})
	}
}

func TestCmdFetch(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (configPath string, cleanup func())
		wantErr bool
	}{
		{
			name: "with valid config",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.ini")
				dbPath := filepath.Join(tmpDir, "planet.db")
				outputDir := filepath.Join(tmpDir, "public")
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					t.Fatal(err)
				}

				configContent := `[planet]
name = Test Planet
link = https://example.com
owner_name = Test
owner_email = test@example.com
output_dir = ` + outputDir + `
days = 7

[database]
path = ` + dbPath + `
`
				if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
					t.Fatal(err)
				}

				return configPath, func() {}
			},
			wantErr: false, // Should auto-create database and succeed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath, cleanup := tt.setup(t)
			defer cleanup()

			var buf bytes.Buffer
			opts := FetchOptions{
				ConfigPath: configPath,
				Verbose:    false,
				Output:     &buf,
			}

			err := cmdFetch(opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("cmdFetch() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check that output contains success messages
			output := buf.String()
			if !tt.wantErr && !strings.Contains(output, "complete") {
				t.Errorf("cmdFetch() output should contain 'complete', got: %q", output)
			}
		})
	}
}

func TestCmdGenerate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (configPath string, cleanup func())
		days    int
		wantErr bool
	}{
		{
			name: "with valid config",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.ini")
				dbPath := filepath.Join(tmpDir, "planet.db")
				outputDir := filepath.Join(tmpDir, "public")
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					t.Fatal(err)
				}

				configContent := `[planet]
name = Test Planet
link = https://example.com
owner_name = Test
owner_email = test@example.com
output_dir = ` + outputDir + `
days = 7

[database]
path = ` + dbPath + `
`
				if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
					t.Fatal(err)
				}

				// Create database
				repo, err := repository.New(dbPath)
				if err != nil {
					t.Fatal(err)
				}
				repo.Close()

				return configPath, func() {}
			},
			days:    0, // Use config default
			wantErr: false,
		},
		{
			name: "with days override",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.ini")
				dbPath := filepath.Join(tmpDir, "planet.db")
				outputDir := filepath.Join(tmpDir, "public")
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					t.Fatal(err)
				}

				configContent := `[planet]
name = Test Planet
link = https://example.com
owner_name = Test
owner_email = test@example.com
output_dir = ` + outputDir + `
days = 7

[database]
path = ` + dbPath + `
`
				if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
					t.Fatal(err)
				}

				// Create database
				repo, err := repository.New(dbPath)
				if err != nil {
					t.Fatal(err)
				}
				repo.Close()

				return configPath, func() {}
			},
			days:    30, // Override to 30 days
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath, cleanup := tt.setup(t)
			defer cleanup()

			var buf bytes.Buffer
			opts := GenerateOptions{
				ConfigPath: configPath,
				Days:       tt.days,
				Output:     &buf,
			}

			err := cmdGenerate(opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("cmdGenerate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check success output
			output := buf.String()
			if !tt.wantErr && !strings.Contains(output, "complete") {
				t.Errorf("cmdGenerate() output should contain 'complete', got: %q", output)
			}
		})
	}
}
