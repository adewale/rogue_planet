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
