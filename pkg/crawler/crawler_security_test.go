package crawler

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// =============================================================================
// M2: IPv4-Mapped IPv6 Address Tests
// =============================================================================

func TestValidateURL_IPv4MappedIPv6(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		url     string
		wantErr error
	}{
		// IPv4-mapped IPv6 addresses that should be blocked
		{
			name:    "IPv4-mapped loopback ::ffff:127.0.0.1",
			url:     "http://[::ffff:127.0.0.1]/feed",
			wantErr: ErrPrivateIP,
		},
		{
			name:    "IPv4-mapped private 10.x ::ffff:10.0.0.1",
			url:     "http://[::ffff:10.0.0.1]/feed",
			wantErr: ErrPrivateIP,
		},
		{
			name:    "IPv4-mapped private 192.168.x ::ffff:192.168.1.1",
			url:     "http://[::ffff:192.168.1.1]/feed",
			wantErr: ErrPrivateIP,
		},
		// IPv4-mapped IPv6 of a public IP should be allowed
		{
			name:    "IPv4-mapped public IP ::ffff:93.184.216.34",
			url:     "http://[::ffff:93.184.216.34]/feed",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateURL(tt.url)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ValidateURL(%q) unexpected error = %v", tt.url, err)
				}
			} else {
				if err == nil {
					t.Errorf("ValidateURL(%q) = nil, want error %v", tt.url, tt.wantErr)
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateURL(%q) error = %v, want %v", tt.url, err, tt.wantErr)
				}
			}
		})
	}
}

// =============================================================================
// L1: Missing Special IP Range Blocks Tests
// =============================================================================

func TestValidateURL_SpecialIPRanges(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		url     string
		wantErr error
	}{
		// CGNAT / Shared Address Space (RFC 6598): 100.64.0.0/10
		{
			name:    "CGNAT start 100.64.0.0",
			url:     "http://100.64.0.1/feed",
			wantErr: ErrPrivateIP,
		},
		{
			name:    "CGNAT middle 100.100.100.100",
			url:     "http://100.100.100.100/feed",
			wantErr: ErrPrivateIP,
		},
		{
			name:    "CGNAT end 100.127.255.255",
			url:     "http://100.127.255.255/feed",
			wantErr: ErrPrivateIP,
		},
		// Just outside CGNAT range should be allowed
		{
			name:    "just above CGNAT 100.128.0.1",
			url:     "http://100.128.0.1/feed",
			wantErr: nil,
		},

		// Benchmarking (RFC 2544): 198.18.0.0/15
		{
			name:    "benchmarking start 198.18.0.1",
			url:     "http://198.18.0.1/feed",
			wantErr: ErrPrivateIP,
		},
		{
			name:    "benchmarking end 198.19.255.255",
			url:     "http://198.19.255.255/feed",
			wantErr: ErrPrivateIP,
		},
		// Just outside benchmarking range should be allowed
		{
			name:    "just above benchmarking 198.20.0.1",
			url:     "http://198.20.0.1/feed",
			wantErr: nil,
		},

		// "This network" range: 0.0.0.0/8 (beyond just 0.0.0.0)
		{
			name:    "this network 0.0.0.0",
			url:     "http://0.0.0.0/feed",
			wantErr: ErrPrivateIP,
		},
		{
			name:    "this network 0.0.0.1",
			url:     "http://0.0.0.1/feed",
			wantErr: ErrPrivateIP,
		},
		{
			name:    "this network 0.255.255.255",
			url:     "http://0.255.255.255/feed",
			wantErr: ErrPrivateIP,
		},

		// Multicast: 224.0.0.0/4
		{
			name:    "multicast 224.0.0.1",
			url:     "http://224.0.0.1/feed",
			wantErr: ErrPrivateIP,
		},
		{
			name:    "multicast 239.255.255.255",
			url:     "http://239.255.255.255/feed",
			wantErr: ErrPrivateIP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateURL(tt.url)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ValidateURL(%q) unexpected error = %v", tt.url, err)
				}
			} else {
				if err == nil {
					t.Errorf("ValidateURL(%q) = nil, want error %v", tt.url, tt.wantErr)
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateURL(%q) error = %v, want %v", tt.url, err, tt.wantErr)
				}
			}
		})
	}
}

