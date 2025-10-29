// Package crawler provides HTTP fetching for RSS/Atom feeds with conditional request support.
//
// The crawler implements proper HTTP caching using ETag and Last-Modified headers,
// SSRF prevention, and response size limiting. It is designed to be a well-behaved
// feed fetcher that minimizes bandwidth and server load.
package crawler

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// MaxFeedSize limits response body size to 10MB
	MaxFeedSize = 10 * 1024 * 1024
	// DefaultTimeout for HTTP requests
	DefaultTimeout = 30 * time.Second
	// MaxRedirects prevents redirect loops
	MaxRedirects = 5
	// UserAgent identifies the bot
	UserAgent = "RoguePlanet/0.3 (+https://github.com/adewale/rogue_planet)"
)

var (
	ErrInvalidURL      = errors.New("invalid URL")
	ErrPrivateIP       = errors.New("private or internal IP not allowed")
	ErrInvalidScheme   = errors.New("only http and https schemes allowed")
	ErrMaxSizeExceeded = errors.New("response body exceeds maximum size")
)

// FeedCache stores HTTP caching headers for conditional requests
type FeedCache struct {
	URL          string
	ETag         string // Stored exactly as received, including quotes
	LastModified string // Stored exactly as received
	LastFetched  time.Time
}

// FeedResponse contains the fetched feed data and metadata
type FeedResponse struct {
	Body        []byte
	StatusCode  int
	NotModified bool      // True if 304 Not Modified was returned
	NewCache    FeedCache // Updated cache headers for storage
	FinalURL    string    // URL after redirects (for 301 permanent redirects)
	FetchTime   time.Time
}

// Crawler handles HTTP fetching with proper conditional request support
type Crawler struct {
	client        *http.Client
	userAgent     string
	maxSize       int64
	skipSSRFCheck bool // For testing only - allows local URLs
}

// New creates a new Crawler with default settings
func New() *Crawler {
	// Configure HTTP transport with connection pooling
	transport := &http.Transport{
		// Connection pooling settings
		MaxIdleConns:        100,              // Total idle connections across all hosts
		MaxIdleConnsPerHost: 10,               // Idle connections per host (key for feed fetching)
		MaxConnsPerHost:     20,               // Maximum active connections per host
		IdleConnTimeout:     90 * time.Second, // Keep idle connections for reuse

		// Timeouts for connection establishment
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second, // TCP connection timeout
			KeepAlive: 30 * time.Second, // TCP keep-alive
		}).DialContext,

		// TLS handshake timeout
		TLSHandshakeTimeout: 10 * time.Second,

		// Response header timeout
		ResponseHeaderTimeout: 10 * time.Second,

		// Expect Continue timeout
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &Crawler{
		client: &http.Client{
			Transport: transport,
			Timeout:   DefaultTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= MaxRedirects {
					return fmt.Errorf("stopped after %d redirects", MaxRedirects)
				}
				return nil
			},
		},
		userAgent:     UserAgent,
		maxSize:       MaxFeedSize,
		skipSSRFCheck: false,
	}
}

// NewWithUserAgent creates a Crawler with a custom user agent
func NewWithUserAgent(userAgent string) *Crawler {
	c := New()
	if userAgent != "" {
		c.userAgent = userAgent
	}
	return c
}

// NewForTesting creates a Crawler that allows local URLs (for testing only)
func NewForTesting() *Crawler {
	c := New()
	c.skipSSRFCheck = true
	return c
}

// CrawlerConfig contains configuration options for HTTP connection pooling
type CrawlerConfig struct {
	UserAgent              string
	MaxIdleConns           int
	MaxIdleConnsPerHost    int
	MaxConnsPerHost        int
	IdleConnTimeoutSeconds int
}

// NewWithConfig creates a Crawler with custom configuration
func NewWithConfig(cfg CrawlerConfig) *Crawler {
	// Configure HTTP transport with custom connection pooling
	transport := &http.Transport{
		// Connection pooling settings
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:     cfg.MaxConnsPerHost,
		IdleConnTimeout:     time.Duration(cfg.IdleConnTimeoutSeconds) * time.Second,

		// Timeouts for connection establishment
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second, // TCP connection timeout
			KeepAlive: 30 * time.Second, // TCP keep-alive
		}).DialContext,

		// TLS handshake timeout
		TLSHandshakeTimeout: 10 * time.Second,

		// Response header timeout
		ResponseHeaderTimeout: 10 * time.Second,

		// Expect Continue timeout
		ExpectContinueTimeout: 1 * time.Second,
	}

	userAgent := cfg.UserAgent
	if userAgent == "" {
		userAgent = UserAgent
	}

	return &Crawler{
		client: &http.Client{
			Transport: transport,
			Timeout:   DefaultTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= MaxRedirects {
					return fmt.Errorf("stopped after %d redirects", MaxRedirects)
				}
				return nil
			},
		},
		userAgent:     userAgent,
		maxSize:       MaxFeedSize,
		skipSSRFCheck: false,
	}
}

