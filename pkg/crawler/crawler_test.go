package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestValidateURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid http", "http://example.com/feed", false},
		{"valid https", "https://example.com/feed", false},
		{"localhost rejected", "http://localhost/feed", true},
		{"127.0.0.1 rejected", "http://127.0.0.1/feed", true},
		{"::1 rejected", "http://[::1]/feed", true},
		{"0.0.0.0 rejected", "http://0.0.0.0/feed", true},
		{"private 10.x rejected", "http://10.0.0.1/feed", true},
		{"private 192.168.x rejected", "http://192.168.1.1/feed", true},
		{"private 172.16.x rejected", "http://172.16.0.1/feed", true},
		{"link-local rejected", "http://169.254.169.254/feed", true},
		{"ftp scheme rejected", "ftp://example.com/feed", true},
		{"file scheme rejected", "file:///etc/passwd", true},
		{"invalid URL", "not a url", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFetch(t *testing.T) {
	t.Parallel()
	t.Run("successful fetch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify User-Agent
			if !strings.Contains(r.Header.Get("User-Agent"), "RoguePlanet") {
				t.Errorf("User-Agent not set correctly: %s", r.Header.Get("User-Agent"))
			}

			w.Header().Set("ETag", `"abc123"`)
			w.Header().Set("Last-Modified", "Mon, 02 Jan 2006 15:04:05 GMT")
			w.Header().Set("Content-Type", "application/rss+xml")
			if _, err := w.Write([]byte(`<?xml version="1.0"?><rss version="2.0"><channel></channel></rss>`)); err != nil {
				t.Errorf("Write error: %v", err)
			}
		}))
		defer server.Close()

		crawler := NewForTesting()
		resp, err := crawler.Fetch(context.Background(), server.URL, FeedCache{})

		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
		}

		if resp.NotModified {
			t.Error("NotModified = true, want false")
		}

		if resp.NewCache.ETag != `"abc123"` {
			t.Errorf("ETag = %q, want %q", resp.NewCache.ETag, `"abc123"`)
		}

		if resp.NewCache.LastModified != "Mon, 02 Jan 2006 15:04:05 GMT" {
			t.Errorf("LastModified = %q, want %q", resp.NewCache.LastModified, "Mon, 02 Jan 2006 15:04:05 GMT")
		}

		if len(resp.Body) == 0 {
			t.Error("Body is empty")
		}
	})

	t.Run("304 not modified", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check conditional headers
			if r.Header.Get("If-None-Match") != `"abc123"` {
				t.Errorf("If-None-Match = %q, want %q", r.Header.Get("If-None-Match"), `"abc123"`)
			}
			if r.Header.Get("If-Modified-Since") != "Mon, 02 Jan 2006 15:04:05 GMT" {
				t.Errorf("If-Modified-Since = %q, want %q", r.Header.Get("If-Modified-Since"), "Mon, 02 Jan 2006 15:04:05 GMT")
			}

			w.WriteHeader(http.StatusNotModified)
		}))
		defer server.Close()

		crawler := NewForTesting()
		cache := FeedCache{
			URL:          server.URL,
			ETag:         `"abc123"`,
			LastModified: "Mon, 02 Jan 2006 15:04:05 GMT",
			LastFetched:  time.Now(),
		}

		resp, err := crawler.Fetch(context.Background(), server.URL, cache)

		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}

		if !resp.NotModified {
			t.Error("NotModified = false, want true")
		}

		if resp.StatusCode != 304 {
			t.Errorf("StatusCode = %d, want 304", resp.StatusCode)
		}

		// Cache should be preserved
		if resp.NewCache.ETag != `"abc123"` {
			t.Errorf("ETag = %q, want %q", resp.NewCache.ETag, `"abc123"`)
		}
	})

	t.Run("redirect handling", func(t *testing.T) {
		finalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := w.Write([]byte(`<?xml version="1.0"?><rss version="2.0"><channel></channel></rss>`)); err != nil {
				t.Errorf("Write error: %v", err)
			}
		}))
		defer finalServer.Close()

		redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, finalServer.URL, http.StatusMovedPermanently)
		}))
		defer redirectServer.Close()

		crawler := NewForTesting()
		resp, err := crawler.Fetch(context.Background(), redirectServer.URL, FeedCache{})

		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}

		if resp.FinalURL == redirectServer.URL {
			t.Error("FinalURL should be different after redirect")
		}

		if !strings.Contains(resp.FinalURL, finalServer.URL) {
			t.Errorf("FinalURL = %s, want to contain %s", resp.FinalURL, finalServer.URL)
		}
	})

	t.Run("SSRF prevention", func(t *testing.T) {
		crawler := New()

		dangerousURLs := []string{
			"http://localhost/feed",
			"http://127.0.0.1/feed",
			"http://[::1]/feed",
			"http://10.0.0.1/feed",
			"http://192.168.1.1/feed",
			"http://169.254.169.254/latest/meta-data/",
		}

		for _, url := range dangerousURLs {
			_, err := crawler.Fetch(context.Background(), url, FeedCache{})
			if err == nil {
				t.Errorf("Expected error for dangerous URL %s, got nil", url)
			}
		}
	})

	t.Run("timeout handling", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
			if _, err := w.Write([]byte("too slow")); err != nil {
				t.Errorf("Write error: %v", err)
			}
		}))
		defer server.Close()

		crawler := NewForTesting()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := crawler.Fetch(ctx, server.URL, FeedCache{})
		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
	})

	t.Run("max size limit", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Send 11MB of data
			largeData := make([]byte, 11*1024*1024)
			if _, err := w.Write(largeData); err != nil {
				// Expected: client closes connection when size limit is reached
				if !strings.Contains(err.Error(), "connection reset") && !strings.Contains(err.Error(), "broken pipe") {
					t.Errorf("Unexpected write error: %v", err)
				}
			}
		}))
		defer server.Close()

		crawler := NewForTesting()
		_, err := crawler.Fetch(context.Background(), server.URL, FeedCache{})

		if err == nil {
			t.Error("Expected max size error, got nil")
		}

		if !strings.Contains(err.Error(), "maximum size") {
			t.Errorf("Expected 'maximum size' error, got: %v", err)
		}
	})
}

