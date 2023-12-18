package rlutils

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/2manymws/rl"
)

// Helper function to generate a request context suitable for testing.
func createTestContext(limit int, windowLen time.Duration, next http.Handler) *rl.Context {
	return &rl.Context{
		StatusCode: http.StatusTooManyRequests,
		Err:        rl.ErrRateLimitExceeded,
		// Limiter, RequestLimit, and other fields would be initialized here as needed
		Next:               next,
		RequestLimit:       limit,
		RateLimitRemaining: 0,
		WindowLen:          windowLen,
	}
}
func TestBaseLimiter_isTargetExtensions(t *testing.T) {
	tests := []struct {
		name             string
		targetExtensions []string
		requestPath      string
		want             bool
	}{
		{
			name:             "Target extension .txt",
			targetExtensions: []string{".txt"},
			requestPath:      "/file.txt",
			want:             true,
		},
		{
			name:             "Target extension .jpg among others",
			targetExtensions: []string{".png", ".jpg", ".gif"},
			requestPath:      "/image.jpg",
			want:             true,
		},
		{
			name:             "Target extension case insensitive",
			targetExtensions: []string{".TXT"},
			requestPath:      "/file.txt",
			want:             true,
		},
		{
			name:             "No target extensions set",
			targetExtensions: []string{},
			requestPath:      "/file.txt",
			want:             true,
		},
		{
			name:             "Extension not in target list",
			targetExtensions: []string{".png", ".jpg", ".gif"},
			requestPath:      "/file.txt",
			want:             false,
		},
		{
			name:             "Request path without extension",
			targetExtensions: []string{".txt"},
			requestPath:      "/file",
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a BaseLimiter with specific target extensions.
			bl := &BaseLimiter{
				targetExtensions: tt.targetExtensions,
			}
			// Create an HTTP request with the test case path.
			req := httptest.NewRequest("GET", tt.requestPath, nil)
			// Check if the request extension is in the target list.
			got := bl.isTargetExtensions(req)
			if got != tt.want {
				t.Errorf("isTargetExtensions() = %v, want %v", got, tt.want)
			}
		})
	}
}
