// Package config provides configuration file parsing for the feed aggregator.
//
// The config package reads INI-format configuration files and provides
// validated configuration values with sensible defaults. It supports
// forward compatibility by ignoring unknown sections and keys.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Configuration validation constants define acceptable ranges for config values.
// These limits ensure resource safety and prevent misconfiguration.
const (
	// Concurrency limits
	MinConcurrentFetches = 1
	MaxConcurrentFetches = 50 // Prevents resource exhaustion

	// Retry limits
	MinMaxRetries = 0
	MaxMaxRetries = 10 // Reasonable retry limit

	// HTTP connection pool limits
	MinMaxIdleConns        = 10   // Minimum for connection reuse
	MaxMaxIdleConns        = 1000 // Prevents memory bloat
	MinMaxIdleConnsPerHost = 1
	MaxMaxIdleConnsPerHost = 100
	MinMaxConnsPerHost     = 1
	MaxMaxConnsPerHost     = 200

	// Timeout limits (seconds)
	MinIdleConnTimeout       = 10  // 10 seconds
	MaxIdleConnTimeout       = 600 // 10 minutes
	MinHTTPTimeout           = 5   // 5 seconds
	MaxHTTPTimeout           = 300 // 5 minutes
	MinDialTimeout           = 1   // 1 second
	MaxDialTimeout           = 60  // 1 minute
	MinTLSHandshakeTimeout   = 1   // 1 second
	MaxTLSHandshakeTimeout   = 60  // 1 minute
	MinResponseHeaderTimeout = 1   // 1 second
	MaxResponseHeaderTimeout = 60  // 1 minute

	// Rate limiting
	MinRequestsPerMinute = 1
	MaxRequestsPerMinute = 600 // 10 requests/second max
	MinRateLimitBurst    = 1
	MaxRateLimitBurst    = 50

	// Content limits
	MinDays = 1 // At least 1 day of content
)

// Config represents the application configuration
type Config struct {
	Planet   PlanetConfig
	Database DatabaseConfig
	Feeds    []string
}

// PlanetConfig contains planet-level settings
type PlanetConfig struct {
	Name              string
	Link              string
	OwnerName         string
	OwnerEmail        string
	OutputDir         string
	Days              int
	LogLevel          string
	ConcurrentFetch   int
	UserAgent         string
	GroupByDate       bool
	Template          string
	FilterByFirstSeen bool
	SortBy            string

	// HTTP connection pooling and retry settings
	MaxRetries             int // Number of retry attempts for failed requests (default: 3)
	MaxIdleConns           int // Total idle connections across all hosts (default: 100)
	MaxIdleConnsPerHost    int // Idle connections per host (default: 10)
	MaxConnsPerHost        int // Maximum active connections per host (default: 20)
	IdleConnTimeoutSeconds int // Idle connection timeout in seconds (default: 90)

	// HTTP timeout settings
	HTTPTimeoutSeconds           int // Overall HTTP request timeout (default: 30)
	DialTimeoutSeconds           int // TCP connection timeout (default: 10)
	TLSHandshakeTimeoutSeconds   int // TLS handshake timeout (default: 10)
	ResponseHeaderTimeoutSeconds int // Response header timeout (default: 10)

	// Rate limiting settings (per domain)
	RequestsPerMinute int // Maximum requests per domain per minute (default: 60)
	RateLimitBurst    int // Burst size for rate limiter (default: 10)
}

// DatabaseConfig contains database settings
type DatabaseConfig struct {
	Path string
}

// Default returns a configuration with default values
func Default() *Config {
	return &Config{
		Planet: PlanetConfig{
			Name:              "My Planet",
			Link:              "",
			OwnerName:         "",
			OwnerEmail:        "",
			OutputDir:         "./public",
			Days:              7,
			LogLevel:          "info",
			ConcurrentFetch:   5,
			UserAgent:         "RoguePlanet/0.1",
			GroupByDate:       true,
			FilterByFirstSeen: false,
			SortBy:            "published",

			// HTTP connection pooling and retry defaults
			MaxRetries:             3,
			MaxIdleConns:           100,
			MaxIdleConnsPerHost:    10,
			MaxConnsPerHost:        20,
			IdleConnTimeoutSeconds: 90,

			// HTTP timeout defaults
			HTTPTimeoutSeconds:           30,
			DialTimeoutSeconds:           10,
			TLSHandshakeTimeoutSeconds:   10,
			ResponseHeaderTimeoutSeconds: 10,

			// Rate limiting defaults
			RequestsPerMinute: 60,
			RateLimitBurst:    10,
		},
		Database: DatabaseConfig{
			Path: "./data/planet.db",
		},
		Feeds: []string{},
	}
}