func TestFetchWithRetry(t *testing.T) {
	t.Parallel()
	t.Run("succeeds on second attempt", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if _, err := w.Write([]byte(`<?xml version="1.0"?><rss version="2.0"><channel></channel></rss>`)); err != nil {
				t.Errorf("Write error: %v", err)
			}
		}))
		defer server.Close()

		crawler := NewForTesting()
		resp, err := crawler.FetchWithRetry(context.Background(), server.URL, FeedCache{}, 3)

		if err != nil {
			t.Fatalf("FetchWithRetry() error = %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
		}

		if attempts != 2 {
			t.Errorf("attempts = %d, want 2", attempts)
		}
	})

	t.Run("no retry on SSRF error", func(t *testing.T) {
		crawler := New()
		_, err := crawler.FetchWithRetry(context.Background(), "http://localhost/feed", FeedCache{}, 3)

		if err == nil {
			t.Error("Expected SSRF error, got nil")
		}

		// Should fail immediately without retries
	})
}

func TestFetchInvalidContentType(t *testing.T) {
	t.Parallel()
	t.Run("HTML instead of feed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			if _, err := w.Write([]byte("<html><body>This is not a feed</body></html>")); err != nil {
				t.Errorf("Write error: %v", err)
			}
		}))
		defer server.Close()

		crawler := NewForTesting()
		resp, err := crawler.Fetch(context.Background(), server.URL, FeedCache{})

		// The crawler doesn't validate content-type (intentional - many feeds have incorrect headers)
		// It just fetches the data and lets the parser handle it
		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}

		// Response should contain the HTML data
		if len(resp.Body) == 0 {
			t.Error("Body is empty")
		}

		// Parser will fail later when trying to parse this as a feed
		if !strings.Contains(string(resp.Body), "This is not a feed") {
			t.Error("Expected HTML body to be returned")
		}
	})

	t.Run("JSON instead of XML feed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"error": "not a feed"}`)); err != nil {
				t.Errorf("Write error: %v", err)
			}
		}))
		defer server.Close()

		crawler := NewForTesting()
		resp, err := crawler.Fetch(context.Background(), server.URL, FeedCache{})

		// The crawler accepts any content-type
		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}

		if len(resp.Body) == 0 {
			t.Error("Body is empty")
		}
	})
}

