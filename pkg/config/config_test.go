package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefault(t *testing.T) {
	t.Parallel()
	config := Default()

	if config.Planet.Name != "My Planet" {
		t.Errorf("Default name = %q, want %q", config.Planet.Name, "My Planet")
	}

	if config.Planet.Days != 7 {
		t.Errorf("Default days = %d, want 7", config.Planet.Days)
	}

	if config.Planet.ConcurrentFetch != 5 {
		t.Errorf("Default concurrent_fetch = %d, want 5", config.Planet.ConcurrentFetch)
	}

	if config.Database.Path != "./data/planet.db" {
		t.Errorf("Default database path = %q, want %q", config.Database.Path, "./data/planet.db")
	}
}

func TestLoadFromFile(t *testing.T) {
	t.Parallel()
	t.Run("valid config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.ini")

		configContent := `[planet]
name = Test Planet
link = https://example.com
owner_name = John Doe
owner_email = john@example.com
output_dir = ./output
days = 14
log_level = debug
concurrent_fetches = 10
group_by_date = true

[database]
path = ./test.db
`

		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		config, err := LoadFromFile(configPath)
		if err != nil {
			t.Fatalf("LoadFromFile() error = %v", err)
		}

		if config.Planet.Name != "Test Planet" {
			t.Errorf("Name = %q, want %q", config.Planet.Name, "Test Planet")
		}

		if config.Planet.Link != "https://example.com" {
			t.Errorf("Link = %q, want %q", config.Planet.Link, "https://example.com")
		}

		if config.Planet.OwnerName != "John Doe" {
			t.Errorf("OwnerName = %q, want %q", config.Planet.OwnerName, "John Doe")
		}

		if config.Planet.Days != 14 {
			t.Errorf("Days = %d, want 14", config.Planet.Days)
		}

		if config.Planet.ConcurrentFetch != 10 {
			t.Errorf("ConcurrentFetch = %d, want 10", config.Planet.ConcurrentFetch)
		}

		if config.Database.Path != "./test.db" {
			t.Errorf("Database path = %q, want %q", config.Database.Path, "./test.db")
		}
	})

	t.Run("with comments", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.ini")

		configContent := `# This is a comment
[planet]
name = Test Planet
# Another comment
days = 10
; Semicolon comment
link = https://example.com
`

		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		config, err := LoadFromFile(configPath)
		if err != nil {
			t.Fatalf("LoadFromFile() error = %v", err)
		}

		if config.Planet.Name != "Test Planet" {
			t.Errorf("Name = %q, want %q", config.Planet.Name, "Test Planet")
		}
	})

	t.Run("with quoted values", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.ini")

		configContent := `[planet]
name = "Test Planet"
link = 'https://example.com'
`

		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		config, err := LoadFromFile(configPath)
		if err != nil {
			t.Fatalf("LoadFromFile() error = %v", err)
		}

		if config.Planet.Name != "Test Planet" {
			t.Errorf("Name = %q, want %q (quotes should be removed)", config.Planet.Name, "Test Planet")
		}
	})

	t.Run("invalid days value", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.ini")

		configContent := `[planet]
days = not_a_number
`

		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		_, err = LoadFromFile(configPath)
		if err == nil {
			t.Error("Expected error for invalid days value")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := LoadFromFile("/nonexistent/config.ini")
		if err == nil {
			t.Error("Expected error for missing file")
		}
	})
}