// LoadFromFile loads configuration from an INI file
func LoadFromFile(path string) (config *Config, err error) {
	file, openErr := os.Open(path)
	if openErr != nil {
		return nil, fmt.Errorf("open config file: %w", openErr)
	}
	// Close file when done. Close errors during read are rarely actionable.
	defer file.Close()

	config = Default()
	scanner := bufio.NewScanner(file)
	currentSection := ""

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Check for section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}

		// Parse key-value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line %d: %s", lineNum, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		// Apply configuration based on section
		if err := config.set(currentSection, key, value); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	return config, nil
}

// set applies a configuration value
func (c *Config) set(section, key, value string) error {
	switch section {
	case "planet", "":
		return c.setPlanet(key, value)
	case "database":
		return c.setDatabase(key, value)
	default:
		// Unknown sections are ignored for forward compatibility
		return nil
	}
}

// setPlanet sets planet configuration values
func (c *Config) setPlanet(key, value string) error {
	switch key {
	case "name":
		c.Planet.Name = value
	case "link":
		c.Planet.Link = value
	case "owner_name":
		c.Planet.OwnerName = value
	case "owner_email":
		c.Planet.OwnerEmail = value
	case "output_dir":
		c.Planet.OutputDir = value
	case "days":
		days, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid days value: %s", value)
		}
		if days < MinDays {
			return fmt.Errorf("days must be >= %d", MinDays)
		}
		c.Planet.Days = days
	case "log_level":
		c.Planet.LogLevel = strings.ToLower(value)
	case "concurrent_fetches":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid concurrent_fetches value: %s", value)
		}
		if n < MinConcurrentFetches || n > MaxConcurrentFetches {
			return fmt.Errorf("concurrent_fetches must be between %d and %d", MinConcurrentFetches, MaxConcurrentFetches)
		}
		c.Planet.ConcurrentFetch = n
	case "user_agent":
		c.Planet.UserAgent = value
	case "group_by_date":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid group_by_date value: %s", value)
		}
		c.Planet.GroupByDate = b
	case "template":
		c.Planet.Template = value
	case "filter_by_first_seen":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid filter_by_first_seen value: %s", value)
		}
		c.Planet.FilterByFirstSeen = b
	case "sort_by":
		if value != "" && value != "published" && value != "first_seen" {
			return fmt.Errorf("sort_by must be 'published' or 'first_seen', got: %s", value)
		}
		c.Planet.SortBy = value
	case "max_retries":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid max_retries value: %s", value)
		}
		if n < MinMaxRetries || n > MaxMaxRetries {
			return fmt.Errorf("max_retries must be between %d and %d", MinMaxRetries, MaxMaxRetries)
		}
		c.Planet.MaxRetries = n
	case "max_idle_conns":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid max_idle_conns value: %s", value)
		}
		if n < MinMaxIdleConns || n > MaxMaxIdleConns {
			return fmt.Errorf("max_idle_conns must be between %d and %d", MinMaxIdleConns, MaxMaxIdleConns)
		}
		c.Planet.MaxIdleConns = n
	case "max_idle_conns_per_host":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid max_idle_conns_per_host value: %s", value)
		}
		if n < MinMaxIdleConnsPerHost || n > MaxMaxIdleConnsPerHost {
			return fmt.Errorf("max_idle_conns_per_host must be between %d and %d", MinMaxIdleConnsPerHost, MaxMaxIdleConnsPerHost)
		}
		c.Planet.MaxIdleConnsPerHost = n
	case "max_conns_per_host":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid max_conns_per_host value: %s", value)
		}
		if n < MinMaxConnsPerHost || n > MaxMaxConnsPerHost {
			return fmt.Errorf("max_conns_per_host must be between %d and %d", MinMaxConnsPerHost, MaxMaxConnsPerHost)
		}
		c.Planet.MaxConnsPerHost = n
	case "idle_conn_timeout_seconds":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid idle_conn_timeout_seconds value: %s", value)
		}
		if n < MinIdleConnTimeout || n > MaxIdleConnTimeout {
			return fmt.Errorf("idle_conn_timeout_seconds must be between %d and %d", MinIdleConnTimeout, MaxIdleConnTimeout)
		}
		c.Planet.IdleConnTimeoutSeconds = n
	case "http_timeout_seconds":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid http_timeout_seconds value: %s", value)
		}
		if n < MinHTTPTimeout || n > MaxHTTPTimeout {
			return fmt.Errorf("http_timeout_seconds must be between %d and %d", MinHTTPTimeout, MaxHTTPTimeout)
		}
		c.Planet.HTTPTimeoutSeconds = n
	case "dial_timeout_seconds":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid dial_timeout_seconds value: %s", value)
		}
		if n < MinDialTimeout || n > MaxDialTimeout {
			return fmt.Errorf("dial_timeout_seconds must be between %d and %d", MinDialTimeout, MaxDialTimeout)
		}
		c.Planet.DialTimeoutSeconds = n
	case "tls_handshake_timeout_seconds":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid tls_handshake_timeout_seconds value: %s", value)
		}
		if n < MinTLSHandshakeTimeout || n > MaxTLSHandshakeTimeout {
			return fmt.Errorf("tls_handshake_timeout_seconds must be between %d and %d", MinTLSHandshakeTimeout, MaxTLSHandshakeTimeout)
		}
		c.Planet.TLSHandshakeTimeoutSeconds = n
	case "response_header_timeout_seconds":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid response_header_timeout_seconds value: %s", value)
		}
		if n < MinResponseHeaderTimeout || n > MaxResponseHeaderTimeout {
			return fmt.Errorf("response_header_timeout_seconds must be between %d and %d", MinResponseHeaderTimeout, MaxResponseHeaderTimeout)
		}
		c.Planet.ResponseHeaderTimeoutSeconds = n
	case "requests_per_minute":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid requests_per_minute value: %s", value)
		}
		if n < MinRequestsPerMinute || n > MaxRequestsPerMinute {
			return fmt.Errorf("requests_per_minute must be between %d and %d", MinRequestsPerMinute, MaxRequestsPerMinute)
		}
		c.Planet.RequestsPerMinute = n
	case "rate_limit_burst":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid rate_limit_burst value: %s", value)
		}
		if n < MinRateLimitBurst || n > MaxRateLimitBurst {
			return fmt.Errorf("rate_limit_burst must be between %d and %d", MinRateLimitBurst, MaxRateLimitBurst)
		}
		c.Planet.RateLimitBurst = n
	default:
		// Unknown keys are ignored for forward compatibility
		return nil
	}
	return nil
}

