package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adewale/rogue_planet/pkg/logging"
	"github.com/adewale/rogue_planet/pkg/repository"
)

func TestCmdAddFeed(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	t.Run("missing URL", func(t *testing.T) {
		var buf bytes.Buffer
		opts := RemoveFeedOptions{
			URL:        "",
			ConfigPath: "./config.ini",
			Output:     &buf,
			Force:      true,
		}
		err := cmdRemoveFeed(opts)
		if err == nil {
			t.Error("cmdRemoveFeed() expected error for missing URL, got nil")
		}
	})

	t.Run("remove feed with --force", func(t *testing.T) {
		// Setup test environment
		tmpDir := t.TempDir()

		// Create database directory
		dataDir := filepath.Join(tmpDir, "data")
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			t.Fatalf("Failed to create data directory: %v", err)
		}

		// Create database and add feed
		dbPath := filepath.Join(dataDir, "planet.db")

		// Create config file with database path
		configPath := filepath.Join(tmpDir, "config.ini")
		configContent := fmt.Sprintf(`[planet]
name = Test Planet

[database]
path = %s
`, dbPath)
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}
		repo, err := repository.New(dbPath)
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}

		ctx := context.Background()

		feedURL := "https://example.com/feed"
		feedID, err := repo.AddFeed(ctx, feedURL, "Test Feed")
		if err != nil {
			t.Fatalf("Failed to add feed: %v", err)
		}

		// Add some entries
		now := time.Now()
		for i := 0; i < 3; i++ {
			entry := &repository.Entry{
				FeedID:      feedID,
				EntryID:     fmt.Sprintf("entry%d", i),
				Title:       fmt.Sprintf("Entry %d", i),
				Link:        fmt.Sprintf("https://example.com/entry%d", i),
				Published:   now,
				Updated:     now,
				FirstSeen:   now,
				Content:     "Test content",
				ContentType: "html",
			}
			if err := repo.UpsertEntry(ctx, entry); err != nil {
				t.Fatalf("Failed to add entry: %v", err)
			}
		}
		repo.Close()

		// Remove feed with --force flag
		var buf bytes.Buffer
		opts := RemoveFeedOptions{
			URL:        feedURL,
			ConfigPath: configPath,
			Output:     &buf,
			Force:      true,
		}

		if err := cmdRemoveFeed(opts); err != nil {
			t.Fatalf("cmdRemoveFeed() error = %v", err)
		}

		// Check output includes entry count
		output := buf.String()
		if !strings.Contains(output, "3 entries deleted") {
			t.Errorf("Output should mention 3 entries deleted, got: %s", output)
		}
		if !strings.Contains(output, feedURL) {
			t.Errorf("Output should mention feed URL, got: %s", output)
		}

		// Verify feed was removed from database
		repo, err = repository.New(dbPath)
		if err != nil {
			t.Fatalf("Failed to reopen repository: %v", err)
		}
		defer repo.Close()

		_, err = repo.GetFeedByURL(ctx, feedURL)
		if err != repository.ErrFeedNotFound {
			t.Errorf("Feed should be removed, got error: %v", err)
		}

		// Verify entries were cascade deleted
		count, err := repo.GetEntryCountForFeed(ctx, feedID)
		if err != nil {
			t.Fatalf("GetEntryCountForFeed() error = %v", err)
		}
		if count != 0 {
			t.Errorf("Entries should be cascade deleted, got count: %d", count)
		}
	})

	t.Run("feed not found", func(t *testing.T) {
		tmpDir := t.TempDir()

		dataDir := filepath.Join(tmpDir, "data")
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			t.Fatalf("Failed to create data directory: %v", err)
		}

		dbPath := filepath.Join(dataDir, "planet.db")

		configPath := filepath.Join(tmpDir, "config.ini")
		configContent := fmt.Sprintf(`[planet]
name = Test Planet

[database]
path = %s
`, dbPath)
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}
		repo, err := repository.New(dbPath)
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}
		repo.Close()

		var buf bytes.Buffer
		opts := RemoveFeedOptions{
			URL:        "https://nonexistent.com/feed",
			ConfigPath: configPath,
			Output:     &buf,
			Force:      true,
		}

		err = cmdRemoveFeed(opts)
		if err == nil {
			t.Error("cmdRemoveFeed() expected error for non-existent feed, got nil")
		}
		if !strings.Contains(err.Error(), "feed not found") {
			t.Errorf("Error should mention 'feed not found', got: %v", err)
		}
	})

	// Test #5: Interactive Confirmation Test
	t.Run("interactive confirmation - accept with y", func(t *testing.T) {
		tmpDir := t.TempDir()
		dataDir := filepath.Join(tmpDir, "data")
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			t.Fatalf("Failed to create data directory: %v", err)
		}

		dbPath := filepath.Join(dataDir, "planet.db")
		configPath := filepath.Join(tmpDir, "config.ini")
		configContent := fmt.Sprintf(`[planet]
name = Test Planet

[database]
path = %s
`, dbPath)
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		repo, err := repository.New(dbPath)
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}

		ctx := context.Background()

		feedURL := "https://example.com/feed"
		feedID, err := repo.AddFeed(ctx, feedURL, "Test Feed")
		if err != nil {
			t.Fatalf("Failed to add feed: %v", err)
		}

		// Add an entry
		now := time.Now()
		entry := &repository.Entry{
			FeedID:      feedID,
			EntryID:     "entry1",
			Title:       "Entry 1",
			Link:        "https://example.com/entry1",
			Published:   now,
			Updated:     now,
			FirstSeen:   now,
			Content:     "Test content",
			ContentType: "html",
		}
		if err := repo.UpsertEntry(ctx, entry); err != nil {
			t.Fatalf("Failed to add entry: %v", err)
		}
		repo.Close()

		// Mock stdin with "y\n"
		input := strings.NewReader("y\n")
		var buf bytes.Buffer

		opts := RemoveFeedOptions{
			URL:        feedURL,
			ConfigPath: configPath,
			Output:     &buf,
			Input:      input,
			Force:      false,
		}

		if err := cmdRemoveFeed(opts); err != nil {
			t.Fatalf("cmdRemoveFeed() should succeed with 'y' input, got error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Feed: https://example.com/feed") {
			t.Errorf("Output should show feed URL, got: %s", output)
		}
		if !strings.Contains(output, "Remove this feed and all 1 entries? (y/N):") {
			t.Errorf("Output should show confirmation prompt, got: %s", output)
		}
		if !strings.Contains(output, "✓ Removed feed") {
			t.Errorf("Output should show success message, got: %s", output)
		}

		// Verify feed was removed
		repo, err = repository.New(dbPath)
		if err != nil {
			t.Fatalf("Failed to reopen repository: %v", err)
		}
		defer repo.Close()

		_, err = repo.GetFeedByURL(ctx, feedURL)
		if err != repository.ErrFeedNotFound {
			t.Errorf("Feed should be removed, got error: %v", err)
		}
	})

	t.Run("interactive confirmation - cancel with n", func(t *testing.T) {
		tmpDir := t.TempDir()
		dataDir := filepath.Join(tmpDir, "data")
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			t.Fatalf("Failed to create data directory: %v", err)
		}

		dbPath := filepath.Join(dataDir, "planet.db")
		configPath := filepath.Join(tmpDir, "config.ini")
		configContent := fmt.Sprintf(`[planet]
name = Test Planet

[database]
path = %s
`, dbPath)
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		repo, err := repository.New(dbPath)
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}

		ctx := context.Background()

		feedURL := "https://example.com/feed"
		_, err = repo.AddFeed(ctx, feedURL, "Test Feed")
		if err != nil {
			t.Fatalf("Failed to add feed: %v", err)
		}
		repo.Close()

		// Mock stdin with "n\n"
		input := strings.NewReader("n\n")
		var buf bytes.Buffer

		opts := RemoveFeedOptions{
			URL:        feedURL,
			ConfigPath: configPath,
			Output:     &buf,
			Input:      input,
			Force:      false,
		}

		err = cmdRemoveFeed(opts)
		if err == nil {
			t.Error("cmdRemoveFeed() should return error when user cancels")
		}

		// Check that error is ErrUserCancelled
		if _, ok := err.(*ErrUserCancelled); !ok {
			t.Errorf("Error should be *ErrUserCancelled, got: %T", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Cancelled.") {
			t.Errorf("Output should show 'Cancelled.', got: %s", output)
		}

		// Verify feed was NOT removed
		repo, err = repository.New(dbPath)
		if err != nil {
			t.Fatalf("Failed to reopen repository: %v", err)
		}
		defer repo.Close()

		feed, err := repo.GetFeedByURL(ctx, feedURL)
		if err != nil {
			t.Errorf("Feed should still exist after cancellation, got error: %v", err)
		}
		if feed == nil {
			t.Error("Feed should still exist after cancellation")
		}
	})

	t.Run("interactive confirmation - accept with yes", func(t *testing.T) {
		tmpDir := t.TempDir()
		dataDir := filepath.Join(tmpDir, "data")
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			t.Fatalf("Failed to create data directory: %v", err)
		}

		dbPath := filepath.Join(dataDir, "planet.db")
		configPath := filepath.Join(tmpDir, "config.ini")
		configContent := fmt.Sprintf(`[planet]
name = Test Planet

[database]
path = %s
`, dbPath)
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		repo, err := repository.New(dbPath)
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}

		ctx := context.Background()

		feedURL := "https://example.com/feed"
		_, err = repo.AddFeed(ctx, feedURL, "Test Feed")
		if err != nil {
			t.Fatalf("Failed to add feed: %v", err)
		}
		repo.Close()

		// Mock stdin with "yes\n"
		input := strings.NewReader("yes\n")
		var buf bytes.Buffer

		opts := RemoveFeedOptions{
			URL:        feedURL,
			ConfigPath: configPath,
			Output:     &buf,
			Input:      input,
			Force:      false,
		}

		if err := cmdRemoveFeed(opts); err != nil {
			t.Fatalf("cmdRemoveFeed() should succeed with 'yes' input, got error: %v", err)
		}
	})

	// Test #7: Non-Interactive Error Test
	t.Run("non-interactive without --force", func(t *testing.T) {
		tmpDir := t.TempDir()
		dataDir := filepath.Join(tmpDir, "data")
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			t.Fatalf("Failed to create data directory: %v", err)
		}

		dbPath := filepath.Join(dataDir, "planet.db")
		configPath := filepath.Join(tmpDir, "config.ini")
		configContent := fmt.Sprintf(`[planet]
name = Test Planet

[database]
path = %s
`, dbPath)
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		repo, err := repository.New(dbPath)
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}

		ctx := context.Background()

		feedURL := "https://example.com/feed"
		_, err = repo.AddFeed(ctx, feedURL, "Test Feed")
		if err != nil {
			t.Fatalf("Failed to add feed: %v", err)
		}
		repo.Close()

		// Create a pipe to simulate piped input (non-terminal os.File)
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("Failed to create pipe: %v", err)
		}
		defer r.Close()

		// Write input to pipe
		go func() {
			if _, err := w.Write([]byte("y\n")); err != nil {
				// If write fails, close and return - test will fail when it doesn't get input
				w.Close()
				return
			}
			w.Close()
		}()

		var buf bytes.Buffer

		opts := RemoveFeedOptions{
			URL:        feedURL,
			ConfigPath: configPath,
			Output:     &buf,
			Input:      r, // This is an os.File but not a terminal (it's a pipe)
			Force:      false,
		}

		err = cmdRemoveFeed(opts)
		if err == nil {
			t.Error("cmdRemoveFeed() should return error in non-interactive mode without --force")
		}

		if !strings.Contains(err.Error(), "cannot prompt for confirmation in non-interactive mode") {
			t.Errorf("Error should mention non-interactive mode, got: %v", err)
		}
		if !strings.Contains(err.Error(), "Use --force to skip confirmation") {
			t.Errorf("Error should mention --force flag, got: %v", err)
		}
	})
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
	t.Parallel()
	var buf bytes.Buffer
	opts := PruneOptions{
		ConfigPath: "/nonexistent/config.ini",
		Days:       90,
		DryRun:     true,
		Output:     &buf,
	}

	// With dry-run, should print message and return without error
	err := cmdPrune(context.Background(), opts)
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
	t.Parallel()
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
	t.Parallel()
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
				Logger:     logging.New("info"),
			}

			err := cmdUpdate(context.Background(), opts)
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
	t.Parallel()
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
				Logger:     logging.New("info"),
			}

			err := cmdFetch(context.Background(), opts)
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
	t.Parallel()
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

			err := cmdGenerate(context.Background(), opts)
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