// TestConnectionPooling verifies that the crawler is configured with proper connection pooling
func TestConnectionPooling(t *testing.T) {
	t.Parallel()
	crawler := New()

	// Verify client exists and has a transport
	if crawler.client == nil {
		t.Fatal("client is nil")
	}

	transport, ok := crawler.client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("client.Transport is not *http.Transport")
	}

	// Verify connection pooling settings
	tests := []struct {
		name     string
		got      interface{}
		want     interface{}
		testFunc func() bool
	}{
		{
			name: "MaxIdleConns",
			got:  transport.MaxIdleConns,
			want: 100,
			testFunc: func() bool {
				return transport.MaxIdleConns == 100
			},
		},
		{
			name: "MaxIdleConnsPerHost",
			got:  transport.MaxIdleConnsPerHost,
			want: 10,
			testFunc: func() bool {
				return transport.MaxIdleConnsPerHost == 10
			},
		},
		{
			name: "MaxConnsPerHost",
			got:  transport.MaxConnsPerHost,
			want: 20,
			testFunc: func() bool {
				return transport.MaxConnsPerHost == 20
			},
		},
		{
			name: "IdleConnTimeout",
			got:  transport.IdleConnTimeout,
			want: 90 * time.Second,
			testFunc: func() bool {
				return transport.IdleConnTimeout == 90*time.Second
			},
		},
		{
			name: "TLSHandshakeTimeout",
			got:  transport.TLSHandshakeTimeout,
			want: 10 * time.Second,
			testFunc: func() bool {
				return transport.TLSHandshakeTimeout == 10*time.Second
			},
		},
		{
			name: "ResponseHeaderTimeout",
			got:  transport.ResponseHeaderTimeout,
			want: 10 * time.Second,
			testFunc: func() bool {
				return transport.ResponseHeaderTimeout == 10*time.Second
			},
		},
		{
			name: "ExpectContinueTimeout",
			got:  transport.ExpectContinueTimeout,
			want: 1 * time.Second,
			testFunc: func() bool {
				return transport.ExpectContinueTimeout == 1*time.Second
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.testFunc() {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}

	// Verify DialContext is configured
	if transport.DialContext == nil {
		t.Error("DialContext is nil, should be configured")
	}
}

// TestConnectionReuseIntegration tests that connections are actually reused
func TestConnectionReuseIntegration(t *testing.T) {
	t.Parallel()
	connectionCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connectionCount++
		w.Header().Set("Content-Type", "application/rss+xml")
		if _, err := w.Write([]byte(`<?xml version="1.0"?><rss version="2.0"><channel><title>Test</title></channel></rss>`)); err != nil {
			t.Errorf("Write error: %v", err)
		}
	}))
	defer server.Close()

	crawler := NewForTesting()

	// Make multiple requests to the same server
	for i := 0; i < 5; i++ {
		resp, err := crawler.Fetch(context.Background(), server.URL, FeedCache{})
		if err != nil {
			t.Fatalf("Fetch #%d error: %v", i+1, err)
		}
		if resp.StatusCode != 200 {
			t.Errorf("Fetch #%d status = %d, want 200", i+1, resp.StatusCode)
		}
	}

	// All 5 requests should have been made
	if connectionCount != 5 {
		t.Errorf("Handler called %d times, want 5", connectionCount)
	}

	// Note: We can't directly verify connection reuse in this test because
	// httptest.Server doesn't expose connection metrics. The real benefit is
	// measured in production or with more sophisticated network benchmarks.
	// This test at least verifies that the configured transport doesn't
	// break normal request flow.
}

func TestParseRetryAfter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		header   string
		expected time.Duration
	}{
		{
			name:     "seconds format",
			header:   "120",
			expected: 120 * time.Second,
		},
		{
			name:     "seconds with whitespace",
			header:   "  60  ",
			expected: 60 * time.Second,
		},
		{
			name:     "HTTP-date format",
			header:   time.Now().Add(2 * time.Minute).Format(time.RFC1123),
			expected: 2 * time.Minute, // Approximately, we'll check range
		},
		{
			name:     "empty string",
			header:   "",
			expected: 0,
		},
		{
			name:     "invalid seconds",
			header:   "abc",
			expected: 0,
		},
		{
			name:     "negative seconds",
			header:   "-60",
			expected: 0,
		},
		{
			name:     "zero seconds",
			header:   "0",
			expected: 0,
		},
		{
			name:     "exceeds max (24 hours)",
			header:   "90000", // > 86400 seconds
			expected: 0,
		},
		{
			name:     "past HTTP-date",
			header:   time.Now().Add(-1 * time.Hour).Format(time.RFC1123),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRetryAfter(tt.header)

			// For HTTP-date tests, check within a reasonable range
			if strings.Contains(tt.name, "HTTP-date") && tt.expected > 0 {
				if result < tt.expected-5*time.Second || result > tt.expected+5*time.Second {
					t.Errorf("parseRetryAfter(%q) = %v, want approximately %v", tt.header, result, tt.expected)
				}
			} else {
				if result != tt.expected {
					t.Errorf("parseRetryAfter(%q) = %v, want %v", tt.header, result, tt.expected)
				}
			}
		})
	}
}