func TestLoadFeedsFile(t *testing.T) {
	t.Parallel()
	t.Run("valid feeds file", func(t *testing.T) {
		tmpDir := t.TempDir()
		feedsPath := filepath.Join(tmpDir, "feeds.txt")

		feedsContent := `https://example.com/feed1.xml
https://example.com/feed2.xml
# This is a comment
https://example.com/feed3.xml

https://example.com/feed4.xml
`

		err := os.WriteFile(feedsPath, []byte(feedsContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write feeds file: %v", err)
		}

		feeds, err := LoadFeedsFile(feedsPath)
		if err != nil {
			t.Fatalf("LoadFeedsFile() error = %v", err)
		}

		if len(feeds) != 4 {
			t.Errorf("len(feeds) = %d, want 4", len(feeds))
		}

		expected := []string{
			"https://example.com/feed1.xml",
			"https://example.com/feed2.xml",
			"https://example.com/feed3.xml",
			"https://example.com/feed4.xml",
		}

		for i, feed := range feeds {
			if feed != expected[i] {
				t.Errorf("feeds[%d] = %q, want %q", i, feed, expected[i])
			}
		}
	})

	t.Run("empty lines and comments only", func(t *testing.T) {
		tmpDir := t.TempDir()
		feedsPath := filepath.Join(tmpDir, "feeds.txt")

		feedsContent := `# All comments

# No actual feeds
`

		err := os.WriteFile(feedsPath, []byte(feedsContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write feeds file: %v", err)
		}

		feeds, err := LoadFeedsFile(feedsPath)
		if err != nil {
			t.Fatalf("LoadFeedsFile() error = %v", err)
		}

		if len(feeds) != 0 {
			t.Errorf("len(feeds) = %d, want 0", len(feeds))
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := LoadFeedsFile("/nonexistent/feeds.txt")
		if err == nil {
			t.Error("Expected error for missing file")
		}
	})
}

func TestValidate(t *testing.T) {
	t.Parallel()
	t.Run("valid config", func(t *testing.T) {
		config := Default()
		config.Planet.Name = "Test Planet"

		err := config.Validate()
		if err != nil {
			t.Errorf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		config := Default()
		config.Planet.Name = ""

		err := config.Validate()
		if err == nil {
			t.Error("Expected error for empty name")
		}
	})

	t.Run("invalid days", func(t *testing.T) {
		config := Default()
		config.Planet.Days = 0

		err := config.Validate()
		if err == nil {
			t.Error("Expected error for days = 0")
		}
	})

	t.Run("invalid concurrent_fetches low", func(t *testing.T) {
		config := Default()
		config.Planet.ConcurrentFetch = 0

		err := config.Validate()
		if err == nil {
			t.Error("Expected error for concurrent_fetches = 0")
		}
	})

	t.Run("invalid concurrent_fetches high", func(t *testing.T) {
		config := Default()
		config.Planet.ConcurrentFetch = 100

		err := config.Validate()
		if err == nil {
			t.Error("Expected error for concurrent_fetches = 100")
		}
	})

	t.Run("empty database path", func(t *testing.T) {
		config := Default()
		config.Database.Path = ""

		err := config.Validate()
		if err == nil {
			t.Error("Expected error for empty database path")
		}
	})

	t.Run("database path with parent reference", func(t *testing.T) {
		config := Default()
		config.Database.Path = "../etc/passwd"

		err := config.Validate()
		if err == nil {
			t.Error("Expected error for database path with ..")
		}
		if err != nil && !strings.Contains(err.Error(), "parent directory") {
			t.Errorf("Expected parent directory error, got: %v", err)
		}
	})

	t.Run("output dir with parent reference", func(t *testing.T) {
		config := Default()
		config.Planet.OutputDir = "../../somewhere"

		err := config.Validate()
		if err == nil {
			t.Error("Expected error for output dir with ..")
		}
		if err != nil && !strings.Contains(err.Error(), "parent directory") {
			t.Errorf("Expected parent directory error, got: %v", err)
		}
	})

	t.Run("template path with parent reference", func(t *testing.T) {
		config := Default()
		config.Planet.Template = "../../../etc/passwd"

		err := config.Validate()
		if err == nil {
			t.Error("Expected error for template path with ..")
		}
		if err != nil && !strings.Contains(err.Error(), "parent directory") {
			t.Errorf("Expected parent directory error, got: %v", err)
		}
	})

	t.Run("empty template path allowed", func(t *testing.T) {
		config := Default()
		config.Planet.Template = "" // Should be valid - uses default

		err := config.Validate()
		if err != nil {
			t.Errorf("Validate() error = %v, want nil for empty template", err)
		}
	})
}

func TestSetPlanet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		key       string
		value     string
		wantErr   bool
		checkFunc func(*Config) bool
	}{
		// String fields
		{
			name:  "set name",
			key:   "name",
			value: "New Name",
			checkFunc: func(c *Config) bool {
				return c.Planet.Name == "New Name"
			},
		},
		{
			name:  "set link",
			key:   "link",
			value: "https://planet.example.com",
			checkFunc: func(c *Config) bool {
				return c.Planet.Link == "https://planet.example.com"
			},
		},
		{
			name:  "set owner_name",
			key:   "owner_name",
			value: "Jane Doe",
			checkFunc: func(c *Config) bool {
				return c.Planet.OwnerName == "Jane Doe"
			},
		},
		{
			name:  "set owner_email",
			key:   "owner_email",
			value: "jane@example.com",
			checkFunc: func(c *Config) bool {
				return c.Planet.OwnerEmail == "jane@example.com"
			},
		},
		{
			name:  "set output_dir",
			key:   "output_dir",
			value: "./custom_output",
			checkFunc: func(c *Config) bool {
				return c.Planet.OutputDir == "./custom_output"
			},
		},
		{
			name:  "set log_level",
			key:   "log_level",
			value: "DEBUG",
			checkFunc: func(c *Config) bool {
				return c.Planet.LogLevel == "debug"
			},
		},
		{
			name:  "set user_agent",
			key:   "user_agent",
			value: "CustomBot/1.0",
			checkFunc: func(c *Config) bool {
				return c.Planet.UserAgent == "CustomBot/1.0"
			},
		},
		{
			name:  "set template",
			key:   "template",
			value: "./themes/classic/template.html",
			checkFunc: func(c *Config) bool {
				return c.Planet.Template == "./themes/classic/template.html"
			},
		},
		// Integer fields
		{
			name:  "set days valid",
			key:   "days",
			value: "30",
			checkFunc: func(c *Config) bool {
				return c.Planet.Days == 30
			},
		},
		{
			name:    "set days invalid",
			key:     "days",
			value:   "invalid",
			wantErr: true,
		},
		{
			name:    "set days negative",
			key:     "days",
			value:   "-1",
			wantErr: true,
		},
		{
			name:    "set days zero",
			key:     "days",
			value:   "0",
			wantErr: true,
		},
		{
			name:  "set concurrent_fetches valid",
			key:   "concurrent_fetches",
			value: "10",
			checkFunc: func(c *Config) bool {
				return c.Planet.ConcurrentFetch == 10
			},
		},
		{
			name:  "set concurrent_fetches min",
			key:   "concurrent_fetches",
			value: "1",
			checkFunc: func(c *Config) bool {
				return c.Planet.ConcurrentFetch == 1
			},
		},
		{
			name:  "set concurrent_fetches max",
			key:   "concurrent_fetches",
			value: "50",
			checkFunc: func(c *Config) bool {
				return c.Planet.ConcurrentFetch == 50
			},
		},
		{
			name:    "set concurrent_fetches too low",
			key:     "concurrent_fetches",
			value:   "0",
			wantErr: true,
		},
		{
			name:    "set concurrent_fetches too high",
			key:     "concurrent_fetches",
			value:   "51",
			wantErr: true,
		},
		{
			name:    "set concurrent_fetches invalid",
			key:     "concurrent_fetches",
			value:   "not_a_number",
			wantErr: true,
		},
		// Boolean fields
		{
			name:  "set group_by_date true",
			key:   "group_by_date",
			value: "true",
			checkFunc: func(c *Config) bool {
				return c.Planet.GroupByDate == true
			},
		},
		{
			name:  "set group_by_date false",
			key:   "group_by_date",
			value: "false",
			checkFunc: func(c *Config) bool {
				return c.Planet.GroupByDate == false
			},
		},
		{
			name:  "set group_by_date 1",
			key:   "group_by_date",
			value: "1",
			checkFunc: func(c *Config) bool {
				return c.Planet.GroupByDate == true
			},
		},
		{
			name:  "set group_by_date 0",
			key:   "group_by_date",
			value: "0",
			checkFunc: func(c *Config) bool {
				return c.Planet.GroupByDate == false
			},
		},
		{
			name:    "set group_by_date invalid",
			key:     "group_by_date",
			value:   "maybe",
			wantErr: true,
		},
		// Entry spam prevention fields
		{
			name:  "set filter_by_first_seen true",
			key:   "filter_by_first_seen",
			value: "true",
			checkFunc: func(c *Config) bool {
				return c.Planet.FilterByFirstSeen == true
			},
		},
		{
			name:  "set filter_by_first_seen false",
			key:   "filter_by_first_seen",
			value: "false",
			checkFunc: func(c *Config) bool {
				return c.Planet.FilterByFirstSeen == false
			},
		},
		{
			name:    "set filter_by_first_seen invalid",
			key:     "filter_by_first_seen",
			value:   "maybe",
			wantErr: true,
		},
		{
			name:  "set sort_by published",
			key:   "sort_by",
			value: "published",
			checkFunc: func(c *Config) bool {
				return c.Planet.SortBy == "published"
			},
		},
		{
			name:  "set sort_by first_seen",
			key:   "sort_by",
			value: "first_seen",
			checkFunc: func(c *Config) bool {
				return c.Planet.SortBy == "first_seen"
			},
		},
		{
			name:    "set sort_by invalid",
			key:     "sort_by",
			value:   "foobar",
			wantErr: true,
		},
		// Unknown key
		{
			name:  "unknown key ignored",
			key:   "unknown_setting",
			value: "value",
			checkFunc: func(c *Config) bool {
				return true // Should not error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Default()
			err := config.setPlanet(tt.key, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("setPlanet() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.checkFunc != nil && !tt.checkFunc(config) {
				t.Errorf("setPlanet() did not set value correctly")
			}
		})
	}
}

// TestAllConfigFieldsFromExample tests that all fields in examples/config.ini are parsed correctly
func TestAllConfigFieldsFromExample(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.ini")

	// This matches examples/config.ini with all documented fields
	configContent := `[planet]
name = My Planet
link = https://planet.example.com
owner_name = Your Name
owner_email = you@example.com
output_dir = ./public
days = 7
log_level = info
concurrent_fetches = 5
group_by_date = true
user_agent = MyPlanet/1.0
template = ./themes/classic/template.html
filter_by_first_seen = false
sort_by = published

[database]
path = ./data/planet.db
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	// Verify all planet fields
	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"name", config.Planet.Name, "My Planet"},
		{"link", config.Planet.Link, "https://planet.example.com"},
		{"owner_name", config.Planet.OwnerName, "Your Name"},
		{"owner_email", config.Planet.OwnerEmail, "you@example.com"},
		{"output_dir", config.Planet.OutputDir, "./public"},
		{"days", config.Planet.Days, 7},
		{"log_level", config.Planet.LogLevel, "info"},
		{"concurrent_fetches", config.Planet.ConcurrentFetch, 5},
		{"group_by_date", config.Planet.GroupByDate, true},
		{"user_agent", config.Planet.UserAgent, "MyPlanet/1.0"},
		{"template", config.Planet.Template, "./themes/classic/template.html"},
		{"filter_by_first_seen", config.Planet.FilterByFirstSeen, false},
		{"sort_by", config.Planet.SortBy, "published"},
		{"database.path", config.Database.Path, "./data/planet.db"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

// TestSetDatabase tests database configuration parsing
func TestSetDatabase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		key       string
		value     string
		wantErr   bool
		checkFunc func(*Config) bool
	}{
		{
			name:  "set path",
			key:   "path",
			value: "./custom.db",
			checkFunc: func(c *Config) bool {
				return c.Database.Path == "./custom.db"
			},
		},
		{
			name:  "unknown database key ignored",
			key:   "unknown_db_option",
			value: "value",
			checkFunc: func(c *Config) bool {
				return true // Should not error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Default()
			err := config.setDatabase(tt.key, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("setDatabase() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.checkFunc != nil && !tt.checkFunc(config) {
				t.Errorf("setDatabase() did not set value correctly")
			}
		})
	}
}

// TestConnectionPoolingConfig tests HTTP connection pooling and retry configuration
func TestConnectionPoolingConfig(t *testing.T) {
	t.Parallel()
	t.Run("defaults", func(t *testing.T) {
		config := Default()

		if config.Planet.MaxRetries != 3 {
			t.Errorf("Default MaxRetries = %d, want 3", config.Planet.MaxRetries)
		}
		if config.Planet.MaxIdleConns != 100 {
			t.Errorf("Default MaxIdleConns = %d, want 100", config.Planet.MaxIdleConns)
		}
		if config.Planet.MaxIdleConnsPerHost != 10 {
			t.Errorf("Default MaxIdleConnsPerHost = %d, want 10", config.Planet.MaxIdleConnsPerHost)
		}
		if config.Planet.MaxConnsPerHost != 20 {
			t.Errorf("Default MaxConnsPerHost = %d, want 20", config.Planet.MaxConnsPerHost)
		}
		if config.Planet.IdleConnTimeoutSeconds != 90 {
			t.Errorf("Default IdleConnTimeoutSeconds = %d, want 90", config.Planet.IdleConnTimeoutSeconds)
		}
	})

	t.Run("custom values", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.ini")

		configContent := `[planet]
name = Test Planet
link = https://example.com
max_retries = 5
max_idle_conns = 200
max_idle_conns_per_host = 20
max_conns_per_host = 50
idle_conn_timeout_seconds = 120
`

		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		config, err := LoadFromFile(configPath)
		if err != nil {
			t.Fatalf("LoadFromFile() error = %v", err)
		}

		if config.Planet.MaxRetries != 5 {
			t.Errorf("MaxRetries = %d, want 5", config.Planet.MaxRetries)
		}
		if config.Planet.MaxIdleConns != 200 {
			t.Errorf("MaxIdleConns = %d, want 200", config.Planet.MaxIdleConns)
		}
		if config.Planet.MaxIdleConnsPerHost != 20 {
			t.Errorf("MaxIdleConnsPerHost = %d, want 20", config.Planet.MaxIdleConnsPerHost)
		}
		if config.Planet.MaxConnsPerHost != 50 {
			t.Errorf("MaxConnsPerHost = %d, want 50", config.Planet.MaxConnsPerHost)
		}
		if config.Planet.IdleConnTimeoutSeconds != 120 {
			t.Errorf("IdleConnTimeoutSeconds = %d, want 120", config.Planet.IdleConnTimeoutSeconds)
		}
	})

	t.Run("invalid max_retries", func(t *testing.T) {
		config := Default()
		err := config.setPlanet("max_retries", "11")
		if err == nil {
			t.Error("Expected error for max_retries > 10")
		}
		err = config.setPlanet("max_retries", "-1")
		if err == nil {
			t.Error("Expected error for max_retries < 0")
		}
	})

	t.Run("invalid max_idle_conns", func(t *testing.T) {
		config := Default()
		err := config.setPlanet("max_idle_conns", "5")
		if err == nil {
			t.Error("Expected error for max_idle_conns < 10")
		}
		err = config.setPlanet("max_idle_conns", "1001")
		if err == nil {
			t.Error("Expected error for max_idle_conns > 1000")
		}
	})

	t.Run("invalid max_idle_conns_per_host", func(t *testing.T) {
		config := Default()
		err := config.setPlanet("max_idle_conns_per_host", "0")
		if err == nil {
			t.Error("Expected error for max_idle_conns_per_host < 1")
		}
		err = config.setPlanet("max_idle_conns_per_host", "101")
		if err == nil {
			t.Error("Expected error for max_idle_conns_per_host > 100")
		}
	})

	t.Run("invalid max_conns_per_host", func(t *testing.T) {
		config := Default()
		err := config.setPlanet("max_conns_per_host", "0")
		if err == nil {
			t.Error("Expected error for max_conns_per_host < 1")
		}
		err = config.setPlanet("max_conns_per_host", "201")
		if err == nil {
			t.Error("Expected error for max_conns_per_host > 200")
		}
	})

	t.Run("invalid idle_conn_timeout_seconds", func(t *testing.T) {
		config := Default()
		err := config.setPlanet("idle_conn_timeout_seconds", "5")
		if err == nil {
			t.Error("Expected error for idle_conn_timeout_seconds < 10")
		}
		err = config.setPlanet("idle_conn_timeout_seconds", "601")
		if err == nil {
			t.Error("Expected error for idle_conn_timeout_seconds > 600")
		}
	})
}