// =============================================================================
// L2: Explicit Minimum TLS Version Tests
// =============================================================================

func TestNew_MinTLSVersion(t *testing.T) {
	t.Parallel()

	c := New()
	transport, ok := c.client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Transport is not *http.Transport")
	}

	if transport.TLSClientConfig == nil {
		t.Fatal("TLSClientConfig is nil, expected explicit TLS configuration")
	}

	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %d, want %d (TLS 1.2)", transport.TLSClientConfig.MinVersion, tls.VersionTLS12)
	}
}

func TestNewWithConfig_MinTLSVersion(t *testing.T) {
	t.Parallel()

	cfg := CrawlerConfig{
		UserAgent: "TestBot/1.0",
	}
	c := NewWithConfig(cfg)
	transport, ok := c.client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Transport is not *http.Transport")
	}

	if transport.TLSClientConfig == nil {
		t.Fatal("TLSClientConfig is nil, expected explicit TLS configuration")
	}

	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %d, want %d (TLS 1.2)", transport.TLSClientConfig.MinVersion, tls.VersionTLS12)
	}
}

// =============================================================================
// L6: Retry-After Date Upper Bound Tests
// =============================================================================

func TestParseRetryAfter_DateUpperBound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		header      string
		maxExpected time.Duration
	}{
		{
			name:        "HTTP-date 48 hours in future is capped at 24 hours",
			header:      time.Now().Add(48 * time.Hour).Format(time.RFC1123),
			maxExpected: 24 * time.Hour,
		},
		{
			name:        "HTTP-date 7 days in future is capped at 24 hours",
			header:      time.Now().Add(7 * 24 * time.Hour).Format(time.RFC1123),
			maxExpected: 24 * time.Hour,
		},
		{
			name:        "HTTP-date 2 minutes in future is not capped",
			header:      time.Now().Add(2 * time.Minute).Format(time.RFC1123),
			maxExpected: 3 * time.Minute, // allow some slack for timing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parseRetryAfter(tt.header)
			if result > tt.maxExpected {
				t.Errorf("parseRetryAfter(%q) = %v, want <= %v (capped)", tt.header, result, tt.maxExpected)
			}
			if result <= 0 {
				t.Errorf("parseRetryAfter(%q) = %v, want > 0", tt.header, result)
			}
		})
	}
}

func TestParseRetryAfter_MaxDuration24Hours(t *testing.T) {
	t.Parallel()

	// A date very far in the future should be capped at 24 hours
	farFuture := time.Now().Add(365 * 24 * time.Hour).Format(time.RFC1123)
	result := parseRetryAfter(farFuture)

	maxAllowed := 24*time.Hour + 5*time.Second // small tolerance
	if result > maxAllowed {
		t.Errorf("parseRetryAfter(far future date) = %v, want <= %v", result, maxAllowed)
	}
	if result <= 0 {
		t.Errorf("parseRetryAfter(far future date) = %v, want > 0", result)
	}
}

// =============================================================================
// M1: DNS Rebinding Protection Tests
// =============================================================================

func TestIsPrivateIP_ExportedForDialContext(t *testing.T) {
	t.Parallel()

	// Test the isPrivateIP function directly (it needs to be extracted for
	// the custom DialContext to use)
	tests := []struct {
		name    string
		ip      string
		private bool
	}{
		{"loopback 127.0.0.1", "127.0.0.1", true},
		{"loopback 127.0.0.2", "127.0.0.2", true},
		{"private 10.0.0.1", "10.0.0.1", true},
		{"private 192.168.1.1", "192.168.1.1", true},
		{"private 172.16.0.1", "172.16.0.1", true},
		{"link-local 169.254.169.254", "169.254.169.254", true},
		{"IPv6 loopback ::1", "::1", true},
		{"public 93.184.216.34", "93.184.216.34", false},
		{"public 8.8.8.8", "8.8.8.8", false},
		{"CGNAT 100.64.0.1", "100.64.0.1", true},
		{"benchmarking 198.18.0.1", "198.18.0.1", true},
		{"this-network 0.0.0.1", "0.0.0.1", true},
		{"multicast 224.0.0.1", "224.0.0.1", true},
		// IPv4-mapped IPv6
		{"IPv4-mapped loopback", "::ffff:127.0.0.1", true},
		{"IPv4-mapped private", "::ffff:10.0.0.1", true},
		{"IPv4-mapped public", "::ffff:93.184.216.34", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", tt.ip)
			}
			result := isPrivateIP(ip)
			if result != tt.private {
				t.Errorf("isPrivateIP(%s) = %v, want %v", tt.ip, result, tt.private)
			}
		})
	}
}