// ValidateURL checks if a URL is safe to fetch (SSRF prevention)
func ValidateURL(rawURL string) error {
	// Handle empty string explicitly
	if rawURL == "" {
		return ErrInvalidURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	// url.Parse can succeed but return invalid URLs (e.g., "not a url" parses but has no scheme)
	// Explicitly check for presence of scheme
	if parsed.Scheme == "" {
		return ErrInvalidURL
	}

	// Only allow http and https
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return ErrInvalidScheme
	}

	// Check hostname
	host := parsed.Hostname()

	// Block known internal hostnames
	internalHosts := []string{"localhost", "127.0.0.1", "::1", "0.0.0.0"}
	for _, blocked := range internalHosts {
		if strings.EqualFold(host, blocked) {
			return ErrPrivateIP
		}
	}

	// Try to parse as IP address
	ip := net.ParseIP(host)
	if ip != nil {
		// Block loopback
		if ip.IsLoopback() {
			return ErrPrivateIP
		}
		// Block private networks (RFC 1918)
		if ip.IsPrivate() {
			return ErrPrivateIP
		}
		// Block link-local
		if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return ErrPrivateIP
		}
	}

	return nil
}

// Fetch fetches a feed with conditional request support
func (c *Crawler) Fetch(ctx context.Context, feedURL string, cache FeedCache) (*FeedResponse, error) {
	// Validate URL for SSRF prevention (unless testing mode)
	if !c.skipSSRFCheck {
		if err := ValidateURL(feedURL); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set User-Agent
	req.Header.Set("User-Agent", c.userAgent)

	// Set conditional request headers if we have cached values
	if cache.LastModified != "" {
		req.Header.Set("If-Modified-Since", cache.LastModified)
	}
	if cache.ETag != "" {
		req.Header.Set("If-None-Match", cache.ETag)
	}

	// Request compression
	req.Header.Set("Accept-Encoding", "gzip, deflate")

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	// Prepare response
	fetchTime := time.Now()
	finalURL := resp.Request.URL.String()

	// Handle 304 Not Modified
	if resp.StatusCode == http.StatusNotModified {
		return &FeedResponse{
			StatusCode:  resp.StatusCode,
			NotModified: true,
			NewCache: FeedCache{
				URL:          finalURL,
				ETag:         cache.ETag,
				LastModified: cache.LastModified,
				LastFetched:  fetchTime,
			},
			FinalURL:  finalURL,
			FetchTime: fetchTime,
		}, nil
	}

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		return &FeedResponse{
			StatusCode: resp.StatusCode,
			FinalURL:   finalURL,
			FetchTime:  fetchTime,
		}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Handle gzip decompression if needed
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// Limit response body size - add 1 to detect when limit is exceeded
	limitedReader := io.LimitedReader{
		R: reader,
		N: c.maxSize + 1,
	}

	// Read body
	body, err := io.ReadAll(&limitedReader)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	// Check if we exceeded the size limit
	// We set limitedReader.N to maxSize + 1
	// If we read more than maxSize bytes, the body length will exceed maxSize
	if int64(len(body)) > c.maxSize {
		return nil, ErrMaxSizeExceeded
	}

	// Extract new cache headers (EXACTLY as received)
	newCache := FeedCache{
		URL:          finalURL,
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
		LastFetched:  fetchTime,
	}

	return &FeedResponse{
		Body:        body,
		StatusCode:  resp.StatusCode,
		NotModified: false,
		NewCache:    newCache,
		FinalURL:    finalURL,
		FetchTime:   fetchTime,
	}, nil
}

// FetchWithRetry attempts to fetch with exponential backoff
func (c *Crawler) FetchWithRetry(ctx context.Context, feedURL string, cache FeedCache, maxRetries int) (*FeedResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s, 8s...
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		resp, err := c.Fetch(ctx, feedURL, cache)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't retry on certain errors
		if errors.Is(err, ErrInvalidURL) ||
			errors.Is(err, ErrPrivateIP) ||
			errors.Is(err, ErrInvalidScheme) ||
			errors.Is(err, ErrMaxSizeExceeded) {
			return nil, err
		}

		// Don't retry on 4xx client errors (except 429)
		if resp != nil && resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
			return resp, err
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