// setDatabase sets database configuration values
func (c *Config) setDatabase(key, value string) error {
	switch key {
	case "path":
		c.Database.Path = value
	default:
		// Unknown keys are ignored
		return nil
	}
	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Planet.Name == "" {
		return fmt.Errorf("planet name is required")
	}

	if c.Planet.Days < 1 {
		return fmt.Errorf("days must be >= 1")
	}

	if c.Planet.ConcurrentFetch < 1 || c.Planet.ConcurrentFetch > 50 {
		return fmt.Errorf("concurrent_fetches must be between 1 and 50")
	}

	if c.Database.Path == "" {
		return fmt.Errorf("database path is required")
	}

	// Path validation - prevent path traversal attacks
	// Reject parent directory references in paths (the main security concern)
	if strings.Contains(c.Database.Path, "..") {
		return fmt.Errorf("database path must not contain parent directory references (..): %s", c.Database.Path)
	}
	if strings.Contains(c.Planet.OutputDir, "..") {
		return fmt.Errorf("output directory must not contain parent directory references (..): %s", c.Planet.OutputDir)
	}
	// Validate template path if specified (empty is allowed - uses default template)
	if c.Planet.Template != "" && strings.Contains(c.Planet.Template, "..") {
		return fmt.Errorf("template path must not contain parent directory references (..): %s", c.Planet.Template)
	}

	// Set default and validate sort_by
	if c.Planet.SortBy == "" {
		c.Planet.SortBy = "published"
	}
	if c.Planet.SortBy != "published" && c.Planet.SortBy != "first_seen" {
		return fmt.Errorf("sort_by must be 'published' or 'first_seen', got: %s", c.Planet.SortBy)
	}

	return nil
}

// LoadFeedsFile loads feed URLs from a text file.
// Each line should contain a single URL. Lines starting with '#' are comments.
// Empty lines are ignored.
//
// Returns a slice of feed URLs or an error if the file cannot be read.
func LoadFeedsFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read feeds file: %w", err)
	}

	var urls []string
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		urls = append(urls, line)
	}

	return urls, nil
}