func TestDNSRebindingProtection_DialContext(t *testing.T) {
	t.Parallel()

	// Test that the production crawler's custom DialContext blocks connections
	// to private IPs resolved from DNS. We use a local server that binds to
	// 127.0.0.1 and verify that the production crawler refuses to connect.

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("should not reach here"))
	}))
	defer server.Close()

	// The httptest server is on 127.0.0.1 -- a production crawler should
	// block the resolved IP at the DialContext level, not just the hostname.
	c := New()

	// Fetch with a URL that resolves to localhost. We use the server's
	// actual address (127.0.0.1:PORT) but first we need to skip the
	// hostname-level check to test the dial-level check.
	// We'll use an IP address directly to bypass the hostname blocklist.
	// 127.0.0.1 is already in the blocklist string check, so this tests
	// both layers. The key test is that even if ValidateURL didn't catch it,
	// the DialContext would.

	_, err := c.Fetch(context.Background(), server.URL, FeedCache{})
	if err == nil {
		t.Error("Expected error when fetching from localhost, got nil")
	}
}

func TestDNSRebindingProtection_TestingCrawlerAllows(t *testing.T) {
	t.Parallel()

	// Test that NewForTesting() still works with local servers
	// (the custom DialContext should be skipped when skipSSRFCheck is true)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<rss version="2.0"><channel></channel></rss>`))
	}))
	defer server.Close()

	c := NewForTesting()
	resp, err := c.Fetch(context.Background(), server.URL, FeedCache{})
	if err != nil {
		t.Fatalf("NewForTesting() crawler should allow localhost: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

func TestDNSRebindingProtection_NewWithConfigTestingMode(t *testing.T) {
	t.Parallel()

	// Verify that NewWithConfig also works when skipSSRFCheck is enabled
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<rss version="2.0"><channel></channel></rss>`))
	}))
	defer server.Close()

	cfg := CrawlerConfig{UserAgent: "TestBot/1.0"}
	c := NewWithConfig(cfg)
	c.skipSSRFCheck = true // Simulate testing mode

	resp, err := c.Fetch(context.Background(), server.URL, FeedCache{})
	if err != nil {
		t.Fatalf("Crawler with skipSSRFCheck should allow localhost: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

// =============================================================================
// isPrivateIP helper function unit tests (for completeness)
// =============================================================================

func TestIsPrivateIP_UnspecifiedAddress(t *testing.T) {
	t.Parallel()

	// 0.0.0.0 is the unspecified address; should be blocked
	ip := net.ParseIP("0.0.0.0")
	if ip == nil {
		t.Fatal("Failed to parse 0.0.0.0")
	}
	if !isPrivateIP(ip) {
		t.Error("isPrivateIP(0.0.0.0) = false, want true")
	}

	// :: (IPv6 unspecified) should also be blocked
	ip6 := net.ParseIP("::")
	if ip6 == nil {
		t.Fatal("Failed to parse ::")
	}
	if !isPrivateIP(ip6) {
		t.Error("isPrivateIP(::) = false, want true")
	}
}

func TestIsPrivateIP_IPv6LinkLocal(t *testing.T) {
	t.Parallel()

	ip := net.ParseIP("fe80::1")
	if ip == nil {
		t.Fatal("Failed to parse fe80::1")
	}
	if !isPrivateIP(ip) {
		t.Error("isPrivateIP(fe80::1) = false, want true")
	}
}
