package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewWithUserAgent(t *testing.T) {
	tests := []struct {
		name              string
		userAgent         string
		expectedUserAgent string
	}{
		{
			name:              "custom user agent",
			userAgent:         "MyBot/1.0",
			expectedUserAgent: "MyBot/1.0",
		},
		{
			name:              "empty user agent uses default",
			userAgent:         "",
			expectedUserAgent: UserAgent,
		},
		{
			name:              "custom user agent with URL",
			userAgent:         "CustomPlanet/2.0 (+https://example.com)",
			expectedUserAgent: "CustomPlanet/2.0 (+https://example.com)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server that echoes back User-Agent header
			receivedUA := ""
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedUA = r.Header.Get("User-Agent")
				w.Header().Set("Content-Type", "application/rss+xml")
				w.Write([]byte(`<?xml version="1.0"?><rss version="2.0"><channel><title>Test</title></channel></rss>`))
			}))
			defer server.Close()

			// Create crawler with custom user agent
			c := NewWithUserAgent(tt.userAgent)
			c.skipSSRFCheck = true // Allow localhost for testing

			// Make request
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cache := FeedCache{URL: server.URL}
			_, err := c.Fetch(ctx, server.URL, cache)
			if err != nil {
				t.Fatalf("Fetch() error = %v", err)
			}

			// Verify User-Agent was sent correctly
			if receivedUA != tt.expectedUserAgent {
				t.Errorf("User-Agent = %q, want %q", receivedUA, tt.expectedUserAgent)
			}
		})
	}
}

func TestNewWithUserAgent_PreservesOtherSettings(t *testing.T) {
	c := NewWithUserAgent("CustomBot/1.0")

	// Verify other settings are preserved
	if c.maxSize != MaxFeedSize {
		t.Errorf("maxSize = %d, want %d", c.maxSize, MaxFeedSize)
	}

	if c.skipSSRFCheck != false {
		t.Errorf("skipSSRFCheck = %v, want false", c.skipSSRFCheck)
	}

	if c.client.Timeout != DefaultTimeout {
		t.Errorf("timeout = %v, want %v", c.client.Timeout, DefaultTimeout)
	}
}
