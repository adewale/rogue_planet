package crawler

import (
	"compress/gzip"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Comprehensive tests for SSRF prevention - all edge cases
func TestValidateURL_Comprehensive(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr error
	}{
		// Valid URLs
		{"valid http", "http://example.com/feed.xml", nil},
		{"valid https", "https://example.com/feed.xml", nil},
		{"valid with port", "http://example.com:8080/feed.xml", nil},
		{"valid public IPv4", "http://93.184.216.34/feed.xml", nil}, // example.com IP

		// Invalid schemes
		{"ftp scheme", "ftp://example.com/feed.xml", ErrInvalidScheme},
		{"file scheme", "file:///etc/passwd", ErrInvalidScheme},
		{"data URI", "data:text/html,<script>alert(1)</script>", ErrInvalidScheme},

		// Localhost variations
		{"localhost by name", "http://localhost/feed.xml", ErrPrivateIP},
		{"localhost uppercase", "http://LOCALHOST/feed.xml", ErrPrivateIP},
		{"127.0.0.1", "http://127.0.0.1/feed.xml", ErrPrivateIP},
		{"127.0.0.2", "http://127.0.0.2/feed.xml", ErrPrivateIP},
		{"IPv6 localhost", "http://[::1]/feed.xml", ErrPrivateIP},
		{"0.0.0.0", "http://0.0.0.0/feed.xml", ErrPrivateIP},

		// Private IPv4 ranges (RFC 1918)
		{"private 10.x.x.x", "http://10.0.0.1/feed.xml", ErrPrivateIP},
		{"private 10.255.255.255", "http://10.255.255.255/feed.xml", ErrPrivateIP},
		{"private 192.168.x.x", "http://192.168.1.1/feed.xml", ErrPrivateIP},
		{"private 192.168.255.255", "http://192.168.255.255/feed.xml", ErrPrivateIP},
		{"private 172.16.x.x", "http://172.16.0.1/feed.xml", ErrPrivateIP},
		{"private 172.31.x.x", "http://172.31.255.255/feed.xml", ErrPrivateIP},

		// Link-local addresses
		{"link-local IPv4", "http://169.254.0.1/feed.xml", ErrPrivateIP},
		{"AWS metadata service", "http://169.254.169.254/latest/meta-data/", ErrPrivateIP},
		{"link-local IPv6", "http://[fe80::1]/feed.xml", ErrPrivateIP},

		// Malformed URLs
		{"malformed URL", "http://[invalid", ErrInvalidURL},
		{"not a URL", "not a url", ErrInvalidURL},
		{"empty string", "", ErrInvalidURL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ValidateURL() error = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("ValidateURL() error = nil, want %v", tt.wantErr)
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateURL() error = %v, want %v", err, tt.wantErr)
				}
			}
		})
	}
}

