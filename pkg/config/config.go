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
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close config file: %w", closeErr)
		}
	}()

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
		if days < 1 {
			return fmt.Errorf("days must be >= 1")
		}
		c.Planet.Days = days
	case "log_level":
		c.Planet.LogLevel = strings.ToLower(value)
	case "concurrent_fetches":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid concurrent_fetches value: %s", value)
		}
		if n < 1 || n > 50 {
			return fmt.Errorf("concurrent_fetches must be between 1 and 50")
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

	// Set default and validate sort_by
	if c.Planet.SortBy == "" {
		c.Planet.SortBy = "published"
	}
	if c.Planet.SortBy != "published" && c.Planet.SortBy != "first_seen" {
		return fmt.Errorf("sort_by must be 'published' or 'first_seen', got: %s", c.Planet.SortBy)
	}

	return nil
}

// LoadFeedsFile loads feed URLs from a text file
// One URL per line, lines starting with # are treated as comments
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