func TestFetchWithRetry_RespectsRetryAfter(t *testing.T) {
	t.Parallel()
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		if attemptCount < 3 {
			// Return 429 with Retry-After header
			w.Header().Set("Retry-After", "1") // 1 second
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		// Third attempt succeeds
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("<rss><channel><title>Test</title></channel></rss>"))
		if err != nil {
			t.Errorf("Write error: %v", err)
		}
	}))
	defer server.Close()

	crawler := NewForTesting()
	ctx := context.Background()

	startTime := time.Now()
	resp, err := crawler.FetchWithRetry(ctx, server.URL, FeedCache{}, 3)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("FetchWithRetry error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}

	if attemptCount != 3 {
		t.Errorf("attemptCount = %d, want 3", attemptCount)
	}

	// Should have waited at least 2 seconds (2 retries × 1 second Retry-After)
	// Allow some tolerance for timing
	if duration < 1800*time.Millisecond {
		t.Errorf("Duration = %v, expected at least 2 seconds (respecting Retry-After)", duration)
	}
}

func TestFetch_CapturesRetryAfter(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	crawler := NewForTesting()
	resp, err := crawler.Fetch(context.Background(), server.URL, FeedCache{})

	if err == nil {
		t.Fatal("Expected error for 429 response")
	}

	if resp == nil {
		t.Fatal("Expected response even with error")
	}

	if resp.StatusCode != 429 {
		t.Errorf("StatusCode = %d, want 429", resp.StatusCode)
	}

	expected := 60 * time.Second
	if resp.RetryAfter != expected {
		t.Errorf("RetryAfter = %v, want %v", resp.RetryAfter, expected)
	}
}

func TestFetch_Tracks301PermanentRedirect(t *testing.T) {
	t.Parallel()
	// Create a server that redirects with 301
	redirectCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/old-feed" {
			redirectCount++
			// 301 Permanent Redirect
			w.Header().Set("Location", "/new-feed")
			w.WriteHeader(http.StatusMovedPermanently)
			return
		}
		// Final destination
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("<rss><channel><title>Test</title></channel></rss>"))
		if err != nil {
			t.Errorf("Write error: %v", err)
		}
	}))
	defer server.Close()

	crawler := NewForTesting()
	resp, err := crawler.Fetch(context.Background(), server.URL+"/old-feed", FeedCache{})

	if err != nil {
		t.Fatalf("Fetch error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}

	if !resp.PermanentRedirect {
		t.Error("PermanentRedirect = false, want true (301 redirect was encountered)")
	}

	if !strings.HasSuffix(resp.FinalURL, "/new-feed") {
		t.Errorf("FinalURL = %s, want to end with /new-feed", resp.FinalURL)
	}

	if redirectCount != 1 {
		t.Errorf("redirectCount = %d, want 1", redirectCount)
	}
}

func TestFetch_Distinguishes301From302(t *testing.T) {
	t.Parallel()
	// Test that 302 redirects are NOT marked as permanent
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/temp-redirect" {
			// 302 Temporary Redirect
			w.Header().Set("Location", "/current-feed")
			w.WriteHeader(http.StatusFound)
			return
		}
		// Final destination
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("<rss><channel><title>Test</title></channel></rss>"))
		if err != nil {
			t.Errorf("Write error: %v", err)
		}
	}))
	defer server.Close()

	crawler := NewForTesting()
	resp, err := crawler.Fetch(context.Background(), server.URL+"/temp-redirect", FeedCache{})

	if err != nil {
		t.Fatalf("Fetch error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}

	if resp.PermanentRedirect {
		t.Error("PermanentRedirect = true, want false (302 redirect is temporary, not permanent)")
	}

	if !strings.HasSuffix(resp.FinalURL, "/current-feed") {
		t.Errorf("FinalURL = %s, want to end with /current-feed", resp.FinalURL)
	}
}