// Test HTTP conditional requests with all ETag formats
func TestFetch_ConditionalRequests_ETagFormats(t *testing.T) {
	tests := []struct {
		name            string
		etag            string
		wantIfNoneMatch string
	}{
		{"ETag with quotes", `"abc123"`, `"abc123"`},
		{"ETag without quotes", `abc123`, `abc123`},
		{"Weak ETag", `W/"abc123"`, `W/"abc123"`},
		{"ETag with special chars", `"abc-123_456"`, `"abc-123_456"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ifNoneMatch := r.Header.Get("If-None-Match")
				if ifNoneMatch != tt.wantIfNoneMatch {
					t.Errorf("If-None-Match = %q, want %q", ifNoneMatch, tt.wantIfNoneMatch)
				}
				w.WriteHeader(http.StatusNotModified)
			}))
			defer server.Close()

			crawler := NewForTesting()
			cache := FeedCache{
				URL:         server.URL,
				ETag:        tt.etag,
				LastFetched: time.Now(),
			}

			_, err := crawler.Fetch(context.Background(), server.URL, cache)
			if err != nil {
				t.Fatalf("Fetch() error = %v", err)
			}
		})
	}
}

// Test conditional requests with only ETag, only Last-Modified, and both
func TestFetch_ConditionalRequests_Combinations(t *testing.T) {
	tests := []struct {
		name          string
		cache         FeedCache
		expectETag    bool
		expectLastMod bool
	}{
		{
			name: "only ETag",
			cache: FeedCache{
				ETag: `"abc123"`,
			},
			expectETag:    true,
			expectLastMod: false,
		},
		{
			name: "only Last-Modified",
			cache: FeedCache{
				LastModified: "Mon, 02 Jan 2006 15:04:05 GMT",
			},
			expectETag:    false,
			expectLastMod: true,
		},
		{
			name: "both ETag and Last-Modified",
			cache: FeedCache{
				ETag:         `"abc123"`,
				LastModified: "Mon, 02 Jan 2006 15:04:05 GMT",
			},
			expectETag:    true,
			expectLastMod: true,
		},
		{
			name:          "neither",
			cache:         FeedCache{},
			expectETag:    false,
			expectLastMod: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ifNoneMatch := r.Header.Get("If-None-Match")
				ifModifiedSince := r.Header.Get("If-Modified-Since")

				if tt.expectETag && ifNoneMatch == "" {
					t.Error("Expected If-None-Match header, got empty")
				}
				if !tt.expectETag && ifNoneMatch != "" {
					t.Errorf("Expected no If-None-Match header, got %q", ifNoneMatch)
				}
				if tt.expectLastMod && ifModifiedSince == "" {
					t.Error("Expected If-Modified-Since header, got empty")
				}
				if !tt.expectLastMod && ifModifiedSince != "" {
					t.Errorf("Expected no If-Modified-Since header, got %q", ifModifiedSince)
				}

				w.Header().Set("ETag", `"new123"`)
				w.Header().Set("Last-Modified", "Tue, 03 Jan 2006 15:04:05 GMT")
				if _, err := w.Write([]byte("content")); err != nil {
					t.Errorf("Write error: %v", err)
				}
			}))
			defer server.Close()

			crawler := NewForTesting()
			_, err := crawler.Fetch(context.Background(), server.URL, tt.cache)
			if err != nil {
				t.Fatalf("Fetch() error = %v", err)
			}
		})
	}
}

// Test gzip decompression
func TestFetch_GzipDecompression(t *testing.T) {
	originalContent := []byte("<?xml version=\"1.0\"?><rss version=\"2.0\"><channel><title>Test</title></channel></rss>")

	t.Run("gzip encoded response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check Accept-Encoding header
			acceptEncoding := r.Header.Get("Accept-Encoding")
			if !strings.Contains(acceptEncoding, "gzip") {
				t.Errorf("Accept-Encoding missing gzip: %s", acceptEncoding)
			}

			w.Header().Set("Content-Encoding", "gzip")
			gzWriter := gzip.NewWriter(w)
			if _, err := gzWriter.Write(originalContent); err != nil {
				t.Errorf("Write error: %v", err)
			}
			gzWriter.Close()
		}))
		defer server.Close()

		crawler := NewForTesting()
		resp, err := crawler.Fetch(context.Background(), server.URL, FeedCache{})

		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}

		if string(resp.Body) != string(originalContent) {
			t.Errorf("Body not decompressed correctly.\nGot: %s\nWant: %s", resp.Body, originalContent)
		}
	})

	t.Run("uncompressed response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := w.Write(originalContent); err != nil {
				t.Errorf("Write error: %v", err)
			}
		}))
		defer server.Close()

		crawler := NewForTesting()
		resp, err := crawler.Fetch(context.Background(), server.URL, FeedCache{})

		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}

		if string(resp.Body) != string(originalContent) {
			t.Errorf("Body = %s, want %s", resp.Body, originalContent)
		}
	})
}

// Test various HTTP status codes
func TestFetch_HTTPStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{"200 OK", http.StatusOK, false},
		{"304 Not Modified", http.StatusNotModified, false},
		{"404 Not Found", http.StatusNotFound, true},
		{"429 Too Many Requests", http.StatusTooManyRequests, true},
		{"500 Internal Server Error", http.StatusInternalServerError, true},
		{"502 Bad Gateway", http.StatusBadGateway, true},
		{"503 Service Unavailable", http.StatusServiceUnavailable, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					if _, err := w.Write([]byte("content")); err != nil {
						t.Errorf("Write error: %v", err)
					}
				}
			}))
			defer server.Close()

			crawler := NewForTesting()
			resp, err := crawler.Fetch(context.Background(), server.URL, FeedCache{})

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if resp != nil && resp.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", resp.StatusCode, tt.statusCode)
			}
		})
	}
}

// Test size limits
func TestFetch_SizeLimits(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{"under 10MB", 5 * 1024 * 1024, false},
		{"exactly 10MB", 10 * 1024 * 1024, false},
		{"over 10MB", 11 * 1024 * 1024, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data := make([]byte, tt.size)
				w.Write(data)
			}))
			defer server.Close()

			crawler := NewForTesting()
			_, err := crawler.Fetch(context.Background(), server.URL, FeedCache{})

			if tt.wantErr && err == nil {
				t.Error("Expected size limit error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

// Test redirect handling
func TestFetch_Redirects(t *testing.T) {
	t.Run("3 redirects - success", func(t *testing.T) {
		redirectCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if redirectCount < 3 {
				redirectCount++
				http.Redirect(w, r, "/redirect", http.StatusFound)
			} else {
				w.Write([]byte("final content"))
			}
		}))
		defer server.Close()

		crawler := NewForTesting()
		resp, err := crawler.Fetch(context.Background(), server.URL, FeedCache{})

		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}
		if redirectCount != 3 {
			t.Errorf("redirectCount = %d, want 3", redirectCount)
		}
		if string(resp.Body) != "final content" {
			t.Errorf("Body = %s, want 'final content'", resp.Body)
		}
	})

	t.Run("6 redirects - exceeds max", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/redirect", http.StatusFound)
		}))
		defer server.Close()

		crawler := NewForTesting()
		_, err := crawler.Fetch(context.Background(), server.URL, FeedCache{})

		if err == nil {
			t.Error("Expected redirect limit error, got nil")
		}
	})

	t.Run("301 permanent redirect", func(t *testing.T) {
		finalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("moved content"))
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
		if !strings.Contains(resp.FinalURL, finalServer.URL) {
			t.Errorf("FinalURL = %s, should contain %s", resp.FinalURL, finalServer.URL)
		}
	})
}

// Test FetchWithRetry comprehensive scenarios
func TestFetchWithRetry_Comprehensive(t *testing.T) {
	t.Run("success on first try - no retries", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			w.Write([]byte("content"))
		}))
		defer server.Close()

		crawler := NewForTesting()
		_, err := crawler.FetchWithRetry(context.Background(), server.URL, FeedCache{}, 3)

		if err != nil {
			t.Fatalf("FetchWithRetry() error = %v", err)
		}
		if attempts != 1 {
			t.Errorf("attempts = %d, want 1", attempts)
		}
	})

	t.Run("transient error, success on third try", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write([]byte("content"))
		}))
		defer server.Close()

		crawler := NewForTesting()
		_, err := crawler.FetchWithRetry(context.Background(), server.URL, FeedCache{}, 3)

		if err != nil {
			t.Fatalf("FetchWithRetry() error = %v", err)
		}
		if attempts != 3 {
			t.Errorf("attempts = %d, want 3", attempts)
		}
	})

	t.Run("max retries exceeded", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		crawler := NewForTesting()
		_, err := crawler.FetchWithRetry(context.Background(), server.URL, FeedCache{}, 2)

		if err == nil {
			t.Error("Expected max retries error, got nil")
		}
		if !strings.Contains(err.Error(), "max retries") {
			t.Errorf("Expected 'max retries' error, got: %v", err)
		}
	})

	t.Run("no retry on ErrInvalidURL", func(t *testing.T) {
		crawler := New()
		_, err := crawler.FetchWithRetry(context.Background(), "ftp://example.com/feed", FeedCache{}, 3)

		if err == nil {
			t.Error("Expected invalid scheme error, got nil")
		}
		if !errors.Is(err, ErrInvalidScheme) {
			t.Errorf("Expected ErrInvalidScheme, got %v", err)
		}
	})

	t.Run("no retry on ErrPrivateIP", func(t *testing.T) {
		crawler := New()
		_, err := crawler.FetchWithRetry(context.Background(), "http://127.0.0.1/feed", FeedCache{}, 3)

		if err == nil {
			t.Error("Expected private IP error, got nil")
		}
		if !errors.Is(err, ErrPrivateIP) {
			t.Errorf("Expected ErrPrivateIP, got %v", err)
		}
	})

	t.Run("no retry on 404", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		crawler := NewForTesting()
		_, err := crawler.FetchWithRetry(context.Background(), server.URL, FeedCache{}, 3)

		if err == nil {
			t.Error("Expected 404 error, got nil")
		}
		if attempts != 1 {
			t.Errorf("Should not retry on 404, attempts = %d", attempts)
		}
	})

	t.Run("retry on 429", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts == 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			w.Write([]byte("content"))
		}))
		defer server.Close()

		crawler := NewForTesting()
		_, err := crawler.FetchWithRetry(context.Background(), server.URL, FeedCache{}, 3)

		if err != nil {
			t.Fatalf("FetchWithRetry() error = %v", err)
		}
		if attempts < 2 {
			t.Errorf("Should retry on 429, attempts = %d", attempts)
		}
	})

	t.Run("context cancelled during retry", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		crawler := NewForTesting()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := crawler.FetchWithRetry(ctx, server.URL, FeedCache{}, 3)

		if err == nil {
			t.Error("Expected context cancelled error, got nil")
		}
	})
}

// Test constructor variants
func TestConstructors(t *testing.T) {
	t.Run("New() creates default crawler", func(t *testing.T) {
		crawler := New()
		if crawler == nil {
			t.Fatal("New() returned nil")
		}
		if crawler.userAgent != UserAgent {
			t.Errorf("userAgent = %s, want %s", crawler.userAgent, UserAgent)
		}
		if crawler.maxSize != MaxFeedSize {
			t.Errorf("maxSize = %d, want %d", crawler.maxSize, MaxFeedSize)
		}
		if crawler.skipSSRFCheck {
			t.Error("skipSSRFCheck should be false by default")
		}
	})

	t.Run("NewWithUserAgent() sets custom UA", func(t *testing.T) {
		customUA := "MyBot/1.0"
		crawler := NewWithUserAgent(customUA)
		if crawler.userAgent != customUA {
			t.Errorf("userAgent = %s, want %s", crawler.userAgent, customUA)
		}
	})

	t.Run("NewWithUserAgent('') uses default", func(t *testing.T) {
		crawler := NewWithUserAgent("")
		if crawler.userAgent != UserAgent {
			t.Errorf("userAgent = %s, want %s", crawler.userAgent, UserAgent)
		}
	})

	t.Run("NewForTesting() disables SSRF checks", func(t *testing.T) {
		crawler := NewForTesting()
		if !crawler.skipSSRFCheck {
			t.Error("skipSSRFCheck should be true for testing")
		}
	})
}

// Test cache update on successful response
func TestFetch_CacheUpdate(t *testing.T) {
	t.Run("cache updated on 200 response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("ETag", `"new-etag"`)
			w.Header().Set("Last-Modified", "Wed, 04 Jan 2006 15:04:05 GMT")
			w.Write([]byte("new content"))
		}))
		defer server.Close()

		crawler := NewForTesting()
		oldCache := FeedCache{
			URL:          server.URL,
			ETag:         `"old-etag"`,
			LastModified: "Mon, 02 Jan 2006 15:04:05 GMT",
			LastFetched:  time.Now().Add(-1 * time.Hour),
		}

		resp, err := crawler.Fetch(context.Background(), server.URL, oldCache)

		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}

		if resp.NewCache.ETag != `"new-etag"` {
			t.Errorf("NewCache.ETag = %q, want %q", resp.NewCache.ETag, `"new-etag"`)
		}
		if resp.NewCache.LastModified != "Wed, 04 Jan 2006 15:04:05 GMT" {
			t.Errorf("NewCache.LastModified = %q, want %q", resp.NewCache.LastModified, "Wed, 04 Jan 2006 15:04:05 GMT")
		}
		if resp.NewCache.LastFetched.Before(time.Now().Add(-1 * time.Second)) {
			t.Error("NewCache.LastFetched should be recent")
		}
	})

	t.Run("cache preserved on 304 response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotModified)
		}))
		defer server.Close()

		crawler := NewForTesting()
		oldCache := FeedCache{
			URL:          server.URL,
			ETag:         `"preserved-etag"`,
			LastModified: "Mon, 02 Jan 2006 15:04:05 GMT",
			LastFetched:  time.Now().Add(-1 * time.Hour),
		}

		resp, err := crawler.Fetch(context.Background(), server.URL, oldCache)

		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}

		if resp.NewCache.ETag != `"preserved-etag"` {
			t.Errorf("NewCache.ETag = %q, want %q (preserved)", resp.NewCache.ETag, `"preserved-etag"`)
		}
		if resp.NewCache.LastModified != "Mon, 02 Jan 2006 15:04:05 GMT" {
			t.Errorf("NewCache.LastModified = %q, want %q (preserved)", resp.NewCache.LastModified, "Mon, 02 Jan 2006 15:04:05 GMT")
		}
	})
}

// Test User-Agent header
func TestFetch_UserAgent(t *testing.T) {
	t.Run("default user agent", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ua := r.Header.Get("User-Agent")
			if ua != UserAgent {
				t.Errorf("User-Agent = %s, want %s", ua, UserAgent)
			}
			w.Write([]byte("content"))
		}))
		defer server.Close()

		crawler := NewForTesting()
		_, err := crawler.Fetch(context.Background(), server.URL, FeedCache{})
		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}
	})

	t.Run("custom user agent", func(t *testing.T) {
		customUA := "TestBot/2.0"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ua := r.Header.Get("User-Agent")
			if ua != customUA {
				t.Errorf("User-Agent = %s, want %s", ua, customUA)
			}
			w.Write([]byte("content"))
		}))
		defer server.Close()

		crawler := NewForTesting()
		crawler.userAgent = customUA
		_, err := crawler.Fetch(context.Background(), server.URL, FeedCache{})
		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}
	})
}
