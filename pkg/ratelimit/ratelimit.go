// Package ratelimit provides per-domain rate limiting for HTTP requests.
//
// The rate limiter ensures the aggregator behaves as a good HTTP citizen
// by limiting requests per domain to prevent overwhelming servers.
package ratelimit

import (
	"context"
	"net/url"
	"sync"

	"golang.org/x/time/rate"
)

// Manager manages per-domain rate limiters
type Manager struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	limit    rate.Limit // Requests per second
	burst    int        // Burst size
}

// New creates a new rate limiter manager.
// requestsPerMinute: maximum requests per domain per minute
// burst: maximum burst size (allows temporary spikes)
func New(requestsPerMinute int, burst int) *Manager {
	// Convert requests/minute to requests/second
	reqPerSec := float64(requestsPerMinute) / 60.0

	return &Manager{
		limiters: make(map[string]*rate.Limiter),
		limit:    rate.Limit(reqPerSec),
		burst:    burst,
	}
}

// Wait blocks until the request to the given URL is allowed by the rate limiter.
// Returns an error if the context is cancelled while waiting.
func (m *Manager) Wait(ctx context.Context, feedURL string) error {
	domain, err := extractDomain(feedURL)
	if err != nil {
		// If we can't extract domain, allow the request (fail open)
		return nil
	}

	limiter := m.getLimiter(domain)
	return limiter.Wait(ctx)
}

// Allow checks if a request to the given URL would be allowed without blocking.
// Returns true if the request can proceed immediately, false if it would be rate-limited.
func (m *Manager) Allow(feedURL string) bool {
	domain, err := extractDomain(feedURL)
	if err != nil {
		// If we can't extract domain, allow the request (fail open)
		return true
	}

	limiter := m.getLimiter(domain)
	return limiter.Allow()
}

// getLimiter returns the rate limiter for a domain, creating it if necessary
func (m *Manager) getLimiter(domain string) *rate.Limiter {
	m.mu.RLock()
	limiter, exists := m.limiters[domain]
	m.mu.RUnlock()

	if exists {
		return limiter
	}

	// Create new limiter for this domain
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check again in case another goroutine created it while we were waiting for lock
	if limiter, exists := m.limiters[domain]; exists {
		return limiter
	}

	limiter = rate.NewLimiter(m.limit, m.burst)
	m.limiters[domain] = limiter
	return limiter
}

// extractDomain extracts the domain (host) from a URL
func extractDomain(feedURL string) (string, error) {
	parsed, err := url.Parse(feedURL)
	if err != nil {
		return "", err
	}

	// Return hostname without port
	return parsed.Hostname(), nil
}

// Stats returns statistics about the rate limiter
type Stats struct {
	TotalDomains int
	Limiters     map[string]LimiterStats
}

// LimiterStats contains statistics for a single domain's rate limiter
type LimiterStats struct {
	Domain            string
	TokensAvailable   int
	Burst             int
	RequestsPerMinute float64
}

// Stats returns current statistics about all rate limiters
func (m *Manager) Stats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := Stats{
		TotalDomains: len(m.limiters),
		Limiters:     make(map[string]LimiterStats),
	}

	for domain, limiter := range m.limiters {
		tokens := limiter.Tokens()
		stats.Limiters[domain] = LimiterStats{
			Domain:            domain,
			TokensAvailable:   int(tokens),
			Burst:             m.burst,
			RequestsPerMinute: float64(m.limit) * 60,
		}
	}

	return stats
}

// ResetAll clears all rate limiters (useful for testing)
func (m *Manager) ResetAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.limiters = make(map[string]*rate.Limiter)
}

// SetLimit updates the rate limit for all current and future limiters
func (m *Manager) SetLimit(requestsPerMinute int, burst int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	reqPerSec := float64(requestsPerMinute) / 60.0
	m.limit = rate.Limit(reqPerSec)
	m.burst = burst

	// Update existing limiters
	for domain := range m.limiters {
		m.limiters[domain] = rate.NewLimiter(m.limit, m.burst)
	}
}