// TestSignalHandlingGracefulShutdown verifies that closing the signal channel
// during normal shutdown does not produce spurious "Received signal <nil>" log messages
func TestSignalHandlingGracefulShutdown(t *testing.T) {
	// Set up temporary directory and config
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
log_level = info

[database]
path = ` + dbPath + `
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create database and add a feed (to exercise signal handling code)
	repo, err := repository.New(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Add a feed (will fail to fetch, but that's okay - we just need to test signal handling)
	_, err = repo.AddFeed(ctx, "https://example.com/feed.xml", "Example Feed")
	if err != nil {
		repo.Close()
		t.Fatal(err)
	}
	repo.Close()

	// Capture log output
	var logBuf bytes.Buffer
	originalLogOutput := log.Writer()
	originalLogFlags := log.Flags()
	log.SetOutput(&logBuf)
	log.SetFlags(0) // Remove timestamp for easier testing
	defer func() {
		log.SetOutput(originalLogOutput)
		log.SetFlags(originalLogFlags)
	}()

	// Run cmdFetch (which calls fetchFeeds with signal handling)
	var outputBuf bytes.Buffer
	opts := FetchOptions{
		ConfigPath: configPath,
		Verbose:    false,
		Output:     &outputBuf,
		Logger:     logging.New("info"),
	}

	err = cmdFetch(context.Background(), opts)
	if err != nil {
		t.Fatalf("cmdFetch() error = %v", err)
	}

	// Check that log output does NOT contain the spurious signal message
	logOutput := logBuf.String()
	if strings.Contains(logOutput, "Received signal <nil>") {
		t.Errorf("Log output contains spurious 'Received signal <nil>' message during normal shutdown.\nLog output:\n%s", logOutput)
	}

	// Verify expected log messages are present
	if !strings.Contains(logOutput, "Fetching 1 feeds") {
		t.Errorf("Log output should contain 'Fetching 1 feeds', got:\n%s", logOutput)
	}
	if !strings.Contains(logOutput, "Completed fetching all feeds") {
		t.Errorf("Log output should contain 'Completed fetching all feeds', got:\n%s", logOutput)
	}
}
