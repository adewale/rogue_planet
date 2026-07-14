package normalizer

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// Tests for L14: Future date clamping

func TestFutureDateClamping(t *testing.T) {
	t.Parallel()
	n := New()

	fetchTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		pubDate         string // RFC3339
		expectedClamped bool
	}{
		{
			name:            "date 1 day in the future is clamped",
			pubDate:         fetchTime.Add(24 * time.Hour).Format(time.RFC3339),
			expectedClamped: true,
		},
		{
			name:            "date 2 hours in the future is clamped",
			pubDate:         fetchTime.Add(2 * time.Hour).Format(time.RFC3339),
			expectedClamped: true,
		},
		{
			name:            "date 30 minutes in the future is within tolerance",
			pubDate:         fetchTime.Add(30 * time.Minute).Format(time.RFC3339),
			expectedClamped: false,
		},
		{
			name:            "date in the past is not clamped",
			pubDate:         fetchTime.Add(-24 * time.Hour).Format(time.RFC3339),
			expectedClamped: false,
		},
		{
			name:            "date exactly at tolerance boundary is not clamped",
			pubDate:         fetchTime.Add(FutureDateTolerance).Format(time.RFC3339),
			expectedClamped: false,
		},
		{
			name:            "date just past tolerance is clamped",
			pubDate:         fetchTime.Add(FutureDateTolerance + time.Second).Format(time.RFC3339),
			expectedClamped: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedData := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Test Feed</title>
  <link href="https://example.com"/>
  <entry>
    <title>Future Entry</title>
    <link href="https://example.com/post"/>
    <id>future-entry-1</id>
    <published>%s</published>
    <updated>%s</updated>
    <content>Content</content>
  </entry>
</feed>`, tt.pubDate, tt.pubDate)

			_, entries, err := n.Parse(context.Background(), []byte(feedData), "https://example.com/feed", fetchTime)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(entries) != 1 {
				t.Fatalf("len(entries) = %d, want 1", len(entries))
			}

			entry := entries[0]

			if tt.expectedClamped {
				// Published should be clamped to fetchTime
				if !entry.Published.Equal(fetchTime) {
					t.Errorf("Published = %v, want %v (clamped to fetchTime)", entry.Published, fetchTime)
				}
				// Updated should also be clamped
				if !entry.Updated.Equal(fetchTime) {
					t.Errorf("Updated = %v, want %v (clamped to fetchTime)", entry.Updated, fetchTime)
				}
			} else {
				// Published should NOT be clamped
				if entry.Published.Equal(fetchTime) && tt.pubDate != fetchTime.Format(time.RFC3339) {
					t.Errorf("Published = %v, should NOT have been clamped to fetchTime", entry.Published)
				}
			}
		})
	}
}

func TestFutureDateTolerance_Constant(t *testing.T) {
	t.Parallel()

	// Verify the tolerance constant is 1 hour
	if FutureDateTolerance != 1*time.Hour {
		t.Errorf("FutureDateTolerance = %v, want 1h", FutureDateTolerance)
	}
}
